// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cwmp

import (
	"context"
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"github.com/nokia/corteca-cli/internal/tui"
	"encoding/xml"
	"errors"

	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultCWMPPort = 7547
)

func init() {
	device.RegisterDeviceType("cwmp", NewCWMPDevice)
	device.RegisterDeviceType("cwmps", NewCWMPDevice)
}

type CWMPDevice struct {
	server    *http.Server
	in        chan messages.Message
	out       chan *messages.Envelope
	log       io.Writer
	currentID string
}

type CWMPConfig struct {
	configuration.HttpClientEndpoint `yaml:",inline"`
	Server                           configuration.HttpServerEndpoint `yaml:"server"`
}

func (d *CWMPDevice) NewSessionID() {
	if uuid, err := uuid.NewV7(); err != nil {
		panic(err)
	} else {
		d.currentID = uuid.String()
	}
}

func (d *CWMPDevice) SetSessionID(id string) {
	d.currentID = id
}

func (d *CWMPDevice) ResetSessionID() {
	d.currentID = ""
}

func NewCWMPDevice(c *configuration.DeviceConfig, log io.Writer) (device.Device, error) {
	cwmpconfig := CWMPConfig{}
	if err := c.Decode(&cwmpconfig); err != nil {
		return nil, err
	}

	d := CWMPDevice{
		log:       log,
		in:        make(chan messages.Message),
		out:       make(chan *messages.Envelope),
		currentID: "",
	}
	if err := d.initServer(&cwmpconfig.Server); err != nil {
		return nil, err
	}
	if err := d.sendConnectionRequest(&cwmpconfig.HttpClientEndpoint); err != nil {
		tui.LogError("Failed sending connection request: %s", err.Error())
	}
	tui.DisplaySuccessMsg("Waiting for CPE to establish connection...")
	return &d, nil
}

func (d *CWMPDevice) BeginSequence() error {
	d.ResetSessionID()
	ctx, cancel := context.WithTimeout(context.Background(), configuration.DefaultMaxTimeout)
	defer cancel()
	tui.LogNormal("Waiting for (ready) message...")
	if _, err := d.expectRPC(ctx, func(m messages.Message) bool { return m == nil }); err != nil {
		return err
	}
	return nil
}

func (d *CWMPDevice) ExecuteCommand(ctx context.Context, cmd *configuration.SequenceCmd) (any, error) {
	rpc, err := d.createRPCFromCmd(cmd)
	if err != nil {
		return nil, err
	} else {
		d.NewSessionID()
		tui.LogNormal("Sending '%s' RPC...", rpc.GetName())
		env := d.newEnvelope(rpc)
		if err := d.pushEnvelope(ctx, &env); err != nil {
			return nil, err
		}
	}

	tui.LogNormal("Waiting for response...")
	resp, err := d.pullMessage(ctx)
	if err != nil {
		return nil, err
	}
	if fault, ok := resp.(messages.Fault); ok {
		return nil, fmt.Errorf("%s (faultcode: %d)", fault.Detail.FaultString, fault.Detail.FaultCode)
	} else if err := rpc.ValidateResponse(resp); err != nil {
		return nil, err
	}

	d.ResetSessionID()
	if async, ok := rpc.(messages.AsyncRPC); ok {
		return d.handleAsyncRPC(ctx, async)
	}
	return resp, nil
}

func (d *CWMPDevice) EndSequence() error {
	ctx, cancel := context.WithTimeout(context.Background(), configuration.DefaultMaxTimeout)
	defer cancel()
	return d.pushEnvelope(ctx, nil)
}

func (d *CWMPDevice) GetProtocol() string {
	return "cwmp"
}

func (c *CWMPDevice) Close() {
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil {
		tui.LogError("Server Shutdown Failed: %s", err)
		return
	}
	tui.LogNormal("Server stopped gracefully!")
}

func (d *CWMPDevice) handleAsyncRPC(ctx context.Context, rpc messages.AsyncRPC) (messages.Message, error) {
	tui.LogNormal("Expecting async notification for '%s'...", rpc.GetName())
	// send blank response to end session
	if err := d.pushEnvelope(ctx, nil); err != nil {
		return nil, err
	}
	// wait until an RPC with the same command key arrives
	notif, err := d.expectRPC(ctx, func(m messages.Message) bool { return rpc.Match(m) })
	if err != nil {
		return nil, err
	}
	// respond to the notification RPC
	if err := d.pushEnvelope(ctx, d.respondToRPC(notif)); err != nil {
		return nil, err
	}
	// wait until a "ready" message arrives
	if _, err := d.expectRPC(ctx, func(m messages.Message) bool { return m == nil }); err != nil {
		return nil, err
	}
	// return notification RPC payload
	return notif, rpc.ValidateResponse(notif)
}

func (c *CWMPDevice) sendConnectionRequest(cpe *configuration.HttpClientEndpoint) error {
	client, err := cpe.NewHttpClient()
	if err != nil {
		return err
	}
	url, err := url.Parse(cpe.Addr.String())
	if err != nil {
		return err
	}
	// convert the scheme to https
	switch url.Scheme {
	case "cwmp":
		url.Scheme = "http"
	case "cwmps":
		url.Scheme = "https"
	default:
		panic(fmt.Sprintf("unexpected scheme '%s'", url.Scheme))
	}
	if url.Port() == "" {
		url.Host = net.JoinHostPort(url.Host, strconv.FormatInt(DefaultCWMPPort, 10))
	}
	resp, err := client.Get(url.String())
	if err != nil {
		return fmt.Errorf("error sending connection request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = c.log.Write(fmt.Appendf([]byte(""), "[%s] Connection Request (response: %s)\n", time.Now().Format(time.DateTime), resp.Status))
	_, _ = io.Copy(io.Discard, resp.Body)
	tui.LogNormal("Connection Request sent to %s; status code: %d", url.String(), resp.StatusCode)
	return nil
}

func (d *CWMPDevice) initServer(config *configuration.HttpServerEndpoint) error {

	// TODO: if no url or server is empty we should assume listen on http://0.0.0.0:DefaultCWMPPort
	u, err := url.Parse(config.Addr.String())
	if err != nil {
		return err
	}

	if u.Port() == "" {
		u.Host = net.JoinHostPort(u.Host, strconv.Itoa(DefaultCWMPPort))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleHTTPRequest)
	d.server = &http.Server{
		Addr:    u.Host,
		Handler: mux,
	}

	tui.DisplaySuccessMsg(fmt.Sprintf("Starting CWMP server on %s...", d.server.Addr))
	// Run server in a goroutine
	go func() {
		var err error
		switch u.Scheme {
		case "http":
			err = d.server.ListenAndServe()
		case "https":
			err = d.server.ListenAndServeTLS(config.Certificate.String(), config.Key.String())
		}
		if err != nil && err != http.ErrServerClosed {
			tui.LogError("Failed to start server: %s", err)
			os.Exit(1)
		}
	}()

	return nil
}

func (d *CWMPDevice) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: implement pullMessage and pullEnvelope to be context cancellation aware
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, _ = d.log.Write(fmt.Appendf([]byte(""), "[%s] IN: %s %s %s\n",
		time.Now().Format(time.DateTime),
		r.Method,
		r.RequestURI,
		r.Proto))
	// parse response while logging it
	tee := io.TeeReader(r.Body, d.log)
	if env, err := messages.ParseEnvelopeXML(tee); err != nil {
		if errors.Is(err, io.EOF) {
			d.in <- nil
		} else {
			tui.LogError("Malformed request received: %s", err.Error())
			http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusBadRequest)
			return
		}
	} else {
		_, _ = d.log.Write([]byte("\n"))
		if len(env.Body.Messages) > 1 {
			tui.LogNormal("%d messages received; ignoring all but first", len(env.Body.Messages))
		} else if len(env.Body.Messages) == 0 {
			tui.LogError("No message received")
			http.Error(w, "Bad request; no message received", http.StatusBadRequest)
			return
		}
		d.in <- env.Body.Messages[0]
		d.SetSessionID(env.GetID())
	}

	_, _ = d.log.Write(fmt.Appendf([]byte(""), "[%s] OUT:\n", time.Now().Format(time.DateTime)))
	resp := <-d.out
	d.writeHTTPResponse(w, http.StatusOK, resp)
	_, _ = d.log.Write([]byte("\n--------------------------------------------------------------------------------\n"))
}

// write a reply to the response
func (d *CWMPDevice) writeHTTPResponse(w http.ResponseWriter, statusCode int, resp *messages.Envelope) {
	w.WriteHeader(statusCode)
	if resp != nil {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		tee := io.MultiWriter(w, d.log)
		enc := xml.NewEncoder(tee)
		enc.Indent("", "\t")
		if err := enc.Encode(resp); err != nil {
			panic(err)
		}
	}
}

func (d *CWMPDevice) respondToRPC(r messages.Message) *messages.Envelope {
	var env messages.Envelope
	if rpc, ok := r.(messages.ACSMethod); ok {
		resp := rpc.GenerateResponse()
		env = d.newEnvelope(resp)
	} else {
		env = d.newEnvelope(messages.NewFault(8000, "Method not supported"))
	}
	return &env
}

func (d *CWMPDevice) createRPCFromCmd(cmd *configuration.SequenceCmd) (messages.SyncRPC, error) {
	rpcName := cmd.Cmd.String()
	switch rpcName {
	case messages.ChangeDUState{}.GetName():
		var m messages.ChangeDUState
		return m, cmd.Decode(&m)
	case messages.GetParameterNames{}.GetName():
		var m messages.GetParameterNames
		return m, cmd.Decode(&m)
	case messages.GetParameterValues{}.GetName():
		var m messages.GetParameterValues
		return m, cmd.Decode(&m)
	case messages.GetRPCMethods{}.GetName():
		var m messages.GetRPCMethods
		return m, cmd.Decode(&m)
	case messages.SetParameterValues{}.GetName():
		var m messages.SetParameterValues
		return m, cmd.Decode(&m)
	default:
		return nil, fmt.Errorf("unknown RPC '%s'", rpcName)
	}
}

func (d *CWMPDevice) expectRPC(ctx context.Context, matcher func(messages.Message) bool) (messages.Message, error) {
	// loop until message arrives or context expires
	for {
		rpc, err := d.pullMessage(ctx)
		if err != nil {
			return nil, err
		}
		if matcher(rpc) {
			return rpc, nil
		} else {
			env := d.respondToRPC(rpc)
			if err := d.pushEnvelope(ctx, env); err != nil {
				return nil, err
			}
		}
	}
}

func (d *CWMPDevice) newEnvelope(msg ...messages.Message) messages.Envelope {
	env := messages.Envelope{}
	env.Header = &messages.EnvelopeHeader{
		ID: messages.IDStruct{MustUnderstand: "1", Value: d.currentID},
	}
	env.Body = messages.EnvelopeBody{Messages: msg}
	return env
}

func (d *CWMPDevice) pullMessage(ctx context.Context) (messages.Message, error) {
	select {
	case rpc := <-d.in:
		return rpc, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout while waiting for incoming message")
	}
}

func (d *CWMPDevice) pushEnvelope(ctx context.Context, env *messages.Envelope) error {
	select {
	case d.out <- env:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout while sending message")
	}
}

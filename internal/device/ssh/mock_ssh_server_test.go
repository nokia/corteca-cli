// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package ssh_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// cmdHandlerFunc is called by the mock server for every exec request it receives.
// It returns the stdout to send back and the exit code to report.
type cmdHandlerFunc func(cmd string) (stdout string, exitCode uint32)

// withQuaggaProbe wraps a cmdHandlerFunc so that the "ps | grep ash" probe fired
// unconditionally by NewSSHDevice is handled transparently (returns "ash", exit 0,
// which signals to the device that Quagga is not active and no further action is
// required). All other commands are forwarded to inner.
func withQuaggaProbe(inner cmdHandlerFunc) cmdHandlerFunc {
	return func(cmd string) (string, uint32) {
		if strings.TrimSpace(cmd) == "ps | grep ash" {
			return "ash", 0
		}
		return inner(cmd)
	}
}

// startTestServer starts an in-process SSH server on a random loopback port and
// returns its address. The server accepts:
//   - password authentication when password is non-empty
//   - public-key authentication when authorizedKey is non-nil
//
// handler is called for every exec request the server receives.
// The server and its goroutines are torn down via t.Cleanup.
func startTestServer(t *testing.T, expectedUsername, password string, authorizedKey ssh.PublicKey, handler cmdHandlerFunc) string {
	t.Helper()

	// Generate a fresh host key for every server instance.
	hostKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("startTestServer: generate host key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatalf("startTestServer: create signer: %v", err)
	}

	cfg := &ssh.ServerConfig{}
	cfg.AddHostKey(signer)

	if password != "" {
		cfg.PasswordCallback = func(meta ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if expectedUsername != "" && meta.User() != expectedUsername {
				return nil, fmt.Errorf("wrong username: got %q, want %q", meta.User(), expectedUsername)
			}
			if string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("wrong password")
		}
	}

	if authorizedKey != nil {
		cfg.PublicKeyCallback = func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if expectedUsername != "" && meta.User() != expectedUsername {
				return nil, fmt.Errorf("wrong username: got %q, want %q", meta.User(), expectedUsername)
			}
			if bytes.Equal(key.Marshal(), authorizedKey.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("unauthorized key")
		}
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startTestServer: listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed — server done
			}
			go serveConn(conn, cfg, handler)
		}
	}()

	return ln.Addr().String()
}

func serveConn(conn net.Conn, cfg *ssh.ServerConfig, handler cmdHandlerFunc) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
	if err != nil {
		return // auth failure or protocol error — nothing to do
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}
		ch, requests, err := newChan.Accept()
		if err != nil {
			return
		}
		go serveSession(ch, requests, handler)
	}
}

func serveSession(ch ssh.Channel, reqs <-chan *ssh.Request, handler cmdHandlerFunc) {
	// defer ch.Close() is the key: when serveSession returns after handling the
	// exec request, the channel is closed. This causes the client-side
	// s.wait goroutine's "for msg := range reqs" loop to exit, which populates
	// s.exitStatus and unblocks session.Wait() — preventing a deadlock.
	defer ch.Close()

	for req := range reqs {
		switch req.Type {
		case "exec":
			if len(req.Payload) < 4 {
				req.Reply(false, nil)
				continue
			}
			n := binary.BigEndian.Uint32(req.Payload[:4])
			if uint32(len(req.Payload)) < 4+n {
				req.Reply(false, nil)
				continue
			}
			cmd := string(req.Payload[4 : 4+n])
			req.Reply(true, nil)

			// Call the handler synchronously. For normal commands this returns
			// quickly. For the context-cancellation test the handler blocks on
			// a channel until t.Cleanup fires — that is fine because by the
			// time the blocking handler is running, executeCommandString has
			// already returned ctx.Err() to the test goroutine.
			stdout, exitCode := handler(cmd)
			if stdout != "" {
				ch.Write([]byte(stdout)) //nolint:errcheck
			}
			exitStatus := make([]byte, 4)
			binary.BigEndian.PutUint32(exitStatus, exitCode)
			ch.SendRequest("exit-status", false, exitStatus) //nolint:errcheck

			// Return so that defer ch.Close() fires immediately, closing the
			// SSH channel and unblocking the client's session.Wait().
			return

		default:
			// Covers signal requests (e.g. SIGKILL on context cancel) and any
			// other channel requests the client may send while the handler is
			// running or between commands.
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

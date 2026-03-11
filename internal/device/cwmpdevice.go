package device

import (
	"bufio"
	"context"
	"corteca/internal/configuration"
	"corteca/internal/cwmp/messages"
	"corteca/internal/cwmp/models"
	"corteca/internal/dispatcher"
	"corteca/internal/tui"

	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	digestAuthClient "github.com/xinsnake/go-http-digest-auth-client"
)

const (
	defaultCWMPPort = 7547
)

type CWMPDevice struct {
	endpoint       configuration.Endpoint
	protocol       string
	server         *http.Server
	resultChan     chan *models.ResultsMessage
	taskChannel    chan messages.Message
	readWriter     *bufio.ReadWriter
	connection     net.Conn
	Log            *Logger
	lastCommandKey string
}

func NewCWMPDevice(endpoint configuration.Endpoint, logfile string) (*CWMPDevice, error) {
	device, _ := newCWMP(endpoint, logfile)

	device.protocol = "http"

	return device, nil
}

func NewCWMPsDevice(endpoint configuration.Endpoint, logfile string) (*CWMPDevice, error) {
	device, _ := newCWMP(endpoint, logfile)

	device.protocol = "https"

	return device, nil
}

func newCWMP(endpoint configuration.Endpoint, logfile string) (*CWMPDevice, error) {
	logger := &Logger{}
	logger.SetLogFile(logfile)

	return &CWMPDevice{
		endpoint:       endpoint,
		resultChan:     make(chan *models.ResultsMessage),
		taskChannel:    make(chan messages.Message),
		Log:            logger,
		lastCommandKey: "",
	}, nil

}

func (c *CWMPDevice) initServer(address string) error {
	c.server = &http.Server{
		Addr:    address,
		Handler: nil, // uses default mux
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	http.HandleFunc("/", c.handleTr069)
	tui.DisplaySuccessMsg(fmt.Sprintf("Starting CWMP server on address %v...", address))
	// Run server in a goroutine
	go func() {
		if err := c.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			tui.DisplayErrorMsg(fmt.Sprintf("Failed to start server: %s", err))
			os.Exit(1)
		}
	}()

	return nil
}

func (d *CWMPDevice) GetProtocol() int {
	return ConnectionCWMP
}

func (d *CWMPDevice) DiscoverTargetCPUArch(dispatcher dispatcher.Dispatcher) (string, error) {
	if configuration.GetCmdContext().Device.DeployDevice.DeviceArch == "" {
		return "", fmt.Errorf("error discovering device architecture")
	}
	return configuration.GetCmdContext().Device.DeployDevice.DeviceArch, nil
}

func (c *CWMPDevice) Connect() (dispatcher.Dispatcher, error) {
	var address string
	u, err := url.Parse(c.endpoint.CwmpServerAddr)
	if err != nil {
		return nil, err
	}

	if u.Port() == "" {
		address = u.Host + ":" + strconv.Itoa(defaultCWMPPort)
	} else {
		address = u.Host
	}

	if err = c.initServer(address); err != nil {
		return nil, err
	}

	connectionReqURL, err := url.Parse(c.endpoint.Addr.String())
	if err != nil {
		return nil, err
	}
	err = c.checkConnReqValues(connectionReqURL)
	if err != nil {
		tui.DisplayErrorMsg(fmt.Sprintf("skipping connection request to CPE device: %s", err))
	} else {
		err := c.SendConnectionRequest()
		if err != nil {
			tui.DisplayErrorMsg(err.Error())
		}
	}

	tui.DisplaySuccessMsg("Waiting for CPE to establish connection...")

	return dispatcher.NewCWMPDispatcher(c.taskChannel, c.resultChan), nil
}

func (c *CWMPDevice) Close() {
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil {
		tui.DisplayErrorMsg(fmt.Sprintf("Server Shutdown Failed: %s", err))
		return
	}

	tui.DisplaySuccessMsg("Server stopped gracefully!")
}

func (c *CWMPDevice) checkConnReqValues(u *url.URL) error {

	if u.Hostname() == "" {
		return fmt.Errorf("connection request URL: empty hostname")
	} else {
		if u.Port() == "" {
			return fmt.Errorf("connection request URL: empty port")
		}

		c.endpoint.Addr.RawTemplate = c.protocol + "://" + u.Host
	}

	user := u.User.Username()
	if user == "" && configuration.GetCmdContext().Device.Username.String() == "" {
		return fmt.Errorf("connection request username: is empty")
	} else if user != "" {
		c.endpoint.Username.RawTemplate = user
	} else {
		c.endpoint.Username.RawTemplate = configuration.GetCmdContext().Device.Username.String()
	}

	pass, ok := u.User.Password()
	if !ok && configuration.GetCmdContext().Device.Password.String() == "" {
		return fmt.Errorf("connection request password: is empty")
	} else if ok {
		c.endpoint.Password.RawTemplate = pass
	} else {
		c.endpoint.Password.RawTemplate = configuration.GetCmdContext().Device.Password.String()
	}

	return nil
}

func (c *CWMPDevice) SendConnectionRequest() error {
	dr := digestAuthClient.NewRequest(
		c.endpoint.Username.String(),
		c.endpoint.Password.String(),
		"GET",
		c.endpoint.Addr.String(),
		"",
	)
	resp, err := dr.Execute()
	if err != nil {
		return fmt.Errorf("error sending Connection Request: %v", err)
	}
	defer resp.Body.Close()

	tui.LogNormal("Connection Request sent. Status code: %d", resp.StatusCode)
	return nil
}

func (c *CWMPDevice) handleTr069(w http.ResponseWriter, r *http.Request) {
	var err error
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	c.connection, c.readWriter, err = hj.Hijack()
	if err != nil {
		tui.LogError("Hijacking error: %v", err)
		return
	}

	defer c.connection.Close()

	var requestBody []byte
	if r.Method == "POST" {
		// receive posted data
		requestBody, err = io.ReadAll(r.Body)
		if err != nil {
			tui.LogError("tr069 read body error")
			return
		}
	}

	var msg messages.Message
	msg, err = messages.ParseXML(requestBody)
	if err != nil {
		tui.DisplayErrorMsg(err.Error())
	}

	msgResponse, _ := c.createResponseMessage(msg)
	c.sendReply(msgResponse)
	tui.DisplaySuccessMsg("Connection with CPE established")
	c.handleSessionFlow()
}

func ParseMessage(reader *bufio.Reader, contentLength int) (messages.Message, error) {
	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, fmt.Errorf("reading HTTP body error: %v", err)
	}

	return messages.ParseXML(body)
}

func createOutgoingFaultMsg(reply messages.Message) (xml []byte) {
	switch reply.GetName() {
	case "InformResponse":
		fault := messages.NewFault()
		fault.MsgFaultCode = "8002"
		fault.MsgFaultString = "error creating inform response"
		fault.ID = reply.GetID()
		fault.CwmpFaultCode = "Server"
		fault.CwmpFaultString = "CWMP fault"
		xml, _ = fault.CreateXML()
	}
	return xml
}

// Generates a response and a result type message according to the incoming message
// If response message is nil, then no response shall be send to CPE
func (c *CWMPDevice) createResponseMessage(msg messages.Message) (response messages.Message, result *models.ResultsMessage) {
	switch msg.GetName() {
	case "Inform":
		inform := msg.(*messages.Inform)
		configuration.GetCmdContext().Device.Addr = configuration.TemplateField{RawTemplate: inform.Params["Device.ManagementServer.ConnectionRequestURL"]}
		informResponse := new(messages.InformResponse)
		informResponse.ID = inform.ID
		informResponse.MaxEnvelopes = 1

		response = informResponse
	case "GetParameterNamesResponse":
		result = models.NewResulMessage()
		result.Code = 0
		result.Message = msg.(*messages.GetParameterNamesResponse)

		response = nil
	case "GetParameterValuesResponse":
		result = models.NewResulMessage()
		result.Code = 0
		result.Message = msg.(*messages.GetParameterValuesResponse)

		response = nil
	case "SetParameterValuesResponse":
		status := msg.(*messages.SetParameterValuesResponse).Status
		result = models.NewResulMessage()
		result.Code = status
		result.Message = msg.(*messages.SetParameterValuesResponse)

		response = nil
	case "ChangeDUStateResponse":
		response = nil
	case "DUStateChangeComplete":
		ducomplete := msg.(*messages.DUStateChangeComplete)
		//if complete is from a previous task do not send results. wait for task
		if ducomplete.CommandKey == c.lastCommandKey {
			result = models.NewResulMessage()
			result.Code = ducomplete.Fault.FaultCode
			result.Message = ducomplete
			c.lastCommandKey = ""
		} else {
			tui.LogError("CommandKey not matching. %s != %s", ducomplete.CommandKey, c.lastCommandKey)
		}

		completeResp := messages.NewDUStateCompleteResponse()
		completeResp.ID = ducomplete.ID
		response = completeResp
	case "Fault":
		faultMsg := msg.(*messages.Fault)
		result = models.NewResulMessage()
		result.Code, _ = strconv.Atoi(faultMsg.MsgFaultCode)
		result.Message = faultMsg
		response = nil
	default:
		fault := messages.NewFault()
		fault.CwmpFaultCode = "8002"
		fault.CwmpFaultString = "internal error"
		fault.ID = msg.GetID()
		fault.CwmpFaultCode = "Server"
		fault.CwmpFaultString = "CWMP fault"
		response = fault

		result = models.NewResulMessage()
		result.Code, _ = strconv.Atoi(fault.MsgFaultCode)
		result.Message = fault
	}

	return response, result
}

func (c *CWMPDevice) sendReply(msg messages.Message) {
	if msg == nil {
		c.sendEmptyResponse()
		return
	}
	response, err := msg.CreateXML()
	if err != nil {
		response = createOutgoingFaultMsg(msg)
	}
	fmt.Fprintf(c.readWriter, "HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\nContent-Length: %d\r\n\r\n%s", len(response), response)
	c.readWriter.Flush()
}

func (c *CWMPDevice) sendEmptyResponse() {
	fmt.Fprint(c.readWriter, "HTTP/1.1 204 No Content\r\nContent-Length: 0\r\n\r\n")
	c.readWriter.Flush()
}

func (c *CWMPDevice) handleSessionFlow() {
	reader := bufio.NewReader(c.connection)
	for {
		c.connection.SetReadDeadline(time.Now().Add(10 * time.Second))
		headers, err := readHeaders(reader)

		if err != nil {
			// if err == io.EOF {
			// 	log.Println("Connection closed by the CPE")
			// }
			if err != io.EOF {
				tui.LogError("Reading HTTP headers error: %v", err)
			}
			// Stop reading headers in case of error or connetion termination by the CPE
			break
		}

		contentLength := GetMessageLength(headers)

		var msg messages.Message
		if contentLength == 0 {
			if err = c.sendRPC(); err != nil {
				error_res := models.NewResulMessage()
				error_res.Code = -1
				c.resultChan <- error_res
			}
			continue
		} else {
			msg, err = ParseMessage(reader, contentLength)
			if err != nil {
				c.sendEmptyResponse()
				err_msg := messages.NewFault()
				err_msg.MsgFaultCode = "-1"
				err_msg.MsgFaultString = err.Error()
				c.resultChan <- &models.ResultsMessage{Code: -1, Message: err_msg}
				continue
			}
		}

		resp, res := c.createResponseMessage(msg)
		c.sendReply(resp)
		// if there are results to return to dispatcher
		// send to CPE empty message (no more to send)
		// and send results to dispatcher
		if res != nil {
			c.sendEmptyResponse()
			c.resultChan <- res
		}
	}
}

func readHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "POST") || strings.HasPrefix(line, "GET") {
			headers[":request"] = line
		} else {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				headers[strings.ToLower(key)] = value
			}
		}
	}
	return headers, nil
}

func GetMessageLength(headers map[string]string) int {
	if val, ok := headers["content-length"]; ok {
		var n int
		fmt.Sscanf(val, "%d", &n)
		return n
	}
	return 0
}

func (c *CWMPDevice) sendRPC() error {
	msg := <-c.taskChannel
	if msg.GetName() == "ChangeDUState" {
		c.lastCommandKey = msg.(*messages.ChangeDUState).CommandKey
	}

	rpcXML, err := msg.CreateXML()
	if err != nil {
		c.sendEmptyResponse()
		return err
	}

	fmt.Fprintf(c.readWriter, "HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\nContent-Length: %d\r\n\r\n%s", len(rpcXML), string(rpcXML))
	c.readWriter.Flush()
	c.sendEmptyResponse()
	return nil
}

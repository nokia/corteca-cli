//go:build exclude

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/cwmp/messages"
	"github.com/nokia/corteca-cli/internal/cwmp/models"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCWMPDevice(t *testing.T) {
	endpoint := configuration.Endpoint{
		Addr: configuration.TemplateField{RawTemplate: "cwmp://localhost:7547"},
	}
	device, err := NewCWMPDevice(endpoint, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	u, err := url.Parse(device.endpoint.Addr.String())
	if err != nil {
		t.Errorf("parse error: %s", err.Error())
	}

	if u.Scheme != "cwmp" {
		t.Errorf("Expected endpoint scheme to be 'cwmp', got %v", u.Scheme)
	}

	if u.Host != "localhost:7547" {
		t.Errorf("Expected endpoint host to be 'localhost', got %v", u.Host)
	}

	if u.Port() != "7547" {
		t.Errorf("Expected port 7547', got %v", u.Port())
	}

}

func TestGetCWMPProtocol(t *testing.T) {
	device := &CWMPDevice{}
	expected := ConnectionCWMP
	if device.GetProtocol() != expected {
		t.Errorf("Expected protocol %d, got %d", expected, device.GetProtocol())
	}
}

// You can mock initServer to avoid starting a real server
func TestInitServerFailure(t *testing.T) {
	device := &CWMPDevice{}
	err := device.initServer("10.184.48.57:-1") // Invalid address
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}

func resetConfig(u, p, a string) {
	configuration.GetCmdContext().Device.Username = configuration.TemplateField{RawTemplate: u}
	configuration.GetCmdContext().Device.Password = configuration.TemplateField{RawTemplate: p}
	configuration.GetCmdContext().Device.Addr = configuration.TemplateField{RawTemplate: a}
}

func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("bad url %q: %v", s, err)
	}
	return u
}

func TestCheckConnReqValues(t *testing.T) {
	tests := []struct {
		name         string
		rawURL       string
		cfgUser      string
		cfgPass      string
		cfgAddr      string
		expectErrSub string // substring to check in error, empty means expect no error
		wantAddr     string
		wantUser     string
		wantPass     string
	}{
		{
			name:         "Valid: URL has host:port and creds override config",
			rawURL:       "http://admin:secret@localhost:7547/path",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "",
			wantAddr:     "http://localhost:7547",
			wantUser:     "admin",
			wantPass:     "secret",
		},
		{
			name:         "Valid: URL has no creds, config supplies both",
			rawURL:       "http://localhost:7547",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "",
			wantAddr:     "http://localhost:7547",
			wantUser:     "cfgUser",
			wantPass:     "cfgPass",
		},
		{
			name:         "Valid: username only in URL, password from config",
			rawURL:       "http://admin@localhost:7547",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "",
			wantAddr:     "http://localhost:7547",
			wantUser:     "admin",
			wantPass:     "cfgPass",
		},
		{
			name:         "Valid: password only in URL, username from config",
			rawURL:       "http://:p4ss@localhost:7547",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass", // URL should win for password, but username must come from config
			expectErrSub: "",
			wantAddr:     "http://localhost:7547",
			wantUser:     "cfgUser",
			wantPass:     "p4ss",
		},
		{
			name:         "Error: empty hostname",
			rawURL:       "http://:7547",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "empty hostname",
		},
		{
			name:         "Error: empty port",
			rawURL:       "http://localhost", // no :port
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "empty port",
		},
		{
			name:         "Error: creds missing in both URL and config -> username error first",
			rawURL:       "http://localhost:7547",
			cfgUser:      "",
			cfgPass:      "",
			expectErrSub: "username: is empty",
		},
		{
			name:         "Valid: IPv6 host with port",
			rawURL:       "http://[2001:db8::1]:7547",
			cfgUser:      "cfgUser",
			cfgPass:      "cfgPass",
			expectErrSub: "",
			wantAddr:     "http://[2001:db8::1]:7547",
			wantUser:     "cfgUser",
			wantPass:     "cfgPass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfig(tt.cfgUser, tt.cfgPass, tt.cfgAddr)

			u := mustURL(t, tt.rawURL)

			var dev CWMPDevice
			dev.protocol = u.Scheme

			err := dev.checkConnReqValues(u)

			if tt.expectErrSub != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectErrSub) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectErrSub, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := dev.endpoint.Addr.RawTemplate; got != tt.wantAddr {
				t.Errorf("Addr.RawTemplate mismatch: got %q, want %q", got, tt.wantAddr)
			}
			if got := dev.endpoint.Username.RawTemplate; got != tt.wantUser {
				t.Errorf("Username.RawTemplate mismatch: got %q, want %q", got, tt.wantUser)
			}
			if got := dev.endpoint.Password.RawTemplate; got != tt.wantPass {
				t.Errorf("Password.RawTemplate mismatch: got %q, want %q", got, tt.wantPass)
			}
		})
	}
}

func TestCreateOutgoingFaultMsg(t *testing.T) {
	tests := []struct {
		name                string
		incomingMsg         messages.Message
		expectedXMLFaultMsg string
	}{
		{
			incomingMsg: &messages.InformResponse{
				ID:           "5",
				NoMore:       0,
				MaxEnvelopes: 1,
			},
			expectedXMLFaultMsg: `  <SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
      <SOAP-ENV:Header>
          <cwmp:ID SOAP-ENV:mustUnderstand="1">5</cwmp:ID>
      </SOAP-ENV:Header>
      <SOAP-ENV:Body>
          <SOAP-ENV:Fault>
              <faultcode>Server</faultcode>
              <faultstring>CWMP fault</faultstring>
              <detail>
                  <cwmp:Fault>
                      <FaultCode>8002</FaultCode>
                      <FaultString>error creating inform response</FaultString>
                      <SetParameterValuesFault>
                          <ParameterName></ParameterName>
                          <FaultCode></FaultCode>
                          <FaultString></FaultString>
                          <ParameterKey></ParameterKey>
                      </SetParameterValuesFault>
                  </cwmp:Fault>
              </detail>
          </SOAP-ENV:Fault>
      </SOAP-ENV:Body>
  </SOAP-ENV:Envelope>`,
		},
	}

	for _, testStruct := range tests {
		t.Run(testStruct.name, func(t *testing.T) {
			xml := createOutgoingFaultMsg(testStruct.incomingMsg)
			assert.Equal(t, testStruct.expectedXMLFaultMsg, string(xml), "XMLs should be the same")
		})
	}
}

func TestCreateResponseMessage(t *testing.T) {
	tests := []struct {
		name                string
		incomingMsg         messages.Message
		expectedResponseMsg messages.Message
		expectedResultMsg   *models.ResultsMessage
	}{
		{
			name: "Inform",
			incomingMsg: &messages.Inform{
				ID: "5",
				Params: map[string]string{
					"Device.ManagementServer.ConnectionRequestURL": "http://10.42.0.249:7547",
				},
				MaxEnvelopes: 1,
			},
			expectedResponseMsg: &messages.InformResponse{
				ID:           "5",
				MaxEnvelopes: 1,
			},
		},
		{
			name: "GetParameterNamesResponse",
			incomingMsg: &messages.GetParameterNamesResponse{
				ID:      "5",
				XMLName: "GetParameterNamesResponse",
				ParameterList: messages.ParameterList{
					Parameters: []messages.ParameterInfoStruct{
						{
							Name:     "Device.DeviceInfo.ModelName",
							Writable: false,
						},
					}},
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 0,
				Message: &messages.GetParameterNamesResponse{
					ID:      "5",
					XMLName: "GetParameterNamesResponse",
					ParameterList: messages.ParameterList{
						Parameters: []messages.ParameterInfoStruct{
							{
								Name:     "Device.DeviceInfo.ModelName",
								Writable: false,
							},
						}},
				},
			},
		},
		{
			name: "GetParameterValuesResponse",
			incomingMsg: &messages.GetParameterValuesResponse{
				ID: "5",
				ParameterList: []messages.ParameterValuesInfoStruct{
					{
						Name:  "Device.DeviceInfo.ModelName",
						Value: "Beacon 10",
					}},
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 0,
				Message: &messages.GetParameterValuesResponse{
					ID: "5",
					ParameterList: []messages.ParameterValuesInfoStruct{
						{
							Name:  "Device.DeviceInfo.ModelName",
							Value: "Beacon 10",
						}},
				},
			},
		},
		{
			name: "SetParameterValuesResponse Status 0",
			incomingMsg: &messages.SetParameterValuesResponse{
				ID:     "5",
				Status: 0,
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 0,
				Message: &messages.SetParameterValuesResponse{
					ID:     "5",
					Status: 0,
				},
			},
		},
		{
			name: "SetParameterValuesResponse Status 1",
			incomingMsg: &messages.SetParameterValuesResponse{
				ID:     "5",
				Status: 1,
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 1,
				Message: &messages.SetParameterValuesResponse{
					ID:     "5",
					Status: 1,
				},
			},
		},
		{
			name: "DUStateChangeComplete",
			incomingMsg: &messages.DUStateChangeComplete{
				ID:                   "5",
				Name:                 "DUStateChangeComplete",
				UUID:                 "5cb0dbaa-37ae-59a2-a46d-b96999b01991",
				DeploymentUnitRef:    "DeploymentUnitRef",
				Version:              "1.0",
				ExecutionUnitRefList: []string{"ExecutionUnitRefList"},
				StartTime:            "15:48:32",
				CompleteTime:         "16:48:32",
				Fault: messages.FaultStruct{
					FaultCode:   0,
					FaultString: "No fault",
				},
				CommandKey: "commandKey",
			},
			expectedResponseMsg: &messages.ChangeDUStateCompleteResponse{
				ID: "5",
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 0,
				Message: &messages.DUStateChangeComplete{
					ID:                   "5",
					Name:                 "DUStateChangeComplete",
					UUID:                 "5cb0dbaa-37ae-59a2-a46d-b96999b01991",
					DeploymentUnitRef:    "DeploymentUnitRef",
					Version:              "1.0",
					ExecutionUnitRefList: []string{"ExecutionUnitRefList"},
					Fault: messages.FaultStruct{
						FaultCode:   0,
						FaultString: "No fault",
					},
					StartTime:    "15:48:32",
					CompleteTime: "16:48:32",
					CommandKey:   "commandKey",
				},
			},
		},
		{
			name: "Fault",
			incomingMsg: &messages.Fault{
				ID:             "5",
				MsgFaultCode:   "9001",
				MsgFaultString: "Request Denied",
			},
			expectedResultMsg: &models.ResultsMessage{
				Code: 9001,
				Message: &messages.Fault{
					ID:             "5",
					MsgFaultCode:   "9001",
					MsgFaultString: "Request Denied",
				},
			},
		},
	}

	device := &CWMPDevice{lastCommandKey: "commandKey"}
	for _, testStruct := range tests {
		t.Run(testStruct.name, func(t *testing.T) {
			responseMsg, resultMsg := device.createResponseMessage(testStruct.incomingMsg)
			if responseMsg != nil {
				assert.Equal(t, testStruct.expectedResponseMsg, responseMsg, "Response messages should be the same")
			}
			if resultMsg != nil {
				assert.Equal(t, testStruct.expectedResultMsg.Code, resultMsg.Code, "Result codes should be the same")
				assert.Equal(t, testStruct.expectedResultMsg.Message, resultMsg.Message, "Result messages should be the same")
			}
		})
	}
}

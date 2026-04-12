package messages_test

import (
	"bytes"
	"corteca/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	GetRPCMethodsInputXML = `<cwmp:GetRPCMethods></cwmp:GetRPCMethods>`

	GetRPCMethodsResponseInputXML = `<cwmp:GetRPCMethodsResponse>
  <MethodList soap-enc:array="xsd:string[3]">
    <string>Inform</string>
    <string>GetRPCMethods</string>
    <string>DUStateChangeComplete</string>
  </MethodList>
</cwmp:GetRPCMethodsResponse>`

	GetRPCMethodsResponseInputYAML = `MethodList:
    - Inform
    - GetRPCMethods
    - DUStateChangeComplete
`
)

var GetRPCMethodsInputMsg = messages.GetRPCMethods{}

var GetRPCMethodsResponseInputMsg = messages.GetRPCMethods{}.GenerateResponse()

func TestGetRPCMethodsParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetRPCMethodsInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetRPCMethods{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}
}

func TestGetRPCMethodsSerializeToXML(t *testing.T) {
	msg := GetRPCMethodsInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetRPCMethodsInputXML, buf.String())
}

func TestGetRPCMethodsResponseParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetRPCMethodsResponseInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetRPCMethodsResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 3, len(msg.MethodList.Methods))
	assert.Equal(t, "Inform", msg.MethodList.Methods[0])
	assert.Equal(t, "GetRPCMethods", msg.MethodList.Methods[1])
	assert.Equal(t, "DUStateChangeComplete", msg.MethodList.Methods[2])
}

func TestGetRPCMethodsResponseSerializeToXML(t *testing.T) {
	msg := GetRPCMethodsResponseInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetRPCMethodsResponseInputXML, buf.String())
}

func TestGetRPCMethodsResponseParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(GetRPCMethodsResponseInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.GetRPCMethodsResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 3, len(msg.MethodList.Methods))
	assert.Equal(t, "Inform", msg.MethodList.Methods[0])
	assert.Equal(t, "GetRPCMethods", msg.MethodList.Methods[1])
	assert.Equal(t, "DUStateChangeComplete", msg.MethodList.Methods[2])
}

func TestGetRPCMethodsResponseSerializeToYAML(t *testing.T) {
	msg := GetRPCMethodsResponseInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetRPCMethodsResponseInputYAML, outbuf.String())
}

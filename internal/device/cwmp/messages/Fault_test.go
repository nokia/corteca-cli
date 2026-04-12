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
	FaultSimpleInputXML = `<soap-env:Fault>
  <faultcode>Client</faultcode>
  <faultstring>CWMP fault</faultstring>
  <detail>
    <cwmp:Fault>
      <FaultCode>9027</FaultCode>
      <FaultString>System Resources Exceeded</FaultString>
    </cwmp:Fault>
  </detail>
</soap-env:Fault>`

	FaultInputXML = `<soap-env:Fault>
  <faultcode>Client</faultcode>
  <faultstring>CWMP fault</faultstring>
  <detail>
    <cwmp:Fault>
      <FaultCode>9003</FaultCode>
      <FaultString>Invalid arguments</FaultString>
      <SetParameterValuesFault>
        <ParameterName>Device.SomeParam</ParameterName>
        <FaultCode>9008</FaultCode>
        <FaultString>Attempt to set a non-writable parameter</FaultString>
      </SetParameterValuesFault>
    </cwmp:Fault>
  </detail>
</soap-env:Fault>`

	FaultInputYAML = `faultcode: Client
faultstring: CWMP fault
Detail:
    FaultCode: 9003
    FaultString: Invalid arguments
    SetParameterValuesFault:
        ParameterName: Device.SomeParam
        FaultCode: "9008"
        FaultString: Attempt to set a non-writable parameter
`
)

var FaultInputMsg = messages.Fault{
	FaultCode:   "Client",
	FaultString: "CWMP fault",
	Detail: messages.CwmpFaultStruct{
		FaultStruct: messages.FaultStruct{
			FaultCode:   9003,
			FaultString: "Invalid arguments",
		},
		SetParameterValuesFault: &messages.SetParameterValuesFaultStruct{
			ParameterName: "Device.SomeParam",
			FaultCode:     "9008",
			FaultString:   "Attempt to set a non-writable parameter",
		},
	},
}

func TestFaultParseFromXMLNoSetParameterValuesFault(t *testing.T) {
	buf := bytes.NewBufferString(FaultSimpleInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.Fault{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Client", msg.FaultCode)
	assert.Equal(t, "CWMP fault", msg.FaultString)
	assert.Equal(t, uint(9027), msg.Detail.FaultCode)
	assert.Equal(t, "System Resources Exceeded", msg.Detail.FaultString)
	assert.Nil(t, msg.Detail.SetParameterValuesFault)
}

func TestFaultParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(FaultInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.Fault{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Client", msg.FaultCode)
	assert.Equal(t, "CWMP fault", msg.FaultString)

	assert.Equal(t, uint(9003), msg.Detail.FaultCode)
	assert.Equal(t, "Invalid arguments", msg.Detail.FaultString)

	assert.NotNil(t, msg.Detail.SetParameterValuesFault)
	assert.Equal(t, "Device.SomeParam", msg.Detail.SetParameterValuesFault.ParameterName)
	assert.Equal(t, "9008", msg.Detail.SetParameterValuesFault.FaultCode)
	assert.Equal(t, "Attempt to set a non-writable parameter", msg.Detail.SetParameterValuesFault.FaultString)
}

func TestFaultSerializeToXML(t *testing.T) {
	msg := FaultInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, FaultInputXML, buf.String())
}

func TestFaultParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(FaultInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.Fault{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Client", msg.FaultCode)
	assert.Equal(t, "CWMP fault", msg.FaultString)

	assert.Equal(t, uint(9003), msg.Detail.FaultCode)
	assert.Equal(t, "Invalid arguments", msg.Detail.FaultString)

	assert.NotNil(t, msg.Detail.SetParameterValuesFault)
	assert.Equal(t, "Device.SomeParam", msg.Detail.SetParameterValuesFault.ParameterName)
	assert.Equal(t, "9008", msg.Detail.SetParameterValuesFault.FaultCode)
	assert.Equal(t, "Attempt to set a non-writable parameter", msg.Detail.SetParameterValuesFault.FaultString)
}

func TestFaultSerializeToYAML(t *testing.T) {
	msg := FaultInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, FaultInputYAML, outbuf.String())
}

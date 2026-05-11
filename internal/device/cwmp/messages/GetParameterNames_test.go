package messages_test

import (
	"bytes"
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	GetParameterNamesInputXML = `<cwmp:GetParameterNames>
  <ParameterPath>Device.SoftwareModules.</ParameterPath>
  <NextLevel>true</NextLevel>
</cwmp:GetParameterNames>`

	GetParameterNamesInputYAML = `ParameterPath: Device.SoftwareModules.
NextLevel: true
`

	GetParameterNamesResponseInputXML = `<cwmp:GetParameterNamesResponse>
  <ParameterList>
    <ParameterInfoStruct>
      <Name>Device.SoftwareModules.</Name>
      <Writable>false</Writable>
    </ParameterInfoStruct>
    <ParameterInfoStruct>
      <Name>Device.SoftwareModules.ExecutionUnit.1.Version</Name>
      <Writable>false</Writable>
    </ParameterInfoStruct>
  </ParameterList>
</cwmp:GetParameterNamesResponse>`

	GetParameterNamesResponseInputYAML = `ParameterList:
    - Name: Device.SoftwareModules.
      Writable: false
    - Name: Device.SoftwareModules.ExecutionUnit.1.Version
      Writable: false
`
)

var GetParameterNamesInputMsg = messages.GetParameterNames{
	ParameterPath: configuration.T("Device.SoftwareModules."),
	NextLevel:     true,
}

var GetParameterNamesResponseInputMsg = messages.GetParameterNamesResponse{
	ParameterList: []messages.ParameterInfoStruct{
		{Name: "Device.SoftwareModules.", Writable: false},
		{Name: "Device.SoftwareModules.ExecutionUnit.1.Version", Writable: false},
	},
}

func TestGetParameterNamesParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterNamesInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetParameterNames{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Device.SoftwareModules.", msg.ParameterPath.String())
	assert.Equal(t, true, msg.NextLevel)
}

func TestGetParameterNamesSerializeToXML(t *testing.T) {
	msg := GetParameterNamesInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterNamesInputXML, buf.String())
}

func TestGetParameterNamesParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterNamesInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.GetParameterNames{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Device.SoftwareModules.", msg.ParameterPath.String())
	assert.Equal(t, true, msg.NextLevel)
}

func TestGetParameterNamesSerializeToYAML(t *testing.T) {
	msg := GetParameterNamesInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterNamesInputYAML, outbuf.String())
}

func TestGetParameterNamesResponseParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterNamesResponseInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetParameterNamesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList))
	assert.Equal(t, "Device.SoftwareModules.", msg.ParameterList[0].Name)
	assert.Equal(t, false, msg.ParameterList[0].Writable)
	assert.Equal(t, "Device.SoftwareModules.ExecutionUnit.1.Version", msg.ParameterList[1].Name)
	assert.Equal(t, false, msg.ParameterList[1].Writable)
}

func TestGetParameterNamesResponseSerializeToXML(t *testing.T) {
	msg := GetParameterNamesResponseInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterNamesResponseInputXML, buf.String())
}

func TestGetParameterNamesResponseParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterNamesResponseInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.GetParameterNamesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList))
	assert.Equal(t, "Device.SoftwareModules.", msg.ParameterList[0].Name)
	assert.Equal(t, false, msg.ParameterList[0].Writable)
	assert.Equal(t, "Device.SoftwareModules.ExecutionUnit.1.Version", msg.ParameterList[1].Name)
	assert.Equal(t, false, msg.ParameterList[1].Writable)
}

func TestGetParameterNamesResponseSerializeToYAML(t *testing.T) {
	msg := GetParameterNamesResponseInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterNamesResponseInputYAML, outbuf.String())
}

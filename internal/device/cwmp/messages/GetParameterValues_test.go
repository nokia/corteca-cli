// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

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
	GetParameterValuesInputXML = `<cwmp:GetParameterValues>
  <ParameterNames soap-enc:array="xsd:string[2]">
    <string>Device.DeviceInfo.SoftwareVersion</string>
    <string>Device.DeviceInfo.HardwareVersion</string>
  </ParameterNames>
</cwmp:GetParameterValues>`

	GetParameterValuesInputYAML = `ParameterNames:
    - Device.DeviceInfo.SoftwareVersion
    - Device.DeviceInfo.HardwareVersion
`

	GetParameterValuesResponseInputXML = `<cwmp:GetParameterValuesResponse>
  <ParameterList soap-enc:array="cwmp:ParameterValueStruct[2]">
    <ParameterValueStruct>
      <Name>Device.DeviceInfo.SoftwareVersion</Name>
      <Value xsi:type="xsd:string">1.0.0</Value>
    </ParameterValueStruct>
    <ParameterValueStruct>
      <Name>Device.DeviceInfo.HardwareVersion</Name>
      <Value xsi:type="xsd:string">RevA</Value>
    </ParameterValueStruct>
  </ParameterList>
</cwmp:GetParameterValuesResponse>`

	GetParameterValuesResponseInputYAML = `ParameterList:
    - Name: Device.DeviceInfo.SoftwareVersion
      Type: xsd:string
      Value: 1.0.0
    - Name: Device.DeviceInfo.HardwareVersion
      Type: xsd:string
      Value: RevA
`
)

var GetParameterValuesInputMsg = messages.GetParameterValues{
	ParameterNames: messages.ParameterNameListStruct{
		Params: []configuration.TemplateField{
			configuration.T("Device.DeviceInfo.SoftwareVersion"),
			configuration.T("Device.DeviceInfo.HardwareVersion"),
		},
	},
}

var GetParameterValuesResponseInputMsg = messages.GetParameterValuesResponse{
	ParameterList: messages.ParameterValueListStruct{
		Params: []messages.ParameterValueStruct{
			{
				Name:    configuration.T("Device.DeviceInfo.SoftwareVersion"),
				Content: messages.NodeStruct{Type: messages.XsdString, Value: configuration.T("1.0.0")},
			},
			{
				Name:    configuration.T("Device.DeviceInfo.HardwareVersion"),
				Content: messages.NodeStruct{Type: messages.XsdString, Value: configuration.T("RevA")},
			},
		},
	},
}

func TestGetParameterValuesParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterValuesInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetParameterValues{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterNames.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterNames.Params[0].String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterNames.Params[1].String())
}

func TestGetParameterValuesSerializeToXML(t *testing.T) {
	msg := GetParameterValuesInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterValuesInputXML, buf.String())
}

func TestGetParameterValuesParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterValuesInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.GetParameterValues{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterNames.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterNames.Params[0].String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterNames.Params[1].String())
}

func TestGetParameterValuesSerializeToYAML(t *testing.T) {
	msg := GetParameterValuesInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterValuesInputYAML, outbuf.String())
}

func TestGetParameterValuesResponseParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterValuesResponseInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.GetParameterValuesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "1.0.0", msg.ParameterList.Params[0].Content.Value.String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterList.Params[1].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[1].Content.Type)
	assert.Equal(t, "RevA", msg.ParameterList.Params[1].Content.Value.String())
}

func TestGetParameterValuesResponseSerializeToXML(t *testing.T) {
	msg := GetParameterValuesResponseInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterValuesResponseInputXML, buf.String())
}

func TestGetParameterValuesResponseParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(GetParameterValuesResponseInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.GetParameterValuesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "1.0.0", msg.ParameterList.Params[0].Content.Value.String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterList.Params[1].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[1].Content.Type)
	assert.Equal(t, "RevA", msg.ParameterList.Params[1].Content.Value.String())
}

func TestGetParameterValuesResponseSerializeToYAML(t *testing.T) {
	msg := GetParameterValuesResponseInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, GetParameterValuesResponseInputYAML, outbuf.String())
}

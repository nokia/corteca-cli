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
	SetParameterValuesInputXML = `<cwmp:SetParameterValues>
  <ParameterList soap-enc:array="cwmp:ParameterValueStruct[2]">
    <ParameterValueStruct>
      <Name>Device.DeviceInfo.SoftwareVersion</Name>
      <Value xsi:type="xsd:string">2.0.0</Value>
    </ParameterValueStruct>
    <ParameterValueStruct>
      <Name>Device.DeviceInfo.HardwareVersion</Name>
      <Value xsi:type="xsd:string">RevB</Value>
    </ParameterValueStruct>
  </ParameterList>
  <ParameterKey>mykey</ParameterKey>
</cwmp:SetParameterValues>`

	SetParameterValuesInputYAML = `ParameterList:
    - Name: Device.DeviceInfo.SoftwareVersion
      Type: xsd:string
      Value: 2.0.0
    - Name: Device.DeviceInfo.HardwareVersion
      Type: xsd:string
      Value: RevB
ParameterKey: mykey
`

	SetParameterValuesResponseInputXML = `<cwmp:SetParameterValuesResponse>
  <Status>0</Status>
</cwmp:SetParameterValuesResponse>`

	SetParameterValuesResponseInputYAML = `Status: 0
`
)

var SetParameterValuesInputMsg = messages.SetParameterValues{
	ParameterList: messages.ParameterValueListStruct{
		Params: []messages.ParameterValueStruct{
			{
				Name:    configuration.T("Device.DeviceInfo.SoftwareVersion"),
				Content: messages.NodeStruct{Type: messages.XsdString, Value: configuration.T("2.0.0")},
			},
			{
				Name:    configuration.T("Device.DeviceInfo.HardwareVersion"),
				Content: messages.NodeStruct{Type: messages.XsdString, Value: configuration.T("RevB")},
			},
		},
	},
	ParameterKey: "mykey",
}

var SetParameterValuesResponseInputMsg = messages.SetParameterValuesResponse{
	Status: 0,
}

func TestSetParameterValuesParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(SetParameterValuesInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.SetParameterValues{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "2.0.0", msg.ParameterList.Params[0].Content.Value.String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterList.Params[1].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[1].Content.Type)
	assert.Equal(t, "RevB", msg.ParameterList.Params[1].Content.Value.String())
	assert.Equal(t, "mykey", msg.ParameterKey)
}

func TestSetParameterValuesSerializeToXML(t *testing.T) {
	msg := SetParameterValuesInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, SetParameterValuesInputXML, buf.String())
}

func TestSetParameterValuesParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(SetParameterValuesInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.SetParameterValues{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, 2, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "2.0.0", msg.ParameterList.Params[0].Content.Value.String())
	assert.Equal(t, "Device.DeviceInfo.HardwareVersion", msg.ParameterList.Params[1].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[1].Content.Type)
	assert.Equal(t, "RevB", msg.ParameterList.Params[1].Content.Value.String())
	assert.Equal(t, "mykey", msg.ParameterKey)
}

func TestSetParameterValuesSerializeToYAML(t *testing.T) {
	msg := SetParameterValuesInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, SetParameterValuesInputYAML, outbuf.String())
}

func TestSetParameterValuesResponseParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(SetParameterValuesResponseInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.SetParameterValuesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, uint(0), msg.Status)
}

func TestSetParameterValuesResponseSerializeToXML(t *testing.T) {
	msg := SetParameterValuesResponseInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, SetParameterValuesResponseInputXML, buf.String())
}

func TestSetParameterValuesResponseParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(SetParameterValuesResponseInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.SetParameterValuesResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, uint(0), msg.Status)
}

func TestSetParameterValuesResponseSerializeToYAML(t *testing.T) {
	msg := SetParameterValuesResponseInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, SetParameterValuesResponseInputYAML, outbuf.String())
}

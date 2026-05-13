// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages_test

import (
	"bytes"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	BlankEnvelopeInputXML = ` <?xml version="1.0" encoding="UTF-8"?>
<soap-env:Envelope
    xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:cwmp="urn:dslforum-org:cwmp-1-2"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema">

  <soap-env:Body></soap-env:Body>
</soap-env:Envelope>`

	EnvelopeInputXML = ` <?xml version="1.0" encoding="UTF-8"?>
<soap-env:Envelope
    xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:cwmp="urn:dslforum-org:cwmp-1-2"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema">

  <soap-env:Header>
    <cwmp:ID soap-env:mustUnderstand="1">123456789</cwmp:ID>
  </soap-env:Header>

  <soap-env:Body>
    <cwmp:Inform>
      <!-- Device Identification -->
      <DeviceId>
        <Manufacturer>ExampleCorp</Manufacturer>
        <OUI>ABCDEF</OUI>
        <ProductClass>RouterModelX</ProductClass>
        <SerialNumber>1234567890</SerialNumber>
      </DeviceId>
      <!-- Event List -->
      <Event soap-enc:arrayType="cwmp:EventStruct[2]"
             xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/">
        <EventStruct>
          <EventCode>0 BOOTSTRAP</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
        <EventStruct>
          <EventCode>1 BOOT</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
      </Event>
      <!-- Max Envelopes -->
      <MaxEnvelopes>1</MaxEnvelopes>
      <!-- Current Time -->
      <CurrentTime>2026-03-29T12:00:00Z</CurrentTime>
      <!-- Retry Count -->
      <RetryCount>3</RetryCount>
      <!-- Parameter List -->
      <ParameterList soap-enc:arrayType="cwmp:ParameterValueStruct[4]"
                     xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">1.0.0</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.HardwareVersion</Name>
          <Value xsi:type="xsd:string">RevA</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.ManagementServer.ConnectionRequestURL</Name>
          <Value xsi:type="xsd:string">http://192.168.1.1:7547/</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress</Name>
          <Value xsi:type="xsd:string">203.0.113.45</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:Inform>
  </soap-env:Body>
</soap-env:Envelope>`

	InvalidEnvelopeInputXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
 <soap-env:Body>
  <cwmp:InvalidRPC></cwmp:InvalidRPC>
 </soap-env:Body>
</soap-env:Envelope>`

	EnvelopeOutputXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Header>
    <cwmp:ID soap-env:mustUnderstand="1">testEnvelope</cwmp:ID>
  </soap-env:Header>
  <soap-env:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer></Manufacturer>
        <OUI></OUI>
        <ProductClass></ProductClass>
        <SerialNumber></SerialNumber>
      </DeviceId>
      <Event soap-enc:array="cwmp:EventStruct[0]"></Event>
      <MaxEnvelopes>0</MaxEnvelopes>
      <CurrentTime></CurrentTime>
      <RetryCount>0</RetryCount>
      <ParameterList soap-enc:array="cwmp:ParameterValueStruct[0]"></ParameterList>
    </cwmp:Inform>
  </soap-env:Body>
</soap-env:Envelope>`

	BlankEnvelopeOutputXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body></soap-env:Body>
</soap-env:Envelope>`
)

func TestEnvelopeParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(EnvelopeInputXML)
	dec := xml.NewDecoder(buf)
	env := messages.Envelope{}
	if err := dec.Decode(&env); err != nil {
		t.Logf("Failed parsing xml input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "123456789", env.GetID())
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.Inform)
	assert.Equal(t, true, ok)
	assert.Equal(t, "ABCDEF", msg.DeviceId.OUI)
	assert.Equal(t, "RouterModelX", msg.DeviceId.ProductClass)
	assert.Equal(t, "1234567890", msg.DeviceId.SerialNumber)
	assert.Equal(t, uint(1), msg.MaxEnvelopes)
	assert.Equal(t, "2026-03-29T12:00:00Z", msg.CurrentTime)
	assert.Equal(t, uint(3), msg.RetryCount)
	assert.Equal(t, 2, len(msg.Event.Events))
	assert.Equal(t, messages.EventBoot, msg.Event.Events[1].EventCode)
	assert.Equal(t, 4, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress", msg.ParameterList.Params[3].Name.String())
	assert.Equal(t, "203.0.113.45", msg.ParameterList.Params[3].Content.Value.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[3].Content.Type)
}

func TestBlankEnvelopeParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(BlankEnvelopeInputXML)
	dec := xml.NewDecoder(buf)
	env := messages.Envelope{}
	if err := dec.Decode(&env); err != nil {
		t.Logf("Failed parsing xml input: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, "", env.GetID())
	assert.Empty(t, env.GetBody())
}

func TestInvalidRPCEnvelope(t *testing.T) {
	buf := bytes.NewBufferString(InvalidEnvelopeInputXML)
	dec := xml.NewDecoder(buf)
	env := messages.Envelope{}
	assert.NotNil(t, dec.Decode(&env))
}

func TestEnvelopeSerializeToXML(t *testing.T) {
	env := messages.NewEnvelope("testEnvelope", &messages.Inform{})
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(env); err != nil {
		t.Logf("Failed generating xml output: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, EnvelopeOutputXML, buf.String())
}

func TestBlankEnvelopeSerializeToXML(t *testing.T) {
	env := messages.NewEnvelope("", nil)
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(env); err != nil {
		t.Logf("Failed generating xml output: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, BlankEnvelopeOutputXML, buf.String())
}

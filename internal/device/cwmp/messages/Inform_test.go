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
	InformInputXML = `
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
  <RetryCount>5</RetryCount>
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
`

	InformOutputXML = `<cwmp:Inform>
  <DeviceId>
    <Manufacturer>Nokia</Manufacturer>
    <OUI>NOKIA</OUI>
    <ProductClass>Beacon 9</ProductClass>
    <SerialNumber>SN1234567890</SerialNumber>
  </DeviceId>
  <Event soap-enc:array="cwmp:EventStruct[2]">
    <EventStruct>
      <EventCode>1 BOOT</EventCode>
      <CommandKey>12345</CommandKey>
    </EventStruct>
    <EventStruct>
      <EventCode>6 CONNECTION REQUEST</EventCode>
      <CommandKey>67890</CommandKey>
    </EventStruct>
  </Event>
  <MaxEnvelopes>1</MaxEnvelopes>
  <CurrentTime>2026-03-29T23:45:00Z</CurrentTime>
  <RetryCount>2</RetryCount>
  <ParameterList soap-enc:array="cwmp:ParameterValueStruct[1]">
    <ParameterValueStruct>
      <Name>Device.SoftwareModules.ExecutionUnit.1.Version</Name>
      <Value xsi:type="xsd:string">1.0.0</Value>
    </ParameterValueStruct>
  </ParameterList>
</cwmp:Inform>`

	InformInputYAML = `DeviceId:
    Manufacturer: Nokia
    OUI: NOKIA
    ProductClass: Beacon 9
    SerialNumber: SN1234567890
Event:
    - EventCode: 1 BOOT
      CommandKey: "12345"
    - EventCode: 6 CONNECTION REQUEST
      CommandKey: "67890"
MaxEnvelopes: 1
CurrentTime: "2026-03-29T23:45:00Z"
RetryCount: 2
ParameterList:
    - Name: Device.SoftwareModules.ExecutionUnit.1.Version
      Type: xsd:string
      Value: 1.0.0
`

	InformResponseInputXML = `<cwmp:InformResponse>
  <MaxEnvelopes>1</MaxEnvelopes>
</cwmp:InformResponse>`
)

func TestInformParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(InformInputXML)
	enc := xml.NewDecoder(buf)
	msg := messages.Inform{}
	if err := enc.Decode(&msg); err != nil {
		t.Logf("Failed parsing xml input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "ExampleCorp", msg.DeviceId.Manufacturer)
	assert.Equal(t, "ABCDEF", msg.DeviceId.OUI)
	assert.Equal(t, "RouterModelX", msg.DeviceId.ProductClass)
	assert.Equal(t, "1234567890", msg.DeviceId.SerialNumber)
	assert.Equal(t, uint(1), msg.MaxEnvelopes)
	assert.Equal(t, "2026-03-29T12:00:00Z", msg.CurrentTime)
	assert.Equal(t, uint(5), msg.RetryCount)
	assert.Equal(t, 2, len(msg.Event.Events))
	assert.Equal(t, messages.EventBoot, msg.Event.Events[1].EventCode)
	assert.Equal(t, 4, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress", msg.ParameterList.Params[3].Name.String())
	assert.Equal(t, "203.0.113.45", msg.ParameterList.Params[3].Content.Value.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[3].Content.Type)
}

func TestInformSerializeToXML(t *testing.T) {
	msg := messages.Inform{
		DeviceId: messages.DeviceIDStruct{
			Manufacturer: "Nokia",
			OUI:          "NOKIA",
			ProductClass: "Beacon 9",
			SerialNumber: "SN1234567890",
		},
		Event: messages.EventList{
			Events: []messages.EventStruct{
				messages.EventStruct{EventCode: messages.EventBoot, CommandKey: "12345"},
				messages.EventStruct{EventCode: messages.EventConnectionRequest, CommandKey: "67890"},
			},
		},
		MaxEnvelopes: 1,
		RetryCount:   2,
		CurrentTime:  "2026-03-29T23:45:00Z",
		ParameterList: messages.ParameterValueListStruct{
			Params: []messages.ParameterValueStruct{
				messages.ParameterValueStruct{
					Name: configuration.T("Device.SoftwareModules.ExecutionUnit.1.Version"),
					Content: messages.NodeStruct{
						Type:  messages.XsdString,
						Value: configuration.T("1.0.0"),
					},
				},
			},
		},
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating xml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, InformOutputXML, buf.String())
}

func TestInformUmarshalFromYaml(t *testing.T) {
	buf := bytes.NewBufferString(InformInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.Inform{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing yaml input: %s", err.Error())
		t.FailNow()
	}
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(outbuf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating xml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, InformOutputXML, outbuf.String())
}

func TestInformMarshalToYaml(t *testing.T) {
	buf := bytes.NewBufferString(InformOutputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.Inform{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing xml input: %s", err.Error())
		t.FailNow()
	}
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating yaml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, InformInputYAML, outbuf.String())
}

func TestInformResponseParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(InformResponseInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.InformResponse{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, uint(1), msg.MaxEnvelopes)
}

func TestInformResponseSerializeToXML(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	msg := messages.InformResponse{MaxEnvelopes: 1}
	if err := enc.Encode(&msg); err != nil {
		t.Logf("Failed generating XML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, InformResponseInputXML, buf.String())
}

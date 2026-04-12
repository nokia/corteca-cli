package messages

import (
	"encoding/xml"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Inform struct {
	XMLName       xml.Name                 `xml:"Inform" yaml:"-"`
	DeviceId      DeviceIDStruct           `yaml:"DeviceId"`
	Event         EventList                `yaml:"Event"`
	MaxEnvelopes  uint                     `yaml:"MaxEnvelopes"`
	CurrentTime   string                   `yaml:"CurrentTime"`
	RetryCount    uint                     `yaml:"RetryCount"`
	ParameterList ParameterValueListStruct `yaml:"ParameterList"`
}

// custom marshaller to add prefix to name
func (i Inform) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "Inform")
	type Alias Inform
	return e.EncodeElement(Alias(i), start)
}

type DeviceIDStruct struct {
	Manufacturer string `yaml:"Manufacturer"`
	OUI          string `yaml:"OUI"`
	ProductClass string `yaml:"ProductClass"`
	SerialNumber string `yaml:"SerialNumber"`
}

// EventStruct event
type EventList struct {
	Events []EventStruct `xml:"EventStruct"`
}

// custom marshaller to add type attribute
func (el EventList) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: SoapArray},
		Value: fmt.Sprintf("cwmp:EventStruct[%d]", len(el.Events)),
	})
	type Alias EventList
	return enc.EncodeElement(Alias(el), start)
}

func (e EventList) MarshalYAML() (any, error) {
	return e.Events, nil
}

func (e *EventList) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&e.Events)
}

type EventStruct struct {
	EventCode  string `yaml:"EventCode"`
	CommandKey string `yaml:"CommandKey"`
}

func (msg Inform) GetName() string           { return "Inform" }
func (msg Inform) GenerateResponse() Message { return InformResponse{MaxEnvelopes: 1} }

type InformResponse struct {
	XMLName      xml.Name `xml:"InformResponse" yaml:"-" json:"-"`
	MaxEnvelopes uint     `yaml:"MaxEnvelopes"`
}

func (i InformResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "InformResponse")
	type Alias InformResponse
	return e.EncodeElement(Alias(i), start)
}

func (msg InformResponse) GetName() string { return "InformResponse" }

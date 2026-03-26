package messages

import (
	"encoding/xml"

	"github.com/beevik/etree"
)

const (
	//XsdString string type
	XsdString string = "xsd:string"
	//XsdUnsignedint uint type
	XsdUnsignedint string = "xsd:unsignedInt"
)

const (
	//SoapArray array type
	SoapArray string = "SOAP-ENC:Array"
)

const (
	//EventBootStrap first connection
	EventBootStrap string = "0 BOOTSTRAP"
	//EventBoot reset or power on
	EventBoot string = "1 BOOT"
	//EventPeriodic periodic inform
	EventPeriodic string = "2 PERIODIC"
	//EventScheduled scheduled infrorm
	EventScheduled string = "3 SCHEDULED"
	//EventValueChange value change event
	EventValueChange string = "4 VALUE CHANGE"
	//EventKicked acs notify cpe
	EventKicked string = "5 KICKED"
	//EventConnectionRequest cpe request connection
	EventConnectionRequest string = "6 CONNECTION REQUEST"
	//EventTransferComplete download complete
	EventTransferComplete string = "7 TRANSFER COMPLETE"
)

// Message tr069 msg interface
type Message interface {
	Parse(doc *etree.Document) error
	CreateXML() ([]byte, error)
	GetName() string
	GetID() string
}

// Envelope tr069 body
type Envelope struct {
	XMLName   xml.Name    `xml:"SOAP-ENV:Envelope"`
	XmlnsEnv  string      `xml:"xmlns:SOAP-ENV,attr"`
	XmlnsEnc  string      `xml:"xmlns:SOAP-ENC,attr"`
	XmlnsXsd  string      `xml:"xmlns:xsd,attr"`
	XmlnsXsi  string      `xml:"xmlns:xsi,attr"`
	XmlnsCwmp string      `xml:"xmlns:cwmp,attr"`
	Header    interface{} `xml:"SOAP-ENV:Header"`
	Body      interface{} `xml:"SOAP-ENV:Body"`
}

// HeaderStruct tr069 header
type HeaderStruct struct {
	ID     IDStruct    `xml:"cwmp:ID"`
	NoMore interface{} `xml:"cwmp:NoMoreRequests,omitempty"`
}

// IDStruct msg id
type IDStruct struct {
	Attr  string `xml:"SOAP-ENV:mustUnderstand,attr,omitempty"`
	Value string `xml:",chardata"`
}

// NodeStruct node
type NodeStruct struct {
	Type  interface{} `xml:"xsi:type,attr"`
	Value string      `xml:",chardata"`
}

// EventStruct event
type EventStruct struct {
	Type   string            `xml:"SOAP-ENC:arrayType,attr"`
	Events []EventNodeStruct `xml:"EventStruct"`
}

// EventNodeStruct event node
type EventNodeStruct struct {
	EventCode  NodeStruct `xml:"EventCode"`
	CommandKey string     `xml:"CommandKey"`
}

// ParameterListStruct param list
type ParameterListStruct struct {
	Type   string                 `xml:"SOAP-ENC:arrayType,attr"`
	Params []ParameterValueStruct `xml:"ParameterValueStruct"`
}

// ParameterValueStruct param value
type ParameterValueStruct struct {
	Name  NodeStruct `xml:"Name"`
	Value NodeStruct `xml:"Value"`
}

// ValueStruct value
type ValueStruct struct {
	Type  string
	Value string
}

// FaultStruct error
type FaultStruct struct {
	FaultCode   int
	FaultString string
}

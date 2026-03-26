package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

type ParameterValues struct {
	XMLName        string
	ID             string
	CommandKey     string
	PrintFormat    string
	ParameterNames []string
}

type BodyStruct struct {
	GetParameterValues GetParameterValues `xml:"cwmp:GetParameterValues"`
}

type GetParameterValues struct {
	ParameterNames ParameterNames `xml:"ParameterNames"`
}

type ParameterNames struct {
	XmlnsSoapEnc string   `xml:"xmlns:soap-enc,attr"`
	ArrayType    string   `xml:"soap-enc:arrayType,attr"`
	Strings      []string `xml:"string"`
}

// GetName get msg type
func (msg *ParameterValues) GetName() string {
	return "GetParameterValues"
}

// GetID get msg id
func (msg *ParameterValues) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

func (msg *ParameterValues) CreateXML() ([]byte, error) {
	envelope := Envelope{
		XmlnsEnv:  "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsEnc:  "http://schemas.xmlsoap.org/soap/encoding/",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsCwmp: "urn:dslforum-org:cwmp-1-0",
		Header:    HeaderStruct{ID: IDStruct{Attr: "1", Value: msg.ID}},
		Body: BodyStruct{
			GetParameterValues: GetParameterValues{
				ParameterNames: ParameterNames{
					XmlnsSoapEnc: "http://schemas.xmlsoap.org/soap/encoding/",
					ArrayType:    fmt.Sprintf("xsd:string[%v]", len(msg.ParameterNames)),
					Strings:      msg.ParameterNames,
				},
			},
		},
	}
	return xml.MarshalIndent(envelope, "", "  ")
}

func (msg *ParameterValues) Parse(doc *etree.Document) error {
	return nil
}

func NewGetParameterValues() *ParameterValues {
	paramValStruct := new(ParameterValues)
	paramValStruct.ID = paramValStruct.GetID()
	paramValStruct.XMLName = paramValStruct.GetName()
	return paramValStruct
}

package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

type SetParameterValues struct {
	ID            string
	ParameterList []ParameterVal
	ParameterKey  string
}

type Body struct {
	SetParameterValues SetParameterValuesStruct `xml:"cwmp:SetParameterValues"`
}

type SetParameterValuesStruct struct {
	SetParameterList SetParameterList `xml:"ParameterList"`
	ParameterKey     string           `xml:"parameterKey"`
}

type SetParameterList struct {
	XmlnsSoapEnc string         `xml:"xmlns:soap-enc,attr"`
	ArrayType    string         `xml:"soap-enc:arrayType,attr"`
	Parameters   []ParameterVal `xml:"ParameterValueStruct"`
}

type ParameterVal struct {
	Name  string `xml:"Name"`
	Value Values `xml:"Value" yaml:"Value"`
}

type Values struct {
	Type  string `xml:"xsi:type,attr"`
	Value string `xml:",chardata"`
}

// GetName get msg type
func (msg *SetParameterValues) GetName() string {
	return "SetParameterValues"
}

// GetID get msg id
func (msg *SetParameterValues) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

func (msg *SetParameterValues) CreateXML() ([]byte, error) {
	envelope := Envelope{
		XmlnsEnv:  "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsEnc:  "http://schemas.xmlsoap.org/soap/encoding/",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsCwmp: "urn:dslforum-org:cwmp-1-0",
		Header:    HeaderStruct{ID: IDStruct{Attr: "1", Value: msg.ID}},
		Body: Body{
			SetParameterValues: SetParameterValuesStruct{
				ParameterKey: msg.ParameterKey,
				SetParameterList: SetParameterList{
					XmlnsSoapEnc: "http://schemas.xmlsoap.org/soap/encoding/",
					ArrayType:    fmt.Sprintf("cwmp:ParameterValueStruct[%v]", len(msg.ParameterList)),
					Parameters:   msg.ParameterList,
				},
			},
		},
	}
	return xml.MarshalIndent(envelope, "", "  ")
}

func (msg *SetParameterValues) Parse(doc *etree.Document) error {
	return nil
}

func NewSetParameterValues() *SetParameterValues {
	paramValStruct := new(SetParameterValues)
	paramValStruct.ID = paramValStruct.GetID()
	return paramValStruct
}

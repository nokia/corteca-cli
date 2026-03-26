package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

// Fault error response
type Fault struct {
	ID                      string
	Name                    string
	NoMore                  int
	CwmpFaultCode           string
	CwmpFaultString         string
	MsgFaultCode            string
	MsgFaultString          string
	SetParameterValuesFault SetParameterValuesFaultStruct
}

type faultBodyStruct struct {
	Fault faultStruct `xml:"SOAP-ENV:Fault"`
}
type faultStruct struct {
	FaultCode   string            `xml:"faultcode"`
	FaultString string            `xml:"faultstring"`
	FaultDetail faultDetailStruct `xml:"detail"`
}

type faultDetailStruct struct {
	CwmpFault cwmpFaultStruct `xml:"cwmp:Fault"`
}

type cwmpFaultStruct struct {
	FaultCode               string
	FaultString             string
	SetParameterValuesFault SetParameterValuesFaultStruct
}

// SetParameterValuesFaultStruct setParameterValues Fault
type SetParameterValuesFaultStruct struct {
	ParameterName string
	FaultCode     string
	FaultString   string
	ParameterKey  string
}

// NewFault create Fault object
func NewFault() (m *Fault) {
	m = &Fault{}
	m.ID = m.GetID()
	m.Name = m.GetName()
	return m
}

// GetName get msg type
func (msg *Fault) GetName() string {
	return "Fault"
}

// GetID get msg id
func (msg *Fault) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *Fault) CreateXML() ([]byte, error) {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}
	setParamFault := SetParameterValuesFaultStruct{
		FaultCode:     msg.SetParameterValuesFault.FaultCode,
		FaultString:   msg.SetParameterValuesFault.FaultString,
		ParameterName: msg.SetParameterValuesFault.ParameterName,
		ParameterKey:  msg.SetParameterValuesFault.ParameterKey,
	}
	cwmpFault := cwmpFaultStruct{
		FaultCode:               msg.MsgFaultCode,
		FaultString:             msg.MsgFaultString,
		SetParameterValuesFault: setParamFault,
	}
	detail := faultDetailStruct{CwmpFault: cwmpFault}
	fault := faultStruct{
		FaultCode:   msg.CwmpFaultCode,
		FaultString: msg.CwmpFaultString,
		FaultDetail: detail,
	}
	env.Body = faultBodyStruct{fault}
	return xml.MarshalIndent(env, "  ", "    ")
}

// Parse decode from xml
func (msg *Fault) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
	faultNode := doc.FindElement("//Fault")
	msg.CwmpFaultCode = faultNode.SelectElement("faultcode").Text()
	msg.CwmpFaultString = faultNode.SelectElement("faultstring").Text()
	detailNode := faultNode.FindElement("//detail")
	detailFaultNode := detailNode.FindElement("cwmp:Fault")
	msg.MsgFaultCode = detailFaultNode.SelectElement("FaultCode").Text()
	msg.MsgFaultString = detailFaultNode.SelectElement("FaultString").Text()
	setParamFaultNode := detailFaultNode.FindElement("//SetParameterValuesFault")
	if setParamFaultNode != nil {
		msg.SetParameterValuesFault.FaultCode = setParamFaultNode.SelectElement("FaultCode").Text()
		msg.SetParameterValuesFault.FaultString = setParamFaultNode.SelectElement("FaultString").Text()
		msg.SetParameterValuesFault.ParameterName = setParamFaultNode.SelectElement("ParameterName").Text()
	}
	return nil
}

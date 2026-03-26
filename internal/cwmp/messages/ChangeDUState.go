package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

type ChangeDUState struct {
	ID            string
	Name          string
	OperationType string
	Operation     DeploymentUnitOperationStruct
	CommandKey    string
}

type ChangeDUStateBodyStruct struct {
	Body ChangeDUStateStruct `xml:"cwmp:ChangeDUState"`
}

type ChangeDUStateStruct struct {
	XMLName    xml.Name                        `xml:"cwmp:ChangeDUState"`
	XmlnsCwmp  string                          `xml:"xmlns:cwmp,attr"`
	CommandKey string                          `xml:"CommandKey"`
	Operations []DeploymentUnitOperationStruct `xml:"Operations"`
}

type DeploymentUnitOperationStruct struct {
	XmlnsXsi        string `xml:"xmlns:xsi,attr"`
	XmlnsXsiType    string `xml:"xsi:type,attr"`
	URL             string `xml:"URL"`
	UUID            string `xml:"UUID"`
	Username        string `xml:"Username,omitempty"`
	Password        string `xml:"Password,omitempty"`
	ExecutionEnvRef string `xml:"ExecutionEnvRef"`
	Version         string `xml:"Version"`
}

func NewChangeDUState() *ChangeDUState {
	changeDUState := new(ChangeDUState)
	changeDUState.ID = changeDUState.GetID()
	changeDUState.Name = changeDUState.GetName()
	return changeDUState
}

// GetName get msg type
func (msg *ChangeDUState) GetName() string {
	return "ChangeDUState"
}

func (msg *ChangeDUState) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *ChangeDUState) CreateXML() ([]byte, error) {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}
	var operation = DeploymentUnitOperationStruct{}
	switch msg.OperationType {
	case "install":
		operation = DeploymentUnitOperationStruct{
			URL:      msg.Operation.URL,
			UUID:     msg.Operation.UUID,
			Password: msg.Operation.Password,
			Username: msg.Operation.Username,
		}
	case "uninstall":
		operation = DeploymentUnitOperationStruct{
			UUID:    msg.Operation.UUID,
			Version: msg.Operation.Version,
		}
	case "update":
		operation = DeploymentUnitOperationStruct{
			URL:  msg.Operation.URL,
			UUID: msg.Operation.UUID,
		}
	default:
		return nil, fmt.Errorf("operation %s not supported", msg.OperationType)
	}

	operation.XmlnsXsiType = fmt.Sprintf("cwmp:%sOpStruct", msg.OperationType)
	operation.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	changeDUState := ChangeDUStateStruct{CommandKey: msg.CommandKey, Operations: []DeploymentUnitOperationStruct{operation}}
	changeDUState.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Body = ChangeDUStateBodyStruct{changeDUState}
	return xml.MarshalIndent(env, "  ", "    ")
}

func (msg *ChangeDUState) Parse(doc *etree.Document) error {
	return nil
}

package messages

import (
	"encoding/xml"
	"fmt"

	"strconv"
	"time"

	"github.com/beevik/etree"
)

type DUStateChangeComplete struct {
	ID                   string
	Name                 string
	UUID                 string
	DeploymentUnitRef    string
	Version              string
	ExecutionUnitRefList []string
	Fault                FaultStruct
	StartTime            string
	CompleteTime         string
	CommandKey           string
}

type ChangeDUStateCompleteHeaderStruct struct {
	ID     IDStruct    `xml:"cwmp:ID"`
	NoMore interface{} `xml:"cwmp:NoMoreRequests,omitempty"`
}

type ChangeDUStateCompleteStruct struct {
	XMLName    xml.Name                     `xml:"DUStateChangeComplete"`
	CommandKey string                       `xml:"CommandKey"`
	Results    []DeploymentUnitResultStruct `xml:"Results"`
}

type DeploymentUnitResultStruct struct {
	UUID                 string                `xml:"UUID"`
	DeploymentUnitRef    string                `xml:"DeploymentUnitRef"`
	Version              string                `xml:"Version"`
	ExecutionUnitRefList []string              `xml:"ExecutionUnitRefList>string"`
	OperationPerformed   string                `xml:"OperationPerformed"`
	StartTime            string                `xml:"StartTime"`
	CompleteTime         string                `xml:"CompleteTime"`
	Fault                DUCompleteFaultStruct `xml:"Fault"`
}

type DUCompleteFaultStruct struct {
	FaultCode   int    `xml:"FaultCode"`
	FaultString string `xml:"FaultString"`
}

// GetName get msg type
func (msg *DUStateChangeComplete) GetName() string {
	return msg.Name
}

// GetID get msg id
func (msg *DUStateChangeComplete) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

func (msg *DUStateChangeComplete) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
	msg.Name = "DUStateChangeComplete"

	elemUUID := doc.FindElement("//UUID")
	if elemUUID == nil {
		return fmt.Errorf("failed to parse uuid")
	}
	msg.UUID = doc.FindElement("//UUID").Text()

	elemDepRef := doc.FindElement("//DeploymentUnitRef")
	if elemDepRef == nil {
		return fmt.Errorf("failed to parse DeploymentUnitRef")
	}
	msg.DeploymentUnitRef = elemDepRef.Text()

	elemStartTime := doc.FindElement("//StartTime")
	if elemStartTime == nil {
		return fmt.Errorf("failed to parse StartTime")
	}
	msg.StartTime = elemStartTime.Text()

	elemCompleteTime := doc.FindElement("//CompleteTime")
	if elemCompleteTime == nil {
		return fmt.Errorf("failed to parse CompleteTime")
	}
	msg.CompleteTime = elemCompleteTime.Text()

	elemFaultCode := doc.FindElement("//FaultCode")
	if elemFaultCode == nil {
		return fmt.Errorf("failed to parse FaultCode")
	}
	msg.Fault.FaultCode, _ = strconv.Atoi(elemFaultCode.Text())

	elemFaultString := doc.FindElement("//FaultString")
	if elemFaultString == nil {
		return fmt.Errorf("failed to parse FaultString")
	}
	msg.Fault.FaultString = elemFaultString.Text()

	msg.CommandKey = doc.FindElement("//CommandKey").Text()
	return nil
}

func (msg *DUStateChangeComplete) CreateXML() ([]byte, error) {
	return nil, nil
}

func NewChangeDUStateComplete() *DUStateChangeComplete {
	return &DUStateChangeComplete{}
}

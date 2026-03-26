package messages

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beevik/etree"
)

type SetParameterValuesResponse struct {
	ID      string
	Status  int
}

// GetName get msg type
func (msg *SetParameterValuesResponse) GetName() string {
	return "SetParameterValuesResponse"
}

// GetID get msg id
func (msg *SetParameterValuesResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

func (msg *SetParameterValuesResponse) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
	msg.Status, _ = strconv.Atoi(doc.FindElement("//SetParameterValuesResponse/Status").Text())
	return nil
}

func (msg *SetParameterValuesResponse) CreateXML() ([]byte, error) {
	return nil, fmt.Errorf("createXML should not be called on a response message, it's present to satisfy the Message interface")
}

func NewSetParameterValuesResponse() *SetParameterValuesResponse {
	return &SetParameterValuesResponse{}
}

package messages

import (
	"fmt"
	"time"
	
	"github.com/beevik/etree"
)

type ChangeDUStateResponse struct {
    ID              string
	Name            string
}

func (msg *ChangeDUStateResponse) GetName() string {
    return "ChangeDUStateResponse"
}

func (msg *ChangeDUStateResponse) GetID() string {
    if len(msg.ID) < 1 {
        msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
    }
    return msg.ID
}

// CreateXML encode into xml
func (msg *ChangeDUStateResponse) CreateXML() ([]byte, error) {
    return nil, nil
}

func (msg *ChangeDUStateResponse) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
    return nil
}

func NewChangeDUStateResponse() *ChangeDUStateResponse {
    changeDUState := new(ChangeDUStateResponse)
    changeDUState.ID = changeDUState.GetID()
    changeDUState.Name = changeDUState.GetName()
    return changeDUState
}

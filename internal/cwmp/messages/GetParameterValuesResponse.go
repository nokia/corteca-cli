package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

type GetParameterValuesResponse struct {
	ID            string	`xml:"ID" json:"ID"`
	XMLName       string	`xml:"Name" json:"Name"`
	ParameterList []ParameterValuesInfoStruct	`xml:"ParameterList" json:"ParameterList"`
}

type ParameterValuesInfoStruct struct {
	Name  string `xml:"name" json:"name"`
	Value string `xml:"value" json:"value"`
}

// GetName get msg type
func (msg *GetParameterValuesResponse) GetName() string {
	return "GetParameterValuesResponse"
}

// GetID get msg id
func (msg *GetParameterValuesResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

func (msg *GetParameterValuesResponse) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
	msg.XMLName = "GetParameterValuesResponse"
	for _, param := range doc.FindElements("//ParameterList/ParameterValueStruct") {
		msg.ParameterList = append(msg.ParameterList, ParameterValuesInfoStruct{
			Name:  param.SelectElement("Name").Text(),
			Value: param.SelectElement("Value").Text(),
		})
	}
	return nil
}

func (msg *GetParameterValuesResponse) CreateXML() ([]byte, error) {
	xmlStr, err := xml.Marshal(msg)
	fmt.Println(string(xmlStr))
	return xmlStr, err
	//return nil, fmt.Errorf("createXML should not be called on a response message, it's present to satisfy the Message interface")
}

func NewGetParameterValuesResponse() *GetParameterValuesResponse {
	return &GetParameterValuesResponse{}
}

package messages

import (
	"fmt"

	"github.com/beevik/etree"
)

type ParameterInfoStruct struct {
	Name     string `xml:"Name"`
	Writable bool   `xml:"Writable"`
}

type ParameterList struct {
	Parameters []ParameterInfoStruct `xml:"ParameterInfoStruct"`
}

type GetParameterNamesResponse struct {
	XMLName       string        `xml:"cwmp:GetParameterNamesResponse"`
	ParameterList ParameterList `xml:"ParameterList"`
	ID            string        `xml:"ID,attr"`
}

func NewGetParameterNamesResponse() *GetParameterNamesResponse {
	return &GetParameterNamesResponse{}
}

func (resp *GetParameterNamesResponse) GetID() string {
	return resp.ID
}

func (resp *GetParameterNamesResponse) Parse(doc *etree.Document) error {
	body := doc.FindElement("//cwmp:GetParameterNamesResponse")
	if body == nil {
		return fmt.Errorf("GetParameterNamesResponse element not found")
	}

	if idElem := doc.FindElement("//cwmp:ID"); idElem != nil {
		resp.ID = idElem.Text()
	}

	parameterList := body.FindElement("ParameterList")
	if parameterList == nil {
		return fmt.Errorf("ParameterList element not found")
	}

	for _, p := range parameterList.SelectElements("ParameterInfoStruct") {
		name := p.FindElement("Name")
		writable := p.FindElement("Writable")

		if name != nil && writable != nil {
			resp.ParameterList.Parameters = append(resp.ParameterList.Parameters, ParameterInfoStruct{
				Name:     name.Text(),
				Writable: writable.Text() == "1" || writable.Text() == "true",
			})
		}
	}

	return nil
}

func (resp *GetParameterNamesResponse) GetName() string {
	return "GetParameterNamesResponse"
}

func (resp *GetParameterNamesResponse) CreateXML() ([]byte, error) {
	return nil, fmt.Errorf("should not be called on a response, it's present to satisfy the Message interface")
}

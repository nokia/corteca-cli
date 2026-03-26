package messages

import (
	"encoding/xml"

	"github.com/beevik/etree"
)


type changeDUStateCompleteResponseBody struct {
	Response changeDUStateCompleteResponseStruct `xml:"cwmp:ChangeDUStateCompleteResponse"`
}

type changeDUStateCompleteResponseStruct struct{}


type ChangeDUStateCompleteResponse struct {
	ID      string
	Name    string
}

func (msg *ChangeDUStateCompleteResponse) GetName() string {
	return "DUStateChangeCompleteResponse"
}


func (msg *ChangeDUStateCompleteResponse) GetID() string {
	return msg.ID
}

func (msg *ChangeDUStateCompleteResponse) CreateXML() ([]byte, error) {
	env := Envelope{
		XmlnsEnv:  "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsEnc:  "http://schemas.xmlsoap.org/soap/encoding/",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsCwmp: "urn:dslforum-org:cwmp-1-0",
		Header: HeaderStruct{
			ID: IDStruct{
				Attr:  "1",
				Value: msg.GetID(), // You should define GetID() on your type
			},
		},
		Body: changeDUStateCompleteResponseBody{
			Response: changeDUStateCompleteResponseStruct{},
		},
	}

	return xml.MarshalIndent(env, "  ", "    ")
}

func (msg *ChangeDUStateCompleteResponse) Parse(doc *etree.Document) error {
	return nil
}

func NewDUStateCompleteResponse() *ChangeDUStateCompleteResponse {
	return &ChangeDUStateCompleteResponse{}
}

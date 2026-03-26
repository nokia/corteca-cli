package messages

import (
	"fmt"

	"github.com/beevik/etree"
)

func ParseXML(data []byte) (msg Message, err error) {
	doc := etree.NewDocument()
	doc.ReadFromBytes(data)

	envelope := doc.SelectElement("Envelope")
	if envelope == nil {
		return nil, fmt.Errorf("Envelope not found")
	}

	var body *etree.Element
	for _, elem := range envelope.ChildElements() {
		if elem.Tag == "Body" || elem.Tag == "soap-env:Body" {
			body = elem
			break
		}
	}

	if body != nil {
		bodyContent := body.ChildElements()[0]
		name := bodyContent.Tag
		switch name {
		case "Inform":
			msg = NewInform()
			err = msg.Parse(doc)
		case "DUStateChangeComplete":
			msg = NewChangeDUStateComplete()
			err = msg.Parse(doc)
		case "ChangeDUStateResponse":
			msg = NewChangeDUStateResponse()
			err = msg.Parse(doc)
		case "GetParameterNamesResponse":
			msg = NewGetParameterNamesResponse()
			err = msg.Parse(doc)
			if err != nil {
				return nil, fmt.Errorf("GetParameterNamesResponse ParseXML generated error: %v", err)
			}
		case "GetParameterValuesResponse":
			msg = NewGetParameterValuesResponse()
			err = msg.Parse(doc)
		case "SetParameterValuesResponse":
			msg = NewSetParameterValuesResponse()
			err = msg.Parse(doc)
		case "Fault":
			msg = NewFault()
			err = msg.Parse(doc)
		case "GetRPCMethodsResponse":
			s, _ := doc.WriteToString()
			fmt.Println("msg: ", s)
			return nil, fmt.Errorf("message %s not supported", name)
		default:
			return nil, fmt.Errorf("unknown message %s", name)
		}

		return msg, err
	} else {
		return nil, fmt.Errorf("body element not found")
	}
}

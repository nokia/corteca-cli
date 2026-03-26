package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

// InformResponse infrom response
type InformResponse struct {
	ID           string
	Name         string
	NoMore       int
	MaxEnvelopes int
}

type informResponseBodyStruct struct {
	Body informResponseStruct `xml:"cwmp:InformResponse"`
}

type informResponseStruct struct {
	MaxEnvelopes int `xml:"MaxEnvelopes"`
}

// GetID get msg id
func (msg *InformResponse) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// GetName get msg type
func (msg *InformResponse) GetName() string {
	return "InformResponse"
}

// CreateXML encode into xml
func (msg *InformResponse) CreateXML() ([]byte, error) {
	env := Envelope{}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	id := IDStruct{Attr: "1", Value: msg.GetID()}
	env.Header = HeaderStruct{ID: id, NoMore: msg.NoMore}
	infromResp := informResponseStruct{MaxEnvelopes: msg.MaxEnvelopes}
	env.Body = informResponseBodyStruct{infromResp}
	//output, err := xml.Marshal(env)
	output, err := xml.MarshalIndent(env, "  ", "    ")
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Parse decode from xml
func (msg *InformResponse) Parse(doc *etree.Document) error {
	return fmt.Errorf("parse should not be called on an outgoing message, it's present to satisfy the Message interface")
}

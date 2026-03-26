package messages

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/beevik/etree"
)

type GetParameterNames struct {
	Name          string
	ID            string
	ParameterPath string
	NextLevel     bool
}

type GetParameterNamesBodyStruct struct {
	XMLName xml.Name             `xml:"soap-env:Body"`
	Body    GetParameterNamesRPC `xml:"cwmp:GetParameterNames"`
}

type GetParameterNamesRPC struct {
	ParameterPath string `xml:"ParameterPath"`
	NextLevel     bool   `xml:"NextLevel"`
}

func (g *GetParameterNames) GetName() string {
	return "GetParameterNames"
}

func (g *GetParameterNames) CreateXML() ([]byte, error) {
	rpc := GetParameterNamesRPC{
		ParameterPath: g.ParameterPath,
		NextLevel:     g.NextLevel,
	}

	envelope := Envelope{
		XmlnsEnv:  "http://schemas.xmlsoap.org/soap/envelope/",
		XmlnsEnc:  "http://schemas.xmlsoap.org/soap/encoding/",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsCwmp: "urn:dslforum-org:cwmp-1-0",
		Header:    HeaderStruct{ID: IDStruct{Attr: "1", Value: g.ID}},
		Body:      GetParameterNamesBodyStruct{Body: rpc},
	}

	return xml.MarshalIndent(envelope, "", "  ")
}

func (g *GetParameterNames) GetID() string {
	if len(g.ID) < 1 {
		g.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", g.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return g.ID
}

func (g *GetParameterNames) Parse(doc *etree.Document) error {
	return nil
}

func NewGetParameterNames() *GetParameterNames {
	getParamNames := new(GetParameterNames)
	getParamNames.ID = getParamNames.GetID()
	getParamNames.Name = getParamNames.GetName()
	return getParamNames
}

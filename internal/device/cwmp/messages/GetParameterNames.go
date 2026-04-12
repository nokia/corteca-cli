package messages

import (
	c "corteca/internal/configuration"
	"encoding/xml"
)

type GetParameterNames struct {
	XMLName       xml.Name        `xml:"GetParameterNames" yaml:"-"`
	ParameterPath c.TemplateField `yaml:"ParameterPath"`
	NextLevel     bool            `yaml:"NextLevel"`
}

func (msg GetParameterNames) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "GetParameterNames")
	type Alias GetParameterNames
	return enc.EncodeElement(Alias(msg), start)
}

func (msg GetParameterNames) GetName() string { return "GetParameterNames" }
func (msg GetParameterNames) ValidateResponse(resp Message) error {
	return ExpectMessage[GetParameterNamesResponse](resp)
}

type GetParameterNamesResponse struct {
	XMLName       xml.Name              `xml:"GetParameterNamesResponse" yaml:"-"`
	ParameterList []ParameterInfoStruct `xml:"ParameterList>ParameterInfoStruct" yaml:"ParameterList"`
}

func (msg GetParameterNamesResponse) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name)
	type Alias GetParameterNamesResponse
	return enc.EncodeElement(Alias(msg), start)
}

type ParameterInfoStruct struct {
	Name     string `yaml:"Name"`
	Writable bool   `yaml:"Writable"`
}

func (msg GetParameterNamesResponse) GetName() string { return "GetParameterNamesResponse" }

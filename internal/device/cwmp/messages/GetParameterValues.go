package messages

import (
	"encoding/xml"
)

type GetParameterValues struct {
	XMLName        xml.Name                `xml:"GetParameterValues" yaml:"-"`
	ParameterNames ParameterNameListStruct `yaml:"ParameterNames"`
}

func (msg GetParameterValues) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "GetParameterValues")
	type Alias GetParameterValues
	return enc.EncodeElement(Alias(msg), start)
}

func (msg GetParameterValues) GetName() string { return "GetParameterValues" }
func (msg GetParameterValues) ValidateResponse(resp Message) error {
	return ExpectMessage[GetParameterValuesResponse](resp)
}

type GetParameterValuesResponse struct {
	XMLName       xml.Name                 `xml:"GetParameterValuesResponse" yaml:"-"`
	ParameterList ParameterValueListStruct `yaml:"ParameterList"`
}

func (msg GetParameterValuesResponse) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name)
	type Alias GetParameterValuesResponse
	return enc.EncodeElement(Alias(msg), start)
}

func (msg GetParameterValuesResponse) GetName() string { return "GetParameterValuesResponse" }

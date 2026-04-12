package messages

import (
	"encoding/xml"
)

type SetParameterValues struct {
	XMLName       xml.Name                 `xml:"SetParameterValues" yaml:"-"`
	ParameterList ParameterValueListStruct `yaml:"ParameterList"`
	ParameterKey  string                   `yaml:"ParameterKey"`
}

func (msg SetParameterValues) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "SetParameterValues")
	type Alias SetParameterValues
	return enc.EncodeElement(Alias(msg), start)
}

func (msg SetParameterValues) GetName() string { return "SetParameterValues" }
func (msg SetParameterValues) ValidateResponse(resp Message) error {
	return ExpectMessage[SetParameterValuesResponse](resp)
}

type SetParameterValuesResponse struct {
	XMLName xml.Name `xml:"SetParameterValuesResponse" yaml:"-"`
	Status  uint     `yaml:"Status"`
}

func (msg SetParameterValuesResponse) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name)
	type Alias SetParameterValuesResponse
	return enc.EncodeElement(Alias(msg), start)
}

func (msg SetParameterValuesResponse) GetName() string { return "SetParameterValuesResponse" }

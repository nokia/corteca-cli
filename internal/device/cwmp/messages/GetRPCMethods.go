package messages

import (
	"encoding/xml"
	"fmt"

	"gopkg.in/yaml.v3"
)

type GetRPCMethods struct {
	XMLName xml.Name `xml:"GetRPCMethods" yaml:"-"`
}

func (msg GetRPCMethods) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "GetRPCMethods")
	type Alias GetRPCMethods
	return enc.EncodeElement(Alias(msg), start)
}

func (msg GetRPCMethods) GetName() string { return "GetRPCMethods" }
func (msg GetRPCMethods) ValidateResponse(resp Message) error {
	return ExpectMessage[GetRPCMethodsResponse](resp)
}
func (msg GetRPCMethods) GenerateResponse() Message {
	return GetRPCMethodsResponse{
		MethodList: MethodListStruct{
			Methods: []string{
				"Inform",
				"GetRPCMethods",
				"DUStateChangeComplete",
			},
		},
	}
}

type GetRPCMethodsResponse struct {
	XMLName    xml.Name         `xml:"GetRPCMethodsResponse" yaml:"-"`
	MethodList MethodListStruct `yaml:"MethodList"`
}

func (msg GetRPCMethodsResponse) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "GetRPCMethodsResponse")
	type Alias GetRPCMethodsResponse
	return enc.EncodeElement(Alias(msg), start)
}

type MethodListStruct struct {
	Methods []string `xml:"string"`
}

func (pl MethodListStruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, XmlAttr(SoapArray, fmt.Sprintf("xsd:string[%d]", len(pl.Methods))))
	type Alias MethodListStruct
	return enc.EncodeElement(Alias(pl), start)
}

func (pl MethodListStruct) MarshalYAML() (any, error) {
	return pl.Methods, nil
}

func (pl *MethodListStruct) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&pl.Methods)
}

func (msg GetRPCMethodsResponse) GetName() string { return "GetRPCMethodsResponse" }

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"encoding/xml"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	XsdString      string = "xsd:string"
	XsdUnsignedint string = "xsd:unsignedInt"
)

const (
	SoapArray string = "soap-enc:array"
	XsiType   string = "xsi:type"
)

const (
	EventBootStrap         string = "0 BOOTSTRAP"
	EventBoot              string = "1 BOOT"
	EventPeriodic          string = "2 PERIODIC"
	EventScheduled         string = "3 SCHEDULED"
	EventValueChange       string = "4 VALUE CHANGE"
	EventKicked            string = "5 KICKED"
	EventConnectionRequest string = "6 CONNECTION REQUEST"
	EventTransferComplete  string = "7 TRANSFER COMPLETE"
)

// helpers
func PrefixName(prefix string, name *xml.Name, elems ...string) {
	if len(elems) > 0 {
		name.Local = strings.Join(elems, ".")
	}
	name.Local = fmt.Sprintf("%s:%s", prefix, name.Local)
}

func PrefixSoapEnv(name *xml.Name, elems ...string) {
	PrefixName("soap-env", name, elems...)
}

func PrefixCwmp(name *xml.Name, elems ...string) {
	PrefixName("cwmp", name, elems...)
}

func XmlAttr(name, value string) xml.Attr {
	return xml.Attr{
		Name:  xml.Name{Local: name},
		Value: value,
	}
}

type Message interface {
	GetName() string
}

type SyncRPC interface {
	Message
	ValidateResponse(Message) error
}

type AsyncRPC interface {
	SyncRPC
	Match(m Message) bool
}

type ACSMethod interface {
	Message
	GenerateResponse() Message
}

func ExpectMessage[T Message](m Message) error {
	if _, ok := m.(T); !ok {
		return fmt.Errorf("unexpected %s received", m.GetName())
	}
	return nil
}

type NodeStruct struct {
	Type  string                      `xml:"type,attr,omitempty" yaml:"Type,omitempty"`
	Value configuration.TemplateField `xml:",chardata" yaml:"Value"`
}

func (ns NodeStruct) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		Type  string `xml:"xsi:type,attr,omitempty"`
		Value string `xml:",chardata"`
	}{Type: ns.Type, Value: ns.Value.String()}, start)
}

type ParameterNameListStruct struct {
	Params []configuration.TemplateField `xml:"string"`
}

func (pl ParameterNameListStruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, XmlAttr(SoapArray, fmt.Sprintf("xsd:string[%d]", len(pl.Params))))
	type Alias ParameterNameListStruct
	return enc.EncodeElement(Alias(pl), start)
}

func (pl ParameterNameListStruct) MarshalYAML() (any, error) {
	return pl.Params, nil
}

func (pl *ParameterNameListStruct) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&pl.Params)
}

type ParameterValueListStruct struct {
	Params []ParameterValueStruct `xml:"ParameterValueStruct"`
}

func (pl ParameterValueListStruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, XmlAttr(SoapArray, fmt.Sprintf("cwmp:ParameterValueStruct[%d]", len(pl.Params))))
	type Alias ParameterValueListStruct
	return enc.EncodeElement(Alias(pl), start)
}

func (pl ParameterValueListStruct) MarshalYAML() (any, error) {
	return pl.Params, nil
}

func (pl *ParameterValueListStruct) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&pl.Params)
}

type ParameterValueStruct struct {
	Name    configuration.TemplateField `xml:"Name" yaml:"Name"`
	Content NodeStruct                  `xml:"Value" yaml:",inline"`
}

type FaultStruct struct {
	FaultCode   uint   `yaml:"FaultCode"`
	FaultString string `yaml:"FaultString"`
}

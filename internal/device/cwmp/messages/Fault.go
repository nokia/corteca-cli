// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages

import (
	"encoding/xml"
)

func NewFault(code uint, msg string) Fault {
	return Fault{
		FaultCode:   "Server",
		FaultString: "CWMP Fault",
		Detail: CwmpFaultStruct{
			FaultStruct: FaultStruct{
				FaultCode:   code,
				FaultString: msg,
			},
		},
	}
}

type Fault struct {
	XMLName     xml.Name        `xml:"Fault" yaml:"-"`
	FaultCode   string          `xml:"faultcode" yaml:"faultcode"`
	FaultString string          `xml:"faultstring" yaml:"faultstring"`
	Detail      CwmpFaultStruct `xml:"detail>Fault" yaml:"Detail"`
}

func (f Fault) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixSoapEnv(&start.Name, "Fault")
	type Alias Fault
	return enc.EncodeElement(Alias(f), start)
}

type CwmpFaultStruct struct {
	XMLName                 xml.Name `xml:"Fault" yaml:"-"`
	FaultStruct             `yaml:",inline"`
	SetParameterValuesFault *SetParameterValuesFaultStruct `xml:"SetParameterValuesFault,omitempty" yaml:"SetParameterValuesFault,omitempty"`
}

func (cf CwmpFaultStruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name)
	type Alias CwmpFaultStruct
	return enc.EncodeElement(Alias(cf), start)
}

type SetParameterValuesFaultStruct struct {
	ParameterName string `yaml:"ParameterName"`
	FaultCode     string `yaml:"FaultCode"`
	FaultString   string `yaml:"FaultString"`
}

func (msg Fault) GetName() string { return "Fault" }

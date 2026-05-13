// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages

import (
	"encoding/xml"
)

type DUStateChangeComplete struct {
	XMLName    xml.Name         `xml:"DUStateChangeComplete" yaml:"-"`
	CommandKey string           `yaml:"CommandKey"`
	Results    []OpResultStruct `xml:"Results>OpResultStruct" yaml:"Results"`
}

func (msg DUStateChangeComplete) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "DUStateChangeComplete")
	type Alias DUStateChangeComplete
	return enc.EncodeElement(Alias(msg), start)
}

type OpResultStruct struct {
	XMLName              xml.Name    `xml:"OpResultStruct" yaml:"-"`
	UUID                 string      `yaml:"UUID"`
	DeploymentUnitRef    string      `yaml:"DeploymentUnitRef"`
	Version              string      `yaml:"Version"`
	CurrentState         string      `yaml:"CurrentState"`
	Resolved             bool        `yaml:"Resolved"`
	ExecutionUnitRefList string      `yaml:"ExecutionUnitRefList"`
	StartTime            string      `yaml:"StartTime"`
	CompleteTime         string      `yaml:"CompleteTime"`
	Fault                FaultStruct `yaml:"Fault"`
}

func (m DUStateChangeComplete) GetName() string { return "DUStateChangeComplete" }
func (m DUStateChangeComplete) ValidateResponse(msg Message) error {
	return ExpectMessage[DUStateChangeCompleteResponse](msg)
}
func (m DUStateChangeComplete) GenerateResponse() Message { return DUStateChangeCompleteResponse{} }

type DUStateChangeCompleteResponse struct {
	XMLName xml.Name `xml:"DUStateChangeCompleteResponse"`
}

func (msg DUStateChangeCompleteResponse) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "DUStateChangeCompleteResponse")
	type Alias DUStateChangeCompleteResponse
	return enc.EncodeElement(Alias(msg), start)
}

func (m DUStateChangeCompleteResponse) GetName() string { return "DUStateChangeCompleteResponse" }

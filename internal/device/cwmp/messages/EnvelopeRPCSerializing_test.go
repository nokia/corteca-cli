// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages_test

import (
	"bytes"
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

const envelopeHeader = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Header>
    <cwmp:ID soap-env:mustUnderstand="1">test-id</cwmp:ID>
  </soap-env:Header>
  <soap-env:Body>`

const envelopeFooter = `
  </soap-env:Body>
</soap-env:Envelope>`

const (
	ChangeDUStateEnvelopeOutputXML = envelopeHeader + `
    <cwmp:ChangeDUState>
      <CommandKey>testkey</CommandKey>
      <Operations>
        <InstallOpStruct>
          <URL>http://example.com/app.ipk</URL>
          <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
          <Username>user</Username>
          <Password>pass</Password>
          <ExecutionEnvRef>generic</ExecutionEnvRef>
        </InstallOpStruct>
      </Operations>
    </cwmp:ChangeDUState>` + envelopeFooter

	FaultEnvelopeOutputXML = envelopeHeader + `
    <soap-env:Fault>
      <faultcode>Server</faultcode>
      <faultstring>CWMP Fault</faultstring>
      <detail>
        <cwmp:Fault>
          <FaultCode>9027</FaultCode>
          <FaultString>System Resources Exceeded</FaultString>
        </cwmp:Fault>
      </detail>
    </soap-env:Fault>` + envelopeFooter

	GetParameterNamesEnvelopeOutputXML = envelopeHeader + `
    <cwmp:GetParameterNames>
      <ParameterPath>Device.SoftwareModules.</ParameterPath>
      <NextLevel>true</NextLevel>
    </cwmp:GetParameterNames>` + envelopeFooter

	GetParameterValuesEnvelopeOutputXML = envelopeHeader + `
    <cwmp:GetParameterValues>
      <ParameterNames soap-enc:array="xsd:string[2]">
        <string>Device.DeviceInfo.SoftwareVersion</string>
        <string>Device.DeviceInfo.HardwareVersion</string>
      </ParameterNames>
    </cwmp:GetParameterValues>` + envelopeFooter

	SetParameterValuesEnvelopeOutputXML = envelopeHeader + `
    <cwmp:SetParameterValues>
      <ParameterList soap-enc:array="cwmp:ParameterValueStruct[2]">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">2.0.0</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.HardwareVersion</Name>
          <Value xsi:type="xsd:string">RevB</Value>
        </ParameterValueStruct>
      </ParameterList>
      <ParameterKey>mykey</ParameterKey>
    </cwmp:SetParameterValues>` + envelopeFooter

	GetRPCMethodsEnvelopeOutputXML = envelopeHeader + `
    <cwmp:GetRPCMethods></cwmp:GetRPCMethods>` + envelopeFooter

	InformResponseAndChangeDUStateEnvelopeOutputXML = envelopeHeader + `
    <cwmp:InformResponse>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:InformResponse>
    <cwmp:ChangeDUState>
      <CommandKey>testkey</CommandKey>
      <Operations>
        <InstallOpStruct>
          <URL>http://example.com/app.ipk</URL>
          <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
          <Username>user</Username>
          <Password>pass</Password>
          <ExecutionEnvRef>generic</ExecutionEnvRef>
        </InstallOpStruct>
      </Operations>
    </cwmp:ChangeDUState>` + envelopeFooter
)

func serializeEnvelope(t *testing.T, env messages.Envelope) string {
	t.Helper()
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(env); err != nil {
		t.Logf("Failed serializing envelope: %s", err.Error())
		t.FailNow()
	}
	return buf.String()
}

func testChangeDUState() messages.ChangeDUState {
	return messages.ChangeDUState{
		CommandKey: configuration.T("testkey"),
		Operations: messages.DUOperationStruct{
			Op: []messages.DUOperation{
				messages.InstallOpStruct{
					URL:             configuration.T("http://example.com/app.ipk"),
					UUID:            configuration.T("c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d"),
					Username:        configuration.T("user"),
					Password:        configuration.T("pass"),
					ExecutionEnvRef: configuration.T("generic"),
				},
			},
		},
	}
}

func TestEnvelopeSerializeChangeDUState(t *testing.T) {
	msg := testChangeDUState()
	env := messages.NewEnvelope("test-id", &msg)
	assert.Equal(t, ChangeDUStateEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeFault(t *testing.T) {
	msg := messages.NewFault(9027, "System Resources Exceeded")
	env := messages.NewEnvelope("test-id", &msg)
	assert.Equal(t, FaultEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeGetParameterNames(t *testing.T) {
	msg := GetParameterNamesInputMsg
	env := messages.NewEnvelope("test-id", &msg)
	assert.Equal(t, GetParameterNamesEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeGetParameterValues(t *testing.T) {
	msg := GetParameterValuesInputMsg
	env := messages.NewEnvelope("test-id", &msg)
	assert.Equal(t, GetParameterValuesEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeSetParameterValues(t *testing.T) {
	msg := SetParameterValuesInputMsg
	env := messages.NewEnvelope("test-id", &msg)
	assert.Equal(t, SetParameterValuesEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeGetRPCMethods(t *testing.T) {
	msg := GetRPCMethodsInputMsg
	env := messages.NewEnvelope("test-id", msg)
	assert.Equal(t, GetRPCMethodsEnvelopeOutputXML, serializeEnvelope(t, env))
}

func TestEnvelopeSerializeInformResponseAndChangeDUState(t *testing.T) {
	inform := messages.InformResponse{MaxEnvelopes: 1}
	cds := testChangeDUState()
	env := messages.NewEnvelope("test-id", &inform, &cds)
	assert.Equal(t, InformResponseAndChangeDUStateEnvelopeOutputXML, serializeEnvelope(t, env))
}

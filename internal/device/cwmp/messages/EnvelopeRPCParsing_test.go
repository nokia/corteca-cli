// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package messages_test

import (
	"bytes"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	InformResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:InformResponse>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:InformResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	ChangeDUStateEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:ChangeDUState>
      <CommandKey>testkey</CommandKey>
      <Operations>
        <InstallOpStruct>
          <URL>http://example.com/app.ipk</URL>
          <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
          <ExecutionEnvRef>generic</ExecutionEnvRef>
        </InstallOpStruct>
      </Operations>
    </cwmp:ChangeDUState>
  </soap-env:Body>
</soap-env:Envelope>`

	ChangeDUStateResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:ChangeDUStateResponse></cwmp:ChangeDUStateResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	DUStateChangeCompleteEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:DUStateChangeComplete>
      <CommandKey>testkey</CommandKey>
      <Results>
        <OpResultStruct>
          <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
          <DeploymentUnitRef>Device.SoftwareModules.DeploymentUnit.1</DeploymentUnitRef>
          <Version>1.0.0</Version>
          <CurrentState>Installed</CurrentState>
          <Resolved>true</Resolved>
          <ExecutionUnitRefList>exec-1</ExecutionUnitRefList>
          <StartTime>2026-04-08T10:00:00Z</StartTime>
          <CompleteTime>2026-04-08T10:01:30Z</CompleteTime>
          <Fault>
            <FaultCode>0</FaultCode>
            <FaultString></FaultString>
          </Fault>
        </OpResultStruct>
      </Results>
    </cwmp:DUStateChangeComplete>
  </soap-env:Body>
</soap-env:Envelope>`

	DUStateChangeCompleteResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:DUStateChangeCompleteResponse></cwmp:DUStateChangeCompleteResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	FaultEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <soap-env:Fault>
      <faultcode>Server</faultcode>
      <faultstring>CWMP Fault</faultstring>
      <detail>
        <cwmp:Fault>
          <FaultCode>9003</FaultCode>
          <FaultString>Invalid arguments</FaultString>
        </cwmp:Fault>
      </detail>
    </soap-env:Fault>
  </soap-env:Body>
</soap-env:Envelope>`

	GetParameterNamesEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:GetParameterNames>
      <ParameterPath>Device.DeviceInfo.</ParameterPath>
      <NextLevel>true</NextLevel>
    </cwmp:GetParameterNames>
  </soap-env:Body>
</soap-env:Envelope>`

	GetParameterNamesResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:GetParameterNamesResponse>
      <ParameterList>
        <ParameterInfoStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Writable>false</Writable>
        </ParameterInfoStruct>
      </ParameterList>
    </cwmp:GetParameterNamesResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	GetParameterValuesEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/">
  <soap-env:Body>
    <cwmp:GetParameterValues>
      <ParameterNames soap-enc:array="xsd:string[1]">
        <string>Device.DeviceInfo.SoftwareVersion</string>
      </ParameterNames>
    </cwmp:GetParameterValues>
  </soap-env:Body>
</soap-env:Envelope>`

	GetParameterValuesResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/">
  <soap-env:Body>
    <cwmp:GetParameterValuesResponse>
      <ParameterList soap-enc:array="cwmp:ParameterValueStruct[1]">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">1.0.0</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:GetParameterValuesResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	SetParameterValuesEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/">
  <soap-env:Body>
    <cwmp:SetParameterValues>
      <ParameterList soap-enc:array="cwmp:ParameterValueStruct[1]">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value xsi:type="xsd:string">2.0.0</Value>
        </ParameterValueStruct>
      </ParameterList>
      <ParameterKey>mykey</ParameterKey>
    </cwmp:SetParameterValues>
  </soap-env:Body>
</soap-env:Envelope>`

	SetParameterValuesResponseEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:SetParameterValuesResponse>
      <Status>0</Status>
    </cwmp:SetParameterValuesResponse>
  </soap-env:Body>
</soap-env:Envelope>`

	MultipleMessagesEnvelopeXML = `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <soap-env:Body>
    <cwmp:InformResponse>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:InformResponse>
    <cwmp:ChangeDUState>
      <CommandKey>testkey</CommandKey>
      <Operations>
        <InstallOpStruct>
          <URL>http://example.com/app.ipk</URL>
          <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
          <ExecutionEnvRef>generic</ExecutionEnvRef>
        </InstallOpStruct>
      </Operations>
    </cwmp:ChangeDUState>
  </soap-env:Body>
</soap-env:Envelope>`
)

func parseEnvelope(t *testing.T, xmlInput string) messages.Envelope {
	t.Helper()
	buf := bytes.NewBufferString(xmlInput)
	if env, err := messages.ParseEnvelopeXML(buf); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
		return messages.Envelope{}
	} else {
		return *env
	}
}

func TestEnvelopeParseInformResponse(t *testing.T) {
	env := parseEnvelope(t, InformResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.InformResponse)
	require.True(t, ok)
	assert.Equal(t, uint(1), msg.MaxEnvelopes)
}

func TestEnvelopeParseChangeDUState(t *testing.T) {
	env := parseEnvelope(t, ChangeDUStateEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.ChangeDUState)
	require.True(t, ok)
	assert.Equal(t, "testkey", msg.CommandKey.String())
	assert.Equal(t, 1, len(msg.Operations.Op))
	install, ok := msg.Operations.Op[0].(messages.InstallOpStruct)
	require.True(t, ok)
	assert.Equal(t, "http://example.com/app.ipk", install.URL.String())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", install.UUID.String())
	assert.Equal(t, "generic", install.ExecutionEnvRef.String())
}

func TestEnvelopeParseChangeDUStateResponse(t *testing.T) {
	env := parseEnvelope(t, ChangeDUStateResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	_, ok := env.GetBody()[0].(messages.ChangeDUStateResponse)
	require.True(t, ok)
}

func TestEnvelopeParseDUStateChangeComplete(t *testing.T) {
	env := parseEnvelope(t, DUStateChangeCompleteEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.DUStateChangeComplete)
	require.True(t, ok)
	assert.Equal(t, "testkey", msg.CommandKey)
	assert.Equal(t, 1, len(msg.Results))
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Results[0].UUID)
	assert.Equal(t, "Device.SoftwareModules.DeploymentUnit.1", msg.Results[0].DeploymentUnitRef)
	assert.Equal(t, "Installed", msg.Results[0].CurrentState)
	assert.Equal(t, uint(0), msg.Results[0].Fault.FaultCode)
}

func TestEnvelopeParseDUStateChangeCompleteResponse(t *testing.T) {
	env := parseEnvelope(t, DUStateChangeCompleteResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	_, ok := env.GetBody()[0].(messages.DUStateChangeCompleteResponse)
	require.True(t, ok)
}

func TestEnvelopeParseFault(t *testing.T) {
	env := parseEnvelope(t, FaultEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.Fault)
	require.True(t, ok)
	assert.Equal(t, "Server", msg.FaultCode)
	assert.Equal(t, "CWMP Fault", msg.FaultString)
	assert.Equal(t, uint(9003), msg.Detail.FaultCode)
	assert.Equal(t, "Invalid arguments", msg.Detail.FaultString)
}

func TestEnvelopeParseGetParameterNames(t *testing.T) {
	env := parseEnvelope(t, GetParameterNamesEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.GetParameterNames)
	require.True(t, ok)
	assert.Equal(t, "Device.DeviceInfo.", msg.ParameterPath.String())
	assert.Equal(t, true, msg.NextLevel)
}

func TestEnvelopeParseGetParameterNamesResponse(t *testing.T) {
	env := parseEnvelope(t, GetParameterNamesResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.GetParameterNamesResponse)
	require.True(t, ok)
	assert.Equal(t, 1, len(msg.ParameterList))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList[0].Name)
	assert.Equal(t, false, msg.ParameterList[0].Writable)
}

func TestEnvelopeParseGetParameterValues(t *testing.T) {
	env := parseEnvelope(t, GetParameterValuesEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.GetParameterValues)
	require.True(t, ok)
	assert.Equal(t, 1, len(msg.ParameterNames.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterNames.Params[0].String())
}

func TestEnvelopeParseGetParameterValuesResponse(t *testing.T) {
	env := parseEnvelope(t, GetParameterValuesResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.GetParameterValuesResponse)
	require.True(t, ok)
	assert.Equal(t, 1, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "1.0.0", msg.ParameterList.Params[0].Content.Value.String())
}

func TestEnvelopeParseSetParameterValues(t *testing.T) {
	env := parseEnvelope(t, SetParameterValuesEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.SetParameterValues)
	require.True(t, ok)
	assert.Equal(t, 1, len(msg.ParameterList.Params))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", msg.ParameterList.Params[0].Name.String())
	assert.Equal(t, messages.XsdString, msg.ParameterList.Params[0].Content.Type)
	assert.Equal(t, "2.0.0", msg.ParameterList.Params[0].Content.Value.String())
	assert.Equal(t, "mykey", msg.ParameterKey)
}

func TestEnvelopeParseSetParameterValuesResponse(t *testing.T) {
	env := parseEnvelope(t, SetParameterValuesResponseEnvelopeXML)
	require.Equal(t, 1, len(env.GetBody()))
	msg, ok := env.GetBody()[0].(messages.SetParameterValuesResponse)
	require.True(t, ok)
	assert.Equal(t, uint(0), msg.Status)
}

func TestEnvelopeParseMultipleMessages(t *testing.T) {
	env := parseEnvelope(t, MultipleMessagesEnvelopeXML)
	require.Equal(t, 2, len(env.GetBody()))

	inform, ok := env.GetBody()[0].(messages.InformResponse)
	require.True(t, ok, "expected GetBody()[0] to be messages.InformResponse")
	assert.Equal(t, uint(1), inform.MaxEnvelopes)

	cds, ok := env.GetBody()[1].(messages.ChangeDUState)
	require.True(t, ok, "expected GetBody()[1] to be *messages.ChangeDUState")
	assert.Equal(t, "testkey", cds.CommandKey.String())
	require.Equal(t, 1, len(cds.Operations.Op))
	install, ok := cds.Operations.Op[0].(messages.InstallOpStruct)
	require.True(t, ok, "expected Operations.Op[0] to be messages.InstallOpStruct")
	assert.Equal(t, "http://example.com/app.ipk", install.URL.String())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", install.UUID.String())
	assert.Equal(t, "generic", install.ExecutionEnvRef.String())
}

package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCreateXML(t *testing.T) {
	expectedOutput := `<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">5</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:SetParameterValues>
      <ParameterList xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" soap-enc:arrayType="cwmp:ParameterValueStruct[2]">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.ProvisioningCode</Name>
          <Value xsi:type="string">MyCustomModel123</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.X_ALU-COM_FriendlyName</Name>
          <Value xsi:type="string">myval</Value>
        </ParameterValueStruct>
      </ParameterList>
      <parameterKey>paramKey</parameterKey>
    </cwmp:SetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	setParamValues := NewSetParameterValues()
	setParamValues.ParameterList = []ParameterVal{
		{
			Name:  "Device.DeviceInfo.ProvisioningCode",
			Value: Values{Type: "string", Value: "MyCustomModel123"},
		},
		{
			Name:  "Device.DeviceInfo.X_ALU-COM_FriendlyName",
			Value: Values{Type: "string", Value: "myval"},
		},
	}
	setParamValues.ID = "5"
	setParamValues.ParameterKey = "paramKey"

	rpcXML, err := setParamValues.CreateXML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedOutput, string(rpcXML), "Strings should be equal")
}

func TestSetParseXML(t *testing.T) {
	expectedMsg := NewSetParameterValuesResponse()
	expectedMsg.ID = "5"
	expectedMsg.Status = 0

	body := `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
<soap-env:Header>
<cwmp:ID soap-env:mustUnderstand="1">5</cwmp:ID>
</soap-env:Header>
<soap-env:Body>
<cwmp:SetParameterValuesResponse>
<Status>0</Status>
</cwmp:SetParameterValuesResponse>
</soap-env:Body>
</soap-env:Envelope>`
	msg, err := ParseXML([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedMsg, msg, "Messages should be the same")
}

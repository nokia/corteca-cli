package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateXML(t *testing.T) {
	expectedOutput := `<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">5</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:GetParameterValues>
      <ParameterNames xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" soap-enc:arrayType="xsd:string[2]">
        <string>Device.DeviceInfo.ProvisioningCode</string>
        <string>Device.DeviceInfo.X_ALU-COM_FriendlyName</string>
      </ParameterNames>
    </cwmp:GetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	getParamValues := NewGetParameterValues()
	getParamValues.ParameterNames = []string{"Device.DeviceInfo.ProvisioningCode", "Device.DeviceInfo.X_ALU-COM_FriendlyName"}
	getParamValues.ID = "5"
	getParamValues.XMLName = "name"

	rpcXML, err := getParamValues.CreateXML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedOutput, string(rpcXML), "Strings should be equal")
}

func TestParseXML(t *testing.T) {
	expectedMsg := NewGetParameterValuesResponse()
	expectedMsg.ID = "5"
	expectedMsg.XMLName = "GetParameterValuesResponse"
	expectedMsg.ParameterList = append(expectedMsg.ParameterList, ParameterValuesInfoStruct{
		Name:  "Device.DeviceInfo.ProvisioningCode",
		Value: "MyCustomModel123",
	}, ParameterValuesInfoStruct{
		Name:  "Device.DeviceInfo.X_ALU-COM_FriendlyName",
		Value: "myval",
	})

	requestBody := `<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
<soap-env:Header>
<cwmp:ID soap-env:mustUnderstand="1">5</cwmp:ID>
</soap-env:Header>
<soap-env:Body>
<cwmp:GetParameterValuesResponse>
<ParameterList xsi:type="soap-enc:Array" soap-enc:arrayType="cwmp:ParameterValueStruct[2]">
<ParameterValueStruct>
<Name>Device.DeviceInfo.ProvisioningCode</Name>
<Value xsi:type="xsd:string">MyCustomModel123</Value>
</ParameterValueStruct>
<ParameterValueStruct>
<Name>Device.DeviceInfo.X_ALU-COM_FriendlyName</Name>
<Value xsi:type="xsd:string">myval</Value>
</ParameterValueStruct>
</ParameterList>
</cwmp:GetParameterValuesResponse>
</soap-env:Body>
</soap-env:Envelope>
`
	msg, err := ParseXML([]byte(requestBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedMsg, msg, "Messages should be the same")
}

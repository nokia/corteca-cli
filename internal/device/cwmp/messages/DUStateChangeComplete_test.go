package messages_test

import (
	"bytes"
	"corteca/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	DUStateChangeCompleteInputXML = `<cwmp:DUStateChangeComplete>
  <CommandKey>Foobar</CommandKey>
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
    <OpResultStruct>
      <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
      <DeploymentUnitRef>Device.SoftwareModules.DeploymentUnit.2</DeploymentUnitRef>
      <Version>1.0.0</Version>
      <CurrentState>Failed</CurrentState>
      <Resolved>true</Resolved>
      <ExecutionUnitRefList>exec-1</ExecutionUnitRefList>
      <StartTime>2026-04-08T10:00:00Z</StartTime>
      <CompleteTime>2026-04-08T10:01:30Z</CompleteTime>
      <Fault>
        <FaultCode>9027</FaultCode>
        <FaultString>System Resources Exceeded</FaultString>
      </Fault>
    </OpResultStruct>
  </Results>
</cwmp:DUStateChangeComplete>`

	DUStateChangeCompleteInputYAML = `CommandKey: Foobar
Results:
    - UUID: c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d
      DeploymentUnitRef: Device.SoftwareModules.DeploymentUnit.1
      Version: 1.0.0
      CurrentState: Installed
      Resolved: true
      ExecutionUnitRefList: exec-1
      StartTime: "2026-04-08T10:00:00Z"
      CompleteTime: "2026-04-08T10:01:30Z"
      Fault:
        FaultCode: 0
        FaultString: ""
    - UUID: c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d
      DeploymentUnitRef: Device.SoftwareModules.DeploymentUnit.2
      Version: 1.0.0
      CurrentState: Failed
      Resolved: true
      ExecutionUnitRefList: exec-1
      StartTime: "2026-04-08T10:00:00Z"
      CompleteTime: "2026-04-08T10:01:30Z"
      Fault:
        FaultCode: 9027
        FaultString: System Resources Exceeded
`
)

var DUStateChangeCompleteInputMsg = messages.DUStateChangeComplete{
	CommandKey: "Foobar",
	Results: []messages.OpResultStruct{
		{
			UUID:                 "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d",
			DeploymentUnitRef:    "Device.SoftwareModules.DeploymentUnit.1",
			Version:              "1.0.0",
			CurrentState:         "Installed",
			Resolved:             true,
			ExecutionUnitRefList: "exec-1",
			StartTime:            "2026-04-08T10:00:00Z",
			CompleteTime:         "2026-04-08T10:01:30Z",
			Fault:                messages.FaultStruct{FaultCode: 0, FaultString: ""},
		},
		{
			UUID:                 "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d",
			DeploymentUnitRef:    "Device.SoftwareModules.DeploymentUnit.2",
			Version:              "1.0.0",
			CurrentState:         "Failed",
			Resolved:             true,
			ExecutionUnitRefList: "exec-1",
			StartTime:            "2026-04-08T10:00:00Z",
			CompleteTime:         "2026-04-08T10:01:30Z",
			Fault:                messages.FaultStruct{FaultCode: 9027, FaultString: "System Resources Exceeded"},
		},
	},
}

func TestDUStateChangeCompleteParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(DUStateChangeCompleteInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.DUStateChangeComplete{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Foobar", msg.CommandKey)
	assert.Equal(t, 2, len(msg.Results))

	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Results[0].UUID)
	assert.Equal(t, "Device.SoftwareModules.DeploymentUnit.1", msg.Results[0].DeploymentUnitRef)
	assert.Equal(t, "1.0.0", msg.Results[0].Version)
	assert.Equal(t, "Installed", msg.Results[0].CurrentState)
	assert.Equal(t, true, msg.Results[0].Resolved)
	assert.Equal(t, "exec-1", msg.Results[0].ExecutionUnitRefList)
	assert.Equal(t, "2026-04-08T10:00:00Z", msg.Results[0].StartTime)
	assert.Equal(t, "2026-04-08T10:01:30Z", msg.Results[0].CompleteTime)
	assert.Equal(t, uint(0), msg.Results[0].Fault.FaultCode)
	assert.Equal(t, "", msg.Results[0].Fault.FaultString)

	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Results[1].UUID)
	assert.Equal(t, "Device.SoftwareModules.DeploymentUnit.2", msg.Results[1].DeploymentUnitRef)
	assert.Equal(t, "1.0.0", msg.Results[1].Version)
	assert.Equal(t, "Failed", msg.Results[1].CurrentState)
	assert.Equal(t, true, msg.Results[1].Resolved)
	assert.Equal(t, "exec-1", msg.Results[1].ExecutionUnitRefList)
	assert.Equal(t, "2026-04-08T10:00:00Z", msg.Results[1].StartTime)
	assert.Equal(t, "2026-04-08T10:01:30Z", msg.Results[1].CompleteTime)
	assert.Equal(t, uint(9027), msg.Results[1].Fault.FaultCode)
	assert.Equal(t, "System Resources Exceeded", msg.Results[1].Fault.FaultString)
}

func TestDUStateChangeCompleteSerializeToXML(t *testing.T) {
	msg := DUStateChangeCompleteInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating xml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, DUStateChangeCompleteInputXML, buf.String())
}

func TestDUStateChangeCompleteParseFromYAML(t *testing.T) {
	buf := bytes.NewBufferString(DUStateChangeCompleteInputYAML)
	dec := yaml.NewDecoder(buf)
	msg := messages.DUStateChangeComplete{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing YAML input: %s", err.Error())
		t.FailNow()
	}

	assert.Equal(t, "Foobar", msg.CommandKey)
	assert.Equal(t, 2, len(msg.Results))

	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Results[0].UUID)
	assert.Equal(t, "Device.SoftwareModules.DeploymentUnit.1", msg.Results[0].DeploymentUnitRef)
	assert.Equal(t, "1.0.0", msg.Results[0].Version)
	assert.Equal(t, "Installed", msg.Results[0].CurrentState)
	assert.Equal(t, true, msg.Results[0].Resolved)
	assert.Equal(t, "exec-1", msg.Results[0].ExecutionUnitRefList)
	assert.Equal(t, "2026-04-08T10:00:00Z", msg.Results[0].StartTime)
	assert.Equal(t, "2026-04-08T10:01:30Z", msg.Results[0].CompleteTime)
	assert.Equal(t, uint(0), msg.Results[0].Fault.FaultCode)
	assert.Equal(t, "", msg.Results[0].Fault.FaultString)

	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Results[1].UUID)
	assert.Equal(t, "Device.SoftwareModules.DeploymentUnit.2", msg.Results[1].DeploymentUnitRef)
	assert.Equal(t, "1.0.0", msg.Results[1].Version)
	assert.Equal(t, "Failed", msg.Results[1].CurrentState)
	assert.Equal(t, true, msg.Results[1].Resolved)
	assert.Equal(t, "exec-1", msg.Results[1].ExecutionUnitRefList)
	assert.Equal(t, "2026-04-08T10:00:00Z", msg.Results[1].StartTime)
	assert.Equal(t, "2026-04-08T10:01:30Z", msg.Results[1].CompleteTime)
	assert.Equal(t, uint(9027), msg.Results[1].Fault.FaultCode)
	assert.Equal(t, "System Resources Exceeded", msg.Results[1].Fault.FaultString)
}

func TestDUStateChangeCompleteSerializeToYAML(t *testing.T) {
	msg := DUStateChangeCompleteInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating YAML output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, DUStateChangeCompleteInputYAML, outbuf.String())
}

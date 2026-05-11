package messages_test

import (
	"bytes"
	"github.com/nokia/corteca-cli/internal/configuration"
	c "github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device/cwmp/messages"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	ChangeDUStateInputXML = `<cwmp:ChangeDUState>
  <CommandKey>foo</CommandKey>
  <Operations>
    <InstallOpStruct>
      <URL>http://example.com/some/image:1.0.0</URL>
      <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
      <Username>some.user@example.com</Username>
      <Password>somepassword</Password>
      <ExecutionEnvRef>generic</ExecutionEnvRef>
    </InstallOpStruct>
    <UpdateOpStruct>
      <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
      <Version>1.0.0</Version>
      <URL>http://example.com/some/image:1.0.0</URL>
      <Username>some.user@example.com</Username>
      <Password>somepassword</Password>
    </UpdateOpStruct>
    <UninstallOpStruct>
      <UUID>c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d</UUID>
      <Version>1.0.0</Version>
      <ExecutionEnvRef>generic</ExecutionEnvRef>
    </UninstallOpStruct>
  </Operations>
</cwmp:ChangeDUState>`

	ChangeDUStateInputYaml = `CommandKey: foo
Operations:
    - !InstallOpStruct
      URL: http://example.com/some/${app.name}:${app.version}
      UUID: ${app.duid}
      Username: some.user@example.com
      Password: somepassword
      ExecutionEnvRef: generic
    - !UpdateOpStruct
      UUID: c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d
      Version: 1.0.0
      URL: http://example.com/some/image:1.0.0
      Username: some.user@example.com
      Password: somepassword
    - !UninstallOpStruct
      UUID: c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d
      Version: 1.0.0
      ExecutionEnvRef: generic
`
)

var ChangeDUStateInputMsg = messages.ChangeDUState{
	CommandKey: c.TemplateField{RawTemplate: "foo"},
	Operations: messages.DUOperationStruct{
		Op: []messages.DUOperation{
			messages.InstallOpStruct{
				URL:             c.TemplateField{RawTemplate: "http://example.com/some/${app.name}:${app.version}"},
				UUID:            c.TemplateField{RawTemplate: "${app.duid}"},
				Username:        c.TemplateField{RawTemplate: "some.user@example.com"},
				Password:        c.TemplateField{RawTemplate: "somepassword"},
				ExecutionEnvRef: c.TemplateField{RawTemplate: "generic"},
			},
			messages.UpdateOpStruct{
				URL:      c.TemplateField{RawTemplate: "http://example.com/some/image:1.0.0"},
				UUID:     c.TemplateField{RawTemplate: "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d"},
				Username: c.TemplateField{RawTemplate: "some.user@example.com"},
				Password: c.TemplateField{RawTemplate: "somepassword"},
				Version:  c.TemplateField{RawTemplate: "1.0.0"},
			},
			messages.UninstallOpStruct{
				UUID:            c.TemplateField{RawTemplate: "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d"},
				ExecutionEnvRef: c.TemplateField{RawTemplate: "generic"},
				Version:         c.TemplateField{RawTemplate: "1.0.0"},
			},
		},
	},
}

func setupContext() {
	configuration.ResetContext()
	configuration.GetCmdContext().App.Name = "image"
	configuration.GetCmdContext().App.Version = "1.0.0"
	configuration.GetCmdContext().App.DUID = "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d"
}

func TestChangeDUStateParseFromXML(t *testing.T) {
	buf := bytes.NewBufferString(ChangeDUStateInputXML)
	dec := xml.NewDecoder(buf)
	msg := messages.ChangeDUState{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing XML input: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, "foo", msg.CommandKey.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[0].(messages.InstallOpStruct) })
	assert.Equal(t, "http://example.com/some/image:1.0.0", msg.Operations.Op[0].(messages.InstallOpStruct).URL.String())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[0].(messages.InstallOpStruct).UUID.String())
	assert.Equal(t, "some.user@example.com", msg.Operations.Op[0].(messages.InstallOpStruct).Username.String())
	assert.Equal(t, "somepassword", msg.Operations.Op[0].(messages.InstallOpStruct).Password.String())
	assert.Equal(t, "generic", msg.Operations.Op[0].(messages.InstallOpStruct).ExecutionEnvRef.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[1].(messages.UpdateOpStruct) })
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[1].(messages.UpdateOpStruct).UUID.String())
	assert.Equal(t, "1.0.0", msg.Operations.Op[1].(messages.UpdateOpStruct).Version.String())
	assert.Equal(t, "http://example.com/some/image:1.0.0", msg.Operations.Op[1].(messages.UpdateOpStruct).URL.String())
	assert.Equal(t, "some.user@example.com", msg.Operations.Op[1].(messages.UpdateOpStruct).Username.String())
	assert.Equal(t, "somepassword", msg.Operations.Op[1].(messages.UpdateOpStruct).Password.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[2].(messages.UninstallOpStruct) })
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[2].(messages.UninstallOpStruct).UUID.String())
	assert.Equal(t, "1.0.0", msg.Operations.Op[2].(messages.UninstallOpStruct).Version.String())
	assert.Equal(t, "generic", msg.Operations.Op[2].(messages.UninstallOpStruct).ExecutionEnvRef.String())

}

func TestChangeDUStateSerializeToXML(t *testing.T) {
	setupContext()
	msg := ChangeDUStateInputMsg
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating yaml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, ChangeDUStateInputXML, buf.String())
}

func TestChangeDUStateUmarshalFromYaml(t *testing.T) {
	setupContext()
	buf := bytes.NewBufferString(ChangeDUStateInputYaml)
	dec := yaml.NewDecoder(buf)
	msg := messages.ChangeDUState{}
	if err := dec.Decode(&msg); err != nil {
		t.Logf("Failed parsing yaml input: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, "foo", msg.CommandKey.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[0].(messages.InstallOpStruct) })
	assert.Equal(t, "InstallOpStruct", msg.Operations.Op[0].GetOpType())
	assert.Equal(t, "http://example.com/some/image:1.0.0", msg.Operations.Op[0].(messages.InstallOpStruct).URL.String())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[0].(messages.InstallOpStruct).UUID.String())
	assert.Equal(t, "some.user@example.com", msg.Operations.Op[0].(messages.InstallOpStruct).Username.String())
	assert.Equal(t, "somepassword", msg.Operations.Op[0].(messages.InstallOpStruct).Password.String())
	assert.Equal(t, "generic", msg.Operations.Op[0].(messages.InstallOpStruct).ExecutionEnvRef.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[1].(messages.UpdateOpStruct) })
	assert.Equal(t, "UpdateOpStruct", msg.Operations.Op[1].GetOpType())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[1].(messages.UpdateOpStruct).UUID.String())
	assert.Equal(t, "1.0.0", msg.Operations.Op[1].(messages.UpdateOpStruct).Version.String())
	assert.Equal(t, "http://example.com/some/image:1.0.0", msg.Operations.Op[1].(messages.UpdateOpStruct).URL.String())
	assert.Equal(t, "some.user@example.com", msg.Operations.Op[1].(messages.UpdateOpStruct).Username.String())
	assert.Equal(t, "somepassword", msg.Operations.Op[1].(messages.UpdateOpStruct).Password.String())

	assert.NotPanics(t, func() { _ = msg.Operations.Op[2].(messages.UninstallOpStruct) })
	assert.Equal(t, "UninstallOpStruct", msg.Operations.Op[2].GetOpType())
	assert.Equal(t, "c0c4328b-18a4-4b3b-b1da-e8ea8d8f457d", msg.Operations.Op[2].(messages.UninstallOpStruct).UUID.String())
	assert.Equal(t, "1.0.0", msg.Operations.Op[2].(messages.UninstallOpStruct).Version.String())
	assert.Equal(t, "generic", msg.Operations.Op[2].(messages.UninstallOpStruct).ExecutionEnvRef.String())
}

func TestChangeDUStateMarshalToYaml(t *testing.T) {
	msg := ChangeDUStateInputMsg
	outbuf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := yaml.NewEncoder(outbuf)
	enc.SetIndent(4)
	if err := enc.Encode(msg); err != nil {
		t.Logf("Failed generating yaml output: %s", err.Error())
		t.FailNow()
	}
	assert.Equal(t, ChangeDUStateInputYaml, outbuf.String())
}


package device

import (
	"io"
	"testing"

	"corteca/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ssh"
)

// --- Mock Interfaces ---

type MockSSHClient struct {
	mock.Mock
}

func (m *MockSSHClient) NewSession() (*ssh.Session, error) {
	args := m.Called()
	return args.Get(0).(*ssh.Session), args.Error(1)
}

func (m *MockSSHClient) Close() error {
	return m.Called().Error(0)
}

type MockSSHSession struct {
	mock.Mock
}

func (m *MockSSHSession) Run(cmd string) error {
	return m.Called(cmd).Error(0)
}

func (m *MockSSHSession) StdinPipe() (io.WriteCloser, error) {
	args := m.Called()
	return args.Get(0).(io.WriteCloser), args.Error(1)
}

func (m *MockSSHSession) Shell() error {
	return m.Called().Error(0)
}

func (m *MockSSHSession) Close() error {
	return m.Called().Error(0)
}

func (m *MockSSHSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return m.Called(term, h, w, modes).Error(0)
}

// --- Tests ---

func TestNewSSHDevice(t *testing.T) {
	endpoint := configuration.Endpoint{
		Addr:           configuration.TemplateField{RawTemplate: "ssh://user:pass@localhost:22"},
		Auth:           "password",
		Username:       configuration.TemplateField{RawTemplate: "user"},
		Password:       configuration.TemplateField{RawTemplate: "password"},
		Password2:      configuration.TemplateField{RawTemplate: "password2"},
		PrivateKeyFile: configuration.TemplateField{RawTemplate: "/path/to/key"},
	}
	device, err := NewSSHDevice(endpoint, "test.log")
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, endpoint.Addr.String(), device.urlInfo.String())
	assert.Equal(t, endpoint.Auth, device.auth)
	assert.Equal(t, endpoint.Token.String(), device.token)
	assert.Equal(t, endpoint.Password2.String(), device.password2)
	assert.Equal(t, endpoint.PrivateKeyFile.String(), device.keyFile)
}

func TestGetProtocol(t *testing.T) {
	device := &SSHDevice{}
	assert.Equal(t, ConnectionSSH, device.GetProtocol())
}

package device_test

import (
	"corteca/internal/configuration"
	"corteca/internal/device"
	"errors"
	"os"
	"testing"
)

// MockDispatcher simulates dispatcher.Dispatcher behavior
type MockDispatcher struct {
	Responses   map[string]string
	Failures    map[string]error
	printFormat string
}

func (m *MockDispatcher) ExecuteCommand(cmd any) (string, error) {
	commandStr, ok := cmd.(string)
	if !ok {
		return "", errors.New("invalid command type")
	}
	if err, exists := m.Failures[commandStr]; exists {
		return "", err
	}
	if output, exists := m.Responses[commandStr]; exists {
		return output, nil
	}
	return "", nil
}

func (m *MockDispatcher) SetPrintFormat(format string) {
	m.printFormat = format
}

// --- Tests ---

func TestNewDevice_SSH(t *testing.T) {
	endpoint := configuration.Endpoint{
		Addr: configuration.TemplateField{RawTemplate: "ssh://user@localhost"},
	}
	dev, err := device.NewDevice(endpoint, "stdout")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dev.GetProtocol() != device.ConnectionSSH {
		t.Errorf("expected SSH protocol, got %d", dev.GetProtocol())
	}
}

func TestNewDevice_Unsupported(t *testing.T) {
	endpoint := configuration.Endpoint{
		Addr: configuration.TemplateField{RawTemplate: "ftp://localhost"},
	}
	_, err := device.NewDevice(endpoint, "stdout")
	if err == nil {
		t.Fatal("expected error for unsupported protocol")
	}
}

func TestNewLogger_Stdout(t *testing.T) {
	logger, err := device.NewLogger("stdout")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger == nil || logger.LogFile != os.Stdout {
		t.Error("expected logger to use stdout")
	}
}

func TestNewLogger_File(t *testing.T) {
	filename := "test.log"
	defer os.Remove(filename)

	logger, err := device.NewLogger(filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger == nil || logger.LogFile == nil {
		t.Error("expected logger to open file")
	}
}

func TestDetectContainerFramework_OCI(t *testing.T) {
	mock := &MockDispatcher{
		Responses: map[string]string{
			"lcm list": "running",
		},
	}
	result := device.DetectContainerFramework(mock)
	if result != "oci" {
		t.Errorf("expected 'oci', got %s", result)
	}
}

func TestDetectContainerFramework_RootFS(t *testing.T) {
	mock := &MockDispatcher{
		Failures: map[string]error{
			"lcm list": errors.New("not found"),
		},
		Responses: map[string]string{
			"pgrep PluginMgr": "PluginMgr",
		},
	}
	result := device.DetectContainerFramework(mock)
	if result != "rootfs" {
		t.Errorf("expected 'rootfs', got %s", result)
	}
}

func TestDetectContainerFramework_Unknown(t *testing.T) {
	mock := &MockDispatcher{
		Failures: map[string]error{
			"lcm list":        errors.New("not found"),
			"pgrep PluginMgr": errors.New("not found"),
		},
	}
	result := device.DetectContainerFramework(mock)
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

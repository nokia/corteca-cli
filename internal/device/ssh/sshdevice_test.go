// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package ssh_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nokia/corteca-cli/internal/configuration"
	devssh "github.com/nokia/corteca-cli/internal/device/ssh"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Test helpers
// =============================================================================

const testPassword = "s3cr3t-test-password"

// mustDeviceConfig unmarshals yamlStr into a *configuration.DeviceConfig,
// ensuring the internal raw yaml.Node is populated (required by DeviceConfig.Decode).
func mustDeviceConfig(t *testing.T, yamlStr string) *configuration.DeviceConfig {
	t.Helper()
	var cfg configuration.DeviceConfig
	if err := yaml.Unmarshal([]byte(yamlStr), &cfg); err != nil {
		t.Fatalf("mustDeviceConfig: %v", err)
	}
	return &cfg
}

// mustSequenceCmd unmarshals yamlStr into a *configuration.SequenceCmd,
// ensuring the internal raw yaml.Node is populated (required by SequenceCmd.Decode,
// which is called inside SSHDevice.ExecuteCommand to decode the params field).
func mustSequenceCmd(t *testing.T, yamlStr string) *configuration.SequenceCmd {
	t.Helper()
	var cmd configuration.SequenceCmd
	if err := yaml.Unmarshal([]byte(yamlStr), &cmd); err != nil {
		t.Fatalf("mustSequenceCmd: %v", err)
	}
	return &cmd
}

// =============================================================================
// Authentication tests
// =============================================================================

// TestSSHDevice_PasswordAuth verifies that the correct username and password
// are sent to the server under all four combinations of URL-embedded and
// explicit-field credentials. The priority rule is: explicit Username/Password
// fields in the device config take precedence over credentials embedded in the
// addr URL.
func TestSSHDevice_PasswordAuth(t *testing.T) {
	tests := []struct {
		name         string
		expectedUser string
		expectedPass string
		buildCfg     func(addr string) string
	}{
		{
			// Baseline: both username and password come from the addr URL.
			name:         "url_credentials_only",
			expectedUser: "url-user",
			expectedPass: "url-pass",
			buildCfg: func(addr string) string {
				return fmt.Sprintf("addr: ssh://url-user:url-pass@%s", addr)
			},
		},
		{
			// Both explicit fields are set; they must override the URL credentials.
			name:         "explicit_fields_override_url",
			expectedUser: "explicit-user",
			expectedPass: "explicit-pass",
			buildCfg: func(addr string) string {
				return fmt.Sprintf(
					"addr: ssh://url-user:url-pass@%s\nusername: explicit-user\npassword: explicit-pass\n",
					addr,
				)
			},
		},
		{
			// Only the explicit username field is set; it must override the URL
			// username while the password still comes from the URL.
			name:         "explicit_username_overrides_url",
			expectedUser: "explicit-user",
			expectedPass: "url-pass",
			buildCfg: func(addr string) string {
				return fmt.Sprintf(
					"addr: ssh://url-user:url-pass@%s\nusername: explicit-user\n",
					addr,
				)
			},
		},
		{
			// Only the explicit password field is set; it must override the URL
			// password while the username still comes from the URL.
			name:         "explicit_password_overrides_url",
			expectedUser: "url-user",
			expectedPass: "explicit-pass",
			buildCfg: func(addr string) string {
				return fmt.Sprintf(
					"addr: ssh://url-user:url-pass@%s\npassword: explicit-pass\n",
					addr,
				)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr := startTestServer(t, tc.expectedUser, tc.expectedPass, nil,
				withQuaggaProbe(func(cmd string) (string, uint32) {
					return "", 0
				}),
			)

			cfg := mustDeviceConfig(t, tc.buildCfg(addr))
			dev, err := devssh.NewSSHDevice(cfg, io.Discard)
			if err != nil {
				t.Fatalf("expected successful connection, got: %v", err)
			}
			dev.Close()
		})
	}
}

// TestSSHDevice_WrongPassword_ReturnsError verifies that NewSSHDevice returns an
// error when the supplied password is rejected by the server.
func TestSSHDevice_WrongPassword_ReturnsError(t *testing.T) {
	addr := startTestServer(t, "", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		return "", 0
	}))

	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", "wrong-password", addr))
	_, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err == nil {
		t.Fatal("expected error with wrong password, got nil")
	}
}

// TestSSHDevice_PublicKeyAuth verifies that NewSSHDevice connects successfully
// when a valid ECDSA private key file is provided and the server accepts the
// corresponding public key.
func TestSSHDevice_PublicKeyAuth(t *testing.T) {
	// Generate a fresh ECDSA P-256 key pair for this test.
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate client key: %v", err)
	}
	sshPubKey, err := ssh.NewPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatalf("create SSH public key: %v", err)
	}

	// Write the private key to a temporary PEM file so that connectSSHClient
	// can read it via os.ReadFile / ssh.ParsePrivateKey.
	keyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		t.Fatalf("marshal EC private key: %v", err)
	}
	keyFile, err := os.CreateTemp(t.TempDir(), "test-key-*.pem")
	if err != nil {
		t.Fatalf("create temp key file: %v", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		t.Fatalf("write PEM key: %v", err)
	}
	_ = keyFile.Close()

	// Start a server that only accepts the generated public key (no password).
	addr := startTestServer(t, "testuser", "", sshPubKey, withQuaggaProbe(func(cmd string) (string, uint32) {
		return "", 0
	}))

	cfg := mustDeviceConfig(t, fmt.Sprintf(
		"addr: ssh://testuser@%s\nprivateKeyFile: %s\n",
		addr, keyFile.Name(),
	))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("expected successful connection with public key, got: %v", err)
	}
	dev.Close()
}

// =============================================================================
// Protocol / lifecycle tests
// =============================================================================

// TestSSHDevice_GetProtocol verifies that GetProtocol returns "ssh".
func TestSSHDevice_GetProtocol(t *testing.T) {
	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		return "", 0
	}))
	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer dev.Close()

	if got := dev.GetProtocol(); got != "ssh" {
		t.Errorf("GetProtocol: expected %q, got %q", "ssh", got)
	}
}

// TestSSHDevice_BeginAndEndSequence verifies that both BeginSequence and
// EndSequence are no-ops that return nil.
func TestSSHDevice_BeginAndEndSequence(t *testing.T) {
	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		return "", 0
	}))
	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer dev.Close()

	if err := dev.BeginSequence(); err != nil {
		t.Errorf("BeginSequence: expected nil, got %v", err)
	}
	if err := dev.EndSequence(); err != nil {
		t.Errorf("EndSequence: expected nil, got %v", err)
	}
}

// =============================================================================
// ExecuteCommand tests
// =============================================================================

// TestSSHDevice_ExecuteCommand_OutputCaptured verifies that stdout produced by
// the remote command is written to the log writer supplied to NewSSHDevice.
func TestSSHDevice_ExecuteCommand_OutputCaptured(t *testing.T) {
	const want = "hello from server"

	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		return want + "\n", 0
	}))

	var logBuf bytes.Buffer
	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, &logBuf)
	if err != nil {
		t.Fatalf("unexpected error creating device: %v", err)
	}
	defer dev.Close()

	cmd := mustSequenceCmd(t, "cmd: echo-test")
	if _, err := dev.ExecuteCommand(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error executing command: %v", err)
	}

	if !strings.Contains(logBuf.String(), want) {
		t.Errorf("log writer: expected to contain %q, got %q", want, logBuf.String())
	}
}

// TestSSHDevice_ExecuteCommand_ParamsConcatenated verifies that the Cmd field
// and the Params array are joined into a single space-separated string before
// being sent to the server.
func TestSSHDevice_ExecuteCommand_ParamsConcatenated(t *testing.T) {
	received := make(chan string, 1)

	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		received <- strings.TrimSpace(cmd)
		return "", 0
	}))

	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error creating device: %v", err)
	}
	defer dev.Close()

	cmd := mustSequenceCmd(t, `
cmd: echo
params:
  - hello
  - world
`)
	if _, err := dev.ExecuteCommand(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error executing command: %v", err)
	}

	select {
	case got := <-received:
		const want = "echo hello world"
		if got != want {
			t.Errorf("server received %q; want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server to receive command")
	}
}

// TestSSHDevice_ExecuteCommand_ExitError verifies that a non-zero exit code from
// the remote command causes ExecuteCommand to return a non-nil error whose
// message includes the exit code, and that the captured stdout is returned as
// the result value.
func TestSSHDevice_ExecuteCommand_ExitError(t *testing.T) {
	const cmdOutput = "something went wrong\n"

	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		return cmdOutput, 1
	}))

	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error creating device: %v", err)
	}
	defer dev.Close()

	cmd := mustSequenceCmd(t, "cmd: failing-cmd")
	result, err := dev.ExecuteCommand(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error for non-zero exit code, got nil")
	}
	if !strings.Contains(err.Error(), "exit code (1)") {
		t.Errorf("error %q does not mention exit code 1", err.Error())
	}
	if result == nil {
		t.Error("expected non-nil result buffer when command fails with an exit error")
	}
}

// TestSSHDevice_ExecuteCommand_ContextCancellation verifies that cancelling the
// context while a command is in progress causes ExecuteCommand to return
// context.Canceled promptly.
func TestSSHDevice_ExecuteCommand_ContextCancellation(t *testing.T) {
	// unblock is closed by t.Cleanup to release the blocking handler goroutine
	// after the test has finished, preventing a goroutine leak.
	unblock := make(chan struct{})
	t.Cleanup(func() { close(unblock) })

	addr := startTestServer(t, "testuser", testPassword, nil, withQuaggaProbe(func(cmd string) (string, uint32) {
		<-unblock // block until the test is done
		return "", 1
	}))

	cfg := mustDeviceConfig(t, fmt.Sprintf("addr: ssh://testuser:%s@%s", testPassword, addr))
	dev, err := devssh.NewSSHDevice(cfg, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error creating device: %v", err)
	}
	defer dev.Close()

	ctx, cancel := context.WithCancel(context.Background())

	cmd := mustSequenceCmd(t, "cmd: sleep-forever")
	errCh := make(chan error, 1)
	go func() {
		_, err := dev.ExecuteCommand(ctx, cmd)
		errCh <- err
	}()

	// Give the command a moment to reach the server before cancelling.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ExecuteCommand to return after context cancellation")
	}
}

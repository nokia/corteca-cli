// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device_test

import (
	"context"
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device"
	"io"
	"testing"
)

// mockDevice is a minimal Device implementation used in tests.
type mockDevice struct {
	protocol string
}

func (m *mockDevice) Close() {}

func (m *mockDevice) GetProtocol() string {
	return m.protocol
}

func (m *mockDevice) BeginSequence() error {
	return nil
}

func (m *mockDevice) ExecuteCommand(_ context.Context, _ *configuration.SequenceCmd) (any, error) {
	return nil, nil
}

func (m *mockDevice) EndSequence() error {
	return nil
}

// makeCreator returns a DeviceCreator that records whether it was called and
// returns a mockDevice with the given protocol label.
func makeCreator(protocol string, called *bool) device.DeviceCreator {
	return func(cfg *configuration.DeviceConfig, log io.Writer) (device.Device, error) {
		*called = true
		return &mockDevice{protocol: protocol}, nil
	}
}

func TestNewDevice_CorrectCreatorIsDispatched(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		protocol string
	}{
		{name: "alpha schema", schema: "alpha", protocol: "alpha"},
		{name: "beta schema", schema: "beta", protocol: "beta"},
		{name: "gamma schema", schema: "gamma", protocol: "gamma"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calledAlpha := false
			calledBeta := false
			calledGamma := false

			device.RegisterDeviceType("alpha", makeCreator("alpha", &calledAlpha))
			device.RegisterDeviceType("beta", makeCreator("beta", &calledBeta))
			device.RegisterDeviceType("gamma", makeCreator("gamma", &calledGamma))

			cfg := &configuration.DeviceConfig{
				Endpoint: configuration.Endpoint{
					Addr: configuration.T(tc.schema + "://some-host"),
				},
			}

			dev, err := device.NewDevice(cfg, io.Discard)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dev.GetProtocol() != tc.protocol {
				t.Errorf("expected protocol %q, got %q", tc.protocol, dev.GetProtocol())
			}

			// Verify exactly the right creator was called.
			if tc.schema == "alpha" && !calledAlpha {
				t.Error("expected alpha creator to be called")
			}
			if tc.schema == "beta" && !calledBeta {
				t.Error("expected beta creator to be called")
			}
			if tc.schema == "gamma" && !calledGamma {
				t.Error("expected gamma creator to be called")
			}

			// Verify the other creators were NOT called.
			if tc.schema != "alpha" && calledAlpha {
				t.Error("alpha creator should not have been called")
			}
			if tc.schema != "beta" && calledBeta {
				t.Error("beta creator should not have been called")
			}
			if tc.schema != "gamma" && calledGamma {
				t.Error("gamma creator should not have been called")
			}
		})
	}
}

func TestNewDevice_UnknownSchema_ReturnsError(t *testing.T) {
	device.RegisterDeviceType("alpha", makeCreator("alpha", new(bool)))
	device.RegisterDeviceType("beta", makeCreator("beta", new(bool)))
	device.RegisterDeviceType("gamma", makeCreator("gamma", new(bool)))

	cfg := &configuration.DeviceConfig{
		Endpoint: configuration.Endpoint{
			Addr: configuration.T("unknown://some-host"),
		},
	}

	dev, err := device.NewDevice(cfg, io.Discard)
	if err == nil {
		t.Fatal("expected an error for unknown schema, got nil")
	}
	if dev != nil {
		t.Errorf("expected nil device for unknown schema, got %v", dev)
	}
}

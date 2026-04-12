// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"corteca/internal/configuration"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Device defines the interface for device operations
type Device interface {
	Close()
	GetProtocol() string
	configuration.CommandExecutor
}

type DeviceCreator func(*configuration.DeviceConfig, io.Writer) (Device, error)

var deviceTypeRegistry map[string]DeviceCreator

func RegisterDeviceType(typename string, creator DeviceCreator) {
	deviceTypeRegistry[strings.ToLower(typename)] = creator
}

func init() {
	deviceTypeRegistry = make(map[string]DeviceCreator)
}

// NewDevice is a factory method that creates a Device based on the endpoint protocol
func NewDevice(config *configuration.DeviceConfig, log io.Writer) (Device, error) {
	u, err := url.Parse(config.Addr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint address: %w", err)
	}
	typename := strings.ToLower(u.Scheme)
	if creator, found := deviceTypeRegistry[typename]; found {
		return creator(config, log)
	} else {
		return nil, fmt.Errorf("unsupported device connection type '%s'", typename)
	}
}

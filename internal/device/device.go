// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"corteca/internal/configuration"
	"corteca/internal/dispatcher"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Connection types
const (
	ConnectionTelnet = iota
	ConnectionTelnetS
	ConnectionSSH
	ConnectionFIFO
	ConnectionCWMP
)

// Command constants
const (
	cmdLCMList       = "lcm list"
	cmdGrepPluginMgr = "pgrep PluginMgr"
)

// Logger handles logging to a file or standard output
type Logger struct {
	LogFile *os.File
}

// Device defines the interface for device operations
type Device interface {
	Connect() (dispatcher.Dispatcher, error)
	Close()
	GetProtocol() int
}

// NewDevice is a factory method that creates a Device based on the endpoint protocol
func NewDevice(endpoint configuration.Endpoint, logfile string) (Device, error) {
	u, err := url.Parse(endpoint.Addr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint address: %w", err)
	}

	switch strings.ToLower(u.Scheme) {
	case "ssh":
		return NewSSHDevice(endpoint, logfile)
	case "cwmp":
		return NewCWMPDevice(endpoint, logfile)
	case "cwmps":
		return NewCWMPsDevice(endpoint, logfile)
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", u.Scheme)
	}
}

// NewLogger initializes a Logger instance
func NewLogger(filename string) (*Logger, error) {
	logger := &Logger{}
	if err := logger.SetLogFile(filename); err != nil {
		return nil, err
	}
	return logger, nil
}

// SetLogFile configures the log output destination
func (logger *Logger) SetLogFile(filename string) error {
	switch filename {
	case "stdout":
		logger.LogFile = os.Stdout
	case "stderr":
		logger.LogFile = os.Stderr
	default:
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logger.LogFile = file
	}
	return nil
}

// DetectContainerFramework returns the container framework type based on command results
func DetectContainerFramework(d dispatcher.Dispatcher) string {
	if _, err := d.ExecuteCommand(cmdLCMList); err == nil {
		return "oci"
	}
	if _, err := d.ExecuteCommand(cmdGrepPluginMgr); err == nil {
		return "rootfs"
	}
	return ""
}

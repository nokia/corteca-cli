// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"bytes"
	"corteca/internal/configuration"
	"corteca/internal/templating"
	"corteca/internal/tui"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	MaxNumRetries  = 3
	DefaultSSHPort = 22
)

const (
	CONNECTION_TELNET = iota
	CONNECTION_TELNETS
	CONNECTION_SSH
	CONNECTION_FIFO
)

type Connection struct {
	// Deliberately hidden from the user
	handler   any
	errorChan <-chan error
	logFile   *os.File
	Type      int
	Address   string
}

func Connect(device *configuration.Endpoint, logFile string) (*Connection, error) {
	var (
		conn *Connection
		err  error
	)
	u, err := url.Parse(device.Addr)
	if err != nil {
		return nil, err
	}
	protocol := strings.ToLower(u.Scheme)
	switch protocol {
	case "ssh":
		conn, err = connectSSH(u, device)
	case "telnet":
		return nil, fmt.Errorf("telnet not supported yet; to be supported soon")
	case "telnets":
		return nil, fmt.Errorf("telnets not supported yet; to be supported soon")
	case "fifo":
		return nil, fmt.Errorf("pipe (fifo) not supported yet; to be supported soon")
	default:
		return nil, fmt.Errorf("unsupported target communication protocol '%s'", protocol)
	}
	// if a logfile was specified, open it and append stream output
	if err == nil && logFile != "" {
		err = conn.SetLogFile(logFile)
	}
	return conn, err
}

func connectSSH(u *url.URL, device *configuration.Endpoint) (*Connection, error) {
	sshConfig := ssh.ClientConfig{
		User:            u.User.Username(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: remove and either use $HOME/.ssh/known_hosts or prompt each time
	}

	// add default ssh port if not specified
	if u.Port() == "" {
		u.Host = u.Host + fmt.Sprintf(":%v", DefaultSSHPort)
	}

	if password, passwordSet := u.User.Password(); passwordSet {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		switch device.Auth {
		case configuration.AUTH_SSH_PASSWORD:
			sshConfig.Auth = []ssh.AuthMethod{ssh.RetryableAuthMethod(ssh.KeyboardInteractive(keyboardChallenge), MaxNumRetries)}
		case configuration.AUTH_SSH_PUBLIC_KEY:
			keyFile := device.PrivateKeyFile
			if keyFile == "" {
				return nil, fmt.Errorf("cannot use ssh public key auth method without a key file")
			}
			key, err := os.ReadFile(keyFile)
			if err != nil {
				return nil, err
			}
			signer, err := ssh.ParsePrivateKey(key)
			// TODO: test if err is PassphraseMissingError and re-parse with ssh.ParsePrivateKeyWithPassphrase()
			if err != nil {
				return nil, err
			}
			sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
		}
	}

	client, err := ssh.Dial("tcp", u.Host, &sshConfig)
	if err != nil {
		return nil, err
	}

	return &Connection{
		handler:   client,
		errorChan: nil,
		Type:      CONNECTION_SSH,
		Address:   u.Scheme + "://" + u.Host,
	}, nil
}

func (c *Connection) SetLogFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		c.Close()
		return err
	}
	f.WriteString(fmt.Sprintf("\n=== New connection to %v on %v ===\n", c.Address, time.Now().Format(time.DateTime)))
	switch c.Type {
	case CONNECTION_SSH:
		c.logFile = f
	}
	return nil
}

func (c *Connection) SendCmd(cmd string) (string, string, error) {
	switch c.Type {
	case CONNECTION_SSH:
		session, err := c.handler.(*ssh.Client).NewSession()
		if err != nil {
			return "", "", nil
		}
		defer session.Close()

		var outBuff, errBuff bytes.Buffer
		mwOut := io.MultiWriter(c.logFile, &outBuff)
		mwErr := io.MultiWriter(c.logFile, &errBuff)
		session.Stdout = mwOut
		session.Stderr = mwErr

		c.logFile.WriteString(fmt.Sprintf("%v\n", cmd))

		err = session.Run(cmd)

		if err != nil {
			if status, ok := err.(*ssh.ExitError); ok {
				c.logFile.WriteString(fmt.Sprintf("exit code (%v)\n", status.ExitStatus()))
				return outBuff.String(), errBuff.String(), fmt.Errorf("exit code (%v)", status.ExitStatus())
			} else {
				return "", "", err
			}
		} else {
			return strings.TrimSpace(outBuff.String()), strings.TrimSpace(errBuff.String()), nil
		}
	default:
		return "", "", fmt.Errorf("unsupported connection type (%v)", c.Type)
	}
}

func (c *Connection) ExecuteSequence(title string, sequence []configuration.SequenceCmd, context any) error {
	for idx, cmd := range sequence {
		fmt.Printf("Executing %v sequence step %v/%v...\n", title, idx+1, len(sequence))
		attempts := cmd.Retries + 1
		for {
			err := c.ExecuteCommand(cmd, context)
			attempts--
			if err != nil {
				if attempts == 0 {
					return err
				} else {
					fmt.Printf("Command failed (%v); will retry %v more time(s).\n", err.Error(), attempts)
				}
			}
			if cmd.Delay > 0 {
				fmt.Printf("=> Waiting for %v millisecond(s)...\n", cmd.Delay)
				time.Sleep(time.Duration(cmd.Delay) * time.Millisecond)
			}
			if err == nil {
				break
			}
		}
	}
	return nil
}

func (c *Connection) ExecuteCommand(cmd configuration.SequenceCmd, context any) error {
	if cmd.Cmd != "" {
		cmdStr, err := templating.RenderTemplateString(cmd.Cmd, context)
		if err != nil {
			return err
		}
		fmt.Printf("=> Send cmd: '%v'...\n", cmdStr)

		if err != nil {
			return err
		}
		output, _, err := c.SendCmd(cmdStr)
		if err != nil {
			return err
		}
		// if specified expected output, validate against actual
		if cmd.Output != "" {
			outputStr, err := templating.RenderTemplateString(cmd.Output, context)
			if err != nil {
				return err
			}
			if outputStr != output {
				return fmt.Errorf("cmd '%v' validation failed; expected output '%v', actual '%v'", cmdStr, outputStr, output)
			} else {
				fmt.Printf("=> Cmd output validated: %v\n", outputStr)
			}
		}
	}
	return nil
}

func (c *Connection) Close() error {
	switch c.Type {
	case CONNECTION_SSH:
		c.logFile.Sync()
		c.logFile.Close()
		return nil
	default:
		return fmt.Errorf("unsupported connection type (%v)", c.Type)
	}
}

func keyboardChallenge(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
	if name != "" {
		tui.DisplayHelpMsg(name)
	}
	if instruction != "" {
		tui.DisplayHelpMsg(instruction)
	}
	answers = make([]string, len(questions))
	for idx, question := range questions {
		var err error
		if !echos[idx] {
			answers[idx], err = tui.PromptForPassword(question)
		} else {
			answers[idx], err = tui.PromptForValue(question, "")
		}
		if err != nil {
			return nil, err
		}
	}
	return answers, nil
}

// Map values should be inline with toolchain output
var machineToArchMap = map[string]string{
	"aarch64": "armv8",
	"armv7l":  "armv7",
}

const cpuArchDiscoveryCmd = "uname -m"

func DiscoverTargetCPUarch(c Connection) (string, error) {

	cpuArch, _, err := c.SendCmd(cpuArchDiscoveryCmd)
	if err != nil {
		return "", err
	}

	result, ok := machineToArchMap[cpuArch]

	if !ok {
		return "", fmt.Errorf("unsupported CPU arch %v", cpuArch)
	}

	return result, nil
}

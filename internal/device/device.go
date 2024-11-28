// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package device

import (
	"bytes"
	"corteca/internal/tui"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const DEACTIVATE_QUAGGA_CMD = "sed -i 's#/usr/bin/vtysh#/bin/ash#' /etc/passwd"

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

const (
	authSshPasswordName  = "password"
	authSshPublicKeyName = "publicKey"
)

type Connection struct {
	// Deliberately hidden from the user
	handler   any
	errorChan <-chan error
	logFile   *os.File
	Type      int
	Address   string
}

func Connect(addr, authType, privateKeyFile, password2, logFile string) (*Connection, error) {
	var (
		conn *Connection
		err  error
	)
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	authType = strings.ToLower(authType)
	protocol := strings.ToLower(u.Scheme)
	switch protocol {
	case "ssh":
		conn, err = connectSSH(u, authType, password2, privateKeyFile)
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

func connectSSH(u *url.URL, authType, password2, keyFile string) (*Connection, error) {
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
		switch authType {
		case authSshPasswordName:
			password, err := tui.PromptForPassword(fmt.Sprintf("%s@%s's password", u.User, u.Host))
			if err != nil {
				return nil, err
			}
			sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
		case authSshPublicKeyName:
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

	isQuaggaActive, err := hasQuagga(client)
	if err != nil {
		return nil, err
	}

	if isQuaggaActive {
		if err := deactivateQuagga(client, password2); err != nil {
			return nil, err
		}

		client.Close()

		client, err = ssh.Dial("tcp", u.Host, &sshConfig)
		if err != nil {
			return nil, err
		}
	}

	return &Connection{
		handler:   client,
		errorChan: nil,
		Type:      CONNECTION_SSH,
		Address:   u.Scheme + "://" + u.Host,
	}, nil
}

func hasQuagga(client *ssh.Client) (bool, error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	var stdOut bytes.Buffer
	session.Stdout = &stdOut
	err = session.Run("ps | grep ash")

	if err != nil {
		return true, nil
	} else if strings.Contains(stdOut.String(), "ash") {
		return false, nil
	}

	return false, nil
}

func deactivateQuagga(client *ssh.Client, password2 string) error {
	var err error

	if password2 == "" {
		password2, err = tui.PromptForPassword("Enter Password2")
		if err != nil {
			return err
		}
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return err
	}

	stdInPipe, err := session.StdinPipe()
	if err != nil {
		return err
	}

	if err = session.Shell(); err != nil {
		return err
	}

	commands := []string{
		"shell",
		password2,
		DEACTIVATE_QUAGGA_CMD,
	}

	for _, cmd := range commands {
		_, err := stdInPipe.Write([]byte(cmd + "\n"))

		if err != nil {
			return fmt.Errorf("error while deactivate Quagga: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (c *Connection) SetLogFile(filename string) error {
	var output *os.File
	var err error
	switch filename {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		output, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			c.Close()
			return err
		}
	}
	output.WriteString(fmt.Sprintf("\n=== New connection to %s on %s ===\n", c.Address, time.Now().Format(time.DateTime)))
	switch c.Type {
	case CONNECTION_SSH:
		c.logFile = output
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

const cpuArchDiscoveryCmd = "uname -m"

func DiscoverTargetCPUarch(c Connection) (string, error) {

	cpuArch, _, err := c.SendCmd(cpuArchDiscoveryCmd)
	if err != nil {
		return "", err
	}
	return cpuArch, nil
}

const lcmList = "lcm list"
const grepPluginMgr = "pgrep PluginMgr"

func ContainerFrameworkType(connectionToDevice Connection) string {

	_, _, err := connectionToDevice.SendCmd(lcmList)
	if err == nil {
		return "oci"
	}
	_, _, err = connectionToDevice.SendCmd(grepPluginMgr)
	if err == nil {
		return "rootfs"
	}
	return ""
}

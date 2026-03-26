package device

import (
	"bytes"
	"corteca/internal/configuration"
	"corteca/internal/dispatcher"
	"corteca/internal/tui"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	deactivateQuaggaCmd = "sed -i 's#/usr/bin/vtysh#/bin/ash#' /etc/passwd"
	maxNumRetries       = 3
	defaultSSHPort      = 22

	authSSHPassword  = "password"
	authSSHPublicKey = "publicKey"
	cmdCPUArch        = "uname -m"
)

type SSHClient interface {
	NewSession() (*ssh.Session, error)
	Close() error
}

type SSHDevice struct {
	client   	SSHClient
	urlInfo		*url.URL
	auth        string
	password2   string
	keyFile 	string
	token       string
	log      	*Logger
}

func NewSSHDevice(endpoint configuration.Endpoint, logfile string) (*SSHDevice, error) {
	log := &Logger{}
	if logfile != "" {
		log.SetLogFile(logfile)
	}

	u, err := url.Parse(endpoint.Addr.String())
	if err != nil {
		return nil, err
	}

	if u.Port() == "" {
		u.Host = u.Host + fmt.Sprintf(":%v", defaultSSHPort)
	}

	return &SSHDevice{
		urlInfo: u,
		log: log,
		auth: endpoint.Auth,
		password2: endpoint.Password2.String(),
		token: endpoint.Token.String(),
		keyFile: endpoint.PrivateKeyFile.String(),
	}, nil
}

func (d *SSHDevice)DiscoverTargetCPUArch(dispatcher dispatcher.Dispatcher) (string, error) {
	output, err := dispatcher.ExecuteCommand(cmdCPUArch)
	if err != nil {
		return "", fmt.Errorf("failed to discover CPU architecture: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (d *SSHDevice) GetProtocol() int {
	return ConnectionSSH
}

func (d *SSHDevice) Connect() (dispatcher.Dispatcher, error) {
	

	d.log.LogFile.WriteString(fmt.Sprintf("\n=== New connection to %s on %s ===\n", d.urlInfo.Host, time.Now().Format(time.DateTime)))

	sshConfig, err := d.buildSSHConfig(d.urlInfo)
	if err != nil {
		return nil, err
	}

	if err := d.connectClient(d.urlInfo.Host, sshConfig); err != nil {
		return nil, err
	}

	if active, err := hasQuagga(d.client); err != nil {
		return nil, err
	} else if active {
		if err := deactivateQuagga(d.client, d.password2); err != nil {
			return nil, err
		}
		d.client.Close()
		if err := d.connectClient(d.urlInfo.Host, sshConfig); err != nil {
			return nil, err
		}
	}

	return dispatcher.NewSSHDispatcher(d.client.(*ssh.Client)), nil
}

func (d *SSHDevice) buildSSHConfig(u *url.URL) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            u.User.Username(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Replace with secure method
	}

	if passwd, ok := u.User.Password(); ok {
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
		return config, nil
	}

	switch d.auth {
	case authSSHPassword:
		password, err := tui.PromptForPassword(fmt.Sprintf("%s@%s's password", u.User.Username(), u.Host))
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	case authSSHPublicKey:
		keyPath := d.keyFile
		if keyPath == "" {
			return nil, fmt.Errorf("missing private key file for public key authentication")
		}
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	return config, nil
}

func (d *SSHDevice) connectClient(host string, config *ssh.ClientConfig) error {
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return err
	}
	d.client = client
	return nil
}

func hasQuagga(client SSHClient) (bool, error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output

	if err := session.Run("ps | grep ash"); err != nil {
		return true, nil
	}

	return !strings.Contains(output.String(), "ash"), nil
}

func deactivateQuagga(client SSHClient, password2 string) error {
	if password2 == "" {
		var err error
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

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return err
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	commands := []string{"shell", password2, deactivateQuaggaCmd}
	for _, cmd := range commands {
		if _, err := stdin.Write([]byte(cmd + "\n")); err != nil {
			return fmt.Errorf("failed to run command %q: %w", cmd, err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *SSHDevice) Close() {
	if d.log != nil && d.log.LogFile != nil {
		d.log.LogFile.Sync()
		d.log.LogFile.Close()
	}
}

package ssh

import (
	"bytes"
	"context"
	"corteca/internal/configuration"
	"corteca/internal/device"
	"corteca/internal/tui"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	stdssh "golang.org/x/crypto/ssh"
)

// syncWriter wraps an io.Writer with a mutex so that concurrent writes from
// the SSH session's stdout and stderr goroutines are serialised safely.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

const (
	deactivateQuaggaCmd = "sed -i 's#/usr/bin/vtysh#/bin/ash#' /etc/passwd"
	maxNumRetries       = 3
	defaultSSHPort      = "22"

	authSSHPassword  = "password"
	authSSHPublicKey = "publicKey"
	cmdCPUArch       = "uname -m"
)

type SSHDevice struct {
	client *stdssh.Client
	log    *syncWriter
}

func init() {
	device.RegisterDeviceType("ssh", NewSSHDevice)
}

func NewSSHDevice(c *configuration.DeviceConfig, log io.Writer) (device.Device, error) {
	d := SSHDevice{
		log: &syncWriter{w: log},
	}
	var (
		err       error
		sshconfig configuration.SSHClientEndpoint
	)
	c.Decode(&sshconfig)

	if err = d.connectSSHClient(&sshconfig); err != nil {
		return nil, err
	}

	passwd2 := sshconfig.Password2.String()

	if active, err := hasQuagga(d.client); err != nil {
		return nil, err
	} else if active && len(passwd2) > 0 {
		if err := deactivateQuagga(d.client, passwd2); err != nil {
			return nil, err
		}
		d.client.Close()
		tui.LogNormal("Need to reconnect...")
		return NewSSHDevice(c, log)
	}
	return &d, nil
}

func (d *SSHDevice) BeginSequence() error {
	return nil
}

func (d *SSHDevice) ExecuteCommand(ctx context.Context, cmd *configuration.SequenceCmd) (any, error) {
	// interpret cmd.params as SSHParams (i.e. array of strings)
	var params struct {
		Params []configuration.TemplateField `yaml:"params"`
	}
	if err := cmd.Decode(&params); err != nil {
		return nil, fmt.Errorf("incompatible command parameters specified; array of strings expected")
	}

	// render params in case they use templates
	paramsRendered := make([]string, len(params.Params))
	for i := 0; i < len(params.Params); i++ {
		paramsRendered[i] = params.Params[i].String()
	}

	// concatenate into a single string
	cmdString := fmt.Sprintf("%s %s", cmd.Cmd.String(), strings.Join(paramsRendered, " "))

	return d.executeCommandString(ctx, cmdString)
}

func (d *SSHDevice) executeCommandString(ctx context.Context, cmd string) (any, error) {
	session, err := d.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("cannot start SSH command session: %w", err)
	}
	defer session.Close()
	output := bytes.NewBuffer(make([]byte, 0, 512))
	session.Stdout = io.MultiWriter(d.log, output)
	session.Stderr = d.log

	// start command
	if err = session.Start(cmd); err != nil {
		return nil, err
	}

	// create channel to synchronize
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case err := <-done:
		var exitError *stdssh.ExitError
		if errors.As(err, &exitError) {
			return output, fmt.Errorf("exit code (%d)", exitError.ExitStatus())
		} else {
			return nil, err
		}

	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return nil, ctx.Err()
	}
}

func (d *SSHDevice) EndSequence() error {
	return nil
}

func (d *SSHDevice) GetProtocol() string {
	return "ssh"
}

func (d *SSHDevice) connectSSHClient(sshconfig *configuration.SSHClientEndpoint) error {
	u, err := url.Parse(sshconfig.Addr.String())
	if err != nil {
		return err
	}
	if u.Port() == "" {
		u.Host = net.JoinHostPort(u.Host, defaultSSHPort)
	}

	// Determine username: explicit Username field takes priority over the URL.
	username := u.User.Username()
	if explicitUser := sshconfig.Username.String(); len(explicitUser) > 0 {
		username = explicitUser
	}

	config := &stdssh.ClientConfig{
		User:            username,
		HostKeyCallback: stdssh.InsecureIgnoreHostKey(), // TODO: Replace with secure method
		Auth:            make([]stdssh.AuthMethod, 0, 2),
	}

	// add keyfile, if present
	keyPath := sshconfig.PrivateKeyFile.String()
	if len(keyPath) > 0 {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("cannot read private key file %s: %w", keyPath, err)
		}
		signer, err := stdssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("cannot parse private key file %s: %w", keyPath, err)
		}
		config.Auth = append(config.Auth, stdssh.PublicKeys(signer))
	}

	// add password, if present
	passwd := sshconfig.Password.String()
	passwdPresent := len(passwd) > 0
	if passwdPresent {
		config.Auth = append(config.Auth, stdssh.Password(passwd))
	} else if passwd, passwdPresent = u.User.Password(); passwdPresent {
		config.Auth = append(config.Auth, stdssh.Password(passwd))
	}

	// add prompt for password if no other methods exist
	if len(config.Auth) == 0 {
		// FIXME:
		// the below results in always asking for a password even if the SSH server is not asking for one
		// should use something like: config.Auth = append(config.Auth, stdssh.KeyboardInteractive(...))
		passwd, err := tui.PromptForPassword(fmt.Sprintf("%s@%s's password", u.User.Username(), u.Host))
		if err != nil {
			return err
		}
		config.Auth = append(config.Auth, stdssh.Password(passwd))
	}

	// connect to stdssh server
	d.client, err = stdssh.Dial("tcp", u.Host, config)
	if err != nil {
		fmt.Fprintf(d.log, "\n=== New connection to %s at %s ===\n", u.Host, time.Now().Format(time.DateTime))
	}
	return err
}

func hasQuagga(client *stdssh.Client) (bool, error) {
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

func deactivateQuagga(client *stdssh.Client, password2 string) error {
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

	if err := session.RequestPty("xterm", 80, 40, stdssh.TerminalModes{}); err != nil {
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
	d.client.Close()
}

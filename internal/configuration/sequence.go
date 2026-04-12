package configuration

import (
	"context"
	"corteca/internal/tui"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

var cmdRegularExpression *regexp.Regexp

func init() {
	cmdRegularExpression = regexp.MustCompile(`^\s*\$\((.+)\)\s*$`)
}

const (
	DefaultMaxTimeout = 5 * time.Minute
)

var (
	ErrAbortSequence = errors.New("fatal error")
)

type SequenceMap map[string]Sequence

type Sequence []SequenceCmd

type SequenceCmd struct {
	Cmd           TemplateField `yaml:"cmd"`
	Delay         time.Duration `yaml:"duration,omitempty"`
	Timeout       time.Duration `yaml:"timeout,omitempty"`
	Retries       uint          `yaml:"retries,omitempty"`
	IgnoreFailure *bool         `yaml:"ignoreFailure,omitempty"`
	raw           *yaml.Node
}

func parseDuration(value string, defaultvalue time.Duration) (time.Duration, error) {
	if len(value) > 0 {
		return time.ParseDuration(value)
	} else {
		return defaultvalue, nil
	}
}

func (cmd *SequenceCmd) UnmarshalYAML(value *yaml.Node) error {
	cmd.raw = value
	var proxy struct {
		Cmd           TemplateField `yaml:"cmd"`
		Delay         string        `yaml:"duration"`
		Timeout       string        `yaml:"timeout"`
		Retries       uint          `yaml:"retries"`
		IgnoreFailure *bool         `yaml:"ignoreFailure"`
	}
	if err := value.Decode(&proxy); err != nil {
		return err
	}
	cmd.Cmd = proxy.Cmd
	if d, err := parseDuration(proxy.Delay, 0); err != nil {
		return err
	} else {
		cmd.Delay = d
	}
	if d, err := parseDuration(proxy.Timeout, 0); err != nil {
		return err
	} else {
		cmd.Timeout = d
	}
	cmd.Retries = proxy.Retries
	cmd.IgnoreFailure = proxy.IgnoreFailure
	return nil
}

func (cmd SequenceCmd) MarshalYAML() (any, error) {
	return cmd.raw, nil
}

func (cmd *SequenceCmd) Decode(v any) error {
	return cmd.raw.Decode(v)
}

type CommandExecutor interface {
	BeginSequence() error
	ExecuteCommand(context.Context, *SequenceCmd) (any, error)
	EndSequence() error
}

func (sm *SequenceMap) Execute(executor CommandExecutor, seqName string, skipinit bool) error {
	seq, found := (*sm)[seqName]
	if !found {
		return fmt.Errorf("sequence '%s' was not found", seqName)
	}
	tui.LogNormal("Executing sequence '%s'", seqName)
	if !skipinit {
		if err := executor.BeginSequence(); err != nil {
			return fmt.Errorf("failed to initialize sequence: %w", err)
		}
	}
	for idx, step := range seq {
		if refSeqName, found := findRefToSequence(step.Cmd.String()); found {
			if err := sm.Execute(executor, refSeqName, true); err != nil {
				return err
			}
			continue
		}
		res, err := executeStep(&step, executor)
		if err != nil {
			return fmt.Errorf("sequence '%s' failed at step %d: %w", seqName, idx+1, err)
		}
		// TODO: provide option to suppress output
		tui.SetOutputColor(tui.CBlue, os.Stdout)
		enc := yaml.NewEncoder(os.Stdout)
		enc.Encode(res)
		tui.ResetOutputColor(os.Stdout)
	}
	if !skipinit {
		if err := executor.EndSequence(); err != nil {
			return fmt.Errorf("failed to shutdown sequence: %w", err)
		}
	}
	return nil
}

func findRefToSequence(expr string) (string, bool) {
	if cmdRefRegex := cmdRegularExpression.FindStringSubmatch(expr); len(cmdRefRegex) == 2 {
		return cmdRefRegex[1], true
	} else {
		return "", false
	}
}

func createContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultMaxTimeout
	}
	return context.WithTimeout(context.Background(), timeout)
}

func executeStep(step *SequenceCmd, executor CommandExecutor) (any, error) {
	attempts := step.Retries + 1
	var (
		res any
		err error
	)
	for attempts > 0 {
		if step.Delay > 0 {
			tui.LogNormal("Waiting for %s", step.Delay.String())
			time.Sleep(step.Delay)
		}
		ctx, cancel := createContext(step.Timeout)
		defer cancel()
		res, err = executor.ExecuteCommand(ctx, step)
		attempts--
		if err != nil {
			if step.IgnoreFailure != nil && (*step.IgnoreFailure) {
				return res, nil
			} else {
				tui.LogError("Command failed: %s", err.Error())
				if attempts > 0 {
					tui.LogNormal("Will retry %d more time(s)", attempts)
				} else {
					return res, err
				}
			}
		} else {
			break
		}
	}
	return res, nil
}

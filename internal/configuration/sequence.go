package configuration

import (
	"corteca/internal/dispatcher"
	"corteca/internal/tui"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type SequenceMap map[string]Sequence

func (sm *SequenceMap) UnmarshalYAML(data *yaml.Node) error {
	sequences := make(map[string]struct {
		Type  string      `yaml:"type"`
		Steps []yaml.Node `yaml:"steps"`
	})
	if err := data.Decode(sequences); err != nil {
		fmt.Println("Decode error")
		return err
	}

	for seqName, rawSeq := range sequences {
		switch rawSeq.Type {
		case "ssh":
			steps := make([]StringCmd, len(rawSeq.Steps))

			for i, node := range rawSeq.Steps {
				var step StringCmd
				if err := node.Decode(&step); err != nil {
					return fmt.Errorf("error decoding step %d in sequence %q: %w", i, seqName, err)
				}
				steps[i] = step
			}

			(*sm)[seqName] = NewStringSequence(rawSeq.Type, steps)
		case "cwmp":
			steps := make([]CwmpCmd, len(rawSeq.Steps))

			for i, node := range rawSeq.Steps {
				var step CwmpCmd
				if err := node.Decode(&step); err != nil {
					return fmt.Errorf("error decoding step %d in sequence %q: %w", i, seqName, err)
				}
				steps[i] = step
			}

			(*sm)[seqName] = NewCwmpSequence(rawSeq.Type, steps)
		}
	}
	return nil
}

func (sm *SequenceMap) Execute(dispatcher dispatcher.Dispatcher, seq string) error {
	selectedSequence, found := (*sm)[seq]
	if !found {
		return fmt.Errorf("sequence '%s' was not found", seq)
	}
	tui.LogNormal("Executing sequence %s", seq)
	return selectedSequence.Execute(dispatcher, *sm)
}

type Sequence interface {
	GetType() string
	Execute(dispatcher.Dispatcher, SequenceMap) error
}

func NewStringSequence(tp string, steps []StringCmd) *StringSequence {
	return &StringSequence{tp, steps}
}

type StringSequence struct {
	Type  string
	Steps []StringCmd
}

func (sq *StringSequence) GetType() string {
	return sq.Type
}

func (sq *StringSequence) Execute(dispathcer dispatcher.Dispatcher, sm SequenceMap) error {
	for idx, step := range sq.Steps {
		if seqName, found := findRefToSequence(step.Cmd.String()); found {
			err := sm.Execute(dispathcer, seqName)
			if err != nil {
				return fmt.Errorf("reference sequence failed at step %d: %w", idx+1, err)
			}
		} else {

			if out, err := step.Execute(dispathcer); err != nil {
				return fmt.Errorf("sequence failed at step %d: %w", idx+1, err)
			} else {
				tui.LogOutData(out)
			}
		}
	}
	return nil
}

func findRefToSequence(seqCmd string) (string, bool) {
	if cmdRefRegex := cmdRegularExpression.FindStringSubmatch(seqCmd); len(cmdRefRegex) == 2 {
		return cmdRefRegex[1], true
	} else {
		return "", false
	}
}

type SequenceCmd interface {
	Execute(dispatcher.Dispatcher) error
}

type CwmpSequence struct {
	Type  string
	Steps []CwmpCmd
}

func NewCwmpSequence(tp string, steps []CwmpCmd) *CwmpSequence {
	return &CwmpSequence{tp, steps}
}

func (sq *CwmpSequence) GetType() string {
	return sq.Type
}

func (sq *CwmpSequence) Execute(dispathcer dispatcher.Dispatcher, sm SequenceMap) error {
	for idx, step := range sq.Steps {
		// Cmd, Operation and PrintFormat should be lowercase in order for the code to work
		step.Cmd.RawTemplate = strings.ToLower(step.Cmd.String())
		step.Operation = strings.ToLower(step.Operation)
		step.PrintFormat = strings.ToLower(step.PrintFormat)
		if seqName, found := findRefToSequence(step.Cmd.String()); found {
			err := sm.Execute(dispathcer, seqName)
			if err != nil {
				return fmt.Errorf("reference sequence failed at step %d: %w", idx+1, err)
			}
		} else {
			if cmdResults, err := step.Execute(dispathcer); err != nil {
				return fmt.Errorf("sequence failed at step %d: %w", idx+1, err)
			} else {
				if cmdResults != "" {
					tui.LogOutData(cmdResults)
				}
			}
		}
	}
	return nil
}

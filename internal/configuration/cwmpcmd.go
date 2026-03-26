package configuration

import (
	"corteca/internal/cwmp/messages"
	"corteca/internal/dispatcher"
	"corteca/internal/tui"

	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type CwmpCmd struct {
	Cmd             TemplateField         `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Operation       string                `yaml:"operation,omitempty" json:"operation,omitempty"`
	ParameterList   []ParameterListValues `yaml:"parameterList,omitempty"`
	ParameterNames  []string              `yaml:"parameterNames,omitempty"`
	ParameterKey    string                `yaml:"parameterKey,omitempty"`
	Url             TemplateField         `yaml:"url,omitempty" json:"url,omitempty"`
	Username        TemplateField         `yaml:"username,omitempty" json:"username,omitempty"`
	Password        TemplateField         `yaml:"password,omitempty" json:"password,omitempty"`
	UUID            TemplateField         `yaml:"uuid,omitempty" json:"uuid,omitempty"`
	Version         TemplateField         `yaml:"version,omitempty" json:"version,omitempty"`
	ExecutionEnvRef TemplateField         `yaml:"executionenvref,omitempty" json:"executionenvref,omitempty"`
	Delay           uint                  `yaml:"delay,omitempty" json:"delay,omitempty"`
	Retries         uint                  `yaml:"retries,omitempty" json:"retries,omitempty"`
	IgnoreFailure   bool                  `yaml:"ignoreFailure,omitempty" json:"ignoreFailure,omitempty"`
	ParameterPath   string                `yaml:"parameterPath,omitempty" json:"path,omitempty"`
	NextLevel       bool                  `yaml:"nextLevel,omitempty" json:"nextLevel,omitempty"`
	PrintFormat     string                `yaml:"printFormat,omitempty" json:"printFormat,omitempty"`
}

type ParameterListValues struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	Type  string `yaml:"type"`
}

func (sqCmd *CwmpCmd) Execute(dispatcher dispatcher.Dispatcher) (cmdResults string, err error) {
	attempts := sqCmd.Retries + 1

	for attempts > 0 {
		if sqCmd.Cmd.String() != "" {
			cmdResults, err = sqCmd.executeCommand(dispatcher)
		}

		attempts--

		if !sqCmd.IgnoreFailure && err != nil {
			if attempts == 0 {
				return "", err
			} else {
				tui.LogError("Command failed (%s); will retry %d more time(s).", err.Error(), attempts)
			}
		}

		if sqCmd.Delay > 0 {
			tui.LogNormal("=> Waiting for %d millisecond(s)...", sqCmd.Delay)
			time.Sleep(time.Duration(sqCmd.Delay) * time.Millisecond)
		}
		if err == nil {
			break
		}

	}

	return cmdResults, nil
}

func (sqCmd *CwmpCmd) executeCommand(dispatcher dispatcher.Dispatcher) (string, error) {
	var msg messages.Message
	dispatcher.SetPrintFormat(sqCmd.PrintFormat)
	switch sqCmd.Cmd.String() {
	case "change_du_state":
		tui.LogNormal("=> Cmd:\n    ChangeDUState\n=> Operation:\n    %s", sqCmd.Operation)
		dustate := messages.NewChangeDUState()
		var operation messages.DeploymentUnitOperationStruct
		dustate.CommandKey = strconv.FormatInt(rand.Int63n(9000000000)+1000000000, 10)
		dustate.OperationType = sqCmd.Operation
		operation.UUID = sqCmd.UUID.String()
		operation.URL = sqCmd.Url.String()
		operation.ExecutionEnvRef = sqCmd.ExecutionEnvRef.String()
		operation.Password = sqCmd.Password.String()
		operation.Username = sqCmd.Username.String()
		operation.Version = sqCmd.Version.String()
		dustate.Operation = operation
		msg = dustate
	case "get_parameter_names":
		if sqCmd.PrintFormat != "json" {
			tui.LogNormal("=> Cmd:\n    GetParameterNames\n=> Parameter:\n    %s", sqCmd.ParameterPath)
		}
		getParamNames := messages.NewGetParameterNames()
		getParamNames.ParameterPath = sqCmd.ParameterPath
		getParamNames.NextLevel = sqCmd.NextLevel
		msg = getParamNames
	case "get_parameter_values":
		if len(sqCmd.ParameterNames) == 0 {
			return "", fmt.Errorf("empty parameter names list %v", sqCmd.ParameterNames)
		}
		if sqCmd.PrintFormat != "json" {
			tui.LogNormal("=> Cmd:\n    GetParameterValues\n=> Parameter(s):")
			for _, parameterName := range sqCmd.ParameterNames {
				tui.LogNormal("    %s", parameterName)
			}
		}

		getParamValues := messages.NewGetParameterValues()
		getParamValues.ParameterNames = sqCmd.ParameterNames
		getParamValues.PrintFormat = sqCmd.PrintFormat
		msg = getParamValues
	case "set_parameter_values":
		tui.LogNormal("=> Cmd:\n    SetParameterValues\n=> Parameter(s) & New Value(s):")
		setParamValues := messages.NewSetParameterValues()
		for _, parameter := range sqCmd.ParameterList {
			tui.LogNormal("    %s: %v", parameter.Name, parameter.Value)
			paramval := messages.ParameterVal{}
			paramval.Name = parameter.Name
			paramval.Value.Value = parameter.Value
			paramval.Value.Type = parameter.Type
			setParamValues.ParameterList = append(setParamValues.ParameterList, paramval)
		}
		setParamValues.ParameterKey = sqCmd.ParameterKey
		msg = setParamValues
	default:
		tui.LogNormal("=> Cmd: '%s'", sqCmd.Cmd.String())

		err := fmt.Errorf("rpc %s not implemented", sqCmd.Cmd.String())
		return "", err
	}

	cmdResults, err := dispatcher.ExecuteCommand(msg)
	return cmdResults, err
}

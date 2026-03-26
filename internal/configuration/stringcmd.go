package configuration

import (
    "corteca/internal/dispatcher"
    "time"
    "fmt"
)

type StringCmd struct {
    Cmd           TemplateField `yaml:"cmd,omitempty"`
    Delay         uint          `yaml:"delay,omitempty"`
    Retries       uint          `yaml:"retries,omitempty"`
    IgnoreFailure bool          `yaml:"ignoreFailure,omitempty"`
}

func (sqCmd *StringCmd) Execute(dispatcher dispatcher.Dispatcher) (out string, err error) {
    attempts := sqCmd.Retries + 1

    for attempts > 0 {
        out, err = sqCmd.executeCommand(dispatcher)

        attempts--

        if !sqCmd.IgnoreFailure && err != nil {
            if attempts == 0 {
                    return "", err
                } else {
                    fmt.Printf("Command failed (%s); will retry %d more time(s).\n", err.Error(), attempts)
                }
        }

        if sqCmd.Delay > 0 {
            fmt.Printf("=> Waiting for %d millisecond(s)...\n", sqCmd.Delay)
            time.Sleep(time.Duration(sqCmd.Delay) * time.Millisecond)
        }
        if err == nil {
                break
        }
    }

    return out, nil
}

func (sqCmd *StringCmd) executeCommand(dispatcher dispatcher.Dispatcher) (string, error) {
    
    cmdStr := sqCmd.Cmd.String()
    
    fmt.Printf("=> Send cmd: '%s'...\n", cmdStr)
    out, err := dispatcher.ExecuteCommand(cmdStr)
    
    return out, err
}

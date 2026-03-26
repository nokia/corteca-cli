package dispatcher

import (
	"bytes"
	"corteca/internal/cwmp/messages"
	"corteca/internal/cwmp/models"
	"encoding/json"
	"fmt"
	"io"
    "strings"

	"golang.org/x/crypto/ssh"
)

type Dispatcher interface {
    ExecuteCommand(any) (string, error)
    SetPrintFormat(string)
}

type SSHDispatcher struct {
    client *ssh.Client
}

func NewSSHDispatcher(client *ssh.Client) *SSHDispatcher {
    return &SSHDispatcher{client: client}
}

func (ssh_dispatcher *SSHDispatcher) SetPrintFormat(format string) {
}

func (ssd_dispatcher *SSHDispatcher) ExecuteCommand(cmd any) (string, error) {
    session, err := ssd_dispatcher.client.NewSession()
    if err != nil {
        return "", nil
    }
    defer session.Close()

    var outBuff bytes.Buffer
    mwOut := io.MultiWriter(&outBuff)
    session.Stdout = mwOut
    session.Stderr = mwOut

    err = session.Run(cmd.(string))

    if err != nil {
        if status, ok := err.(*ssh.ExitError); ok {
            return outBuff.String(), fmt.Errorf("exit code (%v)", status.ExitStatus())
        } else {
            return outBuff.String(), err
        }
    } else {
        return outBuff.String(), nil
    }
}

type CWMPDispatcher struct {
    taskChannel     chan messages.Message
    resultChannel   chan *models.ResultsMessage
    printFormat     string
}

func NewCWMPDispatcher(taskChan chan messages.Message, resultChannel chan *models.ResultsMessage) *CWMPDispatcher {
    return &CWMPDispatcher{taskChannel: taskChan, resultChannel: resultChannel}
}

func (d *CWMPDispatcher) SetPrintFormat(format string) {
    d.printFormat = format
}

func formatMessageOutput(result *models.ResultsMessage, printFormat string) (string, error) {
    var output string

    if result.Message == nil {
        return "", fmt.Errorf("empty message")
    }

    msg := result.Message

    switch msg.GetName() {
        case "Inform":
        case "GetParameterNamesResponse":
            var builder strings.Builder
            if printFormat == "json" {
                parameterValuesList, err := json.Marshal(msg.(*messages.GetParameterNamesResponse).ParameterList.Parameters)
                if err != nil {
                    return "", err
                }

                output = string(parameterValuesList)
            } else {
                for _, parameter := range msg.(*messages.GetParameterNamesResponse).ParameterList.Parameters {
                    builder.WriteString(fmt.Sprintf("- Name: %-s\n", parameter.Name))
                    builder.WriteString(fmt.Sprintf("  Writable: %-v\n", parameter.Writable))
                }
                output = builder.String()
            }
        case "GetParameterValuesResponse":
            var resultStr strings.Builder
            
            if printFormat == "json" {
                parameterValuesList, err := json.Marshal(msg.(*messages.GetParameterValuesResponse).ParameterList)
                if err != nil {
                    return "", err
                }
                output = string(parameterValuesList)
            } else {
                resultStr.WriteString("\n************** Parameter(s) Value(s) **************\n")
                for _, parameter := range msg.(*messages.GetParameterValuesResponse).ParameterList {
                    resultStr.WriteString(fmt.Sprintf("%s: %s\n", parameter.Name, parameter.Value))
                }
                resultStr.WriteString("***************************************************\n")
                output = resultStr.String()
            }
        case "SetParameterValuesResponse":
            if result.Code == 0 {
                output = "All parameters changes have been validated and applied"
            } else {
                output = "All Parameter changes have been validated and committed, but some or all are not yet applied (e.g A reboot is required before the new values are applied)"
            }
        case "ChangeDUStateResponse":
        case "DUStateChangeComplete":
            ducomplete := msg.(*messages.DUStateChangeComplete)
            output = ducomplete.Fault.FaultString
        case "Fault":
            output = result.Message.(*messages.Fault).MsgFaultString
        default:
            output = "internal error"
        }

        return output, nil
}

func (d *CWMPDispatcher) ExecuteCommand(cmd any) (string, error) {
    task, ok := cmd.(messages.Message)
    
    if ok {
        //Send task to cwmp server
        d.taskChannel <- task
        //wait for results
        result := <-d.resultChannel
        
        output , err := formatMessageOutput(result, d.printFormat)
        
        if result.Code != 0 || err != nil {
            return "", fmt.Errorf("task \"%s\" with error \"%s\" (cd: %d)", task.GetName(), output, result.Code)
        }
        return output, nil
    } else {
        return "", fmt.Errorf("cmd is not a valid cwmp task")
    }
}

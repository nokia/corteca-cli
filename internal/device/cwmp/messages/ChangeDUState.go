package messages

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"encoding/xml"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type ChangeDUState struct {
	XMLName    xml.Name                    `xml:"ChangeDUState" yaml:"-"`
	CommandKey configuration.TemplateField `yaml:"CommandKey"`
	Operations DUOperationStruct           `yaml:"Operations"`
}

func (cdu ChangeDUState) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name, "ChangeDUState")
	type Alias ChangeDUState
	return enc.EncodeElement(Alias(cdu), start)
}

type DUOperationStruct struct {
	Op []DUOperation
}

func (ops *DUOperationStruct) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			op, err := decodeOpXML(dec, &tok)
			if err != nil {
				return err
			}
			ops.Op = append(ops.Op, op)
		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

func decodeOpXML(dec *xml.Decoder, start *xml.StartElement) (DUOperation, error) {
	switch start.Name.Local {
	case InstallOpStruct{}.GetOpType():
		op := InstallOpStruct{}
		return op, dec.DecodeElement(&op, start)
	case UpdateOpStruct{}.GetOpType():
		op := UpdateOpStruct{}
		return op, dec.DecodeElement(&op, start)
	case UninstallOpStruct{}.GetOpType():
		op := UninstallOpStruct{}
		return op, dec.DecodeElement(&op, start)
	default:
		return nil, fmt.Errorf("unknown optype '%s'", start.Name.Local)
	}
}

type DUOperation interface {
	GetOpType() string
}

func (ops DUOperationStruct) MarshalYAML() (any, error) {
	nodes := make([]yaml.Node, len(ops.Op))
	for i, op := range ops.Op {
		if err := nodes[i].Encode(op); err != nil {
			return nil, err
		}
		nodes[i].Tag = fmt.Sprintf("!%s", op.GetOpType())
	}
	return nodes, nil
}

func (ops *DUOperationStruct) UnmarshalYAML(value *yaml.Node) error {
	proxy := []yaml.Node{}
	if err := value.Decode(&proxy); err != nil {
		return err
	}
	for _, node := range proxy {
		op, err := decodeOpYAML(&node)
		if err != nil {
			return err
		}
		ops.Op = append(ops.Op, op)
	}
	return nil
}

func decodeOpYAML(node *yaml.Node) (DUOperation, error) {
	typeTag, _ := strings.CutPrefix(node.Tag, "!")
	switch typeTag {
	case InstallOpStruct{}.GetOpType():
		op := InstallOpStruct{}
		return op, node.Decode(&op)
	case UpdateOpStruct{}.GetOpType():
		op := UpdateOpStruct{}
		return op, node.Decode(&op)
	case UninstallOpStruct{}.GetOpType():
		op := UninstallOpStruct{}
		return op, node.Decode(&op)
	default:
		return nil, fmt.Errorf("unknown optype '%s'", typeTag)
	}
}

type InstallOpStruct struct {
	XMLName         xml.Name                    `xml:"InstallOpStruct" yaml:"-"`
	URL             configuration.TemplateField `xml:"URL" yaml:"URL"`
	UUID            configuration.TemplateField `xml:"UUID" yaml:"UUID"`
	Username        configuration.TemplateField `xml:"Username,omitempty" yaml:"Username,omitempty"`
	Password        configuration.TemplateField `xml:"Password,omitempty" yaml:"Password,omitempty"`
	ExecutionEnvRef configuration.TemplateField `xml:"ExecutionEnvRef" yaml:"ExecutionEnvRef"`
}

func (op InstallOpStruct) GetOpType() string { return "InstallOpStruct" }

type UpdateOpStruct struct {
	XMLName  xml.Name                    `xml:"UpdateOpStruct" yaml:"-"`
	UUID     configuration.TemplateField `xml:"UUID" yaml:"UUID"`
	Version  configuration.TemplateField `xml:"Version" yaml:"Version"`
	URL      configuration.TemplateField `xml:"URL" yaml:"URL"`
	Username configuration.TemplateField `xml:"Username,omitempty" yaml:"Username,omitempty"`
	Password configuration.TemplateField `xml:"Password,omitempty" yaml:"Password,omitempty"`
}

func (op UpdateOpStruct) GetOpType() string { return "UpdateOpStruct" }

type UninstallOpStruct struct {
	XMLName         xml.Name                    `xml:"UninstallOpStruct" yaml:"-"`
	UUID            configuration.TemplateField `xml:"UUID" yaml:"UUID"`
	Version         configuration.TemplateField `xml:"Version" yaml:"Version"`
	ExecutionEnvRef configuration.TemplateField `xml:"ExecutionEnvRef" yaml:"ExecutionEnvRef"`
}

func (op UninstallOpStruct) GetOpType() string { return "UninstallOpStruct" }

func (m ChangeDUState) GetName() string { return "ChangeDUState" }
func (m ChangeDUState) ValidateResponse(msg Message) error {
	if resp, ok := msg.(DUStateChangeComplete); ok {
		for i := 0; i < len(resp.Results); i++ {
			if resp.Results[i].Fault.FaultCode != 0 {
				return fmt.Errorf("error(s) occured in DU operation(s)")
			}
		}
		return nil
	} else {
		return ExpectMessage[ChangeDUStateResponse](msg)
	}
}
func (m ChangeDUState) Match(msg Message) bool {
	if r, ok := msg.(DUStateChangeComplete); ok {
		return r.CommandKey == m.CommandKey.String()
	} else {
		return false
	}
}
func (m ChangeDUState) GenerateResponse() Message {
	return ChangeDUStateResponse{}
}

type ChangeDUStateResponse struct {
	XMLName xml.Name `xml:"ChangeDUStateResponse"`
}

func (m ChangeDUStateResponse) GetName() string { return "ChangeDUStateResponse" }

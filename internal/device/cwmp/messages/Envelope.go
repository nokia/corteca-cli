package messages

import (
	"encoding/xml"
	"fmt"
	"io"
)

func ParseEnvelopeXML(input io.Reader) (*Envelope, error) {
	env := Envelope{}
	enc := xml.NewDecoder(input)
	return &env, enc.Decode(&env)
}

type Envelope struct {
	XMLName xml.Name        `xml:"Envelope"`
	Header  *EnvelopeHeader `xml:",omitempty"`
	Body    EnvelopeBody
}

func (e Envelope) GetID() string {
	if e.Header != nil {
		return e.Header.ID.Value
	} else {
		return ""
	}
}

func (e Envelope) GetBody() []Message {
	return e.Body.Messages
}

func NewEnvelope(id string, msg ...Message) Envelope {
	env := Envelope{
		Body: EnvelopeBody{
			Messages: msg,
		},
	}
	if len(id) > 0 {
		env.Header = &EnvelopeHeader{
			ID: IDStruct{Value: id, MustUnderstand: "1"},
		}
	}
	return env
}

func (e Envelope) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixSoapEnv(&start.Name)
	start.Attr = append(start.Attr,
		XmlAttr("xmlns:soap-env", "http://schemas.xmlsoap.org/soap/envelope/"),
		XmlAttr("xmlns:soap-enc", "http://schemas.xmlsoap.org/soap/encoding/"),
		XmlAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema"),
		XmlAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance"),
		XmlAttr("xmlns:cwmp", "urn:dslforum-org:cwmp-1-0"),
	)
	type Alias Envelope
	return enc.EncodeElement(Alias(e), start)
}

type EnvelopeHeader struct {
	ID IDStruct `xml:"ID"`
}

func (eh EnvelopeHeader) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixSoapEnv(&start.Name)
	type Alias EnvelopeHeader
	return enc.EncodeElement(Alias(eh), start)
}

type IDStruct struct {
	MustUnderstand string `xml:"soap-env:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

func (id IDStruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixCwmp(&start.Name)
	type Alias IDStruct
	return enc.EncodeElement(Alias(id), start)
}

type EnvelopeBody struct {
	Messages []Message
}

func (eb EnvelopeBody) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	PrefixSoapEnv(&start.Name)
	type Alias EnvelopeBody
	return enc.EncodeElement(Alias(eb), start)
}

func (eb *EnvelopeBody) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			var msg Message
			switch tok.Name.Local {
			case Inform{}.GetName():
				var m Inform
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case InformResponse{}.GetName():
				var m InformResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case ChangeDUState{}.GetName():
				var m ChangeDUState
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case ChangeDUStateResponse{}.GetName():
				var m ChangeDUStateResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case DUStateChangeComplete{}.GetName():
				var m DUStateChangeComplete
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case DUStateChangeCompleteResponse{}.GetName():
				var m DUStateChangeCompleteResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case Fault{}.GetName():
				var m Fault
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case GetParameterNames{}.GetName():
				var m GetParameterNames
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case GetParameterNamesResponse{}.GetName():
				var m GetParameterNamesResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case GetParameterValues{}.GetName():
				var m GetParameterValues
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case GetParameterValuesResponse{}.GetName():
				var m GetParameterValuesResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case SetParameterValues{}.GetName():
				var m SetParameterValues
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			case SetParameterValuesResponse{}.GetName():
				var m SetParameterValuesResponse
				if err := dec.DecodeElement(&m, &tok); err != nil {
					return err
				}
				msg = m
			default:
				return fmt.Errorf("unknown RPC '%s'", tok.Name.Local)
			}
			eb.Messages = append(eb.Messages, msg)
		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

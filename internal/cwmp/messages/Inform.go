package messages

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/beevik/etree"
)

// Inform tr069 inform (heartbeat)
type Inform struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Manufacturer string            `json:"manufacturer"`
	OUI          string            `json:"oui"`
	ProductClass string            `json:"productClass"`
	Sn           string            `json:"sn"`
	Events       map[string]string `json:"events"`
	MaxEnvelopes int               `json:"maxEnvelopes"`
	CurrentTime  string            `json:"currentTime"`
	RetryCount   int               `json:"retryCount"`
	Params       map[string]string `json:"params"`
}

type informBodyStruct struct {
	Body informStruct `xml:"cwmp:Inform"`
}

type informStruct struct {
	DeviceID     deviceIDStruct      `xml:"DeviceId"`
	Event        EventStruct         `xml:"Event"`
	MaxEnvelopes NodeStruct          `xml:"MaxEnvelopes"`
	CurrentTime  NodeStruct          `xml:"CurrentTime"`
	RetryCount   NodeStruct          `xml:"RetryCount"`
	Params       ParameterListStruct `xml:"ParameterList"`
}

type deviceIDStruct struct {
	Type         string     `xml:"xsi:type,attr"`
	Manufacturer NodeStruct `xml:"Manufacturer"`
	OUI          NodeStruct `xml:"OUI"`
	ProductClass NodeStruct `xml:"ProductClass"`
	SerialNumber NodeStruct `xml:"SerialNumber"`
}

// NewInform create a inform messages
func NewInform() *Inform {
	inform := new(Inform)
	inform.ID = inform.GetID()
	inform.Name = inform.GetName()
	inform.Events = make(map[string]string)
	inform.Params = make(map[string]string)
	return inform
}

// GetName get msg type
func (msg *Inform) GetName() string {
	return "Inform"
}

// GetID get msg id
func (msg *Inform) GetID() string {
	if len(msg.ID) < 1 {
		msg.ID = fmt.Sprintf("ID:intrnl.unset.id.%s%d.%d", msg.GetName(), time.Now().Unix(), time.Now().UnixNano())
	}
	return msg.ID
}

// CreateXML encode into xml
func (msg *Inform) CreateXML() ([]byte, error) {
	env := Envelope{}
	id := IDStruct{"1", msg.GetID()}
	env.XmlnsEnv = "http://schemas.xmlsoap.org/soap/envelope/"
	env.XmlnsEnc = "http://schemas.xmlsoap.org/soap/encoding/"
	env.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	env.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	env.XmlnsCwmp = "urn:dslforum-org:cwmp-1-0"
	env.Header = HeaderStruct{ID: id}
	manufacturer := NodeStruct{Type: XsdString, Value: msg.Manufacturer}
	oui := NodeStruct{Type: XsdString, Value: msg.OUI}
	productClass := NodeStruct{Type: XsdString, Value: msg.ProductClass}
	serialNumber := NodeStruct{Type: XsdString, Value: msg.Sn}
	deviceID := deviceIDStruct{Type: "cwmp:DeviceIdStruct", Manufacturer: manufacturer, OUI: oui, ProductClass: productClass, SerialNumber: serialNumber}
	eventLen := strconv.Itoa(len(msg.Events))
	event := EventStruct{Type: "cwmp:EventStruct[" + eventLen + "]"}
	for k, v := range msg.Events {
		eventCode := NodeStruct{Type: XsdString, Value: k}
		event.Events = append(event.Events, EventNodeStruct{EventCode: eventCode, CommandKey: v})
	}

	maxEnv := strconv.Itoa(msg.MaxEnvelopes)
	maxEnvelopes := NodeStruct{Type: XsdString, Value: maxEnv}
	currentTime := NodeStruct{Type: XsdString, Value: msg.CurrentTime}
	trys := strconv.Itoa(msg.RetryCount)
	retryCount := NodeStruct{Type: XsdString, Value: trys}
	paramLen := strconv.Itoa(len(msg.Params))
	paramList := ParameterListStruct{Type: "cwmp:ParameterValueStruct[" + paramLen + "]"}
	for k, v := range msg.Params {
		param := ParameterValueStruct{
			Name:  NodeStruct{Type: XsdString, Value: k},
			Value: NodeStruct{Type: XsdString, Value: v}}
		paramList.Params = append(paramList.Params, param)
	}
	info := informStruct{DeviceID: deviceID, Event: event, MaxEnvelopes: maxEnvelopes, CurrentTime: currentTime, RetryCount: retryCount, Params: paramList}
	env.Body = informBodyStruct{info}
	output, err := xml.MarshalIndent(env, "  ", "    ")
	//output, err := xml.Marshal(env)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Parse decode from xml
func (msg *Inform) Parse(doc *etree.Document) error {
	msg.ID = doc.FindElement("//ID").Text()
	deviceNode := doc.FindElement("//DeviceId")
	if deviceNode != nil {
		msg.Manufacturer = deviceNode.SelectElement("Manufacturer").Text()
		msg.OUI = deviceNode.SelectElement("OUI").Text()
		msg.ProductClass = deviceNode.SelectElement("ProductClass").Text()
		msg.Sn = deviceNode.SelectElement("SerialNumber").Text()
	}
	informNode := doc.FindElement("//Inform")
	if informNode != nil {
		var err error
		msg.CurrentTime = informNode.SelectElement("CurrentTime").Text()
		msg.MaxEnvelopes, err = strconv.Atoi(informNode.SelectElement("MaxEnvelopes").Text())
		if err != nil {
			return err
		}
		msg.RetryCount, err = strconv.Atoi(informNode.SelectElement("RetryCount").Text())
		if err != nil {
			return err
		}
	}

	eventNode := doc.FindElement("//Event")
	if eventNode != nil {
		//msg.Events = make(map[string]string)
		var code string
		for _, event := range eventNode.ChildElements() {
			if event != nil {
				code = event.SelectElement("EventCode").Text()
				msg.Events[code] = event.SelectElement("CommandKey").Text()
			}
		}
	}

	paramsNode := doc.FindElement("//ParameterList")
	if paramsNode != nil {
		//msg.Params = make(map[string]string)
		var name string
		for _, param := range paramsNode.ChildElements() {
			if param != nil {
				name = param.SelectElement("Name").Text()
				msg.Params[name] = param.SelectElement("Value").Text()
			}
		}
	}
	return nil
}

// IsEvent is a connect request or others
func (msg *Inform) IsEvent(event string) bool {
	if _, ok := msg.Events[event]; ok {
		return true
	}
	return false
}

// GetParam get param in inform
func (msg *Inform) GetParam(name string) (value string) {
	value = msg.Params[name]
	return
}

// GetConfigVersion get current config version
func (msg *Inform) GetConfigVersion() (version string) {
	version = msg.GetParam("InternetGatewayDevice.DeviceConfig.ConfigVersion")
	return
}

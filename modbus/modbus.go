package modbus

import (
	"bytes"
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"text/template"
	"time"

	"github.com/goburrow/modbus"
)

type input interface {
	read(*modbus.Client) ([]State, error)
	getReportTemplate() *template.Template
	getTopicTemplate() *template.Template
}

type State struct {
	Value       interface{}
	LastChanged uint16
	Address     uint16
}

type Modbus struct {
	metrics       *metrics
	Connected     bool
	Blocks        []*input
	modbusHandler *modbus.Client
	readInterval  time.Duration
}

func getInputs() []*input {
	conf := configuration.GetConfiguration()
	inputs := []*input{}
	addresses := conf.Modbus.Addresses
	for _, value := range addresses {
		switch value.AddressType {
		case "coil":
			var input input = NewCoil(value.Start, value.Count, value.Topic, value.ReportFormat)
			inputs = append(inputs, &input)
		}
	}
	return inputs
}

func NewModbus(address string, port uint16, readInterval time.Duration) *Modbus {
	modbusHandler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", address, port))
	log.Println("INFO", fmt.Sprintf("Set modbus server address to tcp://%s:%d", address, port))
	modbusHandler.Timeout = 10 * time.Second
	modbusHandler.SlaveId = 0xFF
	err := modbusHandler.Connect()
	if err != nil {
		log.Println("ERROR", err.Error())
		panic(err)
	}
	log.Println("INFO", "Connected to modbus server")
	client := modbus.NewClient(modbusHandler)
	return &Modbus{
		metrics:       newMetrics(uint16(readInterval.Milliseconds())),
		Connected:     false,
		Blocks:        getInputs(),
		modbusHandler: &client,
		readInterval:  readInterval,
	}
}

func (m Modbus) StartThread(mqtt *mqtt.Mqtt) {
	startTime := time.Now()
	time.AfterFunc(m.readInterval, func() {
		m.StartThread(mqtt)
	})
	for _, block := range m.Blocks {
		values, err := (*block).read(m.modbusHandler)
		if err != nil {
			m.Connected = false
			log.Println("ERROR", "%v\n", err)
		} else {
			m.Connected = true
		}

		// use transforms here

		// burde man have have en som reporterede tiden den havde været holdt nede?
		// med mulighed for invert?

		for _, value := range values {
			if value.LastChanged == 0 {
				go func() {
					topic := executeTemplate((*block).getTopicTemplate(), value)
					report := executeTemplate((*block).getReportTemplate(), value)
					mqtt.Publish(topic, report)
				}()
			}
		}
	}

	elapsed := time.Since(startTime)
	(*m.metrics).addRead(uint16(elapsed / time.Millisecond))
}

func executeTemplate(tmpl *template.Template, data interface{}) string {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalln("ERROR", err.Error())
	}
	return buf.String()
}

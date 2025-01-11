package modbus

import (
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"time"

	"github.com/goburrow/modbus"
)

type input interface {
	read(*modbus.Client) ([]State, error)
}

type State struct {
	value       interface{}
	lastChanged uint16
	address     uint16
}

type Modbus struct {
	metrics       *metrics
	Connected     bool
	Blocks        []*input
	modbusHandler *modbus.Client
	readInterval  time.Duration
}

// can i make an interface, for transforms? Just read in coils (perhaps create more types for show)

func getInputs() []*input {
	conf := configuration.GetConfiguration()
	inputs := []*input{}
	addresses := conf.Modbus.ReadAddresses
	for _, value := range addresses {
		switch value.AddressType {
		case "coil":
			var input input = NewCoil(value.Start, value.Count)
			inputs = append(inputs, &input)
		}
	}
	return inputs
}

func NewModbus(address string, port uint16, readInterval time.Duration) *Modbus {
	modbusHandler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", address, port))
	log.Println("INFO", fmt.Sprintf("Set modbus server address to %s:%d", address, port))
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

		for _, value := range values {
			if value.lastChanged == 0 {
				go func() {
					mqtt.Publish("output",
						fmt.Sprintf("{\"address\": %d, \"on\": %t}", value.address, value.value))
				}()
			}
		}
	}

	elapsed := time.Since(startTime)
	(*m.metrics).addRead(uint16(elapsed / time.Millisecond))
}

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
	read(*modbus.Client, time.Duration) ([]State, error)
	getConfiguration() configuration.Address
}

type State struct {
	Value       interface{}
	LastChanged uint16
	Address     uint16
	Changed     bool
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
			var input input = NewCoil(value)
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

func (modbus Modbus) StartThread(mqtt *mqtt.Mqtt) {
	startTime := time.Now()
	time.AfterFunc(modbus.readInterval, func() {
		modbus.StartThread(mqtt)
	})
	for _, block := range modbus.Blocks {
		values, err := (*block).read(modbus.modbusHandler, modbus.readInterval)
		if err != nil {
			modbus.Connected = false
			log.Println("ERROR", "%v\n", err)
		} else {
			modbus.Connected = true
		}

		for _, value := range values {
			go func() {
				report(value, (*block).getConfiguration(), *mqtt)
			}()
		}
	}

	elapsed := time.Since(startTime)
	(*modbus.metrics).addRead(uint16(elapsed / time.Millisecond))
}

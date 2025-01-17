package modbus

import (
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"time"

	"github.com/goburrow/modbus"
)

type block interface {
	read(*modbus.Client, time.Duration) ([]input, error)
}

type input struct {
	State             State
	Address           uint16
	ScanInterval      uint16
	BlockStartAddress uint16
	BlockAddressCount uint16
	BlockType         string

	templates *templates
}

func (input input) getTemplates() templates {
	return *input.templates
}

type State struct {
	Value       any
	LastChanged uint16
	Changed     bool
}

type Modbus struct {
	metrics       *metrics
	Connected     bool
	Blocks        []*block
	modbusHandler *modbus.Client
	readInterval  time.Duration
}

func getInputs() []*block {
	conf := configuration.GetConfiguration()
	inputs := []*block{}
	addresses := conf.Modbus.Addresses
	for index, value := range addresses {
		switch value.Type {
		case "coil":
			address := &addresses[index]
			var input block = NewCoilBlock(address)
			inputs = append(inputs, &input)
		}
	}
	return inputs
}

func NewModbus(address string, port uint16, readInterval time.Duration) *Modbus {
	modbusHandler := modbus.NewTCPClientHandler(
		fmt.Sprintf("%s:%d", address, port),
	)
	log.Println("INFO", fmt.Sprintf(
		"Set modbus server address to tcp://%s:%d", address, port),
	)
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
				report(value, *mqtt)
			}()
		}
	}

	elapsed := time.Since(startTime)
	(*modbus.metrics).addRead(uint16(elapsed / time.Millisecond))
}

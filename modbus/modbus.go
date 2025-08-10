package modbus

import (
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"os"
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
	running       bool
	Blocks        []*block
	client        *modbus.Client
	modbusHandler *modbus.TCPClientHandler
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
	metrics := newMetrics(uint16(readInterval.Milliseconds()))
	metrics.setConnected(0)

	modbusHandler := modbus.NewTCPClientHandler(
		fmt.Sprintf("%s:%d", address, port),
	)
	log.Println("INFO", fmt.Sprintf(
		"Set modbus server address to tcp://%s:%d", address, port),
	)
	modbusHandler.Timeout = 500 * time.Millisecond
	modbusHandler.SlaveId = 0xFF

	client := modbus.NewClient(modbusHandler)

	return &Modbus{
		metrics:       metrics,
		Connected:     false,
		running:       false,
		Blocks:        getInputs(),
		client:        &client,
		modbusHandler: modbusHandler,
		readInterval:  readInterval,
	}
}

func (modbus *Modbus) Connect(retries int) {

	if modbus.Connected {
		log.Println("INFO", "Already connected to modbus server")
		return
	}

	retry := 0
	for {
		err := modbus.modbusHandler.Connect()
		if err == nil {
			modbus.Connected = true
			modbus.metrics.setConnected(1)
			log.Println("INFO", "Connected to modbus server")
			return
		}
		retry++
		log.Println("ERROR", fmt.Sprintf("Failed to connect to modbus client, retrying (%d)", retry))
		if retries > 0 && retry >= retries {
			log.Println("ERROR", fmt.Sprintf("Could not connect to modbus client after %d attempts, exiting", retry))
			os.Exit(1)
		}
		time.Sleep(10 * time.Second)
	}
}

func (modbus *Modbus) StartThread(mqtt *mqtt.Mqtt) {

	for {
		if !modbus.running {
			break
		}
		time.Sleep(1 * time.Second)
	}
	modbus.running = true

	time.AfterFunc(modbus.readInterval, func() {
		modbus.StartThread(mqtt)
	})

	if !modbus.Connected {
		modbus.Connect(0)
	}

	startTime := time.Now()

	for _, block := range modbus.Blocks {
		values, err := (*block).read(modbus.client, modbus.readInterval)
		if err != nil {
			modbus.modbusHandler.Close()
			log.Println("ERROR", fmt.Sprintf("Failed to read from modbus block: %s", err.Error()))
			modbus.metrics.setConnected(0)
			modbus.Connected = false
			modbus.running = false
			return
		}

		for _, value := range values {
			go func() {
				report(value, *mqtt)
			}()
		}
	}

	elapsed := time.Since(startTime)
	(*modbus.metrics).addRead(uint16(elapsed / time.Millisecond))
	modbus.running = false
}

func (modbus Modbus) IsConnected() bool {
	return modbus.Connected
}

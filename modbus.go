package main

import (
	"fmt"
	"log"
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

type coil struct {
	startAddress uint16
	count        uint16
	states       []State
}

func NewCoil(startAddress uint16, count uint16) *coil {
	states := make([]State, count)
	for index, state := range states {
		state.address = startAddress + uint16(index)
		state.lastChanged = 0
		state.value = false
	}
	return &coil{
		startAddress: startAddress,
		count:        count,
		states:       states,
	}
}

func bytesToBoolArray(bytes []byte) []bool {
	var res []bool
	for _, b := range bytes {
		var result [8]bool
		for i := 0; i < 8; i++ {
			result[7-i] = (b & (1 << i)) != 0
		}
		res = append(res, result[:]...)
	}
	return res
}

func (c *coil) read(client *modbus.Client) ([]State, error) {
	res, err := (*client).ReadCoils(c.startAddress, c.count)
	results := bytesToBoolArray(res)
	if err != nil {
		log.Printf("%v\n", err)
	}
	for index, value := range results {
		state := &c.states[index]
		if value == state.value {
			state.lastChanged++
		} else {
			state.lastChanged = 0
		}
		c.states[index].value = value
	}
	return c.states, err
}

type Stats struct {
	readTimeMin   uint16
	readTimeMax   uint16
	overtimeCount uint64
	readCount     uint64
	overtime      uint64
}

func (s *Stats) addRead(readTime uint16) {
	s.readCount++
	if readTime > uint16(s.overtime) {
		s.overtimeCount++
	}
	if readTime > s.readTimeMax {
		s.readTimeMax = readTime
	}
	if readTime < s.readTimeMin {
		s.readTimeMin = readTime
	}
}

type Modbus struct {
	Stats         *Stats
	Connected     bool
	Blocks        []*input
	modbusHandler *modbus.Client
}

// can i make an interface, for transforms? Just read in coils (perhaps create more types for show)

func getInputs() []*input {
	conf := GetConfiguration()
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

func NewModbus(address string, port uint16) *Modbus {
	modbusHandler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", address, port))
	modbusHandler.Timeout = 10 * time.Second
	modbusHandler.SlaveId = 0xFF
	// modbusHandler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	err := modbusHandler.Connect()
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
	client := modbus.NewClient(modbusHandler)
	return &Modbus{
		Stats: &Stats{
			readTimeMin:   65535,
			readTimeMax:   0,
			overtimeCount: 0,
			readCount:     0,
			overtime:      0,
		},
		Connected:     false,
		Blocks:        getInputs(),
		modbusHandler: &client,
	}
}

func (m Modbus) startThread(t time.Duration, mqtt *Mqtt) {
	// conf := GetConfiguration()
	startTime := time.Now()
	time.AfterFunc(t, func() {
		m.startThread(t, mqtt)
	})
	for _, block := range m.Blocks {
		values, err := (*block).read(m.modbusHandler)
		if err != nil {
			m.Connected = false
			log.Printf("%v\n", err)
		} else {
			m.Connected = true
		}
		// use transforms here

		for _, value := range values {
			if value.lastChanged == 0 {
				mqtt.Publish("output",
					fmt.Sprintf("{\"address\": %d, \"on\": %t", value.address, value.value))
			}
		}
	}

	elapsed := time.Since(startTime)
	(*m.Stats).addRead(uint16(elapsed / time.Millisecond))
}

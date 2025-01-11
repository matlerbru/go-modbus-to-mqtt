package modbus

import (
	"log"

	"github.com/goburrow/modbus"
)

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

package modbus

import (
	"log"
	"modbus-to-mqtt/configuration"
	"time"

	"github.com/goburrow/modbus"
)

type coil struct {
	states        []State
	configuration configuration.Address
}

func NewCoil(address configuration.Address) *coil {
	states := make([]State, address.Count)
	templates := generateTemplates(address)
	for index := range states {
		states[index].Address = address.Start + uint16(index)
		states[index].Value = false
		states[index].templates = templates
	}

	return &coil{
		states:        states,
		configuration: address,
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

func (coil *coil) read(client *modbus.Client, interval time.Duration) ([]State, error) {
	res, err := (*client).ReadCoils(coil.configuration.Start, coil.configuration.Count)
	results := bytesToBoolArray(res)
	if err != nil {
		log.Printf("%v\n", err)
	}
	for index, value := range results {
		state := &coil.states[index]

		if state.Changed {
			state.LastChanged = 0
		} else {
			state.LastChanged += uint16(interval.Milliseconds())
		}

		state.Changed = value != state.Value
		coil.states[index].Value = value
	}
	return coil.states, err
}

func (coil coil) getConfiguration() configuration.Address {
	return coil.configuration
}

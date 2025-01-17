package modbus

import (
	"log"
	"modbus-to-mqtt/configuration"
	"time"

	"github.com/goburrow/modbus"
)

type coilBlock struct {
	coils        []input
	startAddress uint16
	AddressCount uint16
}

func NewCoilBlock(block *configuration.Block) coilBlock {

	conf := configuration.GetConfiguration()
	states := make([]State, block.Count)

	for index := range states {
		states[index].Value = false
	}

	var coils []input

	for index, state := range states {
		coils = append(coils, input{
			State:             state,
			Address:           block.Start + uint16(index),
			ScanInterval:      conf.Modbus.ScanInterval,
			BlockStartAddress: block.Start,
			BlockAddressCount: block.Count,
			BlockType:         block.Type,

			templates: generateTemplates(block),
		})
	}

	return coilBlock{
		coils:        coils,
		startAddress: block.Start,
		AddressCount: block.Count,
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

func (coilBlock coilBlock) read(client *modbus.Client, interval time.Duration) ([]input, error) {

	res, err := (*client).ReadCoils(coilBlock.startAddress, coilBlock.AddressCount)
	results := bytesToBoolArray(res)
	if err != nil {
		log.Printf("%v\n", err)
	}

	count := coilBlock.AddressCount
	for index, value := range results {

		if uint16(index) == count {
			break
		}

		state := &coilBlock.coils[index].State

		if state.Changed {
			state.LastChanged = 0
		} else {
			state.LastChanged += uint16(interval.Milliseconds())
		}

		state.Changed = value != state.Value
		coilBlock.coils[index].State.Value = value
	}
	return coilBlock.coils, err
}

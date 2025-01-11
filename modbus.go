package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

type ModbusMetrics struct {
	readCounter         prometheus.Counter
	maximimReadTime     prometheus.Gauge
	minimumReadTime     prometheus.Gauge
	readOvertimeCounter prometheus.Counter

	overtimeLimit uint16
	maxReadTime   uint16
	minReadTime   uint16
}

func NewModbusMetrics(overtimeLimit uint16) *ModbusMetrics {
	return &ModbusMetrics{
		readCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_read_count",
			Help: "Total number of times modbus server has been read",
		}),
		maximimReadTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_read_time_max",
			Help: "Largest read time for modbus server",
		}),
		minimumReadTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_read_time_min",
			Help: "Smallest read time for modbus server",
		}),
		readOvertimeCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_read_overtime_count",
			Help: "Total number of times modbus server read has been above the allowed time",
		}),
		overtimeLimit: overtimeLimit,
		maxReadTime:   0,
		minReadTime:   65535,
	}
}

func (metrics *ModbusMetrics) addRead(readTime uint16) {
	metrics.readCounter.Inc()

	if readTime > uint16(metrics.overtimeLimit) {
		metrics.readOvertimeCounter.Inc()
	}

	if readTime > metrics.maxReadTime {
		metrics.maxReadTime = readTime
		metrics.maximimReadTime.Set(float64(metrics.maxReadTime))
	}
	if readTime < metrics.minReadTime {
		metrics.minReadTime = readTime
		metrics.minimumReadTime.Set(float64(metrics.minReadTime))
	}
}

type Modbus struct {
	Stats         *ModbusMetrics
	Connected     bool
	Blocks        []*input
	modbusHandler *modbus.Client
	readInterval  time.Duration
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

func NewModbus(address string, port uint16, readInterval time.Duration) *Modbus {
	modbusHandler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", address, port))
	modbusHandler.Timeout = 10 * time.Second
	modbusHandler.SlaveId = 0xFF
	err := modbusHandler.Connect()
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
	client := modbus.NewClient(modbusHandler)
	return &Modbus{
		Stats:         NewModbusMetrics(uint16(readInterval.Milliseconds())),
		Connected:     false,
		Blocks:        getInputs(),
		modbusHandler: &client,
		readInterval:  readInterval,
	}
}

func (m Modbus) startThread(mqtt *Mqtt) {

	startTime := time.Now()
	time.AfterFunc(m.readInterval, func() {
		m.startThread(mqtt)
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
				mqtt.Publish("output",
					fmt.Sprintf("{\"address\": %d, \"on\": %t}", value.address, value.value))
			}
		}
	}

	elapsed := time.Since(startTime)
	(*m.Stats).addRead(uint16(elapsed / time.Millisecond))
}

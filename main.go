package main

import (
	"modbus-to-mqtt/configuration"
	"time"
)

func main() {
	conf := configuration.GetConfiguration()

	mqtt := NewMqtt(conf.Mqtt.Address, conf.Mqtt.Port, conf.Mqtt.Qos)
	mqtt.SetBaseTopic(conf.Mqtt.MainTopic)
	mqtt.Connect()

	modbus := NewModbus(conf.Modbus.Address, conf.Modbus.Port, time.Duration(conf.Modbus.ScanInterval)*time.Millisecond)
	modbus.startThread(mqtt)

	if conf.Metrics.Enabled {
		metricsExorter := NewMetricsExporter(modbus.Stats)
		metricsExorter.serve()
	}

	select {}
}

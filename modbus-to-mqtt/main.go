package main

import (
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/modbus"
	"modbus-to-mqtt/mqtt"
	"time"
)

func main() {
	conf := configuration.GetConfiguration()

	mqtt := mqtt.NewMqtt(conf.Mqtt.Address, conf.Mqtt.Port, conf.Mqtt.Qos)
	mqtt.SetBaseTopic(conf.Mqtt.MainTopic)
	mqtt.Connect()

	modbus := modbus.NewModbus(
		conf.Modbus.Address,
		conf.Modbus.Port,
		time.Duration(conf.Modbus.ScanInterval)*time.Millisecond,
	)
	modbus.StartThread(mqtt)

	if conf.Metrics.Enabled {
		go func() {
			metricsExporter := NewMetricsExporter()
			metricsExporter.serve()
		}()
	}

	select {}
}

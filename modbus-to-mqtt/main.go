package main

import (
	"fmt"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/modbus"
	"modbus-to-mqtt/mqtt"
	"time"
)

func main() {
	conf := configuration.GetConfiguration()

	if conf.Metrics.Enabled {
		go func() {
			metricsExporter := NewMetricsExporter()
			metricsExporter.serve()
		}()
	}

	mqtt := mqtt.NewMqtt(conf.Mqtt.Address, conf.Mqtt.Port, conf.Mqtt.Qos)
	mqtt.SetBaseTopic(conf.Mqtt.MainTopic)
	mqtt.Connect(12)

	modbus := modbus.NewModbus(
		conf.Modbus.Address,
		conf.Modbus.Port,
		time.Duration(conf.Modbus.ScanInterval)*time.Millisecond,
	)
	modbus.Connect(12)
	modbus.StartThread(mqtt)

	healthCheck := NewHealthCheck(func() error {
		if !mqtt.IsConnected() {
			return fmt.Errorf("MQTT client is not connected")
		}
		if !modbus.IsConnected() {
			return fmt.Errorf("Modbus client is not connected")
		}
		return nil
	})
	healthCheck.serve()
	select {}
}

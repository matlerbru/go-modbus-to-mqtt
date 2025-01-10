package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	client    mqtt.Client
	baseTopic string
	qos uint8
}

func onConnect(client mqtt.Client) {
	log.Println("Connected to mqtt broker")
}

func onDisconnect(client mqtt.Client, error error) {
	log.Printf("Disconnected from mqtt broker: %s", error.Error())
	os.Exit(1)
}

func NewMqtt(broker string, port uint16, qos uint8) *Mqtt {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	options.OnConnect = onConnect
	options.OnConnectionLost = onDisconnect

	m := Mqtt{qos: qos}
	m.client = mqtt.NewClient(options)
	return &m
}

func (m Mqtt) Connect() {
	logged := false
	for {
		token := m.client.Connect()
		res := token.WaitTimeout(1 * time.Second)
		if res {
			break
		}
		if !logged {
			log.Println("Failed to connect to mqtt broker")
			logged = true
		}
		time.Sleep(5 * time.Second)
	}
}

func (m Mqtt) Publish(topic string, message string) {
	conf := GetConfiguration()
	t := m.client.Publish(m.baseTopic+topic, conf.Mqtt.Qos, false, message)
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			log.Println(t.Error())
		}
	}()
}

func (m *Mqtt) SetBaseTopic(topic string) {
	m.baseTopic = topic + "/"
}

func (m Mqtt) IsConnected() bool {
	return m.client.IsConnected()
}

package mqtt

import (
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	metrics   *metrics
	client    mqtt.Client
	baseTopic string
	qos       uint8
}

func onConnect(client mqtt.Client) {
	log.Println("INFO", "Connected to mqtt broker")
}

func onDisconnect(client mqtt.Client, error error) {
	log.Println("ERROR", "Disconnected from mqtt broker: %s", error.Error())
	os.Exit(1)
}

func NewMqtt(broker string, port uint16, qos uint8) *Mqtt {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	log.Println("INFO", fmt.Sprintf("Set broker address to mqtt://%s:%d", broker, port))
	options.OnConnect = onConnect
	options.OnConnectionLost = onDisconnect

	m := Mqtt{metrics: newMetrics(), qos: qos}
	m.client = mqtt.NewClient(options)
	return &m
}

func (m Mqtt) Connect() {
	for {
		token := m.client.Connect()
		res := token.WaitTimeout(1 * time.Second)
		if res {
			break
		}
	}
}

func (m Mqtt) Publish(topic string, message string) {
	conf := configuration.GetConfiguration()
	t := m.client.Publish(m.baseTopic+topic, conf.Mqtt.Qos, false, message)
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			log.Println("ERROR", t.Error())
		} else {
			log.Println("INFO", fmt.Sprintf("Published message on topic %s, payload: '%s', qos: %d", m.baseTopic+topic, message, conf.Mqtt.Qos))
			m.metrics.incrementPublishCounter()
		}
	}()
}

func (m *Mqtt) SetBaseTopic(topic string) {
	log.Println("INFO", fmt.Sprintf("Base topic set to %s", topic))
	m.baseTopic = topic + "/"
}

func (m Mqtt) IsConnected() bool {
	return m.client.IsConnected()
}

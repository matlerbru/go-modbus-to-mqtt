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

var instances []*Mqtt

func onConnect(client mqtt.Client) {
	log.Println("INFO", "Connected to MQTT broker")

	for _, instance := range instances {
		if instance.client == client {
			instance.metrics.setConnected(1)
			break
		}
	}
}

func onDisconnect(client mqtt.Client, err error) {
	log.Println("ERROR", fmt.Sprintf("Disconnected from MQTT broker: %s", err.Error()))

	for _, instance := range instances {
		if instance.client == client {
			instance.metrics.setConnected(0)
			instance.Connect(0)
			break
		}
	}
}

func NewMqtt(broker string, port uint16, qos uint8) *Mqtt {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	log.Println("INFO", fmt.Sprintf("Set broker address to mqtt://%s:%d", broker, port))
	options.OnConnect = onConnect
	options.OnConnectionLost = onDisconnect

	m := Mqtt{metrics: newMetrics(), qos: qos}
	m.client = mqtt.NewClient(options)
	m.metrics.setConnected(0)

	instances = append(instances, &m)

	return &m
}

func (m Mqtt) Connect(retries int) {
	retry := 0
	log.Println("INFO", "Connecting to MQTT broker")
	for {
		token := m.client.Connect()
		res := token.WaitTimeout(1 * time.Second)
		if res && token.Error() == nil {
			break
		}
		retry++
		log.Println("ERROR", fmt.Sprintf("Failed to connect to MQTT broker, retrying (%d)", retry))
		if retries > 0 && retry >= retries {
			log.Println("ERROR", fmt.Sprintf("Could not connect to MQTT broker after %d attempts, exiting", retry))
			os.Exit(1)
		}
		time.Sleep(10 * time.Second)
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
			log.Println("INFO", fmt.Sprintf(
				"Published message on topic %s, payload: '%s', qos: %d",
				m.baseTopic+topic,
				message,
				conf.Mqtt.Qos,
			))
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

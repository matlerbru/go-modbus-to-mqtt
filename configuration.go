package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type config struct {
	Mqtt struct {
		Address   string `yaml:"address"`
		Port      uint16 `yaml:"port"`
		MainTopic string `yaml:"mainTopic"`
		Qos       uint8  `yaml:"qos"`
	} `yaml:"mqtt"`
	Modbus struct {
		Address       string `yaml:"address"`
		Port          uint16 `yaml:"port"`
		ReadAddresses []struct {
			AddressType string `yaml:"type"`
			Start       uint16 `yaml:"start"`
			Count       uint16 `yaml:"count"`
		} `yaml:"readAddresses"`
		ScanInterval uint16 `yaml:"scanInterval"`
	} `yaml:"modbus"`
	Metrics struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"metrics"`
	// longPressTime uint16 `yaml:"longPressTime"`
}

type conf struct {
	configuration config
	once          sync.Once
}

func (conf *conf) readConfiguration(path string) {
	log.Println("INFO", fmt.Sprintf("Reading configuration from %s", path))
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Println("ERROR", err.Error())
		panic(err)
	}
	var c config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Println("ERROR", err.Error())
		panic(err)
	}
	conf.configuration = c
}

var c = conf{
	once: sync.Once{},
}

func GetConfiguration() *config {
	c.once.Do(func() {

		configPath := os.Getenv("MODBUS_TO_MQTT_CONFIG_PATH")
		if configPath == "" {
			configPath = "config.yaml"
		}
		c.readConfiguration(configPath)
	})
	return &c.configuration
}

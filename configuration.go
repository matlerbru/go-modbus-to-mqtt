package main

import (
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
	// longPressTime uint16 `yaml:"longPressTime"`
}

type conf struct {
	configuration config
	once          sync.Once
}

func (conf *conf) readConfiguration(path string) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var c config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic(err)
	}
	conf.configuration = c
}

var c = conf{
	once: sync.Once{},
}

func GetConfiguration() *config {
	c.once.Do(func() {
		c.readConfiguration("config.yaml")
	})
	return &c.configuration
}

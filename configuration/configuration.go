package configuration

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type config struct {
	Mqtt    Mqtt    `yaml:"mqtt"`
	Modbus  Modbus  `yaml:"modbus"`
	Metrics Metrics `yaml:"metrics"`
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

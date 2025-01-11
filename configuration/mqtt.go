package configuration

type mqtt struct {
	Address   string `yaml:"address"`
	Port      uint16 `yaml:"port"`
	MainTopic string `yaml:"mainTopic"`
	Qos       uint8  `yaml:"qos"`
}

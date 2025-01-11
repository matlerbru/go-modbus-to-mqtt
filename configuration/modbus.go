package configuration

type modbus struct {
	Address       string `yaml:"address"`
	Port          uint16 `yaml:"port"`
	ReadAddresses []struct {
		AddressType string `yaml:"type"`
		Start       uint16 `yaml:"start"`
		Count       uint16 `yaml:"count"`
	} `yaml:"readAddresses"`
	ScanInterval uint16 `yaml:"scanInterval"`
}

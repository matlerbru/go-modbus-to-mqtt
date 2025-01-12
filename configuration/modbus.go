package configuration

type modbus struct {
	Address   string `yaml:"address"`
	Port      uint16 `yaml:"port"`
	Addresses []struct {
		AddressType  string `yaml:"type"`
		Start        uint16 `yaml:"start"`
		Count        uint16 `yaml:"count"`
		Topic        string `yaml:topic`
		ReportFormat string `yaml:"reportFormat"`
	} `yaml:"addresses"`
	ScanInterval uint16 `yaml:"scanInterval"`
}

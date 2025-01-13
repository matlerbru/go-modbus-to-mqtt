package configuration

type Modbus struct {
	Address      string    `yaml:"address"`
	Port         uint16    `yaml:"port"`
	Addresses    []Address `yaml:"addresses"`
	ScanInterval uint16    `yaml:"scanInterval"`
}

type Address struct {
	AddressType string   `yaml:"type"`
	Start       uint16   `yaml:"start"`
	Count       uint16   `yaml:"count"`
	Topic       string   `yaml:"topic"`
	Report      []Report `yaml:"report"`
}

type Report struct {
	EnabledTemplate string `yaml:"enabled"`
	FormatTemplate  string `yaml:"format"`
}

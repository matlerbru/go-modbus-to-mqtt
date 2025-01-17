package configuration

type Modbus struct {
	Address      string  `yaml:"address"`
	Port         uint16  `yaml:"port"`
	Addresses    []Block `yaml:"blocks"`
	ScanInterval uint16  `yaml:"scanInterval"`
}

type Block struct {
	Type   string   `yaml:"type"`
	Start  uint16   `yaml:"start"`
	Count  uint16   `yaml:"count"`
	Topic  string   `yaml:"topic"`
	Report []Report `yaml:"report"`
}

type Report struct {
	SendOn       string `yaml:"sendOn"`
	Format       string `yaml:"format"`
	OnlyOnChange bool   `yaml:"onlyOnChange"`
}

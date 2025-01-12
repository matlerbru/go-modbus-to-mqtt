package modbus

import (
	"fmt"
	"log"
	"text/template"

	"github.com/goburrow/modbus"
)

type coil struct {
	startAddress   uint16
	count          uint16
	states         []State
	topicTemplate  *template.Template
	reportTemplate *template.Template
}

func NewCoil(startAddress uint16, count uint16, topic string, reportFormat string) *coil {
	states := make([]State, count)
	for index := range states {
		states[index].Address = startAddress + uint16(index)
	}

	reportTemplate, err := template.New("template").Parse(reportFormat)
	if err != nil {
		log.Fatalln("ERROR", fmt.Sprintf("Unable to parse reportFormat: %s", err.Error()))
	}
	topicTemplate, err := template.New("template").Parse(topic)
	if err != nil {
		log.Fatalln("ERROR", fmt.Sprintf("Unable to parse reportFormat: %s", err.Error()))
	}

	return &coil{
		startAddress:   startAddress,
		count:          count,
		states:         states,
		topicTemplate:  topicTemplate,
		reportTemplate: reportTemplate,
	}
}

func bytesToBoolArray(bytes []byte) []bool {
	var res []bool
	for _, b := range bytes {
		var result [8]bool
		for i := 0; i < 8; i++ {
			result[7-i] = (b & (1 << i)) != 0
		}
		res = append(res, result[:]...)
	}
	return res
}

func (c *coil) read(client *modbus.Client) ([]State, error) {
	res, err := (*client).ReadCoils(c.startAddress, c.count)
	results := bytesToBoolArray(res)
	if err != nil {
		log.Printf("%v\n", err)
	}
	for index, value := range results {
		state := &c.states[index]
		if value == state.Value {
			state.LastChanged++
		} else {
			state.LastChanged = 0
		}
		c.states[index].Value = value
	}
	return c.states, err
}

func (c coil) getReportTemplate() *template.Template {
	return c.reportTemplate
}

func (c coil) getTopicTemplate() *template.Template {
	return c.topicTemplate
}

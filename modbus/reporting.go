package modbus

import (
	"bytes"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"strconv"
	"text/template"
)

func report(state State, mqtt mqtt.Mqtt) {

	var topicBuffer bytes.Buffer
	state.templates.topic.Execute(&topicBuffer, state)
	topic := topicBuffer.String()

	for _, report := range state.templates.report {

		var enabledBuffer bytes.Buffer
		report.enabled.Execute(&enabledBuffer, state)
		enabled, err := strconv.ParseBool(enabledBuffer.String())
		if err != nil {
			log.Println("ERROR", err.Error())
		}
		if !enabled {
			continue
		}

		var formatBuffer bytes.Buffer
		report.format.Execute(&formatBuffer, state)

		mqtt.Publish(topic, formatBuffer.String())
	}
}

func generateTemplates(address configuration.Address) *templates {

	topic, err := template.New("topic").Parse(address.Topic)
	if err != nil {
		log.Fatalln("ERROR", err.Error())
	}

	var reports []reportTemplate

	for _, report := range address.Report {

		enabled, err := template.New("enabled").Parse(report.EnabledTemplate)
		if err != nil {
			log.Fatalln("ERROR", err.Error())
		}

		format, err := template.New("format").Parse(report.FormatTemplate)
		if err != nil {
			log.Fatalln("ERROR", err.Error())
		}

		reports = append(reports, reportTemplate{enabled: enabled, format: format})

	}

	return &templates{
		topic:  topic,
		report: reports,
	}
}

type templates struct {
	topic  *template.Template
	report []reportTemplate
}

type reportTemplate struct {
	enabled *template.Template
	format  *template.Template
}

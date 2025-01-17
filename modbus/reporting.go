package modbus

import (
	"bytes"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"strconv"
	"strings"
	"text/template"
)

func report(input input, mqtt mqtt.Mqtt) {

	var reports []string

	for _, report := range input.getTemplates().reports {

		if report.onlyOnChange && !input.State.Changed {
			continue
		}

		enabled, err := strconv.ParseBool(report.sendOn.Execute(input))
		if err != nil {
			log.Println("ERROR", err)
			continue
		}

		if !enabled {
			continue
		}
		report := report.format.Execute(input)
		reports = append(reports, report)
	}

	topic := input.templates.topic.Execute(input)

	for _, report := range reports {
		mqtt.Publish(topic, report)
	}
}

func generateTemplates(block *configuration.Block) *templates {

	topic := NewTemplateable(block.Topic)

	var reports []reportTemplateables

	for _, report := range block.Report {
		sendOn := NewTemplateable(report.SendOn)
		format := NewTemplateable(report.Format)

		reports = append(reports, reportTemplateables{
			sendOn:       sendOn,
			format:       format,
			onlyOnChange: report.OnlyOnChange,
		})
	}

	return &templates{
		topic:   topic,
		reports: reports,
	}
}

type templateable struct {
	template *template.Template
	raw      string
	enabled  bool
}

func NewTemplateable(text string) templateable {
	templ, err := template.New("").Parse(text)
	if err != nil {
		log.Fatalln("ERROR", err.Error())
	}

	return templateable{
		template: templ,
		raw:      text,
		enabled:  strings.Contains(text, "{{") && strings.Contains(text, "}}"),
	}
}

func (templateable templateable) Execute(structure any) string {
	if templateable.enabled {
		var buffer bytes.Buffer
		templateable.template.Execute(&buffer, structure)
		return buffer.String()

	}
	return templateable.raw
}

type templates struct {
	topic   templateable
	reports []reportTemplateables
}

type reportTemplateables struct {
	sendOn       templateable
	format       templateable
	onlyOnChange bool
}

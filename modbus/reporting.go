package modbus

import (
	"bytes"
	"fmt"
	"log"
	"modbus-to-mqtt/configuration"
	"modbus-to-mqtt/mqtt"
	"strconv"
	"text/template"
)

func report(state State, address configuration.Address, mqtt mqtt.Mqtt) {

	for _, report := range address.Report {
		enabledStr, err := executeTemplate("enabled", report.EnabledTemplate, state)
		if err != nil {
			log.Println("ERROR", fmt.Sprintf("Error in 'enabled' template: %s", err.Error()))
			continue
		}

		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			log.Println("ERROR", fmt.Sprintf("Invalid boolean value in 'enabled' template: %s", err.Error()))
			continue
		}

		if !enabled {
			continue
		}

		templatedReport, err := executeTemplate("format", report.FormatTemplate, state)
		if err != nil {
			log.Println("ERROR", fmt.Sprintf("Error in 'format' template: %s", err.Error()))
			continue
		}

		templatedTopic, err := executeTemplate("topic", address.Topic, state)
		if err != nil {
			log.Println("ERROR", fmt.Sprintf("Error in 'topic' template: %s", err.Error()))
			continue
		}

		mqtt.Publish(templatedTopic, templatedReport)
	}
}

func executeTemplate(templateName, templateString string, data interface{}) (string, error) {
	tmpl, err := template.New(templateName).Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template '%s': %w", templateName, err)
	}
	var result bytes.Buffer
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}
	return result.String(), nil
}

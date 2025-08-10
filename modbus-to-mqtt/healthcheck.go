package main

import (
	"log"
	"net/http"

	"modbus-to-mqtt/modbus"
	"modbus-to-mqtt/mqtt"
)

type HealthCheck struct {
	mqtt   *mqtt.Mqtt
	modbus *modbus.Modbus
}

func NewHealthCheck(mqtt *mqtt.Mqtt, modbus *modbus.Modbus) *HealthCheck {
	return &HealthCheck{
		mqtt:   mqtt,
		modbus: modbus,
	}
}

func (healthCheck *HealthCheck) check() bool {
	if !healthCheck.mqtt.IsConnected() {
		return false
	}
	if !healthCheck.modbus.IsConnected() {
		return false
	}
	return true
}

func (healthCheck *HealthCheck) serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", "no-store")

		if healthCheck.check() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("UNHEALTHY"))
	})

	log.Println("INFO", "Serving health at :8080/health")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("ERROR HealthCheck server stopped with error: %v", err)
	}
}

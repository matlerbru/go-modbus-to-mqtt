package main

import (
	"log"
	"net/http"
)

type HealthCheck struct {
	callback func() error
}

func NewHealthCheck(callback func() error) *HealthCheck {
	return &HealthCheck{
		callback: callback,
	}
}

func (HealthCheck *HealthCheck) serve() {
	log.Println("INFO", "Serving health at localhost:8080/health")
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if HealthCheck.callback() != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Error"))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}
	})
	err := http.ListenAndServe(":8080", nil)
	log.Println("ERROR", "HealthCheck server stopped with error: %v", err)
}

package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsExporter struct {
}

func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{}
}

func (metrics *MetricsExporter) serve() {
	log.Println("INFO", "Serving metrics at localhost:9090/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9090", nil)
	log.Println("ERROR", "Metrics server stopped with error: %v", err)
}

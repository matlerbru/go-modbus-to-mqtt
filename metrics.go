package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsExporter struct {
}

func NewMetricsExporter(stats *ModbusMetrics) *MetricsExporter {
	return &MetricsExporter{}
}

func (metrics *MetricsExporter) serve() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9090", nil)
}

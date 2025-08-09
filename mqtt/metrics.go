package mqtt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
)

type metrics struct {
	connected      prometheus.Gauge
	publishCounter prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		publishCounter: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_mqtt_publish_count",
			Help: "Total number of messages published by mqtt",
		}),
		connected: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_mqtt_connected",
			Help: "Indicates if the mqtt client is connected",
		}),
	}
}

func (metrics *metrics) setConnected(value float64) {
	if value < 0.0 || value > 1.0 {
		log.Println("ERROR", "Invalid value for connected metric, must be 0 or 1")
		return
	}
	metrics.connected.Set(value)
}

func (metrics *metrics) incrementPublishCounter() {
	metrics.publishCounter.Inc()
}

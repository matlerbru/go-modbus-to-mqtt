package mqtt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	publishCounter prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		publishCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_mqtt_publish_count",
			Help: "Total number of messages published by mqtt",
		}),
	}
}

func (metrics *metrics) incrementPublishCounter() {
	metrics.publishCounter.Inc()
}

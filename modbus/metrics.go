package modbus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	readCounter         prometheus.Counter
	maximimReadTime     prometheus.Gauge
	minimumReadTime     prometheus.Gauge
	readOvertimeCounter prometheus.Counter

	overtimeLimit uint16
	maxReadTime   uint16
	minReadTime   uint16
}

func newMetrics(overtimeLimit uint16) *metrics {
	return &metrics{
		readCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_modbus_read_count",
			Help: "Total number of times modbus server has been read",
		}),
		maximimReadTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_read_time_max",
			Help: "Largest read time for modbus server",
		}),
		minimumReadTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_read_time_min",
			Help: "Smallest read time for modbus server",
		}),
		readOvertimeCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_modbus_read_overtime_count",
			Help: "Total number of times modbus server read has been above the allowed time",
		}),
		overtimeLimit: overtimeLimit,
		maxReadTime:   0,
		minReadTime:   65535,
	}
}

func (metrics *metrics) addRead(readTime uint16) {
	metrics.readCounter.Inc()

	if readTime > uint16(metrics.overtimeLimit) {
		metrics.readOvertimeCounter.Inc()
	}

	if readTime > metrics.maxReadTime {
		metrics.maxReadTime = readTime
		metrics.maximimReadTime.Set(float64(metrics.maxReadTime))
	}
	if readTime < metrics.minReadTime {
		metrics.minReadTime = readTime
		metrics.minimumReadTime.Set(float64(metrics.minReadTime))
	}
}

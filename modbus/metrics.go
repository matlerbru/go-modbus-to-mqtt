package modbus

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	readCounter         prometheus.Counter
	readTime            prometheus.Gauge
	maximimReadTime     prometheus.Gauge
	readOvertimeCounter prometheus.Counter

	overtimeLimit uint16
	maxReadTime   uint16

	connected prometheus.Gauge
}

func newMetrics(overtimeLimit uint16) *metrics {
	return &metrics{
		connected: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_connected",
			Help: "Indicates if the modbus client is connected",
		}),
		readCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "modbus_to_mqtt_modbus_read_count",
			Help: "Times modbus server is been read",
		}),
		readTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_read_time",
			Help: "Read time for modbus server",
		}),
		maximimReadTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_read_time_max",
			Help: "Largest read time for modbus server",
		}),
		readOvertimeCounter: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "modbus_to_mqtt_modbus_read_overtime_count",
			Help: "Total number of times modbus server read has been above the allowed time",
		}),
		overtimeLimit: overtimeLimit,
		maxReadTime:   0,
	}
}

func (metrics *metrics) setConnected(value float64) {
	if value < 0.0 || value > 1.0 {
		log.Println("ERROR", "Invalid value for connected metric, must be 0 or 1")
		return
	}
	metrics.connected.Set(value)
}

func (metrics *metrics) addRead(readTime uint16) {
	metrics.readCounter.Inc()
	metrics.readTime.Set(float64(readTime))

	if readTime > uint16(metrics.overtimeLimit) {
		metrics.readOvertimeCounter.Inc()
		log.Println("WARNING", "Read time is above the allowed limit, read time:", readTime)
	}

	if readTime > metrics.maxReadTime {
		metrics.maxReadTime = readTime
		metrics.maximimReadTime.Set(float64(metrics.maxReadTime))
	}
}

package statusHandler

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-testing/core"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusStatusHandler will define the handler which will update prometheus metrics
type PrometheusStatusHandler struct {
	prometheusGaugeMetrics sync.Map
}

// InitMetricsMap will init the map of prometheus metrics
func (psh *PrometheusStatusHandler) InitMetricsMap() {
	psh.prometheusGaugeMetrics = sync.Map{}
}

// will create a prometheus gauge and add it to the sync map
func (psh *PrometheusStatusHandler) addMetric(name string, help string) {
	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	psh.prometheusGaugeMetrics.Store(name, metric)
}

// InitMetrics will declare and init all the metrics which should be used for Prometheus
func (psh *PrometheusStatusHandler) InitMetrics() {
	psh.InitMetricsMap()

	psh.addMetric(core.MetricSynchronizedRound, "The round where the synchronized blockchain is")
	psh.addMetric(core.MetricNonce, "The nonce for the node")
	psh.addMetric(core.MetricCurrentRound, "The current round where the node is")
	psh.addMetric(core.MetricNumConnectedPeers, "The current number of peers connected")
	psh.addMetric(core.MetricIsSyncing, "The synchronization state. If it's in process of syncing will be 1"+
		" and if it's synchronized will be 0")

	psh.prometheusGaugeMetrics.Range(func(key, value interface{}) bool {
		gauge := value.(prometheus.Gauge)
		_ = prometheus.Register(gauge)
		return true
	})
}

// NewPrometheusStatusHandler will return an instance of a PrometheusStatusHandler
func NewPrometheusStatusHandler() *PrometheusStatusHandler {
	psh := new(PrometheusStatusHandler)
	psh.InitMetrics()
	return psh
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *PrometheusStatusHandler) IsInterfaceNil() bool {
	if psh == nil {
		return true
	}
	return false
}

// Increment will be used for incrementing the value for a key
func (psh *PrometheusStatusHandler) Increment(key string) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Inc()
	}
}

// AddUint64 will be used for increase the value for a key with a specific value
func (psh *PrometheusStatusHandler) AddUint64(key string, val uint64) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Add(float64(val))
	}
}

// Decrement will be used for decrementing the value for a key
func (psh *PrometheusStatusHandler) Decrement(key string) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Dec()
	}
}

// SetInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetInt64Value(key string, value int64) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Set(float64(value))
	}
}

// SetUInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetUInt64Value(key string, value uint64) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Set(float64(value))
	}
}

// SetStringValue method - will update the value for a key
func (psh *PrometheusStatusHandler) SetStringValue(key string, value string) {
}

// Close will unregister Prometheus metrics
func (psh *PrometheusStatusHandler) Close() {
	psh.prometheusGaugeMetrics.Range(func(key, value interface{}) bool {
		gauge := value.(prometheus.Gauge)
		prometheus.Unregister(gauge)
		return true
	})
}

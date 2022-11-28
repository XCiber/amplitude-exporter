package amplitude

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type MetricInfo struct {
	desc   *prometheus.Desc
	value  float64
	key    string
	acc    float64
	mType  prometheus.ValueType
	lock   sync.RWMutex
	offset int
}

func (mi *MetricInfo) GetValue() float64 {
	mi.lock.RLock()
	defer mi.lock.RUnlock()
	switch mi.mType {
	case prometheus.GaugeValue:
		return mi.value
	default:
		return mi.value + mi.acc
	}
}

func (mi *MetricInfo) GetKey() string {
	mi.lock.RLock()
	defer mi.lock.RUnlock()
	return mi.key
}

func (mi *MetricInfo) Add(key string, value float64, previousKey string, previousValue float64) {
	mi.lock.Lock()
	defer mi.lock.Unlock()

	if value < 0 {
		value = 0
	}
	if previousValue < 0 {
		previousValue = 0
	}

	switch mi.key {
	case key:
	case previousKey:
		if mi.acc <= previousValue {
			mi.value += previousValue
		} else {
			mi.value += mi.acc
		}
		mi.key = key
	default:
		mi.value += mi.acc
		mi.key = key
	}
	mi.acc = value
}

func (mi *MetricInfo) Set(key string, value float64) {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	mi.key = key
	mi.value = value
}

func (mi *MetricInfo) Reset() {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	mi.value = 0
	mi.acc = 0
	mi.key = ""
}

func (mi *MetricInfo) GetPromMetric() (prometheus.Metric, error) {
	return prometheus.NewConstMetric(
		mi.desc,
		mi.mType,
		mi.GetValue())
}

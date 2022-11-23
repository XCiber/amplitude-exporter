package amplitude

import (
	_ "embed"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed test.json
var testJson []byte

func TestExporter(t *testing.T) {
	var r ChartResponse
	err := json.Unmarshal(testJson, &r)
	require.NoError(t, err)
	require.Greater(t, len(r.Data.XValues), 1, "no x values")
	require.Greater(t, len(r.Data.Series[0]), 1, "no y values")
	lastKey := r.Data.XValues[len(r.Data.XValues)-1]
	lastValue := r.Data.Series[0][len(r.Data.Series[0])-1].Value
	previousKey := r.Data.XValues[len(r.Data.XValues)-2]
	previousValue := r.Data.Series[0][len(r.Data.Series[0])-2].Value
	log.Infof("Receive last %s=%f", lastKey, lastValue)
	log.Infof("Receive previous %s=%f", previousKey, previousValue)

	g := newGauge()
	g.Set(previousKey, previousValue)
	require.Equal(t, previousValue, g.GetValue())
	require.Equal(t, float64(10), g.GetValue())

	c := newCounter()
	c.Add(lastKey, lastValue, previousKey, previousValue)
	require.Equal(t, lastValue, c.GetValue())
	require.Equal(t, float64(22), c.GetValue())
}

func newCounter() *MetricInfo {
	return &MetricInfo{
		desc: prometheus.NewDesc(
			"counter",
			"Number of requests",
			[]string{"project", "chart", "key"},
			nil,
		),
		mType: prometheus.CounterValue,
	}
}

func newGauge() *MetricInfo {
	return &MetricInfo{
		desc: prometheus.NewDesc(
			"gauge",
			"Number of requests",
			[]string{"project", "chart", "key"},
			nil,
		),
		mType: prometheus.GaugeValue,
	}
}

func TestNewCounter(t *testing.T) {
	// create metric
	m := newCounter()
	require.Equal(t, float64(0), m.GetValue())
}

func TestCounterFirstAdd(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	require.Equal(t, float64(5), m.GetValue())
}

func TestCounterAddSameKey(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:05", 7, "10:00", 3)
	require.Equal(t, float64(7), m.GetValue())
}

func TestCounterNewKeyAddValue(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:10", 1, "10:05", 5)
	require.Equal(t, float64(6), m.GetValue())
}

func TestCounterNewKeyAddPrevious(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:10", 0, "10:05", 7)
	require.Equal(t, float64(7), m.GetValue())
}

func TestCounterNewKeyAddBoth(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:10", 1, "10:05", 7)
	require.Equal(t, float64(8), m.GetValue())
}

func TestCounterNewKeyAddPreviousLess(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:10", 3, "10:05", 3)
	require.Equal(t, float64(8), m.GetValue())
}

func TestCounterNewKeyAddNegative(t *testing.T) {
	m := newCounter()
	m.Add("10:05", 5, "10:00", 3)
	m.Add("10:10", -3, "10:05", 3)
	require.Equal(t, float64(5), m.GetValue())
}

func TestNewGauge(t *testing.T) {
	// create metric
	m := newGauge()
	require.Equal(t, float64(0), m.GetValue())
}

func TestGaugeSetKey(t *testing.T) {
	// create metric
	m := newGauge()
	m.Set("10:00", 5)
	require.Equal(t, float64(5), m.GetValue())
}

func TestGaugeSetSameKey(t *testing.T) {
	// create metric
	m := newGauge()
	m.Set("10:00", 5)
	m.Set("10:00", 7)
	require.Equal(t, float64(7), m.GetValue())
}

func TestGaugeSetNewKey(t *testing.T) {
	// create metric
	m := newGauge()
	m.Set("10:00", 5)
	m.Set("10:05", 15)
	require.Equal(t, float64(15), m.GetValue())
}

func TestGaugeSetNewKeyNegative(t *testing.T) {
	// create metric
	m := newGauge()
	m.Set("10:00", 5)
	m.Set("10:05", -15)
	require.Equal(t, float64(-15), m.GetValue())
}

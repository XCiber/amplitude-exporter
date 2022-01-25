package amplitude

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const (
	namespace = "amplitude"
	baseURL   = "https://amplitude.com/api/3/chart/"
)

type Exporter struct {
	mutex        sync.RWMutex
	up           prometheus.Gauge
	totalScrapes prometheus.Counter
	projects     *Projects
	client       *http.Client
	metrics      map[string]*MetricInfo
	timer        *time.Ticker
}

type MetricInfo struct {
	desc   *prometheus.Desc
	value  float64
	key    string
	acc    float64
	labels []string
}

func (mi *MetricInfo) Inc(key string, value float64, previousKey string, previousValue float64) {
	if mi.key == key {
		mi.acc = value
	} else {
		if mi.key == previousKey {
			mi.value += previousValue
		} else {
			mi.value += mi.acc
		}
		mi.key = key
		mi.acc = value
	}
}

func (mi *MetricInfo) GetValue() float64 {
	return mi.value + mi.acc
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		ch <- m.desc
	}
	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	e.totalScrapes.Inc()
	for _, metric := range e.metrics {
		cm, err := prometheus.NewConstMetric(
			metric.desc,
			prometheus.CounterValue,
			metric.GetValue(), metric.labels...)
		if err != nil {
			continue
		}
		ch <- cm
	}

	ch <- e.up
	ch <- e.totalScrapes
}

type Option func(e *Exporter)

func SetProjects(p *Projects) Option {
	return func(e *Exporter) {
		e.projects = p
		m := map[string]*MetricInfo{}
		for _, project := range *p {
			for _, chart := range project.Charts {
				m[fmt.Sprintf("%s/%s", project.Name, chart.ID)] = &MetricInfo{
					desc: prometheus.NewDesc(
						prometheus.BuildFQName(namespace, chart.Subsystem, chart.Name),
						chart.HelpString, chart.Labels, nil),
					value:  0,
					key:    "",
					acc:    0,
					labels: chart.Labels,
				}
			}
		}
		e.metrics = m
	}
}

func (e *Exporter) StartScrape() {
	e.timer = time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-e.timer.C:
				e.scrape()
			}
		}
	}()
}

func SetHTTPClient(client *http.Client) Option {
	return func(e *Exporter) {
		e.client = client
	}
}

func New(opts ...Option) *Exporter {
	e := &Exporter{
		client: http.DefaultClient,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of amplitude successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total Amplitude scrapes.",
		}),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *Exporter) scrape() {
	up := 1

	for _, project := range *e.projects {
		for _, chart := range project.Charts {
			cd, err := GetChartData(chart.ID, e.client, project.ApiId, project.ApiKey)
			if err != nil {
				log.Errorf("%s(%s): %v", project.Name, chart.ID, err)
				up = 0
				continue
			}
			if metric, found := e.metrics[fmt.Sprintf("%s/%s", project.Name, chart.ID)]; found {
				if len(cd.Data.XValues) < 2 || len(cd.Data.Series[0]) < 2 {
					up = 0
					continue
				}
				previousKey := cd.Data.XValues[len(cd.Data.XValues)-2]
				previousValue := cd.Data.Series[0][len(cd.Data.Series[0])-2].Value
				lastKey := cd.Data.XValues[len(cd.Data.XValues)-1]
				lastValue := cd.Data.Series[0][len(cd.Data.Series[0])-1].Value
				metric.Inc(lastKey, float64(lastValue), previousKey, float64(previousValue))
				log.Debugf("Receive %s[%s]=%d", metric.desc.String(), lastKey, lastValue)
			}
		}
	}

	e.up.Set(float64(up))
}

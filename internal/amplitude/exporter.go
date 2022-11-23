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

type Option func(e *Exporter)

type Exporter struct {
	mutex        sync.RWMutex
	up           prometheus.Gauge
	totalScrapes prometheus.Counter
	projects     *Projects
	client       *http.Client
	metrics      map[string]*MetricInfo
	timer        *time.Ticker
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
		m, err := metric.GetPromMetric()
		if m != nil {
			ch <- m
			if err != nil {
				log.Debugf("metric: %v, err: %v", m, err)
			}
			continue
		}
		if err != nil {
			log.Errorf("Error getting metric: %v", err)
		}
	}

	ch <- e.up
	ch <- e.totalScrapes
}

func (e *Exporter) StartScrape(interval time.Duration) {
	e.timer = time.NewTicker(interval)
	go func() {
		for {
			<-e.timer.C
			e.scrape()
		}
	}()
}

func (e *Exporter) scrape() {
	up := 1

	for _, project := range *e.projects {
		for _, chart := range project.Charts {

			// Get the chart data
			cd, err := chart.GetChartData(e.client, project.ApiId, project.ApiKey)
			if err != nil {
				log.Errorf("%s(%s): %v", project.Name, chart.ID, err)
				up = 0
				continue
			}

			// Update the metric
			if metric, found := e.metrics[fmt.Sprintf("%s/%s", project.Name, chart.ID)]; found {

				var lastKey, previousKey string
				var lastValue, previousValue float64

				if len(cd.Data.XValues) > 0 && len(cd.Data.Series[0]) > 0 {
					lastKey = cd.Data.XValues[len(cd.Data.XValues)-1]
					lastValue = cd.Data.Series[0][len(cd.Data.Series[0])-1].Value
				}

				if len(cd.Data.XValues) > 1 && len(cd.Data.Series[0]) > 1 {
					previousKey = cd.Data.XValues[len(cd.Data.XValues)-2]
					previousValue = cd.Data.Series[0][len(cd.Data.Series[0])-2].Value
				}

				switch metric.mType {
				case prometheus.GaugeValue:
					metric.Set(lastKey, lastValue)
				default:
					metric.Add(lastKey, lastValue, previousKey, previousValue)
				}
				log.Debugf("Receive %s[%s]=%f(%f)", metric.desc.String(), lastKey, lastValue, metric.GetValue())
			}
		}
	}

	e.up.Set(float64(up))
}

func SetProjects(p *Projects) Option {
	return func(e *Exporter) {
		e.projects = p
		m := map[string]*MetricInfo{}
		for _, project := range *p {
			for _, chart := range project.Charts {
				m[fmt.Sprintf("%s/%s", project.Name, chart.ID)] = chart.newMetric()
			}
		}
		e.metrics = m
	}
}

func SetHTTPClient(client *http.Client) Option {
	return func(e *Exporter) {
		e.client = client
	}
}

func New(opts ...Option) *Exporter {
	e := &Exporter{
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

	if e.client == nil {
		e.client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return e
}

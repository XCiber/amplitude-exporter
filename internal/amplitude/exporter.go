package amplitude

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

const (
	namespace = "amplitude"
	baseURL   = "https://amplitude.com/api/3/chart/"
)

var (
	amplitudeUp = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of Amplitude successful.", nil, nil)
)

type Exporter struct {
	mutex        sync.RWMutex
	up           prometheus.Gauge
	totalScrapes prometheus.Counter
	projects     *Projects
	client       *http.Client
	metrics      map[string]*prometheus.Desc
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range e.metrics {
		ch <- desc
	}
	ch <- amplitudeUp
	ch <- e.totalScrapes.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(amplitudeUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes
}

type Option func(e *Exporter)

func SetProjects(p *Projects) Option {
	return func(e *Exporter) {
		e.projects = p
		m := map[string]*prometheus.Desc{}
		for _, project := range *p {
			for _, chart := range project.Charts {
				m[fmt.Sprintf("%s/%s", project.Name, chart.ID)] = prometheus.NewDesc(
					prometheus.BuildFQName(namespace, chart.Subsystem, chart.Name),
					chart.HelpString, chart.Labels, nil)
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

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	for _, project := range *e.projects {
		for _, chart := range project.Charts {
			cd, err := GetChartData(chart.ID, e.client, project.ApiId, project.ApiKey)
			if err != nil {
				log.Errorf("%s(%s): %v", project.Name, chart.ID, err)
				continue
			}
			if desc, found := e.metrics[fmt.Sprintf("%s/%s", project.Name, chart.ID)]; found {
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					float64(cd.Data.Series[0][len(cd.Data.Series[0])-1].Value), chart.Labels...)
			}
		}
	}

	return 1
}

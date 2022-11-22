package amplitude

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type Chart struct {
	ID         string            `mapstructure:"id"`
	Labels     []string          `mapstructure:"labels"`
	Name       string            `mapstructure:"name"`
	Subsystem  string            `mapstructure:"subsystem"`
	HelpString string            `mapstructure:"helpString"`
	Type       string            `mapstructure:"type"`
	Tags       map[string]string `mapstructure:"tags"`
}

type Project struct {
	Name   string  `mapstructure:"name"`
	ApiId  string  `mapstructure:"apiId"`
	ApiKey string  `mapstructure:"apiKey"`
	Charts []Chart `mapstructure:"charts"`
}

type ChartResponse struct {
	Data struct {
		XValues []string `json:"xValues"`
		Series  [][]struct {
			SetId string  `json:"setId"`
			Value float64 `json:"value"`
		} `json:"series"`
	} `json:"data"`
}

type Projects []Project

func (c *Chart) GetChartData(client *http.Client, username string, passwd string) (ChartResponse, error) {
	cr := ChartResponse{}
	url := fmt.Sprintf("%s%s/query", baseURL, c.ID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(username, passwd)
	res, err := client.Do(req)
	if err != nil {
		return cr, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("response body close: ", err)
		}
	}(res.Body)
	bodyText, err := io.ReadAll(res.Body)
	if err != nil {
		return cr, err
	}
	err = json.Unmarshal(bodyText, &cr)
	if err != nil {
		return ChartResponse{}, err
	}
	return cr, nil
}

func (c *Chart) newMetric() *MetricInfo {
	var t prometheus.ValueType
	switch c.Type {
	case "counter":
		t = prometheus.CounterValue
	case "gauge":
		t = prometheus.GaugeValue
	default:
		t = prometheus.CounterValue
	}
	return &MetricInfo{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, c.Subsystem, c.Name),
			c.HelpString, c.Labels, c.Tags),
		value:  0,
		key:    "",
		acc:    0,
		labels: c.Labels,
		mType:  t,
	}
}

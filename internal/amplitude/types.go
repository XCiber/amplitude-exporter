package amplitude

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type Chart struct {
	ID         string
	Labels     []string
	Name       string
	Subsystem  string
	HelpString string
	Type       string
}

type Project struct {
	Name   string
	ApiId  string
	ApiKey string
	Charts []Chart
}

type ChartResponse struct {
	Data struct {
		XValues []string `json:"xValues"`
		Series  [][]struct {
			SetId string  `json:"setId"`
			Value float64 `json:"value"`
		} `json:"series"`
		SeriesCollapsed [][]struct {
			SetId string  `json:"setId"`
			Value float64 `json:"value"`
		} `json:"seriesCollapsed"`
		SeriesLabels []int `json:"seriesLabels"`
		//SeriesMeta   []struct {
		//	SegmentIndex int    `json:"segmentIndex"`
		//	EventIndex   int    `json:"eventIndex"`
		//	FormulaIndex int    `json:"formulaIndex"`
		//	Formula      string `json:"formula"`
		//} `json:"seriesMeta"`
	} `json:"data"`
	TimeComputed                       int64         `json:"timeComputed"`
	WasCached                          bool          `json:"wasCached"`
	CacheFreshness                     string        `json:"cacheFreshness"`
	NovaRuntime                        int           `json:"novaRuntime"`
	NovaRequestDuration                int           `json:"novaRequestDuration"`
	NovaCost                           int           `json:"novaCost"`
	ThrottleTime                       int           `json:"throttleTime"`
	MinSampleRate                      float64       `json:"minSampleRate"`
	TransformationIds                  []interface{} `json:"transformationIds"`
	Backend                            string        `json:"backend"`
	RealtimeDataMissing                bool          `json:"realtimeDataMissing"`
	TimedOutRealtimeData               bool          `json:"timedOutRealtimeData"`
	PartialMergedAndNewUserInformation bool          `json:"partialMergedAndNewUserInformation"`
	PrunedResult                       bool          `json:"prunedResult"`
	HitChunkGroupByLimit               bool          `json:"hitChunkGroupByLimit"`
	Subcluster                         int           `json:"subcluster"`
	MillisSinceComputed                int           `json:"millisSinceComputed"`
	QueryIds                           []string      `json:"queryIds"`
}

type Projects []Project

func GetChartData(chartId string, client *http.Client, username string, passwd string) (ChartResponse, error) {
	cr := ChartResponse{}
	url := fmt.Sprintf("%s%s/query", baseURL, chartId)
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

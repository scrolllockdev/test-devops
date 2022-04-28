package agent

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/scrolllockdev/test-devops/internal/agent/config"
	"github.com/scrolllockdev/test-devops/internal/agent/converter"
	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

func InitLogger() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

type Agent struct {
	client          *http.Client
	pollInterval    time.Duration
	reportInterval  time.Duration
	endpoint        string
	updatesEndpoint string
	key             string
}

func (a *Agent) postRequest(ctx context.Context, value []byte, endpoint string) (int, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(value))
	if err != nil {
		return 0, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	return response.StatusCode, nil
}

func (a *Agent) Init(cfg config.Config) *Agent {

	InitLogger()

	a.client = &http.Client{}

	transport := &http.Transport{}
	transport.MaxIdleConns = 60

	a.client.Transport = transport
	a.pollInterval = cfg.PollInterval
	a.reportInterval = cfg.ReportInterval
	a.endpoint = getEndpoint(cfg.ServerAddress, "update")
	a.updatesEndpoint = getEndpoint(cfg.ServerAddress, "updates")
	a.key = cfg.Key
	return a
}

func (a *Agent) Run(ctx context.Context) {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	metrics := storage.Storage{
		GaugeStorage:   make(map[string]storage.Gauge),
		CounterStorage: make(map[string]storage.Counter),
	}
	var counter storage.Counter = 0
	go func(ctx context.Context) {
		for {
			select {
			case t := <-pollTicker.C:
				var rtm runtime.MemStats
				runtime.ReadMemStats(&rtm)
				counter++
				metrics.SaveMetrics(rtm, storage.RandomValue(), counter)
				log.Infof("metrics are collected in %s\n", t)
			case <-reportTicker.C:
				metricsArray, err := converter.StorageToArray(metrics, a.key)
				if err != nil {
					log.Errorln(err)
				}
				if statusCode, err := a.postRequest(ctx, metricsArray, a.updatesEndpoint); err != nil {
					log.Errorln(err)
				} else {
					log.Infof("response status-code - %d\n", statusCode)
				}
			case t := <-ctx.Done():
				pollTicker.Stop()
				reportTicker.Stop()
				log.Infof("tickers stopped in %s\n", t)
				return
			}
		}
	}(ctx)
}

func getEndpoint(address, route string) string {

	endpoint := url.URL{
		Scheme: "http",
		Host:   address,
		Path:   path.Join(route),
	}

	return endpoint.String() + "/"
}

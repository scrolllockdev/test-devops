package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"

	"github.com/scrolllockdev/test-devops/internal/agent/config"
	"github.com/scrolllockdev/test-devops/internal/agent/converter"
	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

type Agent struct {
	client         *http.Client
	pollInterval   time.Duration
	reportInterval time.Duration
	endpoint       string
	key            string
}

func (a *Agent) postRequest(ctx context.Context, value []byte) error {

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewBuffer(value))
	if err != nil {
		fmt.Printf("error creating a new request with context %s: %v", a.endpoint, err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		fmt.Printf("error sending the request %s: %v \n", a.endpoint, err)
		return err
	}
	defer response.Body.Close()

	fmt.Printf("response status-code %s\n", response.Status)

	return nil
}

func (a *Agent) Init(cfg config.Config) *Agent {
	a.client = &http.Client{}

	transport := &http.Transport{}
	transport.MaxIdleConns = 60

	a.client.Transport = transport
	a.pollInterval = cfg.PollInterval
	a.reportInterval = cfg.ReportInterval
	a.endpoint = getEndpoint(cfg.ServerAddress)
	a.key = cfg.Key

	return a
}

func (a *Agent) Run(ctx context.Context) {
	pollTicker := time.NewTicker(a.pollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(a.reportInterval)
	defer reportTicker.Stop()

	done := make(chan bool)

	metrics := storage.Storage{
		GaugeStorage:   make(map[string]storage.Gauge),
		CounterStorage: make(map[string]storage.Counter),
	}
	var counter storage.Counter = 0

	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-pollTicker.C:
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
			counter++
			metrics.SaveMetrics(rtm, storage.RandomValue(), counter)
			fmt.Printf("metrics are collected in %s\n", t)
		case t := <-reportTicker.C:
			counter = 0
			for key, value := range metrics.GaugeStorage {
				body := converter.GaugeToJSON(key, value, a.key)
				if err := a.postRequest(ctx, body); err != nil {
					fmt.Println(err)
				}
			}
			fmt.Printf("gauge metrics are sended in %s\n", t)

			for key, value := range metrics.CounterStorage {
				body := converter.CounterToJSON(key, value, a.key)
				if err := a.postRequest(ctx, body); err != nil {
					fmt.Println(err)
				}
			}
			fmt.Printf("counter metrics are sended in %s\n", t)

		}
	}
}

func getEndpoint(address string) string {

	endpoint := url.URL{
		Scheme: "http",
		Host:   address,
		Path:   path.Join("update"),
	}

	return endpoint.String() + "/"
}

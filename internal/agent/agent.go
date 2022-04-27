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
	client          *http.Client
	pollInterval    time.Duration
	reportInterval  time.Duration
	endpoint        string
	updatesEndpoint string
	key             string
}

func (a *Agent) postRequest(ctx context.Context, value []byte, endpoint string) error {

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(value))
	if err != nil {
		fmt.Printf("error creating a new request with context %s: %v", endpoint, err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		fmt.Printf("error sending the request %s: %v \n", endpoint, err)
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
				fmt.Printf("metrics are collected in %s\n", t)
				// metricsArray, err := converter.StorageToArray(metrics, a.key)
				// if err != nil {
				// 	fmt.Println(err)
				// }
				// if err := a.postRequest(ctx, metricsArray, a.updatesEndpoint); err != nil {
				// 	fmt.Println(err)
				// }
			case t := <-reportTicker.C:
				metricsArray, err := converter.StorageToArray(metrics, a.key)
				if err != nil {
					fmt.Println(err)
				}
				if err := a.postRequest(ctx, metricsArray, a.updatesEndpoint); err != nil {
					fmt.Println(err)
				}
				// counter = 0
				// for key, value := range metrics.GaugeStorage {
				// 	body, err := converter.GaugeToJSON(key, value, a.key)
				// 	if err != nil {
				// 		fmt.Println(err)
				// 		return
				// 	}
				// 	if err := a.postRequest(ctx, body, a.endpoint); err != nil {
				// 		fmt.Println(err)
				// 	}
				// }
				// fmt.Printf("gauge metrics are sended in %s\n", t)
				// for key, value := range metrics.CounterStorage {
				// 	body, err := converter.CounterToJSON(key, value, a.key)
				// 	if err != nil {
				// 		fmt.Println(err)
				// 		return
				// 	}
				// 	if err := a.postRequest(ctx, body, a.endpoint); err != nil {
				// 		fmt.Println(err)
				// 	}
				// }
				fmt.Printf("counter metrics are sended in %s\n", t)
			case <-ctx.Done():
				pollTicker.Stop()
				reportTicker.Stop()
				fmt.Println("Tickers stopped!")
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

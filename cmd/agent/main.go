package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"time"
)

func randomValue() gauge {
	return gauge(rand.Float64())
}

func updateMetricURL(serverAddress string) string {

	resultURL := url.URL{
		Scheme: "http",
		Host:   serverAddress,
		Path:   path.Join("update"),
	}
	return resultURL.String() + "/"
}


func sendMetric(metric []byte, client *http.Client, cfg Config) error {
	endpoint := updateMetricURL(cfg.ServerAddress)
	fmt.Printf("Отправка метрик на %s\n", endpoint)
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(metric))
	if err != nil {
		fmt.Printf("Error while reading %s: %v", endpoint, err)
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error while sending request %s: %v \n", endpoint, err)
		fmt.Println(response)
		return err
	}
	defer response.Body.Close()

	fmt.Printf("Статус-код %s\n", response.Status)

	return nil

}
func convertCounterMetricToJSON(name, mType string, value counter) []byte {
	mValue := int64(value)

	metric := Metrics{
		ID:    name,
		MType: mType,
		Delta: &mValue,
	}

	metricJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal", err)
		panic(err)
	}

	return metricJSON
}
func convertGaugeMetricToJSON(name, mType string, value gauge) []byte {
	mValue := float64(value)

	metric := Metrics{
		ID:    name,
		MType: mType,
		Value: &mValue,
	}

	metricJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal", err)
		panic(err)
	}
	
	return metricJSON
}

func main() {

	var cfg Config
	err := cfg.readConfig()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(cfg)

	client := &http.Client{}
	transport := &http.Transport{}
	transport.MaxIdleConns = 60
	client.Transport = transport

	pollTicker := time.NewTicker(cfg.PollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

	done := make(chan bool)

	var metrics MetricsStorage
	metrics.counterMetricsStorage = make(map[string]counter)
	metrics.gaugeMetricsStorage = make(map[string]gauge)
	metricsCounter := 0
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-pollTicker.C:
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
			metricsCounter++
			metrics.addMetrics(rtm, randomValue(), counter(metricsCounter))
			fmt.Printf("Метрики собраны в %s\n", t)
		case <-reportTicker.C:
			metricsCounter = 0
			for name, value := range metrics.gaugeMetricsStorage {
				if err := sendMetric(convertGaugeMetricToJSON(name, "gauge", value), client, cfg); err != nil {
					fmt.Println(err, string(convertGaugeMetricToJSON(name, "gauge", value)))
				} else {
					fmt.Println("Success send metric", string(convertGaugeMetricToJSON(name, "gauge", value)))
				}
			}
			for name, value := range metrics.counterMetricsStorage {
				if err := sendMetric(convertCounterMetricToJSON(name, "counter", value), client, cfg); err != nil {
					fmt.Println("Success send metric", string(convertCounterMetricToJSON(name, "counter", value)))
					fmt.Println(err)
				}
			}
		}
	}
}

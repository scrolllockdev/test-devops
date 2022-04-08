package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func main() {

	var cfg Config
	cfg.readConfig()

	metricsStorage := &Storage{Storage: make([]Metrics, 0), config: cfg}

	if metricsStorage.config.Restore {
		err := metricsStorage.readFromDbFile()
		if err != nil {
			fmt.Println(err)
		}
	}

	if metricsStorage.config.StoreInterval != 0 {
		go metricsStorage.storeWithInterval()
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gzipHandle)

	fmt.Println(cfg)

	r.Post("/update/{type}/{name}/{value}", metricsStorage.updateMetric)
	r.Get("/value/{MetricType}/{MetricName}", metricsStorage.currentMetric)
	r.Get("/", metricsStorage.allMetrics)
	// 4 - increment
	r.Post("/update/", metricsStorage.updateMetricFromBody)
	r.Post("/value/", metricsStorage.metricValue)

	if err := http.ListenAndServe(metricsStorage.config.ServerAddress, r); err != nil {
		fmt.Println(err)
	}

}

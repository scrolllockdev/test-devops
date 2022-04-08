package main

import "runtime"

// gauge - Тип, содержащий значение метрик
type gauge float64

// counter - Тип, содержащий значение счетчика измерения метрик
type counter int64

// MetricsStorage - Структура метрик
type MetricsStorage struct {
	gaugeMetricsStorage   map[string]gauge
	counterMetricsStorage map[string]counter
}

// addMetrics - Получение текущих метрик
func (m *MetricsStorage) addMetrics(stats runtime.MemStats, randomValue gauge, counter counter) {

	m.gaugeMetricsStorage["Alloc"] = gauge(stats.Alloc)
	m.gaugeMetricsStorage["BuckHashSys"] = gauge(stats.BuckHashSys)
	m.gaugeMetricsStorage["Frees"] = gauge(stats.Frees)
	m.gaugeMetricsStorage["GCCPUFraction"] = gauge(stats.GCCPUFraction)
	m.gaugeMetricsStorage["GCSys"] = gauge(stats.GCSys)
	m.gaugeMetricsStorage["HeapAlloc"] = gauge(stats.HeapAlloc)
	m.gaugeMetricsStorage["HeapIdle"] = gauge(stats.HeapIdle)
	m.gaugeMetricsStorage["HeapInuse"] = gauge(stats.HeapInuse)
	m.gaugeMetricsStorage["HeapObjects"] = gauge(stats.HeapObjects)
	m.gaugeMetricsStorage["HeapReleased"] = gauge(stats.HeapReleased)
	m.gaugeMetricsStorage["HeapSys"] = gauge(stats.HeapSys)
	m.gaugeMetricsStorage["LastGC"] = gauge(stats.LastGC)
	m.gaugeMetricsStorage["Lookups"] = gauge(stats.Lookups)
	m.gaugeMetricsStorage["MCacheInuse"] = gauge(stats.MCacheInuse)
	m.gaugeMetricsStorage["MCacheSys"] = gauge(stats.MSpanInuse)
	m.gaugeMetricsStorage["MSpanInuse"] = gauge(stats.MSpanInuse)
	m.gaugeMetricsStorage["MSpanSys"] = gauge(stats.MSpanSys)
	m.gaugeMetricsStorage["Mallocs"] = gauge(stats.Mallocs)
	m.gaugeMetricsStorage["NextGC"] = gauge(stats.NextGC)
	m.gaugeMetricsStorage["NumForcedGC"] = gauge(stats.NumForcedGC)
	m.gaugeMetricsStorage["NumGC"] = gauge(stats.NumGC)
	m.gaugeMetricsStorage["OtherSys"] = gauge(stats.OtherSys)
	m.gaugeMetricsStorage["PauseTotalNs"] = gauge(stats.PauseTotalNs)
	m.gaugeMetricsStorage["StackInuse"] = gauge(stats.StackInuse)
	m.gaugeMetricsStorage["StackSys"] = gauge(stats.StackSys)
	m.gaugeMetricsStorage["Sys"] = gauge(stats.Sys)
	m.gaugeMetricsStorage["TotalAlloc"] = gauge(stats.TotalAlloc)
	m.counterMetricsStorage["PollCount"] = counter
	m.gaugeMetricsStorage["RandomValue"] = randomValue
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

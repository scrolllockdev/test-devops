package storage

import (
	"math/rand"
	"runtime"
)

type Gauge float64
type Counter int64

func RandomValue() Gauge {
	return Gauge(rand.Float64())
}

type Storage struct {
	GaugeStorage   map[string]Gauge
	CounterStorage map[string]Counter
}

func (s *Storage) SaveMetrics(stats runtime.MemStats, randomValue Gauge, counter Counter) {

	s.GaugeStorage["Alloc"] = Gauge(stats.Alloc)
	s.GaugeStorage["BuckHashSys"] = Gauge(stats.BuckHashSys)
	s.GaugeStorage["Frees"] = Gauge(stats.Frees)
	s.GaugeStorage["GCCPUFraction"] = Gauge(stats.GCCPUFraction)
	s.GaugeStorage["GCSys"] = Gauge(stats.GCSys)
	s.GaugeStorage["HeapAlloc"] = Gauge(stats.HeapAlloc)
	s.GaugeStorage["HeapIdle"] = Gauge(stats.HeapIdle)
	s.GaugeStorage["HeapInuse"] = Gauge(stats.HeapInuse)
	s.GaugeStorage["HeapObjects"] = Gauge(stats.HeapObjects)
	s.GaugeStorage["HeapReleased"] = Gauge(stats.HeapReleased)
	s.GaugeStorage["HeapSys"] = Gauge(stats.HeapSys)
	s.GaugeStorage["LastGC"] = Gauge(stats.LastGC)
	s.GaugeStorage["Lookups"] = Gauge(stats.Lookups)
	s.GaugeStorage["MCacheInuse"] = Gauge(stats.MCacheInuse)
	s.GaugeStorage["MCacheSys"] = Gauge(stats.MSpanInuse)
	s.GaugeStorage["MSpanInuse"] = Gauge(stats.MSpanInuse)
	s.GaugeStorage["MSpanSys"] = Gauge(stats.MSpanSys)
	s.GaugeStorage["Mallocs"] = Gauge(stats.Mallocs)
	s.GaugeStorage["NextGC"] = Gauge(stats.NextGC)
	s.GaugeStorage["NumForcedGC"] = Gauge(stats.NumForcedGC)
	s.GaugeStorage["NumGC"] = Gauge(stats.NumGC)
	s.GaugeStorage["OtherSys"] = Gauge(stats.OtherSys)
	s.GaugeStorage["PauseTotalNs"] = Gauge(stats.PauseTotalNs)
	s.GaugeStorage["StackInuse"] = Gauge(stats.StackInuse)
	s.GaugeStorage["StackSys"] = Gauge(stats.StackSys)
	s.GaugeStorage["Sys"] = Gauge(stats.Sys)
	s.GaugeStorage["TotalAlloc"] = Gauge(stats.TotalAlloc)
	s.GaugeStorage["RandomValue"] = randomValue
	s.CounterStorage["PollCount"] = counter

}

func (s *Storage) getGauges() map[string]Gauge {
	return s.GaugeStorage
}

func (s *Storage) getCounters() map[string]Counter {
	return s.CounterStorage
}

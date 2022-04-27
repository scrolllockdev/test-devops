package converter

import (
	"encoding/json"
	"fmt"

	"github.com/scrolllockdev/test-devops/internal/agent/hash"
	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func StorageToArray(storage storage.Storage, secretKey string) ([]byte, error) {
	s := []Metrics{}
	for key, value := range storage.GaugeStorage {
		gVal := float64(value)
		gMetric := Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: &gVal,
			Hash:  hash.Hash(fmt.Sprintf("%s:gauge:%f", key, value), key),
		}
		s = append(s, gMetric)
	}
	for key, value := range storage.CounterStorage {
		cVal := int64(value)
		cMetric := Metrics{
			ID:    key,
			MType: "counter",
			Delta: &cVal,
			Value: nil,
			Hash:  hash.Hash(fmt.Sprintf("%s:counter:%d", key, cVal), key),
		}
		s = append(s, cMetric)
	}

	metrcisArray, err := json.Marshal(s)
	if err != nil {
		fmt.Println("something wrong with marshal metrics array", err)
		return nil, err
	}

	return metrcisArray, nil
}

func GaugeToJSON(name string, value storage.Gauge, key string) ([]byte, error) {

	gValue := float64(value)
	metric := Metrics{
		ID:    name,
		MType: "gauge",
		Value: &gValue,
	}

	if key != "" {
		metric.Hash = hash.Hash(fmt.Sprintf("%s:gauge:%f", name, value), key)
	}

	gaugeJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal gauge value", err)
		return nil, err
	}
	return gaugeJSON, nil
}

func CounterToJSON(name string, value storage.Counter, key string) ([]byte, error) {
	mValue := int64(value)

	metric := Metrics{
		ID:    name,
		MType: "counter",
		Delta: &mValue,
	}

	if key != "" {
		metric.Hash = hash.Hash(fmt.Sprintf("%s:counter:%d", name, mValue), key)
	}

	counterJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal counter value", err)
		return nil, err
	}

	return counterJSON, nil
}

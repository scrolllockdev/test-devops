package converter

import (
	"encoding/json"
	"fmt"

	"github.com/scrolllockdev/test-devops/internal/agent/hash"
	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func gaugeHashModel(name string, value float64) string {
	return fmt.Sprintf("%s:gauge:%f", name, value)
}

func counterHashModel(name string, value int64) string {
	return fmt.Sprintf("%s:counter:%d", name, value)
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
			Hash:  hash.Hash(gaugeHashModel(key, gVal), secretKey),
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
			Hash:  hash.Hash(counterHashModel(key, cVal), secretKey),
		}
		s = append(s, cMetric)
	}
	metrcisArray, err := json.Marshal(s)
	if err != nil {
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
		metric.Hash = hash.Hash(gaugeHashModel(name, gValue), key)
	}

	gaugeJSON, err := json.Marshal(metric)
	if err != nil {
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
		metric.Hash = hash.Hash(counterHashModel(name, mValue), key)
	}

	counterJSON, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	return counterJSON, nil
}

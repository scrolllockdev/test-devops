package converter

import (
	"encoding/json"
	"fmt"

	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func GaugeToJSON(name string, value storage.Gauge) []byte {

	gValue := float64(value)
	metric := Metrics{
		ID:    name,
		MType: "gauge",
		Value: &gValue,
	}

	gaugeJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal gauge value", err)
		panic(err)
	}

	return gaugeJSON
}

func CounterToJSON(name string, value storage.Counter) []byte {
	mValue := int64(value)

	metric := Metrics{
		ID:    name,
		MType: "counter",
		Delta: &mValue,
	}

	counterJSON, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("something wrong with marshal counter value", err)
		panic(err)
	}

	return counterJSON
}
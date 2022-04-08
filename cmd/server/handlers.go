package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type Storage struct {
	Storage []Metrics `json:"Storage"`
	config  Config
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (storage *Storage) updateMetric(w http.ResponseWriter, r *http.Request) {

	stat := strings.Split(r.URL.String(), "/")

	resp := make(map[string]string)
	resp["error"] = "Something wrong"
	jsonResp, _ := json.Marshal(resp)

	w.Header().Set("content-type", "text/plain")

	if len(stat) != 5 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(jsonResp)
		return
	} else {
		switch stat[2] {
		case "gauge":
			value, err := strconv.ParseFloat(stat[4], 64)
			fmt.Print(err)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonResp)
				return
			}

			metricIsUpdated := false
			for index, item := range storage.Storage {
				if item.ID == stat[3] {
					storage.Storage[index].Value = &value
					metricIsUpdated = true
					break
				}
			}

			if !metricIsUpdated {

				metric := Metrics{
					ID:    stat[3],
					MType: "gauge",
					Value: &value,
				}

				storage.Storage = append(storage.Storage, metric)
			}

			w.WriteHeader(http.StatusOK)
			w.Write(jsonResp)
		case "counter":
			value, errorCounter := strconv.ParseInt(stat[4], 10, 64)
			if errorCounter != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonResp)
				return
			}

			metricIsUpdated := false
			for index, item := range storage.Storage {
				if item.ID == stat[3] {
					delta := *storage.Storage[index].Delta
					delta += value
					storage.Storage[index].Delta = &delta
					metricIsUpdated = true
					break
				}
			}

			if !metricIsUpdated {

				metric := Metrics{
					ID:    stat[3],
					MType: "counter",
					Delta: &value,
				}
				storage.Storage = append(storage.Storage, metric)
			}

			w.WriteHeader(http.StatusOK)
			w.Write(jsonResp)
			return
		default:
			w.WriteHeader(http.StatusNotImplemented)
			w.Write(jsonResp)
		}
	}
	if storage.config.StoreInterval == 0 && storage.config.StorePath != "" {
		fmt.Println("store!!")
		storeToFile(*storage)
	}
	return
}

func (storage *Storage) updateMetricFromBody(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content type", http.StatusInternalServerError)
	}

	var metric Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case "gauge":
		metricIsUpdated := false
		for index, item := range storage.Storage {
			if item.ID == metric.ID {
				storage.Storage[index].Value = metric.Value
				metricIsUpdated = true
				break
			}
		}
		if !metricIsUpdated {
			storage.Storage = append(storage.Storage, metric)
		}
		w.WriteHeader(http.StatusOK)
	case "counter":
		metricIsUpdated := false
		for index, item := range storage.Storage {
			if item.ID == metric.ID {
				delta := *storage.Storage[index].Delta
				delta += *metric.Delta
				storage.Storage[index].Delta = &delta
				metricIsUpdated = true
				break
			}
		}
		if !metricIsUpdated {
			storage.Storage = append(storage.Storage, metric)
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "500 not implement", http.StatusNotImplemented)
		return
	}
	if storage.config.StoreInterval == 0 && storage.config.StorePath != "" {
		storeToFile(*storage)
	}
	return
}

func (storage *Storage) currentMetric(w http.ResponseWriter, r *http.Request) {

	stat := strings.Split(r.URL.String(), "/")

	resp := make(map[string]string)
	resp["error"] = "Something wrong"
	jsonResp, _ := json.Marshal(resp)

	w.Header().Set("content-type", "text/plain")

	if len(stat) != 4 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(jsonResp)
		return
	} else {
		switch stat[2] {
		case "gauge":
			for _, item := range storage.Storage {
				if stat[3] == item.ID {
					_, err := w.Write([]byte(fmt.Sprintf("%v", *item.Value)))
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write(jsonResp)
						return
					}
					return
				}
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonResp)
			return
		case "counter":
			metricIndex := 0
			metricFounded := false
			for index, item := range storage.Storage {
				if item.ID == stat[3] {
					metricIndex = index
					metricFounded = true
					break
				}
			}
			if !metricFounded {
				w.WriteHeader(http.StatusNotFound)
				w.Write(jsonResp)
				return
			}
			_, err := w.Write([]byte(strconv.FormatInt(*storage.Storage[metricIndex].Delta, 10)))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonResp)
				return
			}
			return
		}
	}
}

func (storage *Storage) metricValue(w http.ResponseWriter, r *http.Request) {

	resp := make(map[string]string)
	resp["error"] = "Something wrong"
	jsonResp, _ := json.Marshal(resp)

	w.Header().Set("application-type", "text/plain")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResp)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metric Metrics
	if err := json.Unmarshal(b, &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResp)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricFounded := false
	for _, item := range storage.Storage {
		if item.ID == metric.ID {
			metricJSON, err := json.Marshal(item)
			if err != nil {
				fmt.Println(err)
			}
			_, err = w.Write(metricJSON)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write(jsonResp)
				//http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			metricFounded = true
			break
		}
	}
	if !metricFounded {
		w.WriteHeader(http.StatusNotFound)
		w.Write(jsonResp)
		//http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func (storage *Storage) allMetrics(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	tmp, err := template.ParseFiles(path.Join(pwd, "currentMetrics.html"))
	if err != nil {
		http.Error(w, "501 server error", http.StatusInternalServerError)
		return
	}

	err = tmp.Execute(w, storage.Storage)
	if err != nil {
		http.Error(w, "500 server error", http.StatusInternalServerError)
		return
	}
}

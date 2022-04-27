package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/scrolllockdev/test-devops/internal/model"
	"github.com/scrolllockdev/test-devops/internal/server/database"

	_ "github.com/lib/pq"

	"github.com/scrolllockdev/test-devops/internal/server/config"
	"github.com/scrolllockdev/test-devops/internal/server/middlewares"
	s "github.com/scrolllockdev/test-devops/internal/server/storage"
)

func UpdateMetrics(storage *s.Storage, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid content type", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var metrics []model.Metric
		if err := json.Unmarshal(body, &metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, metric := range metrics {
			if metric.MType == "gauge" {
				metric.Delta = nil
				storage.UpdateGaugeStorage(metric.ID, *metric.Value)
			} else if metric.MType == "counter" {
				metric.Value = nil
				storage.UpdateCounterStorage(metric.ID, *metric.Delta)
			} else {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
		}
		if err = database.MultiplyUpdates(r.Context(), db, metrics, "storage"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func UpdateMetricFromAddress(storage *s.Storage, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stat := strings.Split(r.URL.String(), "/")
		if len(stat) != 5 {
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}
		switch stat[2] {
		case "gauge":
			value, err := strconv.ParseFloat(stat[4], 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			storage.UpdateGaugeStorage(stat[3], value)
		case "counter":
			value, err := strconv.ParseInt(stat[4], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			storage.UpdateCounterStorage(stat[3], value)
		default:
			http.Error(w, "not implemented", http.StatusNotImplemented)
			return
		}
		if config.StoreInterval == 0 && config.StoragePath != "" {
			if err := storage.StoreToFile(config.StoragePath); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMetricFromAddress(storage *s.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		stat := strings.Split(r.URL.String(), "/")
		if len(stat) != 4 {
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}
		switch stat[2] {
		case "gauge":
			for _, item := range storage.Storage {
				if stat[3] == item.ID && stat[2] == item.MType {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(fmt.Sprintf("%v", *item.Value)))
					return
				}
			}
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		case "counter":
			metricIndex := 0
			for index, item := range storage.Storage {
				if item.ID == stat[3] && item.MType == stat[2] {
					metricIndex = index
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(strconv.FormatInt(*storage.Storage[metricIndex].Delta, 10)))
					return
				}
			}
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		default:
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}
	}
}

func AllMetrics(storage *s.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept-Encoding", "gzip")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		for index := range storage.Storage {
			item := storage.Storage[index]
			val := ""
			if item.MType == "gauge" {
				val = strconv.FormatFloat(*item.Value, 'e', -1, 64)
			} else {
				val = strconv.FormatInt(*item.Delta, 10)
			}

			metric := fmt.Sprintf("%s - %s - %s<br>", item.ID, item.MType, val)
			io.WriteString(w, "<html><body>"+metric+"</body></html>")
		}
	}
}

func UpdateMetricFromBody(storage *s.Storage, config config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid content type", http.StatusInternalServerError)
			return
		}
		var metric model.Metric
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			storage.UpdateGaugeStorage(metric.ID, *metric.Value)
			err := database.InsertGaugeValueToTable(r.Context(), db, metric, "storage")
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(http.StatusOK)
		case "counter":
			if metric.Delta == nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			storage.UpdateCounterStorage(metric.ID, *metric.Delta)
			err := database.InsertCounterValueToTable(r.Context(), db, metric, "storage")
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "not implemented", http.StatusNotImplemented)
			return
		}
		if config.StoreInterval == 0 && config.StoragePath != "" {
			if err := storage.StoreToFile(config.StoragePath); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func GetMetricValueFromBody(storage *s.Storage, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var metric model.Metric
		if err := json.Unmarshal(b, &metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, item := range storage.Storage {
			if item.ID == metric.ID && item.MType == metric.MType {
				if len(key) != 0 {
					hash, err := middlewares.Hash(item, key)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					item.Hash = hash
				}
				metricJSON, err := json.Marshal(item)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(metricJSON)
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func PingDB(databaseDSN string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if len(databaseDSN) == 0 {
			http.Error(w, "database dsn is empty", http.StatusInternalServerError)
			return
		}

		db, err := sql.Open("postgres", databaseDSN)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

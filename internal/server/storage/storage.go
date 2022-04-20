package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	mw "github.com/scrolllockdev/test-devops/internal/server/middlewares"
	"github.com/scrolllockdev/test-devops/internal/server/model"
)

type Storage struct {
	Storage []model.Metric `json:"Storage"`
}

func (s *Storage) DirExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Storage) CreateDir() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Mkdir(path.Join(pwd, "tmp"), 0755)
	return nil
}

func (s *Storage) UpdateMetric(stat []string) (error, int) {
	if len(stat) != 5 {
		return errors.New("page not found"), http.StatusNotFound
	} else {
		switch stat[2] {
		case "gauge":
			value, err := strconv.ParseFloat(stat[4], 64)
			if err != nil {
				return err, http.StatusBadRequest
			}
			metricIsUpdated := false
			for index, item := range s.Storage {
				if item.ID == stat[3] && item.MType == stat[2] {
					s.Storage[index].Value = &value
					metricIsUpdated = true
					break
				}
			}

			if !metricIsUpdated {
				metric := model.Metric{
					ID:    stat[3],
					MType: "gauge",
					Value: &value,
				}

				s.Storage = append(s.Storage, metric)
			}

			return nil, http.StatusOK
		case "counter":
			value, err := strconv.ParseInt(stat[4], 10, 64)
			if err != nil {
				return err, http.StatusBadRequest
			}
			metricIsUpdated := false
			for index, item := range s.Storage {
				if item.ID == stat[3] && item.MType == stat[2] {
					delta := *s.Storage[index].Delta
					delta += value
					s.Storage[index].Delta = &delta
					metricIsUpdated = true
					break
				}
			}
			if !metricIsUpdated {
				metric := model.Metric{
					ID:    stat[3],
					MType: "counter",
					Delta: &value,
				}
				s.Storage = append(s.Storage, metric)
			}
			return nil, http.StatusOK
		default:
			return errors.New("status not implemented"), http.StatusNotImplemented
		}
	}
}

func (s *Storage) GetMetric(stat []string) (error, int, []byte) {
	if len(stat) != 4 {
		return errors.New("page not found"), http.StatusNotFound, nil
	} else {
		switch stat[2] {
		case "gauge":
			for _, item := range s.Storage {
				if stat[3] == item.ID && stat[2] == item.MType {
					return nil, http.StatusOK, []byte(fmt.Sprintf("%v", *item.Value))
				}
			}
			return errors.New("metric not found"), http.StatusNotFound, nil
		case "counter":
			metricIndex := 0
			metricFounded := false
			for index, item := range s.Storage {
				if item.ID == stat[3] && item.MType == stat[2] {
					metricIndex = index
					metricFounded = true
					break
				}
			}
			if !metricFounded {
				return errors.New("metric not found"), http.StatusNotFound, nil
			}
			return nil, http.StatusOK, []byte(strconv.FormatInt(*s.Storage[metricIndex].Delta, 10))
		default:
			return errors.New("metric not found"), http.StatusNotFound, nil
		}
	}
}

func (s *Storage) UpdateMetricFromRequest(r *http.Request) (error, int) {

	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("invalid content type"), http.StatusInternalServerError
	}

	var metric model.Metric

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		return err, http.StatusBadRequest
	}

	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return errors.New("bad request"), http.StatusBadRequest
		}
		metricIsUpdated := false
		for index, item := range s.Storage {
			if item.ID == metric.ID && item.MType == metric.MType {
				s.Storage[index].Value = metric.Value
				s.Storage[index].IncomingHash = metric.Hash
				metricIsUpdated = true
				break
			}
		}
		if !metricIsUpdated {
			s.Storage = append(s.Storage, metric)
		}
		return nil, http.StatusOK
	case "counter":
		if metric.Delta == nil {
			return errors.New("bad request"), http.StatusBadRequest
		}
		metricIsUpdated := false
		for index, item := range s.Storage {
			if item.ID == metric.ID && item.MType == metric.MType {
				delta := *s.Storage[index].Delta
				delta += *metric.Delta
				s.Storage[index].Delta = &delta
				s.Storage[index].IncomingHash = metric.Hash
				metricIsUpdated = true
				break
			}
		}
		if !metricIsUpdated {
			s.Storage = append(s.Storage, metric)
		}
		return nil, http.StatusOK
	default:
		return errors.New("not implemented"), http.StatusNotImplemented
	}
}

func (s *Storage) GetValueMetricFromBody(r *http.Request, key string) (error, int, []byte) {

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err, http.StatusBadRequest, nil
	}
	defer r.Body.Close()

	var metric model.Metric
	if err := json.Unmarshal(b, &metric); err != nil {
		return err, http.StatusBadRequest, nil
	}

	for _, item := range s.Storage {
		if item.ID == metric.ID && item.MType == metric.MType {
			if len(key) != 0 {
				fmt.Println(item)
				hash, err := mw.Hash(item, key)
				if err != nil {
					return err, http.StatusBadRequest, nil
				}
				item.Hash = hash
			}
			metricJSON, err := json.Marshal(item)
			if err != nil {
				return err, http.StatusInternalServerError, nil
			}
			return nil, http.StatusOK, metricJSON
		}
	}
	return errors.New("not found"), http.StatusNotFound, nil
}

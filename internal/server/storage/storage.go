package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/scrolllockdev/test-devops/internal/model"
)

type Storage struct {
	Storage []model.Metric `json:"storage"`
}

func (storage *Storage) RestoreFromFile(storePath string) error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := storage.DirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		fmt.Println("nothing to restore")
		return nil
	}
	file, err := os.OpenFile(path.Join(pwd, storePath), os.O_RDONLY, 0755)
	var buf bytes.Buffer
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				buf.Write(line)
				break
			} else {
				return err
			}
		}
		buf.Write(line)
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf.Bytes(), &storage)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println("metrics restored")
	return nil
}

func (storage *Storage) StoreToFile(storePath string) error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := storage.DirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		if err := storage.CreateDir(); err != nil {
			return err
		}
	}
	data, _ := json.MarshalIndent(storage, "", "  ")
	data = append(data, '\n')
	file, err := os.OpenFile(path.Join(pwd, storePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func (storage *Storage) DirExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (storage *Storage) CreateDir() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Mkdir(path.Join(pwd, "tmp"), 0755)
	return nil
}

func (storage *Storage) UpdateCounterStorage(id string, value int64) {
	metricType := "counter"
	metricIsUpdated := false
	for index, item := range storage.Storage {
		if item.ID == id && item.MType == metricType {
			delta := *storage.Storage[index].Delta
			delta += value
			storage.Storage[index].Delta = &delta
			metricIsUpdated = true
			break
		}
	}
	if !metricIsUpdated {
		metric := model.Metric{
			ID:    id,
			MType: metricType,
			Delta: &value,
		}
		storage.Storage = append(storage.Storage, metric)
	}
}

func (storage *Storage) UpdateGaugeStorage(id string, value float64) {
	metricType := "gauge"
	metricIsUpdated := false
	for index, item := range storage.Storage {
		if item.ID == id && item.MType == metricType {
			storage.Storage[index].Value = &value
			metricIsUpdated = true
			break
		}
	}
	if !metricIsUpdated {
		metric := model.Metric{
			ID:    id,
			MType: metricType,
			Value: &value,
		}
		storage.Storage = append(storage.Storage, metric)
	}
}

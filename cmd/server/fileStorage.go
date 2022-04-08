package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

func tmpDirExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createTmpDir() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Mkdir(path.Join(pwd, "tmp"), 0755)
	return nil
}

func storeToFile(storage Storage) error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := tmpDirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		createTmpDir()
	}

	data, _ := json.MarshalIndent(storage, "", "  ")

	data = append(data, '\n')

	file, err := os.OpenFile(path.Join(pwd, storage.config.StorePath), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

func (storage *Storage) readFromDbFile() error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := tmpDirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		return errors.New("nothing to restore")
	}
	file, err := os.OpenFile(path.Join(pwd, storage.config.StorePath), os.O_RDONLY, 0755)
	var buf bytes.Buffer
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				buf.Write(line)
				break // end of the input
			} else {
				fmt.Println(err) // something bad happened
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
	return nil
}

func (storage *Storage) storeWithInterval() error {
	tickerStorage := time.NewTicker(storage.config.StoreInterval)
	defer tickerStorage.Stop()
	done := make(chan bool)
	for {
		select {
		case <-done:
			storeToFile(*storage)
			return nil
		case <-tickerStorage.C:
			fmt.Println("Метрики сохранены")
			storeToFile(*storage)
		}
	}
	return nil
}

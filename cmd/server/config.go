package main

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StorePath     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func (config *Config) readConfig() error {

	flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "server address")
	flag.BoolVar(&config.Restore, "r", true, "restore from db file")
	flag.DurationVar(&config.StoreInterval, "i", 300*time.Second, "store interval")
	flag.StringVar(&config.StorePath, "f", "tmp/devops-metrics-db.json", "path to db file")
	flag.Parse()

	serverAddress, exist := os.LookupEnv("ADDRESS")
	if exist {
		config.ServerAddress = serverAddress
	}

	storePath, exist := os.LookupEnv("STORE_FILE")
	if exist {
		config.StorePath = storePath
	}

	restore, exist := os.LookupEnv("RESTORE")
	if exist {
		boolVal, err := strconv.ParseBool(restore)
		if err != nil {
			os.Exit(1)
		}
		config.Restore = boolVal
	}
	duration, exist := os.LookupEnv("STORE_INTERVAL")
	if exist {
		durationVal, err := time.ParseDuration(duration)
		if err != nil {
			os.Exit(2)
		}
		config.StoreInterval = durationVal
	}

	return nil
}

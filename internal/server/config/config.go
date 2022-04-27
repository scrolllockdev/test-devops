package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoragePath   string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Shutdown      time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s"`
	Key           string        `env:"KEY"`
	DatabaseDsn   string        `env:"DATABASE_DSN"`
}

func (config *Config) ReadConfig() error {

	flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "server address")
	flag.BoolVar(&config.Restore, "r", true, "restore from db file")
	flag.DurationVar(&config.StoreInterval, "i", 300*time.Second, "store interval")
	flag.StringVar(&config.StoragePath, "f", "tmp/devops-metrics-db.json", "path to db file")
	flag.StringVar(&config.Key, "k", "", "key for sha256")
	flag.StringVar(&config.DatabaseDsn, "d", "", "database dsn")
	flag.Parse()

	databaseDsn, exist := os.LookupEnv("DATABASE_DSN")
	if exist {
		config.DatabaseDsn = databaseDsn
	}

	key, exist := os.LookupEnv("KEY")
	if exist {
		config.Key = key
	}

	serverAddress, exist := os.LookupEnv("ADDRESS")
	if exist {
		config.ServerAddress = serverAddress
	}

	storePath, exist := os.LookupEnv("STORE_FILE")
	if exist {
		config.StoragePath = storePath
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

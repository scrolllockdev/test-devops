package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	ServerAddress  string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

func (config *Config) ReadConfig() error {

	flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "server address")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "poll interval")
	flag.Parse()

	serverAddress, exist := os.LookupEnv("ADDRESS")
	if exist {
		config.ServerAddress = serverAddress
	}

	reportDuration, exist := os.LookupEnv("REPORT_INTERVAL")
	if exist {
		durationVal, err := time.ParseDuration(reportDuration)
		if err != nil {
			return err
		}
		config.ReportInterval = durationVal
	}

	pollDuration, exist := os.LookupEnv("POLL_INTERVAL")
	if exist {
		durationVal, err := time.ParseDuration(pollDuration)
		if err != nil {
			return err
		}
		config.PollInterval = durationVal
	}

	return nil
}

package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	App     AppConfig     `yaml:"app"`
	Metrics MetricsConfig `yaml:"metrics"`
}

type AppConfig struct {
	Port           int    `yaml:"port" env:"PORT" env-default:"8888"`
	LogLevel       string `yaml:"log_level" env:"LOG_LEVEL" env-default:"debug"`
	GoogleJsonPath string `yaml:"google_json_path" env:"GOOGLE_JSON_PATH" env-default:"./google.json"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED" env-default:"false"`
	Host    string `yaml:"host" env:"METRICS_HOST" env-default:"localhost"`
	Port    int    `yaml:"port" env:"METRICS_PORT" env-default:"8888"`
}

const (
	flagConfigPathName = "config"
	envConfigPathName  = "CONFIG_PATH"
)

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		var configPath string
		flag.StringVar(&configPath, flagConfigPathName, "", "path to the config file")
		flag.Parse()

		if path, ok := os.LookupEnv(envConfigPathName); ok {
			configPath = path
		}

		instance = &Config{}

		if configPath != "" {
			if readErr := cleanenv.ReadConfig(configPath, instance); readErr != nil {
				description, descrErr := cleanenv.GetDescription(instance, nil)
				if descrErr != nil {
					panic(descrErr)
				}
				slog.Info(description)
				slog.Error("failed to read config", slog.String("apperrors", readErr.Error()), slog.String("path", configPath))
				os.Exit(1)
			}
		} else {
			err := cleanenv.ReadEnv(instance)
			if err != nil {
				slog.Error("Failed to apply environment variables", slog.String("apperrors", err.Error()))
				os.Exit(1)
			}
		}

	})
	return instance
}

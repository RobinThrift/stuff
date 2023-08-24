package config

import "os"

type Config struct {
	Addr string `json:"address"`

	Database Database `json:"database"`

	LogLevel  string `json:"logLevel"`
	LogFormat string `json:"logFormat"`
}

type Database struct {
	Path string `json:"path"`
}

func NewConfigFromEnv() (*Config, error) {
	return &Config{
		Addr: getEnvDefault("STUFF_ADDRESS", ":8080"),
	}, nil
}

func getEnvDefault(key string, d string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	return v
}

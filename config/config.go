package config

type Config struct {
	Database Database `json:"database"`

	LogLevel  string `json:"logLevel"`
	LogFormat string `json:"logFormat"`
}

type Database struct {
	Path string `json:"path"`
}

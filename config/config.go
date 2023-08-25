package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr string `json:"address"`

	Database Database `json:"database"`

	Auth Auth `json:"auth"`

	LogLevel  string `json:"logLevel"`
	LogFormat string `json:"logFormat"`
}

type Database struct {
	Path string `json:"path"`
}

type Auth struct {
	Local LocalAuth `json:"local"`
}

type LocalAuth struct {
	InitialAdminPassword string       `json:"initialAdminPassword"`
	Argon2Params         Argon2Params `json:"argon2"`
}

type Argon2Params struct {
	KeyLen  uint32 `json:"keyLen"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	Time    uint32 `json:"time"`
	Version int    `json:"version"`
}

func NewConfigFromEnv() (*Config, error) {
	return &Config{
		Addr: getEnvDefault("STUFF_ADDRESS", ":8080"),

		Database: Database{
			Path: getEnvDefault("STUFF_DATABASE_PATH", "stuff.db"),
		},

		LogLevel:  getEnvDefault("STUFF_LOG_LEVEL", "info"),
		LogFormat: getEnvDefault("STUFF_LOG_FORMAT", "json"),

		Auth: Auth{
			Local: LocalAuth{
				InitialAdminPassword: getEnvDefault("STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD", ""),
				// Recommended setting can be found at https://tools.ietf.org/html/draft-irtf-cfrg-argon2-03#section-4
				Argon2Params: Argon2Params{
					KeyLen:  uint32(getEnvIntDefault("STUFF_AUTH_LOCAL_ARGON2_KEYLEN", 32)),
					Memory:  uint32(getEnvIntDefault("STUFF_AUTH_LOCAL_ARGON2_MEMORY", 131072)), // 4GiB
					Threads: uint8(getEnvIntDefault("STUFF_AUTH_LOCAL_ARGON2_THREADS", 4)),
					Time:    uint32(getEnvIntDefault("STUFF_AUTH_LOCAL_ARGON2_TIME", 1)),
					Version: 0x13, // constant
				},
			},
		},
	}, nil
}

func getEnvDefault(key string, d string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	return v
}

func getEnvIntDefault(key string, d int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return d
	}

	return i
}

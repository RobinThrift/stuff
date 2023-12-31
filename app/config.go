package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr string `json:"address"`

	BaseURL string `json:"baseURL"`

	UseSecureCookies bool

	Database Database `json:"database"`
	FileDir  string   `json:"fileDir"`
	TmpDir   string

	TagAlgorithm string `json:"tagAlgorithm"`

	DefaultCurrency  string `json:"defaultCurrency"`
	DecimalSeparator string `json:"decimalSeparator"`

	Auth Auth `json:"auth"`

	LogLevel  string `json:"logLevel"`
	LogFormat string `json:"logFormat"`
}

type Database struct {
	Path      string        `json:"path"`
	Timeout   time.Duration `json:"timeout"`
	EnableWAL bool          `json:"enableWAL"`
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
	addr := getEnvDefault("STUFF_ADDRESS", ":8080")
	defaultBaseURL := addr
	if defaultBaseURL[0] == ':' {
		defaultBaseURL = "localhost" + defaultBaseURL
	}
	baseURL := getEnvDefault("STUFF_BASE_URL", defaultBaseURL)

	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	return &Config{
		Addr: addr,

		BaseURL:          baseURL,
		UseSecureCookies: getEnvBoolDefault("STUFF_USE_SECURE_COOKIES", true),

		Database: Database{
			Path:      getEnvDefault("STUFF_DATABASE_PATH", "stuff.db"),
			EnableWAL: getEnvBoolDefault("STUFF_DATABASE_ENABLE_WAL", false),
			Timeout:   getEnvDurationDefault("STUFF_DATABASE_TIMEOUT", time.Millisecond*500),
		},

		FileDir: getEnvDefault("STUFF_FILE_DIR", "files"),
		TmpDir:  getEnvDefault("STUFF_TMP_DIR", getEnvDefault("TMPDIR", "")),

		TagAlgorithm: getEnvDefault("STUFF_TAG_ALGORITHM", "nanoid"),

		DefaultCurrency:  getEnvDefault("STUFF_DEFAULT_CURRENCY", "EUR"),
		DecimalSeparator: getEnvDefault("STUFF_DECIMAL_SEPARATOR", ","),

		Auth: Auth{
			Local: LocalAuth{
				InitialAdminPassword: getEnvDefault("STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD", ""),
				// Recommended settings can be found at https://tools.ietf.org/html/draft-irtf-cfrg-argon2-03#section-4
				Argon2Params: Argon2Params{
					KeyLen:  getEnvUint32Default("STUFF_AUTH_LOCAL_ARGON2_KEYLEN", 32),
					Memory:  getEnvUint32Default("STUFF_AUTH_LOCAL_ARGON2_MEMORY", 131072), // 4GiB
					Threads: getEnvUint8Default("STUFF_AUTH_LOCAL_ARGON2_THREADS", 4),
					Time:    getEnvUint32Default("STUFF_AUTH_LOCAL_ARGON2_TIME", 1),
					Version: 0x13, // constant
				},
			},
		},

		LogLevel:  getEnvDefault("STUFF_LOG_LEVEL", "info"),
		LogFormat: getEnvDefault("STUFF_LOG_FORMAT", "json"),
	}, nil
}

func getEnvDefault(key string, d string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	return v
}

func getEnvUint8Default(key string, d uint8) uint8 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	i, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return d
	}

	return uint8(i)
}

func getEnvUint32Default(key string, d uint32) uint32 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	i, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return d
	}

	return uint32(i)
}

func getEnvBoolDefault(key string, d bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}

	return b
}

func getEnvDurationDefault(key string, d time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return d
	}

	p, err := time.ParseDuration(v)
	if err != nil {
		return d
	}

	return p
}

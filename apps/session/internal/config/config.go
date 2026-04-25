package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App        App        `env-prefix:"SESSION_APP_"`
	Logger     Logger     `env-prefix:"SESSION_LOG_"`
	HTTP       HTTP       `env-prefix:"SESSION_APP_"`
	Redis      Redis      `env-prefix:"SESSION_REDIS_"`
	Internal   Internal   `env-prefix:"SESSION_INTERNAL_"`
	Management Management `env-prefix:"SESSION_MANAGEMENT_"`
	WS         WS         `env-prefix:"SESSION_WS_"`
	Game       Game       `env-prefix:"SESSION_GAME_"`
}

type App struct {
	Name string `env:"NAME" env-default:"Quiz Session"`
}

type Logger struct {
	Level  string `env:"LEVEL"  env-default:"info"`
	Format string `env:"FORMAT" env-default:"json"`
}

type HTTP struct {
	Port string `env:"PORT" env-default:"8000"`
}

type Redis struct {
	Addr     string `env:"ADDR"     env-default:"localhost:6379"`
	Password string `env:"PASSWORD" env-default:"redis_password"`
	DB       int    `env:"DB"       env-default:"0"`
}

type Internal struct {
	ServiceName     string   `env:"SERVICE_NAME"     env-default:"session"`
	AllowedServices []string `env:"ALLOWED_SERVICES" env-separator:"," env-default:"management"`
	Token           string   `env:"TOKEN"            env-required:"true"`
}

type Management struct {
	BaseURL        string `env:"BASE_URL"        env-default:"http://management:8000"`
	TimeoutSeconds int    `env:"TIMEOUT_SECONDS" env-default:"5"`
}

type WS struct {
	ReadLimitBytes int `env:"READ_LIMIT_BYTES" env-default:"65536"`
}

type Game struct {
	RevealDurationSeconds int `env:"REVEAL_DURATION_SECONDS" env-default:"5"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config from env: %w", err)
	}

	cfg.Management.BaseURL = strings.TrimRight(cfg.Management.BaseURL, "/")
	cfg.Internal.AllowedServices = normalizeStringList(cfg.Internal.AllowedServices)

	return &cfg, nil
}

func (h HTTP) Address() string {
	return ":" + h.Port
}

func (i Internal) Allows(service string) bool {
	service = strings.TrimSpace(service)
	if service == "" {
		return false
	}

	for _, allowed := range i.AllowedServices {
		if strings.EqualFold(allowed, service) {
			return true
		}
	}

	return false
}

func (m Management) Timeout() time.Duration {
	return time.Duration(m.TimeoutSeconds) * time.Second
}

func (g Game) RevealDuration() time.Duration {
	return time.Duration(g.RevealDurationSeconds) * time.Second
}

func normalizeStringList(values []string) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		result = append(result, value)
	}

	return result
}

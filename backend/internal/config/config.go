package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
)

type NestingConfig struct {
	Workers int `env:"WORKERS" env-default:"2"`
	Queue   int `env:"QUEUE" env-default:"4"`
}

type Config struct {
	Nesting  NestingConfig  `env-prefix:"NESTING_"`
	App      AppConfig      `env-prefix:"APP_"`
	HTTP     HTTPConfig     `env-prefix:"HTTP_"`
	Log      LogConfig      `env-prefix:"LOG_"`
	DB       DBConfig       `env-prefix:"DB_"`
	Security SecurityConfig `env-prefix:"SECURITY_"`
	JWT      JWTConfig      `env-prefix:"JWT_"`
}

type AppConfig struct {
	Env string `env:"ENV" env-default:"local"`
}

type HTTPConfig struct {
	Port          int    `env:"PORT" env-default:"8082"`
	BrowserOrigin string `env:"BROWSER_ORIGIN" env-default:"http://localhost:5173"`
}

type LogConfig struct {
	Level string `env:"LEVEL" env-default:"debug"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     int    `env:"PORT" env-default:"5435"`
	User     string `env:"USER" env-default:"postgres"`
	Password string `env:"PASSWORD" env-default:"postgres"`
	DBName   string `env:"NAME" env-default:"nesting_optimizer"`
	SSLMode  string `env:"SSL_MODE" env-default:"disable"`
}

type SecurityConfig struct {
	BcryptCost int `env:"BCRYPT_COST" env-default:"12"`
}

type JWTConfig struct {
	Issuer        string `env:"ISSUER" env-default:"nesting-optimizer"`
	AccessSecret  string `env:"ACCESS_SECRET" env-default:"dev-access-secret"`
	AccessTTLMin  int    `env:"ACCESS_TTL_MIN" env-default:"15"`
	RefreshSecret string `env:"REFRESH_SECRET" env-default:"dev-refresh-secret"`
	RefreshTTLMin int    `env:"REFRESH_TTL_MIN" env-default:"10080"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c DBConfig) DSN() string {
	u := url.URL{Scheme: "postgres", User: url.UserPassword(c.User, c.Password), Host: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Path: "/" + c.DBName}
	q := url.Values{}
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

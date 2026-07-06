package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      AppConfig      `env-prefix:"APP_"`
	HTTP     HTTPConfig     `env-prefix:"HTTP_"`
	Log      LogConfig      `env-prefix:"LOG_"`
	DB       DBConfig       `env-prefix:"DB_"`
	Security SecurityConfig `env-prefix:"SECURITY_"`
}

type AppConfig struct {
	Env string `env:"ENV" env-default:"local"`
}

type HTTPConfig struct {
	Port int `env:"PORT" env-default:"8082"`
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

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}
	return cfg, nil
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"prod"`
	HTTPServer `yaml:"http_server"`
	Storage    `yaml:"storage"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

type Storage struct {
	Host        string `yaml:"host" env:"DB_HOST" env-required:"true"`
	User        string `yaml:"user" env:"DB_USER" env-required:"true"`
	Password    string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	DBName      string `yaml:"dbname" env:"DB_NAME" env-required:"true"`
	SSLMode     string `yaml:"sslmode" env:"DB_SSL_MODE" env-default:"disable"`
	StoragePool `yaml:"storage_pool"`
}

func (s *Storage) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		s.User, s.Password, s.Host, s.DBName, s.SSLMode)
}

type StoragePool struct {
	MaxConns        int32         `yaml:"max_conns" env:"DB_MAX_CONNS" env-default:"20"`
	MinConns        int32         `yaml:"min_conns" env:"DB_MIN_CONNS" env-default:"5"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME" env-default:"30m"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT" env-default:"5m"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME" env-default:"5m"`
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file does not exist: %s", configPath)
		}
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("failed to read config file: %s", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			help, _ := cleanenv.GetDescription(&cfg, nil)
			log.Fatalf("invalid configuration (env): \n%s\n error: %v", help, err)
		}
	}

	return &cfg
}

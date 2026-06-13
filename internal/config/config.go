package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"prod"`
	HTTPServer `yaml:"http_server"`
	Storage    `yaml:"storage"`
}

type HTTPServer struct {
	Address string `yaml:"address" validate:"required"`
}

type Storage struct {
	Host        string `yaml:"host" validate:"required"`
	User        string `yaml:"user" validate:"required"`
	Password    string `yaml:"password" validate:"required"`
	DBName      string `yaml:"dbname" validate:"required"`
	SSLMode     string `yaml:"sslmode" env-default:"disable"`
	StoragePool `yaml:"storage_pool"`
}

type StoragePool struct {
	MaxConns        int32         `yaml:"max_conns" env-default:"20"`
	MinConns        int32         `yaml:"min_conns" env-default:"5"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env-default:"30m"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout" env-default:"5m"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env-default:"5m"`
}

func MustLoad() *Config {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", *configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	return &cfg
}

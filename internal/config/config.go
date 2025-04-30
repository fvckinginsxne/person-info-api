package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server ServerConfig `env-prefix:"SERVER_" env-required:"true"`
	DB     DBConfig     `env-prefix:"DB_" env-required:"true"`
}

type ServerConfig struct {
	Host        string        `env:"HOST" env-default:"localhost"`
	Port        int           `env:"PORT" env-default:"8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"15s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     string `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-requited:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	Name     string `env:"NAME" env-required:"true"`
}

// MustLoad Load config file and panic if error occurs
func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config file path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file path")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string     `toml:"env" env-default:"local"`
	LogLevel string     `toml:"log_level" env-default:"info"`
	Http     HttpConfig `toml:"http" env-required:"true"`
	DB       DBConfig   `toml:"db" env-required:"true"`
}

type HttpConfig struct {
	Port    int `toml:"port"`
	Timeout int `toml:"timeout"`
}

type DBConfig struct {
	Host string `toml:"host" env-required:"true"`
	Port int    `toml:"port"`
	User string `toml:"user"`
}

var (
	cfg  *Config
	once sync.Once
)

func ParseConfig(path string) *Config {
	once.Do(func() {
		cfg = &Config{}
		if path != "" {
			if err := cleanenv.ReadConfig(path, cfg); err != nil {
				panic(err)
			}
		} else {
			if err := cleanenv.ReadEnv(cfg); err != nil {
				panic(err)
			}
		}
	})
	return cfg
}

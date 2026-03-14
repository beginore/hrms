package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `toml:"env"`
	LogLevel string `toml:"log_level" env:"LOG_LEVEL" env-default:"info"`

	AWS struct {
		Region          string `toml:"region" env:"AWS_REGION" env-default:"eu-west-1"`
		AccessKeyID     string `toml:"access_key_id" env:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `toml:"secret_access_key" env:"AWS_SECRET_ACCESS_KEY"`
	} `toml:"aws"`

	Cognito struct {
		UserPoolID      string `toml:"user_pool_id" env:"COGNITO_USER_POOL_ID"`
		AppClientID     string `toml:"app_client_id" env:"COGNITO_APP_CLIENT_ID"`
		AppClientSecret string `toml:"app_client_secret" env:"COGNITO_APP_CLIENT_SECRET"`
	} `toml:"cognito"`

	Database struct {
		DSN string `toml:"dsn" env:"DB_DSN" env-description:"PostgreSQL connection string"`
	} `toml:"database"`

	SES struct {
		SenderEmail string `toml:"sender_email" env:"SES_SENDER_EMAIL" env-description:"Verified SES sender email"`
	} `toml:"ses"`
}

var (
	cfg  *Config
	once sync.Once
)

func ParseConfig(explicitConfigPath string) *Config {
	once.Do(func() {
		cfg = &Config{}

		var configFile string

		if explicitConfigPath != "" {
			configFile = explicitConfigPath
		} else {
			_, filename, _, _ := runtime.Caller(0)
			root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
			localPath := filepath.Join(root, "configs", "local", "config.toml")

			if _, err := os.Stat(localPath); err == nil {
				configFile = localPath
				fmt.Printf("[config] Using local developer config: %s\n", localPath)
			}
		}

		var err error

		if configFile != "" {
			err = cleanenv.ReadConfig(configFile, cfg)
			if err != nil {
				panic(fmt.Errorf("failed to read config file %s: %w", configFile, err))
			}
			err = cleanenv.ReadEnv(cfg)
			if err != nil {
				panic(fmt.Errorf("failed to apply env overrides after file: %w", err))
			}
		} else {
			err = cleanenv.ReadEnv(cfg)
			if err != nil {
				panic(fmt.Errorf("failed to read environment variables: %w", err))
			}
		}

		if cfg.Cognito.UserPoolID == "" {
			panic("COGNITO_USER_POOL_ID is required (set in config file or environment variable)")
		}
		if cfg.Cognito.AppClientID == "" {
			panic("COGNITO_APP_CLIENT_ID is required (set in config file or environment variable)")
		}
		if cfg.Cognito.AppClientSecret == "" {
			fmt.Println("[WARN] COGNITO_APP_CLIENT_SECRET not set — assuming public client (no SecretHash)")
		}

		if cfg.Database.DSN == "" {
			panic("Database DSN is required (add [database] dsn = \"...\" to config.toml or set DB_DSN env var)")
		}

		if cfg.SES.SenderEmail == "" {
			panic("SES sender email is required (add [ses] sender_email = \"...\" to config.toml or set SES_SENDER_EMAIL env var)")
		}
	})

	return cfg
}

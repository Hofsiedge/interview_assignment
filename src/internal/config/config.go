package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	CoinGeckoToken  string `env:"COIN_GECKO_TOKEN" env-description:"CoinGecko demo API auth token" env-required:"true"`
	PollingInterval uint   `env:"POLLING_INTERVAL" env-description:"Polling interval in seconds (positive integer)" env-required:"true"`
	PostgresURL     string `env:"POSTGRES_URL" env-required:"true"`
}

func ReadConfig() (Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)

	return cfg, err
}

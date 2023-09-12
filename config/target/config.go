package target

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	DBAddress      string        `env:"DB_ADDRESS"`
	SleepDuration  time.Duration `env:"HANDLE_SLEEP_TIME" envDefault:"1s"`
	TickerDuration time.Duration `env:"TICKER" envDefault:"10s"`
	Strategy       string        `env:"STRATEGY"`
	Port           int           `env:"PORT" envDefault:"8082"`
}

func LoadConfig() (Config, error) {
	conf := Config{}

	if err := env.Parse(&conf); err != nil {
		return conf, err
	}

	return conf, nil
}

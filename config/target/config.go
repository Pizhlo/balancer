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
	Port           int           `env:"PORT" envDefault:"8081"`
}

func LoadConfig(path string) (Config, error) {
	// viper.AddConfigPath(path)
	// viper.SetConfigName("app")
	// viper.SetConfigType("env")

	// viper.AutomaticEnv()

	conf := Config{}

	if err := env.Parse(&conf); err != nil {
		return conf, err
	}

	//conf.SleepDuration
	// sleepDur := os.Getenv("HANDLE_SLEEP_TIME")
	// conf.DBAddress = os.Getenv("DB_ADDRESS")
	// //conf.TickerDuration
	// ticker := os.Getenv("TICKER")
	// conf.Strategy = os.Getenv("STRATEGY")

	// err := viper.ReadInConfig()
	// if err != nil {
	// 	return conf, err
	// }

	// err = viper.Unmarshal(&conf)
	// if err != nil {
	// 	return conf, err
	// }

	return conf, nil
}

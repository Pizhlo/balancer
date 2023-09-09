package target

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBAddress      string        `mapstructure:"DB_ADDRESS"`
	SleepDuration  time.Duration `mapstructure:"HANDLE_SLEEP_TIME"`
	TickerDuration time.Duration `mapstructure:"TICKER"`
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	conf := Config{}

	err := viper.ReadInConfig()
	if err != nil {
		return conf, err
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

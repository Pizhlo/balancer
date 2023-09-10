package balancer

import (
	"os"
)

type Config struct {
	BalancerPort string `mapstructure:"BALANCER_PORT"`
	DBAddress    string `mapstructure:"DB_ADDRESS"`
	Strategy     string `mapstructure:"STRATEGY"`
}

func LoadConfig() (Config, error) {
	conf := Config{}

	conf.BalancerPort = os.Getenv("BALANCER_PORT")
	conf.DBAddress = os.Getenv("DB_ADDRESS")
	conf.Strategy = os.Getenv("STRATEGY")

	return conf, nil
}

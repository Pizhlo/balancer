package balancer

import (
	"os"
)

type Config struct {
	BalancerPort string `mapstructure:"BALANCER_PORT"`
	DBAddress    string `mapstructure:"DB_ADDRESS"`
	Strategy     string `mapstructure:"STRATEGY"`
}

func LoadConfig(path string) (Config, error) {
	// viper.AddConfigPath(".")
	// viper.SetConfigName("app")
	// viper.SetConfigType("env")

	// viper.AutomaticEnv()

	conf := Config{}

	conf.BalancerPort = os.Getenv("BALANCER_PORT")
	conf.DBAddress = os.Getenv("DB_ADDRESS")
	conf.Strategy = os.Getenv("STRATEGY")

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

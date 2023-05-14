package config

import "github.com/caarlos0/env/v8"

type Influx struct {
	Host     string `env:"INFLUX_HOST" envDefault:"localhost"`
	Port     int    `env:"INFLUX_PORT" envDefault:"8086"`
	Username string `env:"INFLUX_USERNAME" envDefault:"root"`
	Password string `env:"INFLUX_PASSWORD" envDefault:"root"`
	Bucket   string `env:"INFLUX_BUCKET" envDefault:"sensors"`
	Org      string `env:"INFLUX_ORG" envDefault:"sensors"`
	Token    string `env:"INFLUX_TOKEN" envDefault:""`
}

func BuildInflux(cfg *Config) error {
	influx := &Influx{}
	if err := env.Parse(influx); err != nil {
		return err
	}
	cfg.Influx = *influx

	return nil
}

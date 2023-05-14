package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	env "github.com/caarlos0/env/v8"
)

type Config struct {
	Local
	MQTT
	Influx
}

func Build() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildLocal(&cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildMQTT(&cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildInflux(&cfg); err != nil {
		return nil, logs.Error(err)
	}

	return &cfg, nil
}

package config

import (
	env "github.com/caarlos0/env/v8"
)

type Local struct {
	KeepLocal            bool   `env:"LOCAL_ONLY" envDefault:"false" json:"keep_local,omitempty"`
	Development          bool   `env:"DEVELOPMENT" envDefault:"false" json:"development,omitempty"`
	SleepTime            int    `env:"SLEEP_TIME" envDefault:"5" json:"sleep_time,omitempty"`
	TempChangePercentage string `env:"TEMP_CHANGE_PERCENTAGE" envDefault:"10" json:"temp_change_percentage,omitempty"`
	ErrorLimit           int    `env:"ERROR_LIMIT" envDefault:"5" json:"error_limit,omitempty"`
}

func BuildLocal(cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.Local = *local

	return nil
}

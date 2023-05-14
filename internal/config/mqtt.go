package config

import (
	env "github.com/caarlos0/env/v8"
)

type MQTT struct {
	Host     string `env:"MQTT_HOST" envDefault:"localhost" json:"host,omitempty"`
	Port     int    `env:"MQTT_PORT" envDefault:"1883" json:"port,omitempty"`
	Username string `env:"MQTT_USERNAME" envDefault:"" json:"username,omitempty"`
	Password string `env:"MQTT_PASSWORD" envDefault:"" json:"password,omitempty"`
	Topic    string `env:"MQTT_TOPIC" envDefault:"" json:"topic,omitempty"`
}

func BuildMQTT(cfg *Config) error {
	mqtt := &MQTT{}
	if err := env.Parse(mqtt); err != nil {
		return err
	}
	cfg.MQTT = *mqtt

	return nil
}

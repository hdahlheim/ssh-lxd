package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Keys []string
}

type Config struct {
	Auth map[string]Host
}

var cfg Config

func GetConfig() *Config {
	return &cfg
}

func LoadConfig() error {
	raw, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return err
	}

	return nil
}

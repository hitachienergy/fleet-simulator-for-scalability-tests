package main

import (
	"gopkg.in/yaml.v3"
)

type ThingsBoardConfig struct {
	Client ThingsBoardClientConfig `yaml:"client"`
	Server ThingsBoardServerConfig `yaml:"server"`
}

type ThingsBoardServerConfig struct {
	DevicesEndpoint string `yaml:"devicesEndpoint"`
}

type ThingsBoardClientConfig struct {
	Args map[string]interface{} `yaml:"args"`
}

func ParseConfig(data []byte) (config *ThingsBoardConfig, err error) {
	if err != nil {
		return &ThingsBoardConfig{}, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

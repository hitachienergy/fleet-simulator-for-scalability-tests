package main

import (
	"gopkg.in/yaml.v3"
)

type HawkbitConfig struct {
	Client HawkbitClientConfig `yaml:"client"`
	Server HawkbitServerConfig `yaml:"server"`
}

type HawkbitServerConfig struct {
	DevicesEndpoint string `yaml:"devicesEndpoint"`
}

type HawkbitClientConfig struct {
	Args map[string]interface{} `yaml:"args"`
}

func ParseConfig(data []byte) (config *HawkbitConfig, err error) {
	if err != nil {
		return &HawkbitConfig{}, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

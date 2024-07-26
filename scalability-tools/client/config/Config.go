package config

import (
	"gopkg.in/yaml.v3"
)

// Config represents the mandatory structure of the configuration yaml file
type Config struct {
	LogLevel   string              `yaml:"logLevel"`
	Target     string              `yaml:"target"`
	Client     ClientDefaultConfig `yaml:"client"`
	Output     OutputDefaultConfig `yaml:"output"`
	Simulation SimulationConfig    `yaml:"simulation"`
}

type ClientDefaultConfig struct {
	Template            string `yaml:"template"`
	Factory             string `yaml:"factory"`
	Number              int    `yaml:"numberOfDevices"`
	DevicesRegisterMode string `yaml:"devicesRegisterMode"`
	NamePrefix          string `yaml:"namePrefix"`
}

type OutputDefaultConfig struct {
	Path string `yaml:"path"`
}

// ParseConfig parses the configuration file for the program to use
func ParseConfig(data []byte) (config Config, err error) {
	err = yaml.Unmarshal(data, &config)
	return config, err
}

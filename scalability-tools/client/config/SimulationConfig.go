package config

import (
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SimulationConfig represents the device-side simulation related configuration
type SimulationConfig struct {
	Task      string           `yaml:"task"`
	DummyWork DummyWorkDetails `yaml:"dummyWork"`
	Crash     CrashDetails     `yaml:"crash"`
	Seed      int64            `yaml:"seed"`
}

// SimulationDetails represents the details of device-side simulation related configuration
type DummyWorkDetails struct {
	AffectedNum int          `yaml:"number"`
	Percentage  Percentage   `yaml:"percent"`
	Duration    TimeDuration `yaml:"duration"`
	Variation   TimeDuration `yaml:"variation"`
	Period      TimeDuration `yaml:"period"`
}

type CrashDetails struct {
	AffectedNum int          `yaml:"number"`
	Percentage  Percentage   `yaml:"percent"`
	Within      TimeDuration `yaml:"within"`
}

type TimeDuration time.Duration
type Percentage float64

// UnmarshalYAML overwrites the parser for the Pecentage struct
func (p *Percentage) UnmarshalYAML(value *yaml.Node) error {
	percentageStr := value.Value
	percentageStr = strings.TrimSuffix(percentageStr, "%")

	percentage, err := strconv.ParseFloat(percentageStr, 64)
	if err != nil {
		return err
	}
	*p = Percentage(percentage / 100.0)
	return nil
}

// UnmarshalYAML overwrites the parser for the Pecentage struct
func (p *TimeDuration) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*p = TimeDuration(duration)
	return nil
}

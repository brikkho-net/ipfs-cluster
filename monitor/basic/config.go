package basic

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ipfs/ipfs-cluster/config"
	"github.com/kelseyhightower/envconfig"
)

const configKey = "monbasic"
const envConfigKey = "cluster_monbasic"

// Default values for this Config.
const (
	DefaultCheckInterval    = 15 * time.Second
	DefaultFailureThreshold = 3.0
)

// Config allows to initialize a Monitor and customize some parameters.
type Config struct {
	config.Saver

	CheckInterval    time.Duration
	FailureThreshold float64
}

type jsonConfig struct {
	CheckInterval    string  `json:"check_interval"`
	FailureThreshold float64 `json:"failure_threshold"`
}

// ConfigKey provides a human-friendly identifier for this type of Config.
func (cfg *Config) ConfigKey() string {
	return configKey
}

// Default sets the fields of this Config to sensible values.
func (cfg *Config) Default() error {
	cfg.CheckInterval = DefaultCheckInterval
	cfg.FailureThreshold = DefaultFailureThreshold
	return nil
}

// ApplyEnvVars fills in any Config fields found
// as environment variables.
func (cfg *Config) ApplyEnvVars() error {
	jcfg := cfg.toJSONConfig()

	err := envconfig.Process(envConfigKey, jcfg)
	if err != nil {
		return err
	}

	return cfg.applyJSONConfig(jcfg)
}

// Validate checks that the fields of this Config have working values,
// at least in appearance.
func (cfg *Config) Validate() error {
	if cfg.CheckInterval <= 0 {
		return errors.New("basic.check_interval too low")
	}

	if cfg.FailureThreshold <= 0 {
		return errors.New("basic.failure_threshold too low")
	}

	return nil
}

// LoadJSON sets the fields of this Config to the values defined by the JSON
// representation of it, as generated by ToJSON.
func (cfg *Config) LoadJSON(raw []byte) error {
	jcfg := &jsonConfig{}
	err := json.Unmarshal(raw, jcfg)
	if err != nil {
		logger.Error("Error unmarshaling basic monitor config")
		return err
	}

	cfg.Default()

	return cfg.applyJSONConfig(jcfg)
}

func (cfg *Config) applyJSONConfig(jcfg *jsonConfig) error {
	interval, _ := time.ParseDuration(jcfg.CheckInterval)
	cfg.CheckInterval = interval
	cfg.FailureThreshold = jcfg.FailureThreshold

	return cfg.Validate()
}

// ToJSON generates a human-friendly JSON representation of this Config.
func (cfg *Config) ToJSON() ([]byte, error) {
	jcfg := cfg.toJSONConfig()

	return json.MarshalIndent(jcfg, "", "    ")
}

func (cfg *Config) toJSONConfig() *jsonConfig {
	return &jsonConfig{
		CheckInterval:    cfg.CheckInterval.String(),
		FailureThreshold: cfg.FailureThreshold,
	}
}

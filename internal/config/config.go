package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Load reads the YAML configuration from the given file path
// and unmarshals it into a Config struct.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Optional: set defaults
	v.SetDefault("data_dir", "./var")
	v.SetDefault("bin_path", "./bin")

	// Read in the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validation: ensure at least one node is defined
	if len(cfg.Cluster) == 0 {
		return nil, fmt.Errorf("config error: cluster must contain at least one node")
	}

	return &cfg, nil
}

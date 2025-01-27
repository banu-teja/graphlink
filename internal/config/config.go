package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Neo4jURI      string `mapstructure:"neo4j_uri"`
	Neo4jUsername string `mapstructure:"neo4j_username"`
	Neo4jPassword string `mapstructure:"neo4j_password"`
	GRPCPort      int    `mapstructure:"grpc_port"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Look for config in the working directory

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; proceed with defaults and/or environment variables
			fmt.Println("Config file not found. Using environment variables and defaults.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default values if not provided in config file or environment variables
	if cfg.Neo4jURI == "" {
		cfg.Neo4jURI = "bolt://localhost:7687"
	}
	if cfg.Neo4jUsername == "" {
		cfg.Neo4jUsername = "neo4j"
	}
	if cfg.Neo4jPassword == "" {
		cfg.Neo4jPassword = "password"
	}
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50051
	}

	return &cfg, nil
}

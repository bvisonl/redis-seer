package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config - Global configuration variable
var Config SeerConfig

// SeerConfig - Configuration structure
type SeerConfig struct {
	Debug           bool                          `yaml:"debug"`
	Port            int                           `yaml:"port"`
	MonitorInterval int                           `yaml:"monitorInterval"`
	Servers         map[string]*RedisServerConfig `yaml:"servers"` // TODO: Oh no, map of struct pointer...
	CurrentMaster   string                        `yaml:"master"`
}

// RedisServerConfig - Redis Server configuration structure
type RedisServerConfig struct {
	Alias    string `yaml:"alias"`
	Database int    `yaml:"db"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Enabled  bool   `yaml:"enabled"`
	Alive    bool
}

// LoadConfig - Load the configuration from ./config.yml
func LoadConfig(configPath string) {
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	// Initialize default values
	Config.CurrentMaster = "redis1"
	for _, server := range Config.Servers {
		server.Alive = true
	}

}

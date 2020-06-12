package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// Constants
const (
	SERVER_STATUS_ACTIVE   = 1
	SERVER_STATUS_INACTIVE = 2
	SERVER_STATUS_DISABLED = 6
)

// Global variables
var (
	seerConfig      SeerConfig
	monitorInterval time.Duration
)

// SeerConfig - Configuration structure
type SeerConfig struct {
	Debug           bool                          `yaml:"debug"`
	Port            int                           `yaml:"port"`
	Database        int                           `yaml:"db"`
	MonitorInterval int                           `yaml:"monitorInterval"`
	Servers         map[string]*RedisServerConfig `yaml:"servers"`
	SelectionMode   string                        `yaml:"selectionMode"`
	Master          string
}

// RedisServerConfig - Redis Server configuration structure
type RedisServerConfig struct {
	Alias   string `yaml:"alias"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Enabled bool   `yaml:"enabled"`
	Status  int
}

// LoadConfig - Load the configuration from ./config.yml
func loadConfig(configPath string) {

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&seerConfig)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	// Interval between monitoring commands
	interval := strconv.Itoa(seerConfig.MonitorInterval) + "s"
	monitorInterval, err = time.ParseDuration(interval)

	if err != nil {
		log.Printf("Invalid interval provided %s for the monitor. Error: %s\r\n", interval, err)
		os.Exit(1)
		return
	}

}

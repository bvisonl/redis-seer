package main

import (
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v2"
)

// YamlConfig - Configuration structure
type SeerConfig struct {
	Debug           bool                          `yaml:"debug"`
	Port            int                           `yaml:"port"`
	MonitorInterval int                           `yaml:"monitorInterval"`
	Servers         map[string]*RedisServerConfig `yaml:"servers"` // TODO: Oh no, map of struct pointer...
	CurrentMaster   string
}

type RedisServerConfig struct {
	Alias    string `yaml:"alias"`
	Database int    `yaml:"db"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Alive    bool
}

type RedisConnection struct {
	ServerConfig RedisServerConfig
	Connection   *net.Conn
}

// Config - Global configuration variable
var Config SeerConfig

// RedisMasterKey - Current redis master
var RedisMasterKey string

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
	for _, server := range Config.Servers {
		server.Alive = true
	}
}

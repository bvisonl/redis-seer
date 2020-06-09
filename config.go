package main

import (
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Debug   		bool                   		`yaml:"debug"`
	Port    		int                    		`yaml:"port"`
	MonitorInterval int 						`yaml:"monitorInterval"`
	// TODO: Oh no, map of struct pointer...
	Servers 		map[string]*RedisServer 	`yaml:"servers"`
}

type RedisServer struct {
	Alias    string `yaml:"alias"`
	Database int    `yaml:"db"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Alive	 bool
}

type RedisConnection struct {
	Name       string
	Host     string
	Port     int
	Connection *net.Conn
}

// Config - Global configuration variable
var Config YamlConfig

// RedisMasterKey - Current redis master
var RedisMasterKey string

// LoadConfig - Load the configuration from ./config.yml
func LoadConfig() {
	f, err := os.Open("./config.yml")
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

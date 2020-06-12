package main

import (
	"flag"
	"log"
)

func main() {

	// Configurations by flag
	configPath := flag.String("c", "./config.yml", "Configuration file path")
	debug := flag.Bool("d", false, "Enable debugging")
	flag.Parse()

	log.Println(*configPath)

	// Load configuration
	loadConfig(*configPath)

	if *debug == true {
		seerConfig.Debug = true
	}

	// Start the monitoring service
	go startMonitor()

	// Start listening for connections
	startServer()
}

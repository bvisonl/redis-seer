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
	LoadConfig(*configPath)

	if *debug == true {
		Config.Debug = true
	}

	// Start the monitoring service
	go StartMonitor()

	// Start listening for connections
	StartServer()
}

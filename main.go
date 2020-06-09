package main

func main() {

	// Load servers configuration
	LoadConfig()

	// Start the monitoring service
	go StartMonitor()

	// Start listening for connections
	StartServer()
}

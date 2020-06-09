package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"os"
	"strconv"
	"math/rand"

	"github.com/bvisonl/redis-standalone-proxy/redis"
)

func StartServer() {
	port := Config.Port

	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatalf("Unable to start server on port %d. Error: %s", port, err)
		os.Exit(1)
	}

	defer listen.Close()

	log.Printf("Listening on :%d", port)

	for {
		proxyConnection, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go initProxy(proxyConnection)
	}
}

// Per connection functions

func initProxy(proxyConnection net.Conn) {

	defer proxyConnection.Close()

	// Create a new connection with all redis servers
	redisConnections := make(map[string]*RedisConnection, 0)

	// Establish connection with all the servers
	for key, server := range Config.Servers {

		redisConnection, err := net.Dial("tcp4", server.Host+":"+strconv.Itoa(server.Port))

		redisConnections[key] = &RedisConnection{
			Name:       key,
			Host: 	    server.Host,
			Port:		server.Port,
			Connection: &redisConnection,
		}

		if err != nil {
			log.Printf("An error occurred connection to server: %s. Error: %s\r\n", key, err)
			continue
		}

	}

	// Start listening for requests and proxying them
	proxy(proxyConnection, redisConnections)

}

func proxy(proxyConnection net.Conn, redisConnections map[string]*RedisConnection) {

	proxyReader := redis.NewReader(proxyConnection)

	for {
		// Start reading data
		proxyData, err := proxyReader.ReadObject()
		if err != nil {
			fmt.Printf("Error reading data. Error: %s\r\n", err)
			break
		}

		// Select a redis server
		// TODO: Implement selection algorithm
		// TODO: readOnlyFromSlave parameter in config.yml should be considered
		selectedRedis := selectServer()

		// Send data to the selected redis
		if redisConnections[selectedRedis].Connection == nil {
			redisConnection, err := net.Dial("tcp4", redisConnections[selectedRedis].Host+":"+strconv.Itoa(redisConnections[selectedRedis].Port))
			if err != nil {
				// THIS SHOULD NOT HAPPEN BECAUSE THE SELECTION PROCESS SHOULD RETURN AN ALIVE REDIS
				log.Printf("An error occurred connection to server: %s. Error: %s\r\n", redisConnections[selectedRedis].Name, err)
				time.Sleep(2 * time.Second)
				continue
			}
			redisConnections[selectedRedis].Connection = &redisConnection
		}

		(*redisConnections[selectedRedis].Connection).Write(proxyData)

		// Get data from selected redis
		redisReader := redis.NewReader((*redisConnections[selectedRedis].Connection))
		redisData, err := redisReader.ReadObject()
		if err != nil {
			fmt.Printf("Error reading data from Redis %s. Error: %s\r\n", redisConnections[selectedRedis].Name, err)
			continue
		}

		// Relay response from redis back to the client
		proxyConnection.Write(redisData)

	}

}



func selectServer() string {

	numServers := len(Config.Servers)
	candidate := rand.Intn(numServers)
	i := 0;

	lastAlive := ""

	for key, server := range Config.Servers {

		if server.Alive == true {
			lastAlive = key
		}

		if i == candidate && server.Alive == true {
			return key
		}

		i++
	}

	return lastAlive


}

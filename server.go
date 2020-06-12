package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/bvisonl/redis-seer/redis"
)

func startServer() {

	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(seerConfig.Port))

	if err != nil {
		log.Fatalf("Unable to start server on port %d. Error: %s", seerConfig.Port, err)
		os.Exit(1)
	}

	defer listen.Close()

	log.Printf("Listening on :%d", seerConfig.Port)

	for {
		proxyConnection, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go initProxy(proxyConnection)
	}
}

func initProxy(proxyConnection net.Conn) {

	defer proxyConnection.Close()

	// Create a new connection with all redis servers
	redisConnections := make(map[string]net.Conn, 0)

	// Establish connection with all the servers
	for key, server := range seerConfig.Servers {
		if server.Enabled == false {
			continue
		}
		redisConnection, err := net.Dial("tcp4", server.Host+":"+strconv.Itoa(server.Port))
		redisConnections[key] = redisConnection
		if err != nil {
			log.Printf("An error occurred connection to server: %s. Error: %s\r\n", key, err)
			continue
		}
		defer redisConnection.Close()
	}

	// Start listening for requests and proxying them
	proxy(proxyConnection, redisConnections)

}

func proxy(proxyConnection net.Conn, redisConnections map[string]net.Conn) {

	proxyReader := redis.NewReader(proxyConnection)

	// Start reading data
	for {
		proxyData, err := proxyReader.ReadObject()

		if err != nil {
			if err == io.EOF {
				continue
			}
			fmt.Printf("Error reading data from client. Error: %s\r\n", err)
			break
		}

		target, err := redis.GetTarget(proxyData)
		if err != nil {
			proxyConnection.Write([]byte("-Error " + err.Error() + "\r\n"))
			continue
		}

	SelectServer:
		selectedRedis, err := selectServer(target)
		if err != nil {
			proxyConnection.Write([]byte("-Error " + err.Error() + "\r\n"))
			continue
		}

		// Send data to the selected redis
		if redisConnections[selectedRedis] == nil {
			redisConnections[selectedRedis], err = net.Dial("tcp4", seerConfig.Servers[selectedRedis].Host+":"+strconv.Itoa(seerConfig.Servers[selectedRedis].Port))
			defer redisConnections[selectedRedis].Close()
			if redisConnections[selectedRedis] == nil {
				goto SelectServer
			}
		}

		(redisConnections[selectedRedis]).Write(proxyData)

		// Get data from selected redis
		redisReader := redis.NewReader((redisConnections[selectedRedis]))
		redisData, err := redisReader.ReadObject()
		if err != nil {
			fmt.Printf("Error reading data from Redis %s. Error: %s\r\n", selectedRedis, err)
			continue
		}

		// Relay response from redis back to the client
		proxyConnection.Write(redisData)

	}

}

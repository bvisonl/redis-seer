package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/bvisonl/redis-seer/redis"
)

func StartMonitor() {

	// Interval between monitoring commands
	interval := strconv.Itoa(Config.MonitorInterval) + "s"
	monitorInterval, err := time.ParseDuration(interval)

	if err != nil {
		log.Printf("Invalid interval provided %s for the monitor. Error: %s\r\n", interval, err)
		return
	}

	// Monitor all servers in the list
	for key, server := range Config.Servers {
		if server.Enabled == false {
			continue
		}
		go monitor(key, server, monitorInterval)
	}
}

func monitor(name string, server *RedisServerConfig, interval time.Duration) {

	redisConnection, err := net.Dial("tcp4", server.Host+":"+strconv.Itoa(server.Port))
	if err != nil {
		log.Printf("Error connecting to redis %s. Will attempt to reconnect in the next interval. Error: %s\r\n", name, err)
		time.Sleep(interval)
		go monitor(name, server, interval)
		return
	}

	log.Printf("Started monitoring %s.\r\n", name)
	server.Alive = true

	for {

		// Send PING to server
		writer := redis.NewRESPWriter(redisConnection)
		writer.WriteCommand("PING")

		// Receive PONG
		redisReader := redis.NewReader(redisConnection)
		_, err := redisReader.ReadObject()
		if err != nil {
			fmt.Printf("Error reading data from Redis %s. Error: %s\r\n", name, err)
			server.Alive = false
			break
		}

		// Wait for monitor interval
		time.Sleep(interval)
	}

	monitor(name, server, interval)

}

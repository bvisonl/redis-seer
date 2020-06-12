package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bvisonl/redis-seer/redis"
)

var (
	monitorConnections map[string]net.Conn = make(map[string]net.Conn, 0)
	mutex                                  = &sync.Mutex{}
)

func startMonitor() {

	// Establish connection with all the servers
	for key, server := range seerConfig.Servers {
		if server.Enabled == false {
			continue
		}
		redisConnection, err := net.Dial("tcp4", server.Host+":"+strconv.Itoa(server.Port))
		monitorConnections[key] = redisConnection
		if err != nil {
			log.Printf("An error occurred connection to server: %s. Error: %s\r\n", key, err)
			continue
		}
		if seerConfig.Master == "" {
			err = getMasterRedis(key, redisConnection)
			if err != nil {
				log.Printf(err.Error())
				os.Exit(1)
			}
		}
		defer redisConnection.Close()
	}

	// Monitor all servers in the list
	for name, redisConnection := range monitorConnections {
		go monitor(name, redisConnection)
	}

	fmt.Println("Current master is " + seerConfig.Master)
}

func getMasterRedis(name string, redisConnection net.Conn) (err error) {
	// Get the replication information
	writer := redis.NewRESPWriter(redisConnection)
	err = writer.WriteCommand("INFO", "replication")
	if err != nil {
		return err
	}

	// Wait for the answer
	reader := redis.NewReader(redisConnection)
	result, err := reader.ReadObject()
	if err != nil {
		return err
	}
	info, _ := redis.InfoToMap(result)

	if info["role"] == "master" {
		seerConfig.Master = name
		return nil
	} else {
		for name, server := range seerConfig.Servers {
			if server.Host == info["master_host"] && strconv.Itoa(server.Port) == info["master_port"] {
				seerConfig.Master = name
				return nil
			}
		}
	}

	return errors.New("Unable to find a suitable Master")
}

func monitor(name string, redisConnection net.Conn) {

Connect:
	// Mark the server as reconnecting
	mutex.Lock()
	seerConfig.Servers[name].Status = SERVER_STATUS_INACTIVE
	mutex.Unlock()

	// Step 1 - Connect to the server
	redisConnection, err := net.Dial("tcp4", seerConfig.Servers[name].Host+":"+strconv.Itoa(seerConfig.Servers[name].Port))
	if err != nil {
		if seerConfig.Master == name {
			seerConfig.Master = "" // To avoid duplicate call of masterFailover
			go masterFailover()
		}

		log.Printf("Error connecting to redis %s. Will attempt to reconnect in the next interval. Error: %s\r\n", name, err)

		// Wait for the next interval to retry
		time.Sleep(monitorInterval)
		goto Connect
	}

	// Close the connection when done
	defer redisConnection.Close()

	// Update server status and connection
	mutex.Lock()
	monitorConnections[name] = redisConnection
	seerConfig.Servers[name].Status = SERVER_STATUS_ACTIVE
	mutex.Unlock()

	log.Printf("Started monitoring %s.\r\n", name)

	// Check the current master
	checkMaster(name, redisConnection)

	for {

		// Send PING to server
		writer := redis.NewRESPWriter(redisConnection)
		err := writer.WriteCommand("PING")
		if err != nil {
			log.Printf("Error sending PING to redis %s. Will attempt to reconnect in the next interval. Error: %s\r\n", name, err)
			time.Sleep(monitorInterval)
			goto Connect
		}

		// Receive PONG
		redisReader := redis.NewReader(redisConnection)
		_, err = redisReader.ReadObject()
		if err != nil {
			fmt.Printf("Error reading data from Redis %s. Error: %s\r\n", name, err)
			time.Sleep(monitorInterval)
			goto Connect
		}

		// Wait for monitor interval
		time.Sleep(monitorInterval)
	}

}

func masterFailover() (err error) {

	// Look for any active Slave
	activeSlave := ""
	for name, server := range seerConfig.Servers {
		if server.Status == SERVER_STATUS_ACTIVE {
			activeSlave = name
		}
	}

	// Get the connection to that slave
	redisConnection := monitorConnections[activeSlave]
	if redisConnection == nil {
		log.Printf("No active servers.")
		return errors.New("No active servers found.")
	}

	// Get the info
	writer := redis.NewRESPWriter(redisConnection)
	err = writer.WriteCommand("SLAVEOF", "NO", "ONE")
	if err != nil {
		return err
	}

	seerConfig.Master = activeSlave
	fmt.Printf("Promoted %s to master\r\n", seerConfig.Master)

	for name, connection := range monitorConnections {
		if name != activeSlave {
			setMasterRedis(name, connection)
		}
	}

	return nil

}

func checkMaster(name string, redisConnection net.Conn) (err error) {

	if name == seerConfig.Master {
		return nil
	}

	// Get the replication information
	writer := redis.NewRESPWriter(redisConnection)
	err = writer.WriteCommand("INFO", "replication")
	if err != nil {
		return err
	}

	// Wait for the answer
	reader := redis.NewReader(redisConnection)
	result, err := reader.ReadObject()
	if err != nil {
		return err
	}
	info, _ := redis.InfoToMap(result)

	if info["role"] == "slave" && (seerConfig.Servers[seerConfig.Master].Host == info["master_host"] || strconv.Itoa(seerConfig.Servers[seerConfig.Master].Port) != info["master_port"]) {
		// If the master is different than the one we have, change it
		setMasterRedis(name, redisConnection)
	}

	return nil
}

func setMasterRedis(name string, redisConnection net.Conn) (err error) {
	writer := redis.NewRESPWriter(redisConnection)
	err = writer.WriteCommand("SLAVEOF", seerConfig.Servers[seerConfig.Master].Host, strconv.Itoa(seerConfig.Servers[seerConfig.Master].Port))
	if err != nil {
		return err
	}

	reader := redis.NewReader(redisConnection)
	_, err = reader.ReadObject()
	if err != nil {
		return err
	}

	return nil
}

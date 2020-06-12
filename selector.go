package main

import (
	"errors"
	"math/rand"
)

const (
	MODE_RANDOM = "random"
	MODE_RR     = "round-robin"
	MODE_LOAD   = "load"
)

func selectServer(target string) (serverName string, err error) {
	if target == "Master" {
		return seerConfig.Master, nil
	}

	switch seerConfig.SelectionMode {
	default:
		return selectServerRandom()
		break
	}

	return seerConfig.Master, nil
}

func selectServerRandom() (selectedServer string, err error) {
	numServers := len(seerConfig.Servers)
	candidate := rand.Intn(numServers)
	i := 0

	selectedServer = ""

	for key, server := range seerConfig.Servers {

		if server.Status == SERVER_STATUS_ACTIVE {
			selectedServer = key
		}

		if i == candidate && server.Status == SERVER_STATUS_ACTIVE {
			return key, nil
		}

		i++
	}

	if selectedServer != "" {
		return "", errors.New("Unable to find an available server.")
	}

	return selectedServer, nil

}

/*
	Copyright (C) 2023  Martijn van der Kleijn

	This file is part of HomeDash sidecar.

	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ContainerUpdate struct {
	Uuid       string          `json:"uuid"`
	Containers []ContainerInfo `json:"containers"`
}

type ContainerInfo struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	Icon    string `json:"icon"`
	Comment string `json:"comment"`
}

var homedashServer string
var homedashInterval string
var myUUID string

func main() {
	// log.SetLevel(log.DebugLevel)

	homedashServer = os.Getenv("HOMEDASH_SERVER")
	if homedashServer == "" {
		log.Fatalf("Environment variable HOMEDASH_SERVER not set or empty.")
	}

	homedashInterval = os.Getenv("HOMEDASH_INTERVAL")
	if homedashInterval == "" {
		homedashInterval = "10"
		log.Warnf("Environment variable HOMEDASH_INTERVAL not set or empty. Defaulting to once every %v minutes.", homedashInterval)
	}

	interval, err := strconv.Atoi(homedashInterval)
	if err != nil {
		log.Fatalf("Failed to parse HOMEDASH_INTERVAL time interval: %s", err)
	}

	homedashServer = homedashServer + "/api/v1/applications"

	log.Debugf("Will attempt to connect to %s once every %s", homedashServer, homedashInterval)

	// Identify myself
	myUUID = uuid.New().String()

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	// Create the options to filter containers
	options := types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "homedash.name"),
			filters.Arg("status", "running"),
		),
	}

	// Check for old data and clean up every X minutes
	go func() {
		for {
			applications := retrieveContainers(cli, options)

			log.Debugf("Trying to add: %v", applications)

			err = postContainerInfo(applications)
			if err != nil {
				log.Println(err)
			}

			time.Sleep(time.Duration(interval) * time.Minute)
		}
	}()

	select {}
}

func retrieveContainers(cli *client.Client, options types.ContainerListOptions) []ContainerInfo {
	applications := make([]ContainerInfo, 0)

	// Get a list of containers based on the filter
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		log.Fatal(err)
	}

	// Print the container information for each active container and post to REST API
	for _, container := range containers {
		containerInfo := ContainerInfo{
			Name:    container.Labels["homedash.name"],
			Url:     container.Labels["homedash.url"],
			Icon:    container.Labels["homedash.icon"],
			Comment: container.Labels["homedash.comment"],
		}
		applications = append(applications, containerInfo)
	}

	return applications
}

func postContainerInfo(applications []ContainerInfo) error {
	containerUpdate := ContainerUpdate{
		Uuid:       myUUID,
		Containers: applications,
	}

	payload, err := json.Marshal(containerUpdate)
	if err != nil {
		return err
	}

	// Create a new HTTP request to the REST API endpoint
	req, err := http.NewRequest("POST", homedashServer, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Log the response status code for posterity
	body, _ := io.ReadAll(resp.Body)
	log.Infof("%d - %s", resp.StatusCode, body)

	return nil
}

/*
Copyright (C) 2023 Martijn van der Kleijn
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
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"code.vanderkleijn.net/homedash-sidecar/internal/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
	Swarm   bool   `json:"swarm_container"`
}

var logLevel = new(slog.LevelVar)

func main() {
	// Setup logging
	logLevel.Set(slog.LevelInfo) // Default to INFO level
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Set the default logger for packages that might use slog.Default()
	slog.SetDefault(logger)

	config.SetLogger(logger)
	cfg := config.Load()
	logger.Debug("homedash server url set", "url", cfg.Server)

	// Update logging level based on config
	logLevel.Set(cfg.Loglevel.ToSlogLevel())

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("unable to connect with Docker API", "error", err)
		os.Exit(1)
	}

	options := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("status", "running"),
		),
	}

	// Check for old data and clean up every X minutes
	go func() {
		for {
			postApps(getApps(cli, options, cfg), cfg)
			time.Sleep(cfg.Interval)
		}
	}()

	select {}
}

func getApps(cli *client.Client, options container.ListOptions, cfg *config.Config) []ContainerInfo {
	logger := config.GetLogger()

	// Get a list of all running containers
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		logger.Error("unable to determine set of running containers", "error", err)
	}

	// Cache for service labels to avoid repeated API calls
	serviceLabelsCache := make(map[string]map[string]string)

	applications := make([]ContainerInfo, 0)

	// Process each container
	for _, container := range containers {
		isSwarmContainer := false

		// Start with container labels
		labels := make(map[string]string)
		for k, v := range container.Labels {
			if strings.HasPrefix(k, cfg.LabelPrefix) {
				labels[k] = v
			}
		}

		// Check if this is a Swarm service container and get service labels
		if serviceID, exists := container.Labels["com.docker.swarm.service.id"]; exists {
			isSwarmContainer = true

			// Check cache first
			if serviceLabels, cached := serviceLabelsCache[serviceID]; cached {
				logger.Debug("cache hit", "serviceID", serviceID, "labels", serviceLabels)

				// Merge service labels, giving them precedence
				for k, v := range serviceLabels {
					if strings.HasPrefix(k, cfg.LabelPrefix) {
						labels[k] = v
					}
				}
			} else {
				// Fetch service info from Docker API
				service, _, err := cli.ServiceInspectWithRaw(context.Background(), serviceID, types.ServiceInspectOptions{})
				if err != nil {
					logger.Warn("failed to inspect service", "serviceID", serviceID, "error", err)
				} else {
					// Cache the service labels
					serviceLabelsCache[serviceID] = service.Spec.Labels
					logger.Debug("caching service labels", "serviceID", serviceID, "labels", service.Spec.Labels)

					// Merge service labels, giving them precedence
					for k, v := range service.Spec.Labels {
						if strings.HasPrefix(k, cfg.LabelPrefix) {
							labels[k] = v
						}
					}
				}
			}
		}

		// Only include containers that have at least the homedash.name label
		if name, exists := labels[cfg.LabelPrefix+"name"]; exists {
			containerInfo := ContainerInfo{
				Name:    name,
				Url:     labels[cfg.LabelPrefix+"url"],
				Icon:    labels[cfg.LabelPrefix+"icon"],
				Comment: labels[cfg.LabelPrefix+"comment"],
				Swarm:   isSwarmContainer,
			}
			applications = append(applications, containerInfo)
			logger.Debug("found application", "container_info", containerInfo)
		}
	}

	return applications
}

func postApps(applications []ContainerInfo, cfg *config.Config) {
	logger := config.GetLogger()

	logger.Debug("attempting to add apps to server", "apps", applications)

	containerUpdate := ContainerUpdate{
		Uuid:       cfg.UUID,
		Containers: applications,
	}

	payload, err := json.Marshal(containerUpdate)
	if err != nil {
		logger.Error("problem marshalling payload for transmission to server", "error", err)
	}

	// Create a new HTTP request to the REST API endpoint
	logger.Debug("transmitting payload", "payload", string(payload))
	req, err := http.NewRequest("POST", cfg.Server, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("problem creating HTTP request", "error", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("problem transmitting payload to server", "error", err)
	}
	defer resp.Body.Close()

	// Log the response status code for posterity
	body, _ := io.ReadAll(resp.Body)
	logger.Debug("server response", "status_code", resp.StatusCode, "body", string(body))
}

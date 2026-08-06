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
	"github.com/docker/docker/api/types/swarm"
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

	client := &http.Client{Timeout: 30 * time.Second}

	// Check for old data and clean up every X minutes
	go func() {
		for {
			postApps(client, getApps(cli, options, cfg), cfg)
			time.Sleep(cfg.Interval)
		}
	}()

	select {}
}

func getApps(cli *client.Client, options container.ListOptions, cfg *config.Config) []ContainerInfo {
	logger := config.GetLogger()

	applications := make([]ContainerInfo, 0)

	// Try Swarm mode first (cluster-wide discovery from manager)
	swarmApps, isSwarmManager := getSwarmApps(cli, cfg)
	if isSwarmManager {
		logger.Debug("running on a Swarm manager, using cluster-wide service discovery")
		applications = append(applications, swarmApps...)
	} else {
		logger.Debug("not a Swarm manager (or Swarm inactive), skipping cluster-wide service discovery")
	}

	// Also check local containers for:
	// - Non-Swarm containers (standalone Docker / Compose)
	// - Swarm containers with container-level labels that override service labels
	localApps := getLocalContainerApps(cli, options, cfg)

	// Merge, avoiding duplicates (prefer local container labels over service labels)
	applications = mergeApps(applications, localApps)

	return applications
}

// getSwarmApps queries all services in the Swarm cluster (manager-only).
// It returns (apps, true) if this node is an active Swarm manager, otherwise (nil, false).
func getSwarmApps(cli *client.Client, cfg *config.Config) ([]ContainerInfo, bool) {
	logger := config.GetLogger()

	info, err := cli.Info(context.Background())
	if err != nil {
		logger.Debug("failed to get Docker info", "error", err)
		return nil, false
	}

	// Not in Swarm mode at all?
	if info.Swarm.LocalNodeState != swarm.LocalNodeStateActive {
		logger.Debug("node is not in an active Swarm state", "state", info.Swarm.LocalNodeState)
		return nil, false
	}

	// Must be a manager to access cluster state
	if !info.Swarm.ControlAvailable {
		logger.Debug("node is not a Swarm manager, cannot query services")
		return nil, false
	}

	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		logger.Warn("failed to list Swarm services", "error", err)
		return nil, false
	}

	logger.Debug("found Swarm services", "count", len(services))

	applications := make([]ContainerInfo, 0)

	for _, service := range services {
		labels := make(map[string]string)

		// Extract homedash labels from service
		for k, v := range service.Spec.Labels {
			if strings.HasPrefix(k, cfg.LabelPrefix) {
				labels[k] = v
			}
		}

		// Only include services that have at least the homedash.name label
		if name, exists := labels[cfg.LabelPrefix+"name"]; exists {
			containerInfo := ContainerInfo{
				Name:    name,
				Url:     labels[cfg.LabelPrefix+"url"],
				Icon:    labels[cfg.LabelPrefix+"icon"],
				Comment: labels[cfg.LabelPrefix+"comment"],
				Swarm:   true,
			}
			applications = append(applications, containerInfo)
			logger.Debug("found Swarm service", "service", service.Spec.Name, "container_info", containerInfo)
		}
	}

	return applications, true
}

// getLocalContainerApps uses local container inspection (Docker / Compose / Swarm containers
// with container-level labels).
func getLocalContainerApps(cli *client.Client, options container.ListOptions, cfg *config.Config) []ContainerInfo {
	logger := config.GetLogger()

	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		logger.Error("unable to determine set of running containers", "error", err)
		return nil
	}

	applications := make([]ContainerInfo, 0)

	for _, container := range containers {
		isSwarmContainer := false

		labels := make(map[string]string)
		for k, v := range container.Labels {
			if strings.HasPrefix(k, cfg.LabelPrefix) {
				labels[k] = v
			}
		}

		if _, exists := container.Labels["com.docker.swarm.service.id"]; exists {
			isSwarmContainer = true
			// We no longer inspect the service here; getSwarmApps already handles service-level labels.
			// This function is just for container-level labels.
		}

		if name, exists := labels[cfg.LabelPrefix+"name"]; exists {
			containerInfo := ContainerInfo{
				Name:    name,
				Url:     labels[cfg.LabelPrefix+"url"],
				Icon:    labels[cfg.LabelPrefix+"icon"],
				Comment: labels[cfg.LabelPrefix+"comment"],
				Swarm:   isSwarmContainer,
			}
			applications = append(applications, containerInfo)
			logger.Debug("found local container application", "container_info", containerInfo)
		}
	}

	return applications
}

// mergeApps combines Swarm service apps and local container apps, avoiding duplicates.
// Local container labels take precedence over service labels (same name -> override).
func mergeApps(swarmApps, localApps []ContainerInfo) []ContainerInfo {
	appMap := make(map[string]ContainerInfo)

	for _, app := range swarmApps {
		appMap[app.Name] = app
	}
	for _, app := range localApps {
		appMap[app.Name] = app
	}

	result := make([]ContainerInfo, 0, len(appMap))
	for _, app := range appMap {
		result = append(result, app)
	}

	return result
}

func postApps(client *http.Client, applications []ContainerInfo, cfg *config.Config) {
	logger := config.GetLogger()

	logger.Debug("attempting to add apps to server", "apps", applications)

	containerUpdate := ContainerUpdate{
		Uuid:       cfg.UUID,
		Containers: applications,
	}

	payload, err := json.Marshal(containerUpdate)
	if err != nil {
		logger.Error("problem marshalling payload for transmission to server", "error", err)
		return
	}

	// Create a new HTTP request to the REST API endpoint
	logger.Debug("transmitting payload", "payload", string(payload))
	req, err := http.NewRequest("POST", cfg.Server, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("problem creating HTTP request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("problem transmitting payload to server", "error", err)
		return
	}
	defer resp.Body.Close()

	// Log the response status code for posterity
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		logger.Error("problem reading server response", "error", err)
		return
	}
	logger.Debug("server response", "status_code", resp.StatusCode, "body", string(body))
}

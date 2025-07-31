/*
Copyright (C) 2025 Martijn van der Kleijn
This file is part of HomeDash sidecar.

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package config

import (
	"log/slog"
	"os"
	"time"

	"code.vanderkleijn.net/homedash-sidecar/internal/log"
	"github.com/google/uuid"
)

type Config struct {
	Loglevel    log.LogLevel
	Server      string
	Interval    time.Duration
	UUID        string
	LabelPrefix string
}

const (
	envVarLogLevel string = "HOMEDASH_LOG_LEVEL"
	envVarServer   string = "HOMEDASH_SERVER"
	envVarInterval string = "HOMEDASH_INTERVAL"
	envVarUUID     string = "HOMEDASH_SIDECAR_UUID"
	envVarPrefix   string = "HOMEDASH_LABEL_PREFIX"
)

var logger *slog.Logger

func SetLogger(slogger *slog.Logger) {
	logger = slogger
}

func GetLogger() *slog.Logger {
	return logger
}

func Load() *Config {
	logger.Info("loading configuration")

	logLevelStr := getEnv(envVarLogLevel, "INFO")
	logLevel, err := log.ParseLogLevel(logLevelStr)
	if err != nil {
		logger.Info("invalid log level specified", "level", logLevelStr)
	}

	intervalStr := getEnv(envVarInterval, "10m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		logger.Error("unable to parse interval string, using default", "default", "10m")
	}

	logger.Debug("interval set to once per N minutes", "interval", interval.Minutes())

	return &Config{
		Loglevel:    logLevel,
		Server:      getEnv(envVarServer, "") + "/api/v1/applications",
		Interval:    interval,
		UUID:        getEnv(envVarUUID, uuid.New().String()),
		LabelPrefix: getEnv(envVarPrefix, "homedash") + ".",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Let's inform the user if default values are used
	switch key {
	case envVarServer:
		logger.Error("required environment variable not set or empty", "env_var", envVarServer)
		os.Exit(1)
	case envVarUUID:
		logger.Warn("using generated uuid", "uuid", fallback, "env_var", envVarUUID)
	case envVarInterval:
		logger.Warn("using default interval (once per N minutes)", "default", "10m", "env_var", envVarInterval)
	case envVarLogLevel:
		logger.Info("using default loglevel", "default", "INFO", "env_var", envVarLogLevel)
	case envVarPrefix:
		logger.Info("using default label prefix", "default", fallback, "env_var", envVarPrefix)
	}

	return fallback
}

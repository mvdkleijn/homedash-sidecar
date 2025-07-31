/*
Copyright (C) 2025 Martijn van der Kleijn
This file is part of HomeDash sidecar.

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package log

import (
	"fmt"
	"log/slog"
	"strings"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ToSlogLevel converts your LogLevel to slog.Level
func (l LogLevel) ToSlogLevel() slog.Level {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseLogLevel parses a string (case-insensitive) to LogLevel,
// returns error if invalid as well as returning default of LogLevelInfo.
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LogLevelDebug, nil
	case "INFO":
		return LogLevelInfo, nil
	case "WARN", "WARNING":
		return LogLevelWarn, nil
	case "ERROR":
		return LogLevelError, nil
	default:
		return LogLevelInfo, fmt.Errorf("invalid log level: %q", s)
	}
}

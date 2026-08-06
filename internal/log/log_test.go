package log

import "testing"

func TestParseLogLevel(t *testing.T) {
	cases := map[string]LogLevel{
		"DEBUG":   LogLevelDebug,
		"info":    LogLevelInfo,
		"Warn":    LogLevelWarn,
		"warning": LogLevelWarn,
		"ERROR":   LogLevelError,
	}

	for in, want := range cases {
		got, err := ParseLogLevel(in)
		if err != nil {
			t.Errorf("ParseLogLevel(%q) unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseLogLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseLogLevelInvalid(t *testing.T) {
	got, err := ParseLogLevel("bogus")
	if err == nil {
		t.Error("expected error for invalid log level")
	}
	if got != LogLevelInfo {
		t.Errorf("expected LogLevelInfo fallback, got %v", got)
	}
}

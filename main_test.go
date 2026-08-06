package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vanderkleijn.net/homedash-sidecar/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func init() {
	config.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func testConfig(server string) *config.Config {
	return &config.Config{
		Server: server,
		UUID:   "test-uuid",
	}
}

func TestPostAppsTransportErrorDoesNotPanic(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("postApps panicked on transport error: %v", r)
		}
	}()

	postApps(client, []ContainerInfo{{Name: "Test", Url: "http://example.test"}}, testConfig("http://server.test/api/v1/applications"))
}

func TestRunUpdateCycleRecoversFromPanic(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:1"))
	if err != nil {
		t.Fatalf("unable to create docker client: %v", err)
	}

	httpClient := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			panic("boom")
		}),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("runUpdateCycle did not recover panic: %v", r)
		}
	}()

	runUpdateCycle(httpClient, dockerClient, container.ListOptions{}, testConfig("http://server.test/api/v1/applications"))
}

func TestPostAppsSendsPayload(t *testing.T) {
	var got ContainerUpdate
	contentType := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("unable to decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	apps := []ContainerInfo{{Name: "Forgejo", Url: "https://git.example", Icon: "forgejo", Swarm: true}}
	postApps(server.Client(), apps, testConfig(server.URL))

	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %q", contentType)
	}
	if got.Uuid != "test-uuid" {
		t.Errorf("expected uuid test-uuid, got %q", got.Uuid)
	}
	if len(got.Containers) != 1 || got.Containers[0].Name != "Forgejo" {
		t.Errorf("unexpected containers: %+v", got.Containers)
	}
}

func TestMergeAppsLocalOverridesService(t *testing.T) {
	service := []ContainerInfo{{Name: "App", Url: "https://service.example", Icon: "old", Swarm: true}}
	local := []ContainerInfo{{Name: "App", Url: "https://local.example", Icon: "new", Swarm: false}}

	merged := mergeApps(service, local)

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged app, got %d", len(merged))
	}
	if merged[0].Url != "https://local.example" {
		t.Errorf("expected container-level url to take precedence, got %q", merged[0].Url)
	}
	if merged[0].Swarm {
		t.Error("expected merged app to keep container-level Swarm flag")
	}
}

func TestMergeAppsDeduplicatesByName(t *testing.T) {
	service := []ContainerInfo{{Name: "App", Url: "https://a.example"}, {Name: "Other", Url: "https://other.example"}}
	local := []ContainerInfo{{Name: "App", Url: "https://b.example"}}

	merged := mergeApps(service, local)

	if len(merged) != 2 {
		t.Errorf("expected 2 unique apps, got %d: %+v", len(merged), merged)
	}
}

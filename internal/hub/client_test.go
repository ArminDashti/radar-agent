package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientAuthenticatesAndExchangesSamples(t *testing.T) {
	var posted SampleBatch
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/agent/targets":
			_ = json.NewEncoder(w).Encode([]Target{{ID: 7, Host: "example.com", Protocols: []string{"http"}}})
		case "/api/agent/samples":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusAccepted)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", "secret", time.Second)
	targets, err := client.Targets(context.Background())
	if err != nil || len(targets) != 1 || targets[0].ID != 7 {
		t.Fatalf("targets = %#v, err = %v", targets, err)
	}
	latency := 12.5
	sample := Sample{EndpointID: 7, Protocol: "http", ObservedAt: time.Now(), LatencyMS: &latency, OK: true}
	if err := client.Submit(context.Background(), []Sample{sample}); err != nil {
		t.Fatal(err)
	}
	if len(posted.Samples) != 1 || posted.Samples[0].EndpointID != 7 {
		t.Fatalf("posted = %#v", posted)
	}
}

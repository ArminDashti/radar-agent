package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("RADAR_AGENT_TOKEN", "token")
	t.Setenv("RADAR_HUB_URL", "")
	t.Setenv("RADAR_HTTP_TIMEOUT_SEC", "")
	t.Setenv("RADAR_ICMP_TIMEOUT_SEC", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HubURL != "http://127.0.0.1:8088" {
		t.Fatalf("HubURL = %q", cfg.HubURL)
	}
	if cfg.HTTPTimeout != 5*time.Second || cfg.ICMPTimeout != 3*time.Second {
		t.Fatalf("unexpected timeouts: %v, %v", cfg.HTTPTimeout, cfg.ICMPTimeout)
	}
}

func TestLoadRequiresToken(t *testing.T) {
	t.Setenv("RADAR_AGENT_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing token error")
	}
}

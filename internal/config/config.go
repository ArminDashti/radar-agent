package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HubURL      string
	AgentToken  string
	HTTPTimeout time.Duration
	ICMPTimeout time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	httpTimeout, err := durationFromEnv("RADAR_HTTP_TIMEOUT_SEC", 5)
	if err != nil {
		return Config{}, err
	}
	icmpTimeout, err := durationFromEnv("RADAR_ICMP_TIMEOUT_SEC", 3)
	if err != nil {
		return Config{}, err
	}

	token := strings.TrimSpace(os.Getenv("RADAR_AGENT_TOKEN"))
	if token == "" {
		return Config{}, errors.New("RADAR_AGENT_TOKEN is required")
	}
	hubURL := strings.TrimRight(strings.TrimSpace(os.Getenv("RADAR_HUB_URL")), "/")
	if hubURL == "" {
		hubURL = "http://127.0.0.1:8088"
	}
	return Config{
		HubURL: hubURL, AgentToken: token,
		HTTPTimeout: httpTimeout, ICMPTimeout: icmpTimeout,
	}, nil
}

func durationFromEnv(name string, fallback int) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		value = strconv.Itoa(fallback)
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return time.Duration(seconds) * time.Second, nil
}

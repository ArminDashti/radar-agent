package probe

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func HTTP(ctx context.Context, host string, timeout time.Duration) (*float64, bool) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalizeURL(host), nil)
	if err != nil {
		return nil, false
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	latency := float64(time.Since(start).Microseconds()) / 1000
	resp.Body.Close()
	return &latency, true
}

func normalizeURL(host string) string {
	host = strings.TrimSpace(host)
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "https://" + host
}

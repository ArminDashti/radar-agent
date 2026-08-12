package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Target struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Host      string   `json:"host"`
	Protocols []string `json:"protocols"`
}

type Sample struct {
	EndpointID int64     `json:"endpoint_id"`
	Protocol   string    `json:"protocol"`
	ObservedAt time.Time `json:"observed_at"`
	LatencyMS  *float64  `json:"latency_ms"`
	OK         bool      `json:"ok"`
}

type SampleBatch struct {
	Samples []Sample `json:"samples"`
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Targets(ctx context.Context) ([]Target, error) {
	req, err := c.request(ctx, http.MethodGet, "/api/agent/targets", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch targets: %w", err)
	}
	defer resp.Body.Close()
	if err := responseError(resp); err != nil {
		return nil, fmt.Errorf("fetch targets: %w", err)
	}
	var targets []Target
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return nil, fmt.Errorf("decode targets: %w", err)
	}
	return targets, nil
}

func (c *Client) Submit(ctx context.Context, samples []Sample) error {
	body, err := json.Marshal(SampleBatch{Samples: samples})
	if err != nil {
		return fmt.Errorf("encode samples: %w", err)
	}
	req, err := c.request(ctx, http.MethodPost, "/api/agent/samples", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("submit samples: %w", err)
	}
	defer resp.Body.Close()
	if err := responseError(resp); err != nil {
		return fmt.Errorf("submit samples: %w", err)
	}
	return nil
}

func (c *Client) request(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func responseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("hub returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

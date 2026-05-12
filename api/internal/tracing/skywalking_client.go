package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// SWClient is a thin GraphQL client talking to SkyWalking OAP.
type SWClient struct {
	baseURL string
	http    *http.Client
}

// DefaultSkyWalkingURL is used when no override is configured.
const DefaultSkyWalkingURL = "http://127.0.0.1:12800/graphql"

// NewSWClient constructs a client, honoring `skywalking.graphql_url` from viper.
func NewSWClient() *SWClient {
	url := viper.GetString("skywalking.graphql_url")
	if url == "" {
		url = DefaultSkyWalkingURL
	}
	return &SWClient{
		baseURL: url,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// BaseURL exposes the resolved backend URL (for diagnostics).
func (c *SWClient) BaseURL() string { return c.baseURL }

// Query executes a GraphQL request, decoding the `data` field into out.
// Errors from SkyWalking are surfaced verbatim for upstream handling.
func (c *SWClient) Query(ctx context.Context, query string, variables map[string]any, out any) error {
	payload := map[string]any{"query": query}
	if variables != nil {
		payload["variables"] = variables
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal graphql payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("skywalking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("skywalking returned %d: %s", resp.StatusCode, string(raw))
	}

	var wrapper struct {
		Data   json.RawMessage  `json:"data"`
		Errors []map[string]any `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}
	if len(wrapper.Errors) > 0 {
		return fmt.Errorf("graphql errors: %v", wrapper.Errors)
	}
	if out == nil {
		return nil
	}
	if len(wrapper.Data) == 0 || string(wrapper.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(wrapper.Data, out); err != nil {
		return fmt.Errorf("unmarshal graphql data: %w", err)
	}
	return nil
}

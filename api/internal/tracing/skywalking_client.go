package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// SWClient is a thin GraphQL client talking to SkyWalking OAP.
type SWClient struct {
	baseURL string
	http    *http.Client
}

// NewSWClientForTest constructs a client with explicit transport for focused tests.
func NewSWClientForTest(baseURL string, httpClient *http.Client) *SWClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &SWClient{baseURL: baseURL, http: httpClient}
}

// ErrNotConfigured marks the missing OAP GraphQL contract explicitly.
var ErrNotConfigured = errors.New("skywalking graphql endpoint not configured")

// DefaultSkyWalkingURL is kept for compatibility documentation; it is not used
// implicitly because an unverified local OAP would make contract gaps look ready.
const DefaultSkyWalkingURL = "http://127.0.0.1:12800/graphql"

// NewSWClient constructs a client, honoring `skywalking.graphql_url` from viper.
func NewSWClient() *SWClient {
	return &SWClient{
		baseURL: configuredSkyWalkingEndpoint(),
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// BaseURL exposes the resolved backend URL (for diagnostics).
func (c *SWClient) BaseURL() string {
	if c == nil {
		return ""
	}
	if strings.TrimSpace(c.baseURL) != "" {
		return c.baseURL
	}
	return configuredSkyWalkingEndpoint()
}

// Query executes a GraphQL request, decoding the `data` field into out.
// Errors from SkyWalking are intentionally sanitized for upstream handling.
func (c *SWClient) Query(ctx context.Context, query string, variables map[string]any, out any) error {
	endpoint := strings.TrimSpace(c.BaseURL())
	if endpoint == "" {
		return ErrNotConfigured
	}
	payload := map[string]any{"query": query}
	if variables != nil {
		payload["variables"] = variables
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal graphql payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build graphql request: %w", ErrNotConfigured)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return errors.New("skywalking request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("skywalking returned status %d", resp.StatusCode)
	}

	var wrapper struct {
		Data   json.RawMessage  `json:"data"`
		Errors []map[string]any `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}
	if len(wrapper.Errors) > 0 {
		return errors.New("skywalking graphql errors returned")
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

func configuredSkyWalkingEndpoint() string {
	for _, value := range []string{
		viper.GetString("skywalking.graphql_url"),
		os.Getenv("SKYWALKING_GRAPHQL_URL"),
		os.Getenv("SKYWALKING_OAP_GRAPHQL_URL"),
	} {
		if endpoint := strings.TrimSpace(value); endpoint != "" {
			return endpoint
		}
	}
	if oap := strings.TrimRight(strings.TrimSpace(os.Getenv("SKYWALKING_OAP_URL")), "/"); oap != "" {
		if strings.HasSuffix(oap, "/graphql") {
			return oap
		}
		return oap + "/graphql"
	}
	return ""
}

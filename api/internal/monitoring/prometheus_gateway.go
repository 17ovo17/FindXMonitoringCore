package monitoring

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	DefaultInstantTimeout = 5 * time.Second
	DefaultRangeTimeout   = 10 * time.Second
)

var (
	ErrDatasourceNotFound = errors.New("prometheus datasource not found")
	ErrInvalidQuery       = errors.New("invalid prometheus query")
	ErrUpstream           = errors.New("prometheus upstream unavailable")
	ErrInvalidResponse    = errors.New("invalid prometheus response")

	sensitiveWarningPattern = regexp.MustCompile(`(?i)(api[_-]?key|apikey|authorization|password|private|secret|token|cookie|dsn|auth|<\s*(token|password|secret|cookie|auth|authorization|api[_-]?key|apikey|dsn|private)\s*>|%3c\s*(token|password|secret|cookie|auth|authorization|api[_-]?key|apikey|dsn|private)\s*%3e)`)
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type PrometheusGateway struct {
	client HTTPClient
}

type PrometheusCallRequest struct {
	BaseURL string
	Path    string
	Params  url.Values
	Timeout time.Duration
}

type PrometheusQueryRequest struct {
	BaseURL string
	Query   string
	Time    float64
	Timeout time.Duration
}

type PrometheusCallResult struct {
	Body       map[string]any
	Data       map[string]any
	Stats      ResultStats
	Warnings   []string
	LatencyMS  int64
	StatusCode int
	QueryHash  string
}

type ResultStats struct {
	SeriesCount int    `json:"series_count"`
	SampleCount int    `json:"sample_count"`
	ResultType  string `json:"result_type"`
}

type Datasource struct {
	ID   string
	Type string
	URL  string
}

type GatewayError struct {
	Kind       error
	StatusCode int
	Message    string
	Err        error
}

func (e *GatewayError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "prometheus gateway error"
}

func (e *GatewayError) Unwrap() error { return e.Kind }

func NewPrometheusGateway(client HTTPClient) *PrometheusGateway {
	return &PrometheusGateway{client: client}
}

func (g *PrometheusGateway) QueryInstant(ctx context.Context, req PrometheusQueryRequest) (PrometheusCallResult, error) {
	params := url.Values{"query": []string{strings.TrimSpace(req.Query)}}
	if req.Time > 0 {
		params.Set("time", fmt.Sprintf("%v", req.Time))
	}
	out, err := g.Call(ctx, PrometheusCallRequest{
		BaseURL: req.BaseURL, Path: "/api/v1/query", Params: params,
		Timeout: TimeoutOrDefault(req.Timeout, DefaultInstantTimeout),
	})
	out.QueryHash = QueryHash(req.Query)
	return out, err
}

func (g *PrometheusGateway) Call(ctx context.Context, req PrometheusCallRequest) (PrometheusCallResult, error) {
	timeout := TimeoutOrDefault(req.Timeout, DefaultInstantTimeout)
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	endpoint, err := prometheusEndpoint(req.BaseURL, req.Path, req.Params)
	if err != nil {
		return PrometheusCallResult{StatusCode: http.StatusBadRequest}, safeError(ErrInvalidResponse, http.StatusBadRequest)
	}
	started := time.Now()
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return PrometheusCallResult{StatusCode: http.StatusBadRequest}, safeError(ErrInvalidResponse, http.StatusBadRequest)
	}
	client := g.client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	resp, err := client.Do(httpReq)
	out := PrometheusCallResult{LatencyMS: time.Since(started).Milliseconds(), StatusCode: http.StatusServiceUnavailable}
	if err != nil {
		return out, wrapSafeError(ErrUpstream, http.StatusServiceUnavailable, err)
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if readErr != nil {
		return out, wrapSafeError(ErrUpstream, http.StatusServiceUnavailable, readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, safeError(ErrUpstream, http.StatusServiceUnavailable)
	}
	return decodePrometheusResponse(raw, out)
}

func QueryHash(query string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(query)))
	return fmt.Sprintf("%x", sum[:])
}

func ValidatePromQL(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("%w: query required", ErrInvalidQuery)
	}
	if len(query) > 4096 {
		return fmt.Errorf("%w: query too long", ErrInvalidQuery)
	}
	for _, r := range query {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: control characters rejected", ErrInvalidQuery)
		}
	}
	lower := strings.ToLower(query)
	for _, term := range []string{"delete_series", "/api/v1/admin", "api/v1/admin", "/admin/", " admin "} {
		if strings.Contains(lower, term) {
			return fmt.Errorf("%w: rejected by safety policy", ErrInvalidQuery)
		}
	}
	return nil
}

func ResultStatsForData(data map[string]any) ResultStats {
	resultType, _ := data["resultType"].(string)
	series, samples := countPrometheusResult(resultType, data["result"])
	return ResultStats{SeriesCount: series, SampleCount: samples, ResultType: resultType}
}

func ResolvePrometheusDatasource(sources []Datasource, datasourceID, defaultID, fallbackURL string) (string, string, error) {
	id := strings.TrimSpace(datasourceID)
	if id == "" {
		id = strings.TrimSpace(defaultID)
	}
	for _, ds := range sources {
		if ds.ID != id {
			continue
		}
		if !strings.EqualFold(ds.Type, "prometheus") {
			return "", "", ErrDatasourceNotFound
		}
		if base := strings.TrimRight(strings.TrimSpace(firstNonEmpty(ds.URL, fallbackURL)), "/"); base != "" {
			return id, base, nil
		}
		return "", "", ErrDatasourceNotFound
	}
	if id == defaultID {
		if base := strings.TrimRight(strings.TrimSpace(fallbackURL), "/"); base != "" {
			return id, base, nil
		}
	}
	return "", "", ErrDatasourceNotFound
}

func HTTPStatus(err error) int {
	var gatewayErr *GatewayError
	if errors.As(err, &gatewayErr) && gatewayErr.StatusCode > 0 {
		return gatewayErr.StatusCode
	}
	if errors.Is(err, ErrDatasourceNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrInvalidQuery) {
		return http.StatusBadRequest
	}
	return http.StatusServiceUnavailable
}

func TimeoutOrDefault(timeout, fallback time.Duration) time.Duration {
	if timeout <= 0 {
		return fallback
	}
	if timeout > 30*time.Second {
		return 30 * time.Second
	}
	return timeout
}

func decodePrometheusResponse(raw []byte, out PrometheusCallResult) (PrometheusCallResult, error) {
	if err := json.Unmarshal(raw, &out.Body); err != nil {
		out.StatusCode = http.StatusServiceUnavailable
		return out, wrapSafeError(ErrInvalidResponse, http.StatusServiceUnavailable, err)
	}
	out.Warnings = prometheusWarnings(out.Body["warnings"])
	if len(out.Warnings) > 0 {
		out.Body["warnings"] = sanitizedWarningBody(out.Warnings)
	}
	if status, _ := out.Body["status"].(string); status != "success" {
		out.StatusCode = http.StatusServiceUnavailable
		return out, safeError(ErrUpstream, http.StatusServiceUnavailable)
	}
	data, ok := out.Body["data"].(map[string]any)
	if !ok {
		out.StatusCode = http.StatusServiceUnavailable
		return out, safeError(ErrInvalidResponse, http.StatusServiceUnavailable)
	}
	out.Data = data
	out.Stats = ResultStatsForData(data)
	out.StatusCode = http.StatusOK
	return out, nil
}

func prometheusEndpoint(base, path string, params url.Values) (string, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("invalid prometheus endpoint")
	}
	endpoint := base + path
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	return endpoint, nil
}

func prometheusWarnings(raw any) []string {
	items, _ := raw.([]any)
	warnings := make([]string, 0, len(items))
	for _, item := range items {
		if warning, ok := item.(string); ok {
			warnings = append(warnings, sanitizePrometheusWarning(warning))
		}
	}
	return warnings
}

func sanitizedWarningBody(warnings []string) []any {
	out := make([]any, 0, len(warnings))
	for _, warning := range warnings {
		out = append(out, warning)
	}
	return out
}

func sanitizePrometheusWarning(warning string) string {
	warning = strings.TrimSpace(warning)
	if warning == "" {
		return warning
	}
	if sensitiveWarningPattern.MatchString(warning) || warningHasSensitiveQuery(warning) {
		return "prometheus warning redacted: <REDACTED>"
	}
	return warning
}

func warningHasSensitiveQuery(warning string) bool {
	for _, field := range strings.Fields(warning) {
		candidate := strings.Trim(field, `"'(),;`)
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.RawQuery == "" {
			continue
		}
		for key, values := range parsed.Query() {
			if sensitiveWarningPattern.MatchString(key) {
				return true
			}
			for _, value := range values {
				if sensitiveWarningPattern.MatchString(value) {
					return true
				}
			}
		}
	}
	return false
}

func countPrometheusResult(resultType string, result any) (int, int) {
	rows, ok := result.([]any)
	if !ok {
		if result != nil && (resultType == "scalar" || resultType == "string") {
			return 1, 1
		}
		return 0, 0
	}
	if resultType != "matrix" {
		return len(rows), len(rows)
	}
	samples := 0
	for _, item := range rows {
		row, _ := item.(map[string]any)
		values, _ := row["values"].([]any)
		samples += len(values)
	}
	return len(rows), samples
}

func safeError(kind error, status int) error {
	return &GatewayError{Kind: kind, StatusCode: status, Message: kind.Error()}
}

func wrapSafeError(kind error, status int, err error) error {
	return &GatewayError{Kind: kind, StatusCode: status, Message: kind.Error(), Err: err}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

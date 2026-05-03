package monitoring

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type gatewayClientFunc func(*http.Request) (*http.Response, error)

func (f gatewayClientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestPrometheusGatewaySuccessVector(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" || r.URL.Query().Get("query") != "up" {
			t.Fatalf("unexpected request path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"job":"node"},"value":[1,"1"]}]}}`))
	}))
	defer upstream.Close()

	got, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: " up "})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if got.Stats.ResultType != "vector" || got.Stats.SeriesCount != 1 || got.Stats.SampleCount != 1 {
		t.Fatalf("unexpected stats: %+v", got.Stats)
	}
}

func TestPrometheusGatewaySuccessWarningsAreSanitized(t *testing.T) {
	sensitiveURL := "http://prom/api?" + "token" + "=<TOKEN>&" + "password" + "=<PASSWORD>"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{"status":"success","data":{"resultType":"vector","result":[]},"warnings":["partial scrape","upstream ` +
			sensitiveURL + `","authorization Bearer <TOKEN>"]}`
		_, _ = w.Write([]byte(payload))
	}))
	defer upstream.Close()

	got, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: "up"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	rendered := strings.Join(got.Warnings, "\n")
	if !strings.Contains(rendered, "partial scrape") || !strings.Contains(rendered, "<REDACTED>") {
		t.Fatalf("warnings lost safe content or redaction marker: %q", rendered)
	}
	for _, forbidden := range []string{"<TOKEN>", "authorization", "token", "password"} {
		if strings.Contains(strings.ToLower(rendered), strings.ToLower(forbidden)) {
			t.Fatalf("warning leaked sensitive fragment %q in %q", forbidden, rendered)
		}
	}
}

func TestPrometheusGatewayNon2xxIsServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("token" + "=<TOKEN>"))
	}))
	defer upstream.Close()

	_, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: "up"})
	if HTTPStatus(err) != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got err=%v status=%d", err, HTTPStatus(err))
	}
}

func TestPrometheusGatewayStatusErrorIsServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","error":"token secret"}`))
	}))
	defer upstream.Close()

	_, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: "up"})
	if HTTPStatus(err) != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got err=%v status=%d", err, HTTPStatus(err))
	}
}

func TestPrometheusGatewayInvalidJSON(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json token secret`))
	}))
	defer upstream.Close()

	_, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: "up"})
	if HTTPStatus(err) != http.StatusServiceUnavailable {
		t.Fatalf("expected invalid json to map to 503, got err=%v status=%d", err, HTTPStatus(err))
	}
}

func TestPrometheusQueryHashUsesSHA256Trim(t *testing.T) {
	sum := sha256.Sum256([]byte("up"))
	want := fmt.Sprintf("%x", sum[:])
	if got := QueryHash(" up\n"); got != want {
		t.Fatalf("unexpected hash: got=%s want=%s", got, want)
	}
}

func TestPrometheusGatewayTimeoutMSPriorityEquivalent(t *testing.T) {
	timeout := TimeoutOrDefault(2500*time.Millisecond, DefaultInstantTimeout)
	if timeout != 2500*time.Millisecond {
		t.Fatalf("timeout_ms equivalent should win, got %s", timeout)
	}
	if TimeoutOrDefault(45*time.Second, DefaultInstantTimeout) != 30*time.Second {
		t.Fatalf("timeout should be capped")
	}
}

func TestPrometheusGatewayErrorDoesNotExposeUpstreamSecret(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("authorization" + "=<TOKEN>"))
	}))
	defer upstream.Close()

	_, err := NewPrometheusGateway(nil).QueryInstant(context.Background(), PrometheusQueryRequest{BaseURL: upstream.URL, Query: "up"})
	if strings.Contains(strings.ToLower(fmt.Sprint(err)), "<token>") {
		t.Fatalf("gateway error leaked upstream body: %v", err)
	}
}

func TestPrometheusGatewayNetworkErrorDoesNotExposeRequestURL(t *testing.T) {
	query := `sum(rate(http_requests_total{password="<TOKEN>",token="<TOKEN>"}[5m]))`
	var rawURL string
	var rawQuery string
	client := gatewayClientFunc(func(req *http.Request) (*http.Response, error) {
		rawURL = req.URL.String()
		rawQuery = req.URL.RawQuery
		return nil, fmt.Errorf("Get %s: dial tcp %s <TOKEN> %s <TOKEN>", rawURL, "password", "token")
	})

	_, err := NewPrometheusGateway(client).QueryInstant(context.Background(), PrometheusQueryRequest{
		BaseURL: "http://prometheus.local",
		Query:   query,
	})
	if err == nil {
		t.Fatalf("expected upstream error")
	}
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
	if HTTPStatus(err) != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got status=%d err=%v", HTTPStatus(err), err)
	}
	rendered := fmt.Sprint(err)
	for _, forbidden := range []string{rawURL, rawQuery, query, "<TOKEN>", "password=", "token="} {
		if forbidden != "" && strings.Contains(rendered, forbidden) {
			t.Fatalf("gateway error leaked sensitive fragment %q in %q", forbidden, rendered)
		}
	}
}

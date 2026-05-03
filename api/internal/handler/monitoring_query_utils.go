package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/monitoring"

	"github.com/gin-gonic/gin"
)

func resolveMonitoringPrometheus(c *gin.Context, datasourceID string) (string, string, bool) {
	datasourceID = normalizeMonitoringDatasourceID(datasourceID)
	dsID, base, err := monitoring.ResolvePrometheusDatasource(
		monitoringDatasources(), datasourceID, defaultMonitoringDatasourceID, resolvePrometheusURL(defaultPrometheusDatasourceID()),
	)
	if err == nil {
		return dsID, base, true
	}
	writeMonitorError(c, http.StatusNotFound, "prometheus datasource not found")
	return "", "", false
}

func monitoringDatasources() []monitoring.Datasource {
	return monitoring.PrometheusDatasourcesFromConfig()
}

func sanitizeDatasourceURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" {
		return ""
	}
	u.User = nil
	q := u.Query()
	for key := range q {
		if sensitiveURLKey(key) {
			q.Set(key, "<REDACTED>")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func sensitiveURLKey(key string) bool {
	key = strings.ToLower(key)
	for _, marker := range []string{"token", "key", "password", "secret", "cookie", "auth"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func prometheusMatchParams(c *gin.Context) url.Values {
	params := url.Values{}
	for _, value := range append(c.QueryArray("match[]"), c.QueryArray("match")...) {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			params.Add("match[]", trimmed)
		}
	}
	return params
}

func timeoutOrDefault(seconds float64, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	timeout := time.Duration(seconds * float64(time.Second))
	if timeout > 30*time.Second {
		return 30 * time.Second
	}
	return timeout
}

func boundedPositiveInt(raw string, fallback, max int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func normalizeMonitoringDatasourceID(datasourceID string) string {
	if strings.TrimSpace(datasourceID) == "" {
		return defaultMonitoringDatasourceID
	}
	return strings.TrimSpace(datasourceID)
}

func writeMonitorError(c *gin.Context, code int, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "request failed"
	}
	for _, marker := range []string{"token", "password", "secret", "cookie", "authorization", "dsn"} {
		if strings.Contains(strings.ToLower(message), marker) {
			message = "request failed"
			break
		}
	}
	c.JSON(code, gin.H{"error": message})
}

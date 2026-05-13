package handler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	lokiProxyTimeout   = 30 * time.Second
	lokiTailSSETimeout = 5 * time.Minute
)

var lokiHTTPClient = &http.Client{Timeout: lokiProxyTimeout}

func getLokiURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("LOKI_URL")), "/")
}

func LogsQueryProxy(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		blockLokiContract(c, http.StatusServiceUnavailable, "loki datasource is not configured")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read request body failed"})
		return
	}
	defer c.Request.Body.Close()

	target := lokiURL + "/loki/api/v1/query_range"
	req, err := newLokiRequest(c, http.MethodPost, target, strings.NewReader(string(body)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki request failed"})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	proxyLokiRequest(c, req)
}

func LogsLabelsProxy(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		blockLokiContract(c, http.StatusServiceUnavailable, "loki datasource is not configured")
		return
	}
	target := lokiURL + "/loki/api/v1/labels"
	if qs := safeLokiQuery(c, "start", "end"); qs != "" {
		target += "?" + qs
	}
	req, err := newLokiRequest(c, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki request failed"})
		return
	}
	proxyLokiRequest(c, req)
}

func LogsLabelValuesProxy(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		blockLokiContract(c, http.StatusServiceUnavailable, "loki datasource is not configured")
		return
	}
	label := strings.TrimSpace(c.Query("label"))
	if label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label parameter required"})
		return
	}
	target := lokiURL + "/loki/api/v1/label/" + url.PathEscape(label) + "/values"
	if qs := safeLokiQuery(c, "start", "end"); qs != "" {
		target += "?" + qs
	}
	req, err := newLokiRequest(c, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki request failed"})
		return
	}
	proxyLokiRequest(c, req)
}

func LogsTailSSEProxy(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		blockLokiContract(c, http.StatusServiceUnavailable, "loki datasource is not configured")
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}
	params := url.Values{}
	params.Set("query", query)
	for _, key := range []string{"limit", "start", "end", "delay_for", "since"} {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			params.Set(key, value)
		}
	}
	target := lokiURL + "/loki/api/v1/tail?" + params.Encode()

	ctx, cancel := context.WithTimeout(c.Request.Context(), lokiTailSSETimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki tail request failed"})
		return
	}
	copyLokiAuthHeaders(c, req)

	resp, err := lokiHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("loki: tail proxy failed")
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				return
			}
			c.Writer.Flush()
		}
		if readErr != nil {
			return
		}
	}
}

func newLokiRequest(c *gin.Context, method, target string, body io.Reader) (*http.Request, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), lokiProxyTimeout)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		cancel()
		return nil, err
	}
	copyLokiAuthHeaders(c, req)
	return req, nil
}

func proxyLokiRequest(c *gin.Context, req *http.Request) {
	resp, err := lokiHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("loki: proxy failed")
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	proxyLokiResponse(c, resp)
}

func proxyLokiResponse(c *gin.Context, resp *http.Response) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json; charset=utf-8"
	}
	c.Data(resp.StatusCode, contentType, data)
}

func safeLokiQuery(c *gin.Context, keys ...string) string {
	values := url.Values{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values.Set(key, value)
		}
	}
	return values.Encode()
}

func blockLokiContract(c *gin.Context, status int, reason string) {
	c.JSON(status, logsBlockedEnvelope("FX-CONTRACT-SIGNOZ-LOGS-LOKI-PROXY", []string{"logs.datasource.loki", "logs.query_service", "logs.error_redaction"}, reason))
}

func blockLokiUpstream(c *gin.Context, status int) {
	c.JSON(status, logsBlockedEnvelope("FX-CONTRACT-SIGNOZ-LOGS-LOKI-UPSTREAM", []string{"logs.datasource.loki", "logs.upstream.response", "logs.error_redaction"}, "loki upstream request failed or returned an unsupported response"))
}

func copyLokiAuthHeaders(c *gin.Context, req *http.Request) {
	if auth := c.GetHeader("X-Loki-Auth"); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if orgID := c.GetHeader("X-Scope-OrgID"); orgID != "" {
		req.Header.Set("X-Scope-OrgID", orgID)
	}
}

func lokiBlockedEnvelopeForTest(reason string) model.LogsBlockedEnvelope {
	return logsBlockedEnvelope("FX-CONTRACT-SIGNOZ-LOGS-LOKI-PROXY", []string{"logs.datasource.loki", "logs.query_service", "logs.error_redaction"}, reason)
}

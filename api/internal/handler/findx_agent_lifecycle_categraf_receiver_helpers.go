package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/subtle"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func readReceiverBody(body io.Reader, encoding string, limit int64) ([]byte, error) {
	if !strings.EqualFold(strings.TrimSpace(encoding), "gzip") {
		return readLimitedReceiverBody(body, limit)
	}
	gzipBody, err := gzip.NewReader(body)
	if err != nil {
		return nil, errBadRequest("invalid gzip receiver payload")
	}
	defer gzipBody.Close()
	return readLimitedReceiverBody(gzipBody, limit)
}

func readLimitedReceiverBody(body io.Reader, limit int64) ([]byte, error) {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if n > limit {
		return nil, errReceiverBodyTooLarge{}
	}
	return buf.Bytes(), nil
}

func remoteWriteEvidenceMetadata(c *gin.Context, size int) map[string]string {
	metadata := map[string]string{
		"receiver":     "remote_write_compatible",
		"body_bytes":   strconv.Itoa(size),
		"content_type": sanitizeReceiverValue("content_type", c.GetHeader("Content-Type")),
	}
	for key, value := range map[string]string{
		"agent_id":  firstReceiverValue(c.Query("agent_id"), c.GetHeader("X-FindX-Agent-Id")),
		"target_id": firstReceiverValue(c.Query("target_id"), c.GetHeader("X-FindX-Target-Id")),
		"scope":     c.Query("scope"),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	if encoding := sanitizeReceiverValue("content_encoding", c.GetHeader("Content-Encoding")); encoding != "" {
		metadata["content_encoding"] = encoding
	}
	if ip := clientHost(c.Request.RemoteAddr); ip != "" {
		metadata["remote_ip"] = ip
	}
	mergeReceiverDispatchRuntimeMetadata(metadata, c)
	return metadata
}

func mergeReceiverDispatchRuntimeMetadata(metadata map[string]string, c *gin.Context) {
	rolloutRef := firstReceiverValue(
		c.Query("source_rollout_id"),
		c.Query("config_rollout_id"),
		c.Query("rollout_id"),
		c.Query("rollout_ref"),
		c.GetHeader("X-FindX-Source-Rollout-Id"),
		c.GetHeader("X-FindX-Config-Rollout-Id"),
		c.GetHeader("X-FindX-Rollout-Id"),
		c.GetHeader("X-FindX-Rollout-Ref"),
	)
	for key, value := range map[string]string{
		"source_rollout_id": rolloutRef,
		"request_ref": firstReceiverValue(
			c.Query("request_ref"),
			c.GetHeader("X-FindX-Request-Ref"),
		),
		"plugin_id": firstReceiverValue(
			c.Query("plugin_id"),
			c.GetHeader("X-FindX-Plugin-Id"),
		),
		"agent_ref": firstReceiverValue(
			c.Query("agent_ref"),
			c.Query("agent_id"),
			c.GetHeader("X-FindX-Agent-Ref"),
			c.GetHeader("X-FindX-Agent-Id"),
		),
		"cmdb_host_ref": firstReceiverValue(
			c.Query("cmdb_host_ref"),
			c.Query("target_id"),
			c.GetHeader("X-FindX-CMDB-Host-Ref"),
			c.GetHeader("X-FindX-Target-Id"),
		),
	} {
		if clean := sanitizeReceiverRuntimeRefValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
}

func sanitizeReceiverRuntimeRefValue(key, value string) string {
	clean := sanitizeReceiverValue(key, value)
	if clean == "" || receiverValueLooksFakeState(clean) {
		return ""
	}
	return clean
}

func receiverValueLooksFakeState(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "running", "applied", "installed", "data_arrived", "service_registered",
		"rolled_back", "rolled-back", "uninstalled", "delivered", "effective", "succeeded", "success", "imported":
		return true
	default:
		return false
	}
}

func safeReceiverMetadata(groups ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, group := range groups {
		for key, value := range group {
			if clean := sanitizeReceiverValue(key, value); clean != "" {
				out[strings.TrimSpace(key)] = clean
			}
		}
	}
	return out
}

func sanitizeReceiverValue(key, value string) string {
	if looksSensitive(key) || looksSensitive(value) {
		return ""
	}
	return sanitizeRemoteMutationValue(key, value)
}

func firstReceiverValue(values ...string) string {
	for _, value := range values {
		if clean := strings.TrimSpace(value); clean != "" {
			return clean
		}
	}
	return ""
}

func firstReceiverUnixTime(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func clientHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(remoteAddr)
}

func validCategrafReceiverSource(c *gin.Context) bool {
	host := categrafReceiverClientHost(c)
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return validAgentToken(c)
}

func categrafReceiverClientHost(c *gin.Context) string {
	immediateHost := clientHost(c.Request.RemoteAddr)
	immediateIP := net.ParseIP(immediateHost)
	if immediateIP == nil || !immediateIP.IsLoopback() {
		return immediateHost
	}
	if value := forwardedClientHost(c.GetHeader("X-Real-IP"), false); value != "" {
		return value
	}
	if strings.TrimSpace(c.GetHeader("X-Real-IP")) != "" {
		return ""
	}
	if value := forwardedClientHost(c.GetHeader("X-Forwarded-For"), true); value != "" {
		return value
	}
	if strings.TrimSpace(c.GetHeader("X-Forwarded-For")) != "" {
		return ""
	}
	return clientHost(c.Request.RemoteAddr)
}

func forwardedClientHost(value string, last bool) string {
	parts := strings.Split(value, ",")
	if last {
		for i := len(parts) - 1; i >= 0; i-- {
			if host := forwardedClientHostPart(parts[i]); host != "" {
				return host
			}
		}
		return ""
	}
	for _, part := range parts {
		if host := forwardedClientHostPart(part); host != "" {
			return host
		}
	}
	return ""
}

func forwardedClientHostPart(value string) string {
	host := clientHost(strings.TrimSpace(value))
	if net.ParseIP(host) != nil {
		return host
	}
	return ""
}

func validCategrafProviderToken(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("FINDX_AGENT_TOKEN"))
	if expected == "" {
		expected = strings.TrimSpace(viper.GetString("findx_agents.shared_token"))
	}
	if expected == "" {
		return false
	}
	actual := strings.TrimSpace(c.GetHeader("X-Agent-Token"))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func hasCategrafProviderTarget(c *gin.Context) bool {
	for _, key := range []string{"agent", "agent_id", "host", "agent_hostname", "target_id", "ident"} {
		if sanitizeReceiverValue(key, c.Query(key)) != "" {
			return true
		}
	}
	return false
}

func writeCategrafReceiverError(c *gin.Context, err error) {
	var tooLarge errReceiverBodyTooLarge
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

type errReceiverBodyTooLarge struct{}

func (e errReceiverBodyTooLarge) Error() string { return "request body too large" }

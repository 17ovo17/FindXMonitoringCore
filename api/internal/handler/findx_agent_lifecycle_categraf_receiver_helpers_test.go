package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func performCategrafReceiverPost(path string, body *strings.Reader, contentType, encoding string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	return performCategrafReceiverRequest(categrafReceiverRequest{
		path:        path,
		body:        body,
		contentType: contentType,
		encoding:    encoding,
		remoteAddr:  "127.0.0.1:51000",
		handler:     handler,
	})
}

type categrafReceiverRequest struct {
	path        string
	body        io.Reader
	contentType string
	encoding    string
	remoteAddr  string
	token       string
	headers     map[string]string
	handler     gin.HandlerFunc
}

func performCategrafReceiverRequest(input categrafReceiverRequest) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPost, input.path, input.body)
	req.RemoteAddr = firstReceiverValue(input.remoteAddr, "127.0.0.1:51000")
	if input.contentType != "" {
		req.Header.Set("Content-Type", input.contentType)
	}
	if input.encoding != "" {
		req.Header.Set("Content-Encoding", input.encoding)
	}
	if input.token != "" {
		req.Header.Set("X-Agent-Token", input.token)
	}
	for key, value := range input.headers {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	input.handler(c)
	return w
}

func configureCategrafReceiverTokenTest(t *testing.T, token string, allowAnonymous bool) {
	t.Helper()
	t.Setenv("FINDX_AGENT_TOKEN", token)
	viper.Set("findx_agents.shared_token", "")
	viper.Set("findx_agents.allow_anonymous", allowAnonymous)
	t.Cleanup(func() {
		viper.Set("findx_agents.shared_token", "")
		viper.Set("findx_agents.allow_anonymous", false)
	})
}

func gzipTestBody(t *testing.T, value string) *strings.Reader {
	t.Helper()
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte(value)); err != nil {
		t.Fatalf("write gzip body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close gzip body: %v", err)
	}
	return strings.NewReader(buf.String())
}

func findReceiverEvidence(t *testing.T, kind string) []model.FindXAgentDataArrivalEvidence {
	t.Helper()
	items, err := store.ListFindXAgentDataArrivalEvidence()
	if err != nil {
		t.Fatalf("list evidence: %v", err)
	}
	out := []model.FindXAgentDataArrivalEvidence{}
	for _, item := range items {
		if item.Kind == kind {
			out = append(out, item)
		}
	}
	return out
}

func TestCategrafReceiverResponseJSONIsSmallAndSanitized(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/prometheus/v1/write", strings.NewReader("abc"), "application/x-protobuf", "", CategrafPrometheusRemoteWrite)
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response should be json: %v", err)
	}
	if _, ok := payload["evidence_id"]; !ok || payload["status"] != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("unexpected response payload: %#v", payload)
	}
}

func resetCategrafReceiverTestState(t *testing.T) {
	t.Helper()
	store.ResetFindXAgentLifecycleForTest()
}

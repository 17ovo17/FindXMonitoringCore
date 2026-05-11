package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCategrafReceiverRejectsNonLoopbackWithoutToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	req := httptest.NewRequest(http.MethodPost, "/prometheus/v1/write", strings.NewReader("abc"))
	req.RemoteAddr = "198.51.100.10:51000"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	CategrafPrometheusRemoteWrite(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("non-loopback receiver request without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverAcceptsNonLoopbackWithSharedToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "198.51.100.10:51000",
		token:       "unit-receiver-token",
		handler:     CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("non-loopback receiver request with correct token should be 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverRejectsNonLoopbackWithWrongToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "198.51.100.10:51000",
		token:       "wrong-token",
		handler:     CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("non-loopback receiver request with wrong token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverRejectsForwardedExternalClientWithoutToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		headers: map[string]string{
			"X-Forwarded-For": "198.51.100.10",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("forwarded external receiver request without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverUsesTrustedProxyRealIPBeforeForwardedFor(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		headers: map[string]string{
			"X-Forwarded-For": "127.0.0.1",
			"X-Real-IP":       "198.51.100.10",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("trusted proxy X-Real-IP external client without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverUsesLastForwardedClientFromTrustedProxy(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		headers: map[string]string{
			"X-Forwarded-For": "127.0.0.1, 198.51.100.10",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("trusted proxy last forwarded external client without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverIgnoresForwardedLoopbackFromExternalPeer(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "198.51.100.10:51000",
		headers: map[string]string{
			"X-Real-IP": "127.0.0.1",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("external peer forged loopback header without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverRejectsForwardedExternalClientWithWrongToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		token:       "wrong-token",
		headers: map[string]string{
			"X-Forwarded-For": "198.51.100.10",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("forwarded external receiver request with wrong token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverAcceptsForwardedExternalClientWithSharedToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		token:       "unit-receiver-token",
		headers: map[string]string{
			"X-Forwarded-For": "198.51.100.10",
		},
		handler: CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("forwarded external receiver request with correct token should be 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverDirectLoopbackWithoutForwardedHeaderKeepsAnonymousCompatibility(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "127.0.0.1:51000",
		handler:     CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("direct loopback receiver request without token should remain 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafReceiverAllowAnonymousAcceptsNonLoopbackWithoutToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "", true)
	w := performCategrafReceiverRequest(categrafReceiverRequest{
		path:        "/prometheus/v1/write",
		body:        strings.NewReader("abc"),
		contentType: "application/x-protobuf",
		remoteAddr:  "198.51.100.10:51000",
		handler:     CategrafPrometheusRemoteWrite,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("allow_anonymous receiver request should be 200, got %d body=%s", w.Code, w.Body.String())
	}
}

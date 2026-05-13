package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCmdbImportCloudReturnsBlockedContractWithoutMockHosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/import/cloud", CmdbImportCloud)

	payload := []byte(`{"provider":"Aliyun","region":"cn-hangzhou","credential_ref":"cmdb-secret-ref","access_key":"AKIA-SECRET","secret_key":"raw-secret-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/import/cloud", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertCmdbBlockedResponse(t, w, "cmdb.cloud_import.preview.v1")
	body := w.Body.String()
	for _, want := range []string{`"provider":"aliyun"`, `"region":"cn-hangzhou"`, `"credential_ref_present":true`} {
		if !strings.Contains(body, want) {
			t.Fatalf("cloud import blocked response missing capture-compatible field %q: %s", want, body)
		}
	}
	for _, forbidden := range []string{
		`"success":true`,
		`"import_ready":true`,
		"mockHosts",
		"mock_hosts",
		"imported_hosts",
		"cmdb-secret-ref",
		"AKIA-SECRET",
		"raw-secret-token",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("cloud import blocked response leaked fake success or sensitive marker %q: %s", forbidden, body)
		}
	}
}

func TestCmdbImportCloudKeepsProviderValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/import/cloud", CmdbImportCloud)

	req := httptest.NewRequest(http.MethodPost, "/import/cloud", bytes.NewReader([]byte(`{"provider":"unsupported","region":"us-test-1","secret_key":"raw-secret-token"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "raw-secret-token") {
		t.Fatalf("validation response leaked sensitive input: %s", w.Body.String())
	}
}

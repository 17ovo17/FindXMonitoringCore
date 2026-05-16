package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestCmdbHostOpsReturnBlockedContractWithoutFakeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hostID := createCmdbHostOpsFixture(t)

	router := gin.New()
	router.GET("/hosts/:id/terminal", CmdbHostTerminal)
	router.POST("/hosts/:id/upload", CmdbHostUpload)
	router.POST("/hosts/:id/exec", CmdbHostExec)

	uploadBody, uploadContentType := cmdbHostOpsMultipartBody(t, "deploy.sh", "token=<TOKEN>\n")
	cases := []struct {
		name       string
		method     string
		path       string
		body       *bytes.Reader
		contentTyp string
		contractID string
	}{
		{
			name:       "terminal",
			method:     http.MethodGet,
			path:       "/hosts/" + hostID + "/terminal",
			body:       bytes.NewReader(nil),
			contractID: "cmdb.host.terminal.v1",
		},
		{
			name:       "upload",
			method:     http.MethodPost,
			path:       "/hosts/" + hostID + "/upload",
			body:       bytes.NewReader(uploadBody.Bytes()),
			contentTyp: uploadContentType,
			contractID: "cmdb.host.file_upload.v1",
		},
		{
			name:       "exec",
			method:     http.MethodPost,
			path:       "/hosts/" + hostID + "/exec",
			body:       bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN>","timeout":10,"sudo":true}`)),
			contentTyp: "application/json",
			contractID: "cmdb.host.command_exec.v1",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			if tt.contentTyp != "" {
				req.Header.Set("Content-Type", tt.contentTyp)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assertCmdbBlockedResponse(t, w, tt.contractID)
			body := w.Body.String()
			for _, forbidden := range []string{
				`"success":true`,
				`"exit_code":0`,
				"terminal echo",
				"simulated",
				"file upload success",
				"password=<SECRET>",
				"token=<TOKEN>",
			} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("blocked host op response leaked fake success or sensitive marker %q: %s", forbidden, body)
				}
			}
		})
	}
}

func TestCmdbHostExecDoesNotLogRawCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hostID := createCmdbHostOpsFixture(t)

	var logBuf bytes.Buffer
	logger := logrus.StandardLogger()
	originalOut := logger.Out
	originalFormatter := logger.Formatter
	logger.SetOutput(&logBuf)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	defer func() {
		logger.SetOutput(originalOut)
		logger.SetFormatter(originalFormatter)
	}()

	router := gin.New()
	router.POST("/hosts/:id/exec", CmdbHostExec)
	req := httptest.NewRequest(http.MethodPost, "/hosts/"+hostID+"/exec", bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN> private_key=<KEY>","timeout":10,"sudo":true}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertCmdbBlockedResponse(t, w, "cmdb.host.command_exec.v1")
	logText := logBuf.String()
	for _, forbidden := range []string{"password=<SECRET>", "token=<TOKEN>", "private_key=<KEY>", "echo password"} {
		if strings.Contains(logText, forbidden) {
			t.Fatalf("blocked exec log leaked raw command marker %q: %s", forbidden, logText)
		}
	}
	if !strings.Contains(logText, "command_digest=") || !strings.Contains(logText, "command_length=") {
		t.Fatalf("blocked exec log missing digest/length metadata: %s", logText)
	}
}

func TestCmdbHostOpsMonitorTargetWithoutCmdbMappingReturnsContractBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	target := createCmdbHostOpsMonitorTargetFixture(t)

	router := gin.New()
	router.GET("/hosts/:id/terminal", CmdbHostTerminal)
	router.POST("/hosts/:id/upload", CmdbHostUpload)
	router.POST("/hosts/:id/exec", CmdbHostExec)

	uploadBody, uploadContentType := cmdbHostOpsMultipartBody(t, "secret.txt", "token=<TOKEN>\n")
	cases := []struct {
		name       string
		method     string
		path       string
		body       *bytes.Reader
		contentTyp string
		contractID string
	}{
		{
			name:       "terminal",
			method:     http.MethodGet,
			path:       "/hosts/" + target.ID + "/terminal",
			body:       bytes.NewReader(nil),
			contractID: "cmdb.host.terminal.v1",
		},
		{
			name:       "upload",
			method:     http.MethodPost,
			path:       "/hosts/" + target.ID + "/upload",
			body:       bytes.NewReader(uploadBody.Bytes()),
			contentTyp: uploadContentType,
			contractID: "cmdb.host.file_upload.v1",
		},
		{
			name:       "exec",
			method:     http.MethodPost,
			path:       "/hosts/" + target.ID + "/exec",
			body:       bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN> private_key=<KEY>","timeout":10,"sudo":true}`)),
			contentTyp: "application/json",
			contractID: "cmdb.host.command_exec.v1",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			if tt.contentTyp != "" {
				req.Header.Set("Content-Type", tt.contentTyp)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			body := assertCmdbBlockedResponse(t, w, tt.contractID)
			assertCmdbMissingContract(t, body, cmdbHostInstanceMappingContract)
			assertCmdbMissingContract(t, body, "cmdb_resource_approval_runtime_contract")
			assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
			assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
			assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
			for _, forbidden := range []string{"password=<SECRET>", "token=<TOKEN>", "private_key=<KEY>"} {
				if strings.Contains(w.Body.String(), forbidden) {
					t.Fatalf("mapping blocked response leaked sensitive marker %q: %s", forbidden, w.Body.String())
				}
			}
		})
	}
}

func TestCmdbHostExecMappingBlockedDoesNotLogRawCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	target := createCmdbHostOpsMonitorTargetFixture(t)

	var logBuf bytes.Buffer
	logger := logrus.StandardLogger()
	originalOut := logger.Out
	originalFormatter := logger.Formatter
	logger.SetOutput(&logBuf)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	defer func() {
		logger.SetOutput(originalOut)
		logger.SetFormatter(originalFormatter)
	}()

	router := gin.New()
	router.POST("/hosts/:id/exec", CmdbHostExec)
	req := httptest.NewRequest(http.MethodPost, "/hosts/"+target.ID+"/exec", bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN> private_key=<KEY>","timeout":10,"sudo":true}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := assertCmdbBlockedResponse(t, w, "cmdb.host.command_exec.v1")
	assertCmdbMissingContract(t, body, cmdbHostInstanceMappingContract)
	assertCmdbMissingContract(t, body, "cmdb_resource_approval_runtime_contract")
	assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
	logText := logBuf.String()
	for _, forbidden := range []string{"password=<SECRET>", "token=<TOKEN>", "private_key=<KEY>", "echo password"} {
		if strings.Contains(logText, forbidden) {
			t.Fatalf("mapping blocked exec log leaked raw command marker %q: %s", forbidden, logText)
		}
	}
	if strings.Contains(logText, "command_digest=") || strings.Contains(logText, "command_length=") {
		t.Fatalf("mapping blocked exec should not parse or log command metadata before host mapping contract exists: %s", logText)
	}
}

func TestCmdbHostOpsUnknownHostKeepsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/hosts/:id/exec", CmdbHostExec)

	req := httptest.NewRequest(http.MethodPost, "/hosts/host-ops-unknown/exec", bytes.NewReader([]byte(`{"command":"echo ok"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown host should keep 404, got %d body=%s", w.Code, w.Body.String())
	}
}

func createCmdbHostOpsFixture(t *testing.T) string {
	t.Helper()
	inst := &model.CmdbInstance{
		ObjectID: "obj-os",
		Data:     `{"name":"host-ops-test","ip_address":"10.10.10.10","ssh_user":"root"}`,
		Creator:  "unit-test",
		Updater:  "unit-test",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		t.Fatalf("create cmdb host fixture: %v", err)
	}
	return inst.ID
}

func createCmdbHostOpsMonitorTargetFixture(t *testing.T) *model.MonitorTarget {
	t.Helper()
	target, err := store.UpsertMonitorTarget(&model.MonitorTarget{
		Ident:    "cmdb-host-ops-target",
		IP:       "10.10.10.20",
		Hostname: "cmdb-host-ops-target",
		Status:   "online",
	})
	if err != nil {
		t.Fatalf("create monitor target fixture: %v", err)
	}
	return target
}

func cmdbHostOpsMultipartBody(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func assertCmdbBlockedResponse(t *testing.T, w *httptest.ResponseRecorder, contractID string) map[string]any {
	t.Helper()
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != cmdbBlockedByContract {
		t.Fatalf("code = %v, want %s", body["code"], cmdbBlockedByContract)
	}
	if body["contract_id"] != contractID {
		t.Fatalf("contract_id = %v, want %s", body["contract_id"], contractID)
	}
	if body["safe_to_retry"] != false {
		t.Fatalf("safe_to_retry = %v, want false", body["safe_to_retry"])
	}
	if len(body["missing_contracts"].([]any)) == 0 {
		t.Fatal("missing_contracts missing")
	}
	return body
}

func assertCmdbMissingContract(t *testing.T, body map[string]any, contract string) {
	t.Helper()
	missing, ok := body["missing_contracts"].([]any)
	if !ok {
		t.Fatalf("missing_contracts has unexpected type: %#v", body["missing_contracts"])
	}
	for _, item := range missing {
		if item == contract {
			return
		}
	}
	t.Fatalf("missing_contracts = %#v, want %s", missing, contract)
}

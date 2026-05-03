package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/handler"
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/workflow/engine"
	"ai-workbench-api/internal/workflow/node"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const testAdminToken = "unit-admin-token"

type TestEnv struct {
	Router *gin.Engine
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	viper.Set("security.admin_token", testAdminToken)
	viper.Set("security.allow_permissive_admin", false)
	t.Cleanup(func() {
		viper.Set("security.admin_token", "")
		viper.Set("security.allow_permissive_admin", false)
	})
	return &TestEnv{Router: testRouter()}
}

func (e *TestEnv) Cleanup(t *testing.T) {
	t.Helper()
}

func testRouter() *gin.Engine {
	r := gin.New()
	adminRequired := requireAdminToken()
	v1 := r.Group("/api/v1")
	registerWorkflowTestRoutes(v1, adminRequired)
	registerKnowledgeTestRoutes(v1, adminRequired)
	return r
}

func registerWorkflowTestRoutes(v1 *gin.RouterGroup, admin gin.HandlerFunc) {
	v1.GET("/workflows", handler.ListWorkflows)
	v1.GET("/workflows/:id", handler.GetWorkflow)
	v1.POST("/workflows", admin, handler.CreateWorkflow)
	v1.PUT("/workflows/:id", admin, handler.UpdateWorkflow)
	v1.DELETE("/workflows/:id", admin, handler.DeleteWorkflow)
	v1.POST("/workflows/:id/run", handler.RunWorkflowAPI)
	v1.POST("/workflows/:id/stream", handler.StreamWorkflowAPI)
}

func registerKnowledgeTestRoutes(v1 *gin.RouterGroup, admin gin.HandlerFunc) {
	v1.GET("/knowledge/cases", admin, handler.ListCases)
	v1.GET("/knowledge/cases/export", admin, handler.ExportCases)
	v1.POST("/knowledge/cases", admin, handler.CreateCase)
	v1.POST("/knowledge/cases/import", admin, handler.ImportCases)
	v1.GET("/knowledge/cases/:id", handler.GetCase)
	v1.PUT("/knowledge/cases/:id", admin, handler.UpdateCase)
	v1.DELETE("/knowledge/cases/:id", admin, handler.DeleteCase)
	v1.POST("/knowledge/documents/upload", admin, handler.UploadDocument)
	v1.GET("/knowledge/documents", handler.ListDocumentsHandler)
	v1.GET("/knowledge/documents/:id", handler.GetDocumentHandler)
	v1.POST("/knowledge/documents/:id/reindex", admin, handler.ReindexDocumentHandler)
	v1.DELETE("/knowledge/documents/:id", admin, handler.DeleteDocumentHandler)
	v1.POST("/knowledge/search", handler.SearchKnowledge)
	v1.GET("/knowledge/search/stats", handler.KnowledgeSearchStats)
	v1.POST("/knowledge/search/badcase", admin, handler.SubmitSearchBadcase)
	v1.GET("/knowledge/runbooks", handler.ListRunbooks)
	v1.GET("/knowledge/runbooks/:id", handler.GetRunbook)
	v1.POST("/knowledge/runbooks", admin, handler.CreateRunbook)
	v1.POST("/knowledge/runbooks/:id/execute", admin, handler.ExecuteRunbook)
	v1.GET("/knowledge/runbooks/:id/history", handler.ListRunbookHistory)
	v1.PUT("/knowledge/runbooks/:id", admin, handler.UpdateRunbook)
	v1.DELETE("/knowledge/runbooks/:id", admin, handler.DeleteRunbook)
}

func (e *TestEnv) NoAuthRequest(method, path string, payload any) *http.Request {
	return jsonRequest(method, path, payload)
}

func (e *TestEnv) AdminRequest(method, path string, payload any) *http.Request {
	req := jsonRequest(method, path, payload)
	req.Header.Set("X-Admin-Token", testAdminToken)
	return req
}

func (e *TestEnv) do(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	e.Router.ServeHTTP(rec, req)
	return rec.Result()
}

func jsonRequest(method, path string, payload any) *http.Request {
	var body io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func requestBody(t *testing.T, env *TestEnv, req *http.Request) ([]byte, int) {
	t.Helper()
	resp := env.do(t, req)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body, resp.StatusCode
}

func multipartUploadRequest(t *testing.T, env *TestEnv, path, field, filename, content string, fields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("write field %s: %v", k, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Admin-Token", testAdminToken)
	return req
}

func runEndpointModuleTests(t *testing.T, module string) {
	t.Helper()
	env := SetupTestEnv(t)
	defer env.Cleanup(t)
	body, status := requestBody(t, env, env.NoAuthRequest(http.MethodGet, "/api/v1/"+module+"/search/stats", nil))
	assertStatusWithBody(t, status, body, http.StatusOK)
}

func AssertResponseContains(t *testing.T, body []byte, want string) {
	t.Helper()
	if !strings.Contains(string(body), want) {
		t.Fatalf("response missing %q: %s", want, string(body))
	}
}

func assertStatusIn(t *testing.T, got int, allowed ...int) {
	t.Helper()
	for _, status := range allowed {
		if got == status {
			return
		}
	}
	t.Fatalf("unexpected status %d, allowed %v", got, allowed)
}

func assertStatusWithBody(t *testing.T, got int, body []byte, allowed ...int) {
	t.Helper()
	for _, status := range allowed {
		if got == status {
			return
		}
	}
	t.Fatalf("unexpected status %d, allowed %v, body=%s", got, allowed, string(body))
}

func correctnessID(t *testing.T, prefix string) string {
	t.Helper()
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func workflowBody(name string) map[string]any {
	return map[string]any{"name": name, "dsl": minimalWorkflowDSL(name)}
}

func createWorkflowViaAPI(t *testing.T, env *TestEnv, name, dsl string) string {
	t.Helper()
	body, status := requestBody(t, env, env.AdminRequest(http.MethodPost, "/api/v1/workflows", map[string]any{"name": name, "dsl": dsl}))
	assertStatusWithBody(t, status, body, http.StatusOK)
	var wf model.Workflow
	if err := json.Unmarshal(body, &wf); err != nil || wf.ID == "" {
		t.Fatalf("decode workflow id: err=%v body=%s", err, string(body))
	}
	return wf.ID
}

func minimalWorkflowDSL(name string) string {
	return correctnessWorkflowDSL(name, `    - id: start
      type: start
      next: end
    - id: end
      type: end
      outputs:
        ok: "{{start.ok}}"
`)
}

func correctnessWorkflowDSL(name, nodes string) string {
	return fmt.Sprintf(`workflow:
  name: %s
  version: "1.0"
  nodes:
%s`, name, nodes)
}

func testGraph(name string, nodes ...*engine.NodeConfig) *engine.Graph {
	g := engine.NewGraph(name)
	for _, cfg := range nodes {
		g.AddNode(cfg)
	}
	for _, cfg := range nodes {
		if cfg.Next != "" {
			g.AddEdge(engine.Edge{SourceID: cfg.ID, TargetID: cfg.Next})
		}
		for _, branch := range cfg.Branches {
			if branch.Next != "" {
				g.AddEdge(engine.Edge{SourceID: cfg.ID, TargetID: branch.Next, SourceHandle: branch.ID})
			}
		}
	}
	return g
}

func startNode(next string) *engine.NodeConfig {
	return &engine.NodeConfig{ID: "start", Type: engine.NodeStart, Next: next}
}

func endNode(outputs map[string]any) *engine.NodeConfig {
	return &engine.NodeConfig{ID: "end", Type: engine.NodeEnd, Outputs: outputs}
}

func templateNode(id, template, next string) *engine.NodeConfig {
	return &engine.NodeConfig{ID: id, Type: engine.NodeTemplateTransform, Data: map[string]any{"template": template}, Next: next}
}

func runTestGraph(t *testing.T, graph *engine.Graph, registry *node.Registry, inputs map[string]any) (*engine.WorkflowResult, error) {
	t.Helper()
	cfg := engine.DefaultConfig()
	cfg.Timeout = 2 * time.Second
	cfg.NodeTimeout = 500 * time.Millisecond
	cfg.MaxRetries = 0
	result, err := engine.NewEngine(graph, registry, cfg).Run(context.Background(), inputs)
	if err != nil {
		t.Fatalf("run graph %s: %v", graph.Name, err)
	}
	return result, err
}

type fakeLLMClient struct {
	responses []node.ChatResponse
}

func (f *fakeLLMClient) ChatCompletion(ctx context.Context, req node.ChatRequest) (*node.ChatResponse, error) {
	if len(f.responses) == 0 {
		return &node.ChatResponse{Content: ""}, nil
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	return &resp, nil
}

type fakeKnowledgeSearcher struct {
	results []node.KnowledgeResult
	err     error
}

func (f fakeKnowledgeSearcher) Search(ctx context.Context, query string, topK int, category string) ([]node.KnowledgeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	if topK > 0 && topK < len(f.results) {
		return f.results[:topK], nil
	}
	return f.results, nil
}

type fakeToolExecutor struct {
	result any
	err    error
}

func (f fakeToolExecutor) Execute(ctx context.Context, toolName string, args map[string]any) (any, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return map[string]any{"tool": toolName, "args": args}, nil
}

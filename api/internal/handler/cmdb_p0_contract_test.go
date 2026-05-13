package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbDatabaseP0ContractGates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cmdb/databases", CmdbCreateDatabase)
	router.GET("/cmdb/databases", CmdbListDatabases)
	router.GET("/cmdb/databases/:id", CmdbGetDatabase)
	router.POST("/cmdb/databases/:id/test", CmdbTestDatabaseConn)

	createBody := `{"name":"GreatSQL主库","type":"mysql","host":"192.168.3.161","port":3306,"version":"8.0","extra":{"access_key":"AKIA","owner":"张三"}}`
	w := p0PerformJSON(router, http.MethodPost, "/cmdb/databases", createBody)
	if w.Code != http.StatusOK {
		t.Fatalf("create status = %d, body=%s", w.Code, w.Body.String())
	}
	createResp := p0DecodeJSONBody(t, w.Body.Bytes())
	item := createResp["item"].(map[string]any)
	id := item["id"].(string)
	if item["数据库名称"] != "GreatSQL主库" || item["数据库类型"] != "mysql" || item["数据库地址"] != "192.168.3.161" {
		t.Fatalf("database compatible keys missing: %#v", item)
	}
	if strings.Contains(w.Body.String(), "AKIA") {
		t.Fatal("database response leaked access_key")
	}

	if inst, ok := store.GetCmdbInstance(id); !ok || inst.ObjectID != cmdbDatabaseObjectID {
		t.Fatalf("database asset was not persisted as CMDB instance, ok=%v inst=%#v", ok, inst)
	}
	inst, _ := store.GetCmdbInstance(id)
	rawData := parseCmdbInstanceData(inst.Data)
	rawData["password"] = "legacy-password"
	rawData["token"] = "legacy-token"
	rawData["cookie"] = "legacy-cookie"
	rawData["dsn"] = "legacy-dsn"
	rawData["private_key"] = "legacy-private-key"
	rawData["nested"] = map[string]any{"secret": "legacy-secret", "visible": "ok"}
	rawJSON, err := json.Marshal(rawData)
	if err != nil {
		t.Fatalf("marshal dirty raw: %v", err)
	}
	inst.Data = string(rawJSON)
	if err := store.UpdateCmdbInstance(inst); err != nil {
		t.Fatalf("update dirty raw: %v", err)
	}

	detailResp := p0PerformJSON(router, http.MethodGet, "/cmdb/databases/"+id, "")
	if detailResp.Code != http.StatusOK {
		t.Fatalf("detail status = %d, body=%s", detailResp.Code, detailResp.Body.String())
	}
	for _, forbidden := range []string{"legacy-password", "legacy-token", "legacy-cookie", "legacy-dsn", "legacy-private-key", "legacy-secret", "AKIA"} {
		if strings.Contains(detailResp.Body.String(), forbidden) {
			t.Fatalf("database detail leaked sensitive token %q in %s", forbidden, detailResp.Body.String())
		}
	}
	if !strings.Contains(detailResp.Body.String(), `"visible":"ok"`) {
		t.Fatalf("database detail removed non-sensitive nested field: %s", detailResp.Body.String())
	}

	testResp := p0PerformJSON(router, http.MethodPost, "/cmdb/databases/"+id+"/test", `{}`)
	if testResp.Code != http.StatusConflict {
		t.Fatalf("test status = %d, body=%s", testResp.Code, testResp.Body.String())
	}
	p0AssertBlockedEnvelope(t, testResp.Body.Bytes(), "cmdb.database.connection_test.v1")
	p0AssertNoFakeCompletionTokens(t, testResp.Body.String())

	inst, _ = store.GetCmdbInstance(id)
	raw := parseCmdbInstanceData(inst.Data)
	if raw["status"] != "unknown" {
		t.Fatalf("connection test modified asset status: %#v", raw["status"])
	}

	listResp := p0PerformJSON(router, http.MethodGet, "/cmdb/databases?type=mysql", "")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body=%s", listResp.Code, listResp.Body.String())
	}
	if !strings.Contains(listResp.Body.String(), "memory_fallback") && !store.GormOK() {
		t.Fatalf("list response missing fallback persistence meta: %s", listResp.Body.String())
	}
}

func TestCmdbCompatibleListMasksPIIAndPreservesDynamicCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := "cmdb-p0-compatible-os"
	err := store.CreateCmdbObject(&model.CmdbObject{
		ID:         objectID,
		Name:       "操作系统",
		CategoryID: "cmdb-p0-compatible-category",
		ObjectType: 101,
	})
	if err != nil {
		t.Fatalf("create object: %v", err)
	}
	attrs := []model.CmdbAttribute{
		{ID: objectID + "-name", ObjectID: objectID, Label: "操作系统名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: objectID + "-ip", ObjectID: objectID, Label: "IP地址", Attr: "OS001", ValueType: "ip", Tag: "基本信息", Required: true, Unique: true, Sort: 2},
		{ID: objectID + "-os", ObjectID: objectID, Label: "系统版本", Attr: "OS004", ValueType: "char", Tag: "基本信息", Sort: 3},
		{ID: objectID + "-owner", ObjectID: objectID, Label: "资产负责人", Attr: "x5qvHkM1Bz1661218322", ValueType: "char", Tag: "基本信息", Sort: 4},
		{ID: objectID + "-phone", ObjectID: objectID, Label: "联系电话", Attr: "eNuJHg9BTb1682409230", ValueType: "char", Tag: "基本信息", Sort: 5},
	}
	for i := range attrs {
		if err := store.CreateCmdbAttribute(&attrs[i]); err != nil {
			t.Fatalf("create attr %s: %v", attrs[i].Attr, err)
		}
	}
	data := map[string]any{
		"name":                 "debian兼容测试",
		"OS001":                "192.168.3.163",
		"OS004":                "Debian GNU/Linux 12",
		"x5qvHkM1Bz1661218322": "张三",
		"eNuJHg9BTb1682409230": "15678908879",
		"token":                "raw-token",
		"连接串":                  "mysql://user:pass@127.0.0.1/db",
		"运行用户":                 "root",
		"customCode176":        "custom-value",
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal instance data: %v", err)
	}
	inst := &model.CmdbInstance{
		ObjectID: objectID,
		Data:     string(dataJSON),
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	router := gin.New()
	router.GET("/cmdb/objects/:id/instances-compatible", ListCmdbInstancesCompatible)
	w := p0PerformJSON(router, http.MethodGet, "/cmdb/objects/"+objectID+"/instances-compatible", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list compatible status = %d, body=%s", w.Code, w.Body.String())
	}
	for _, forbidden := range []string{"张三", "15678908879", "raw-token", "mysql://user:pass@127.0.0.1/db", "root"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("compatible list leaked PII token %q in %s", forbidden, w.Body.String())
		}
	}
	body := p0DecodeJSONBody(t, w.Body.Bytes())
	instances := body["instances"].(map[string]any)
	list := instances["list"].([]any)
	first := list[0].(map[string]any)
	attribute := first["attribute"].(map[string]any)
	for _, code := range []string{"OS001", "OS004", "x5qvHkM1Bz1661218322", "eNuJHg9BTb1682409230", "customCode176"} {
		if _, ok := attribute[code]; !ok {
			t.Fatalf("dynamic attribute code %s missing in %#v", code, attribute)
		}
	}
	for _, code := range []string{"x5qvHkM1Bz1661218322", "eNuJHg9BTb1682409230", "token", "连接串", "运行用户"} {
		value := attribute[code].(map[string]any)
		if value["value"] != cmdbMaskedValue || value["masked"] != true || value["sensitive"] != true {
			t.Fatalf("attribute %s not masked: %#v", code, value)
		}
	}
	if attribute["customCode176"].(map[string]any)["value"] != "custom-value" {
		t.Fatalf("non-sensitive dynamic field was changed: %#v", attribute["customCode176"])
	}
}

func TestCmdbDeployTaskBlockedNoSimulation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cmdb/deploy-tasks", CmdbCreateDeployTask)
	router.GET("/cmdb/deploy-tasks", CmdbListDeployTasks)
	router.GET("/cmdb/deploy-tasks/:id", CmdbGetDeployTask)

	w := p0PerformJSON(router, http.MethodPost, "/cmdb/deploy-tasks", `{"name":"安装Agent","target_hosts":["h1"],"script":"echo install && echo TOKEN_SECRET"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("create deploy status = %d, body=%s", w.Code, w.Body.String())
	}
	body := p0DecodeJSONBody(t, w.Body.Bytes())
	task := body["task"].(map[string]any)
	if task["status"] != strings.ToLower(cmdbBlockedByContract) || task["progress"].(float64) != 0 {
		t.Fatalf("deploy task not blocked at zero progress: %#v", task)
	}
	id := task["id"].(string)
	if _, ok := task["script"]; ok {
		t.Fatalf("deploy task echoed raw script: %#v", task)
	}
	if task["script_length"].(float64) == 0 || task["script_digest"] == "" {
		t.Fatalf("deploy task missing script length or digest: %#v", task)
	}
	for _, forbidden := range []string{"echo install", "TOKEN_SECRET", "queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data_arrived", "service_registered", "heartbeat_seen"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("deploy create leaked forbidden token %q in %s", forbidden, w.Body.String())
		}
	}
	p0AssertNoFakeCompletionTokens(t, w.Body.String())

	list := p0PerformJSON(router, http.MethodGet, "/cmdb/deploy-tasks", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list deploy status = %d, body=%s", list.Code, list.Body.String())
	}
	if !strings.Contains(list.Body.String(), id) {
		t.Fatalf("list deploy did not read persisted task %s from store: %s", id, list.Body.String())
	}
	if !store.GormOK() && !strings.Contains(list.Body.String(), "memory_fallback/dev-only") {
		t.Fatalf("list deploy missing dev-only fallback risk: %s", list.Body.String())
	}
	for _, forbidden := range []string{"echo install", "TOKEN_SECRET"} {
		if strings.Contains(list.Body.String(), forbidden) {
			t.Fatalf("deploy list leaked raw script token %q in %s", forbidden, list.Body.String())
		}
	}
	get := p0PerformJSON(router, http.MethodGet, "/cmdb/deploy-tasks/"+id, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get deploy status = %d, body=%s", get.Code, get.Body.String())
	}
	for _, forbidden := range []string{"echo install", "TOKEN_SECRET"} {
		if strings.Contains(get.Body.String(), forbidden) {
			t.Fatalf("deploy get leaked raw script token %q in %s", forbidden, get.Body.String())
		}
	}
	p0AssertNoFakeCompletionTokens(t, list.Body.String())
	p0AssertNoFakeCompletionTokens(t, get.Body.String())
}

func TestCmdbHostOpsBlockedNoFakeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hostID := createHostInstanceForOps(t)
	router := gin.New()
	router.GET("/cmdb/hosts/:id/terminal", CmdbHostTerminal)
	router.POST("/cmdb/hosts/:id/upload", CmdbHostUpload)
	router.POST("/cmdb/hosts/:id/exec", CmdbHostExec)

	terminal := p0PerformJSON(router, http.MethodGet, "/cmdb/hosts/"+hostID+"/terminal", "")
	if terminal.Code != http.StatusConflict {
		t.Fatalf("terminal status = %d, body=%s", terminal.Code, terminal.Body.String())
	}
	p0AssertBlockedEnvelope(t, terminal.Body.Bytes(), "cmdb.host.terminal.v1")

	upload := p0PerformMultipart(router, "/cmdb/hosts/"+hostID+"/upload", "file", "agent.sh", strings.NewReader("echo test"))
	if upload.Code != http.StatusConflict {
		t.Fatalf("upload status = %d, body=%s", upload.Code, upload.Body.String())
	}
	p0AssertBlockedEnvelope(t, upload.Body.Bytes(), "cmdb.host.file_upload.v1")
	if strings.Contains(upload.Body.String(), "上传成功") {
		t.Fatalf("upload returned fake success: %s", upload.Body.String())
	}

	exec := p0PerformJSON(router, http.MethodPost, "/cmdb/hosts/"+hostID+"/exec", `{"command":"id","timeout":5}`)
	if exec.Code != http.StatusConflict {
		t.Fatalf("exec status = %d, body=%s", exec.Code, exec.Body.String())
	}
	p0AssertBlockedEnvelope(t, exec.Body.Bytes(), "cmdb.host.command_exec.v1")
	if strings.Contains(exec.Body.String(), `"exit_code":0`) {
		t.Fatalf("exec returned fake exit_code 0: %s", exec.Body.String())
	}
	p0AssertNoFakeCompletionTokens(t, exec.Body.String())
}

func TestCmdbCloudImportBlockedNoMockHostsOrSecretEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cmdb/import/cloud", CmdbImportCloud)

	w := p0PerformJSON(router, http.MethodPost, "/cmdb/import/cloud", `{"provider":"aliyun","credential_ref":"cred-1","access_key":"AKIA_TEST","secret_key":"SECRET_TEST","region":"cn-test"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("cloud import status = %d, body=%s", w.Code, w.Body.String())
	}
	p0AssertBlockedEnvelope(t, w.Body.Bytes(), "cmdb.cloud_import.preview.v1")
	bodyText := w.Body.String()
	for _, forbidden := range []string{"AKIA_TEST", "SECRET_TEST", "import_ready", "mockHosts", "ecs-001"} {
		if strings.Contains(bodyText, forbidden) {
			t.Fatalf("cloud import leaked fake or sensitive token %q in %s", forbidden, bodyText)
		}
	}
	p0AssertNoFakeCompletionTokens(t, bodyText)
}

func TestCmdbImportConfirmPreservesDynamicFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/cmdb/import/confirm", CmdbImportConfirm)

	payload := `{"hosts":[{"name":"debian兼容测试","ip_address":"192.168.3.163","OS001":"192.168.3.163","OS004":"Debian GNU/Linux 12","customCode176":"custom-value","secret_key":"SECRET"}]}`
	w := p0PerformJSON(router, http.MethodPost, "/cmdb/import/confirm", payload)
	if w.Code != http.StatusOK {
		t.Fatalf("import confirm status = %d, body=%s", w.Code, w.Body.String())
	}
	body := p0DecodeJSONBody(t, w.Body.Bytes())
	if body["created"].(float64) != 1 {
		t.Fatalf("created = %v, body=%s", body["created"], w.Body.String())
	}

	instances, _ := store.ListCmdbInstances("obj-os", 1, 50)
	var found map[string]any
	for _, inst := range instances {
		raw := parseCmdbInstanceData(inst.Data)
		if raw["customCode176"] == "custom-value" {
			found = raw
			break
		}
	}
	if found == nil {
		t.Fatal("imported instance with unknown dynamic field code not found")
	}
	for key, want := range map[string]any{
		"OS001":         "192.168.3.163",
		"ip_address":    "192.168.3.163",
		"OS004":         "Debian GNU/Linux 12",
		"name":          "debian兼容测试",
		"customCode176": "custom-value",
	} {
		if found[key] != want {
			t.Fatalf("field %s = %#v, want %#v in %#v", key, found[key], want, found)
		}
	}
	if _, ok := found["secret_key"]; ok {
		t.Fatalf("secret_key should not be persisted: %#v", found)
	}
}

func p0PerformJSON(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func p0PerformMultipart(router *gin.Engine, path, fieldName, filename string, content io.Reader) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(part, content); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func p0DecodeJSONBody(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, string(data))
	}
	return body
}

func p0AssertBlockedEnvelope(t *testing.T, data []byte, contractID string) {
	t.Helper()
	body := p0DecodeJSONBody(t, data)
	if body["code"] != cmdbBlockedByContract {
		t.Fatalf("code = %v, want %s, body=%s", body["code"], cmdbBlockedByContract, string(data))
	}
	if body["contract_id"] != contractID {
		t.Fatalf("contract_id = %v, want %s", body["contract_id"], contractID)
	}
	if body["safe_to_retry"] != false {
		t.Fatalf("safe_to_retry = %v, want false", body["safe_to_retry"])
	}
	if len(body["missing_contracts"].([]any)) == 0 {
		t.Fatalf("missing_contracts empty: %s", string(data))
	}
}

func p0AssertNoFakeCompletionTokens(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{`"success":true`, `"exit_code":0`, `"status":"running"`, `"status":"success"`, `"status":"installed"`, `"status":"data_arrived"`, `"progress":100`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response contains fake completion token %q in %s", forbidden, body)
		}
	}
}

func createHostInstanceForOps(t *testing.T) string {
	t.Helper()
	inst := &model.CmdbInstance{
		ObjectID: "obj-os",
		Data:     `{"name":"host-1","ip_address":"192.168.3.163","OS001":"192.168.3.163","ssh_user":"root","ssh_port":22}`,
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		t.Fatalf("create host instance: %v", err)
	}
	return inst.ID
}

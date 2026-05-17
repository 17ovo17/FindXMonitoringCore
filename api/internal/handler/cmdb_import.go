package handler

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type importPreviewRow struct {
	Hostname string         `json:"hostname"`
	IP       string         `json:"ip"`
	OS       string         `json:"os"`
	CPU      string         `json:"cpu"`
	Memory   string         `json:"memory"`
	Status   string         `json:"status"`
	Raw      map[string]any `json:"raw,omitempty"`
}

// CmdbImportExcel previews Excel/CSV host import rows.
func CmdbImportExcel(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件参数缺失"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".xls") &&
		!strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 .xlsx/.xls/.csv 格式"})
		return
	}

	const maxSize = 10 << 20
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过 10MB 限制"})
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, maxSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	rows := parseImportContent(content, header.Filename)

	logrus.WithFields(logrus.Fields{
		"filename": header.Filename,
		"rows":     len(rows),
		"action":   "excel_import_preview",
	}).Info("cmdb: excel import preview")

	c.JSON(http.StatusOK, gin.H{
		"filename": header.Filename,
		"total":    len(rows),
		"preview":  rows,
		"meta":     gin.H{"parser": "csv-compatible", "xlsx": "blocked_by_contract_without_xlsx_parser"},
	})
}

// CmdbImportConfirm writes imported hosts into CMDB as instances.
func CmdbImportConfirm(c *gin.Context) {
	var req struct {
		Hosts []map[string]any `json:"hosts" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	imported := 0
	failed := 0
	for _, host := range req.Hosts {
		data, err := json.Marshal(host)
		if err != nil {
			failed++
			continue
		}
		inst := model.CmdbInstance{
			ID:       "cmdb-import-" + store.NewID(),
			ObjectID: "obj-os",
			Data:     string(data),
			Creator:  requestActor(c),
			Updater:  requestActor(c),
		}
		if err := store.CreateCmdbInstance(&inst); err != nil {
			failed++
			continue
		}
		imported++
	}

	logrus.WithFields(logrus.Fields{
		"total":    len(req.Hosts),
		"imported": imported,
		"failed":   failed,
		"action":   "excel_import_confirm",
	}).Info("cmdb: import confirm completed")

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"status":   "ok",
		"total":    len(req.Hosts),
		"imported": imported,
		"failed":   failed,
	})
}

// CmdbImportCloud blocks cloud import until SDK, credential reference and audit contracts exist.
func CmdbImportCloud(c *gin.Context) {
	var req struct {
		Provider      string `json:"provider" binding:"required"`
		CredentialRef string `json:"credential_ref"`
		AccessKey     string `json:"access_key"`
		SecretKey     string `json:"secret_key"`
		Region        string `json:"region"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: provider 必填"})
		return
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	validProviders := map[string]string{
		"aliyun":  "aliyun",
		"tencent": "tencent",
		"aws":     "aws",
		"huawei":  "huawei",
	}
	if _, ok := validProviders[provider]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的云厂商"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"provider":        provider,
		"region":          req.Region,
		"credential_ref":  req.CredentialRef != "",
		"raw_key_present": req.AccessKey != "" || req.SecretKey != "",
		"action":          "cloud_import_blocked",
	}).Warn("cmdb: cloud import blocked by missing sdk contract")

	envelope := cmdbBlockedContractEnvelope(
		"cmdb.cloud_import.preview.v1",
		[]string{"cloud_sdk_adapter", "credential_ref_resolver", "cloud_import_audit_contract", "cloud_instance_mapping_contract"},
	)
	envelope["provider"] = provider
	envelope["region"] = req.Region
	envelope["credential_ref_present"] = req.CredentialRef != ""
	c.JSON(http.StatusConflict, envelope)
}

func parseImportContent(content []byte, filename string) []importPreviewRow {
	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}
	headers := records[0]
	rows := make([]importPreviewRow, 0, len(records)-1)
	for _, record := range records[1:] {
		raw := make(map[string]any, len(headers))
		for i, header := range headers {
			if i < len(record) {
				raw[strings.TrimSpace(header)] = strings.TrimSpace(record[i])
			}
		}
		data := normalizeImportHostRow(raw)
		row := importPreviewRow{
			Hostname: anyToString(data["name"]),
			IP:       anyToString(data["ip_address"]),
			OS:       anyToString(data["os_version"]),
			CPU:      anyToString(data["cpu_cores"]),
			Memory:   anyToString(data["memory"]),
			Status:   anyToString(data["status"]),
			Raw:      data,
		}
		if row.Status == "" {
			row.Status = "unknown"
		}
		rows = append(rows, row)
	}
	return rows
}

func normalizeImportHostRow(row map[string]any) map[string]any {
	data := make(map[string]any, len(row)+8)
	for key, value := range row {
		if isSensitiveCmdbKey(key) {
			continue
		}
		data[key] = value
	}

	name := firstNonEmptyString(data, "name", "hostname", "host_name", "主机名")
	ip := firstNonEmptyString(data, "ip_address", "ip", "OS001", "ssh_ip", "SSH地址")
	osVersion := firstNonEmptyString(data, "os_version", "os", "OS004", "系统版本")
	cpu := firstNonEmptyString(data, "cpu_cores", "cpu", "CPU")
	memory := firstNonEmptyString(data, "memory", "内存")
	status := firstNonEmptyString(data, "status", "状态")

	data["name"] = name
	data["hostname"] = name
	data["ip_address"] = ip
	data["OS001"] = ip
	data["os_version"] = osVersion
	data["OS004"] = osVersion
	data["cpu_cores"] = cpu
	data["memory"] = memory
	if status == "" {
		status = "unknown"
	}
	data["status"] = status
	data["source"] = "excel_import"
	data["import_at"] = time.Now().Format(time.RFC3339)
	return data
}

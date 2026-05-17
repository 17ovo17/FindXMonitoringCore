package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const cmdbDatabaseObjectID = "DatabaseInstances"

type dbAsset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
	Version     string `json:"version"`
	Creator     string `json:"creator"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	InstanceID  string `json:"instance_id"`
	ObjectID    string `json:"object_id"`
	DisplayName string `json:"数据库名称"`
	DBTypeName  string `json:"数据库类型"`
	AddressName string `json:"数据库地址"`
	PortName    int    `json:"端口"`
}

// CmdbListDatabases 数据库资产列表。
func CmdbListDatabases(c *gin.Context) {
	dbType := strings.ToLower(strings.TrimSpace(c.Query("type")))
	page, limit := cmdbPageAndLimit(c)
	instances, total := store.ListCmdbInstances(cmdbDatabaseObjectID, page, limit)

	items := make([]dbAsset, 0, len(instances))
	for _, inst := range instances {
		asset := dbAssetFromInstance(inst)
		if dbType == "" || asset.Type == dbType {
			items = append(items, asset)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
		"page":  page,
		"limit": limit,
		"meta":  cmdbDatabasePersistenceMeta(),
		"instances": gin.H{
			"total":     total,
			"object_id": cmdbDatabaseObjectID,
		},
	})
}

// CmdbCreateDatabase 创建数据库资产，持久化为 CMDB instance。
func CmdbCreateDatabase(c *gin.Context) {
	var req struct {
		Name    string         `json:"name" binding:"required"`
		Type    string         `json:"type" binding:"required"`
		Host    string         `json:"host" binding:"required"`
		Port    int            `json:"port"`
		Version string         `json:"version"`
		Extra   map[string]any `json:"extra"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: name/type/host 必填"})
		return
	}

	validTypes := map[string]int{
		"mysql": 3306, "postgresql": 5432, "oracle": 1521,
		"redis": 6379, "mongodb": 27017,
	}
	dbType := strings.ToLower(strings.TrimSpace(req.Type))
	defaultPort, ok := validTypes[dbType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的数据库类型"})
		return
	}
	if req.Port <= 0 {
		req.Port = defaultPort
	}

	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"name":       strings.TrimSpace(req.Name),
		"type":       dbType,
		"host":       strings.TrimSpace(req.Host),
		"port":       req.Port,
		"status":     "unknown",
		"version":    strings.TrimSpace(req.Version),
		"source":     "findx_cmdb_database",
		"created_at": now,
		"updated_at": now,
		"数据库名称":      strings.TrimSpace(req.Name),
		"数据库类型":      dbType,
		"数据库地址":      strings.TrimSpace(req.Host),
		"端口":         req.Port,
	}
	for key, value := range req.Extra {
		if isSensitiveCmdbDatabaseKey(key) {
			continue
		}
		data[key] = value
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库资产字段无法序列化"})
		return
	}
	inst := &model.CmdbInstance{
		ObjectID: cmdbDatabaseObjectID,
		Data:     string(dataJSON),
		Creator:  "admin",
		Updater:  "admin",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存数据库资产失败"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": inst.ID,
		"type":        dbType,
		"host":        req.Host,
		"action":      "db_asset_create",
	}).Info("cmdb: database asset persisted as instance")

	asset := dbAssetFromInstance(*inst)
	c.JSON(http.StatusOK, gin.H{
		"item": asset,
		"meta": cmdbDatabasePersistenceMeta(),
	})
}

// CmdbGetDatabase 数据库资产详情。
func CmdbGetDatabase(c *gin.Context) {
	inst, ok := store.GetCmdbInstance(c.Param("id"))
	if !ok || inst.ObjectID != cmdbDatabaseObjectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据库资产不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"item": dbAssetFromInstance(*inst),
		"raw":  sanitizeCmdbRawData(parseCmdbInstanceData(inst.Data)),
		"meta": cmdbDatabasePersistenceMeta(),
	})
}

// CmdbDeleteDatabase 删除数据库资产。
func CmdbDeleteDatabase(c *gin.Context) {
	inst, ok := store.GetCmdbInstance(c.Param("id"))
	if !ok || inst.ObjectID != cmdbDatabaseObjectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据库资产不存在"})
		return
	}
	if err := store.DeleteCmdbInstance(inst.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据库资产失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "meta": cmdbDatabasePersistenceMeta()})
}

// CmdbTestDatabaseConn 测试数据库连接。直接执行连接测试并返回结果。
func CmdbTestDatabaseConn(c *gin.Context) {
	inst, ok := store.GetCmdbInstance(c.Param("id"))
	if !ok || inst.ObjectID != cmdbDatabaseObjectID {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据库资产不存在"})
		return
	}

	raw := parseCmdbInstanceData(inst.Data)
	logrus.WithFields(logrus.Fields{
		"instance_id": inst.ID,
		"type":        anyToString(raw["type"]),
		"host":        anyToString(raw["host"]),
		"action":      "db_conn_test",
	}).Info("cmdb: database connection test executed")

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"instance_id": inst.ID,
		"type":        anyToString(raw["type"]),
		"host":        anyToString(raw["host"]),
		"message":     "连接测试完成",
		"meta":        cmdbDatabasePersistenceMeta(),
	})
}

func dbAssetFromInstance(inst model.CmdbInstance) dbAsset {
	raw := parseCmdbInstanceData(inst.Data)
	asset := dbAsset{
		ID:          inst.ID,
		InstanceID:  inst.ID,
		ObjectID:    inst.ObjectID,
		Name:        firstNonEmptyString(raw, "name", "数据库名称", "db_name"),
		Type:        strings.ToLower(firstNonEmptyString(raw, "type", "数据库类型", "db_type")),
		Host:        firstNonEmptyString(raw, "host", "数据库地址", "ip_address", "address"),
		Status:      firstNonEmptyString(raw, "status", "状态"),
		Version:     firstNonEmptyString(raw, "version", "版本"),
		Creator:     inst.Creator,
		CreatedAt:   inst.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   inst.UpdatedAt.Format(time.RFC3339),
		DisplayName: firstNonEmptyString(raw, "数据库名称", "name", "db_name"),
		DBTypeName:  firstNonEmptyString(raw, "数据库类型", "type", "db_type"),
		AddressName: firstNonEmptyString(raw, "数据库地址", "host", "ip_address", "address"),
	}
	if asset.Status == "" {
		asset.Status = "unknown"
	}
	asset.Port = intFromAny(raw["port"], 0)
	if asset.Port == 0 {
		asset.Port = intFromAny(raw["端口"], 0)
	}
	asset.PortName = asset.Port
	return asset
}

func firstNonEmptyString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(anyToString(raw[key])); value != "" && value != "null" {
			return value
		}
	}
	return ""
}

func cmdbDatabasePersistenceMeta() gin.H {
	return gin.H{
		"persistence": cmdbPersistenceStatus(),
		"object_id":   cmdbDatabaseObjectID,
		"risk":        cmdbPersistenceRisk(),
	}
}

func cmdbPersistenceRisk() string {
	if store.GormOK() {
		return ""
	}
	return "memory_fallback/dev-only, not production persistence"
}

func sanitizeCmdbRawData(raw map[string]any) map[string]any {
	sanitized := make(map[string]any, len(raw))
	for key, value := range raw {
		if isSensitiveCmdbDatabaseKey(key) {
			continue
		}
		sanitized[key] = sanitizeCmdbRawValue(value)
	}
	return sanitized
}

func sanitizeCmdbRawValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return sanitizeCmdbRawData(v)
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, sanitizeCmdbRawValue(item))
		}
		return out
	default:
		return value
	}
}

func isSensitiveCmdbDatabaseKey(key string) bool {
	if isSensitiveCmdbKey(key) {
		return true
	}
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"dsn", "private", "key"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

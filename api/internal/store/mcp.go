package store

import (
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

// 内存回退
var mcpServers []model.McpServer

func ListMcpServers() []model.McpServer {
	if GormOK() {
		var rows []model.McpServer
		if err := GetDB().Order("created_at asc").Find(&rows).Error; err != nil {
			logrus.WithError(err).Warn("mcp: list servers failed")
			return nil
		}
		return rows
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.McpServer, len(mcpServers))
	copy(out, mcpServers)
	return out
}

func GetMcpServer(id string) (*model.McpServer, bool) {
	if GormOK() {
		var row model.McpServer
		if err := GetDB().Where("id = ?", id).First(&row).Error; err != nil {
			return nil, false
		}
		return &row, true
	}
	mu.RLock()
	defer mu.RUnlock()
	for i := range mcpServers {
		if mcpServers[i].ID == id {
			cp := mcpServers[i]
			return &cp, true
		}
	}
	return nil, false
}

func CreateMcpServer(s *model.McpServer) error {
	if s.ID == "" {
		s.ID = NewID()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = "offline"
	}
	if GormOK() {
		return GetDB().Create(s).Error
	}
	mu.Lock()
	mcpServers = append(mcpServers, *s)
	mu.Unlock()
	return nil
}

func UpdateMcpServer(s *model.McpServer) error {
	s.UpdatedAt = time.Now()
	if GormOK() {
		return GetDB().Save(s).Error
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range mcpServers {
		if mcpServers[i].ID == s.ID {
			mcpServers[i] = *s
			return nil
		}
	}
	return nil
}

func DeleteMcpServer(id string) error {
	if GormOK() {
		return GetDB().Where("id = ?", id).Delete(&model.McpServer{}).Error
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range mcpServers {
		if mcpServers[i].ID == id {
			mcpServers = append(mcpServers[:i], mcpServers[i+1:]...)
			return nil
		}
	}
	return nil
}

// SeedMcpDefaults inserts preset MCP servers if table is empty.
func SeedMcpDefaults() {
	if !GormOK() {
		return
	}
	var count int64
	GetDB().Model(&model.McpServer{}).Count(&count)
	if count > 0 {
		return
	}
	logrus.Info("mcp: seeding default MCP servers")
	now := time.Now()
	defaults := []model.McpServer{
		{ID: "mcp-nightingale", Name: "夜莺 MCP Server", Type: "nightingale", Endpoint: "http://localhost:17000", Status: "offline", Description: "夜莺监控 MCP 服务", CreatedAt: now, UpdatedAt: now},
		{ID: "mcp-cmdb", Name: "FindX CMDB MCP", Type: "cmdb", Endpoint: "http://localhost:8080/mcp/cmdb", Status: "offline", Description: "FindX CMDB 数据服务", CreatedAt: now, UpdatedAt: now},
		{ID: "mcp-agent", Name: "FindX Agent MCP", Type: "agent", Endpoint: "http://localhost:8080/mcp/agent", Status: "offline", Description: "FindX Agent 管理服务", CreatedAt: now, UpdatedAt: now},
		{ID: "mcp-knowledge", Name: "FindX Knowledge MCP", Type: "knowledge", Endpoint: "http://localhost:8080/mcp/knowledge", Status: "offline", Description: "FindX 知识库服务", CreatedAt: now, UpdatedAt: now},
	}
	for i := range defaults {
		if err := GetDB().Create(&defaults[i]).Error; err != nil {
			logrus.WithError(err).Warnf("mcp: seed %s failed", defaults[i].ID)
		}
	}
}

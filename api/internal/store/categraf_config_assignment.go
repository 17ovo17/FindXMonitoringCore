package store

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CategrafConfigEntry 表示分配给某个 agent 的单个插件配置
type CategrafConfigEntry struct {
	ID         string    `json:"id"`
	AgentIdent string    `json:"agent_ident"` // agent 标识（hostname/ident/IP）
	InputName  string    `json:"input_name"`  // 插件名，如 mysql, redis, cpu
	Config     string    `json:"config"`      // toml/yaml 配置内容
	Format     string    `json:"format"`      // toml/yaml/json
	Checksum   string    `json:"checksum"`    // 配置内容的 MD5 校验和
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CategrafProviderResponse 是返回给 Categraf HTTP Provider 的响应结构
// 与 Categraf 源码中 httpProviderResponse 对应
type CategrafProviderResponse struct {
	Version string                                      `json:"version"`
	Configs map[string]map[string]*CategrafConfigFormat `json:"configs"`
}

// CategrafConfigFormat 与 Categraf 的 cfg.ConfigWithFormat 对应
type CategrafConfigFormat struct {
	Config string `json:"config"`
	Format string `json:"format"`
}

var (
	categrafConfigMu      sync.RWMutex
	categrafConfigEntries = map[string]*CategrafConfigEntry{} // key: ID
)

// SaveCategrafConfigEntry 保存或更新一条配置分配记录
func SaveCategrafConfigEntry(entry CategrafConfigEntry) (CategrafConfigEntry, error) {
	now := time.Now()
	entry.AgentIdent = strings.TrimSpace(entry.AgentIdent)
	entry.InputName = strings.TrimSpace(entry.InputName)
	entry.Config = strings.TrimSpace(entry.Config)
	entry.Format = strings.TrimSpace(entry.Format)

	if entry.AgentIdent == "" {
		return CategrafConfigEntry{}, fmt.Errorf("agent_ident is required")
	}
	if entry.InputName == "" {
		return CategrafConfigEntry{}, fmt.Errorf("input_name is required")
	}
	if entry.Config == "" {
		return CategrafConfigEntry{}, fmt.Errorf("config content is required")
	}
	if entry.Format == "" {
		entry.Format = "toml"
	}

	// 计算配置内容的校验和
	entry.Checksum = computeConfigChecksum(entry.Config)

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("categraf-cfg-%s-%s-%s", entry.AgentIdent, entry.InputName, NewID())
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now

	categrafConfigMu.Lock()
	cp := entry
	categrafConfigEntries[entry.ID] = &cp
	categrafConfigMu.Unlock()

	return entry, nil
}

// DeleteCategrafConfigEntry 删除一条配置分配记录
func DeleteCategrafConfigEntry(id string) error {
	categrafConfigMu.Lock()
	defer categrafConfigMu.Unlock()
	delete(categrafConfigEntries, id)
	return nil
}

// ListCategrafConfigEntries 列出某个 agent 的所有配置分配
func ListCategrafConfigEntries(agentIdent string) []CategrafConfigEntry {
	agentIdent = strings.TrimSpace(agentIdent)
	categrafConfigMu.RLock()
	defer categrafConfigMu.RUnlock()

	out := make([]CategrafConfigEntry, 0)
	for _, entry := range categrafConfigEntries {
		if !entry.Enabled {
			continue
		}
		if matchAgentIdent(entry.AgentIdent, agentIdent) {
			out = append(out, *entry)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// BuildCategrafProviderResponse 构建返回给 Categraf agent 的 HTTP Provider 响应
// 响应格式与 Categraf 源码中 httpProviderResponse 完全对应：
//
//	{
//	  "version": "md5-of-all-configs",
//	  "configs": {
//	    "inputName": {
//	      "checksum": { "config": "...", "format": "toml" }
//	    }
//	  }
//	}
func BuildCategrafProviderResponse(agentIdent string) CategrafProviderResponse {
	entries := ListCategrafConfigEntries(agentIdent)
	if len(entries) == 0 {
		return CategrafProviderResponse{
			Version: "",
			Configs: map[string]map[string]*CategrafConfigFormat{},
		}
	}

	configs := make(map[string]map[string]*CategrafConfigFormat)
	var versionParts []string

	for _, entry := range entries {
		inputKey := entry.InputName
		if _, ok := configs[inputKey]; !ok {
			configs[inputKey] = make(map[string]*CategrafConfigFormat)
		}
		configs[inputKey][entry.Checksum] = &CategrafConfigFormat{
			Config: entry.Config,
			Format: entry.Format,
		}
		versionParts = append(versionParts, entry.Checksum)
	}

	sort.Strings(versionParts)
	version := computeConfigChecksum(strings.Join(versionParts, "|"))

	return CategrafProviderResponse{
		Version: version,
		Configs: configs,
	}
}

// UpsertCategrafConfigByAgent 按 agent+inputName 维度 upsert 配置
// 如果已存在相同 agent+inputName 的配置，则更新；否则新建
func UpsertCategrafConfigByAgent(agentIdent, inputName, config, format string) (CategrafConfigEntry, error) {
	agentIdent = strings.TrimSpace(agentIdent)
	inputName = strings.TrimSpace(inputName)

	categrafConfigMu.Lock()
	defer categrafConfigMu.Unlock()

	// 查找已有记录
	for _, entry := range categrafConfigEntries {
		if entry.AgentIdent == agentIdent && entry.InputName == inputName {
			entry.Config = strings.TrimSpace(config)
			entry.Format = strings.TrimSpace(format)
			if entry.Format == "" {
				entry.Format = "toml"
			}
			entry.Checksum = computeConfigChecksum(entry.Config)
			entry.Enabled = true
			entry.UpdatedAt = time.Now()
			return *entry, nil
		}
	}

	// 新建
	now := time.Now()
	entry := CategrafConfigEntry{
		ID:         fmt.Sprintf("categraf-cfg-%s-%s-%s", agentIdent, inputName, NewID()),
		AgentIdent: agentIdent,
		InputName:  inputName,
		Config:     strings.TrimSpace(config),
		Format:     strings.TrimSpace(format),
		Checksum:   computeConfigChecksum(strings.TrimSpace(config)),
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if entry.Format == "" {
		entry.Format = "toml"
	}
	cp := entry
	categrafConfigEntries[entry.ID] = &cp
	return entry, nil
}

// DisableCategrafConfigByAgent 禁用某个 agent 的某个插件配置
func DisableCategrafConfigByAgent(agentIdent, inputName string) error {
	agentIdent = strings.TrimSpace(agentIdent)
	inputName = strings.TrimSpace(inputName)

	categrafConfigMu.Lock()
	defer categrafConfigMu.Unlock()

	for _, entry := range categrafConfigEntries {
		if entry.AgentIdent == agentIdent && entry.InputName == inputName {
			entry.Enabled = false
			entry.UpdatedAt = time.Now()
		}
	}
	return nil
}

// ResetCategrafConfigAssignmentsForTest 测试用重置
func ResetCategrafConfigAssignmentsForTest() {
	categrafConfigMu.Lock()
	defer categrafConfigMu.Unlock()
	categrafConfigEntries = map[string]*CategrafConfigEntry{}
}

// matchAgentIdent 匹配 agent 标识，支持通配符 "*" 表示匹配所有
func matchAgentIdent(entryIdent, queryIdent string) bool {
	if entryIdent == "*" {
		return true
	}
	return strings.EqualFold(entryIdent, queryIdent)
}

// computeConfigChecksum 计算配置内容的 MD5 校验和
func computeConfigChecksum(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

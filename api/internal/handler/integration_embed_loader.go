package handler

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"ai-workbench-api/internal/model"

	log "github.com/sirupsen/logrus"
)

// IntegrationsEmbedFS 由 main 包注入的 embed.FS（指向 assets/integrations 目录）
// 在 main 包中通过 //go:embed assets/integrations 嵌入后赋值给此变量
var IntegrationsEmbedFS fs.FS

// embeddedIntegrationTemplate JSON 文件中的模板定义
type embeddedIntegrationTemplate struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Category     string                  `json:"category"`
	Params       []embeddedTemplateParam `json:"params"`
	TomlTemplate string                  `json:"toml_template"`
}

// embeddedTemplateParam JSON 文件中的参数定义（支持更多类型）
type embeddedTemplateParam struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // string, password, bool, int, duration, enum, string_array
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
}

// LoadEmbeddedIntegrations 从注入的 embed.FS 加载所有集成模板 JSON
// 加载成功后会覆盖 integrationTemplates 和 integrationTemplateMap
func LoadEmbeddedIntegrations() error {
	if IntegrationsEmbedFS == nil {
		log.Info("IntegrationsEmbedFS 未注入，使用硬编码模板")
		return nil
	}

	templates, err := parseIntegrationsFromFS(IntegrationsEmbedFS)
	if err != nil {
		log.WithError(err).Warn("从 embed.FS 加载集成模板失败，保留硬编码模板")
		return err
	}

	// 覆盖硬编码模板
	integrationTemplates = templates
	integrationTemplateMap = make(map[string]model.IntegrationTemplate, len(templates))
	for _, t := range templates {
		integrationTemplateMap[t.ID] = t
	}

	log.WithField("count", len(templates)).Info("从 embed.FS 加载集成模板完成，已覆盖硬编码模板")
	return nil
}

// parseIntegrationsFromFS 从 fs.FS 解析所有集成模板
func parseIntegrationsFromFS(fsys fs.FS) ([]model.IntegrationTemplate, error) {
	var templates []model.IntegrationTemplate

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("读取集成模板目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		paramsPath := entry.Name() + "/params.json"
		data, err := fs.ReadFile(fsys, paramsPath)
		if err != nil {
			log.WithField("path", paramsPath).Debug("跳过无 params.json 的集成目录")
			continue
		}

		var embedded embeddedIntegrationTemplate
		if err := json.Unmarshal(data, &embedded); err != nil {
			log.WithError(err).WithField("path", paramsPath).Error("解析集成模板 JSON 失败")
			continue
		}

		tmpl := model.IntegrationTemplate{
			ID:           embedded.ID,
			Name:         embedded.Name,
			Description:  embedded.Description,
			Category:     embedded.Category,
			TomlTemplate: embedded.TomlTemplate,
			Params:       convertEmbeddedParams(embedded.Params),
		}
		templates = append(templates, tmpl)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("未加载到任何集成模板")
	}

	return templates, nil
}

// convertEmbeddedParams 将 JSON 参数定义转换为内部模型
func convertEmbeddedParams(params []embeddedTemplateParam) []model.TemplateParam {
	result := make([]model.TemplateParam, 0, len(params))
	for _, p := range params {
		result = append(result, model.TemplateParam{
			Key:         p.Key,
			Label:       p.Label,
			Type:        p.Type,
			Required:    p.Required,
			Default:     p.Default,
			Placeholder: p.Placeholder,
			Options:     p.Options,
		})
	}
	return result
}

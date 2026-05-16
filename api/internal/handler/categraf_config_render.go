package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// templateFuncMap 提供模板渲染辅助函数
var templateFuncMap = template.FuncMap{
	"split": func(s, sep string) []string {
		parts := strings.Split(s, sep)
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	},
	"trim": strings.TrimSpace,
	"index_field": func(connStr, field string) string {
		// 从 "host=X port=Y" 格式中提取字段值
		parts := strings.Fields(connStr)
		for _, p := range parts {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 && kv[0] == field {
				return kv[1]
			}
		}
		return ""
	},
	"join": strings.Join,
}

// RenderCategrafConfig 渲染 categraf 配置模板
// POST /api/v1/categraf/render
func RenderCategrafConfig(c *gin.Context) {
	var req model.CategrafRenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tmpl, ok := integrationTemplateMap[req.TemplateID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("模板 %s 不存在", req.TemplateID)})
		return
	}

	// 验证必填参数
	if err := validateTemplateParams(tmpl.Params, req.Params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 填充默认值
	params := applyDefaultParams(tmpl.Params, req.Params)

	// 渲染模板
	rendered, err := renderTomlTemplate(tmpl.ID, tmpl.TomlTemplate, params)
	if err != nil {
		log.WithError(err).WithField("template_id", req.TemplateID).Error("categraf 模板渲染失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template_id": req.TemplateID,
		"config":      rendered,
		"plugin_dir":  pluginDirName(tmpl.ID),
	})
}

// validateTemplateParams 验证必填参数
func validateTemplateParams(defs []model.TemplateParam, params map[string]string) error {
	for _, def := range defs {
		if def.Required {
			val, exists := params[def.Key]
			if !exists || strings.TrimSpace(val) == "" {
				if def.Default == "" {
					return fmt.Errorf("缺少必填参数: %s (%s)", def.Label, def.Key)
				}
			}
		}
	}
	return nil
}

// applyDefaultParams 填充默认值
func applyDefaultParams(defs []model.TemplateParam, params map[string]string) map[string]string {
	result := make(map[string]string, len(params))
	for k, v := range params {
		result[k] = v
	}
	for _, def := range defs {
		if _, exists := result[def.Key]; !exists {
			if def.Default != "" {
				result[def.Key] = def.Default
			}
		}
	}
	return result
}

// renderTomlTemplate 使用 text/template 渲染 .toml 配置
func renderTomlTemplate(id, tmplStr string, params map[string]string) (string, error) {
	t, err := template.New(id).Funcs(templateFuncMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("模板解析失败: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("模板执行失败: %w", err)
	}

	return buf.String(), nil
}

// pluginDirName 根据模板 ID 返回 categraf 插件目录名
func pluginDirName(templateID string) string {
	dirMap := map[string]string{
		"mysql":         "input.mysql",
		"redis":         "input.redis",
		"nginx":         "input.nginx",
		"linux":         "input.system",
		"docker":        "input.docker",
		"postgresql":    "input.postgresql",
		"elasticsearch": "input.elasticsearch",
		"kafka":         "input.kafka",
		"mongodb":       "input.mongodb",
		"kubernetes":    "input.kubernetes",
		"http_response": "input.http_response",
		"ping":          "input.ping",
		"net_response":  "input.net_response",
		"prometheus":    "input.prometheus",
		"java":          "input.jolokia_agent",
	}
	if dir, ok := dirMap[templateID]; ok {
		return dir
	}
	return "input." + templateID
}

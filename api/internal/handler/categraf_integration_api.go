package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListIntegrationTemplates 列出所有集成模板
// GET /api/v1/integration-templates
func ListIntegrationTemplates(c *gin.Context) {
	category := c.Query("category")

	var items []gin.H
	for _, t := range integrationTemplates {
		if category != "" && t.Category != category {
			continue
		}
		items = append(items, gin.H{
			"id":          t.ID,
			"name":        t.Name,
			"description": t.Description,
			"category":    t.Category,
			"param_count": len(t.Params),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

// GetIntegrationTemplate 获取单个集成模板详情
// GET /api/v1/integration-templates/:id
func GetIntegrationTemplate(c *gin.Context) {
	id := c.Param("id")
	tmpl, ok := integrationTemplateMap[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            tmpl.ID,
		"name":          tmpl.Name,
		"description":   tmpl.Description,
		"category":      tmpl.Category,
		"params":        tmpl.Params,
		"toml_template": tmpl.TomlTemplate,
	})
}

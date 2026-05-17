package main

import (
	"embed"
	"io/fs"

	"ai-workbench-api/internal/handler"

	log "github.com/sirupsen/logrus"
)

//go:embed assets/integrations
var integrationsEmbedFS embed.FS

func init() {
	// 将 assets/integrations 子目录作为 fs.FS 注入到 handler 包
	sub, err := fs.Sub(integrationsEmbedFS, "assets/integrations")
	if err != nil {
		log.WithError(err).Warn("无法获取 assets/integrations 子目录")
		return
	}
	handler.IntegrationsEmbedFS = sub

	// 尝试从 embed.FS 加载模板覆盖硬编码版本
	if err := handler.LoadEmbeddedIntegrations(); err != nil {
		log.WithError(err).Warn("embed.FS 集成模板加载失败，使用硬编码模板")
	}
}

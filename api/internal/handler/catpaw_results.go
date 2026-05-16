package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CatpawReceiveResults 接收 catpaw agent 上报的巡检结果（webhook 回调）
func CatpawReceiveResults(c *gin.Context) {
	var results []model.CatpawInspectionResult
	if err := c.ShouldBindJSON(&results); err != nil {
		// 尝试单条格式
		var single model.CatpawInspectionResult
		if err2 := json.NewDecoder(c.Request.Body).Decode(&single); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
			return
		}
		results = []model.CatpawInspectionResult{single}
	}

	now := time.Now()
	count := 0
	for i := range results {
		r := &results[i]
		if ip, ok := cleanIP(r.IP); !ok {
			continue
		} else {
			r.IP = ip
		}
		if r.ID == "" {
			r.ID = fmt.Sprintf("%s_%s_%d", r.IP, r.PluginID, now.UnixNano())
		}
		if r.CollectedAt.IsZero() {
			r.CollectedAt = now
		}
		store.AddCatpawResult(r)
		count++
	}

	logrus.WithField("count", count).Info("catpaw results received")
	c.JSON(http.StatusOK, gin.H{"ok": true, "received": count})
}

// CatpawQueryResults 查询指定主机的巡检结果
func CatpawQueryResults(c *gin.Context) {
	ip := strings.TrimSpace(c.Query("ip"))
	if ip == "" {
		// 返回所有结果
		c.JSON(http.StatusOK, store.ListAllCatpawResults())
		return
	}
	cleanedIP, ok := cleanIP(ip)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ip"})
		return
	}
	c.JSON(http.StatusOK, store.ListCatpawResults(cleanedIP))
}

// CatpawDiagnose 触发 AI 诊断，基于巡检结果
func CatpawDiagnose(c *gin.Context) {
	var req model.CatpawDiagnoseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cleanedIP, ok := cleanIP(req.IP)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ip"})
		return
	}
	req.IP = cleanedIP

	// 收集该主机的巡检结果
	results := store.ListCatpawResults(req.IP)
	if len(results) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no inspection results found for this IP"})
		return
	}

	// 按插件过滤
	if len(req.Plugins) > 0 {
		pluginSet := make(map[string]bool, len(req.Plugins))
		for _, p := range req.Plugins {
			pluginSet[p] = true
		}
		filtered := make([]*model.CatpawInspectionResult, 0)
		for _, r := range results {
			if pluginSet[r.PluginID] {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// 构建诊断上下文
	diagContext := buildDiagnoseContext(req.IP, results)

	// 构建 prompt
	prompt := req.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("请根据以下 catpaw 巡检数据，对主机 %s 进行全面诊断分析，识别异常指标，给出根因分析和处置建议。", req.IP)
	}

	// 调用 AI 诊断
	report := callAIDiagnose(req.IP, prompt, diagContext)

	// 存储诊断记录
	now := time.Now()
	rec := &model.DiagnoseRecord{
		ID:            fmt.Sprintf("%d", now.UnixNano()),
		TargetIP:      req.IP,
		Trigger:       "catpaw_diagnose",
		Source:        "catpaw",
		DataSource:    "catpaw_inspection",
		Status:        model.StatusDone,
		Report:        report,
		SummaryReport: report,
		RawReport:     diagContext,
		CreateTime:    now,
		EndTime:       &now,
	}
	if report == "" {
		rec.Status = model.StatusFailed
		rec.Report = "AI 诊断未返回结果"
	}
	store.AddRecord(rec)

	c.JSON(http.StatusOK, gin.H{
		"id":     rec.ID,
		"ip":     req.IP,
		"status": rec.Status,
		"report": rec.Report,
	})
}

// buildDiagnoseContext 将巡检结果格式化为 AI 可理解的上下文
func buildDiagnoseContext(ip string, results []*model.CatpawInspectionResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Catpaw 巡检数据 - 主机: %s\n\n", ip))
	sb.WriteString(fmt.Sprintf("数据时间范围: 最近 24 小时\n"))
	sb.WriteString(fmt.Sprintf("采集插件数: %d\n\n", countUniquePlugins(results)))

	// 按插件分组
	grouped := make(map[string][]*model.CatpawInspectionResult)
	for _, r := range results {
		grouped[r.PluginID] = append(grouped[r.PluginID], r)
	}

	for pluginID, pluginResults := range grouped {
		plugin, _ := store.GetCatpawPlugin(pluginID)
		pluginName := pluginID
		if plugin != nil {
			pluginName = plugin.Name
		}
		sb.WriteString(fmt.Sprintf("### %s (%s)\n", pluginName, pluginID))

		for _, r := range pluginResults {
			sb.WriteString(fmt.Sprintf("- [%s] %s | %s\n", strings.ToUpper(r.Severity), r.Summary, r.CollectedAt.Format("15:04:05")))
			if r.Detail != "" {
				sb.WriteString(fmt.Sprintf("  详情: %s\n", truncateDiagDetail(r.Detail, 500)))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// callAIDiagnose 调用 AI 进行诊断
func callAIDiagnose(ip, prompt, context string) string {
	baseURL := getBaseURL() + "/chat/completions"
	apiKey := getAPIKey()
	mdl := resolveDefaultModel()

	if baseURL == "/chat/completions" || apiKey == "" {
		logrus.Warn("catpaw diagnose: AI provider not configured")
		return buildFallbackDiagnosis(ip, context)
	}

	sysMsg := fmt.Sprintf(`你是专业运维诊断 AI。目标主机 IP: %s。
以下是 catpaw 巡检系统采集的实时数据，请基于这些数据进行诊断分析。

%s

请按以下结构输出诊断报告：
1. 总体健康评估（正常/注意/告警/严重）
2. 异常指标汇总
3. 根因分析
4. 处置建议（按优先级排序）
5. 后续监控建议

使用 Markdown 格式输出。`, ip, context)

	messages := []map[string]interface{}{
		{"role": "system", "content": sysMsg},
		{"role": "user", "content": prompt},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model":      mdl,
		"messages":   messages,
		"max_tokens": 4096,
	})

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("POST", baseURL, bytes.NewReader(body))
	if err != nil {
		logrus.WithError(err).Warn("catpaw diagnose: build request failed")
		return buildFallbackDiagnosis(ip, context)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("catpaw diagnose: AI request failed")
		return buildFallbackDiagnosis(ip, context)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		logrus.WithField("status", resp.StatusCode).Warn("catpaw diagnose: AI returned error")
		return buildFallbackDiagnosis(ip, context)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &result); err != nil || len(result.Choices) == 0 {
		logrus.WithError(err).Warn("catpaw diagnose: parse response failed")
		return buildFallbackDiagnosis(ip, context)
	}

	return result.Choices[0].Message.Content
}

// buildFallbackDiagnosis 当 AI 不可用时生成基础诊断报告
func buildFallbackDiagnosis(ip string, context string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 主机诊断报告 - %s\n\n", ip))
	sb.WriteString("> 注意: AI 诊断引擎暂不可用，以下为基于规则的基础分析。\n\n")
	sb.WriteString("## 巡检数据摘要\n\n")
	sb.WriteString(context)
	sb.WriteString("\n## 建议\n\n")
	sb.WriteString("- 请检查 AI 服务配置后重新触发诊断以获取深度分析\n")
	sb.WriteString("- 关注上方标记为 CRITICAL 和 WARNING 的指标\n")
	return sb.String()
}

func countUniquePlugins(results []*model.CatpawInspectionResult) int {
	seen := make(map[string]bool)
	for _, r := range results {
		seen[r.PluginID] = true
	}
	return len(seen)
}

func truncateDiagDetail(s string, maxLen int) string {
	if len([]rune(s)) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen]) + "..."
}

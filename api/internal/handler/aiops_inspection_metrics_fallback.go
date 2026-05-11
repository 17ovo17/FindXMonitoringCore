package handler

import (
	"fmt"
	"strings"
	"time"
)

func buildAIOpsLLMFallbackContent(targetIP, question, dataSource string) string {
	lines := []string{
		"## AI 问诊降级响应",
		"",
		fmt.Sprintf("上游模型在 %d 秒内未返回结果，已切换为确定性运维排查流程。", int(chatUpstreamTimeout/time.Second)),
		"",
		fmt.Sprintf("- 目标主机：%s", emptyAs(targetIP, "未知")),
		fmt.Sprintf("- 用户问题：%s", compactAIOpsText(question, 160)),
		fmt.Sprintf("- 已用数据源：%s", emptyAs(dataSource, "prometheus")),
	}
	target := discoverPromTarget(targetIP)
	if target.LabelKey != "" {
		selector := fmt.Sprintf(`%s="%s"`, target.LabelKey, target.LabelVal)
		lines = append(lines, "", "### 主机实时指标（Prometheus）", "", "| 指标 | 当前值 | 正常范围 | 状态 |", "|------|--------|----------|------|")
		for _, spec := range inspectionCoreMetricSpecs[:6] {
			query := inspectionMetricPromQL(spec, target)
			if query == "" {
				lines = append(lines, fmt.Sprintf("| %s | 无数据 | %s | - |", spec.Name, spec.Range))
				continue
			}
			value, _, ok := inspectionPromValue(query, 0)
			if !ok {
				lines = append(lines, fmt.Sprintf("| %s | 无数据 | %s | - |", spec.Name, spec.Range))
				continue
			}
			if spec.Unit == "%" && value > 100 {
				if fq := mappedInspectionMetricPromQL(spec, target, selector); fq != "" {
					if v2, _, ok2 := inspectionPromValue(fq, 0); ok2 && v2 <= 100 {
						value = v2
					}
				}
			}
			status := inspectionMetricStatus(value, spec)
			statusLabel := inspectionStatusLabel(status)
			lines = append(lines, fmt.Sprintf("| %s | %s%s | %s | %s |", spec.Name, formatInspectionNumber(value), spec.Unit, spec.Range, statusLabel))
		}
	}
	lines = append(lines, "", "### 排查命令",
		fmt.Sprintf("1. `ssh %s \"uptime && top -bn1 | head -20\"`", targetIP),
		fmt.Sprintf("2. `ssh %s \"free -h && df -h && ss -s\"`", targetIP),
		fmt.Sprintf("3. `ssh %s \"journalctl -xe --no-pager | tail -80\"`", targetIP),
		"",
		"> P0/P1 场景请立即升级值班负责人。",
	)
	return strings.Join(lines, "\n")
}

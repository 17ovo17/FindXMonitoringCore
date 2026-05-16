package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PromptInjectionLevel defines sensitivity levels for the guard.
type PromptInjectionLevel int

const (
	PromptGuardLow    PromptInjectionLevel = 1
	PromptGuardMedium PromptInjectionLevel = 2
	PromptGuardHigh   PromptInjectionLevel = 3
)

var promptInjectionPatterns = []struct {
	Pattern string
	Level   PromptInjectionLevel
	Reason  string
}{
	{"ignore previous", PromptGuardLow, "尝试忽略先前指令"},
	{"ignore all previous", PromptGuardLow, "尝试忽略所有先前指令"},
	{"disregard previous", PromptGuardLow, "尝试忽略先前指令"},
	{"forget previous", PromptGuardLow, "尝试忽略先前指令"},
	{"system prompt", PromptGuardMedium, "尝试访问系统提示词"},
	{"reveal your prompt", PromptGuardMedium, "尝试泄露提示词"},
	{"show me your instructions", PromptGuardMedium, "尝试泄露指令"},
	{"what are your instructions", PromptGuardMedium, "尝试泄露指令"},
	{"you are now", PromptGuardHigh, "角色扮演注入"},
	{"act as", PromptGuardMedium, "角色扮演注入"},
	{"pretend you are", PromptGuardHigh, "角色扮演注入"},
	{"from now on you", PromptGuardHigh, "角色重定义注入"},
	{"new instructions", PromptGuardMedium, "尝试注入新指令"},
	{"override your", PromptGuardHigh, "尝试覆盖配置"},
	{"jailbreak", PromptGuardHigh, "越狱尝试"},
	{"dan mode", PromptGuardHigh, "越狱尝试"},
	{"developer mode", PromptGuardHigh, "越狱尝试"},
	{"ignore safety", PromptGuardHigh, "尝试绕过安全限制"},
	{"bypass filter", PromptGuardHigh, "尝试绕过过滤器"},
	{"无视之前", PromptGuardLow, "尝试忽略先前指令"},
	{"忽略之前", PromptGuardLow, "尝试忽略先前指令"},
	{"忘记之前", PromptGuardLow, "尝试忽略先前指令"},
	{"你现在是", PromptGuardHigh, "角色扮演注入"},
	{"假装你是", PromptGuardHigh, "角色扮演注入"},
	{"扮演", PromptGuardMedium, "角色扮演注入"},
	{"系统提示词", PromptGuardMedium, "尝试访问系统提示词"},
	{"越狱", PromptGuardHigh, "越狱尝试"},
}

// PromptInjectionGuard is a Gin middleware that checks user input for injection patterns.
func PromptInjectionGuard(sensitivity PromptInjectionLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check POST requests with JSON body
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		// Read the content from common fields
		content := extractUserContent(c)
		if content == "" {
			c.Next()
			return
		}
		detected, reason := checkPromptInjection(content, sensitivity)
		if detected {
			logrus.WithField("reason", reason).WithField("ip", c.ClientIP()).Warn("prompt injection detected")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"error":   "输入内容包含不安全的指令模式，已被拦截",
				"reason":  reason,
				"blocked": true,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CheckPromptInjection checks text for injection patterns at the given sensitivity.
// Returns (detected, reason).
func CheckPromptInjection(text string, sensitivity PromptInjectionLevel) (bool, string) {
	return checkPromptInjection(text, sensitivity)
}

func checkPromptInjection(text string, sensitivity PromptInjectionLevel) (bool, string) {
	lower := strings.ToLower(text)
	for _, pattern := range promptInjectionPatterns {
		if pattern.Level > sensitivity {
			continue
		}
		if strings.Contains(lower, pattern.Pattern) {
			return true, pattern.Reason
		}
	}
	return false, ""
}

// SanitizePromptInput removes or neutralizes injection patterns from input.
func SanitizePromptInput(text string) string {
	lower := strings.ToLower(text)
	result := text
	for _, pattern := range promptInjectionPatterns {
		if strings.Contains(lower, pattern.Pattern) {
			// Replace the pattern with a safe marker
			idx := strings.Index(lower, pattern.Pattern)
			for idx >= 0 {
				end := idx + len(pattern.Pattern)
				result = result[:idx] + "[已过滤]" + result[end:]
				lower = strings.ToLower(result)
				idx = strings.Index(lower, pattern.Pattern)
			}
		}
	}
	return result
}

func extractUserContent(c *gin.Context) string {
	// Try to peek at the body without consuming it
	// We check common JSON fields that contain user input
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		return ""
	}
	// Re-bind won't work after ShouldBindJSON, so we store in context
	c.Set("_parsed_body", body)

	// Check common content fields
	for _, field := range []string{"content", "message", "query", "prompt", "text", "input"} {
		if v, ok := body[field]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	// Check nested messages array
	if msgs, ok := body["messages"]; ok {
		if arr, ok := msgs.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					if role, _ := m["role"].(string); role == "user" {
						if content, ok := m["content"].(string); ok {
							return content
						}
					}
				}
			}
		}
	}
	return ""
}

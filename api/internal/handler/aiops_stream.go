package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const aiopsStreamMaxScanSize = 1024 * 1024

// streamAIOpsResponse streams the LLM completion as SSE frames that the
// React shell expects:
//
//	data: {"type":"content","content":"<delta>"}\n\n
//	...
//	data: {"type":"done"}\n\n
//
// It persists the user message before calling (caller responsibility) and
// persists the assembled assistant message on completion.
func streamAIOpsResponse(c *gin.Context, sessionID string, userMsg string, history []model.ChatMessage) {
	// SSE headers
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("X-Session-ID", sessionID)
	c.Status(http.StatusOK)
	c.Writer.Flush()

	// Config-missing fallback: emit a helpful hint and exit cleanly.
	if strings.TrimSpace(getAPIKey()) == "" || strings.TrimSpace(getBaseURL()) == "" {
		hint := "LLM 未配置。请设置环境变量 AI_WORKBENCH_API_KEY 和 AI_WORKBENCH_BASE_URL。"
		writeAIOpsSSEContent(c, hint)
		writeAIOpsSSEDone(c)
		persistAIOpsAssistantMessage(sessionID, hint)
		return
	}

	// Build LLM request body
	messages := make([]map[string]string, 0, len(history)+2)
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": "你是 FindX 智能运维助手，帮助用户分析指标、排查故障、提供运维建议。",
	})
	for _, msg := range history {
		role := strings.TrimSpace(msg.Role)
		if role == "" {
			continue
		}
		messages = append(messages, map[string]string{"role": role, "content": msg.Content})
	}
	// The caller has already persisted the user message into history, so we
	// only append userMsg when the latest history entry is not that same
	// message. This keeps the prompt correct whether or not the caller
	// pre-persisted.
	if !lastHistoryIsUser(history, userMsg) {
		messages = append(messages, map[string]string{"role": "user", "content": userMsg})
	}

	reqBody := map[string]any{
		"model":    resolveDefaultModel(),
		"messages": messages,
		"stream":   true,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		writeAIOpsSSEContent(c, fmt.Sprintf("序列化请求失败：%v", err))
		writeAIOpsSSEDone(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), chatUpstreamTimeout)
	defer cancel()

	endpoint := strings.TrimRight(getBaseURL(), "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		writeAIOpsSSEContent(c, fmt.Sprintf("构造上游请求失败：%v", err))
		writeAIOpsSSEDone(c)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+getAPIKey())
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: chatUpstreamTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Warn("aiops stream upstream call failed")
		writeAIOpsSSEContent(c, fmt.Sprintf("调用上游失败：%v", err))
		writeAIOpsSSEDone(c)
		return
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logrus.WithError(cerr).Warn("aiops stream body close failed")
		}
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Try to surface upstream error for easier debugging.
		buf := make([]byte, 4096)
		n, _ := resp.Body.Read(buf)
		msg := fmt.Sprintf("上游返回 %d：%s", resp.StatusCode, strings.TrimSpace(string(buf[:n])))
		writeAIOpsSSEContent(c, msg)
		writeAIOpsSSEDone(c)
		persistAIOpsAssistantMessage(sessionID, msg)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), aiopsStreamMaxScanSize)

	var fullText strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			logrus.WithError(err).Warn("aiops stream chunk parse failed")
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullText.WriteString(delta)
		if !writeAIOpsSSEContent(c, delta) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Warn("aiops stream scan failed")
	}
	writeAIOpsSSEDone(c)
	persistAIOpsAssistantMessage(sessionID, fullText.String())
}

// writeAIOpsSSEContent writes a single `data: {"type":"content","content":"..."}` frame.
// Returns false on write error.
func writeAIOpsSSEContent(c *gin.Context, content string) bool {
	b, err := json.Marshal(map[string]any{"type": "content", "content": content})
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", string(b)); err != nil {
		logrus.WithError(err).Warn("aiops stream content write failed")
		return false
	}
	c.Writer.Flush()
	return true
}

func writeAIOpsSSEDone(c *gin.Context) {
	if _, err := c.Writer.WriteString("data: {\"type\":\"done\"}\n\n"); err != nil {
		logrus.WithError(err).Warn("aiops stream done write failed")
		return
	}
	c.Writer.Flush()
}

func persistAIOpsAssistantMessage(sessionID, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	store.AddChatMessage(model.ChatMessage{
		ID:        store.NewID(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   content,
		Model:     resolveDefaultModel(),
		CreatedAt: time.Now(),
	})
}

func lastHistoryIsUser(history []model.ChatMessage, userMsg string) bool {
	if len(history) == 0 {
		return false
	}
	last := history[len(history)-1]
	return last.Role == "user" && strings.TrimSpace(last.Content) == strings.TrimSpace(userMsg)
}

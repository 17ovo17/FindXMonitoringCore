package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// FeishuEvent represents a Feishu webhook event.
type FeishuEvent struct {
	Schema    string          `json:"schema"`
	Header    FeishuHeader    `json:"header"`
	Event     json.RawMessage `json:"event"`
	Challenge string          `json:"challenge"`
	Token     string          `json:"token"`
	Type      string          `json:"type"`
}

// FeishuHeader is the event header.
type FeishuHeader struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"`
	CreateTime string `json:"create_time"`
	Token      string `json:"token"`
	AppID      string `json:"app_id"`
	TenantKey  string `json:"tenant_key"`
}

// FeishuMessageEvent represents a message event payload.
type FeishuMessageEvent struct {
	Sender  FeishuSender  `json:"sender"`
	Message FeishuMessage `json:"message"`
}

// FeishuSender is the message sender info.
type FeishuSender struct {
	SenderID   FeishuSenderID `json:"sender_id"`
	SenderType string         `json:"sender_type"`
	TenantKey  string         `json:"tenant_key"`
}

// FeishuSenderID contains sender identifiers.
type FeishuSenderID struct {
	UnionID string `json:"union_id"`
	UserID  string `json:"user_id"`
	OpenID  string `json:"open_id"`
}

// FeishuMessage is the message content.
type FeishuMessage struct {
	MessageID   string `json:"message_id"`
	RootID      string `json:"root_id"`
	ParentID    string `json:"parent_id"`
	CreateTime  string `json:"create_time"`
	ChatID      string `json:"chat_id"`
	ChatType    string `json:"chat_type"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
}

// FeishuCardCallback represents a card action callback.
type FeishuCardCallback struct {
	OpenID      string         `json:"open_id"`
	UserID      string         `json:"user_id"`
	OpenMsgID   string         `json:"open_message_id"`
	TenantKey   string         `json:"tenant_key"`
	Token       string         `json:"token"`
	Action      FeishuAction   `json:"action"`
}

// FeishuAction is a card button action.
type FeishuAction struct {
	Value map[string]any `json:"value"`
	Tag   string         `json:"tag"`
}

// BotFeishuWebhook handles POST /api/v1/bot/feishu/webhook
func BotFeishuWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "read body failed"})
		return
	}

	// Verify signature if configured
	secret := viper.GetString("bot.feishu.verification_token")
	if secret != "" {
		timestamp := c.GetHeader("X-Lark-Request-Timestamp")
		nonce := c.GetHeader("X-Lark-Request-Nonce")
		signature := c.GetHeader("X-Lark-Signature")
		if !verifyFeishuSignature(secret, timestamp, nonce, string(body), signature) {
			logrus.Warn("feishu webhook signature verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "signature verification failed"})
			return
		}
	}

	var event FeishuEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid event payload"})
		return
	}

	// Handle URL verification challenge
	if event.Type == "url_verification" || event.Challenge != "" {
		c.JSON(http.StatusOK, gin.H{"challenge": event.Challenge})
		return
	}

	// Route by event type
	eventType := event.Header.EventType
	switch eventType {
	case "im.message.receive_v1":
		handleFeishuMessage(c, event)
	default:
		// Card callback or unknown event
		handleFeishuCardCallback(c, body)
	}
}

func handleFeishuMessage(c *gin.Context, event FeishuEvent) {
	var msgEvent FeishuMessageEvent
	if err := json.Unmarshal(event.Event, &msgEvent); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	// Parse message content
	content := extractFeishuTextContent(msgEvent.Message.Content, msgEvent.Message.MessageType)
	if content == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	// Remove @bot mention
	content = removeFeishuMention(content)

	// Check for prompt injection
	if detected, reason := CheckPromptInjection(content, PromptGuardMedium); detected {
		logrus.WithField("reason", reason).Warn("feishu bot: prompt injection blocked")
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	// Route to AI engine
	sessionID := "feishu-" + msgEvent.Message.ChatID
	answer := runAIOpsDiagnosis(sessionID, content, nil, "ops")

	// Format response as Feishu interactive card
	card := buildFeishuResponseCard(content, answer.Content, answer.SummaryCard)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"message_id": store.NewID(),
			"card":       card,
			"content":    answer.Content,
		},
	})
}

func handleFeishuCardCallback(c *gin.Context, body []byte) {
	var callback FeishuCardCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	action := ""
	if v, ok := callback.Action.Value["action"]; ok {
		action = fmt.Sprint(v)
	}

	switch action {
	case "ai_analyze":
		target := fmt.Sprint(callback.Action.Value["target"])
		answer := runAIOpsDiagnosis("feishu-card", "分析 "+target, nil, "ops")
		card := buildFeishuResponseCard("AI 分析: "+target, answer.Content, answer.SummaryCard)
		c.JSON(http.StatusOK, card)
	case "mute":
		c.JSON(http.StatusOK, buildFeishuTextCard("已静默告警"))
	case "ack":
		c.JSON(http.StatusOK, buildFeishuTextCard("已确认告警"))
	default:
		c.JSON(http.StatusOK, gin.H{"code": 0})
	}
}

func extractFeishuTextContent(content, msgType string) string {
	if msgType != "text" {
		return ""
	}
	var textContent struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(content), &textContent); err != nil {
		return content
	}
	return textContent.Text
}

func removeFeishuMention(text string) string {
	// Remove @_user_1 style mentions
	for {
		idx := strings.Index(text, "@_user_")
		if idx < 0 {
			break
		}
		end := idx + 8
		for end < len(text) && text[end] != ' ' && text[end] != '\n' {
			end++
		}
		text = text[:idx] + text[end:]
	}
	return strings.TrimSpace(text)
}

func buildFeishuResponseCard(question, answer string, summary model.AIOpsSummaryCard) map[string]any {
	severityColor := "blue"
	switch summary.Severity {
	case "P0":
		severityColor = "red"
	case "P1":
		severityColor = "orange"
	case "P2":
		severityColor = "yellow"
	}

	elements := []map[string]any{
		{"tag": "div", "text": map[string]any{"tag": "lark_md", "content": answer}},
		{"tag": "hr"},
		{"tag": "div", "text": map[string]any{"tag": "lark_md", "content": fmt.Sprintf("**严重级别**: %s | **影响**: %s", summary.Severity, summary.Impact)}},
		{"tag": "action", "actions": []map[string]any{
			{"tag": "button", "text": map[string]any{"tag": "plain_text", "content": "AI 深度分析"}, "type": "primary", "value": map[string]any{"action": "ai_analyze", "target": question}},
			{"tag": "button", "text": map[string]any{"tag": "plain_text", "content": "静默"}, "type": "default", "value": map[string]any{"action": "mute"}},
			{"tag": "button", "text": map[string]any{"tag": "plain_text", "content": "确认"}, "type": "default", "value": map[string]any{"action": "ack"}},
		}},
	}

	return map[string]any{
		"config":   map[string]any{"wide_screen_mode": true},
		"header":   map[string]any{"title": map[string]any{"tag": "plain_text", "content": "FindX AIOps 诊断结果"}, "template": severityColor},
		"elements": elements,
	}
}

func buildFeishuTextCard(text string) map[string]any {
	return map[string]any{
		"config":   map[string]any{"wide_screen_mode": true},
		"header":   map[string]any{"title": map[string]any{"tag": "plain_text", "content": "FindX AIOps"}, "template": "blue"},
		"elements": []map[string]any{{"tag": "div", "text": map[string]any{"tag": "lark_md", "content": text}}},
	}
}

func verifyFeishuSignature(secret, timestamp, nonce, body, signature string) bool {
	content := timestamp + nonce + secret + body
	h := hmac.New(sha256.New, []byte(content))
	h.Write([]byte{})
	computed := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return computed == signature
}

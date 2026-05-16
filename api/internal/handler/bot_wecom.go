package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// WeComMessage represents a WeCom incoming message.
type WeComMessage struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	FromUser   string   `xml:"FromUserName"`
	CreateTime int64    `xml:"CreateTime"`
	MsgType    string   `xml:"MsgType"`
	Content    string   `xml:"Content"`
	MsgID      string   `xml:"MsgId"`
	AgentID    int      `xml:"AgentID"`
}

// WeComEncryptedMsg represents an encrypted WeCom message.
type WeComEncryptedMsg struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
	AgentID    string   `xml:"AgentID"`
}

// BotWeComWebhook handles POST /api/v1/bot/wecom/webhook
func BotWeComWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "read body failed"})
		return
	}

	token := viper.GetString("bot.wecom.token")
	encodingAESKey := viper.GetString("bot.wecom.encoding_aes_key")

	// Verify signature
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	if token != "" && msgSignature != "" {
		if !verifyWeComSignature(token, timestamp, nonce, "", msgSignature) {
			logrus.Warn("wecom webhook signature verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "signature verification failed"})
			return
		}
	}

	// Try to parse as encrypted message
	var encMsg WeComEncryptedMsg
	if err := xml.Unmarshal(body, &encMsg); err == nil && encMsg.Encrypt != "" {
		decrypted := decryptWeComMessage(encMsg.Encrypt, encodingAESKey)
		if decrypted != "" {
			body = []byte(decrypted)
		}
	}

	// Parse the message
	var msg WeComMessage
	if err := xml.Unmarshal(body, &msg); err != nil {
		// Try JSON format (for enterprise bot webhook)
		var jsonMsg struct {
			MsgType string `json:"MsgType"`
			Text    struct {
				Content string `json:"Content"`
			} `json:"Text"`
			From struct {
				UserID string `json:"UserId"`
			} `json:"From"`
		}
		if jsonErr := json.Unmarshal(body, &jsonMsg); jsonErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid message format"})
			return
		}
		msg.MsgType = jsonMsg.MsgType
		msg.Content = jsonMsg.Text.Content
		msg.FromUser = jsonMsg.From.UserID
	}

	if msg.Content == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	// Check for prompt injection
	if detected, reason := CheckPromptInjection(msg.Content, PromptGuardMedium); detected {
		logrus.WithField("reason", reason).Warn("wecom bot: prompt injection blocked")
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	// Route to AI engine
	sessionID := "wecom-" + msg.FromUser
	answer := runAIOpsDiagnosis(sessionID, msg.Content, nil, "ops")

	// Format response as WeCom markdown card
	card := buildWeComResponseCard(answer.Content)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"message_id": store.NewID(),
			"msgtype":    "markdown",
			"markdown":   card,
		},
	})
}

// BotWeComVerify handles GET /api/v1/bot/wecom/webhook (URL verification)
func BotWeComVerify(c *gin.Context) {
	token := viper.GetString("bot.wecom.token")
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if token != "" && msgSignature != "" {
		if !verifyWeComSignature(token, timestamp, nonce, echostr, msgSignature) {
			c.String(http.StatusUnauthorized, "signature verification failed")
			return
		}
	}

	// Decrypt echostr if it's encrypted
	encodingAESKey := viper.GetString("bot.wecom.encoding_aes_key")
	if encodingAESKey != "" && echostr != "" {
		decrypted := decryptWeComMessage(echostr, encodingAESKey)
		if decrypted != "" {
			c.String(http.StatusOK, decrypted)
			return
		}
	}

	c.String(http.StatusOK, echostr)
}

func verifyWeComSignature(token, timestamp, nonce, encrypt, expectedSignature string) bool {
	strs := []string{token, timestamp, nonce}
	if encrypt != "" {
		strs = append(strs, encrypt)
	}
	sort.Strings(strs)
	combined := strings.Join(strs, "")
	h := sha1.New()
	h.Write([]byte(combined))
	computed := fmt.Sprintf("%x", h.Sum(nil))
	return computed == expectedSignature
}

func decryptWeComMessage(encrypted, encodingAESKey string) string {
	if encodingAESKey == "" {
		return ""
	}
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil || len(aesKey) < 16 {
		return ""
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil || len(ciphertext) < aes.BlockSize {
		return ""
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return ""
	}
	iv := aesKey[:16]
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	if len(plaintext) == 0 {
		return ""
	}
	padLen := int(plaintext[len(plaintext)-1])
	if padLen > len(plaintext) || padLen > aes.BlockSize {
		return ""
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	// Skip random bytes (16) + msg length (4)
	if len(plaintext) < 20 {
		return ""
	}
	msgLen := binary.BigEndian.Uint32(plaintext[16:20])
	if int(msgLen)+20 > len(plaintext) {
		return ""
	}
	return string(plaintext[20 : 20+msgLen])
}

func buildWeComResponseCard(answer string) map[string]any {
	// Truncate for WeCom markdown limit
	if len([]rune(answer)) > 2000 {
		runes := []rune(answer)
		answer = string(runes[:2000]) + "\n\n... [内容已截断]"
	}
	return map[string]any{
		"content": answer,
	}
}
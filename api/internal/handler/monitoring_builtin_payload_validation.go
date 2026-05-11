package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func bindBuiltinComponentInputs(c *gin.Context) ([]model.MonitoringBuiltinComponentInput, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("empty body")
	}
	var items []model.MonitoringBuiltinComponentInput
	if err := json.Unmarshal(body, &items); err == nil {
		return items, nil
	}
	var item model.MonitoringBuiltinComponentInput
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, err
	}
	return []model.MonitoringBuiltinComponentInput{item}, nil
}

func bindBuiltinPayloadInputs(c *gin.Context) ([]model.MonitoringBuiltinPayloadInput, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("empty body")
	}
	var items []model.MonitoringBuiltinPayloadInput
	if err := json.Unmarshal(body, &items); err == nil {
		return items, nil
	}
	var item model.MonitoringBuiltinPayloadInput
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, err
	}
	return []model.MonitoringBuiltinPayloadInput{item}, nil
}

func normalizeBuiltinPayloadContent(payloadType string, raw json.RawMessage) (json.RawMessage, []string) {
	content := bytes.TrimSpace(raw)
	if len(content) == 0 || bytes.Equal(content, []byte("null")) {
		return nil, []string{"content is required"}
	}
	if len(content) > maxBuiltinContentLen {
		return nil, []string{"content is too large"}
	}
	content = unwrapBuiltinStringContent(content)
	if requiresJSONBuiltinContent(payloadType) && !jsonObjectOrArray(content) {
		return nil, []string{"content must be a JSON object or array"}
	}
	if unsafeBuiltinText(string(content)) || unsafeBuiltinJSONContent(content) {
		return nil, []string{"content contains blocked content"}
	}
	return append([]byte{}, content...), nil
}

func unwrapBuiltinStringContent(content []byte) []byte {
	if len(content) == 0 || content[0] != '"' {
		return content
	}
	var text string
	if err := json.Unmarshal(content, &text); err != nil {
		return content
	}
	inner := bytes.TrimSpace([]byte(text))
	if json.Valid(inner) {
		return inner
	}
	encoded, err := json.Marshal(text)
	if err != nil {
		return content
	}
	return encoded
}

func requiresJSONBuiltinContent(payloadType string) bool {
	switch strings.TrimSpace(payloadType) {
	case "dashboard", "alert":
		return true
	default:
		return false
	}
}

func jsonObjectOrArray(content []byte) bool {
	if !json.Valid(content) || len(content) == 0 {
		return false
	}
	first := bytes.TrimSpace(content)[0]
	return first == '{' || first == '['
}

func unsafeBuiltinJSONContent(content []byte) bool {
	if !json.Valid(content) {
		return false
	}
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		return true
	}
	return unsafeBuiltinJSONValue(value, "")
}

func unsafeBuiltinJSONValue(value any, key string) bool {
	if unsafeBuiltinKey(key) {
		return true
	}
	switch typed := value.(type) {
	case map[string]any:
		for childKey, childValue := range typed {
			if unsafeBuiltinJSONValue(childValue, childKey) {
				return true
			}
		}
	case []any:
		for _, item := range typed {
			if unsafeBuiltinJSONValue(item, "") {
				return true
			}
		}
	case string:
		return unsafeBuiltinText(typed)
	}
	return false
}

func unsafeBuiltinKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return false
	}
	for _, marker := range []string{
		"password", "passwd", "token", "secret", "cookie", "authorization",
		"api_key", "apikey", "access_key", "private_key", "privatekey", "dsn",
		"username", "login_user", "account",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return unsafeBuiltinText(key)
}

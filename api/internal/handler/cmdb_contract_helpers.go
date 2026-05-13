package handler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

func anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func cmdbBlockedContractEnvelope(contractID string, missing []string, details ...gin.H) gin.H {
	envelope := gin.H{
		"code":              cmdbBlockedByContract,
		"status":            strings.ToLower(cmdbBlockedByContract),
		"message":           "CMDB 兼容契约尚未接入，当前接口不会返回伪造成功数据",
		"contract_id":       contractID,
		"missing_contracts": missing,
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
	if len(details) > 0 {
		for key, value := range details[0] {
			envelope[key] = value
		}
	}
	return envelope
}

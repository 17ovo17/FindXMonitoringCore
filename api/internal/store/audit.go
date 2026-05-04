package store

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func auditHMACKey() string {
	if value := strings.TrimSpace(os.Getenv("AIW_AUDIT_HMAC_KEY")); value != "" {
		return value
	}
	if value := strings.TrimSpace(viper.GetString("audit.hmac_key")); value != "" {
		return value
	}
	// 开发 fallback 仅保证本地测试可运行，生产环境必须配置 AIW_AUDIT_HMAC_KEY 或 audit.hmac_key。
	host, err := os.Hostname()
	if err != nil {
		logrus.WithError(err).Warn("audit hmac dev fallback hostname unavailable")
		host = "unknown-host"
	}
	return "dev-fallback:" + host + ":" + os.TempDir()
}

// computeAuditHash 计算审计事件的 HMAC 签名，用于防篡改校验。
func computeAuditHash(e *AuditEvent) string {
	content := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		e.Action, e.Target, e.Risk, e.Decision,
		e.Detail, e.Operator, e.Description, e.CreatedAt.Format(time.RFC3339))
	mac := hmac.New(sha256.New, []byte(auditHMACKey()))
	mac.Write([]byte(content))
	return hex.EncodeToString(mac.Sum(nil))
}

// AddAuditEvent persists an audit event to memory and MySQL.
func AddAuditEvent(e AuditEvent) {
	fillAuditDerivedFields(&e)
	e.IntegrityHash = computeAuditHash(&e)
	mu.Lock()
	auditEvents = append([]AuditEvent{e}, auditEvents...)
	if len(auditEvents) > 1000 {
		auditEvents = auditEvents[:1000]
	}
	mu.Unlock()
	if mysqlOK {
		if _, err := db.Exec(`REPLACE INTO audit_events (id,action,target,risk,decision,detail,operator,description,test_batch_id,client_ip,created_at,integrity_hash) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			e.ID, e.Action, e.Target, e.Risk, e.Decision, e.Detail, e.Operator, e.Description, e.TestBatchID, e.ClientIP, e.CreatedAt, e.IntegrityHash); err != nil {
			logAuditPersistWarning("legacy audit persist failed", err, e)
		}
	}
	if _, err := AddMonitorAuditLog(MonitorAuditLogFromLegacy(e)); err != nil {
		logAuditPersistWarning("monitor audit mirror failed", err, e)
	}
}

func logAuditPersistWarning(message string, err error, e AuditEvent) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"id":     e.ID,
		"action": e.Action,
		"target": sanitizeAuditString(e.Target, "target"),
		"status": e.Decision,
	}).Warn(message)
}

// ListAuditEvents returns audit events ordered by created_at desc.
func ListAuditEvents(limit int) []AuditEvent {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	if mysqlOK {
		rows, err := db.Query(`SELECT id,action,target,risk,decision,detail,COALESCE(operator,''),COALESCE(description,''),test_batch_id,client_ip,created_at,COALESCE(integrity_hash,'') FROM audit_events ORDER BY created_at DESC LIMIT ?`, limit)
		if err == nil {
			defer rows.Close()
			out := []AuditEvent{}
			for rows.Next() {
				var e AuditEvent
				if err := rows.Scan(&e.ID, &e.Action, &e.Target, &e.Risk, &e.Decision, &e.Detail, &e.Operator, &e.Description, &e.TestBatchID, &e.ClientIP, &e.CreatedAt, &e.IntegrityHash); err != nil {
					logrus.WithError(err).Warn("legacy audit row scan failed")
					continue
				}
				fillAuditDerivedFields(&e)
				sanitizeAuditEventForOutput(&e)
				out = append(out, e)
			}
			if err := rows.Err(); err != nil {
				logrus.WithError(err).Warn("legacy audit rows iteration failed")
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	if len(auditEvents) < limit {
		limit = len(auditEvents)
	}
	out := make([]AuditEvent, limit)
	copy(out, auditEvents[:limit])
	for i := range out {
		fillAuditDerivedFields(&out[i])
		sanitizeAuditEventForOutput(&out[i])
	}
	return out
}

func fillAuditDerivedFields(e *AuditEvent) {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.Timestamp == "" {
		e.Timestamp = e.CreatedAt.Format(time.RFC3339)
	}
	if e.Operator == "" {
		e.Operator = "anonymous"
	}
	if e.Description == "" {
		e.Description = fmt.Sprintf("%s 操作对象 %s，结果 %s", e.Action, e.Target, e.Decision)
		if e.Detail != "" {
			e.Description += "，说明：" + e.Detail
		}
	}
}

func sanitizeAuditEventForOutput(e *AuditEvent) {
	e.Target = sanitizeAuditString(e.Target, "target")
	e.Detail = sanitizeAuditString(e.Detail, "detail")
	e.Description = sanitizeAuditString(e.Description, "description")
}

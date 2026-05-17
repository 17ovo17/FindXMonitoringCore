package sandbox

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ApprovalStatus 审批状态
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalDenied   ApprovalStatus = "denied"
	ApprovalTimeout  ApprovalStatus = "timeout"
)

// ApprovalRequest 用户确认请求
type ApprovalRequest struct {
	ID          string         `json:"id"`
	ToolName    string         `json:"tool_name"`
	RiskLevel   int            `json:"risk_level"`
	Params      map[string]any `json:"params"`
	Description string         `json:"description"`
	Impact      string         `json:"impact"`
	Status      ApprovalStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy  string         `json:"resolved_by,omitempty"`
	mu          sync.Mutex
	ch          chan ApprovalStatus
}

// AuditEntry 审计日志条目
type AuditEntry struct {
	ID         string         `json:"id"`
	SessionID  string         `json:"session_id"`
	ToolName   string         `json:"tool_name"`
	RiskLevel  int            `json:"risk_level"`
	Params     map[string]any `json:"params"`
	Result     any            `json:"result,omitempty"`
	Error      string         `json:"error,omitempty"`
	Status     string         `json:"status"` // executed, denied, timeout, rolled_back
	Duration   int64          `json:"duration_ms"`
	CreatedAt  time.Time      `json:"created_at"`
	ApprovalID string         `json:"approval_id,omitempty"`
}

// Sandbox AI 执行沙箱
type Sandbox struct {
	policy    Policy
	approvals map[string]*ApprovalRequest
	audit     []AuditEntry
	mu        sync.RWMutex
}

// New 创建沙箱实例
func New(policy Policy) *Sandbox {
	return &Sandbox{
		policy:    policy,
		approvals: make(map[string]*ApprovalRequest),
	}
}

// PreExecute 执行前检查
// 返回值：
// - nil, nil: 直接执行（Level 0，或 Level 1 在 auto_review 模式）
// - *ApprovalRequest, nil: 需要等待用户确认（Level 2）
// - nil, error: 拒绝执行（黑名单命中、权限不足）
func (s *Sandbox) PreExecute(toolName string, riskLevel int, params map[string]any) (*ApprovalRequest, error) {
	// 1. 检查命令黑名单（如果是 remote_exec）
	if cmd, ok := params["command"].(string); ok && toolName == "remote_exec" {
		for _, denied := range s.policy.DeniedCommands {
			if strings.Contains(strings.ToLower(cmd), strings.ToLower(denied)) {
				return nil, fmt.Errorf("命令被安全策略拒绝：包含危险操作 '%s'", denied)
			}
		}
	}

	// 2. 根据 policy mode 和 risk level 决定
	switch s.policy.Mode {
	case ModeReadonly:
		if riskLevel > 0 {
			return nil, fmt.Errorf("当前为只读模式，不允许执行 risk_level=%d 的操作", riskLevel)
		}
		return nil, nil

	case ModeAutoReview:
		if riskLevel >= 2 {
			return nil, fmt.Errorf("当前为自动审查模式，不允许执行 risk_level=%d 的操作，请切换到完全访问模式", riskLevel)
		}
		return nil, nil

	case ModeFullAccess:
		if riskLevel >= 2 {
			req := s.createApprovalRequest(toolName, riskLevel, params)
			return req, nil
		}
		return nil, nil
	}

	return nil, nil
}

// WaitApproval 等待用户确认（阻塞，30s 超时）
func (s *Sandbox) WaitApproval(ctx context.Context, req *ApprovalRequest) (ApprovalStatus, error) {
	select {
	case status := <-req.ch:
		return status, nil
	case <-time.After(30 * time.Second):
		req.mu.Lock()
		req.Status = ApprovalTimeout
		req.mu.Unlock()
		return ApprovalTimeout, nil
	case <-ctx.Done():
		return ApprovalDenied, ctx.Err()
	}
}

// Resolve 用户确认/拒绝
func (s *Sandbox) Resolve(approvalID string, status ApprovalStatus, resolvedBy string) error {
	s.mu.Lock()
	req, ok := s.approvals[approvalID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("approval request not found: %s", approvalID)
	}
	req.mu.Lock()
	defer req.mu.Unlock()
	if req.Status != ApprovalPending {
		return fmt.Errorf("approval already resolved: %s", req.Status)
	}
	now := time.Now()
	req.Status = status
	req.ResolvedAt = &now
	req.ResolvedBy = resolvedBy
	req.ch <- status
	return nil
}

// RecordAudit 记录审计日志
func (s *Sandbox) RecordAudit(entry AuditEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	s.audit = append(s.audit, entry)
	logrus.WithFields(logrus.Fields{
		"tool":       entry.ToolName,
		"risk_level": entry.RiskLevel,
		"status":     entry.Status,
		"duration":   entry.Duration,
	}).Info("sandbox: audit recorded")
}

// GetPolicy 获取当前策略
func (s *Sandbox) GetPolicy() Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policy
}

// SetPolicy 更新策略
func (s *Sandbox) SetPolicy(p Policy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = p
}

// ListAudit 查询审计日志
func (s *Sandbox) ListAudit(limit int) []AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := len(s.audit)
	if limit <= 0 || limit > total {
		limit = total
	}
	start := total - limit
	if start < 0 {
		start = 0
	}
	result := make([]AuditEntry, total-start)
	copy(result, s.audit[start:])
	return result
}

// ListPendingApprovals 获取待确认的请求
func (s *Sandbox) ListPendingApprovals() []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var pending []*ApprovalRequest
	for _, req := range s.approvals {
		if req.Status == ApprovalPending {
			pending = append(pending, req)
		}
	}
	return pending
}

func (s *Sandbox) createApprovalRequest(toolName string, riskLevel int, params map[string]any) *ApprovalRequest {
	id := fmt.Sprintf("approval_%d", time.Now().UnixNano())
	req := &ApprovalRequest{
		ID:          id,
		ToolName:    toolName,
		RiskLevel:   riskLevel,
		Params:      params,
		Description: describeToolAction(toolName, params),
		Impact:      estimateImpact(toolName, params),
		Status:      ApprovalPending,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(30 * time.Second),
		ch:          make(chan ApprovalStatus, 1),
	}
	s.mu.Lock()
	s.approvals[id] = req
	s.mu.Unlock()
	return req
}

func describeToolAction(toolName string, params map[string]any) string {
	switch toolName {
	case "remote_exec":
		return fmt.Sprintf("在 %s 上执行命令: %s", params["ip"], truncate(fmt.Sprint(params["command"]), 100))
	case "notify_send":
		return fmt.Sprintf("发送通知到渠道 %s", params["channel"])
	case "alert_mute":
		return fmt.Sprintf("静默告警: %v", params["id"])
	case "alert_resolve":
		return fmt.Sprintf("关闭告警: %s", params["id"])
	case "workflow_trigger":
		return fmt.Sprintf("触发工作流: %s", params["workflow_id"])
	default:
		return fmt.Sprintf("执行 %s", toolName)
	}
}

func estimateImpact(toolName string, params map[string]any) string {
	switch toolName {
	case "remote_exec":
		return fmt.Sprintf("影响主机: %s", params["ip"])
	case "workflow_trigger":
		return fmt.Sprintf("触发工作流: %s", params["workflow_id"])
	case "alert_mute":
		return "告警将被静默，可能导致后续告警被忽略"
	default:
		return "影响范围有限"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

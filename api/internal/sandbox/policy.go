package sandbox

// PolicyMode 定义沙箱运行模式
type PolicyMode string

const (
	ModeReadonly   PolicyMode = "readonly"    // 只读，只允许 risk_level=0
	ModeAutoReview PolicyMode = "auto_review" // 自动审查，允许 0+1，记录审计
	ModeFullAccess PolicyMode = "full_access" // 完全访问，允许 0+1+2，Level 2 需确认
)

// Policy 定义沙箱安全策略
type Policy struct {
	Mode           PolicyMode `json:"mode"`
	DeniedCommands []string   `json:"denied_commands"` // 命令黑名单
	MaxTimeout     int        `json:"max_timeout"`     // 最大执行超时（秒）
	AuditAll       bool       `json:"audit_all"`       // 是否审计所有操作
}

// DefaultPolicy 默认策略：自动审查模式
var DefaultPolicy = Policy{
	Mode:           ModeAutoReview,
	DeniedCommands: []string{"rm -rf /", "dd if=", "mkfs", "shutdown", "reboot", "halt", "init 0", "init 6", "> /dev/sda"},
	MaxTimeout:     300,
	AuditAll:       true,
}

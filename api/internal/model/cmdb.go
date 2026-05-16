package model

import "time"

// CmdbCategory 模型分类（计算资源、系统软件、应用软件等）
type CmdbCategory struct {
	ID       string `gorm:"primaryKey;size:32" json:"id"`
	Label    string `gorm:"size:64;not null" json:"label"`
	ParentID string `gorm:"size:32;index" json:"parent_id"`
	Sort     int    `gorm:"default:0" json:"sort"`
}

// CmdbObject 模型定义（操作系统、数据库、中间件等）
type CmdbObject struct {
	ID         string    `gorm:"primaryKey;size:32" json:"id"`
	Name       string    `gorm:"size:64;not null" json:"name"`
	CategoryID string    `gorm:"size:32;index" json:"category_id"`
	ObjectType int       `gorm:"default:101" json:"object_type"`
	Icon       string    `gorm:"size:32" json:"icon"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CmdbAttribute 模型属性定义
type CmdbAttribute struct {
	ID         string `gorm:"primaryKey;size:32" json:"id"`
	ObjectID   string `gorm:"size:32;index;not null" json:"object_id"`
	Label      string `gorm:"size:64;not null" json:"label"`
	Attr       string `gorm:"size:64;not null" json:"attr"`       // 属性标识符
	ValueType  string `gorm:"size:16;not null" json:"value_type"` // char/int/float/ip/boolean/enum/array/struct
	Tag        string `gorm:"size:32" json:"tag"`                 // 属性分组标签
	Required   bool   `gorm:"default:false" json:"required"`
	Unique     bool   `gorm:"default:false" json:"unique"`
	Discovery  bool   `gorm:"default:false" json:"discovery"` // Agent 自动发现可填充
	Sort       int    `gorm:"default:0" json:"sort"`
	Unit       string `gorm:"size:16" json:"unit"`
	Options    string `gorm:"type:text" json:"options"` // enum 选项 JSON
	DefaultVal string `gorm:"size:256" json:"default_val"`
}

// CmdbInstance 资产实例
type CmdbInstance struct {
	ID        string    `gorm:"primaryKey;size:32" json:"id"`
	ObjectID  string    `gorm:"size:32;index;not null" json:"object_id"`
	Data      string    `gorm:"type:mediumtext" json:"data"` // JSON 存储属性值
	Creator   string    `gorm:"size:64" json:"creator"`
	Updater   string    `gorm:"size:64" json:"updater"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CmdbRelationType 关联类型定义
type CmdbRelationType struct {
	ID             string `gorm:"primaryKey;size:96" json:"id"`
	Name           string `gorm:"size:32;not null" json:"name"` // belong/default/自定义
	Label          string `gorm:"size:64" json:"label"`
	Mapping        string `gorm:"size:16" json:"mapping"`
	Visible        *bool  `json:"visible,omitempty"`
	Description    string `gorm:"size:256" json:"description"`
	RuleLogic      string `gorm:"size:16" json:"rule_logic"`
	RuleExpression string `gorm:"size:256" json:"rule_expression"`
	RulesJSON      string `gorm:"type:text" json:"-"`
	LeftMin        int    `json:"left_min"`
	RightMin       int    `json:"right_min"`
	LeftMax        int    `json:"left_max"`
	RightMax       int    `json:"right_max"`
	Source         int    `json:"source"`
	LeftAsstName   string `gorm:"size:64" json:"left_asst_name"`
	RightAsstName  string `gorm:"size:64" json:"right_asst_name"`
}

// CmdbInstanceRelation 实例间关联
type CmdbInstanceRelation struct {
	ID               string    `gorm:"primaryKey;size:96" json:"id"`
	SourceInstanceID string    `gorm:"size:32;index;not null" json:"source_instance_id"`
	TargetInstanceID string    `gorm:"size:32;index;not null" json:"target_instance_id"`
	RelationTypeID   string    `gorm:"size:96;not null" json:"relation_type_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type CmdbMonitorBinding struct {
	ID               string    `gorm:"primaryKey;size:64" json:"id"`
	InstanceID       string    `gorm:"size:32;index;not null" json:"instance_id"`
	Host             string    `gorm:"size:128" json:"host"`
	HostID           string    `gorm:"size:128;index" json:"hostid"`
	TemplateID       string    `gorm:"size:128;index" json:"templateid"`
	ServerObjectID   string    `gorm:"size:64" json:"server_object_id"`
	ServerPlatformID string    `gorm:"size:64" json:"server_platform_id"`
	CmdbObjectID     string    `gorm:"size:64;index" json:"cmdb_object_id"`
	GroupJSON        string    `gorm:"column:binding_group;type:text" json:"-"`
	TagsJSON         string    `gorm:"type:text" json:"-"`
	ActiveStatus     string    `gorm:"size:32" json:"active_status"`
	HostType         string    `gorm:"size:64" json:"hosttype"`
	SubType          string    `gorm:"size:64" json:"subtype"`
	HostTypeLabel    string    `gorm:"size:128" json:"hosttypeLabel"`
	SubTypeLabel     string    `gorm:"size:128" json:"subtypeLabel"`
	CmdbAttrID       string    `gorm:"size:128" json:"cmdb_attr_id"`
	ServerAttrID     string    `gorm:"size:128" json:"server_attr_id"`
	ServerModelID    string    `gorm:"size:128" json:"server_model_id"`
	ServerModelName  string    `gorm:"size:128" json:"server_model_name"`
	Attr             string    `gorm:"size:128" json:"attr"`
	AttrStruJSON     string    `gorm:"type:text" json:"-"`
	Queue            string    `gorm:"size:128" json:"queue"`
	Creator          string    `gorm:"size:64" json:"creator"`
	Updater          string    `gorm:"size:64" json:"updater"`
	AuditRef         string    `gorm:"size:128" json:"audit_ref"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CmdbMonitorBindingReceipt struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	BindingID   string    `gorm:"size:64;index;not null" json:"binding_id"`
	InstanceID  string    `gorm:"size:32;index;not null" json:"instance_id"`
	ReceiptType string    `gorm:"size:32;index;not null" json:"receipt_type"`
	Status      string    `gorm:"size:64;index;not null" json:"status"`
	ContractID  string    `gorm:"size:128" json:"contract_id"`
	MissingJSON string    `gorm:"type:text" json:"-"`
	RequestRef  string    `gorm:"size:128;index" json:"request_ref,omitempty"`
	AuditRef    string    `gorm:"size:128" json:"audit_ref"`
	CreatedAt   time.Time `json:"created_at"`
}

type CmdbRelationActionRequest struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	Action         string    `gorm:"size:32;index;not null" json:"action"`
	InstanceID     string    `gorm:"size:32;index;not null" json:"instance_id"`
	NodeID         string    `gorm:"size:32;index" json:"node_id"`
	ObjectID       string    `gorm:"size:32;index" json:"object_id"`
	RelationID     string    `gorm:"size:64;index;not null" json:"relation_id"`
	Actor          string    `gorm:"size:64" json:"actor"`
	Status         string    `gorm:"size:32;index" json:"status"`
	DeliveryStatus string    `gorm:"size:64" json:"delivery_status"`
	EffectStatus   string    `gorm:"size:64" json:"effect_status"`
	ContextJSON    string    `gorm:"type:text" json:"-"`
	AuditRef       string    `gorm:"size:128" json:"audit_ref"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CmdbRelationActionReceipt struct {
	ID              string    `gorm:"primaryKey;size:64" json:"id"`
	ActionRequestID string    `gorm:"size:64;index;not null" json:"action_request_id"`
	Action          string    `gorm:"size:32;index;not null" json:"action"`
	InstanceID      string    `gorm:"size:32;index;not null" json:"instance_id"`
	NodeID          string    `gorm:"size:32;index" json:"node_id"`
	RelationID      string    `gorm:"size:64;index;not null" json:"relation_id"`
	ReceiptType     string    `gorm:"size:32;index;not null" json:"receipt_type"`
	Status          string    `gorm:"size:64;index;not null" json:"status"`
	ContractID      string    `gorm:"size:128" json:"contract_id"`
	MissingJSON     string    `gorm:"type:text" json:"-"`
	RequestRef      string    `gorm:"size:128;index" json:"request_ref,omitempty"`
	AuditRef        string    `gorm:"size:128" json:"audit_ref"`
	CreatedAt       time.Time `json:"created_at"`
}

type CmdbDeployTask struct {
	ID              string    `gorm:"primaryKey;size:64" json:"id"`
	Name            string    `gorm:"size:128;not null" json:"name"`
	TargetHostsJSON string    `gorm:"type:text" json:"-"`
	ScriptLength    int       `json:"script_length"`
	ScriptDigest    string    `gorm:"size:64" json:"script_digest"`
	Status          string    `gorm:"size:64;index" json:"status"`
	Progress        int       `json:"progress"`
	Creator         string    `gorm:"size:64" json:"creator"`
	LogsJSON        string    `gorm:"type:text" json:"-"`
	Code            string    `gorm:"size:64" json:"code"`
	ContractID      string    `gorm:"size:128" json:"contract_id"`
	MissingJSON     string    `gorm:"type:text" json:"-"`
	SafeToRetry     bool      `json:"safe_to_retry"`
	AuditRef        string    `gorm:"size:256" json:"audit_ref"`
	LogRef          string    `gorm:"size:256" json:"log_ref"`
	MetaJSON        string    `gorm:"type:text" json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

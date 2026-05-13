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
	ID    string `gorm:"primaryKey;size:32" json:"id"`
	Name  string `gorm:"size:32;not null" json:"name"` // belong/default/自定义
	Label string `gorm:"size:64" json:"label"`
}

// CmdbInstanceRelation 实例间关联
type CmdbInstanceRelation struct {
	ID               string    `gorm:"primaryKey;size:32" json:"id"`
	SourceInstanceID string    `gorm:"size:32;index;not null" json:"source_instance_id"`
	TargetInstanceID string    `gorm:"size:32;index;not null" json:"target_instance_id"`
	RelationTypeID   string    `gorm:"size:32;not null" json:"relation_type_id"`
	CreatedAt        time.Time `json:"created_at"`
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

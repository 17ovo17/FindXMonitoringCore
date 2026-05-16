package model

import "time"

// AlertPipeline 定义告警事件流水线，事件在到达通知阶段前依次经过各处理器。
type AlertPipeline struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Enabled    bool                `json:"enabled"`
	Priority   int                 `json:"priority"` // 数值越小优先级越高
	Processors []PipelineProcessor `json:"processors"`
	CreatedBy  string              `json:"created_by"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// PipelineProcessor 是流水线中的单个处理步骤。
type PipelineProcessor struct {
	Type       string         `json:"type"`       // relabel, drop, callback, enrich
	Enabled    bool           `json:"enabled"`
	Conditions []LabelMatcher `json:"conditions"` // 匹配条件，全部满足时执行
	Config     map[string]any `json:"config"`     // 处理器特定配置
}

// LabelMatcher 用于匹配事件标签。
type LabelMatcher struct {
	Key      string `json:"key"`
	Operator string `json:"operator"` // =, !=, =~, !~
	Value    string `json:"value"`
}

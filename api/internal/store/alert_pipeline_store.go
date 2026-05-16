package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	alertPipelines   = map[string]*model.AlertPipeline{}
	alertPipelinesMu sync.RWMutex

	ErrAlertPipelineNotFound   = errors.New("alert pipeline not found")
	ErrAlertPipelineValidation = errors.New("alert pipeline validation failed")
)

func newPipelineID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "pl_" + hex.EncodeToString(b)
}

// ListAlertPipelines 返回所有流水线，按 Priority 升序排列。
func ListAlertPipelines() []model.AlertPipeline {
	alertPipelinesMu.RLock()
	defer alertPipelinesMu.RUnlock()
	out := make([]model.AlertPipeline, 0, len(alertPipelines))
	for _, p := range alertPipelines {
		out = append(out, copyAlertPipeline(p))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	return out
}

// ListEnabledAlertPipelines 返回所有已启用的流水线，按 Priority 升序排列。
func ListEnabledAlertPipelines() []model.AlertPipeline {
	all := ListAlertPipelines()
	out := make([]model.AlertPipeline, 0, len(all))
	for _, p := range all {
		if p.Enabled {
			out = append(out, p)
		}
	}
	return out
}

// GetAlertPipeline 按 ID 查询单个流水线。
func GetAlertPipeline(id string) (*model.AlertPipeline, bool) {
	alertPipelinesMu.RLock()
	defer alertPipelinesMu.RUnlock()
	p, ok := alertPipelines[id]
	if !ok {
		return nil, false
	}
	cp := copyAlertPipeline(p)
	return &cp, true
}

// CreateAlertPipeline 创建流水线，自动生成 ID 和时间戳。
func CreateAlertPipeline(p model.AlertPipeline) (model.AlertPipeline, error) {
	if strings.TrimSpace(p.Name) == "" {
		return model.AlertPipeline{}, ErrAlertPipelineValidation
	}
	now := time.Now()
	p.ID = newPipelineID()
	p.CreatedAt = now
	p.UpdatedAt = now
	alertPipelinesMu.Lock()
	defer alertPipelinesMu.Unlock()
	cp := p
	alertPipelines[cp.ID] = &cp
	return p, nil
}

// UpdateAlertPipeline 更新指定 ID 的流水线。
func UpdateAlertPipeline(id string, p model.AlertPipeline) (model.AlertPipeline, error) {
	if strings.TrimSpace(p.Name) == "" {
		return model.AlertPipeline{}, ErrAlertPipelineValidation
	}
	alertPipelinesMu.Lock()
	defer alertPipelinesMu.Unlock()
	existing, ok := alertPipelines[id]
	if !ok {
		return model.AlertPipeline{}, ErrAlertPipelineNotFound
	}
	p.ID = id
	p.CreatedAt = existing.CreatedAt
	p.CreatedBy = existing.CreatedBy
	p.UpdatedAt = time.Now()
	cp := p
	alertPipelines[id] = &cp
	return p, nil
}

// DeleteAlertPipeline 删除指定 ID 的流水线。
func DeleteAlertPipeline(id string) error {
	alertPipelinesMu.Lock()
	defer alertPipelinesMu.Unlock()
	if _, ok := alertPipelines[id]; !ok {
		return ErrAlertPipelineNotFound
	}
	delete(alertPipelines, id)
	return nil
}

func copyAlertPipeline(p *model.AlertPipeline) model.AlertPipeline {
	cp := *p
	if p.Processors != nil {
		cp.Processors = make([]model.PipelineProcessor, len(p.Processors))
		copy(cp.Processors, p.Processors)
	}
	return cp
}

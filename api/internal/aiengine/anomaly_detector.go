package aiengine

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// AnomalyDetector — 无规则异常检测（对标 Datadog Watchdog）
// ---------------------------------------------------------------------------
//
// 核心理念：不需要用户配告警规则，自动发现指标异常模式
//
// 检测算法：
// 1. 基线学习：对每个指标学习正常范围（滑动窗口均值 ± 3σ）
// 2. 突变检测：当前值与基线偏差超过阈值
// 3. 趋势检测：指标持续上升/下降超过 N 分钟
// 4. 周期异常：与同期（昨天/上周同时段）对比异常
// 5. 关联检测：多个指标同时异常 → 可能是同一根因

// AnomalyType 异常类型
type AnomalyType string

const (
	AnomalySpike      AnomalyType = "spike"      // 突增
	AnomalyDrop       AnomalyType = "drop"       // 突降
	AnomalyTrend      AnomalyType = "trend"      // 持续趋势
	AnomalyCyclic     AnomalyType = "cyclic"     // 周期异常
	AnomalyCorrelated AnomalyType = "correlated" // 关联异常
)

// Anomaly 检测到的异常
type Anomaly struct {
	ID               string            `json:"id"`
	MetricName       string            `json:"metric_name"`
	Labels           map[string]string `json:"labels"`
	Type             AnomalyType       `json:"type"`
	Severity         string            `json:"severity"` // low, medium, high, critical
	CurrentValue     float64           `json:"current_value"`
	ExpectedValue    float64           `json:"expected_value"`
	Deviation        float64           `json:"deviation"` // 偏差倍数（σ）
	StartedAt        time.Time         `json:"started_at"`
	Description      string            `json:"description"`
	RelatedAnomalies []string          `json:"related_anomalies,omitempty"`
}

// AnomalyGroup 关联异常组
type AnomalyGroup struct {
	ID        string    `json:"id"`
	Anomalies []string  `json:"anomaly_ids"`
	RootCause string    `json:"root_cause,omitempty"`
	StartedAt time.Time `json:"started_at"`
}

// MetricSample 指标采样点
type MetricSample struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Baseline 指标基线
type Baseline struct {
	MetricKey   string    `json:"metric_key"` // metric_name + labels hash
	Mean        float64   `json:"mean"`
	StdDev      float64   `json:"std_dev"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	SampleCount int       `json:"sample_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AnomalyDetector 无规则异常检测器
type AnomalyDetector struct {
	mu        sync.RWMutex
	baselines map[string]*Baseline  // metricKey -> baseline
	anomalies map[string]*Anomaly   // anomaly ID -> anomaly
	// 检测参数
	sigmaThreshold float64 // 偏差阈值（默认 3σ）
	trendWindow    int     // 趋势检测窗口（分钟）
	correlateWindow time.Duration // 关联检测时间窗口
}

// NewAnomalyDetector 创建异常检测器
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		baselines:       make(map[string]*Baseline),
		anomalies:       make(map[string]*Anomaly),
		sigmaThreshold:  3.0,
		trendWindow:     10,
		correlateWindow: 5 * time.Minute,
	}
}

// Detect 检测异常（定期调用，如每 30s）
func (d *AnomalyDetector) Detect(metrics []MetricSample) []Anomaly {
	detected := make([]Anomaly, 0)

	for _, sample := range metrics {
		metricKey := buildMetricKey(sample.MetricName, sample.Labels)

		d.mu.RLock()
		baseline, hasBaseline := d.baselines[metricKey]
		d.mu.RUnlock()

		if !hasBaseline {
			// 没有基线，先学习
			d.initBaseline(metricKey, sample.Value)
			continue
		}

		// 突变检测：当前值与基线偏差超过阈值
		if anomaly := d.detectSpike(sample, baseline); anomaly != nil {
			detected = append(detected, *anomaly)
		}
	}

	// 存储检测到的异常
	d.mu.Lock()
	for i := range detected {
		d.anomalies[detected[i].ID] = &detected[i]
	}
	d.mu.Unlock()

	return detected
}

// UpdateBaseline 更新基线（增量更新，使用 Welford 在线算法）
func (d *AnomalyDetector) UpdateBaseline(metricKey string, values []float64) {
	if len(values) == 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	baseline, exists := d.baselines[metricKey]
	if !exists {
		baseline = &Baseline{MetricKey: metricKey}
		d.baselines[metricKey] = baseline
	}

	// Welford 在线算法计算均值和标准差
	for _, v := range values {
		baseline.SampleCount++
		delta := v - baseline.Mean
		baseline.Mean += delta / float64(baseline.SampleCount)
		delta2 := v - baseline.Mean
		// 使用 M2 累积方差（存储在 StdDev 字段临时复用）
		if baseline.SampleCount == 1 {
			baseline.StdDev = 0
			baseline.Min = v
			baseline.Max = v
		} else {
			baseline.StdDev += delta * delta2
			if v < baseline.Min {
				baseline.Min = v
			}
			if v > baseline.Max {
				baseline.Max = v
			}
		}
	}

	// 计算最终标准差
	if baseline.SampleCount > 1 {
		variance := baseline.StdDev / float64(baseline.SampleCount-1)
		baseline.StdDev = math.Sqrt(variance)
	}
	baseline.UpdatedAt = time.Now()
}

// CorrelateAnomalies 关联同时段的异常
func (d *AnomalyDetector) CorrelateAnomalies(anomalies []Anomaly) []AnomalyGroup {
	if len(anomalies) < 2 {
		return nil
	}

	groups := make([]AnomalyGroup, 0)
	used := make(map[string]bool)

	for i := range anomalies {
		if used[anomalies[i].ID] {
			continue
		}
		group := AnomalyGroup{
			ID:        uuid.New().String(),
			Anomalies: []string{anomalies[i].ID},
			StartedAt: anomalies[i].StartedAt,
		}

		// 找到时间窗口内的其他异常
		for j := i + 1; j < len(anomalies); j++ {
			if used[anomalies[j].ID] {
				continue
			}
			timeDiff := anomalies[j].StartedAt.Sub(anomalies[i].StartedAt)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff <= d.correlateWindow {
				group.Anomalies = append(group.Anomalies, anomalies[j].ID)
				used[anomalies[j].ID] = true
			}
		}

		if len(group.Anomalies) > 1 {
			used[anomalies[i].ID] = true
			groups = append(groups, group)
		}
	}

	return groups
}

// ListAnomalies 获取当前异常列表
func (d *AnomalyDetector) ListAnomalies() []*Anomaly {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]*Anomaly, 0, len(d.anomalies))
	for _, a := range d.anomalies {
		result = append(result, a)
	}
	return result
}

// detectSpike 突变检测
func (d *AnomalyDetector) detectSpike(sample MetricSample, baseline *Baseline) *Anomaly {
	if baseline.StdDev == 0 {
		return nil
	}

	deviation := (sample.Value - baseline.Mean) / baseline.StdDev

	if math.Abs(deviation) < d.sigmaThreshold {
		return nil
	}

	anomalyType := AnomalySpike
	if deviation < 0 {
		anomalyType = AnomalyDrop
	}

	severity := "low"
	absDeviation := math.Abs(deviation)
	switch {
	case absDeviation >= 6:
		severity = "critical"
	case absDeviation >= 5:
		severity = "high"
	case absDeviation >= 4:
		severity = "medium"
	}

	return &Anomaly{
		ID:            uuid.New().String(),
		MetricName:    sample.MetricName,
		Labels:        sample.Labels,
		Type:          anomalyType,
		Severity:      severity,
		CurrentValue:  sample.Value,
		ExpectedValue: baseline.Mean,
		Deviation:     deviation,
		StartedAt:     sample.Timestamp,
		Description:   buildAnomalyDescription(sample, anomalyType, deviation),
	}
}

// initBaseline 初始化基线
func (d *AnomalyDetector) initBaseline(metricKey string, value float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.baselines[metricKey] = &Baseline{
		MetricKey:   metricKey,
		Mean:        value,
		StdDev:      0,
		Min:         value,
		Max:         value,
		SampleCount: 1,
		UpdatedAt:   time.Now(),
	}
}

// buildMetricKey 构建指标唯一键
func buildMetricKey(metricName string, labels map[string]string) string {
	key := metricName
	for k, v := range labels {
		key += "|" + k + "=" + v
	}
	return key
}

// buildAnomalyDescription 构建异常描述
func buildAnomalyDescription(sample MetricSample, anomalyType AnomalyType, deviation float64) string {
	direction := "突增"
	if anomalyType == AnomalyDrop {
		direction = "突降"
	}
	return fmt.Sprintf("指标 %s %s，偏离基线 %.1fσ，当前值 %.2f",
		sample.MetricName, direction, math.Abs(deviation), sample.Value)
}

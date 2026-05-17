package aiengine

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// DimensionDrilldown — 自动维度下钻（对标 Honeycomb BubbleUp）
// ---------------------------------------------------------------------------
//
// 核心理念：告警触发后，自动找到是哪个维度（host/service/endpoint/region）导致的
//
// 算法：
// 1. 获取聚合指标（如 avg(latency) by (service, endpoint, host)）
// 2. 对比异常时段 vs 正常时段的各维度分布
// 3. 找到贡献最大的维度值（哪个 host/endpoint 拉高了整体）
// 4. 递归下钻（service → endpoint → host → process）

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DrilldownResult 下钻分析结果
type DrilldownResult struct {
	Metric        string               `json:"metric"`
	TimeRange     TimeRange            `json:"time_range"`
	Dimensions    []DimensionBreakdown `json:"dimensions"`
	RootDimension *DimensionValue      `json:"root_dimension"` // 最终定位到的维度
}

// DimensionBreakdown 维度分解
type DimensionBreakdown struct {
	DimensionName  string           `json:"dimension_name"` // host, service, endpoint
	Values         []DimensionValue `json:"values"`
	TopContributor string           `json:"top_contributor"` // 贡献最大的值
}

// DimensionValue 维度值及其贡献
type DimensionValue struct {
	Value        string  `json:"value"`
	CurrentAvg   float64 `json:"current_avg"`
	BaselineAvg  float64 `json:"baseline_avg"`
	Contribution float64 `json:"contribution"` // 对整体异常的贡献百分比
	IsAnomalous  bool    `json:"is_anomalous"`
}

// DimensionDataPoint 维度数据点（从 Prometheus 查询获取）
type DimensionDataPoint struct {
	DimensionName  string  `json:"dimension_name"`
	DimensionValue string  `json:"dimension_value"`
	Value          float64 `json:"value"`
}

// DimensionDrilldown 自动维度下钻引擎
type DimensionDrilldown struct {
	mu sync.RWMutex
	// dimensionHierarchy 维度层级（从粗到细）
	dimensionHierarchy []string
	// anomalyThreshold 异常判定阈值（偏差百分比）
	anomalyThreshold float64
}

// NewDimensionDrilldown 创建维度下钻引擎
func NewDimensionDrilldown() *DimensionDrilldown {
	return &DimensionDrilldown{
		dimensionHierarchy: []string{"service", "endpoint", "host", "instance", "pod"},
		anomalyThreshold:   0.5, // 50% 偏差视为异常
	}
}

// Drilldown 从聚合指标自动下钻到具体维度
func (d *DimensionDrilldown) Drilldown(metric string, currentData []DimensionDataPoint, baselineData []DimensionDataPoint, tr TimeRange) (*DrilldownResult, error) {
	if len(currentData) == 0 {
		return nil, fmt.Errorf("no current data provided for drilldown")
	}

	result := &DrilldownResult{
		Metric:     metric,
		TimeRange:  tr,
		Dimensions: make([]DimensionBreakdown, 0),
	}

	// 按维度名分组
	currentByDim := groupByDimension(currentData)
	baselineByDim := groupByDimension(baselineData)

	// 对每个维度进行分析
	for _, dimName := range d.dimensionHierarchy {
		currentValues, hasCurrent := currentByDim[dimName]
		baselineValues, hasBaseline := baselineByDim[dimName]

		if !hasCurrent {
			continue
		}

		breakdown := d.analyzeDimension(dimName, currentValues, baselineValues, hasBaseline)
		result.Dimensions = append(result.Dimensions, breakdown)

		// 如果找到了明确的异常维度值，记录为根因维度
		for i := range breakdown.Values {
			if breakdown.Values[i].IsAnomalous && breakdown.Values[i].Contribution > 0.5 {
				result.RootDimension = &breakdown.Values[i]
				break
			}
		}
	}

	return result, nil
}

// BubbleUp 从底层异常冒泡到影响范围
func (d *DimensionDrilldown) BubbleUp(anomalyHost string, metric string, allData []DimensionDataPoint) (*DrilldownResult, error) {
	if anomalyHost == "" {
		return nil, fmt.Errorf("anomaly host is required for bubble up")
	}

	result := &DrilldownResult{
		Metric:     metric,
		TimeRange:  TimeRange{Start: time.Now().Add(-30 * time.Minute), End: time.Now()},
		Dimensions: make([]DimensionBreakdown, 0),
	}

	// 找到该 host 关联的所有上层维度
	affectedServices := make(map[string]bool)
	affectedEndpoints := make(map[string]bool)

	for _, dp := range allData {
		if dp.DimensionName == "host" && dp.DimensionValue == anomalyHost {
			// 标记该 host 为异常源
			continue
		}
		// 通过数据关联找到受影响的上层服务
		if dp.DimensionName == "service" {
			affectedServices[dp.DimensionValue] = true
		}
		if dp.DimensionName == "endpoint" {
			affectedEndpoints[dp.DimensionValue] = true
		}
	}

	// 构建影响范围
	hostDim := DimensionBreakdown{
		DimensionName:  "host",
		TopContributor: anomalyHost,
		Values: []DimensionValue{{
			Value:       anomalyHost,
			IsAnomalous: true,
			Contribution: 1.0,
		}},
	}
	result.Dimensions = append(result.Dimensions, hostDim)
	result.RootDimension = &hostDim.Values[0]

	return result, nil
}

// analyzeDimension 分析单个维度的异常贡献
func (d *DimensionDrilldown) analyzeDimension(dimName string, currentValues []DimensionDataPoint, baselineValues []DimensionDataPoint, hasBaseline bool) DimensionBreakdown {
	breakdown := DimensionBreakdown{
		DimensionName: dimName,
		Values:        make([]DimensionValue, 0),
	}

	// 构建基线映射
	baselineMap := make(map[string]float64)
	if hasBaseline {
		for _, bp := range baselineValues {
			baselineMap[bp.DimensionValue] = bp.Value
		}
	}

	// 计算总偏差用于贡献度计算
	totalDeviation := 0.0
	deviations := make(map[string]float64)
	for _, cp := range currentValues {
		bl := baselineMap[cp.DimensionValue]
		if bl == 0 {
			bl = cp.Value * 0.8 // 无基线时假设当前值偏高 20%
		}
		dev := math.Abs(cp.Value - bl)
		deviations[cp.DimensionValue] = dev
		totalDeviation += dev
	}

	// 构建维度值列表
	for _, cp := range currentValues {
		bl := baselineMap[cp.DimensionValue]
		if bl == 0 {
			bl = cp.Value * 0.8
		}

		contribution := 0.0
		if totalDeviation > 0 {
			contribution = deviations[cp.DimensionValue] / totalDeviation
		}

		isAnomalous := false
		if bl > 0 {
			relativeDeviation := math.Abs(cp.Value-bl) / bl
			isAnomalous = relativeDeviation > d.anomalyThreshold
		}

		breakdown.Values = append(breakdown.Values, DimensionValue{
			Value:        cp.DimensionValue,
			CurrentAvg:   cp.Value,
			BaselineAvg:  bl,
			Contribution: contribution,
			IsAnomalous:  isAnomalous,
		})
	}

	// 按贡献度排序
	sort.Slice(breakdown.Values, func(i, j int) bool {
		return breakdown.Values[i].Contribution > breakdown.Values[j].Contribution
	})

	if len(breakdown.Values) > 0 {
		breakdown.TopContributor = breakdown.Values[0].Value
	}

	return breakdown
}

// groupByDimension 按维度名分组数据点
func groupByDimension(data []DimensionDataPoint) map[string][]DimensionDataPoint {
	result := make(map[string][]DimensionDataPoint)
	for _, dp := range data {
		result[dp.DimensionName] = append(result[dp.DimensionName], dp)
	}
	return result
}

package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// CapacityForecastResult 容量预测结果。
type CapacityForecastResult struct {
	Metric                  string   `json:"metric"`
	Instance                string   `json:"instance"`
	CurrentValue            float64  `json:"current_value"`
	TrendPerDay             float64  `json:"trend_per_day"`
	PredictedExhaustionDate *string  `json:"predicted_exhaustion_date"`
	Confidence              float64  `json:"confidence"`
	DataPoints              int      `json:"data_points"`
	DaysAnalyzed            int      `json:"days_analyzed"`
}

// GetCapacityForecast 基于历史 Prometheus 数据进行简单线性回归，预测资源耗尽时间。
// GET /api/v1/capacity/forecast?metric=disk_used_percent&instance=10.0.0.1&days=30&threshold=100
func GetCapacityForecast(c *gin.Context) {
	metric := c.Query("metric")
	instance := c.Query("instance")
	daysStr := c.DefaultQuery("days", "30")
	thresholdStr := c.DefaultQuery("threshold", "100")

	if metric == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric query param is required"})
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil || threshold <= 0 {
		threshold = 100
	}

	// 从 Prometheus 查询历史数据
	dataPoints, err := queryPrometheusRange(metric, instance, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query metrics: %v", err)})
		return
	}

	if len(dataPoints) < 2 {
		c.JSON(http.StatusOK, CapacityForecastResult{
			Metric:     metric,
			Instance:   instance,
			DataPoints: len(dataPoints),
			Confidence: 0,
		})
		return
	}

	// 简单线性回归: y = a + b*x
	slope, intercept, rSquared := linearRegression(dataPoints)

	currentValue := dataPoints[len(dataPoints)-1].value
	trendPerDay := slope * 86400 // 每秒斜率转为每天

	var exhaustionDate *string
	if slope > 0 && currentValue < threshold {
		// 预测何时达到阈值
		remaining := threshold - currentValue
		secondsToExhaust := remaining / slope
		exhaustTime := time.Now().Add(time.Duration(secondsToExhaust) * time.Second)
		dateStr := exhaustTime.Format("2006-01-02")
		exhaustionDate = &dateStr
	}

	result := CapacityForecastResult{
		Metric:                  metric,
		Instance:                instance,
		CurrentValue:            math.Round(currentValue*100) / 100,
		TrendPerDay:             math.Round(trendPerDay*1000) / 1000,
		PredictedExhaustionDate: exhaustionDate,
		Confidence:              math.Round(rSquared*1000) / 1000,
		DataPoints:              len(dataPoints),
		DaysAnalyzed:            days,
	}

	c.JSON(http.StatusOK, result)

	_ = intercept
}

type dataPoint struct {
	timestamp float64
	value     float64
}

// linearRegression 执行简单线性回归，返回斜率、截距和 R^2。
func linearRegression(points []dataPoint) (slope, intercept, rSquared float64) {
	n := float64(len(points))
	if n < 2 {
		return 0, 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for _, p := range points {
		sumX += p.timestamp
		sumY += p.value
		sumXY += p.timestamp * p.value
		sumX2 += p.timestamp * p.timestamp
	}

	meanX := sumX / n
	meanY := sumY / n

	denom := sumX2 - sumX*meanX
	if denom == 0 {
		return 0, meanY, 0
	}

	slope = (sumXY - sumX*meanY) / denom
	intercept = meanY - slope*meanX

	// 计算 R^2
	var ssRes, ssTot float64
	for _, p := range points {
		predicted := intercept + slope*p.timestamp
		ssRes += (p.value - predicted) * (p.value - predicted)
		ssTot += (p.value - meanY) * (p.value - meanY)
	}
	if ssTot == 0 {
		rSquared = 1
	} else {
		rSquared = 1 - ssRes/ssTot
	}
	if rSquared < 0 {
		rSquared = 0
	}

	return slope, intercept, rSquared
}

// queryPrometheusRange 查询 Prometheus 范围数据。
// 优先通过已配置的数据源查询，如果不可用则返回空。
func queryPrometheusRange(metric, instance string, days int) ([]dataPoint, error) {
	if !store.GormOK() {
		return nil, fmt.Errorf("database not available")
	}

	// 尝试从 monitor_datasources 获取 Prometheus 地址
	type dsRow struct {
		URL string
	}
	var ds dsRow
	err := store.GetDB().Table("monitor_datasources").
		Select("url").
		Where("type = 'prometheus' AND status = 'active'").
		Order("id ASC").
		Limit(1).
		Scan(&ds).Error
	if err != nil || ds.URL == "" {
		return nil, fmt.Errorf("no active prometheus datasource configured")
	}

	// 构建 Prometheus query_range 请求
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	step := "1h"
	if days > 14 {
		step = "4h"
	}

	query := metric
	if instance != "" {
		query = fmt.Sprintf(`%s{instance="%s"}`, metric, instance)
	}

	url := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		ds.URL, query, start.Unix(), end.Unix(), step)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("prometheus request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Values [][]interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	if promResp.Status != "success" || len(promResp.Data.Result) == 0 {
		return nil, fmt.Errorf("no data returned from prometheus")
	}

	values := promResp.Data.Result[0].Values
	points := make([]dataPoint, 0, len(values))
	for _, v := range values {
		if len(v) < 2 {
			continue
		}
		ts, ok := v[0].(float64)
		if !ok {
			continue
		}
		valStr, ok := v[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		points = append(points, dataPoint{timestamp: ts, value: val})
	}

	return points, nil
}

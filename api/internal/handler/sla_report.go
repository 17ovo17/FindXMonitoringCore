package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SLAService SLA 服务目标模型。
type SLAService struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;uniqueIndex" json:"name"`
	Description   string    `gorm:"size:512" json:"description"`
	TargetPercent float64   `json:"target_percent"` // 如 99.9
	AlertRuleIDs  string    `gorm:"size:1024" json:"alert_rule_ids"` // 逗号分隔的规则 ID
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (SLAService) TableName() string { return "sla_services" }

// SLAReport SLA 报告结果。
type SLAReport struct {
	ServiceID      uint    `json:"service_id"`
	ServiceName    string  `json:"service_name"`
	TargetPercent  float64 `json:"target_percent"`
	ActualPercent  float64 `json:"actual_percent"`
	TotalMinutes   float64 `json:"total_minutes"`
	DowntimeMin    float64 `json:"downtime_minutes"`
	UptimeMin      float64 `json:"uptime_minutes"`
	IncidentCount  int     `json:"incident_count"`
	Met            bool    `json:"met"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
}

func ensureSLATable() {
	if store.GormOK() {
		store.GetDB().AutoMigrate(&SLAService{})
	}
}

// ListSLAServices 列出所有 SLA 服务目标。
// GET /api/v1/sla/services
func ListSLAServices(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureSLATable()

	var services []SLAService
	if err := store.GetDB().Order("name ASC").Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"services": services, "total": len(services)})
}

// CreateSLAService 创建 SLA 服务目标。
// POST /api/v1/sla/services
func CreateSLAService(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureSLATable()

	var svc SLAService
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if svc.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if svc.TargetPercent <= 0 || svc.TargetPercent > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_percent must be between 0 and 100"})
		return
	}

	svc.ID = 0
	svc.CreatedAt = time.Now()
	svc.UpdatedAt = time.Now()

	if err := store.GetDB().Create(&svc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, svc)
}

// GetSLAReport 计算指定服务在时间段内的 SLA。
// GET /api/v1/sla/services/:id/report?start=2026-05-01&end=2026-05-17
// SLA = (total_minutes - downtime_minutes) / total_minutes * 100
func GetSLAReport(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureSLATable()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var svc SLAService
	if err := store.GetDB().First(&svc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query params are required (YYYY-MM-DD)"})
		return
	}

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}
	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}
	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be after start"})
		return
	}

	totalMinutes := endTime.Sub(startTime).Minutes()

	// 计算宕机时间：查询该服务关联规则的 firing 事件持续时间之和
	var downtimeMinutes float64
	var incidentCount int64

	// 查询 alert_events 中 rule_name 匹配且在时间范围内的 firing 事件
	type downtimeRow struct {
		Minutes float64
	}
	var result downtimeRow

	db := store.GetDB()
	// 使用 starts_at 和 ends_at 计算每个事件的持续时间
	err = db.Table("alert_events").
		Select("COALESCE(SUM(TIMESTAMPDIFF(MINUTE, GREATEST(starts_at, ?), LEAST(COALESCE(ends_at, ?), ?))), 0) as minutes", startTime, endTime, endTime).
		Where("rule_name = ? AND starts_at < ? AND (ends_at IS NULL OR ends_at > ?)", svc.Name, endTime, startTime).
		Scan(&result).Error

	if err != nil {
		// 如果查询失败（表结构不匹配等），返回 0 宕机
		downtimeMinutes = 0
	} else {
		downtimeMinutes = result.Minutes
	}

	db.Table("alert_events").
		Where("rule_name = ? AND starts_at < ? AND (ends_at IS NULL OR ends_at > ?)", svc.Name, endTime, startTime).
		Count(&incidentCount)

	if downtimeMinutes < 0 {
		downtimeMinutes = 0
	}
	if downtimeMinutes > totalMinutes {
		downtimeMinutes = totalMinutes
	}

	uptimeMinutes := totalMinutes - downtimeMinutes
	actualPercent := 0.0
	if totalMinutes > 0 {
		actualPercent = (uptimeMinutes / totalMinutes) * 100
	}

	report := SLAReport{
		ServiceID:     uint(id),
		ServiceName:   svc.Name,
		TargetPercent: svc.TargetPercent,
		ActualPercent: actualPercent,
		TotalMinutes:  totalMinutes,
		DowntimeMin:   downtimeMinutes,
		UptimeMin:     uptimeMinutes,
		IncidentCount: int(incidentCount),
		Met:           actualPercent >= svc.TargetPercent,
		StartTime:     startStr,
		EndTime:       endStr,
	}

	c.JSON(http.StatusOK, report)
}

// RegisterSLARoutes 注册 SLA 相关路由。
func RegisterSLARoutes(v1 *gin.RouterGroup, readRequired, writeRequired gin.HandlerFunc) {
	v1.GET("/sla/services", readRequired, ListSLAServices)
	v1.POST("/sla/services", writeRequired, CreateSLAService)
	v1.GET("/sla/services/:id/report", readRequired, GetSLAReport)
}

// 辅助：确保 fmt 被使用
var _ = fmt.Sprintf

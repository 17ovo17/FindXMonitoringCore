package handler

import (
	"net/http"
	"strconv"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OncallShift 值班排班模型。
type OncallShift struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:128;index" json:"user_id"`
	UserName  string    `gorm:"size:128" json:"user_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Type      string    `gorm:"size:32" json:"type"` // primary | backup
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (OncallShift) TableName() string { return "oncall_shifts" }

func ensureOncallTable() {
	if store.GormOK() {
		store.GetDB().AutoMigrate(&OncallShift{})
	}
}

// GetOncallCalendar 返回指定月份的值班日历数据。
// GET /api/v1/oncall/calendar?month=2026-05
func GetOncallCalendar(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureOncallTable()

	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01"))
	start, err := time.Parse("2006-01", monthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
		return
	}
	end := start.AddDate(0, 1, 0)

	var shifts []OncallShift
	if err := store.GetDB().
		Where("start_time < ? AND end_time > ?", end, start).
		Order("start_time ASC").
		Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"month":  monthStr,
		"shifts": shifts,
		"total":  len(shifts),
	})
}

// CreateOncallShift 创建值班排班。
// POST /api/v1/oncall/shifts
func CreateOncallShift(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureOncallTable()

	var shift OncallShift
	if err := c.ShouldBindJSON(&shift); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if shift.UserID == "" || shift.UserName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and user_name are required"})
		return
	}
	if shift.StartTime.IsZero() || shift.EndTime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}
	if !shift.EndTime.After(shift.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}
	if shift.Type == "" {
		shift.Type = "primary"
	}
	if shift.Type != "primary" && shift.Type != "backup" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be primary or backup"})
		return
	}

	shift.ID = 0
	shift.CreatedAt = time.Now()
	shift.UpdatedAt = time.Now()

	if err := store.GetDB().Create(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shift)
}

// UpdateOncallShift 更新值班排班。
// PUT /api/v1/oncall/shifts/:id
func UpdateOncallShift(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureOncallTable()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var existing OncallShift
	if err := store.GetDB().First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "shift not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var body OncallShift
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if body.UserID != "" {
		updates["user_id"] = body.UserID
	}
	if body.UserName != "" {
		updates["user_name"] = body.UserName
	}
	if !body.StartTime.IsZero() {
		updates["start_time"] = body.StartTime
	}
	if !body.EndTime.IsZero() {
		updates["end_time"] = body.EndTime
	}
	if body.Type == "primary" || body.Type == "backup" {
		updates["type"] = body.Type
	}

	if err := store.GetDB().Model(&OncallShift{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	store.GetDB().First(&existing, id)
	c.JSON(http.StatusOK, existing)
}

// DeleteOncallShift 删除值班排班。
// DELETE /api/v1/oncall/shifts/:id
func DeleteOncallShift(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}
	ensureOncallTable()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result := store.GetDB().Delete(&OncallShift{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "shift not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RegisterOncallRoutes 注册值班日历相关路由（供 routes 文件调用）。
func RegisterOncallRoutes(v1 *gin.RouterGroup, readRequired, writeRequired gin.HandlerFunc) {
	v1.GET("/oncall/calendar", readRequired, GetOncallCalendar)
	v1.POST("/oncall/shifts", writeRequired, CreateOncallShift)
	v1.PUT("/oncall/shifts/:id", writeRequired, UpdateOncallShift)
	v1.DELETE("/oncall/shifts/:id", writeRequired, DeleteOncallShift)
}

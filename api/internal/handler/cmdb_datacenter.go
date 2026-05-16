package handler

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Datacenter 数据中心
type cmdbDatacenter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Floor     string `json:"floor"`
	CreatedAt string `json:"created_at"`
}

// Rack 机柜
type cmdbRack struct {
	ID           string `json:"id"`
	DatacenterID string `json:"datacenter_id"`
	Name         string `json:"name"`
	TotalUnits   int    `json:"total_units"`
	Row          string `json:"row"`
	Column       int    `json:"column"`
	CreatedAt    string `json:"created_at"`
}

// RackUnit 机柜U位
type cmdbRackUnit struct {
	Position   int    `json:"position"`
	ResourceID string `json:"resource_id"`
	Height     int    `json:"height"`
}

var (
	dcMu          sync.RWMutex
	datacenters   []cmdbDatacenter
	racks         []cmdbRack
	rackUnits     map[string][]cmdbRackUnit // key: rackID
)

func init() {
	rackUnits = make(map[string][]cmdbRackUnit)
}

// CmdbListDatacenters 列出所有数据中心
func CmdbListDatacenters(c *gin.Context) {
	dcMu.RLock()
	out := make([]cmdbDatacenter, len(datacenters))
	copy(out, datacenters)
	dcMu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbCreateDatacenter 创建数据中心
func CmdbCreateDatacenter(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Location string `json:"location"`
		Floor    string `json:"floor"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: name 必填"})
		return
	}

	dc := cmdbDatacenter{
		ID:        "dc-" + store.NewID(),
		Name:      req.Name,
		Location:  req.Location,
		Floor:     req.Floor,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	dcMu.Lock()
	datacenters = append(datacenters, dc)
	dcMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"dc_id": dc.ID,
		"name":  dc.Name,
	}).Info("cmdb: datacenter created")

	c.JSON(http.StatusCreated, dc)
}

// CmdbListRacks 列出数据中心下的机柜
func CmdbListRacks(c *gin.Context) {
	dcID := c.Param("id")

	dcMu.RLock()
	var out []cmdbRack
	for _, r := range racks {
		if r.DatacenterID == dcID {
			out = append(out, r)
		}
	}
	dcMu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out), "datacenter_id": dcID})
}

// CmdbCreateRack 在数据中心下创建机柜
func CmdbCreateRack(c *gin.Context) {
	dcID := c.Param("id")

	var req struct {
		Name       string `json:"name" binding:"required"`
		TotalUnits int    `json:"total_units"`
		Row        string `json:"row"`
		Column     int    `json:"column"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: name 必填"})
		return
	}

	totalUnits := req.TotalUnits
	if totalUnits <= 0 {
		totalUnits = 42
	}

	rack := cmdbRack{
		ID:           "rack-" + store.NewID(),
		DatacenterID: dcID,
		Name:         req.Name,
		TotalUnits:   totalUnits,
		Row:          req.Row,
		Column:       req.Column,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	dcMu.Lock()
	racks = append(racks, rack)
	dcMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"rack_id":       rack.ID,
		"datacenter_id": dcID,
		"name":          rack.Name,
	}).Info("cmdb: rack created")

	c.JSON(http.StatusCreated, rack)
}

// CmdbGetRackUnits 获取机柜U位布局
func CmdbGetRackUnits(c *gin.Context) {
	rackID := c.Param("rackId")

	dcMu.RLock()
	var targetRack *cmdbRack
	for i := range racks {
		if racks[i].ID == rackID {
			targetRack = &racks[i]
			break
		}
	}
	units := make([]cmdbRackUnit, len(rackUnits[rackID]))
	copy(units, rackUnits[rackID])
	dcMu.RUnlock()

	if targetRack == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "机柜不存在"})
		return
	}

	// 构建完整 U 位视图
	totalUnits := targetRack.TotalUnits
	layout := make([]gin.H, 0, totalUnits)
	occupied := make(map[int]cmdbRackUnit)
	for _, u := range units {
		for pos := u.Position; pos < u.Position+u.Height; pos++ {
			occupied[pos] = u
		}
	}

	for pos := 1; pos <= totalUnits; pos++ {
		if u, ok := occupied[pos]; ok {
			layout = append(layout, gin.H{
				"position":    pos,
				"resource_id": u.ResourceID,
				"height":      u.Height,
				"status":      "occupied",
			})
		} else {
			layout = append(layout, gin.H{
				"position": pos,
				"status":   "empty",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"rack_id":     rackID,
		"total_units": totalUnits,
		"layout":      layout,
	})
}

// CmdbAssignRackUnit 分配 U 位（辅助接口）
func CmdbAssignRackUnit(c *gin.Context) {
	rackID := c.Param("rackId")

	var req struct {
		Position   int    `json:"position" binding:"required"`
		ResourceID string `json:"resource_id" binding:"required"`
		Height     int    `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: position 和 resource_id 必填"})
		return
	}
	height := req.Height
	if height <= 0 {
		height = 1
	}

	dcMu.Lock()
	var rackExists bool
	for i := range racks {
		if racks[i].ID == rackID {
			rackExists = true
			// 检查位置是否超出范围
			if req.Position < 1 || req.Position+height-1 > racks[i].TotalUnits {
				dcMu.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "U位超出机柜范围"})
				return
			}
			break
		}
	}
	if !rackExists {
		dcMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "机柜不存在"})
		return
	}

	// 检查冲突
	for _, existing := range rackUnits[rackID] {
		existEnd := existing.Position + existing.Height - 1
		newEnd := req.Position + height - 1
		if req.Position <= existEnd && newEnd >= existing.Position {
			dcMu.Unlock()
			c.JSON(http.StatusConflict, gin.H{
				"error":             "U位冲突",
				"conflict_position": existing.Position,
				"conflict_resource": existing.ResourceID,
			})
			return
		}
	}

	unit := cmdbRackUnit{
		Position:   req.Position,
		ResourceID: req.ResourceID,
		Height:     height,
	}
	rackUnits[rackID] = append(rackUnits[rackID], unit)
	dcMu.Unlock()

	c.JSON(http.StatusCreated, unit)
}

// CmdbListModelPresets 列出预置资源模型
func CmdbListModelPresets(c *gin.Context) {
	presets := store.ListCmdbModelPresets()
	c.JSON(http.StatusOK, gin.H{"items": presets, "total": len(presets)})
}

// CmdbGetModelPreset 获取单个预置模型详情
func CmdbGetModelPreset(c *gin.Context) {
	preset, ok := store.GetCmdbModelPreset(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "预置模型不存在"})
		return
	}
	c.JSON(http.StatusOK, preset)
}

// CmdbApplyModelPreset 应用预置模型到 CMDB
func CmdbApplyModelPreset(c *gin.Context) {
	presetID := c.Param("id")
	obj, err := store.ApplyCmdbModelPreset(presetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "应用预置模型失败"})
		return
	}
	if obj == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预置模型不存在"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":   "预置模型已应用",
		"object_id": obj.ID,
		"name":      obj.Name,
	})
}

// cmdbPageAndLimitFromQuery 从查询参数获取分页信息（避免与 cmdb_compatible.go 冲突）
func cmdbPageAndLimitFromQuery(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

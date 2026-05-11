package handler

import (
	"net/http"
	"strconv"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// CmdbTree 返回分类树 + 模型列表 + 实例计数
func CmdbTree(c *gin.Context) {
	categories := store.ListCmdbCategories()
	allObjects := store.ListCmdbObjects("")
	counts := store.CountCmdbInstancesByObject()

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"objects":    allObjects,
		"counts":    counts,
	})
}

// ListCmdbObjects 按分类查询模型列表
func ListCmdbObjects(c *gin.Context) {
	categoryID := c.Query("category_id")
	objects := store.ListCmdbObjects(categoryID)
	c.JSON(http.StatusOK, objects)
}

// CreateCmdbObject 创建模型
func CreateCmdbObject(c *gin.Context) {
	var obj model.CmdbObject
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if err := store.CreateCmdbObject(&obj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模型失败"})
		return
	}
	c.JSON(http.StatusOK, obj)
}

// GetCmdbObject 获取单个模型详情
func GetCmdbObject(c *gin.Context) {
	id := c.Param("id")
	obj, ok := store.GetCmdbObject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}
	c.JSON(http.StatusOK, obj)
}

// UpdateCmdbObject 更新模型
func UpdateCmdbObject(c *gin.Context) {
	id := c.Param("id")
	existing, ok := store.GetCmdbObject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	existing.ID = id
	if err := store.UpdateCmdbObject(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模型失败"})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteCmdbObject 删除模型
func DeleteCmdbObject(c *gin.Context) {
	if err := store.DeleteCmdbObject(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除模型失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ListCmdbAttributes 查询模型属性列表
func ListCmdbAttributes(c *gin.Context) {
	objectID := c.Param("id")
	attrs := store.ListCmdbAttributes(objectID)
	c.JSON(http.StatusOK, attrs)
}

// CreateCmdbAttribute 创建模型属性
func CreateCmdbAttribute(c *gin.Context) {
	var attr model.CmdbAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	attr.ObjectID = c.Param("id")
	if err := store.CreateCmdbAttribute(&attr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建属性失败"})
		return
	}
	c.JSON(http.StatusOK, attr)
}

// UpdateCmdbAttribute 更新模型属性
func UpdateCmdbAttribute(c *gin.Context) {
	var attr model.CmdbAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	attr.ID = c.Param("id")
	if err := store.UpdateCmdbAttribute(&attr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新属性失败"})
		return
	}
	c.JSON(http.StatusOK, attr)
}

// DeleteCmdbAttribute 删除模型属性
func DeleteCmdbAttribute(c *gin.Context) {
	if err := store.DeleteCmdbAttribute(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除属性失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ListCmdbInstances 分页查询实例列表
func ListCmdbInstances(c *gin.Context) {
	objectID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, total := store.ListCmdbInstances(objectID, page, limit)
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

// CreateCmdbInstance 创建实例
func CreateCmdbInstance(c *gin.Context) {
	var inst model.CmdbInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	inst.ObjectID = c.Param("id")
	if err := store.CreateCmdbInstance(&inst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建实例失败"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// GetCmdbInstance 获取单个实例
func GetCmdbInstance(c *gin.Context) {
	inst, ok := store.GetCmdbInstance(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// UpdateCmdbInstance 更新实例
func UpdateCmdbInstance(c *gin.Context) {
	id := c.Param("id")
	existing, ok := store.GetCmdbInstance(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	existing.ID = id
	if err := store.UpdateCmdbInstance(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新实例失败"})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteCmdbInstance 删除实例
func DeleteCmdbInstance(c *gin.Context) {
	if err := store.DeleteCmdbInstance(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除实例失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

package handler

import (
	"errors"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const contractMatrixBlockedMessage = "能力缺少执行器或数据源契约，已阻断"

func ListContractMatrixEntries(c *gin.Context) {
	items, err := store.ListContractMatrixEntries(c.Query("status"), c.Query("domain"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract matrix filter"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func GetContractMatrixEntry(c *gin.Context) {
	item, ok, err := store.GetContractMatrixEntry(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract gap id"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract gap not found"})
		return
	}
	if item.Status != model.ContractStatusReady {
		c.JSON(http.StatusConflict, ContractMatrixBlockedResponse(item))
		return
	}
	c.JSON(http.StatusOK, item)
}

func RegisterContractMatrixEntry(c *gin.Context) {
	var input model.ContractMatrixRegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ContractMatrixBlockedResponse(model.ContractMatrixEntry{
			ID:     "FX-CONTRACT-INVALID-PAYLOAD",
			Status: model.ContractStatusUnsafe,
		}))
		return
	}
	item, err := store.SaveContractMatrixEntry(input)
	if err != nil {
		status := contractMatrixValidationStatus(input.Status)
		response := ContractMatrixBlockedResponse(model.ContractMatrixEntry{
			ID:     "FX-CONTRACT-INVALID-ENTRY",
			Status: status,
		})
		if !errors.Is(err, store.ErrContractMatrixValidation) {
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if item.Status != model.ContractStatusReady {
		c.JSON(http.StatusConflict, ContractMatrixBlockedResponse(item))
		return
	}
	c.JSON(http.StatusOK, item)
}

func contractMatrixValidationStatus(status string) string {
	switch strings.TrimSpace(status) {
	case model.ContractStatusBlocked,
		model.ContractStatusMissingBackend,
		model.ContractStatusMissingDatasource,
		model.ContractStatusMissingExecutor,
		model.ContractStatusUnsafe:
		return strings.TrimSpace(status)
	default:
		return model.ContractStatusUnsafe
	}
}

func ContractMatrixBlockedResponse(item model.ContractMatrixEntry) model.ContractMatrixBlockedResponse {
	return model.ContractMatrixBlockedResponse{
		Code:          model.ContractBlockedByContractCode,
		Message:       contractMatrixBlockedMessage,
		ContractGapID: item.ID,
		Status:        item.Status,
		SafeToRetry:   false,
	}
}

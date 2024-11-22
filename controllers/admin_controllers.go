package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/services"
	"net/http"
)

type LogController struct {
	LogService *services.LogService
}

type MonitoringController struct {
	MonitoringService *services.MonitoringService
}

func NewLogController(logService *services.LogService) *LogController {
	return &LogController{LogService: logService}
}

func NewMonitoringController(monitoringService *services.MonitoringService) *MonitoringController {
	return &MonitoringController{MonitoringService: monitoringService}
}

// CreateLogHandler creates a log entry (useful for testing or manual entries)
func (lc *LogController) CreateLogHandler(c *gin.Context) {
	var logRequest struct {
		Level   string `json:"level" binding:"required"`
		Action  string `json:"action" binding:"required"`
		UserID  string `json:"userId" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&logRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create log entry
	err := lc.LogService.CreateLog(logRequest.Level, logRequest.Action, logRequest.UserID, logRequest.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Log created successfully!"})
}

// FetchLogsHandler retrieves logs based on filters (level, action, etc.)
func (lc *LogController) FetchLogsHandler(c *gin.Context) {
	// Extract query parameters (e.g., level, action)
	filter := map[string]interface{}{}
	if level := c.Query("level"); level != "" {
		filter["level"] = level
	}
	if action := c.Query("action"); action != "" {
		filter["action"] = action
	}

	// Fetch logs
	log, err := lc.LogService.FetchLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": log})
}

// FetchMonitoringDataHandler retrieves aggregated monitoring data
func (mc *MonitoringController) FetchMonitoringDataHandler(c *gin.Context) {
	data, err := mc.MonitoringService.FetchMonitoringData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monitoringData": data})
}

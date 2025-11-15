package handlers

import (
	"github.com/gin-gonic/gin"
)

// AnalyzeAnomalies runs anomaly detection for an assignment
// POST /analyze/:assignment_id
func AnalyzeAnomalies(c *gin.Context) {
	AnalyzeAssignmentImpl(c)
}

// GetAnomalies retrieves anomaly findings for an assignment
// GET /anomalies/:assignment_id
func GetAnomalies(c *gin.Context) {
	GetAnomaliesImpl(c)
}

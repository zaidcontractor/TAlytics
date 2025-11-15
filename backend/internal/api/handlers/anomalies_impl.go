package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
	"talytics/internal/services"
)

var AnomalyService *services.AnomalyService

// InitAnomalyService initializes the anomaly detection service
func InitAnomalyService(svc *services.AnomalyService) {
	AnomalyService = svc
}

// AnalyzeAssignmentImpl runs anomaly detection on an assignment
func AnalyzeAssignmentImpl(c *gin.Context) {
	// Get assignment_id from URL param
	assignmentIDStr := c.Param("assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Verify user is professor or head_ta
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role := userRole.(string)
	if role != "professor" && role != "head_ta" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only professors and head TAs can run anomaly detection"})
		return
	}

	// Verify assignment exists
	exists, err = database.GetAssignmentByID(assignmentID)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	// Check if assignment has grades
	grades, err := database.GetAllGradesForAssignment(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve grades: " + err.Error()})
		return
	}

	if len(grades) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assignment has no grades yet. Please grade submissions first."})
		return
	}

	// Minimum grades requirement for statistical analysis
	if len(grades) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Need at least 5 grades for meaningful anomaly detection"})
		return
	}

	// Generate anomaly report
	reportJSON, err := AnomalyService.GenerateAnomalyReport(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate anomaly report: " + err.Error()})
		return
	}

	// Save report to database
	reportID, err := database.SaveAnomalyReport(assignmentID, reportJSON, "pending")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save anomaly report: " + err.Error()})
		return
	}

	// Parse JSON back to struct for response
	var summary models.AnomalySummary
	if err := json.Unmarshal([]byte(reportJSON), &summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse report: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.AnalyzeResponse{
		Message:  "Anomaly analysis completed successfully",
		ReportID: reportID,
		Summary:  summary,
	})
}

// GetAnomaliesImpl retrieves anomaly report for assignment
func GetAnomaliesImpl(c *gin.Context) {
	// Get assignment_id from URL param
	assignmentIDStr := c.Param("assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Verify user is professor or head_ta
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role := userRole.(string)
	if role != "professor" && role != "head_ta" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only professors and head TAs can view anomaly reports"})
		return
	}

	// Get anomaly report from database
	report, err := database.GetAnomalyReport(assignmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No anomaly report found for this assignment"})
		return
	}

	// Parse JSON blob
	jsonBlob := report["json_blob"].(string)
	status := report["status"].(string)

	var summary models.AnomalySummary
	if err := json.Unmarshal([]byte(jsonBlob), &summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse report: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.GetAnomaliesResponse{
		Report: summary,
		Status: status,
	})
}

// UpdateAnomalyStatusImpl updates anomaly report status
func UpdateAnomalyStatusImpl(c *gin.Context) {
	// Get report_id from URL param
	reportIDStr := c.Param("report_id")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	// Parse request body
	var req models.UpdateAnomalyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify user is professor (only professors can update status)
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role := userRole.(string)
	if role != "professor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only professors can update anomaly report status"})
		return
	}

	// Update status in database
	if err := database.UpdateAnomalyReportStatus(reportID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Report status updated successfully",
		"report_id": reportID,
		"status":    req.Status,
	})
}

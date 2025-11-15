package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
	"talytics/internal/services"
)

var (
	ClaudeService *services.ClaudeService
	PDFService    *services.PDFService
)

// InitServices initializes handler services
func InitServices(claude *services.ClaudeService, pdf *services.PDFService) {
	ClaudeService = claude
	PDFService = pdf
}

// CreateRubric saves a rubric created via UI builder
// POST /rubrics
func CreateRubric(c *gin.Context) {
	var req models.CreateRubricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate assignment exists
	exists, err := database.GetAssignmentByID(req.AssignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate assignment"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	// Check if rubric already exists
	rubricExists, err := database.RubricExists(req.AssignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rubric existence"})
		return
	}
	if rubricExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Rubric already exists for this assignment"})
		return
	}

	// Insert rubric into database
	rubricID, err := database.InsertRubric(req.AssignmentID, req.JSONBlob, req.MaxPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create rubric: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Rubric created successfully",
		"rubric_id":     rubricID,
		"assignment_id": req.AssignmentID,
		"max_points":    req.MaxPoints,
	})
}

// UploadRubricPDF handles PDF upload and parsing
// POST /rubrics/upload
func UploadRubricPDF(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get assignment_id
	assignmentIDStr := c.Request.FormValue("assignment_id")
	if assignmentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignment_id is required"})
		return
	}

	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment_id"})
		return
	}

	// Validate assignment exists
	exists, err := database.GetAssignmentByID(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate assignment"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	// Check if rubric already exists
	rubricExists, err := database.RubricExists(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rubric existence"})
		return
	}
	if rubricExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Rubric already exists for this assignment"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Validate file type
	if header.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be a PDF"})
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Save PDF file
	filePath, err := PDFService.SaveUploadedPDF(fileData, header.Filename, "rubric")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save PDF: %v", err)})
		return
	}

	// Extract text from PDF
	extractedText, err := PDFService.ExtractTextFromPDF(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to extract text from PDF: %v", err)})
		return
	}

	if extractedText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF appears to be empty or text extraction failed"})
		return
	}

	// Parse rubric using Claude
	parsedJSON, err := ClaudeService.ParseRubricFromText(extractedText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse rubric with Claude: %v", err)})
		return
	}

	// Extract max_points from parsed JSON
	var rubricData map[string]interface{}
	if err := json.Unmarshal([]byte(parsedJSON), &rubricData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse rubric JSON"})
		return
	}

	maxPoints, ok := rubricData["max_points"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid rubric structure: missing max_points"})
		return
	}

	// Insert rubric into database
	rubricID, err := database.InsertRubric(assignmentID, parsedJSON, maxPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save rubric: %v", err)})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, models.UploadRubricPDFResponse{
		RubricID:     int(rubricID),
		AssignmentID: assignmentID,
		ParsedRubric: parsedJSON,
		MaxPoints:    maxPoints,
		FilePath:     filePath,
	})
}

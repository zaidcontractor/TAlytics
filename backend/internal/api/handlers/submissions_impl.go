package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
)

// UploadSubmissionImpl handles student submission upload
func UploadSubmissionImpl(c *gin.Context) {
	// Try JSON first
	var req models.UploadSubmissionRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		// JSON request
		submissionID, err := database.CreateSubmission(req.AssignmentID, req.StudentIdentifier, req.Text, req.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":       "Submission created successfully",
			"submission_id": submissionID,
		})
		return
	}

	// Try multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err == nil {
		assignmentIDStr := c.Request.FormValue("assignment_id")
		studentIdentifier := c.Request.FormValue("student_identifier")

		if assignmentIDStr == "" || studentIdentifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "assignment_id and student_identifier are required"})
			return
		}

		var assignmentID int
		if _, err := fmt.Sscanf(assignmentIDStr, "%d", &assignmentID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment_id"})
			return
		}

		// Check for file upload
		file, header, err := c.Request.FormFile("file")
		var text, filePath string

		if err == nil {
			defer file.Close()

			// Read file data
			fileData, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
				return
			}

			// Save PDF file if it's a PDF
			if header.Header.Get("Content-Type") == "application/pdf" {
				savedPath, err := PDFService.SaveUploadedPDF(fileData, header.Filename, "submission")
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
					return
				}
				filePath = savedPath

				// Extract text from PDF
				extractedText, err := PDFService.ExtractTextFromPDF(filePath)
				if err == nil {
					text = extractedText
				}
			} else {
				// For other file types, just save path
				filePath = header.Filename
			}
		} else {
			// No file, check for text field
			text = c.Request.FormValue("text")
		}

		submissionID, err := database.CreateSubmission(assignmentID, studentIdentifier, text, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":       "Submission created successfully",
			"submission_id": submissionID,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
}

// GetAssignedSubmissionsImpl retrieves TA's assigned submissions
func GetAssignedSubmissionsImpl(c *gin.Context) {
	// Extract TA user ID from context
	userID, _ := c.Get("user_id")

	// Get submissions
	submissions, err := database.GetSubmissionsByTA(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": submissions,
	})
}

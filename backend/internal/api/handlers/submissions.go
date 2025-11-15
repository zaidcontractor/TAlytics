package handlers

import (
	"github.com/gin-gonic/gin"
)

// UploadSubmission handles student submission upload
// POST /submissions/upload
func UploadSubmission(c *gin.Context) {
	UploadSubmissionImpl(c)
}

// GetAssignedSubmissions retrieves submissions assigned to a TA
// GET /submissions/assigned
func GetAssignedSubmissions(c *gin.Context) {
	GetAssignedSubmissionsImpl(c)
}

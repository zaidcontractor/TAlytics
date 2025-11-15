package handlers

import (
	"github.com/gin-gonic/gin"
)

// GradeSubmission handles single submission grading with Claude assistance
// POST /grade
func GradeSubmission(c *gin.Context) {
	GradeSubmissionImpl(c)
}

// BatchGrade handles batch grading for multiple submissions
// POST /grade/batch
func BatchGrade(c *gin.Context) {
	BatchGradeImpl(c)
}

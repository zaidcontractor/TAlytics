package handlers

import (
	"github.com/gin-gonic/gin"
)

// CreateAssignment handles assignment creation
// POST /assignments
func CreateAssignment(c *gin.Context) {
	CreateAssignmentImpl(c)
}

// GetAssignmentsByCourse retrieves all assignments for a course
// GET /assignments/course/:course_id
func GetAssignmentsByCourse(c *gin.Context) {
	GetAssignmentsByCourseImpl(c)
}

// GetAssignment retrieves assignment details
// GET /assignments/:id
func GetAssignment(c *gin.Context) {
	GetAssignmentImpl(c)
}

// NotifyTAs opens assignment for TA grading
// POST /assignments/:id/notify-tas
func NotifyTAs(c *gin.Context) {
	NotifyTAsImpl(c)
}

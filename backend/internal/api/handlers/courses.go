package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
)

// CreateCourse handles course creation
// POST /courses
func CreateCourse(c *gin.Context) {
	var req models.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract user ID from context (set by AuthMiddleware)
	userID, _ := c.Get("user_id")

	// Create course
	courseID, err := database.CreateCourse(req.Name, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Course created successfully",
		"course_id": courseID,
		"name":      req.Name,
	})
}

// GetCourses retrieves courses for the authenticated user
// GET /courses
func GetCourses(c *gin.Context) {
	// Extract user ID from context
	userID, _ := c.Get("user_id")

	// Get courses
	courses, err := database.GetCoursesByUser(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve courses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"courses": courses,
	})
}

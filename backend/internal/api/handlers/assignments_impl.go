package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
)

// CreateAssignmentImpl handles assignment creation
func CreateAssignmentImpl(c *gin.Context) {
	var req models.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract user ID from context
	userID, _ := c.Get("user_id")

	// Validate user owns the course
	owns, err := database.UserOwnsCourse(userID.(int), req.CourseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate course ownership"})
		return
	}
	if !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to create assignments for this course"})
		return
	}

	// Create assignment
	assignmentID, err := database.CreateAssignment(req.CourseID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Assignment created successfully",
		"assignment_id": assignmentID,
		"course_id":     req.CourseID,
		"title":         req.Title,
		"status":        "draft",
	})
}

// GetAssignmentsByCourseImpl retrieves all assignments for a course
func GetAssignmentsByCourseImpl(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("course_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Extract user ID from context
	userID, _ := c.Get("user_id")

	// Validate user owns the course
	owns, err := database.UserOwnsCourse(userID.(int), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate course ownership"})
		return
	}
	if !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view assignments for this course"})
		return
	}

	// Get assignments
	assignments, err := database.GetAssignmentsByCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": assignments,
	})
}

// GetAssignmentImpl retrieves assignment details
func GetAssignmentImpl(c *gin.Context) {
	assignmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Get assignment with rubric
	assignment, err := database.GetAssignmentWithRubric(assignmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}

// NotifyTAsImpl distributes submissions to TAs
func NotifyTAsImpl(c *gin.Context) {
	assignmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	// Extract user ID and role from context
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// Get assignment to check course ownership
	assignment, err := database.GetAssignmentWithRubric(assignmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	courseID := int(assignment["course_id"].(int64))

	// Check if user owns the course (professor/head_ta only)
	owns, err := database.UserOwnsCourse(userID.(int), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate course ownership"})
		return
	}

	if !owns && userRole != "professor" && userRole != "head_ta" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only professors and head TAs can distribute submissions"})
		return
	}

	// Get all submissions for this assignment
	submissions, err := database.GetSubmissionsByAssignment(assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get submissions"})
		return
	}

	if len(submissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No submissions found for this assignment"})
		return
	}

	// Get all TAs
	taIDs, err := database.GetTAsForCourse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get TAs"})
		return
	}

	if len(taIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No TAs available to assign submissions"})
		return
	}

	// Distribute submissions using round-robin
	distribution := make(map[int]int)
	for i, submission := range submissions {
		taIndex := i % len(taIDs)
		taID := taIDs[taIndex]
		submissionID := submission["id"].(int)

		err := database.AssignSubmissionToTA(submissionID, taID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to assign submission %d", submissionID)})
			return
		}

		distribution[taID]++
	}

	// Update assignment status to "grading"
	err = database.UpdateAssignmentStatus(assignmentID, "grading")
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to update assignment status: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"assignment_id":     assignmentID,
		"total_submissions": len(submissions),
		"distribution":      distribution,
		"message":           fmt.Sprintf("Assigned %d submissions to %d TAs", len(submissions), len(taIDs)),
	})
}

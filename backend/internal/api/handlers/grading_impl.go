package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"talytics/internal/database"
	"talytics/internal/models"
)

// GradeSubmissionImpl handles individual submission grading with Claude assistance
func GradeSubmissionImpl(c *gin.Context) {
	var req models.GradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	taID := userID.(int)

	// Get submission details with rubric
	submission, err := database.GetSubmissionDetails(req.SubmissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Verify TA is assigned to this submission (or is professor/head_ta)
	userRole, _ := c.Get("user_role")
	role := userRole.(string)

	if role == "grader_ta" {
		assignedTAID, hasTA := submission["assigned_ta_id"]
		if !hasTA || int(assignedTAID.(int64)) != taID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not assigned to this submission"})
			return
		}
	}

	// Check if submission is already graded
	existingGrade, err := database.GetGradeBySubmission(req.SubmissionID)
	if err == nil && existingGrade != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Submission already graded"})
		return
	}

	// Get rubric JSON from submission details
	rubricJSON, hasRubric := submission["rubric_json"]
	if !hasRubric {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No rubric found for this assignment"})
		return
	}

	submissionText := submission["text"].(string)
	rubricStr := rubricJSON.(string)

	// If TA wants Claude's recommendation
	if req.UseClaudeRecommendation {
		// Call Claude to get grading recommendation
		claudeResponse, err := ClaudeService.GenerateGradingRecommendation(submissionText, rubricStr)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to get grading recommendation from Claude: " + err.Error()})
			return
		}

		// Parse Claude's JSON response
		var recommendation models.GradingRecommendation
		if err := json.Unmarshal([]byte(claudeResponse), &recommendation); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Claude response: " + err.Error()})
			return
		}

		// Convert rubric breakdown to JSON string
		breakdownBytes, _ := json.Marshal(recommendation.RubricBreakdown)
		breakdownStr := string(breakdownBytes)

		// Save grade to database
		gradeID, err := database.SaveGrade(
			req.SubmissionID,
			taID,
			recommendation.Score,
			recommendation.Feedback,
			breakdownStr,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save grade: " + err.Error()})
			return
		}

		// Update submission status to graded
		if err := database.UpdateSubmissionGradingStatus(req.SubmissionID, "graded"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update submission status: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, models.GradeResponse{
			Message:              "Submission graded successfully with Claude's recommendation",
			GradeID:              gradeID,
			ClaudeRecommendation: &recommendation,
		})
		return
	}

	// If TA provided their own score/feedback (override)
	if req.Score > 0 {
		// Save TA's custom grade
		gradeID, err := database.SaveGrade(
			req.SubmissionID,
			taID,
			req.Score,
			req.Feedback,
			req.RubricBreakdown,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save grade: " + err.Error()})
			return
		}

		// Update submission status to graded
		if err := database.UpdateSubmissionGradingStatus(req.SubmissionID, "graded"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update submission status: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, models.GradeResponse{
			Message: "Submission graded successfully with custom score",
			GradeID: gradeID,
		})
		return
	}

	// If neither Claude nor custom score provided, just return Claude's recommendation without saving
	claudeResponse, err := ClaudeService.GenerateGradingRecommendation(submissionText, rubricStr)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to get grading recommendation: " + err.Error()})
		return
	}

	var recommendation models.GradingRecommendation
	if err := json.Unmarshal([]byte(claudeResponse), &recommendation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Claude response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.GradeResponse{
		Message:              "Grading recommendation generated (not saved)",
		ClaudeRecommendation: &recommendation,
	})
}

// BatchGradeImpl handles batch grading for an assignment
func BatchGradeImpl(c *gin.Context) {
	var req models.BatchGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	taID := userID.(int)

	// Verify user is professor or head_ta
	userRole, _ := c.Get("user_role")
	role := userRole.(string)
	if role != "professor" && role != "head_ta" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only professors and head TAs can batch grade"})
		return
	}

	// Verify assignment exists and has rubric
	assignmentData, err := database.GetAssignmentWithRubric(req.AssignmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
		return
	}

	rubricData, hasRubric := assignmentData["rubric"]
	if !hasRubric {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assignment must have a rubric before batch grading"})
		return
	}

	rubricMap := rubricData.(map[string]interface{})
	rubricJSON := rubricMap["json_blob"].(string)

	// Get all ungraded submissions for this assignment
	submissions, err := database.GetUngradedSubmissionsForAssignment(req.AssignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ungraded submissions: " + err.Error()})
		return
	}

	if len(submissions) == 0 {
		c.JSON(http.StatusOK, models.BatchGradeResponse{
			AssignmentID: req.AssignmentID,
			TotalGraded:  0,
			Message:      "No ungraded submissions found",
		})
		return
	}

	// Grade each submission
	var gradedIDs []int
	var totalScore float64
	var gradingErrors []string

	for _, submission := range submissions {
		submissionID := submission["id"].(int)
		submissionText := submission["text"].(string)

		// Call Claude for grading recommendation
		claudeResponse, err := ClaudeService.GenerateGradingRecommendation(submissionText, rubricJSON)
		if err != nil {
			gradingErrors = append(gradingErrors, fmt.Sprintf("Submission %d: %s", submissionID, err.Error()))
			continue
		}

		// Parse Claude's response
		var recommendation models.GradingRecommendation
		if err := json.Unmarshal([]byte(claudeResponse), &recommendation); err != nil {
			gradingErrors = append(gradingErrors, fmt.Sprintf("Submission %d: Failed to parse response", submissionID))
			continue
		}

		// If auto-approve, save the grade
		if req.AutoApprove {
			breakdownBytes, _ := json.Marshal(recommendation.RubricBreakdown)
			breakdownStr := string(breakdownBytes)

			_, err := database.SaveGrade(
				submissionID,
				taID,
				recommendation.Score,
				recommendation.Feedback,
				breakdownStr,
			)
			if err != nil {
				gradingErrors = append(gradingErrors, fmt.Sprintf("Submission %d: Failed to save grade", submissionID))
				continue
			}

			// Update submission status
			database.UpdateSubmissionGradingStatus(submissionID, "graded")

			gradedIDs = append(gradedIDs, submissionID)
			totalScore += recommendation.Score
		}
	}

	averageScore := 0.0
	if len(gradedIDs) > 0 {
		averageScore = totalScore / float64(len(gradedIDs))
	}

	message := fmt.Sprintf("Successfully graded %d submissions", len(gradedIDs))
	if len(gradingErrors) > 0 {
		message += fmt.Sprintf(" (%d errors)", len(gradingErrors))
	}

	c.JSON(http.StatusOK, models.BatchGradeResponse{
		AssignmentID:      req.AssignmentID,
		TotalGraded:       len(gradedIDs),
		AverageScore:      averageScore,
		GradedSubmissions: gradedIDs,
		Message:           message,
	})
}

// GetSubmissionGradeImpl retrieves the grade for a specific submission
func GetSubmissionGradeImpl(c *gin.Context) {
	submissionIDStr := c.Param("submission_id")
	submissionID, err := strconv.Atoi(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	grade, err := database.GetGradeBySubmission(submissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grade not found"})
		return
	}

	c.JSON(http.StatusOK, grade)
}

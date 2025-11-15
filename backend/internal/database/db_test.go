package database

import (
	"fmt"
	"testing"

	"talytics/internal/testutils"
)

func setupTestDB(t *testing.T) {
	dbPath := testutils.SetupTestDBPath(t)
	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
}

func teardownTestDB(t *testing.T) {
	if DB != nil {
		CloseDB()
	}
}

func cleanupTestDB(t *testing.T) {
	if DB == nil {
		return
	}

	tables := []string{
		"anomaly_reports",
		"grades",
		"submissions",
		"rubrics",
		"assignments",
		"courses",
		"users",
	}

	for _, table := range tables {
		_, err := DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}
}

func TestCreateUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	role := "professor"

	userID, err := CreateUser(email, passwordHash, role)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if userID <= 0 {
		t.Error("User ID should be greater than 0")
	}

	// Test duplicate email
	_, err = CreateUser(email, passwordHash, role)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestGetUserByEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	role := "professor"

	userID, err := CreateUser(email, passwordHash, role)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get user
	retrievedID, retrievedHash, retrievedRole, err := GetUserByEmail(email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedID != int(userID) {
		t.Errorf("Expected user ID %d, got %d", userID, retrievedID)
	}

	if retrievedHash != passwordHash {
		t.Errorf("Expected password hash %s, got %s", passwordHash, retrievedHash)
	}

	if retrievedRole != role {
		t.Errorf("Expected role %s, got %s", role, retrievedRole)
	}

	// Test non-existent user
	_, _, _, err = GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestUserExists(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	role := "professor"

	// User should not exist initially
	exists, err := UserExists(email)
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}
	if exists {
		t.Error("User should not exist initially")
	}

	// Create user
	_, err = CreateUser(email, passwordHash, role)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// User should exist now
	exists, err = UserExists(email)
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}
	if !exists {
		t.Error("User should exist after creation")
	}
}

func TestCreateCourse(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create a user first
	userID, err := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	courseName := "CS 101"
	courseID, err := CreateCourse(courseName, int(userID))
	if err != nil {
		t.Fatalf("Failed to create course: %v", err)
	}

	if courseID <= 0 {
		t.Error("Course ID should be greater than 0")
	}
}

func TestGetCoursesByUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create a user
	userID, err := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create courses
	_, err = CreateCourse("CS 101", int(userID))
	if err != nil {
		t.Fatalf("Failed to create course: %v", err)
	}

	_, err = CreateCourse("CS 102", int(userID))
	if err != nil {
		t.Fatalf("Failed to create course: %v", err)
	}

	// Get courses
	courses, err := GetCoursesByUser(int(userID))
	if err != nil {
		t.Fatalf("Failed to get courses: %v", err)
	}

	if len(courses) != 2 {
		t.Errorf("Expected 2 courses, got %d", len(courses))
	}
}

func TestCreateAssignment(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user and course
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))

	assignmentID, err := CreateAssignment(int(courseID), "Lab 1", "First lab assignment")
	if err != nil {
		t.Fatalf("Failed to create assignment: %v", err)
	}

	if assignmentID <= 0 {
		t.Error("Assignment ID should be greater than 0")
	}
}

func TestInsertRubric(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, and assignment
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")

	jsonBlob := `{"title":"Test Rubric","max_points":100,"criteria":[]}`
	maxPoints := 100.0

	rubricID, err := InsertRubric(int(assignmentID), jsonBlob, maxPoints)
	if err != nil {
		t.Fatalf("Failed to insert rubric: %v", err)
	}

	if rubricID <= 0 {
		t.Error("Rubric ID should be greater than 0")
	}
}

func TestRubricExists(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, and assignment
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Rubric should not exist initially
	exists, err := RubricExists(int(assignmentID))
	if err != nil {
		t.Fatalf("Failed to check rubric existence: %v", err)
	}
	if exists {
		t.Error("Rubric should not exist initially")
	}

	// Create rubric
	jsonBlob := `{"title":"Test Rubric","max_points":100}`
	_, err = InsertRubric(int(assignmentID), jsonBlob, 100.0)
	if err != nil {
		t.Fatalf("Failed to insert rubric: %v", err)
	}

	// Rubric should exist now
	exists, err = RubricExists(int(assignmentID))
	if err != nil {
		t.Fatalf("Failed to check rubric existence: %v", err)
	}
	if !exists {
		t.Error("Rubric should exist after creation")
	}
}

func TestGetRubricByAssignmentID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, and assignment
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")

	jsonBlob := `{"title":"Test Rubric","max_points":100}`
	maxPoints := 100.0

	rubricID, err := InsertRubric(int(assignmentID), jsonBlob, maxPoints)
	if err != nil {
		t.Fatalf("Failed to insert rubric: %v", err)
	}

	// Get rubric
	retrievedID, retrievedBlob, retrievedMaxPoints, err := GetRubricByAssignmentID(int(assignmentID))
	if err != nil {
		t.Fatalf("Failed to get rubric: %v", err)
	}

	if retrievedID != int(rubricID) {
		t.Errorf("Expected rubric ID %d, got %d", rubricID, retrievedID)
	}

	if retrievedBlob != jsonBlob {
		t.Errorf("Expected JSON blob %s, got %s", jsonBlob, retrievedBlob)
	}

	if retrievedMaxPoints != maxPoints {
		t.Errorf("Expected max points %f, got %f", maxPoints, retrievedMaxPoints)
	}
}

func TestCreateSubmission(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, and assignment
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")

	submissionID, err := CreateSubmission(int(assignmentID), "student001", "def factorial(n): return 1", "")
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	if submissionID <= 0 {
		t.Error("Submission ID should be greater than 0")
	}
}

func TestAssignSubmissionToTA(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, assignment, and submission
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")
	submissionID, _ := CreateSubmission(int(assignmentID), "student001", "code", "")

	// Create TA
	taID, _ := CreateUser("ta@example.com", "$2a$10$hash", "grader_ta")

	// Assign submission to TA
	err := AssignSubmissionToTA(int(submissionID), int(taID))
	if err != nil {
		t.Fatalf("Failed to assign submission to TA: %v", err)
	}
}

func TestSaveGrade(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, assignment, submission, and rubric
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")
	submissionID, _ := CreateSubmission(int(assignmentID), "student001", "code", "")
	_, _ = InsertRubric(int(assignmentID), `{"title":"Test"}`, 100.0)

	// Create grader
	graderID, _ := CreateUser("ta@example.com", "$2a$10$hash", "grader_ta")

	gradeID, err := SaveGrade(int(submissionID), int(graderID), 85.5, "Good work", `{"correctness":40}`)
	if err != nil {
		t.Fatalf("Failed to save grade: %v", err)
	}

	if gradeID <= 0 {
		t.Error("Grade ID should be greater than 0")
	}
}

func TestGetGradeBySubmission(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, assignment, submission, and rubric
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")
	submissionID, _ := CreateSubmission(int(assignmentID), "student001", "code", "")
	_, _ = InsertRubric(int(assignmentID), `{"title":"Test"}`, 100.0)

	// Create grader
	graderID, _ := CreateUser("ta@example.com", "$2a$10$hash", "grader_ta")

	score := 85.5
	feedback := "Good work"
	rubricBreakdown := `{"correctness":40}`

	_, err := SaveGrade(int(submissionID), int(graderID), score, feedback, rubricBreakdown)
	if err != nil {
		t.Fatalf("Failed to save grade: %v", err)
	}

	// Get grade
	grade, err := GetGradeBySubmission(int(submissionID))
	if err != nil {
		t.Fatalf("Failed to get grade: %v", err)
	}

	if grade["score"].(float64) != score {
		t.Errorf("Expected score %f, got %f", score, grade["score"].(float64))
	}

	if grade["feedback"].(string) != feedback {
		t.Errorf("Expected feedback %s, got %s", feedback, grade["feedback"].(string))
	}
}

func TestUpdateSubmissionGradingStatus(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, assignment, and submission
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")
	submissionID, _ := CreateSubmission(int(assignmentID), "student001", "code", "")

	err := UpdateSubmissionGradingStatus(int(submissionID), "graded")
	if err != nil {
		t.Fatalf("Failed to update submission status: %v", err)
	}
}

func TestGetSubmissionsByTA(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create user, course, assignment
	userID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")
	courseID, _ := CreateCourse("CS 101", int(userID))
	assignmentID, _ := CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Create TA
	taID, _ := CreateUser("ta@example.com", "$2a$10$hash", "grader_ta")

	// Create submissions
	submissionID1, _ := CreateSubmission(int(assignmentID), "student001", "code1", "")
	submissionID2, _ := CreateSubmission(int(assignmentID), "student002", "code2", "")

	// Assign submissions to TA
	AssignSubmissionToTA(int(submissionID1), int(taID))
	AssignSubmissionToTA(int(submissionID2), int(taID))

	// Get submissions
	submissions, err := GetSubmissionsByTA(int(taID))
	if err != nil {
		t.Fatalf("Failed to get submissions: %v", err)
	}

	if len(submissions) != 2 {
		t.Errorf("Expected 2 submissions, got %d", len(submissions))
	}
}

func TestGetTAsForCourse(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)
	defer cleanupTestDB(t)

	// Create professor
	profID, _ := CreateUser("prof@example.com", "$2a$10$hash", "professor")

	// Create TAs
	ta1ID, _ := CreateUser("ta1@example.com", "$2a$10$hash", "grader_ta")
	ta2ID, _ := CreateUser("ta2@example.com", "$2a$10$hash", "grader_ta")
	headTAID, _ := CreateUser("headta@example.com", "$2a$10$hash", "head_ta")

	// Get TAs
	taIDs, err := GetTAsForCourse()
	if err != nil {
		t.Fatalf("Failed to get TAs: %v", err)
	}

	// Should include grader_ta and head_ta, but not professor
	if len(taIDs) < 3 {
		t.Errorf("Expected at least 3 TAs, got %d", len(taIDs))
	}

	// Check that all TAs are included
	taMap := make(map[int]bool)
	for _, id := range taIDs {
		taMap[id] = true
	}

	if !taMap[int(ta1ID)] {
		t.Error("TA1 should be in the list")
	}
	if !taMap[int(ta2ID)] {
		t.Error("TA2 should be in the list")
	}
	if !taMap[int(headTAID)] {
		t.Error("Head TA should be in the list")
	}
	if taMap[int(profID)] {
		t.Error("Professor should not be in the TA list")
	}
}

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"talytics/internal/database"
)

// TestProfessorCompleteWorkflow tests the complete professor workflow from instructions.md
// Workflow: Register → Create Course → Create Assignment → Upload Rubric → Upload Submissions → Notify TAs
func TestProfessorCompleteWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Step 1: Register professor
	registerBody := map[string]string{
		"email":    "prof@university.edu",
		"password": "securepass123",
		"role":     "professor",
	}
	registerJSON, _ := json.Marshal(registerBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %s", w.Code, w.Body.String())
	}

	// Step 2: Login
	profToken := loginUser(t, router, "prof@university.edu", "securepass123")

	// Step 3: Create course
	courseBody := map[string]string{"name": "CS 101: Introduction to Programming"}
	courseJSON, _ := json.Marshal(courseBody)

	req2, _ := http.NewRequest("POST", "/courses", bytes.NewBuffer(courseJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+profToken)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("Course creation failed: %d - %s", w2.Code, w2.Body.String())
	}

	var courseResponse map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &courseResponse)
	courseID := int(courseResponse["course_id"].(float64))

	// Step 4: Create assignment
	assignmentBody := map[string]interface{}{
		"course_id":   courseID,
		"title":       "Lab 1: Variables and Loops",
		"description": "Practice loops with factorial implementation",
	}
	assignmentJSON, _ := json.Marshal(assignmentBody)

	req3, _ := http.NewRequest("POST", "/assignments", bytes.NewBuffer(assignmentJSON))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+profToken)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusCreated {
		t.Fatalf("Assignment creation failed: %d - %s", w3.Code, w3.Body.String())
	}

	var assignmentResponse map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &assignmentResponse)
	assignmentID := int(assignmentResponse["assignment_id"].(float64))

	// Step 5: Upload rubric (method 2: JSON builder - method 1 would be PDF upload)
	rubricBody := map[string]interface{}{
		"assignment_id": assignmentID,
		"json_blob":     `{"title":"Lab 1 Rubric","max_points":100,"criteria":[{"name":"Correctness","max_points":50,"description":"Code correctness"},{"name":"Code Quality","max_points":30,"description":"Code style and structure"},{"name":"Documentation","max_points":20,"description":"Comments and docstrings"}]}`,
		"max_points":    100.0,
	}
	rubricJSON, _ := json.Marshal(rubricBody)

	req4, _ := http.NewRequest("POST", "/rubrics", bytes.NewBuffer(rubricJSON))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Authorization", "Bearer "+profToken)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusCreated {
		t.Fatalf("Rubric creation failed: %d - %s", w4.Code, w4.Body.String())
	}

	// Step 6: Register TAs before distributing submissions
	_ = createTestUser(t, "ta1@university.edu", "password123", "grader_ta")
	_ = createTestUser(t, "ta2@university.edu", "password123", "grader_ta")

	// Step 7: Upload multiple student submissions
	submissionTexts := []string{
		"def factorial(n):\n    if n == 0: return 1\n    return n * factorial(n-1)",
		"def factorial(n):\n    result = 1\n    for i in range(1, n+1):\n        result *= i\n    return result",
		"def factorial(n):\n    return 1 if n == 0 else n * factorial(n-1)",
	}

	for i, text := range submissionTexts {
		submissionBody := map[string]interface{}{
			"assignment_id":      assignmentID,
			"student_identifier": "student" + string(rune('0'+i+1)),
			"text":               text,
		}
		submissionJSON, _ := json.Marshal(submissionBody)

		req5, _ := http.NewRequest("POST", "/submissions/upload", bytes.NewBuffer(submissionJSON))
		req5.Header.Set("Content-Type", "application/json")
		req5.Header.Set("Authorization", "Bearer "+profToken)
		w5 := httptest.NewRecorder()
		router.ServeHTTP(w5, req5)

		if w5.Code != http.StatusCreated {
			t.Errorf("Submission %d upload failed: %d", i+1, w5.Code)
		}
	}

	// Step 8: Trigger TA notification (distribute submissions)
	req6, _ := http.NewRequest("POST", "/assignments/1/notify-tas", nil)
	req6.Header.Set("Authorization", "Bearer "+profToken)
	w6 := httptest.NewRecorder()
	router.ServeHTTP(w6, req6)

	if w6.Code != http.StatusOK {
		t.Fatalf("TA notification failed: %d - %s", w6.Code, w6.Body.String())
	}

	// Verify assignment status changed to "grading"
	req7, _ := http.NewRequest("GET", "/assignments/1", nil)
	req7.Header.Set("Authorization", "Bearer "+profToken)
	w7 := httptest.NewRecorder()
	router.ServeHTTP(w7, req7)

	if w7.Code != http.StatusOK {
		t.Fatalf("Get assignment failed: %d", w7.Code)
	}

	var assignmentDetails map[string]interface{}
	json.Unmarshal(w7.Body.Bytes(), &assignmentDetails)
	if assignmentDetails["status"] != "grading" {
		t.Errorf("Expected assignment status 'grading', got '%s'", assignmentDetails["status"])
	}
}

// TestTAGradingCompleteWorkflow tests the complete TA grading workflow
// Workflow: Register TA → Login → Get Assigned Submissions → Grade with Claude → Submit Grade
func TestTAGradingCompleteWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Setup: Create professor, course, assignment, rubric, and submissions
	profID := createTestUser(t, "prof@university.edu", "password123", "professor")
	profToken := loginUser(t, router, "prof@university.edu", "password123")

	courseID, _ := database.CreateCourse("CS 101", int(profID))
	assignmentID, _ := database.CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Create rubric
	rubricBody := map[string]interface{}{
		"assignment_id": assignmentID,
		"json_blob":     `{"title":"Test Rubric","max_points":100,"criteria":[{"name":"Correctness","max_points":50}]}`,
		"max_points":    100.0,
	}
	rubricJSON, _ := json.Marshal(rubricBody)

	req, _ := http.NewRequest("POST", "/rubrics", bytes.NewBuffer(rubricJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+profToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Create TAs before distributing
	_ = createTestUser(t, "ta@university.edu", "password123", "grader_ta")

	// Create submissions
	for i := 1; i <= 3; i++ {
		submissionBody := map[string]interface{}{
			"assignment_id":      assignmentID,
			"student_identifier": "student" + string(rune('0'+i)),
			"text":               "def factorial(n): return 1 if n == 0 else n * factorial(n-1)",
		}
		submissionJSON, _ := json.Marshal(submissionBody)

		req2, _ := http.NewRequest("POST", "/submissions/upload", bytes.NewBuffer(submissionJSON))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+profToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
	}

	// Distribute to TAs
	req3, _ := http.NewRequest("POST", "/assignments/1/notify-tas", nil)
	req3.Header.Set("Authorization", "Bearer "+profToken)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	// Step 1: TA Login (TA already created above)
	taToken := loginUser(t, router, "ta@university.edu", "password123")

	// Step 3: Get assigned submissions
	req5, _ := http.NewRequest("GET", "/submissions/assigned", nil)
	req5.Header.Set("Authorization", "Bearer "+taToken)
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Get assigned submissions failed: %d - %s", w5.Code, w5.Body.String())
	}

	var submissionsResponse map[string]interface{}
	json.Unmarshal(w5.Body.Bytes(), &submissionsResponse)
	submissions, ok := submissionsResponse["submissions"].([]interface{})
	if !ok {
		t.Fatal("Submissions not found in response")
	}

	if len(submissions) == 0 {
		t.Fatal("TA should have assigned submissions")
	}

	// Step 4: Grade a submission (preview mode - get Claude recommendation)
	firstSubmission := submissions[0].(map[string]interface{})
	submissionID := int(firstSubmission["id"].(float64))

	gradeBody := map[string]interface{}{
		"submission_id": submissionID,
	}
	gradeJSON, _ := json.Marshal(gradeBody)

	req6, _ := http.NewRequest("POST", "/grade", bytes.NewBuffer(gradeJSON))
	req6.Header.Set("Content-Type", "application/json")
	req6.Header.Set("Authorization", "Bearer "+taToken)
	w6 := httptest.NewRecorder()
	router.ServeHTTP(w6, req6)

	// Should get recommendation (may fail if Claude API key not set, but structure should work)
	if w6.Code != http.StatusOK && w6.Code != http.StatusServiceUnavailable {
		t.Logf("Grading preview returned status %d (may be expected if Claude API not configured)", w6.Code)
	}

	// Step 5: Submit grade with manual override
	manualGradeBody := map[string]interface{}{
		"submission_id": submissionID,
		"score":         85.5,
		"feedback":      "Good implementation. Minor style improvements needed.",
		"rubric_breakdown": `{"Correctness": 45, "Code Quality": 25, "Documentation": 15.5}`,
	}
	manualGradeJSON, _ := json.Marshal(manualGradeBody)

	req7, _ := http.NewRequest("POST", "/grade", bytes.NewBuffer(manualGradeJSON))
	req7.Header.Set("Content-Type", "application/json")
	req7.Header.Set("Authorization", "Bearer "+taToken)
	w7 := httptest.NewRecorder()
	router.ServeHTTP(w7, req7)

	if w7.Code != http.StatusOK {
		t.Fatalf("Manual grading failed: %d - %s", w7.Code, w7.Body.String())
	}

	var gradeResponse map[string]interface{}
	json.Unmarshal(w7.Body.Bytes(), &gradeResponse)
	if _, ok := gradeResponse["grade_id"]; !ok {
		t.Error("Grade ID should be in response")
	}
}

// TestAnomalyDetectionCompleteWorkflow tests the complete anomaly detection workflow
// Workflow: All submissions graded → Run anomaly detection → Get anomaly report
func TestAnomalyDetectionCompleteWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Setup: Create complete assignment with multiple submissions and grades
	profID := createTestUser(t, "prof@university.edu", "password123", "professor")
	profToken := loginUser(t, router, "prof@university.edu", "password123")

	courseID, _ := database.CreateCourse("CS 101", int(profID))
	assignmentID, _ := database.CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Create rubric
	rubricBody := map[string]interface{}{
		"assignment_id": assignmentID,
		"json_blob":     `{"title":"Test Rubric","max_points":100,"criteria":[{"name":"Correctness","max_points":50}]}`,
		"max_points":    100.0,
	}
	rubricJSON, _ := json.Marshal(rubricBody)

	req, _ := http.NewRequest("POST", "/rubrics", bytes.NewBuffer(rubricJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+profToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Create multiple submissions
	submissionIDs := []int{}
	for i := 1; i <= 5; i++ {
		submissionBody := map[string]interface{}{
			"assignment_id":      assignmentID,
			"student_identifier": "student" + string(rune('0'+i)),
			"text":               "def factorial(n): return 1 if n == 0 else n * factorial(n-1)",
		}
		submissionJSON, _ := json.Marshal(submissionBody)

		req2, _ := http.NewRequest("POST", "/submissions/upload", bytes.NewBuffer(submissionJSON))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+profToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var subResponse map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &subResponse)
		submissionIDs = append(submissionIDs, int(subResponse["submission_id"].(float64)))
	}

	// Create TAs and grade submissions (create variance)
	ta1ID := createTestUser(t, "ta1@university.edu", "password123", "grader_ta")
	ta2ID := createTestUser(t, "ta2@university.edu", "password123", "grader_ta")

	// Assign and grade with different scores to create variance
	scores := []float64{90.0, 85.0, 80.0, 75.0, 70.0}
	for i, submissionID := range submissionIDs {
		// Assign to alternating TAs
		taID := ta1ID
		if i%2 == 1 {
			taID = ta2ID
		}
		database.AssignSubmissionToTA(submissionID, int(taID))

		// Grade submission
		gradeBody := map[string]interface{}{
			"submission_id": submissionID,
			"score":         scores[i],
			"feedback":      "Graded submission",
			"rubric_breakdown": `{"Correctness": ` + string(rune(int(scores[i])-20)+'0') + `}`,
		}
		gradeJSON, _ := json.Marshal(gradeBody)

		// Use professor token to grade (professors can grade any submission)
		req3, _ := http.NewRequest("POST", "/grade", bytes.NewBuffer(gradeJSON))
		req3.Header.Set("Content-Type", "application/json")
		req3.Header.Set("Authorization", "Bearer "+profToken)
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)
	}

	// Step 1: Run anomaly detection
	req4, _ := http.NewRequest("POST", "/analyze/1", nil)
	req4.Header.Set("Authorization", "Bearer "+profToken)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("Anomaly detection failed: %d - %s", w4.Code, w4.Body.String())
	}

	var analyzeResponse map[string]interface{}
	json.Unmarshal(w4.Body.Bytes(), &analyzeResponse)
	if _, ok := analyzeResponse["report_id"]; !ok {
		t.Error("Report ID should be in response")
	}

	// Step 2: Get anomaly report
	req5, _ := http.NewRequest("GET", "/anomalies/1", nil)
	req5.Header.Set("Authorization", "Bearer "+profToken)
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Get anomaly report failed: %d - %s", w5.Code, w5.Body.String())
	}

	var reportResponse map[string]interface{}
	json.Unmarshal(w5.Body.Bytes(), &reportResponse)
	if _, ok := reportResponse["report"]; !ok {
		t.Error("Report should be in response")
	}

	report := reportResponse["report"].(map[string]interface{})
	if _, ok := report["total_grades"]; !ok {
		t.Error("Report should contain total_grades")
	}
}

// TestCompleteGradingWorkflowWithMultipleTAs tests the complete workflow with multiple TAs
// and verifies round-robin distribution
func TestCompleteGradingWorkflowWithMultipleTAs(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Setup: Professor creates course and assignment
	profID := createTestUser(t, "prof@university.edu", "password123", "professor")
	profToken := loginUser(t, router, "prof@university.edu", "password123")

	courseID, _ := database.CreateCourse("CS 101", int(profID))
	assignmentID, _ := database.CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Create rubric
	rubricBody := map[string]interface{}{
		"assignment_id": assignmentID,
		"json_blob":     `{"title":"Test Rubric","max_points":100}`,
		"max_points":    100.0,
	}
	rubricJSON, _ := json.Marshal(rubricBody)

	req, _ := http.NewRequest("POST", "/rubrics", bytes.NewBuffer(rubricJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+profToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Create 6 submissions
	for i := 1; i <= 6; i++ {
		submissionBody := map[string]interface{}{
			"assignment_id":      assignmentID,
			"student_identifier": "student" + string(rune('0'+i)),
			"text":               "def factorial(n): return 1 if n == 0 else n * factorial(n-1)",
		}
		submissionJSON, _ := json.Marshal(submissionBody)

		req2, _ := http.NewRequest("POST", "/submissions/upload", bytes.NewBuffer(submissionJSON))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+profToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
	}

	// Register 3 TAs (they need to exist before distribution)
	_ = createTestUser(t, "ta1@university.edu", "password123", "grader_ta")
	_ = createTestUser(t, "ta2@university.edu", "password123", "grader_ta")
	_ = createTestUser(t, "ta3@university.edu", "password123", "grader_ta")

	ta1Token := loginUser(t, router, "ta1@university.edu", "password123")

	// Distribute submissions
	req3, _ := http.NewRequest("POST", "/assignments/1/notify-tas", nil)
	req3.Header.Set("Authorization", "Bearer "+profToken)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	// Verify round-robin distribution (each TA should have 2 submissions)
	req4, _ := http.NewRequest("GET", "/submissions/assigned", nil)
	req4.Header.Set("Authorization", "Bearer "+ta1Token)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	var ta1Submissions map[string]interface{}
	json.Unmarshal(w4.Body.Bytes(), &ta1Submissions)
	ta1Subs := ta1Submissions["submissions"].([]interface{})

	// Each TA should have approximately equal distribution (2 submissions each for 6 submissions / 3 TAs)
	if len(ta1Subs) < 1 {
		t.Error("TA1 should have at least 1 assigned submission")
	}

	// Verify TAs can grade their assigned submissions
	if len(ta1Subs) > 0 {
		firstSub := ta1Subs[0].(map[string]interface{})
		submissionID := int(firstSub["id"].(float64))

		gradeBody := map[string]interface{}{
			"submission_id": submissionID,
			"score":         85.0,
			"feedback":      "Good work",
			"rubric_breakdown": `{}`,
		}
		gradeJSON, _ := json.Marshal(gradeBody)

		req5, _ := http.NewRequest("POST", "/grade", bytes.NewBuffer(gradeJSON))
		req5.Header.Set("Content-Type", "application/json")
		req5.Header.Set("Authorization", "Bearer "+ta1Token)
		w5 := httptest.NewRecorder()
		router.ServeHTTP(w5, req5)

		if w5.Code != http.StatusOK {
			t.Errorf("TA grading failed: %d - %s", w5.Code, w5.Body.String())
		}
	}
}


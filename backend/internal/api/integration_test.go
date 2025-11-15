package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"talytics/internal/api"
	"talytics/internal/auth"
	"talytics/internal/database"
	"talytics/internal/testutils"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, string) {
	// Setup test database
	dbPath := testutils.SetupTestDBPath(t)
	err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Initialize auth
	auth.InitJWT("test-secret-key")

	// Setup router
	gin.SetMode(gin.TestMode)
	router := api.SetupRouter()

	return router, ""
}

func teardownIntegrationTest(t *testing.T) {
	if database.DB != nil {
		// Cleanup tables
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
			database.DB.Exec("DELETE FROM " + table)
		}
		database.CloseDB()
	}
}

func createTestUser(t *testing.T, email, password, role string) int64 {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	userID, err := database.CreateUser(email, string(hashedPassword), role)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

func loginUser(t *testing.T, router *gin.Engine, email, password string) string {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	token, ok := response["token"].(string)
	if !ok {
		t.Fatal("Token not found in login response")
	}

	return token
}

func TestAuthFlow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Test registration
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "professor",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Test login
	token := loginUser(t, router, "test@example.com", "password123")
	if token == "" {
		t.Error("Login should return a token")
	}
}

func TestCourseWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Create professor
	_ = createTestUser(t, "prof@example.com", "password123", "professor")
	token := loginUser(t, router, "prof@example.com", "password123")

	// Create course
	reqBody := map[string]string{
		"name": "CS 101",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/courses", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	courseID, ok := response["course_id"].(float64)
	if !ok {
		t.Fatal("Course ID not found in response")
	}

	// Get courses
	req2, _ := http.NewRequest("GET", "/courses", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}

	var coursesResponse map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &coursesResponse)
	courses, ok := coursesResponse["courses"].([]interface{})
	if !ok {
		t.Fatal("Courses not found in response")
	}

	if len(courses) != 1 {
		t.Errorf("Expected 1 course, got %d", len(courses))
	}

	// Create assignment
	assignmentBody := map[string]interface{}{
		"course_id":   int(courseID),
		"title":       "Lab 1",
		"description": "First lab assignment",
	}
	assignmentJSON, _ := json.Marshal(assignmentBody)

	req3, _ := http.NewRequest("POST", "/assignments", bytes.NewBuffer(assignmentJSON))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()

	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w3.Code, w3.Body.String())
	}

	var assignmentResponse map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &assignmentResponse)
	assignmentIDFloat, ok := assignmentResponse["assignment_id"].(float64)
	if !ok {
		t.Fatal("Assignment ID not found in response")
	}
	_ = assignmentIDFloat // Use variable

	// Get assignment
	req4, _ := http.NewRequest("GET", "/assignments/1", nil)
	req4.Header.Set("Authorization", "Bearer "+token)
	w4 := httptest.NewRecorder()

	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w4.Code)
	}
}

func TestRubricWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Create professor, course, and assignment
	profID := createTestUser(t, "prof@example.com", "password123", "professor")
	token := loginUser(t, router, "prof@example.com", "password123")

	// Create course and assignment (simplified - would normally use API)
	courseID, _ := database.CreateCourse("CS 101", int(profID))
	assignmentID, _ := database.CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Create rubric manually
	rubricBody := map[string]interface{}{
		"assignment_id": assignmentID,
		"json_blob":     `{"title":"Test Rubric","max_points":100,"criteria":[]}`,
		"max_points":    100.0,
	}
	rubricJSON, _ := json.Marshal(rubricBody)

	req, _ := http.NewRequest("POST", "/rubrics", bytes.NewBuffer(rubricJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestSubmissionWorkflow(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Create professor, course, and assignment
	profID := createTestUser(t, "prof@example.com", "password123", "professor")
	token := loginUser(t, router, "prof@example.com", "password123")

	courseID, _ := database.CreateCourse("CS 101", int(profID))
	assignmentID, _ := database.CreateAssignment(int(courseID), "Lab 1", "First lab")

	// Upload submission
	submissionBody := map[string]interface{}{
		"assignment_id":      assignmentID,
		"student_identifier": "student001",
		"text":               "def factorial(n): return 1 if n == 0 else n * factorial(n-1)",
	}
	submissionJSON, _ := json.Marshal(submissionBody)

	req, _ := http.NewRequest("POST", "/submissions/upload", bytes.NewBuffer(submissionJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Create TA and get assigned submissions
	_ = createTestUser(t, "ta@example.com", "password123", "grader_ta")
	taToken := loginUser(t, router, "ta@example.com", "password123")

	// Get assignment ID from response
	var submissionResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &submissionResponse)

	// Distribute submissions
	req2, _ := http.NewRequest("POST", "/assignments/1/notify-tas", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	// Get assigned submissions
	req3, _ := http.NewRequest("GET", "/submissions/assigned", nil)
	req3.Header.Set("Authorization", "Bearer "+taToken)
	w3 := httptest.NewRecorder()

	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w3.Code)
	}
}

func TestProtectedEndpoints(t *testing.T) {
	router, _ := setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Test accessing protected endpoint without token
	req, _ := http.NewRequest("GET", "/courses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with invalid token
	req2, _ := http.NewRequest("GET", "/courses", nil)
	req2.Header.Set("Authorization", "Bearer invalid.token.here")
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for invalid token, got %d", http.StatusUnauthorized, w2.Code)
	}
}


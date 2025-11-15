package database

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

var DB *sql.DB

// InitDB initializes the SQLite database connection and creates tables
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err = DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Read and execute schema
	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err = DB.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// InsertRubric inserts a new rubric into the database
func InsertRubric(assignmentID int, jsonBlob string, maxPoints float64) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO rubrics (assignment_id, json_blob, max_points) VALUES (?, ?, ?)",
		assignmentID, jsonBlob, maxPoints,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rubric: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetAssignmentByID retrieves an assignment by its ID
func GetAssignmentByID(id int) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM assignments WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check assignment existence: %w", err)
	}
	return exists, nil
}

// RubricExists checks if a rubric already exists for an assignment
func RubricExists(assignmentID int) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM rubrics WHERE assignment_id = ?)", assignmentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check rubric existence: %w", err)
	}
	return exists, nil
}

// GetRubricByAssignmentID retrieves a rubric by assignment ID
func GetRubricByAssignmentID(assignmentID int) (id int, jsonBlob string, maxPoints float64, err error) {
	err = DB.QueryRow(
		"SELECT id, json_blob, max_points FROM rubrics WHERE assignment_id = ? LIMIT 1",
		assignmentID,
	).Scan(&id, &jsonBlob, &maxPoints)

	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to get rubric: %w", err)
	}

	return id, jsonBlob, maxPoints, nil
}

// User Management Functions

// CreateUser inserts a new user into the database
func CreateUser(email, passwordHash, role string) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)",
		email, passwordHash, role,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (id int, passwordHash, role string, err error) {
	err = DB.QueryRow(
		"SELECT id, password_hash, role FROM users WHERE email = ?",
		email,
	).Scan(&id, &passwordHash, &role)

	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get user: %w", err)
	}

	return id, passwordHash, role, nil
}

// UserExists checks if a user with the given email exists
func UserExists(email string) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// Course Management Functions

// CreateCourse inserts a new course into the database
func CreateCourse(name string, createdBy int) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO courses (name, created_by) VALUES (?, ?)",
		name, createdBy,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create course: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetCoursesByUser retrieves all courses created by a user
func GetCoursesByUser(userID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		`SELECT c.id, c.name, c.created_at, COUNT(a.id) as assignment_count
		 FROM courses c
		 LEFT JOIN assignments a ON c.id = a.course_id
		 WHERE c.created_by = ?
		 GROUP BY c.id
		 ORDER BY c.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %w", err)
	}
	defer rows.Close()

	var courses []map[string]interface{}
	for rows.Next() {
		var id, assignmentCount int
		var name, createdAt string
		if err := rows.Scan(&id, &name, &createdAt, &assignmentCount); err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}
		courses = append(courses, map[string]interface{}{
			"id":               id,
			"name":             name,
			"created_at":       createdAt,
			"assignment_count": assignmentCount,
		})
	}

	return courses, nil
}

// GetCourseByID retrieves a course by ID
func GetCourseByID(courseID int) (name string, createdBy int, err error) {
	err = DB.QueryRow(
		"SELECT name, created_by FROM courses WHERE id = ?",
		courseID,
	).Scan(&name, &createdBy)

	if err != nil {
		return "", 0, fmt.Errorf("failed to get course: %w", err)
	}

	return name, createdBy, nil
}

// UserOwnsCourse checks if a user owns a course
func UserOwnsCourse(userID, courseID int) (bool, error) {
	var createdBy int
	err := DB.QueryRow("SELECT created_by FROM courses WHERE id = ?", courseID).Scan(&createdBy)
	if err != nil {
		return false, fmt.Errorf("failed to check course ownership: %w", err)
	}
	return createdBy == userID, nil
}

// Assignment Management Functions

// CreateAssignment inserts a new assignment into the database
func CreateAssignment(courseID int, title, description string) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO assignments (course_id, title, description, status) VALUES (?, ?, ?, 'draft')",
		courseID, title, description,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create assignment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetAssignmentWithRubric retrieves an assignment and its rubric
func GetAssignmentWithRubric(assignmentID int) (map[string]interface{}, error) {
	// Get assignment
	var courseID int
	var title, description, status, createdAt string
	err := DB.QueryRow(
		"SELECT course_id, title, description, status, created_at FROM assignments WHERE id = ?",
		assignmentID,
	).Scan(&courseID, &title, &description, &status, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	result := map[string]interface{}{
		"id":          assignmentID,
		"course_id":   courseID,
		"title":       title,
		"description": description,
		"status":      status,
		"created_at":  createdAt,
	}

	// Try to get rubric
	var rubricID int
	var jsonBlob string
	var maxPoints float64
	err = DB.QueryRow(
		"SELECT id, json_blob, max_points FROM rubrics WHERE assignment_id = ?",
		assignmentID,
	).Scan(&rubricID, &jsonBlob, &maxPoints)

	if err == nil {
		result["rubric"] = map[string]interface{}{
			"id":            rubricID,
			"assignment_id": assignmentID,
			"json_blob":     jsonBlob,
			"max_points":    maxPoints,
		}
	}

	return result, nil
}

// GetAssignmentsByCourse retrieves all assignments for a course
func GetAssignmentsByCourse(courseID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		`SELECT id, course_id, title, description, status, created_at
		 FROM assignments
		 WHERE course_id = ?
		 ORDER BY created_at DESC`,
		courseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}
	defer rows.Close()

	var assignments []map[string]interface{}
	for rows.Next() {
		var id, cID int
		var title, description, status, createdAt string
		if err := rows.Scan(&id, &cID, &title, &description, &status, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, map[string]interface{}{
			"id":          id,
			"course_id":   cID,
			"title":       title,
			"description": description,
			"status":      status,
			"created_at":  createdAt,
		})
	}

	return assignments, nil
}

// UpdateAssignmentStatus updates the status of an assignment
func UpdateAssignmentStatus(assignmentID int, status string) error {
	_, err := DB.Exec(
		"UPDATE assignments SET status = ? WHERE id = ?",
		status, assignmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update assignment status: %w", err)
	}
	return nil
}

// GetSubmissionsByAssignment retrieves all submissions for an assignment
func GetSubmissionsByAssignment(assignmentID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		"SELECT id, student_identifier, assigned_ta_id FROM submissions WHERE assignment_id = ? ORDER BY id",
		assignmentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	defer rows.Close()

	var submissions []map[string]interface{}
	for rows.Next() {
		var id, assignedTAID sql.NullInt64
		var studentIdentifier string
		if err := rows.Scan(&id, &studentIdentifier, &assignedTAID); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}

		submission := map[string]interface{}{
			"id":                 id.Int64,
			"student_identifier": studentIdentifier,
		}
		if assignedTAID.Valid {
			submission["assigned_ta_id"] = assignedTAID.Int64
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// Submission Management Functions

// CreateSubmission inserts a new submission into the database
func CreateSubmission(assignmentID int, studentIdentifier, text, filePath string) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO submissions (assignment_id, student_identifier, text, file_path, graded_status) VALUES (?, ?, ?, ?, 'pending')",
		assignmentID, studentIdentifier, text, filePath,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create submission: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// AssignSubmissionToTA assigns a submission to a TA
func AssignSubmissionToTA(submissionID int, taID int) error {
	_, err := DB.Exec(
		"UPDATE submissions SET assigned_ta_id = ? WHERE id = ?",
		taID, submissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign submission to TA: %w", err)
	}
	return nil
}

// GetSubmissionsByTA retrieves all submissions assigned to a TA
func GetSubmissionsByTA(taID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		`SELECT s.id, s.assignment_id, s.student_identifier, s.text, s.file_path, s.graded_status,
		        a.title as assignment_title, c.name as course_name,
		        r.json_blob as rubric_json, r.max_points as rubric_max_points
		 FROM submissions s
		 JOIN assignments a ON s.assignment_id = a.id
		 JOIN courses c ON a.course_id = c.id
		 LEFT JOIN rubrics r ON a.id = r.assignment_id
		 WHERE s.assigned_ta_id = ?
		 ORDER BY s.id`,
		taID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}
	defer rows.Close()

	var submissions []map[string]interface{}
	for rows.Next() {
		var id, assignmentID int
		var studentIdentifier, text, filePath, gradedStatus, assignmentTitle, courseName string
		var rubricJSON sql.NullString
		var rubricMaxPoints sql.NullFloat64

		if err := rows.Scan(&id, &assignmentID, &studentIdentifier, &text, &filePath, &gradedStatus,
			&assignmentTitle, &courseName, &rubricJSON, &rubricMaxPoints); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}

		submission := map[string]interface{}{
			"id":                 id,
			"assignment_id":      assignmentID,
			"student_identifier": studentIdentifier,
			"text":               text,
			"file_path":          filePath,
			"graded_status":      gradedStatus,
			"assignment_title":   assignmentTitle,
			"course_name":        courseName,
		}

		if rubricJSON.Valid {
			submission["rubric_json"] = rubricJSON.String
		}
		if rubricMaxPoints.Valid {
			submission["rubric_max_points"] = rubricMaxPoints.Float64
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// GetTAsForCourse retrieves all TAs (users with grader_ta or head_ta role)
func GetTAsForCourse() ([]int, error) {
	rows, err := DB.Query("SELECT id FROM users WHERE role IN ('grader_ta', 'head_ta')")
	if err != nil {
		return nil, fmt.Errorf("failed to get TAs: %w", err)
	}
	defer rows.Close()

	var taIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan TA ID: %w", err)
		}
		taIDs = append(taIDs, id)
	}

	return taIDs, nil
}

// Grade Management Functions

// SaveGrade inserts a new grade into the database
func SaveGrade(submissionID, gradedBy int, score float64, feedback string, rubricBreakdown string) (int64, error) {
	// Get rubric_id from submission's assignment
	submission, err := GetSubmissionDetails(submissionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get submission details: %w", err)
	}
	
	assignmentID := submission["assignment_id"].(int64)
	rubricID, _, _, err := GetRubricByAssignmentID(int(assignmentID))
	if err != nil {
		return 0, fmt.Errorf("failed to get rubric: %w", err)
	}

	// Store score as both total_points and score for compatibility
	result, err := DB.Exec(
		"INSERT INTO grades (submission_id, rubric_id, graded_by, score, total_points, feedback, rubric_breakdown, json_blob) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		submissionID, rubricID, gradedBy, score, score, feedback, rubricBreakdown, rubricBreakdown,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save grade: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetGradeBySubmission retrieves a grade by submission ID
func GetGradeBySubmission(submissionID int) (map[string]interface{}, error) {
	var gradeID, gradedBy int
	var score float64
	var feedback, rubricBreakdown, createdAt, graderEmail, graderRole string

	err := DB.QueryRow(
		`SELECT g.id, g.graded_by, g.score, g.feedback, g.rubric_breakdown, g.created_at,
		        u.email as grader_email, u.role as grader_role
		 FROM grades g
		 JOIN users u ON g.graded_by = u.id
		 WHERE g.submission_id = ?
		 ORDER BY g.created_at DESC
		 LIMIT 1`,
		submissionID,
	).Scan(&gradeID, &gradedBy, &score, &feedback, &rubricBreakdown, &createdAt, &graderEmail, &graderRole)

	if err != nil {
		return nil, fmt.Errorf("failed to get grade: %w", err)
	}

	grade := map[string]interface{}{
		"id":                gradeID,
		"submission_id":     submissionID,
		"graded_by":         gradedBy,
		"grader_email":      graderEmail,
		"grader_role":       graderRole,
		"score":             score,
		"feedback":          feedback,
		"rubric_breakdown":  rubricBreakdown,
		"created_at":        createdAt,
	}

	return grade, nil
}

// GetSubmissionDetails retrieves full submission details for grading
func GetSubmissionDetails(submissionID int) (map[string]interface{}, error) {
	var id, assignmentID, assignedTAID sql.NullInt64
	var studentIdentifier, text, filePath, gradedStatus string

	err := DB.QueryRow(
		"SELECT id, assignment_id, student_identifier, text, file_path, assigned_ta_id, graded_status FROM submissions WHERE id = ?",
		submissionID,
	).Scan(&id, &assignmentID, &studentIdentifier, &text, &filePath, &assignedTAID, &gradedStatus)

	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	submission := map[string]interface{}{
		"id":                 id.Int64,
		"assignment_id":      assignmentID.Int64,
		"student_identifier": studentIdentifier,
		"text":               text,
		"file_path":          filePath,
		"graded_status":      gradedStatus,
	}

	if assignedTAID.Valid {
		submission["assigned_ta_id"] = assignedTAID.Int64
	}

	// Get assignment details with rubric
	var title, description, status string
	var rubricJSON sql.NullString
	var maxPoints sql.NullFloat64

	err = DB.QueryRow(
		`SELECT a.title, a.description, a.status, r.json_blob, r.max_points
		 FROM assignments a
		 LEFT JOIN rubrics r ON a.id = r.assignment_id
		 WHERE a.id = ?`,
		assignmentID.Int64,
	).Scan(&title, &description, &status, &rubricJSON, &maxPoints)

	if err == nil {
		submission["assignment_title"] = title
		submission["assignment_description"] = description
		submission["assignment_status"] = status
		if rubricJSON.Valid {
			submission["rubric_json"] = rubricJSON.String
		}
		if maxPoints.Valid {
			submission["rubric_max_points"] = maxPoints.Float64
		}
	}

	return submission, nil
}

// UpdateSubmissionGradingStatus updates the grading status of a submission
func UpdateSubmissionGradingStatus(submissionID int, status string) error {
	_, err := DB.Exec(
		"UPDATE submissions SET graded_status = ? WHERE id = ?",
		status, submissionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission grading status: %w", err)
	}
	return nil
}

// GetUngradedSubmissionsForAssignment retrieves all ungraded submissions for an assignment
func GetUngradedSubmissionsForAssignment(assignmentID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		`SELECT s.id, s.student_identifier, s.text, s.file_path, s.assigned_ta_id
		 FROM submissions s
		 LEFT JOIN grades g ON s.id = g.submission_id
		 WHERE s.assignment_id = ? AND g.id IS NULL
		 ORDER BY s.id`,
		assignmentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ungraded submissions: %w", err)
	}
	defer rows.Close()

	var submissions []map[string]interface{}
	for rows.Next() {
		var id int
		var studentIdentifier, text, filePath string
		var assignedTAID sql.NullInt64

		if err := rows.Scan(&id, &studentIdentifier, &text, &filePath, &assignedTAID); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}

		submission := map[string]interface{}{
			"id":                 id,
			"student_identifier": studentIdentifier,
			"text":               text,
			"file_path":          filePath,
		}

		if assignedTAID.Valid {
			submission["assigned_ta_id"] = int(assignedTAID.Int64)
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// Anomaly Report Management Functions

// GetAllGradesForAssignment retrieves all grades with grader info for statistical analysis
func GetAllGradesForAssignment(assignmentID int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		`SELECT g.id, g.submission_id, g.graded_by, g.score, g.feedback, g.rubric_breakdown,
		        s.student_identifier, u.email as grader_email
		 FROM grades g
		 JOIN submissions s ON g.submission_id = s.id
		 JOIN users u ON g.graded_by = u.id
		 WHERE s.assignment_id = ?
		 ORDER BY g.id`,
		assignmentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []map[string]interface{}
	for rows.Next() {
		var gradeID, submissionID, gradedBy int
		var score float64
		var feedback, rubricBreakdown, studentIdentifier, graderEmail string

		if err := rows.Scan(&gradeID, &submissionID, &gradedBy, &score, &feedback, &rubricBreakdown,
			&studentIdentifier, &graderEmail); err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}

		grade := map[string]interface{}{
			"id":                 gradeID,
			"submission_id":      submissionID,
			"graded_by":          gradedBy,
			"score":              score,
			"feedback":           feedback,
			"rubric_breakdown":   rubricBreakdown,
			"student_identifier": studentIdentifier,
			"grader_email":       graderEmail,
		}
		grades = append(grades, grade)
	}

	return grades, nil
}

// SaveAnomalyReport inserts a new anomaly report into the database
func SaveAnomalyReport(assignmentID int, jsonBlob string, status string) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO anomaly_reports (assignment_id, json_blob, status) VALUES (?, ?, ?)",
		assignmentID, jsonBlob, status,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save anomaly report: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetAnomalyReport retrieves the most recent anomaly report for an assignment
func GetAnomalyReport(assignmentID int) (map[string]interface{}, error) {
	var reportID int
	var jsonBlob, status, createdAt string

	err := DB.QueryRow(
		`SELECT id, json_blob, status, created_at
		 FROM anomaly_reports
		 WHERE assignment_id = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		assignmentID,
	).Scan(&reportID, &jsonBlob, &status, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get anomaly report: %w", err)
	}

	report := map[string]interface{}{
		"id":            reportID,
		"assignment_id": assignmentID,
		"json_blob":     jsonBlob,
		"status":        status,
		"created_at":    createdAt,
	}

	return report, nil
}

// UpdateAnomalyReportStatus updates the status of an anomaly report
func UpdateAnomalyReportStatus(reportID int, status string) error {
	_, err := DB.Exec(
		"UPDATE anomaly_reports SET status = ? WHERE id = ?",
		status, reportID,
	)
	if err != nil {
		return fmt.Errorf("failed to update anomaly report status: %w", err)
	}
	return nil
}

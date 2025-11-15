package models

import "time"

// User represents a system user (professor, head_ta, or grader_ta)
type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"` // professor, head_ta, grader_ta
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Course represents an academic course
type Course struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Assignment represents a course assignment
type Assignment struct {
	ID          int       `json:"id" db:"id"`
	CourseID    int       `json:"course_id" db:"course_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"` // draft, open, grading, completed
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Rubric represents a grading rubric stored as JSON
type Rubric struct {
	ID           int       `json:"id" db:"id"`
	AssignmentID int       `json:"assignment_id" db:"assignment_id"`
	JSONBlob     string    `json:"json_blob" db:"json_blob"`
	MaxPoints    float64   `json:"max_points" db:"max_points"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Submission represents a student submission
type Submission struct {
	ID                int       `json:"id" db:"id"`
	AssignmentID      int       `json:"assignment_id" db:"assignment_id"`
	StudentIdentifier string    `json:"student_identifier" db:"student_identifier"`
	Text              string    `json:"text" db:"text"`
	FilePath          string    `json:"file_path" db:"file_path"`
	GradedStatus      string    `json:"graded_status" db:"graded_status"` // pending, in_progress, graded, regrade_required
	AssignedTAID      *int      `json:"assigned_ta_id" db:"assigned_ta_id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// Grade represents a grading result
type Grade struct {
	ID           int       `json:"id" db:"id"`
	SubmissionID int       `json:"submission_id" db:"submission_id"`
	RubricID     int       `json:"rubric_id" db:"rubric_id"`
	JSONBlob     string    `json:"json_blob" db:"json_blob"`
	TotalPoints  float64   `json:"total_points" db:"total_points"`
	GradedBy     int       `json:"graded_by" db:"graded_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AnomalyReport represents an anomaly detection report
type AnomalyReport struct {
	ID           int       `json:"id" db:"id"`
	AssignmentID int       `json:"assignment_id" db:"assignment_id"`
	JSONBlob     string    `json:"json_blob" db:"json_blob"`
	Status       string    `json:"status" db:"status"` // pending, reviewed, resolved
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Request/Response DTOs

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=professor head_ta grader_ta"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateCourseRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateAssignmentRequest struct {
	CourseID    int    `json:"course_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type CreateRubricRequest struct {
	AssignmentID int     `json:"assignment_id" binding:"required"`
	JSONBlob     string  `json:"json_blob" binding:"required"`
	MaxPoints    float64 `json:"max_points" binding:"required"`
}

type UploadSubmissionRequest struct {
	AssignmentID      int    `json:"assignment_id" binding:"required"`
	StudentIdentifier string `json:"student_identifier" binding:"required"`
	Text              string `json:"text"`
	FilePath          string `json:"file_path"`
}

type GradeRequest struct {
	SubmissionID int     `json:"submission_id" binding:"required"`
	RubricID     int     `json:"rubric_id" binding:"required"`
	JSONBlob     string  `json:"json_blob" binding:"required"`
	TotalPoints  float64 `json:"total_points" binding:"required"`
}

type UploadRubricPDFResponse struct {
	RubricID     int     `json:"rubric_id"`
	AssignmentID int     `json:"assignment_id"`
	ParsedRubric string  `json:"parsed_rubric"`
	MaxPoints    float64 `json:"max_points"`
	FilePath     string  `json:"file_path"`
}

// CourseWithDetails includes course info with assignment count
type CourseWithDetails struct {
	Course
	AssignmentCount int `json:"assignment_count"`
}

// AssignmentWithRubric includes assignment and its rubric
type AssignmentWithRubric struct {
	Assignment
	Rubric *Rubric `json:"rubric,omitempty"`
}

// SubmissionWithDetails includes submission with assignment and rubric
type SubmissionWithDetails struct {
	Submission
	AssignmentTitle string  `json:"assignment_title"`
	CourseName      string  `json:"course_name"`
	RubricJSON      string  `json:"rubric_json,omitempty"`
	RubricMaxPoints float64 `json:"rubric_max_points,omitempty"`
}

// TADistributionResponse shows how submissions were distributed
type TADistributionResponse struct {
	AssignmentID     int         `json:"assignment_id"`
	TotalSubmissions int         `json:"total_submissions"`
	Distribution     map[int]int `json:"distribution"` // TA_ID -> count
	Message          string      `json:"message"`
}

// Grading DTOs

// GradeSubmissionRequest - POST /grade request
type GradeSubmissionRequest struct {
	SubmissionID         int     `json:"submission_id" binding:"required"`
	Score                float64 `json:"score"`                          // Optional, TA can override Claude
	Feedback             string  `json:"feedback"`                       // Optional, TA can add notes
	UseClaudeRecommendation bool `json:"use_claude_recommendation"`     // If true, use Claude's suggestion
	RubricBreakdown      string  `json:"rubric_breakdown"`               // JSON string of criterion -> points
}

// BatchGradeRequest - POST /grade/batch request
type BatchGradeRequest struct {
	AssignmentID int  `json:"assignment_id" binding:"required"`
	AutoApprove  bool `json:"auto_approve"` // If true, auto-save all Claude recommendations
}

// GradingRecommendation - Claude's grading output
type GradingRecommendation struct {
	Score           float64            `json:"score"`
	MaxPoints       float64            `json:"max_points"`
	Feedback        string             `json:"feedback"`
	RubricBreakdown map[string]float64 `json:"rubric_breakdown"` // criterion -> points
	Justification   string             `json:"justification"`
}

// GradeResponse - POST /grade response
type GradeResponse struct {
	Message              string                 `json:"message"`
	GradeID              int64                  `json:"grade_id,omitempty"`
	ClaudeRecommendation *GradingRecommendation `json:"claude_recommendation,omitempty"`
}

// BatchGradeResponse - POST /grade/batch response
type BatchGradeResponse struct {
	AssignmentID      int     `json:"assignment_id"`
	TotalGraded       int     `json:"total_graded"`
	AverageScore      float64 `json:"average_score"`
	GradedSubmissions []int   `json:"graded_submission_ids"`
	Message           string  `json:"message"`
}

// Anomaly Detection DTOs

// AnomalySummary - Summary statistics from analysis
type AnomalySummary struct {
	AssignmentID      int                   `json:"assignment_id"`
	TotalGrades       int                   `json:"total_grades"`
	AverageScore      float64               `json:"average_score"`
	StandardDeviation float64               `json:"standard_deviation"`
	TASeverityIssues  []TASeverityAnomaly   `json:"ta_severity_issues"`
	OutlierGrades     []OutlierAnomaly      `json:"outlier_grades"`
	CriterionIssues   []CriterionAnomaly    `json:"criterion_issues"`
	RegradeRisks      []RegradeRisk         `json:"regrade_risks"`
	GeneratedAt       string                `json:"generated_at"`
}

// TASeverityAnomaly - TA grading too harsh/lenient
type TASeverityAnomaly struct {
	TAID         int     `json:"ta_id"`
	TAEmail      string  `json:"ta_email"`
	AverageScore float64 `json:"average_score"`
	GradesCount  int     `json:"grades_count"`
	Deviation    float64 `json:"deviation"`  // Standard deviations from mean
	Severity     string  `json:"severity"`   // "too_harsh" or "too_lenient"
}

// OutlierAnomaly - Unusual grade (statistical outlier)
type OutlierAnomaly struct {
	SubmissionID      int     `json:"submission_id"`
	StudentIdentifier string  `json:"student_identifier"`
	Score             float64 `json:"score"`
	ZScore            float64 `json:"z_score"`  // How many std devs from mean
	GradedBy          int     `json:"graded_by"`
	GraderEmail       string  `json:"grader_email"`
}

// CriterionAnomaly - Inconsistent scoring on specific criterion
type CriterionAnomaly struct {
	CriterionName      string  `json:"criterion_name"`
	AverageScore       float64 `json:"average_score"`
	StandardDeviation  float64 `json:"standard_deviation"`
	InconsistentGrades []int   `json:"inconsistent_submission_ids"` // Submissions with unusual criterion scores
}

// RegradeRisk - Submission flagged for potential regrade
type RegradeRisk struct {
	SubmissionID      int      `json:"submission_id"`
	StudentIdentifier string   `json:"student_identifier"`
	Score             float64  `json:"score"`
	RiskScore         int      `json:"risk_score"`    // 0-100
	RiskFactors       []string `json:"risk_factors"`  // ["outlier", "harsh_ta", "boundary_grade"]
	GradedBy          int      `json:"graded_by"`
	GraderEmail       string   `json:"grader_email"`
}

// AnalyzeResponse - POST /analyze response
type AnalyzeResponse struct {
	Message  string         `json:"message"`
	ReportID int64          `json:"report_id"`
	Summary  AnomalySummary `json:"summary"`
}

// GetAnomaliesResponse - GET /anomalies response
type GetAnomaliesResponse struct {
	Report AnomalySummary `json:"report"`
	Status string         `json:"status"` // pending, reviewed, resolved
}

// UpdateAnomalyStatusRequest - PATCH /anomalies/:id request
type UpdateAnomalyStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending reviewed resolved"`
}

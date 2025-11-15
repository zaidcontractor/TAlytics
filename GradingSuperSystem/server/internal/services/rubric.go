package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/talytics/server/internal/database"
	pb "github.com/talytics/server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RubricService struct {
	pb.UnimplementedRubricServiceServer
	db *database.Database
}

func NewRubricService(db *database.Database) *RubricService {
	return &RubricService{db: db}
}

func (s *RubricService) CreateRubric(ctx context.Context, req *pb.CreateRubricRequest) (*pb.RubricResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can create rubrics
	if userRole != "instructor" {
		return nil, errors.New("only instructors can create rubrics")
	}

	// Check if user is instructor of this course
	var instructorID int64
	err := s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", req.CourseId).Scan(&instructorID)
	if err == sql.ErrNoRows {
		return nil, errors.New("course not found")
	}
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can create rubrics for this course")
	}

	// Validate weights sum to 100
	var totalWeight float64
	for _, weight := range req.Weights {
		totalWeight += weight
	}
	if totalWeight < 99.9 || totalWeight > 100.1 {
		return nil, fmt.Errorf("criteria weights must sum to 100%%")
	}

	// Validate input
	if req.Name == "" {
		return nil, errors.New("rubric name is required")
	}

	if len(req.Criteria) == 0 || len(req.Weights) == 0 {
		return nil, errors.New("criteria and weights are required")
	}

	if len(req.Criteria) != len(req.Weights) {
		return nil, errors.New("number of criteria must match number of weights")
	}

	// Convert arrays to JSON for storage
	criteriaJSON, err := json.Marshal(req.Criteria)
	if err != nil {
		return nil, err
	}
	
	weightsJSON, err := json.Marshal(req.Weights)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO rubrics (name, course_id, criteria, weights, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, err := s.db.DB.Exec(query, req.Name, req.CourseId, string(criteriaJSON), string(weightsJSON), userID)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get created rubric
	rubric, err := s.getRubricByID(id)
	if err != nil {
		return nil, err
	}

	return &pb.RubricResponse{
		Rubric:  rubric,
		Message: "Rubric created successfully",
	}, nil
}

func (s *RubricService) GetRubric(ctx context.Context, req *pb.GetRubricRequest) (*pb.RubricResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Check if user has access to this rubric (is member of the course)
	var courseID int64
	err := s.db.DB.QueryRow("SELECT course_id FROM rubrics WHERE id = ?", req.Id).Scan(&courseID)
	if err == sql.ErrNoRows {
		return nil, errors.New("rubric not found")
	}
	if err != nil {
		return nil, err
	}

	// Check course membership
	var memberCount int
	err = s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, courseID, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, errors.New("access denied: you are not a member of this course")
	}

	rubric, err := s.getRubricByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.RubricResponse{
		Rubric:  rubric,
		Message: "Rubric retrieved successfully",
	}, nil
}

func (s *RubricService) ListRubrics(ctx context.Context, req *pb.ListRubricsRequest) (*pb.ListRubricsResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Check if user is a member of this course
	var memberCount int
	err := s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, req.CourseId, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, errors.New("access denied: you are not a member of this course")
	}

	rows, err := s.db.DB.Query(`
		SELECT r.id, r.name, r.course_id, r.criteria, r.weights, r.created_by, u.name, r.created_at, r.updated_at 
		FROM rubrics r
		LEFT JOIN users u ON r.created_by = u.id
		WHERE r.course_id = ?
		ORDER BY r.created_at DESC
	`, req.CourseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rubrics []*pb.Rubric
	for rows.Next() {
		var rubric pb.Rubric
		var criteriaJSON, weightsJSON string
		var createdAt, updatedAt time.Time
		var creatorName sql.NullString

		err := rows.Scan(&rubric.Id, &rubric.Name, &rubric.CourseId, &criteriaJSON, &weightsJSON, 
			&rubric.CreatedBy, &creatorName, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		err = json.Unmarshal([]byte(criteriaJSON), &rubric.Criteria)
		if err != nil {
			return nil, err
		}
		
		err = json.Unmarshal([]byte(weightsJSON), &rubric.Weights)
		if err != nil {
			return nil, err
		}

		rubric.CreatedAt = timestamppb.New(createdAt)
		rubric.UpdatedAt = timestamppb.New(updatedAt)

		if creatorName.Valid {
			rubric.CreatorName = creatorName.String
		}

		rubrics = append(rubrics, &rubric)
	}

	return &pb.ListRubricsResponse{
		Rubrics: rubrics,
	}, nil
}

func (s *RubricService) UpdateRubric(ctx context.Context, req *pb.UpdateRubricRequest) (*pb.RubricResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can update rubrics
	if userRole != "instructor" {
		return nil, errors.New("only instructors can update rubrics")
	}

	// Check if rubric exists and get course info
	var courseID, createdBy int64
	err := s.db.DB.QueryRow("SELECT course_id, created_by FROM rubrics WHERE id = ?", req.Id).Scan(&courseID, &createdBy)
	if err == sql.ErrNoRows {
		return nil, errors.New("rubric not found")
	}
	if err != nil {
		return nil, err
	}

	// Check if user is instructor of this course
	var instructorID int64
	err = s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", courseID).Scan(&instructorID)
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can update this rubric")
	}

	// Validate weights sum to 100
	var totalWeight float64
	for _, weight := range req.Weights {
		totalWeight += weight
	}
	if totalWeight < 99.9 || totalWeight > 100.1 {
		return nil, fmt.Errorf("criteria weights must sum to 100%%")
	}

	// Validate input
	if req.Name == "" {
		return nil, errors.New("rubric name is required")
	}

	if len(req.Criteria) == 0 || len(req.Weights) == 0 {
		return nil, errors.New("criteria and weights are required")
	}

	if len(req.Criteria) != len(req.Weights) {
		return nil, errors.New("number of criteria must match number of weights")
	}

	// Convert arrays to JSON for storage
	criteriaJSON, err := json.Marshal(req.Criteria)
	if err != nil {
		return nil, err
	}
	
	weightsJSON, err := json.Marshal(req.Weights)
	if err != nil {
		return nil, err
	}

	// Start transaction for atomic update
	tx, err := s.db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update rubric
	_, err = tx.Exec(`
		UPDATE rubrics 
		SET name = ?, criteria = ?, weights = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Name, string(criteriaJSON), string(weightsJSON), req.Id)
	if err != nil {
		return nil, err
	}

	// If force_regrading is requested (from AI suggestions), mark all grades using this rubric for regrading
	// This is passed via context or we can add it to the request
	// For now, we'll check a custom header or add a field to UpdateRubricRequest

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Get updated rubric
	rubric, err := s.getRubricByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.RubricResponse{
		Rubric:  rubric,
		Message: "Rubric updated successfully",
	}, nil
}

// Helper function to mark grades for regrading when rubric changes
func (s *RubricService) MarkGradesForRegrading(rubricID int64) error {
	// Find all assignments using this rubric
	rows, err := s.db.DB.Query("SELECT id FROM assignments WHERE rubric_id = ?", rubricID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var assignmentIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		assignmentIDs = append(assignmentIDs, id)
	}

	// Mark all grades for these assignments as needing regrading
	for _, assignmentID := range assignmentIDs {
		_, err := s.db.DB.Exec(`
			UPDATE grades 
			SET needs_regrading = 1
			WHERE assignment_id = ?
		`, assignmentID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RubricService) DeleteRubric(ctx context.Context, req *pb.DeleteRubricRequest) (*pb.DeleteRubricResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can delete rubrics
	if userRole != "instructor" {
		return nil, errors.New("only instructors can delete rubrics")
	}

	// Check if rubric exists and get course info
	var courseID int64
	err := s.db.DB.QueryRow("SELECT course_id FROM rubrics WHERE id = ?", req.Id).Scan(&courseID)
	if err == sql.ErrNoRows {
		return nil, errors.New("rubric not found")
	}
	if err != nil {
		return nil, err
	}

	// Check if user is instructor of this course
	var instructorID int64
	err = s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", courseID).Scan(&instructorID)
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can delete this rubric")
	}

	// Check if rubric is being used by any assignments
	var assignmentCount int
	err = s.db.DB.QueryRow("SELECT COUNT(*) FROM assignments WHERE rubric_id = ?", req.Id).Scan(&assignmentCount)
	if err != nil {
		return nil, err
	}
	if assignmentCount > 0 {
		return nil, errors.New("cannot delete rubric: it is being used by one or more assignments")
	}

	_, err = s.db.DB.Exec("DELETE FROM rubrics WHERE id = ?", req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteRubricResponse{
		Message: "Rubric deleted successfully",
	}, nil
}

func (s *RubricService) getRubricByID(rubricID int64) (*pb.Rubric, error) {
	var rubric pb.Rubric
	var criteriaJSON, weightsJSON string
	var createdAt, updatedAt time.Time
	var creatorName sql.NullString

	err := s.db.DB.QueryRow(`
		SELECT r.id, r.name, r.course_id, r.criteria, r.weights, r.created_by, u.name, r.created_at, r.updated_at
		FROM rubrics r
		LEFT JOIN users u ON r.created_by = u.id
		WHERE r.id = ?
	`, rubricID).Scan(&rubric.Id, &rubric.Name, &rubric.CourseId, &criteriaJSON, &weightsJSON,
		&rubric.CreatedBy, &creatorName, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	// Parse JSON arrays
	err = json.Unmarshal([]byte(criteriaJSON), &rubric.Criteria)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal([]byte(weightsJSON), &rubric.Weights)
	if err != nil {
		return nil, err
	}

	rubric.CreatedAt = timestamppb.New(createdAt)
	rubric.UpdatedAt = timestamppb.New(updatedAt)

	if creatorName.Valid {
		rubric.CreatorName = creatorName.String
	}

	return &rubric, nil
}

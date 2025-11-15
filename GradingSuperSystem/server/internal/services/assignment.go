package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/talytics/server/internal/database"
	pb "github.com/talytics/server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AssignmentService struct {
	pb.UnimplementedAssignmentServiceServer
	db *database.Database
}

func NewAssignmentService(db *database.Database) *AssignmentService {
	return &AssignmentService{
		db: db,
	}
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, req *pb.CreateAssignmentRequest) (*pb.AssignmentResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can create assignments
	if userRole != "instructor" {
		return nil, errors.New("only instructors can create assignments")
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
		return nil, errors.New("only the course instructor can create assignments for this course")
	}

	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("assignment name is required")
	}

	// Rubric is now required - validate rubric if provided
	if req.RubricId != 0 {
		var rubricCourseID int64
		err := s.db.DB.QueryRow("SELECT course_id FROM rubrics WHERE id = ?", req.RubricId).Scan(&rubricCourseID)
		if err == sql.ErrNoRows {
			return nil, errors.New("rubric not found")
		}
		if err != nil {
			return nil, err
		}
		if rubricCourseID != req.CourseId {
			return nil, errors.New("rubric does not belong to this course")
		}
	} else {
		return nil, errors.New("rubric is required when creating an assignment")
	}

	// Insert assignment
	result, err := s.db.DB.Exec(`
		INSERT INTO assignments (course_id, name, description, rubric_id, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, req.CourseId, req.Name, req.Description, req.RubricId, userID)
	if err != nil {
		return nil, err
	}

	assignmentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get created assignment
	assignment, err := s.getAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}

	return &pb.AssignmentResponse{
		Assignment: assignment,
		Message:    "Assignment created successfully",
	}, nil
}

func (s *AssignmentService) GetAssignment(ctx context.Context, req *pb.GetAssignmentRequest) (*pb.AssignmentResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Check if user has access to this assignment (is member of the course)
	var courseID int64
	err := s.db.DB.QueryRow("SELECT course_id FROM assignments WHERE id = ?", req.Id).Scan(&courseID)
	if err == sql.ErrNoRows {
		return nil, errors.New("assignment not found")
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

	assignment, err := s.getAssignmentByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.AssignmentResponse{
		Assignment: assignment,
		Message:    "Assignment retrieved successfully",
	}, nil
}

func (s *AssignmentService) ListAssignments(ctx context.Context, req *pb.ListAssignmentsRequest) (*pb.ListAssignmentsResponse, error) {
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

	// Get assignments for this course
	rows, err := s.db.DB.Query(`
		SELECT a.id, a.course_id, a.name, a.description, a.rubric_id, 
		       a.created_by, u.name, a.created_at, a.updated_at
		FROM assignments a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.course_id = ?
		ORDER BY a.created_at DESC
	`, req.CourseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*pb.Assignment
	for rows.Next() {
		var assignment pb.Assignment
		var createdAt, updatedAt time.Time
		var rubricID sql.NullInt64
		var creatorName sql.NullString

		err := rows.Scan(&assignment.Id, &assignment.CourseId, &assignment.Name, 
			&assignment.Description, &rubricID,
			&assignment.CreatedBy, &creatorName, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		assignment.CreatedAt = timestamppb.New(createdAt)
		assignment.UpdatedAt = timestamppb.New(updatedAt)

		if rubricID.Valid {
			assignment.RubricId = rubricID.Int64
			// Get rubric details
			rubric, err := s.getRubricByID(rubricID.Int64)
			if err == nil {
				assignment.Rubric = rubric
			}
		}

		if creatorName.Valid {
			assignment.CreatorName = creatorName.String
		}

		assignments = append(assignments, &assignment)
	}

	return &pb.ListAssignmentsResponse{
		Assignments: assignments,
	}, nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, req *pb.UpdateAssignmentRequest) (*pb.AssignmentResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can update assignments
	if userRole != "instructor" {
		return nil, errors.New("only instructors can update assignments")
	}

	// Check if assignment exists and get course info
	var courseID, createdBy int64
	err := s.db.DB.QueryRow("SELECT course_id, created_by FROM assignments WHERE id = ?", req.Id).Scan(&courseID, &createdBy)
	if err == sql.ErrNoRows {
		return nil, errors.New("assignment not found")
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
		return nil, errors.New("only the course instructor can update this assignment")
	}

	// Rubric is required - validate if provided
	if req.RubricId != 0 {
		var rubricCourseID int64
		err := s.db.DB.QueryRow("SELECT course_id FROM rubrics WHERE id = ?", req.RubricId).Scan(&rubricCourseID)
		if err == sql.ErrNoRows {
			return nil, errors.New("rubric not found")
		}
		if err != nil {
			return nil, err
		}
		if rubricCourseID != courseID {
			return nil, errors.New("rubric does not belong to this course")
		}
	} else {
		return nil, errors.New("rubric is required for assignments")
	}

	// Update assignment
	_, err = s.db.DB.Exec(`
		UPDATE assignments 
		SET name = ?, description = ?, rubric_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Name, req.Description, req.RubricId, req.Id)
	if err != nil {
		return nil, err
	}

	// Get updated assignment
	assignment, err := s.getAssignmentByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.AssignmentResponse{
		Assignment: assignment,
		Message:    "Assignment updated successfully",
	}, nil
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, req *pb.DeleteAssignmentRequest) (*pb.DeleteAssignmentResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can delete assignments
	if userRole != "instructor" {
		return nil, errors.New("only instructors can delete assignments")
	}

	// Check if assignment exists and get course info
	var courseID int64
	err := s.db.DB.QueryRow("SELECT course_id FROM assignments WHERE id = ?", req.Id).Scan(&courseID)
	if err == sql.ErrNoRows {
		return nil, errors.New("assignment not found")
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
		return nil, errors.New("only the course instructor can delete this assignment")
	}

	// Delete assignment (will cascade to grades due to foreign key constraints)
	_, err = s.db.DB.Exec("DELETE FROM assignments WHERE id = ?", req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteAssignmentResponse{
		Message: "Assignment deleted successfully",
	}, nil
}

func (s *AssignmentService) getAssignmentByID(assignmentID int64) (*pb.Assignment, error) {
	var assignment pb.Assignment
	var createdAt, updatedAt time.Time
	var rubricID sql.NullInt64
	var creatorName sql.NullString

	err := s.db.DB.QueryRow(`
		SELECT a.id, a.course_id, a.name, a.description, a.rubric_id, 
		       a.created_by, u.name, a.created_at, a.updated_at
		FROM assignments a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.id = ?
	`, assignmentID).Scan(&assignment.Id, &assignment.CourseId, &assignment.Name, 
		&assignment.Description, &rubricID,
		&assignment.CreatedBy, &creatorName, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	assignment.CreatedAt = timestamppb.New(createdAt)
	assignment.UpdatedAt = timestamppb.New(updatedAt)

	if rubricID.Valid {
		assignment.RubricId = rubricID.Int64
		// Get rubric details
		rubric, err := s.getRubricByID(rubricID.Int64)
		if err == nil {
			assignment.Rubric = rubric
		}
	}

	if creatorName.Valid {
		assignment.CreatorName = creatorName.String
	}

	return &assignment, nil
}

func (s *AssignmentService) getRubricByID(rubricID int64) (*pb.Rubric, error) {
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

	rubric.CreatedAt = timestamppb.New(createdAt)
	rubric.UpdatedAt = timestamppb.New(updatedAt)

	if creatorName.Valid {
		rubric.CreatorName = creatorName.String
	}

	// Parse JSON arrays
	if err := json.Unmarshal([]byte(criteriaJSON), &rubric.Criteria); err != nil {
		return nil, err
	}
	
	if err := json.Unmarshal([]byte(weightsJSON), &rubric.Weights); err != nil {
		return nil, err
	}
	
	return &rubric, nil
}
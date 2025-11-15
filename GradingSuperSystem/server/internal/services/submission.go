package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/talytics/server/internal/database"
	pb "github.com/talytics/server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SubmissionService struct {
	pb.UnimplementedSubmissionServiceServer
	db *database.Database
}

func NewSubmissionService(db *database.Database) *SubmissionService {
	return &SubmissionService{db: db}
}

func (s *SubmissionService) UploadSubmission(ctx context.Context, req *pb.UploadSubmissionRequest) (*pb.SubmissionResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can upload submissions
	if userRole != "instructor" {
		return nil, errors.New("only instructors can upload submissions")
	}

	// Check if assignment exists and user is instructor of the course
	var courseID int64
	err := s.db.DB.QueryRow(`
		SELECT a.course_id FROM assignments a
		WHERE a.id = ?
	`, req.AssignmentId).Scan(&courseID)
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
		return nil, errors.New("only the course instructor can upload submissions")
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join(".", "uploads", fmt.Sprintf("assignment_%d", req.AssignmentId))
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, err
	}

	// Save file
	fileName := fmt.Sprintf("%s_%s.pdf", req.StudentId, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(uploadsDir, fileName)

	if err := os.WriteFile(filePath, req.FileData, 0644); err != nil {
		return nil, err
	}

	// Insert submission record
	result, err := s.db.DB.Exec(`
		INSERT INTO submissions (assignment_id, student_id, student_name, file_path, file_name, uploaded_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, req.AssignmentId, req.StudentId, req.StudentName, filePath, fileName)
	if err != nil {
		os.Remove(filePath) // Clean up file on error
		return nil, err
	}

	submissionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	submission, err := s.getSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	return &pb.SubmissionResponse{
		Submission: submission,
		Message:    "Submission uploaded successfully",
	}, nil
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, req *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Check if user has access to this assignment (is member of the course)
	var courseID int64
	err := s.db.DB.QueryRow(`
		SELECT a.course_id FROM assignments a
		WHERE a.id = ?
	`, req.AssignmentId).Scan(&courseID)
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

	// Get submissions for this assignment
	rows, err := s.db.DB.Query(`
		SELECT id, assignment_id, student_id, student_name, file_path, file_name, uploaded_at
		FROM submissions
		WHERE assignment_id = ?
		ORDER BY student_name ASC
	`, req.AssignmentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*pb.Submission
	for rows.Next() {
		var submission pb.Submission
		var uploadedAt time.Time

		err := rows.Scan(&submission.Id, &submission.AssignmentId, &submission.StudentId,
			&submission.StudentName, &submission.FilePath, &submission.FileName, &uploadedAt)
		if err != nil {
			return nil, err
		}

		submission.UploadedAt = timestamppb.New(uploadedAt)
		submissions = append(submissions, &submission)
	}

	return &pb.ListSubmissionsResponse{
		Submissions: submissions,
	}, nil
}

func (s *SubmissionService) GetSubmission(ctx context.Context, req *pb.GetSubmissionRequest) (*pb.SubmissionResponse, error) {
	userID := ctx.Value("user_id").(int64)

	submission, err := s.getSubmissionByID(req.Id)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this submission's assignment
	var courseID int64
	err = s.db.DB.QueryRow(`
		SELECT a.course_id FROM assignments a
		WHERE a.id = ?
	`, submission.AssignmentId).Scan(&courseID)
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

	return &pb.SubmissionResponse{
		Submission: submission,
		Message:    "Submission retrieved successfully",
	}, nil
}

func (s *SubmissionService) GetSubmissionFile(ctx context.Context, req *pb.GetSubmissionRequest) (*pb.SubmissionFileResponse, error) {
	userID := ctx.Value("user_id").(int64)

	submission, err := s.getSubmissionByID(req.Id)
	if err != nil {
		return nil, err
	}

	// Check if user has access
	var courseID int64
	err = s.db.DB.QueryRow(`
		SELECT a.course_id FROM assignments a
		WHERE a.id = ?
	`, submission.AssignmentId).Scan(&courseID)
	if err != nil {
		return nil, err
	}

	var memberCount int
	err = s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, courseID, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, errors.New("access denied")
	}

	// Read file
	fileData, err := os.ReadFile(submission.FilePath)
	if err != nil {
		return nil, errors.New("failed to read submission file")
	}

	return &pb.SubmissionFileResponse{
		FileData: fileData,
		FileName: submission.FileName,
	}, nil
}

func (s *SubmissionService) DeleteSubmission(ctx context.Context, req *pb.DeleteSubmissionRequest) (*pb.DeleteSubmissionResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can delete submissions
	if userRole != "instructor" {
		return nil, errors.New("only instructors can delete submissions")
	}

	submission, err := s.getSubmissionByID(req.Id)
	if err != nil {
		return nil, err
	}

	// Check if user is instructor of the course
	var courseID, instructorID int64
	err = s.db.DB.QueryRow(`
		SELECT a.course_id, c.instructor_id 
		FROM assignments a
		JOIN courses c ON a.course_id = c.id
		WHERE a.id = ?
	`, submission.AssignmentId).Scan(&courseID, &instructorID)
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can delete submissions")
	}

	// Delete file
	os.Remove(submission.FilePath)

	// Delete database record
	_, err = s.db.DB.Exec("DELETE FROM submissions WHERE id = ?", req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteSubmissionResponse{
		Message: "Submission deleted successfully",
	}, nil
}

func (s *SubmissionService) getSubmissionByID(submissionID int64) (*pb.Submission, error) {
	var submission pb.Submission
	var uploadedAt time.Time

	err := s.db.DB.QueryRow(`
		SELECT id, assignment_id, student_id, student_name, file_path, file_name, uploaded_at
		FROM submissions
		WHERE id = ?
	`, submissionID).Scan(&submission.Id, &submission.AssignmentId, &submission.StudentId,
		&submission.StudentName, &submission.FilePath, &submission.FileName, &uploadedAt)
	if err != nil {
		return nil, err
	}

	submission.UploadedAt = timestamppb.New(uploadedAt)
	return &submission, nil
}

// Helper function to handle streaming file upload (for future use)
func (s *SubmissionService) saveUploadedFile(file io.Reader, destPath string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return err
}

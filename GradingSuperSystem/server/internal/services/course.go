package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"github.com/talytics/server/internal/database"
	pb "github.com/talytics/server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CourseService struct {
	pb.UnimplementedCourseServiceServer
	db         *database.Database
	userService *UserService
}

func NewCourseService(db *database.Database, userService *UserService) *CourseService {
	return &CourseService{
		db:          db,
		userService: userService,
	}
}

func (s *CourseService) CreateCourse(ctx context.Context, req *pb.CreateCourseRequest) (*pb.CourseResponse, error) {
	// Get user from context (assuming middleware sets it)
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)
	
	// Only instructors can create courses
	if userRole != "instructor" {
		return nil, errors.New("only instructors can create courses")
	}

	// Validate input
	if req.Name == "" || req.Code == "" {
		return nil, errors.New("course name and code are required")
	}

	// Check if course code already exists
	var count int
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM courses WHERE code = ?", req.Code).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("course with this code already exists")
	}

	// Generate join code
	joinCode, err := generateCourseJoinCode()
	if err != nil {
		return nil, err
	}

	// Insert course
	result, err := s.db.DB.Exec(`
		INSERT INTO courses (name, code, join_code, instructor_id, description, semester, year, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, req.Name, req.Code, joinCode, userID, req.Description, req.Semester, req.Year)
	if err != nil {
		return nil, err
	}

	courseID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Add creator as course member
	_, err = s.db.DB.Exec(`
		INSERT INTO course_members (course_id, user_id, role, joined_at)
		VALUES (?, ?, 'instructor', CURRENT_TIMESTAMP)
	`, courseID, userID)
	if err != nil {
		return nil, err
	}

	// Get course with members
	course, err := s.getCourseByID(courseID)
	if err != nil {
		return nil, err
	}

	return &pb.CourseResponse{
		Course:  course,
		Message: "Course created successfully",
	}, nil
}

func (s *CourseService) GetCourse(ctx context.Context, req *pb.GetCourseRequest) (*pb.CourseResponse, error) {
	userID := ctx.Value("user_id").(int64)
	
	// Check if user is a member of this course
	var memberCount int
	err := s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, req.Id, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, errors.New("access denied: you are not a member of this course")
	}

	course, err := s.getCourseByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.CourseResponse{
		Course:  course,
		Message: "Course retrieved successfully",
	}, nil
}

func (s *CourseService) ListCourses(ctx context.Context, req *pb.ListCoursesRequest) (*pb.ListCoursesResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Get courses where user is a member
	rows, err := s.db.DB.Query(`
		SELECT c.id, c.name, c.code, c.join_code, c.instructor_id, c.description, c.semester, c.year, c.created_at, c.updated_at
		FROM courses c
		JOIN course_members cm ON c.id = cm.course_id
		WHERE cm.user_id = ?
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*pb.Course
	for rows.Next() {
		var course pb.Course
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(&course.Id, &course.Name, &course.Code, &course.JoinCode, 
			&course.InstructorId, &course.Description, &course.Semester, &course.Year,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		course.CreatedAt = timestamppb.New(createdAt)
		course.UpdatedAt = timestamppb.New(updatedAt)

		// Get members for this course
		members, err := s.getCourseMembers(course.Id)
		if err != nil {
			return nil, err
		}
		course.Members = members

		courses = append(courses, &course)
	}

	return &pb.ListCoursesResponse{
		Courses: courses,
	}, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, req *pb.UpdateCourseRequest) (*pb.CourseResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Check if user is instructor of this course
	if userRole != "instructor" {
		return nil, errors.New("only instructors can update courses")
	}

	var instructorID int64
	err := s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", req.Id).Scan(&instructorID)
	if err == sql.ErrNoRows {
		return nil, errors.New("course not found")
	}
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can update this course")
	}

	// Update course
	_, err = s.db.DB.Exec(`
		UPDATE courses 
		SET name = ?, code = ?, description = ?, semester = ?, year = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Name, req.Code, req.Description, req.Semester, req.Year, req.Id)
	if err != nil {
		return nil, err
	}

	// Get updated course
	course, err := s.getCourseByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.CourseResponse{
		Course:  course,
		Message: "Course updated successfully",
	}, nil
}

func (s *CourseService) JoinCourse(ctx context.Context, req *pb.JoinCourseRequest) (*pb.CourseResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Find course by join code
	var courseID int64
	err := s.db.DB.QueryRow("SELECT id FROM courses WHERE join_code = ?", req.JoinCode).Scan(&courseID)
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid join code")
	}
	if err != nil {
		return nil, err
	}

	// Check if user is already a member
	var memberCount int
	err = s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, courseID, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount > 0 {
		return nil, errors.New("you are already a member of this course")
	}

	// Get user role
	userRole := ctx.Value("user_role").(string)

	// Add user to course
	_, err = s.db.DB.Exec(`
		INSERT INTO course_members (course_id, user_id, role, joined_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, courseID, userID, userRole)
	if err != nil {
		return nil, err
	}

	// Get course
	course, err := s.getCourseByID(courseID)
	if err != nil {
		return nil, err
	}

	return &pb.CourseResponse{
		Course:  course,
		Message: "Successfully joined course",
	}, nil
}

func (s *CourseService) LeaveCourse(ctx context.Context, req *pb.LeaveCourseRequest) (*pb.LeaveCourseResponse, error) {
	userID := ctx.Value("user_id").(int64)

	// Check if user is a member
	var memberCount int
	err := s.db.DB.QueryRow(`
		SELECT COUNT(*) FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, req.CourseId, userID).Scan(&memberCount)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, errors.New("you are not a member of this course")
	}

	// Don't allow course instructor to leave
	var instructorID int64
	err = s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", req.CourseId).Scan(&instructorID)
	if err != nil {
		return nil, err
	}
	if instructorID == userID {
		return nil, errors.New("course instructor cannot leave the course")
	}

	// Remove user from course
	_, err = s.db.DB.Exec(`
		DELETE FROM course_members 
		WHERE course_id = ? AND user_id = ?
	`, req.CourseId, userID)
	if err != nil {
		return nil, err
	}

	return &pb.LeaveCourseResponse{
		Message: "Successfully left course",
	}, nil
}

func (s *CourseService) DeleteCourse(ctx context.Context, req *pb.DeleteCourseRequest) (*pb.DeleteCourseResponse, error) {
	userID := ctx.Value("user_id").(int64)
	userRole := ctx.Value("user_role").(string)

	// Only instructors can delete courses
	if userRole != "instructor" {
		return nil, errors.New("only instructors can delete courses")
	}

	// Check if course exists and verify ownership
	var instructorID int64
	err := s.db.DB.QueryRow("SELECT instructor_id FROM courses WHERE id = ?", req.Id).Scan(&instructorID)
	if err == sql.ErrNoRows {
		return nil, errors.New("course not found")
	}
	if err != nil {
		return nil, err
	}

	if instructorID != userID {
		return nil, errors.New("only the course instructor can delete this course")
	}

	// Delete the course - CASCADE will handle related tables automatically
	// Foreign keys with ON DELETE CASCADE will delete:
	// - course_members
	// - assignments (which cascades to grades)
	// - rubrics
	// - analysis_results
	_, err = s.db.DB.Exec("DELETE FROM courses WHERE id = ?", req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteCourseResponse{
		Message: "Course deleted successfully",
	}, nil
}

func (s *CourseService) getCourseByID(courseID int64) (*pb.Course, error) {
	var course pb.Course
	var createdAt, updatedAt time.Time

	err := s.db.DB.QueryRow(`
		SELECT id, name, code, join_code, instructor_id, description, semester, year, created_at, updated_at
		FROM courses WHERE id = ?
	`, courseID).Scan(&course.Id, &course.Name, &course.Code, &course.JoinCode,
		&course.InstructorId, &course.Description, &course.Semester, &course.Year,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	course.CreatedAt = timestamppb.New(createdAt)
	course.UpdatedAt = timestamppb.New(updatedAt)

	// Get members
	members, err := s.getCourseMembers(courseID)
	if err != nil {
		return nil, err
	}
	course.Members = members

	return &course, nil
}

func (s *CourseService) getCourseMembers(courseID int64) ([]*pb.CourseMember, error) {
	rows, err := s.db.DB.Query(`
		SELECT u.id, u.name, u.email, cm.role, cm.joined_at
		FROM course_members cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.course_id = ?
		ORDER BY cm.joined_at
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*pb.CourseMember
	for rows.Next() {
		var member pb.CourseMember
		var joinedAt time.Time
		
		err := rows.Scan(&member.UserId, &member.Name, &member.Email, &member.Role, &joinedAt)
		if err != nil {
			return nil, err
		}

		member.JoinedAt = timestamppb.New(joinedAt)
		members = append(members, &member)
	}

	return members, nil
}

func generateCourseJoinCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6
	
	bytes := make([]byte, codeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	
	return string(bytes), nil
}
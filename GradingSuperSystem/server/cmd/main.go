package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/talytics/server/internal/database"
	"github.com/talytics/server/internal/middleware"
	"github.com/talytics/server/internal/services"
	pb "github.com/talytics/server/proto"
	"google.golang.org/grpc"
)

const (
	grpcPort = ":50051"
	httpPort = ":5000"
)

func main() {
	// Get the executable's directory to find .env file
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		log.Printf("Looking for .env at: %s", envPath)
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Could not load from exe dir, trying current directory: %v", err)
			godotenv.Load(".env")
		}
	} else {
		godotenv.Load(".env")
	}
	
	// Verify API key loaded
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		log.Printf("✓ Claude API key loaded successfully (length: %d)", len(apiKey))
	} else {
		log.Println("⚠ Warning: CLAUDE_API_KEY not set - AI insights will not work")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Start gRPC server in a goroutine
	go startGRPCServer(db)

	// Start HTTP REST API server
	startHTTPServer(db)
}

func startGRPCServer(db *database.Database) {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcPort, err)
	}

	server := grpc.NewServer()

	// Create services
	userService := services.NewUserService(db)
	courseService := services.NewCourseService(db, userService)
	assignmentService := services.NewAssignmentService(db)
	rubricService := services.NewRubricService(db)
	healthService := services.NewHealthService()

	// Register services
	pb.RegisterUserServiceServer(server, userService)
	pb.RegisterCourseServiceServer(server, courseService)
	pb.RegisterAssignmentServiceServer(server, assignmentService)
	pb.RegisterRubricServiceServer(server, rubricService)
	pb.RegisterHealthServiceServer(server, healthService)

	log.Printf("gRPC server listening on %s", grpcPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

func startHTTPServer(db *database.Database) {
	// Create services
	userService := services.NewUserService(db)
	courseService := services.NewCourseService(db, userService)
	assignmentService := services.NewAssignmentService(db)
	rubricService := services.NewRubricService(db)
	submissionService := services.NewSubmissionService(db)
	healthService := services.NewHealthService()

	// Create authentication middleware
	authMiddleware := middleware.NewAuthMiddleware([]byte("your-256-bit-secret"))

	mux := http.NewServeMux()

	// Health endpoint (public)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		resp, _ := healthService.Check(r.Context(), &pb.HealthCheckRequest{})
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    resp.Status,
			"timestamp": resp.Timestamp,
		})
	})

	// Auth endpoints (public)
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pb.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := userService.Register(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pb.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := userService.Login(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Protected endpoints
	mux.HandleFunc("/api/auth/profile", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		token := r.Context().Value("token").(string)
		
		resp, err := userService.GetProfile(r.Context(), &pb.GetProfileRequest{Token: token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/auth/logout", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		token := r.Context().Value("token").(string)
		resp, err := userService.Logout(r.Context(), &pb.LogoutRequest{Token: token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	}))

	// Course endpoints
	mux.HandleFunc("/api/courses/my-courses", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		resp, err := courseService.ListCourses(r.Context(), &pb.ListCoursesRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Transform protobuf response to match frontend expectations
		courses := make([]map[string]interface{}, 0)
		for _, course := range resp.Courses {
			// Get user's role in this course
			userID := r.Context().Value("user_id").(int64)
			userRole := "member"
			for _, member := range course.Members {
				if member.UserId == userID {
					userRole = member.Role
					break
				}
			}
			
			courseMap := map[string]interface{}{
				"id":          course.Id,
				"name":        course.Name,
				"code":        course.Code,
				"join_code":   course.JoinCode,
				"description": course.Description,
				"semester":    course.Semester,
				"year":        course.Year,
				"created_at":  course.CreatedAt.AsTime(),
				"updated_at":  course.UpdatedAt.AsTime(),
				"role":        userRole,
				"assignment_count": 0, // TODO: Get actual count
				"member_count":     len(course.Members),
			}
			courses = append(courses, courseMap)
		}
		
		response := map[string]interface{}{
			"courses": courses,
		}
		json.NewEncoder(w).Encode(response)
	}))

	mux.HandleFunc("/api/courses/join", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Handle both course_code and join_code for compatibility
		joinCode := requestBody["join_code"]
		if joinCode == "" {
			joinCode = requestBody["course_code"]
		}
		
		if joinCode == "" {
			http.Error(w, "join_code or course_code is required", http.StatusBadRequest)
			return
		}

		req := &pb.JoinCourseRequest{JoinCode: joinCode}
		resp, err := courseService.JoinCourse(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Transform response to match frontend expectations
		course := resp.Course
		courseData := map[string]interface{}{
			"id":          course.Id,
			"name":        course.Name,
			"code":        course.Code,
			"join_code":   course.JoinCode,
			"description": course.Description,
			"semester":    course.Semester,
			"year":        course.Year,
			"created_at":  course.CreatedAt.AsTime(),
			"updated_at":  course.UpdatedAt.AsTime(),
			"role":        "ta", // User joining gets TA role
			"assignment_count": 0,
			"member_count":     len(course.Members),
		}
		json.NewEncoder(w).Encode(courseData)
	}))

	// Course creation endpoint
	mux.HandleFunc("/api/courses", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pb.CreateCourseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		resp, err := courseService.CreateCourse(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Transform response to match frontend expectations
		course := resp.Course
		courseData := map[string]interface{}{
			"id":          course.Id,
			"name":        course.Name,
			"code":        course.Code,
			"join_code":   course.JoinCode,
			"description": course.Description,
			"semester":    course.Semester,
			"year":        course.Year,
			"created_at":  course.CreatedAt.AsTime(),
			"updated_at":  course.UpdatedAt.AsTime(),
			"role":        "instructor",
			"assignment_count": 0,
			"member_count":     len(course.Members),
		}
		json.NewEncoder(w).Encode(courseData)
	}))

	mux.HandleFunc("/api/courses/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Extract ID from URL
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/courses/"), "/")
		if len(pathParts) == 0 {
			http.Error(w, "Course ID required", http.StatusBadRequest)
			return
		}

		courseID, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid course ID", http.StatusBadRequest)
			return
		}

		if len(pathParts) >= 2 {
			if pathParts[1] == "assignments" {
				// List assignments for course
				resp, err := assignmentService.ListAssignments(r.Context(), &pb.ListAssignmentsRequest{CourseId: courseID})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(resp)
				return
			} else if pathParts[1] == "members" {
				// Get course members
				resp, err := courseService.GetCourse(r.Context(), &pb.GetCourseRequest{Id: courseID})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				membersResponse := map[string]interface{}{
					"members": resp.Course.Members,
				}
				json.NewEncoder(w).Encode(membersResponse)
				return
			}
		}

		switch r.Method {
		case "GET":
			resp, err := courseService.GetCourse(r.Context(), &pb.GetCourseRequest{Id: courseID})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(resp)

		case "PUT":
			var req pb.UpdateCourseRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			req.Id = courseID
			
			resp, err := courseService.UpdateCourse(r.Context(), &req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(resp)

		case "DELETE":
			resp, err := courseService.DeleteCourse(r.Context(), &pb.DeleteCourseRequest{Id: courseID})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"message": resp.Message})
		}
	}))

	// Rubric endpoints
	mux.HandleFunc("/api/rubrics", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.Method {
		case "POST":
			var req pb.CreateRubricRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("Error decoding rubric request: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			log.Printf("Creating rubric: course_id=%d, name=%s, criteria_count=%d", req.CourseId, req.Name, len(req.Criteria))
			
			resp, err := rubricService.CreateRubric(r.Context(), &req)
			if err != nil {
				log.Printf("Error creating rubric: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Rubric created successfully: id=%d", resp.Rubric.Id)
			json.NewEncoder(w).Encode(resp)
		case "PUT":
			var requestBody struct {
				Id             int64     `json:"id"`
				Name           string    `json:"name"`
				Criteria       []string  `json:"criteria"`
				Weights        []float64 `json:"weights"`
				ForceRegrading bool      `json:"force_regrading"`
			}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			req := &pb.UpdateRubricRequest{
				Id:       requestBody.Id,
				Name:     requestBody.Name,
				Criteria: requestBody.Criteria,
				Weights:  requestBody.Weights,
			}
			
			resp, err := rubricService.UpdateRubric(r.Context(), req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			// If force_regrading is true, mark all grades for regrading
			if requestBody.ForceRegrading {
				if err := rubricService.MarkGradesForRegrading(requestBody.Id); err != nil {
					log.Printf("Error marking grades for regrading: %v", err)
				} else {
					log.Printf("Marked grades for regrading for rubric %d", requestBody.Id)
				}
			}
			
			json.NewEncoder(w).Encode(resp)
		}
	}))
	
	// AI Rubric Suggestion and individual rubric endpoints
	mux.HandleFunc("/api/rubrics/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/rubrics/")
		parts := strings.Split(path, "/")
		
		if len(parts) >= 2 && parts[1] == "ai-suggest" && r.Method == "POST" {
			rubricID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid rubric ID", http.StatusBadRequest)
				return
			}
			handleAIRubricSuggest(w, r, rubricID, db)
			return
		}
		
		// Handle PUT /api/rubrics/{id} - Update rubric
		if len(parts) >= 1 && r.Method == "PUT" {
			rubricID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid rubric ID", http.StatusBadRequest)
				return
			}
			
			var requestBody struct {
				Name           string    `json:"name"`
				Criteria       []string  `json:"criteria"`
				Weights        []float64 `json:"weights"`
				ForceRegrading bool      `json:"force_regrading"`
			}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			req := &pb.UpdateRubricRequest{
				Id:       rubricID,
				Name:     requestBody.Name,
				Criteria: requestBody.Criteria,
				Weights:  requestBody.Weights,
			}
			
			resp, err := rubricService.UpdateRubric(r.Context(), req)
			if err != nil {
				log.Printf("Error updating rubric %d: %v", rubricID, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			// If force_regrading is true, mark all grades for regrading
			if requestBody.ForceRegrading {
				if err := rubricService.MarkGradesForRegrading(rubricID); err != nil {
					log.Printf("Error marking grades for regrading: %v", err)
				} else {
					log.Printf("Marked grades for regrading for rubric %d", rubricID)
				}
			}
			
			json.NewEncoder(w).Encode(resp)
			return
		}
		
		http.Error(w, "Not found", http.StatusNotFound)
	}))

	// Assignment endpoints
	mux.HandleFunc("/api/assignments", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.Method {
		case "POST":
			var req pb.CreateAssignmentRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				log.Printf("Error decoding assignment request: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			log.Printf("Creating assignment: course_id=%d, name=%s, rubric_id=%d", req.CourseId, req.Name, req.RubricId)
			
			resp, err := assignmentService.CreateAssignment(r.Context(), &req)
			if err != nil {
				log.Printf("Error creating assignment: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))

	// PDF Submission endpoints
	mux.HandleFunc("/api/assignments/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/assignments/")
		pathParts := strings.Split(path, "/")
		
		if len(pathParts) >= 2 && pathParts[1] == "submissions" {
			assignmentID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
				return
			}

			switch r.Method {
			case "POST":
				// Handle PDF uploads
				handlePDFUpload(w, r, assignmentID, submissionService)
				return
			case "GET":
				// List submissions for assignment
				handleListSubmissions(w, r, assignmentID, submissionService)
				return
			}
		}
		
		if len(pathParts) >= 2 && pathParts[1] == "analytics" {
			assignmentID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
				return
			}

			if r.Method == "GET" {
				handleGetAnalytics(w, r, assignmentID, db)
				return
			}
		}
		
		if len(pathParts) >= 2 && pathParts[1] == "ai-insights" {
			assignmentID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
				return
			}

			if r.Method == "POST" {
				handleGetAIInsights(w, r, assignmentID, db)
				return
			}
		}
		
		http.Error(w, "Not found", http.StatusNotFound)
	}))

	// Submissions endpoints
	mux.HandleFunc("/api/submissions/assignment/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Extract assignment ID from URL
		path := strings.TrimPrefix(r.URL.Path, "/api/submissions/assignment/")
		assignmentID, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {
			handleListSubmissions(w, r, assignmentID, submissionService)
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/submissions/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		// Handle individual submission file download
		path := strings.TrimPrefix(r.URL.Path, "/api/submissions/")
		pathParts := strings.Split(path, "/")
		
		if len(pathParts) >= 2 && pathParts[1] == "file" {
			submissionID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid submission ID", http.StatusBadRequest)
				return
			}
			
			if r.Method == "GET" {
				handleGetSubmissionFile(w, r, submissionID, submissionService)
				return
			}
		}
		
		http.Error(w, "Not found", http.StatusNotFound)
	}))

	// Grade endpoints
	mux.HandleFunc("/api/grades", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		log.Printf("Grade endpoint hit: method=%s", r.Method)
		
		if r.Method == "POST" {
			handleSubmitGrade(w, r, db)
			return
		}
		
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/grades/submission/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := strings.TrimPrefix(r.URL.Path, "/api/grades/submission/")
		submissionID, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid submission ID", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {
			handleGetGradeBySubmission(w, r, submissionID, db)
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/grades/assignment/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := strings.TrimPrefix(r.URL.Path, "/api/grades/assignment/")
		assignmentID, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid assignment ID", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {
			handleGetGradesByAssignment(w, r, assignmentID, db)
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/grades/needs-regrading", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		userID := r.Context().Value("user_id").(int64)
		
		// Get all grades by this TA that need regrading
		rows, err := db.DB.Query(`
			SELECT g.id, g.assignment_id, g.submission_id, g.student_id, g.total_score, g.updated_at,
			       a.name as assignment_name, s.student_name
			FROM grades g
			JOIN assignments a ON g.assignment_id = a.id
			JOIN submissions s ON g.submission_id = s.id
			WHERE g.grader_id = ? AND g.needs_regrading = 1
			ORDER BY g.updated_at DESC
		`, userID)
		
		if err != nil {
			log.Printf("Error fetching regrading grades: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		
		var grades []map[string]interface{}
		for rows.Next() {
			var id, assignmentID, submissionID int64
			var studentID, assignmentName, studentName, updatedAt string
			var totalScore float64
			
			err := rows.Scan(&id, &assignmentID, &submissionID, &studentID, &totalScore, &updatedAt, &assignmentName, &studentName)
			if err != nil {
				continue
			}
			
			grade := map[string]interface{}{
				"id":              id,
				"assignment_id":   assignmentID,
				"submission_id":   submissionID,
				"student_id":      studentID,
				"student_name":    studentName,
				"assignment_name": assignmentName,
				"total_score":     totalScore,
				"updated_at":      updatedAt,
			}
			
			grades = append(grades, grade)
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"grades": grades,
		})
	}))

	// Grade comments endpoints
	mux.HandleFunc("/api/grades/", authHandler(authMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		path := strings.TrimPrefix(r.URL.Path, "/api/grades/")
		parts := strings.Split(path, "/")
		
		if len(parts) < 2 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		
		gradeID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid grade ID", http.StatusBadRequest)
			return
		}
		
		if parts[1] == "comments" {
			if r.Method == "GET" {
				handleGetComments(w, r, gradeID, db)
				return
			} else if r.Method == "POST" {
				handlePostComment(w, r, gradeID, db)
				return
			}
		}
		
		http.Error(w, "Not found", http.StatusNotFound)
	}))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Add logging middleware
	loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		mux.ServeHTTP(w, r)
	})

	log.Printf("HTTP REST API server listening on %s", httpPort)
	log.Printf("TAlytics Go server with authentication and course management is ready!")
	
	if err := http.ListenAndServe(httpPort, c.Handler(loggingHandler)); err != nil {
		log.Fatalf("Failed to serve HTTP server: %v", err)
	}
}

// Helper function to wrap handlers with authentication
func authHandler(authMiddleware *middleware.AuthMiddleware, next http.HandlerFunc) http.HandlerFunc {
	return authMiddleware.AuthenticateHTTP(http.HandlerFunc(next)).ServeHTTP
}

// Handle PDF upload for assignments
func handlePDFUpload(w http.ResponseWriter, r *http.Request, assignmentID int64, submissionService *services.SubmissionService) {
	// Parse multipart form (32MB max)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Try to get file from "file" field (single upload)
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		// Fall back to "submissions" field (batch upload)
		files = r.MultipartForm.File["submissions"]
	}
	
	if len(files) == 0 {
		log.Printf("No files found in form. Available fields: %v", r.MultipartForm.File)
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	log.Printf("Processing %d file(s) for assignment %d", len(files), assignmentID)

	uploadedCount := 0

	// Process each uploaded file
	for _, fileHeader := range files {
		// Validate file type (PDF only)
		if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf") {
			http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Read file data
		fileData := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				fileData = append(fileData, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Extract student info from form or filename
		studentID := r.FormValue("student_id")
		studentName := r.FormValue("student_name")
		
		if studentID == "" {
			// Try to extract from filename
			name := strings.TrimSuffix(fileHeader.Filename, ".pdf")
			parts := strings.Split(name, "_")
			if len(parts) >= 2 {
				studentID = parts[0]
				studentName = strings.Join(parts[1:], " ")
			} else {
				studentID = name
				studentName = name
			}
		}

		// Upload to service
		_, err = submissionService.UploadSubmission(r.Context(), &pb.UploadSubmissionRequest{
			AssignmentId: assignmentID,
			StudentId:    studentID,
			StudentName:  studentName,
			FileData:     fileData,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		uploadedCount++
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": "Files uploaded successfully",
		"count":   uploadedCount,
	}
	json.NewEncoder(w).Encode(response)
}

// Handle listing submissions for an assignment
func handleListSubmissions(w http.ResponseWriter, r *http.Request, assignmentID int64, submissionService *services.SubmissionService) {
	w.Header().Set("Content-Type", "application/json")
	
	resp, err := submissionService.ListSubmissions(r.Context(), &pb.ListSubmissionsRequest{
		AssignmentId: assignmentID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

// Handle getting a submission file
func handleGetSubmissionFile(w http.ResponseWriter, r *http.Request, submissionID int64, submissionService *services.SubmissionService) {
	resp, err := submissionService.GetSubmissionFile(r.Context(), &pb.GetSubmissionRequest{
		Id: submissionID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\""+resp.FileName+"\"")
	w.Write(resp.FileData)
}

// Handle submitting a grade
func handleSubmitGrade(w http.ResponseWriter, r *http.Request, db *database.Database) {
	userID := r.Context().Value("user_id").(int64)
	
	var req struct {
		AssignmentID  int64              `json:"assignment_id"`
		SubmissionID  int64              `json:"submission_id"`
		StudentID     string             `json:"student_id"`
		RubricScores  map[string]float64 `json:"rubric_scores"`
		TotalScore    float64            `json:"total_score"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding grade request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Convert rubric scores to JSON
	rubricScoresJSON, err := json.Marshal(req.RubricScores)
	if err != nil {
		http.Error(w, "Error encoding rubric scores", http.StatusInternalServerError)
		return
	}
	
	// Check if grade already exists
	var existingID int64
	err = db.DB.QueryRow(`
		SELECT id FROM grades WHERE submission_id = ?
	`, req.SubmissionID).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		// Insert new grade
		result, err := db.DB.Exec(`
			INSERT INTO grades (assignment_id, submission_id, student_id, grader_id, rubric_scores, total_score, needs_regrading, graded_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, req.AssignmentID, req.SubmissionID, req.StudentID, userID, string(rubricScoresJSON), req.TotalScore)
		
		if err != nil {
			log.Printf("Error inserting grade: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		gradeID, _ := result.LastInsertId()
		existingID = gradeID
		log.Printf("Grade created: id=%d, submission=%d, score=%.2f", gradeID, req.SubmissionID, req.TotalScore)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// Update existing grade and clear needs_regrading flag
		_, err = db.DB.Exec(`
			UPDATE grades SET rubric_scores = ?, total_score = ?, needs_regrading = 0, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, string(rubricScoresJSON), req.TotalScore, existingID)
		
		if err != nil {
			log.Printf("Error updating grade: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		log.Printf("Grade updated: id=%d, submission=%d, score=%.2f", existingID, req.SubmissionID, req.TotalScore)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Grade submitted successfully",
		"grade": map[string]interface{}{
			"id": existingID,
		},
	})
}

// Handle getting a grade by submission
func handleGetGradeBySubmission(w http.ResponseWriter, r *http.Request, submissionID int64, db *database.Database) {
	var grade struct {
		ID           int64              `json:"id"`
		AssignmentID int64              `json:"assignment_id"`
		SubmissionID int64              `json:"submission_id"`
		StudentID    string             `json:"student_id"`
		GraderID     int64              `json:"grader_id"`
		GraderName   string             `json:"grader_name"`
		RubricScores map[string]float64 `json:"rubric_scores"`
		TotalScore   float64            `json:"total_score"`
		GradedAt     string             `json:"graded_at"`
	}
	
	var rubricScoresJSON string
	var graderName sql.NullString
	
	err := db.DB.QueryRow(`
		SELECT g.id, g.assignment_id, g.submission_id, g.student_id, g.grader_id, u.name, g.rubric_scores, g.total_score, g.graded_at
		FROM grades g
		LEFT JOIN users u ON g.grader_id = u.id
		WHERE g.submission_id = ?
	`, submissionID).Scan(&grade.ID, &grade.AssignmentID, &grade.SubmissionID, &grade.StudentID, &grade.GraderID, &graderName, &rubricScoresJSON, &grade.TotalScore, &grade.GradedAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Grade not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching grade: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if graderName.Valid {
		grade.GraderName = graderName.String
	}
	
	// Parse rubric scores JSON
	if err := json.Unmarshal([]byte(rubricScoresJSON), &grade.RubricScores); err != nil {
		http.Error(w, "Error parsing rubric scores", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grade": grade,
	})
}

// Handle getting all grades for an assignment
func handleGetGradesByAssignment(w http.ResponseWriter, r *http.Request, assignmentID int64, db *database.Database) {
	rows, err := db.DB.Query(`
		SELECT g.id, g.assignment_id, g.submission_id, g.student_id, g.grader_id, u.name, g.rubric_scores, g.total_score, g.graded_at,
		       s.student_name, s.file_name
		FROM grades g
		LEFT JOIN users u ON g.grader_id = u.id
		LEFT JOIN submissions s ON g.submission_id = s.id
		WHERE g.assignment_id = ?
		ORDER BY g.graded_at DESC
	`, assignmentID)
	
	if err != nil {
		log.Printf("Error fetching grades: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var grades []map[string]interface{}
	
	for rows.Next() {
		var id, assignmentID, submissionID, graderID int64
		var studentID, rubricScoresJSON, gradedAt, fileName string
		var totalScore float64
		var graderName, studentName sql.NullString
		
		err := rows.Scan(&id, &assignmentID, &submissionID, &studentID, &graderID, &graderName, &rubricScoresJSON, &totalScore, &gradedAt, &studentName, &fileName)
		if err != nil {
			continue
		}
		
		var rubricScores map[string]float64
		json.Unmarshal([]byte(rubricScoresJSON), &rubricScores)
		
		grade := map[string]interface{}{
			"id":            id,
			"assignment_id": assignmentID,
			"submission_id": submissionID,
			"student_id":    studentID,
			"grader_id":     graderID,
			"rubric_scores": rubricScores,
			"total_score":   totalScore,
			"graded_at":     gradedAt,
			"file_name":     fileName,
		}
		
		if graderName.Valid {
			grade["grader_name"] = graderName.String
		}
		if studentName.Valid {
			grade["student_name"] = studentName.String
		}
		
		grades = append(grades, grade)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grades": grades,
	})
}

// Handle posting a comment on a grade
func handlePostComment(w http.ResponseWriter, r *http.Request, gradeID int64, db *database.Database) {
	var req struct {
		Message  string `json:"message"`
		ParentID *int64 `json:"parent_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get user ID from context
	userID := r.Context().Value("user_id").(int64)
	
	var result sql.Result
	var err error
	
	if req.ParentID != nil {
		result, err = db.DB.Exec(`
			INSERT INTO grade_comments (grade_id, user_id, message, parent_id)
			VALUES (?, ?, ?, ?)
		`, gradeID, userID, req.Message, *req.ParentID)
	} else {
		result, err = db.DB.Exec(`
			INSERT INTO grade_comments (grade_id, user_id, message)
			VALUES (?, ?, ?)
		`, gradeID, userID, req.Message)
	}
	
	if err != nil {
		log.Printf("Error posting comment: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	commentID, _ := result.LastInsertId()
	
	// Fetch the created comment with user info
	var comment struct {
		ID        int64  `json:"id"`
		GradeID   int64  `json:"grade_id"`
		UserID    int64  `json:"user_id"`
		UserName  string `json:"user_name"`
		UserRole  string `json:"user_role"`
		Message   string `json:"message"`
		ParentID  *int64 `json:"parent_id"`
		CreatedAt string `json:"created_at"`
	}
	
	var parentID sql.NullInt64
	err = db.DB.QueryRow(`
		SELECT c.id, c.grade_id, c.user_id, u.name, u.role, c.message, c.parent_id, c.created_at
		FROM grade_comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`, commentID).Scan(&comment.ID, &comment.GradeID, &comment.UserID, &comment.UserName, &comment.UserRole, &comment.Message, &parentID, &comment.CreatedAt)
	
	if err != nil {
		log.Printf("Error fetching comment: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comment": comment,
	})
}

// Handle getting all comments for a grade
func handleGetComments(w http.ResponseWriter, r *http.Request, gradeID int64, db *database.Database) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.grade_id, c.user_id, u.name, u.role, c.message, c.parent_id, c.created_at
		FROM grade_comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.grade_id = ?
		ORDER BY c.created_at ASC
	`, gradeID)
	
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var comments []map[string]interface{}
	
	for rows.Next() {
		var id, gradeID, userID int64
		var userName, userRole, message, createdAt string
		var parentID sql.NullInt64
		
		err := rows.Scan(&id, &gradeID, &userID, &userName, &userRole, &message, &parentID, &createdAt)
		if err != nil {
			continue
		}
		
		comment := map[string]interface{}{
			"id":         id,
			"grade_id":   gradeID,
			"user_id":    userID,
			"user_name":  userName,
			"user_role":  userRole,
			"message":    message,
			"created_at": createdAt,
		}
		
		if parentID.Valid {
			comment["parent_id"] = parentID.Int64
		}
		
		comments = append(comments, comment)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comments": comments,
	})
}

// Handle getting analytics for an assignment
func handleGetAnalytics(w http.ResponseWriter, r *http.Request, assignmentID int64, db *database.Database) {
	// Fetch all grades for the assignment with rubric info
	rows, err := db.DB.Query(`
		SELECT g.id, g.rubric_scores, g.total_score, g.grader_id, u.name
		FROM grades g
		LEFT JOIN users u ON g.grader_id = u.id
		WHERE g.assignment_id = ?
	`, assignmentID)
	
	if err != nil {
		log.Printf("Error fetching analytics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	type GradeData struct {
		ID           int64
		RubricScores map[string]float64
		TotalScore   float64
		GraderID     int64
		GraderName   string
	}
	
	var grades []GradeData
	var totalScore float64
	var highestScore float64
	var lowestScore float64 = 999999
	
	// Map to track grader stats
	graderStats := make(map[int64]struct {
		Name           string
		GradeCount     int
		TotalAverage   float64
		CriteriaSum    map[string]float64
		CriteriaCount  map[string]int
	})
	
	// Map to track criteria stats
	criteriaStats := make(map[string]struct {
		Sum   float64
		Count int
		Min   float64
		Max   float64
		Values []float64
	})
	
	for rows.Next() {
		var id, graderID int64
		var rubricScoresJSON, graderName string
		var score float64
		
		err := rows.Scan(&id, &rubricScoresJSON, &score, &graderID, &graderName)
		if err != nil {
			continue
		}
		
		var rubricScores map[string]float64
		json.Unmarshal([]byte(rubricScoresJSON), &rubricScores)
		
		grades = append(grades, GradeData{
			ID:           id,
			RubricScores: rubricScores,
			TotalScore:   score,
			GraderID:     graderID,
			GraderName:   graderName,
		})
		
		totalScore += score
		if score > highestScore {
			highestScore = score
		}
		if score < lowestScore {
			lowestScore = score
		}
		
		// Update grader stats
		if _, exists := graderStats[graderID]; !exists {
			graderStats[graderID] = struct {
				Name           string
				GradeCount     int
				TotalAverage   float64
				CriteriaSum    map[string]float64
				CriteriaCount  map[string]int
			}{
				Name:          graderName,
				GradeCount:    0,
				TotalAverage:  0,
				CriteriaSum:   make(map[string]float64),
				CriteriaCount: make(map[string]int),
			}
		}
		
		stats := graderStats[graderID]
		stats.GradeCount++
		stats.TotalAverage += score
		
		for criterionIdx, criterionScore := range rubricScores {
			stats.CriteriaSum[criterionIdx] += criterionScore
			stats.CriteriaCount[criterionIdx]++
			
			// Update criteria stats
			if _, exists := criteriaStats[criterionIdx]; !exists {
				criteriaStats[criterionIdx] = struct {
					Sum   float64
					Count int
					Min   float64
					Max   float64
					Values []float64
				}{
					Sum:   0,
					Count: 0,
					Min:   999999,
					Max:   0,
					Values: []float64{},
				}
			}
			
			cStats := criteriaStats[criterionIdx]
			cStats.Sum += criterionScore
			cStats.Count++
			cStats.Values = append(cStats.Values, criterionScore)
			if criterionScore < cStats.Min {
				cStats.Min = criterionScore
			}
			if criterionScore > cStats.Max {
				cStats.Max = criterionScore
			}
			criteriaStats[criterionIdx] = cStats
		}
		
		graderStats[graderID] = stats
	}
	
	if len(grades) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_grades": 0,
			"grades":       []interface{}{},
		})
		return
	}
	
	// Calculate overall average and std deviation
	overallAverage := totalScore / float64(len(grades))
	
	var variance float64
	for _, grade := range grades {
		diff := grade.TotalScore - overallAverage
		variance += diff * diff
	}
	stdDeviation := 0.0
	if len(grades) > 1 {
		stdDeviation = variance / float64(len(grades)-1)
		stdDeviation = float64(int(stdDeviation*100)) / 100 // Round to 2 decimals
	}
	
	// Format grader stats
	var graderStatsList []map[string]interface{}
	for graderID, stats := range graderStats {
		criteriaAverages := make(map[string]float64)
		for criterionIdx, sum := range stats.CriteriaSum {
			criteriaAverages[criterionIdx] = sum / float64(stats.CriteriaCount[criterionIdx])
		}
		
		graderStatsList = append(graderStatsList, map[string]interface{}{
			"grader_id":         graderID,
			"grader_name":       stats.Name,
			"grade_count":       stats.GradeCount,
			"average":           stats.TotalAverage / float64(stats.GradeCount),
			"criteria_averages": criteriaAverages,
		})
	}
	
	// Format criteria stats
	var criteriaStatsList []map[string]interface{}
	for criterionIdx, stats := range criteriaStats {
		average := stats.Sum / float64(stats.Count)
		
		// Calculate std dev for this criterion
		var criterionVariance float64
		for _, val := range stats.Values {
			diff := val - average
			criterionVariance += diff * diff
		}
		criterionStdDev := 0.0
		if len(stats.Values) > 1 {
			criterionStdDev = criterionVariance / float64(len(stats.Values)-1)
		}
		
		criteriaStatsList = append(criteriaStatsList, map[string]interface{}{
			"criterion_index": criterionIdx,
			"average":         average,
			"min":             stats.Min,
			"max":             stats.Max,
			"std_dev":         criterionStdDev,
		})
	}
	
	// Get max score from assignment
	var assignmentMaxScore float64
	db.DB.QueryRow("SELECT max_score FROM assignments WHERE id = ?", assignmentID).Scan(&assignmentMaxScore)
	if assignmentMaxScore == 0 {
		assignmentMaxScore = 100
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_grades":      len(grades),
		"overall_average":   overallAverage,
		"highest_score":     highestScore,
		"lowest_score":      lowestScore,
		"std_deviation":     stdDeviation,
		"max_score":         assignmentMaxScore,
		"grades":            grades,
		"grader_stats":      graderStatsList,
		"criteria_stats":    criteriaStatsList,
	})
}

// Handle getting AI insights for analytics
func handleGetAIInsights(w http.ResponseWriter, r *http.Request, assignmentID int64, db *database.Database) {
	log.Printf("AI Insights request received for assignment %d", assignmentID)
	
	var req struct {
		AnalyticsData map[string]interface{} `json:"analytics_data"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("Analytics data received, size: %d fields", len(req.AnalyticsData))
	
	// Get Claude API key from environment
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	if claudeAPIKey == "" {
		log.Printf("ERROR: Claude API key not configured")
		http.Error(w, "Claude API key not configured", http.StatusInternalServerError)
		return
	}
	log.Printf("Claude API key found (length: %d)", len(claudeAPIKey))
	
	// Prepare analytics summary for Claude
	analyticsJSON, _ := json.MarshalIndent(req.AnalyticsData, "", "  ")
	
	prompt := fmt.Sprintf(`You are an educational assessment expert. Analyze this grading data and provide a BRIEF summary.

Analytics Data:
%s

Provide a concise analysis in ONE of these formats (but don't say which one in the response):
- Option 1: 1-2 short paragraphs (max 4-5 sentences total)
- Option 2: 1 paragraph followed by 2-4 bullet points

Focus on the most important insights only:
- Key grading consistency issues (if any)
- Most significant criteria concerns
- 1+ actionable recommendation

Be specific but concise. Reference actual numbers from the data.`, string(analyticsJSON))
	
	log.Printf("Calling Claude API...")
	// Call Claude API
	insights, err := callClaudeAPI(claudeAPIKey, prompt)
	if err != nil {
		log.Printf("Error calling Claude API: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get AI insights: %v", err), http.StatusInternalServerError)
		return
	}
	
	log.Printf("Claude API response received (length: %d)", len(insights))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"insights": insights,
	})
	log.Printf("AI insights response sent successfully")
}

// Call Claude API
func callClaudeAPI(apiKey, prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"
	
	requestBody := map[string]interface{}{
		"model": "claude-3-haiku-20240307",
		"max_tokens": 2000,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", err
	}
	
	log.Printf("Making request to Claude API...")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	
	log.Printf("Claude API responded with status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Claude API error response: %s", string(body))
		return "", fmt.Errorf("Claude API error: %s - %s", resp.Status, string(body))
	}
	
	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", err
	}
	
	if len(result.Content) == 0 {
		log.Printf("No content in Claude response")
		return "", fmt.Errorf("no content in response")
	}
	
	log.Printf("Successfully extracted insights from Claude")
	return result.Content[0].Text, nil
}

// Handle AI rubric suggestions based on grading data
func handleAIRubricSuggest(w http.ResponseWriter, r *http.Request, rubricID int64, db *database.Database) {
	log.Printf("AI Rubric Suggest request received for rubric %d", rubricID)
	
	// Get Claude API key
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	if claudeAPIKey == "" {
		log.Printf("ERROR: Claude API key not configured")
		http.Error(w, "Claude API key not configured", http.StatusInternalServerError)
		return
	}
	
	// 1. Fetch the current rubric
	var rubricName string
	var rubricCriteria, rubricWeights string
	err := db.DB.QueryRow(`
		SELECT name, criteria, weights FROM rubrics WHERE id = ?
	`, rubricID).Scan(&rubricName, &rubricCriteria, &rubricWeights)
	
	if err != nil {
		log.Printf("Error fetching rubric: %v", err)
		http.Error(w, "Rubric not found", http.StatusNotFound)
		return
	}
	
	// Parse criteria and weights
	var criteria []string
	var weights []float64
	json.Unmarshal([]byte(rubricCriteria), &criteria)
	json.Unmarshal([]byte(rubricWeights), &weights)
	
	// Build current rubric structure
	currentRubric := make([]map[string]interface{}, len(criteria))
	for i, criterion := range criteria {
		weight := 0.0
		if i < len(weights) {
			weight = weights[i]
		}
		currentRubric[i] = map[string]interface{}{
			"name":       criterion,
			"max_points": weight,
			"index":      i,
		}
	}
	
	// 2. Fetch all assignments using this rubric
	assignmentRows, err := db.DB.Query(`
		SELECT id FROM assignments WHERE rubric_id = ?
	`, rubricID)
	if err != nil {
		log.Printf("Error fetching assignments: %v", err)
		http.Error(w, "Error fetching assignments", http.StatusInternalServerError)
		return
	}
	defer assignmentRows.Close()
	
	var assignmentIDs []int64
	for assignmentRows.Next() {
		var id int64
		assignmentRows.Scan(&id)
		assignmentIDs = append(assignmentIDs, id)
	}
	
	// 3. Fetch analytics data for all assignments
	analyticsData := make([]map[string]interface{}, 0)
	for _, assignmentID := range assignmentIDs {
		// Get grades with rubric scores
		gradeRows, err := db.DB.Query(`
			SELECT g.id, g.rubric_scores, g.total_score, u.name, u.role
			FROM grades g
			JOIN users u ON g.grader_id = u.id
			WHERE g.assignment_id = ?
		`, assignmentID)
		
		if err != nil {
			continue
		}
		
		var scores []float64
		var graderNames []string
		
		for gradeRows.Next() {
			var gradeID int64
			var rubricScoresJSON string
			var totalScore float64
			var graderName, graderRole string
			
			gradeRows.Scan(&gradeID, &rubricScoresJSON, &totalScore, &graderName, &graderRole)
			
			var rubricScores map[string]float64
			json.Unmarshal([]byte(rubricScoresJSON), &rubricScores)
			
			scores = append(scores, totalScore)
			graderNames = append(graderNames, graderName)
		}
		gradeRows.Close()
		
		if len(scores) > 0 {
			analyticsData = append(analyticsData, map[string]interface{}{
				"assignment_id": assignmentID,
				"grade_count":   len(scores),
				"graders":       graderNames,
			})
		}
	}
	
	// 4. Fetch ALL TA comments for assignments using this rubric
	var allComments []map[string]interface{}
	for _, assignmentID := range assignmentIDs {
		commentRows, err := db.DB.Query(`
			SELECT gc.message, u.name, u.role, gc.created_at
			FROM grade_comments gc
			JOIN grades g ON gc.grade_id = g.id
			JOIN users u ON gc.user_id = u.id
			WHERE g.assignment_id = ? AND u.role = 'TA'
			ORDER BY gc.created_at DESC
		`, assignmentID)
		
		if err != nil {
			continue
		}
		
		for commentRows.Next() {
			var message, userName, userRole, createdAt string
			commentRows.Scan(&message, &userName, &userRole, &createdAt)
			
			allComments = append(allComments, map[string]interface{}{
				"message":    message,
				"ta_name":    userName,
				"created_at": createdAt,
			})
		}
		commentRows.Close()
	}
	
	log.Printf("Collected %d TA comments for rubric analysis", len(allComments))
	
	// 5. Build prompt for Claude
	currentRubricJSON, _ := json.MarshalIndent(currentRubric, "", "  ")
	analyticsJSON, _ := json.MarshalIndent(analyticsData, "", "  ")
	commentsJSON, _ := json.MarshalIndent(allComments, "", "  ")
	
	prompt := fmt.Sprintf(`You are an educational assessment expert. Analyze this grading rubric based on actual usage data and suggest improvements.

CURRENT RUBRIC:
%s

GRADING ANALYTICS:
%s

TA COMMENTS (feedback left during grading):
%s

Based on this data, suggest an improved version of the rubric. ALL SUGGESTIONS MUST KEEP THE TOTAL AT 100 POINTS. IF YOU SPLIT A CRITERIA INTO MULTIPLE SUBCRITERIA, THE SUM OF THOSE POINTS MUST BE THE SAME AS THE ORIGINAL. Consider:

YOU CANNOT, I REPEAT CANNOT, SUGGEST THAT YOU INCREASE OR DECREASE THE POINTS A SPECIFIC CRITERIA IS WORTH, ONLY THE CRITERIA ITSELF CAN CHANGE, NOT ITS VALUE.

1. Are criterion names clear and specific enough based on how TAs are using them?
2. Should any criteria be split into multiple criteria for clarity?
3. Should any criteria be merged?
4. Are the point allocations appropriate based on grading patterns?
5. What issues do the TA comments reveal about the rubric?

Return ONLY a valid JSON array with the same structure as the input rubric, but with your suggested improvements. Each criterion should have:
- "name": string (the criterion name/description)
- "max_points": number (the point value)
- "index": number (position in rubric)
- "change_reason": string (brief explanation of what you changed and why)

Example output format:
[
  {
    "name": "Code Quality and Style",
    "max_points": 25,
    "index": 0,
    "change_reason": "Split from original 'Code' criterion for clarity based on TA comments about ambiguity"
  }
]

Return ONLY the JSON array, no other text.`, string(currentRubricJSON), string(analyticsJSON), string(commentsJSON))
	
	// 6. Call Claude API
	log.Printf("Calling Claude API for rubric suggestions...")
	suggestedRubricText, err := callClaudeAPI(claudeAPIKey, prompt)
	if err != nil {
		log.Printf("Error calling Claude API: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}
	
	// 7. Parse the suggested rubric JSON
	var suggestedRubric []map[string]interface{}
	if err := json.Unmarshal([]byte(suggestedRubricText), &suggestedRubric); err != nil {
		log.Printf("Error parsing Claude response as JSON: %v", err)
		// Try to extract JSON from markdown code blocks
		start := strings.Index(suggestedRubricText, "[")
		end := strings.LastIndex(suggestedRubricText, "]")
		if start >= 0 && end > start {
			jsonPart := suggestedRubricText[start : end+1]
			if err := json.Unmarshal([]byte(jsonPart), &suggestedRubric); err != nil {
				log.Printf("Still failed to parse JSON: %v", err)
				http.Error(w, "Failed to parse AI response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to parse AI response", http.StatusInternalServerError)
			return
		}
	}
	
	log.Printf("Successfully generated AI rubric suggestions")
	
	// 8. Return both old and new rubrics
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_rubric":   currentRubric,
		"suggested_rubric": suggestedRubric,
		"rubric_name":      rubricName,
		"rubric_id":        rubricID,
	})
}

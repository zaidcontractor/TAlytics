package api

import (
	"talytics/internal/api/handlers"
	"talytics/internal/api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes all API routes
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Protected routes (require authentication)
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		// Course routes
		courses := api.Group("/courses")
		{
			courses.POST("", handlers.CreateCourse)
			courses.GET("", handlers.GetCourses)
		}

		// Assignment routes
		assignments := api.Group("/assignments")
		{
			assignments.POST("", handlers.CreateAssignment)
			assignments.GET("/course/:course_id", handlers.GetAssignmentsByCourse)
			assignments.GET("/:id", handlers.GetAssignment)
			assignments.POST("/:id/notify-tas", handlers.NotifyTAs)
		}

		// Rubric routes
		rubrics := api.Group("/rubrics")
		{
			rubrics.POST("", handlers.CreateRubric)
			rubrics.POST("/upload", handlers.UploadRubricPDF)
		}

		// Submission routes
		submissions := api.Group("/submissions")
		{
			submissions.POST("/upload", handlers.UploadSubmission)
			submissions.GET("/assigned", handlers.GetAssignedSubmissions)
		}

		// Grading routes
		grading := api.Group("/grade")
		{
			grading.POST("", handlers.GradeSubmission)
			grading.POST("/batch", handlers.BatchGrade)
		}

		// Anomaly detection routes
		anomalies := api.Group("/")
		{
			anomalies.POST("/analyze/:assignment_id", handlers.AnalyzeAnomalies)
			anomalies.GET("/anomalies/:assignment_id", handlers.GetAnomalies)
		}
	}

	return router
}

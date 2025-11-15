package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite3", "./talytics.db")
	if err != nil {
		return nil, err
	}

	database := &Database{DB: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return database, nil
}

func (d *Database) createTables() error {
	queries := []string{
		// Users table for authentication
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('instructor', 'ta')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Courses table
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			code TEXT UNIQUE NOT NULL,
			join_code TEXT UNIQUE NOT NULL,
			instructor_id INTEGER NOT NULL,
			description TEXT,
			semester TEXT,
			year INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (instructor_id) REFERENCES users (id)
		)`,
		// Course members (instructors and TAs)
		`CREATE TABLE IF NOT EXISTS course_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('instructor', 'ta')),
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			UNIQUE(course_id, user_id)
		)`,
		// Assignments table
		`CREATE TABLE IF NOT EXISTS assignments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			due_date DATETIME,
			max_score REAL NOT NULL DEFAULT 100,
			rubric_id INTEGER,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
			FOREIGN KEY (rubric_id) REFERENCES rubrics (id),
			FOREIGN KEY (created_by) REFERENCES users (id)
		)`,
		// Updated rubrics table with course isolation
		`CREATE TABLE IF NOT EXISTS rubrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			course_id INTEGER NOT NULL,
			criteria TEXT NOT NULL,
			weights TEXT NOT NULL,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
			FOREIGN KEY (created_by) REFERENCES users (id)
		)`,
		// Submissions table for uploaded PDFs
		`CREATE TABLE IF NOT EXISTS submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			assignment_id INTEGER NOT NULL,
			student_id TEXT NOT NULL,
			student_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_name TEXT NOT NULL,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (assignment_id) REFERENCES assignments (id) ON DELETE CASCADE
		)`,
		// Updated grades table with assignment linking and rubric scores
		`CREATE TABLE IF NOT EXISTS grades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			assignment_id INTEGER NOT NULL,
			submission_id INTEGER NOT NULL,
			student_id TEXT NOT NULL,
			grader_id INTEGER NOT NULL,
			rubric_scores TEXT NOT NULL,
			total_score REAL NOT NULL,
			needs_regrading INTEGER DEFAULT 0,
			graded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (assignment_id) REFERENCES assignments (id) ON DELETE CASCADE,
			FOREIGN KEY (submission_id) REFERENCES submissions (id) ON DELETE CASCADE,
			FOREIGN KEY (grader_id) REFERENCES users (id)
		)`,
		// Grade comments for instructor-TA conversations
		`CREATE TABLE IF NOT EXISTS grade_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			grade_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (grade_id) REFERENCES grades (id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (parent_id) REFERENCES grade_comments (id) ON DELETE CASCADE
		)`,
		// Updated analysis results with course context
		`CREATE TABLE IF NOT EXISTS analysis_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			assignment_id INTEGER,
			rubric_id INTEGER,
			analysis_type TEXT NOT NULL,
			results TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
			FOREIGN KEY (assignment_id) REFERENCES assignments (id) ON DELETE CASCADE,
			FOREIGN KEY (rubric_id) REFERENCES rubrics (id)
		)`,
		// User sessions for JWT token management
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token_hash TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := d.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
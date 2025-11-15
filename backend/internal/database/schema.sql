-- TAlytics Database Schema

-- Users table: Stores professor, head TA, and grader TA accounts
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('professor', 'head_ta', 'grader_ta')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Courses table: Stores course information
CREATE TABLE IF NOT EXISTS courses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Assignments table: Stores assignment details for each course
CREATE TABLE IF NOT EXISTS assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft', 'open', 'grading', 'completed')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (course_id) REFERENCES courses(id)
);

-- Rubrics table: Stores rubric data in JSON format
CREATE TABLE IF NOT EXISTS rubrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    json_blob TEXT NOT NULL,
    max_points REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id)
);

-- Submissions table: Stores student submission data
CREATE TABLE IF NOT EXISTS submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    student_identifier TEXT NOT NULL,
    text TEXT,
    file_path TEXT,
    graded_status TEXT NOT NULL DEFAULT 'pending' CHECK(graded_status IN ('pending', 'in_progress', 'graded', 'regrade_required')),
    assigned_ta_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id),
    FOREIGN KEY (assigned_ta_id) REFERENCES users(id)
);

-- Grades table: Stores grading results with rubric breakdown
CREATE TABLE IF NOT EXISTS grades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL,
    rubric_id INTEGER NOT NULL,
    json_blob TEXT NOT NULL,
    total_points REAL NOT NULL,
    graded_by INTEGER NOT NULL,
    score REAL NOT NULL,
    feedback TEXT,
    rubric_breakdown TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (submission_id) REFERENCES submissions(id),
    FOREIGN KEY (rubric_id) REFERENCES rubrics(id),
    FOREIGN KEY (graded_by) REFERENCES users(id)
);

-- Anomaly Reports table: Stores anomaly detection results
CREATE TABLE IF NOT EXISTS anomaly_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    json_blob TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reviewed', 'resolved')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id)
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_courses_created_by ON courses(created_by);
CREATE INDEX IF NOT EXISTS idx_assignments_course_id ON assignments(course_id);
CREATE INDEX IF NOT EXISTS idx_rubrics_assignment_id ON rubrics(assignment_id);
CREATE INDEX IF NOT EXISTS idx_submissions_assignment_id ON submissions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_submissions_assigned_ta_id ON submissions(assigned_ta_id);
CREATE INDEX IF NOT EXISTS idx_grades_submission_id ON grades(submission_id);
CREATE INDEX IF NOT EXISTS idx_anomaly_reports_assignment_id ON anomaly_reports(assignment_id);

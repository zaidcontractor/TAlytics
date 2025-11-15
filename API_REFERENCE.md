# TAlytics API Reference

Complete reference for all implemented endpoints (Steps 1-4).

## Base URL
```
http://localhost:8080
```

## Authentication

All endpoints except `/auth/register` and `/auth/login` require JWT token:
```
Authorization: Bearer <token>
```

---

## Auth Endpoints

### POST /auth/register
Register a new user.

**Request**:
```json
{
  "email": "prof@university.edu",
  "password": "securepass123",
  "role": "professor"  // or "head_ta" or "grader_ta"
}
```

**Response**:
```json
{
  "id": 1,
  "email": "prof@university.edu",
  "role": "professor",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### POST /auth/login
Login and receive JWT token.

**Request**:
```json
{
  "email": "prof@university.edu",
  "password": "securepass123"
}
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "prof@university.edu",
    "role": "professor",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Token expires**: 24 hours

---

## Course Endpoints

### POST /courses
Create a new course.

**Authorization**: Any authenticated user

**Request**:
```json
{
  "name": "CS 101: Introduction to Programming"
}
```

**Response**:
```json
{
  "course_id": 1,
  "name": "CS 101: Introduction to Programming",
  "created_by": 1,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### GET /courses
Get all courses created by the authenticated user.

**Authorization**: Any authenticated user

**Response**:
```json
{
  "courses": [
    {
      "id": 1,
      "name": "CS 101: Introduction to Programming",
      "created_at": "2024-01-15T10:30:00Z",
      "assignment_count": 5
    }
  ]
}
```

---

## Assignment Endpoints

### POST /assignments
Create a new assignment for a course.

**Authorization**: Course owner

**Request**:
```json
{
  "course_id": 1,
  "title": "Lab 1: Variables and Loops",
  "description": "Practice loops with factorial implementation"
}
```

**Response**:
```json
{
  "assignment_id": 1,
  "course_id": 1,
  "title": "Lab 1: Variables and Loops",
  "description": "Practice loops with factorial implementation",
  "status": "draft",
  "created_at": "2024-01-15T10:35:00Z"
}
```

**Status values**: `draft`, `open`, `grading`, `completed`

### GET /assignments/:id
Get assignment details with rubric.

**Authorization**: Any authenticated user

**Response**:
```json
{
  "id": 1,
  "course_id": 1,
  "title": "Lab 1: Variables and Loops",
  "description": "Practice loops with factorial implementation",
  "status": "open",
  "created_at": "2024-01-15T10:35:00Z",
  "rubric": {
    "id": 1,
    "assignment_id": 1,
    "json_blob": "{\"criteria\": [...]}",
    "max_points": 100
  }
}
```

### POST /assignments/:id/notify-tas
Distribute submissions to TAs using round-robin algorithm.

**Authorization**: Professor or course owner

**Response**:
```json
{
  "assignment_id": 1,
  "total_submissions": 30,
  "distribution": {
    "2": 10,
    "3": 10,
    "4": 10
  },
  "message": "Assigned 30 submissions to 3 TAs"
}
```

**Side effects**:
- Updates `assigned_ta_id` for all submissions
- Changes assignment status to `grading`

---

## Rubric Endpoints

### POST /rubrics
Create rubric manually (JSON builder).

**Authorization**: Any authenticated user

**Request**:
```json
{
  "assignment_id": 1,
  "json_blob": "{\"criteria\": [{\"name\": \"Correctness\", \"points\": 50}]}",
  "max_points": 100
}
```

**Response**:
```json
{
  "rubric_id": 1,
  "assignment_id": 1,
  "max_points": 100,
  "message": "Rubric created successfully"
}
```

### POST /rubrics/upload-pdf
Upload rubric PDF and auto-parse with Claude.

**Authorization**: Any authenticated user

**Request** (multipart/form-data):
```
assignment_id: 1
file: rubric.pdf
```

**Response**:
```json
{
  "rubric_id": 1,
  "assignment_id": 1,
  "parsed_rubric": "{\"criteria\": [...]}",
  "max_points": 100,
  "file_path": "uploads/rubrics/rubric_1234567890.pdf"
}
```

**Process**:
1. Saves PDF to disk
2. Extracts text with pdftotext/pdfcpu
3. Sends to Claude for parsing
4. Stores structured JSON

---

## Submission Endpoints

### POST /submissions
Upload a student submission.

**Authorization**: Any authenticated user

**Request (JSON)**:
```json
{
  "assignment_id": 1,
  "student_identifier": "student001@university.edu",
  "text": "def factorial(n):\n    if n == 0: return 1\n    return n * factorial(n-1)"
}
```

**Request (Multipart - PDF upload)**:
```
assignment_id: 1
student_identifier: student001@university.edu
file: submission.pdf
```

**Response**:
```json
{
  "submission_id": 1,
  "assignment_id": 1,
  "student_identifier": "student001@university.edu",
  "graded_status": "pending",
  "created_at": "2024-01-15T11:00:00Z"
}
```

**Grading status values**: `pending`, `in_progress`, `graded`, `regrade_required`

### GET /submissions/assigned
Get submissions assigned to the authenticated TA.

**Authorization**: grader_ta

**Response**:
```json
{
  "submissions": [
    {
      "id": 1,
      "assignment_id": 1,
      "student_identifier": "student001@university.edu",
      "text": "def factorial(n): ...",
      "file_path": "",
      "graded_status": "pending",
      "assignment_title": "Lab 1: Variables and Loops",
      "course_name": "CS 101: Introduction to Programming",
      "rubric_json": "{\"criteria\": [...]}",
      "rubric_max_points": 100
    }
  ]
}
```

---

## Grading Endpoints

### POST /grade
Grade individual submission with Claude assistance.

**Authorization**:
- grader_ta (only assigned submissions)
- professor/head_ta (any submission)

#### Mode A: Preview Recommendation
**Request**:
```json
{
  "submission_id": 1,
  "use_claude_recommendation": false
}
```

**Response**:
```json
{
  "message": "Grading recommendation generated (not saved)",
  "claude_recommendation": {
    "score": 85.5,
    "max_points": 100,
    "feedback": "Good understanding of recursion. Edge case handling could be improved.",
    "rubric_breakdown": {
      "correctness": 40,
      "code_quality": 25.5,
      "edge_cases": 10,
      "documentation": 10
    },
    "justification": "The implementation correctly handles the base case..."
  }
}
```

#### Mode B: Accept Claude's Grade
**Request**:
```json
{
  "submission_id": 1,
  "use_claude_recommendation": true
}
```

**Response**:
```json
{
  "message": "Submission graded successfully with Claude's recommendation",
  "grade_id": 1,
  "claude_recommendation": {
    "score": 85.5,
    ...
  }
}
```

**Side effects**:
- Grade saved to database
- Submission status → `graded`

#### Mode C: Manual Override
**Request**:
```json
{
  "submission_id": 1,
  "score": 90,
  "feedback": "Excellent work! Added extra credit for unit tests.",
  "rubric_breakdown": "{\"correctness\": 45, \"code_quality\": 25, \"edge_cases\": 10, \"documentation\": 10}"
}
```

**Response**:
```json
{
  "message": "Submission graded successfully with custom score",
  "grade_id": 1
}
```

### POST /grade/batch
Batch grade all ungraded submissions for an assignment.

**Authorization**: professor or head_ta only

**Request**:
```json
{
  "assignment_id": 1,
  "auto_approve": true
}
```

**Response**:
```json
{
  "assignment_id": 1,
  "total_graded": 30,
  "average_score": 82.3,
  "graded_submission_ids": [1, 2, 3, 4, ...],
  "message": "Successfully graded 30 submissions"
}
```

**Notes**:
- If `auto_approve: false`, returns recommendations without saving
- Only grades submissions without existing grades

---

## Anomaly Detection Endpoints

**Status**: Not yet implemented (Step 5)

### POST /analyze/:assignment_id
Run anomaly detection on assignment grades.

### GET /anomalies/:assignment_id
Get anomaly report for assignment.

---

## Error Responses

All endpoints return errors in this format:

```json
{
  "error": "Description of what went wrong"
}
```

### Common Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing/invalid token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource doesn't exist
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Claude API error

---

## Environment Variables

Required for server operation:

```bash
# Server
PORT=8080                    # Default: 8080
DB_PATH=../data/talytics.db  # SQLite database path

# Authentication
JWT_SECRET=your-secret-key   # REQUIRED - use strong random value

# Claude AI
CLAUDE_API_KEY=sk-ant-...    # REQUIRED - Anthropic API key
```

---

## Complete API Workflow Example

### Setup
```bash
# Start server
export JWT_SECRET="production-secret-key"
export CLAUDE_API_KEY="sk-ant-api-..."
./bin/talytics

# Register professor
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "prof@uni.edu", "password": "pass123", "role": "professor"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "prof@uni.edu", "password": "pass123"}' | jq -r '.token')
```

### Create Course & Assignment
```bash
# Create course
COURSE=$(curl -s -X POST http://localhost:8080/courses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "CS 101"}')
COURSE_ID=$(echo $COURSE | jq -r '.course_id')

# Create assignment
ASSIGNMENT=$(curl -s -X POST http://localhost:8080/assignments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"course_id\": $COURSE_ID, \"title\": \"Lab 1\", \"description\": \"Loops\"}")
ASSIGNMENT_ID=$(echo $ASSIGNMENT | jq -r '.assignment_id')
```

### Upload Rubric
```bash
# Upload rubric PDF
curl -X POST http://localhost:8080/rubrics/upload-pdf \
  -H "Authorization: Bearer $TOKEN" \
  -F "assignment_id=$ASSIGNMENT_ID" \
  -F "file=@rubric.pdf"
```

### Submit & Grade
```bash
# Upload submission
SUBMISSION=$(curl -s -X POST http://localhost:8080/submissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"assignment_id\": $ASSIGNMENT_ID,
    \"student_identifier\": \"student001\",
    \"text\": \"def factorial(n): return 1 if n == 0 else n * factorial(n-1)\"
  }")
SUBMISSION_ID=$(echo $SUBMISSION | jq -r '.submission_id')

# Notify TAs (distributes submissions)
curl -s -X POST http://localhost:8080/assignments/$ASSIGNMENT_ID/notify-tas \
  -H "Authorization: Bearer $TOKEN"

# Grade with Claude
curl -s -X POST http://localhost:8080/grade \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"submission_id\": $SUBMISSION_ID, \"use_claude_recommendation\": true}"
```

---

## Anomaly Detection Endpoints

### POST /analyze/:assignment_id
Run statistical anomaly detection on assignment grades.

**Authorization**: Professor or head_ta only

**Request**:
```bash
curl -X POST http://localhost:8080/analyze/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "message": "Anomaly analysis completed successfully",
  "report_id": 1,
  "summary": {
    "assignment_id": 1,
    "total_grades": 30,
    "average_score": 82.3,
    "standard_deviation": 12.5,
    "ta_severity_issues": [
      {
        "ta_id": 2,
        "ta_email": "harsh_ta@uni.edu",
        "average_score": 65.2,
        "grades_count": 10,
        "deviation": -2.1,
        "severity": "too_harsh"
      }
    ],
    "outlier_grades": [
      {
        "submission_id": 5,
        "student_identifier": "student005",
        "score": 35.0,
        "z_score": -3.2,
        "graded_by": 2,
        "grader_email": "harsh_ta@uni.edu"
      }
    ],
    "criterion_issues": [
      {
        "criterion_name": "code_quality",
        "average_score": 20.5,
        "standard_deviation": 8.3,
        "inconsistent_submission_ids": [3, 7, 12]
      }
    ],
    "regrade_risks": [
      {
        "submission_id": 5,
        "student_identifier": "student005",
        "score": 59.0,
        "risk_score": 85,
        "risk_factors": ["unusually_low_score", "harsh_grader", "near_boundary_60"],
        "graded_by": 2,
        "grader_email": "harsh_ta@uni.edu"
      }
    ],
    "generated_at": "2024-01-15T14:30:00Z"
  }
}
```

**Detection Algorithms**:
1. **TA Severity Deviation**: Flags TAs grading >1.5 std deviations from mean
2. **Statistical Outliers**: Identifies grades with |z-score| > 2.0
3. **Criterion Inconsistency**: Detects criteria with coefficient of variation > 0.3
4. **Regrade Risk**: Multi-factor scoring (0-100) based on outliers, harsh TAs, boundary grades

**Requirements**:
- Minimum 5 grades needed for statistical validity
- Assignment must have rubric
- All submissions must be graded

**Notes**:
- Report saved to database with status "pending"
- Can be re-run to generate updated analysis

### GET /anomalies/:assignment_id
Retrieve anomaly report for assignment.

**Authorization**: Professor or head_ta only

**Request**:
```bash
curl -X GET http://localhost:8080/anomalies/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "report": {
    "assignment_id": 1,
    "total_grades": 30,
    "average_score": 82.3,
    "ta_severity_issues": [...],
    "outlier_grades": [...],
    "criterion_issues": [...],
    "regrade_risks": [...]
  },
  "status": "pending"
}
```

**Status values**:
- `pending` - Report generated, awaiting review
- `reviewed` - Professor has reviewed
- `resolved` - Issues addressed

---

## Rate Limits

**Claude API**:
- Tier 1: 50 requests/min
- Tier 2: 1,000 requests/min
- Tier 3: 2,000 requests/min

**TAlytics Server**: No rate limits (add nginx if needed)

---

## Changelog

### Step 5 (Current) - Anomaly Detection
- ✅ POST /analyze/:assignment_id (statistical analysis)
- ✅ GET /anomalies/:assignment_id (retrieve report)
- ✅ TA severity deviation detection
- ✅ Outlier grade identification
- ✅ Criterion inconsistency analysis
- ✅ Regrade risk prediction

### Step 4 - Grading
- ✅ POST /grade (individual grading)
- ✅ POST /grade/batch (batch grading)
- ✅ Grade preview mode
- ✅ Manual override support

### Step 3 - Authentication & Workflows
- ✅ POST /auth/register
- ✅ POST /auth/login
- ✅ JWT authentication
- ✅ POST /courses
- ✅ GET /courses
- ✅ POST /assignments
- ✅ GET /assignments/:id
- ✅ POST /assignments/:id/notify-tas
- ✅ POST /submissions
- ✅ GET /submissions/assigned

### Step 2 - PDF & Claude Integration
- ✅ POST /rubrics
- ✅ POST /rubrics/upload-pdf
- ✅ Claude integration
- ✅ PDF extraction

### Step 1 - Backend Scaffolding
- ✅ Backend scaffolding
- ✅ Database schema
- ✅ All endpoint stubs

---

## Backend Implementation Status

**All 15 endpoints implemented**: ✅ 100% COMPLETE

- Auth: 2/2 endpoints
- Courses: 2/2 endpoints
- Assignments: 3/3 endpoints
- Rubrics: 2/2 endpoints
- Submissions: 2/2 endpoints
- Grading: 2/2 endpoints
- Anomaly Detection: 2/2 endpoints

# TAlytics Backend

AI-powered fair and consistent grading assistant with TA workload management and anomaly detection.

## Project Structure

```
backend/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/               # HTTP request handlers (all stubs)
│   │   │   ├── auth.go            # Authentication endpoints
│   │   │   ├── courses.go         # Course management
│   │   │   ├── assignments.go     # Assignment management
│   │   │   ├── rubrics.go         # Rubric creation & upload
│   │   │   ├── submissions.go     # Submission handling
│   │   │   ├── grading.go         # Grading endpoints
│   │   │   └── anomalies.go       # Anomaly detection
│   │   ├── middleware/
│   │   │   └── auth.go            # JWT authentication middleware
│   │   └── routes.go              # Route definitions
│   ├── database/
│   │   ├── db.go                  # Database initialization
│   │   └── schema.sql             # SQLite schema
│   ├── models/
│   │   └── models.go              # Data models and DTOs
│   └── services/
│       ├── claude.go              # Claude API integration (stub)
│       ├── pdf.go                 # PDF extraction service (stub)
│       └── anomaly.go             # Anomaly detection algorithms (stub)
├── go.mod
└── README.md
```

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: SQLite
- **AI**: Claude API (Anthropic)
- **PDF Extraction**: pdftotext (Poppler) / pdfcpu

## API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

### Courses
- `POST /courses` - Create course
- `GET /courses` - Get user's courses

### Assignments
- `POST /assignments` - Create assignment
- `GET /assignments/:id` - Get assignment details
- `POST /assignments/:id/notify-tas` - Open assignment for grading

### Rubrics
- `POST /rubrics` - Create rubric from UI
- `POST /rubrics/upload` - Upload PDF rubric for parsing

### Submissions
- `POST /submissions/upload` - Upload student submission
- `GET /submissions/assigned` - Get TA's assigned submissions

### Grading
- `POST /grade` - Grade submission
- `POST /grade/batch` - Batch grade multiple submissions

### Anomaly Detection
- `POST /analyze/:assignment_id` - Run anomaly detection
- `GET /anomalies/:assignment_id` - Get anomaly report

## Setup Instructions

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Install PDF Tools (Optional for PDF extraction)

**macOS (Homebrew):**
```bash
brew install poppler  # For pdftotext
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install poppler-utils
```

### 3. Set Environment Variables

Create a `.env` file or export variables:

```bash
export PORT=8080
export DB_PATH=./data/talytics.db
export CLAUDE_API_KEY=your_api_key_here
```

### 4. Run the Server

```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

Health check: `http://localhost:8080/health`

## Database Schema

### Tables

1. **users** - System users (professors, TAs)
2. **courses** - Course information
3. **assignments** - Course assignments
4. **rubrics** - Grading rubrics (JSON format)
5. **submissions** - Student submissions
6. **grades** - Grading results
7. **anomaly_reports** - Anomaly detection results

## Current Status

✅ **COMPLETED - Step 1: Backend Scaffolding**
- All directory structure created
- SQLite schema defined
- All API endpoint stubs created
- Service layer stubs created
- Database initialization complete
- Router configuration complete

✅ **COMPLETED - Step 2: PDF & Claude Integration**
- ✅ PDF text extraction implemented (pdftotext/pdfcpu)
- ✅ Rubric ingestion and parsing implemented
- ✅ Claude API integration complete
- ✅ Rubric upload endpoint fully functional
- ✅ Database helpers for rubric operations

⏳ **PENDING - Step 3: Authentication & Workflows**
- Authentication logic (JWT, password hashing)
- Assignment workflows
- TA dashboard logic
- Grading UI backend
- Anomaly detection algorithms
- Reporting features

## Implemented Features

### PDF Text Extraction (`services/pdf.go`)
- Supports pdftotext (Poppler) and pdfcpu
- Automatic tool detection and fallback
- File saving with unique timestamps
- Error handling and validation

### Claude API Integration (`services/claude.go`)
- Full Claude 3.5 Sonnet integration
- Rubric parsing from PDF text
- Grading recommendations (ready for use)
- TA question answering (ready for use)
- Anomaly insights generation (ready for use)

### Rubric Management (`handlers/rubrics.go`)
- **POST /rubrics** - Create rubric from manual JSON
- **POST /rubrics/upload** - Upload PDF, extract text, parse with Claude
- Assignment validation
- Duplicate rubric prevention
- Automatic max_points extraction

### Database Helpers (`database/db.go`)
- `InsertRubric()` - Save rubric to database
- `GetAssignmentByID()` - Validate assignments
- `RubricExists()` - Check for duplicates
- `GetRubricByAssignmentID()` - Retrieve rubrics

## Usage Examples

### Upload a Rubric PDF

```bash
curl -X POST http://localhost:8080/rubrics/upload \
  -F "assignment_id=1" \
  -F "file=@rubric.pdf"
```

**Response:**
```json
{
  "rubric_id": 1,
  "assignment_id": 1,
  "parsed_rubric": "{\"title\":\"Assignment 1 Rubric\",\"max_points\":100,...}",
  "max_points": 100,
  "file_path": "./data/rubrics/rubric_20250115_123045.pdf"
}
```

### Create a Rubric Manually

```bash
curl -X POST http://localhost:8080/rubrics \
  -H "Content-Type: application/json" \
  -d '{
    "assignment_id": 1,
    "json_blob": "{\"title\":\"Manual Rubric\",\"max_points\":100,\"criteria\":[...]}",
    "max_points": 100
  }'
```

## Testing

Run the test script to verify PDF and Claude integration:

```bash
cd backend
./test_rubric_upload.sh
```

## Next Steps

After approval, the following will be implemented in order:

1. **Authentication & Authorization**
   - JWT token generation
   - Password hashing
   - Role-based access control

3. **Assignment Workflows**
   - Submission distribution among TAs
   - Grading workflow logic
   - Progress tracking

4. **Anomaly Detection**
   - TA severity deviation analysis
   - Statistical variance detection
   - Regrade risk prediction

5. **Reporting & Dashboards**
   - Statistical visualizations
   - Anomaly reports
   - Grade exports

## Development Notes

- All handlers return `501 Not Implemented` with stub messages
- Middleware currently allows all requests (no auth enforcement)
- Service methods are placeholder functions
- Database schema is production-ready
- CORS configured for local frontend development

## Testing

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test stub endpoints (should return 501)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123","role":"professor"}'
```

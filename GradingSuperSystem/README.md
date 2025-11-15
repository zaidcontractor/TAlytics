# TAlytics - AI-Powered Grading Standardization System

TAlytics is an intelligent grading quality assurance platform designed to detect inconsistencies and bias in multi-TA grading environments. Built specifically for large undergraduate courses where multiple teaching assistants grade the same assignments or exams.

## 🎯 Problem Solved

In large courses like UVA's Data Structures and Algorithms, multiple TAs grade programming assignments and exams using the same rubric. However, individual interpretation differences lead to:

- **Inconsistent Grading**: Same work receives different scores from different TAs
- **Student Complaints**: Hours spent on regrade requests and disputes  
- **Administrative Overhead**: Professors spending time resolving grading conflicts
- **Unfairness**: Some students unfairly penalized due to TA variance

## 🚀 Solution

TAlytics provides a proactive approach to grading quality control through:

### Core Features

1. **Rubric Management System**
   - Create structured grading criteria with weights
   - Version control for rubric evolution
   - Template library for common assessment types

2. **Grade Data Analysis**
   - CSV upload for existing grade data
   - Real-time statistical analysis
   - Integration-ready APIs for LMS systems

3. **Anomaly Detection Engine**
   - Inter-grader reliability calculation (Krippendorff's Alpha)
   - Outlier detection using Modified Z-Score
   - ANOVA testing for TA variance
   - Temporal drift analysis

4. **Interactive Dashboard**
   - Real-time grading distribution visualization
   - TA performance comparison charts
   - Question-level difficulty analytics
   - Alert system for inconsistencies

5. **Feedback & Correction System**
   - Automated anomaly alerts
   - Guided rubric adjustment workflows
   - Impact simulation for proposed changes
   - Continuous improvement tracking

### Statistical Methods

- **Krippendorff's Alpha**: Inter-grader reliability assessment
- **Modified Z-Score**: Robust outlier detection
- **ANOVA**: Testing for significant differences between TAs
- **Control Charts**: Monitoring grading trends over time
- **State Estimation Framework**: Predictive grading model with feedback loops

## 🏗️ Architecture

### Technology Stack

- **Frontend**: React.js with Chart.js/D3.js for visualizations
- **Backend**: Go with gRPC and Protocol Buffers
- **Database**: SQLite for development, PostgreSQL-ready for production  
- **Analytics**: Advanced statistical analysis engine with Go implementation
- **Communication**: gRPC services with HTTP REST gateway for frontend compatibility
- **Deployment**: Ready for Vercel (frontend) + any Go-compatible hosting (backend)

### Project Structure

```
TAlytics/
├── client/                 # React frontend application
│   ├── src/
│   │   ├── components/     # React components
│   │   │   ├── Dashboard.js
│   │   │   ├── RubricManager.js
│   │   │   ├── GradeUpload.js
│   │   │   └── AnalysisResults.js
│   │   └── App.js
│   └── package.json
├── server/                 # Go backend with gRPC services
│   ├── cmd/               # Main application entry point
│   ├── internal/          # Private Go packages
│   │   ├── services/      # gRPC service implementations
│   │   ├── database/      # Database layer
│   │   └── statistics/    # Statistical analysis engine
│   ├── proto/             # Protocol Buffer definitions
│   └── go.mod
│   │   ├── rubrics.js     # Rubric CRUD operations
│   │   ├── grades.js      # Grade data management
│   │   └── analysis.js    # Statistical analysis
│   ├── database/          # Database setup and models
│   ├── utils/             # Statistical algorithms
│   │   └── statistics.js  # Core statistical functions
│   └── index.js           # Express server
└── package.json           # Root package configuration
```

## 🚀 Getting Started

### Prerequisites

- Go 1.18+ 
- Node.js 14+ and npm
- Git
- Protocol Buffers compiler (protoc)

### Installation

1. **Clone and Install**
   ```bash
   git clone <repository-url>
   cd TAlytics
   npm install && cd client && npm install
   ```

2. **Build Go Backend**
   ```bash
   cd server
   go mod tidy
   go build -o talytics-server ./cmd
   cd ..
   ```

3. **Start Development Servers**
   ```bash
   npm run dev
   ```
   
   This starts both:
   - Frontend: http://localhost:3000
   - Go Backend: http://localhost:5000 (REST API) + :50051 (gRPC)
   - Backend API: http://localhost:5000

### Usage Workflow

1. **Create a Rubric**
   - Navigate to Rubrics tab
   - Define grading criteria with weights
   - Save rubric for future use

2. **Upload Grade Data**
   - Export grading data as CSV
   - Format: `student_id,ta_id,question_id,score,max_score`
   - Upload via Upload Grades tab

3. **Analyze Results**
   - Run anomaly detection analysis
   - Review statistical insights
   - Identify grading inconsistencies

4. **Take Action**
   - Review flagged anomalies
   - Adjust rubric clarity
   - Provide TA feedback

## 📊 Sample Data Format

```csv
student_id,ta_id,question_id,score,max_score
student001,ta_alice,q1,85,100
student001,ta_alice,q2,78,100
student002,ta_bob,q1,76,100
student002,ta_bob,q2,84,100
```

## 🔬 Statistical Analysis

### Anomaly Types Detected

1. **High TA Variance**: Significant differences in mean scores between TAs
2. **Outlier Scores**: Individual scores that deviate significantly from the norm
3. **Low Inter-Grader Reliability**: Overall inconsistency across all graders
4. **Temporal Drift**: Changes in grading severity over time

### Key Metrics

- **Inter-Grader Reliability**: >0.8 excellent, <0.67 concerning
- **Coefficient of Variation**: Measures relative variability
- **Effect Size (Cohen's d)**: Practical significance of differences
- **Modified Z-Score**: Robust outlier detection (threshold: 3.5)

## 🎯 Value Proposition

### For Operations & Staff
- **Time Savings**: Convert reactive regrade hours into proactive minutes
- **Quality Assurance**: Systematic detection of grading issues
- **Scalable Monitoring**: Works regardless of class size
- **Data-Driven Decisions**: Replace anecdotal evidence with statistical analysis

### For Educational Outcomes
- **Fairness**: Ensure equitable assessment across all students
- **Validity**: Grades accurately reflect student learning
- **Continuous Improvement**: Evidence-based rubric refinement
- **TA Development**: Identify training opportunities

## 🏆 Competition Advantages

1. **Novel Application**: First statistical process control system for academic grading
2. **Technical Sophistication**: State estimation framework with feedback loops
3. **Practical Impact**: Solves real pain points in large course management
4. **Scalable Design**: Applicable to any institution with multi-grader courses

## 🔮 Future Enhancements

- **Machine Learning**: Predictive grading models
- **LMS Integration**: Direct Canvas/Blackboard connectivity  
- **Mobile App**: On-the-go grading consistency monitoring
- **Advanced Analytics**: Fourier analysis for periodic grading patterns
- **Collaborative Features**: Multi-institution benchmarking

## 🤝 Contributing

Built for the Claude for Good Hackathon at University of Virginia. Focused on Operations & Staff Support track to streamline educational backbone systems.

## 📄 License

MIT License - see LICENSE file for details.

---

**TAlytics**: Transforming grading from subjective to systematic, one rubric at a time.
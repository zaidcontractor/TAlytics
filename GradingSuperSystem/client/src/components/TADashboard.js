import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import AssignmentDetail from './AssignmentDetail';
import GradingViewer from './GradingViewer';
import './Dashboard.css';
import './TADashboard.css';

const TADashboard = ({ course, onBack }) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [assignments, setAssignments] = useState([]);
  const [pendingGrades, setPendingGrades] = useState([]);
  const [completedGrades, setCompletedGrades] = useState([]);
  const [regradingNeeded, setRegradingNeeded] = useState([]);
  const [selectedAssignment, setSelectedAssignment] = useState(null);
  const [viewingSubmissions, setViewingSubmissions] = useState(false);
  const [loading, setLoading] = useState(false);

  const fetchTAData = useCallback(async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      // Fetch assignments for this course
      const assignmentsResponse = await axios.get(
        `http://localhost:5000/api/courses/${course.id}/assignments`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      const assignmentsData = assignmentsResponse.data.assignments || [];
      setAssignments(assignmentsData);

      // Fetch grades that need regrading for this TA
      const userID = localStorage.getItem('user_id'); // Assuming we store this
      const regradingResponse = await axios.get(
        `http://localhost:5000/api/grades/needs-regrading`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      if (regradingResponse.data.grades) {
        setRegradingNeeded(regradingResponse.data.grades);
      }

      // Mock pending and completed grades (would come from actual grading assignments)
      setPendingGrades([
        {
          id: 1,
          student_name: 'John Doe',
          assignment_name: 'Homework 1',
          submitted_at: new Date().toISOString(),
          priority: 'high',
          pdf_filename: 'john_doe_hw1.pdf'
        },
        {
          id: 2,
          student_name: 'Jane Smith', 
          assignment_name: 'Lab 2',
          submitted_at: new Date().toISOString(),
          priority: 'medium',
          pdf_filename: 'jane_smith_lab2.pdf'
        }
      ]);

      setCompletedGrades([
        {
          id: 3,
          student_name: 'Mike Johnson',
          assignment_name: 'Quiz 1',
          graded_at: new Date().toISOString(),
          score: 85
        }
      ]);

    } catch (error) {
      console.error('Error fetching TA data:', error);
      // If course is not found (404) or access denied (403), redirect back to course list
      if (error.response && (error.response.status === 404 || error.response.status === 403)) {
        alert('This course is no longer available. You will be redirected to the course list.');
        onBack();
      }
    } finally {
      setLoading(false);
    }
  }, [course, onBack]);

  useEffect(() => {
    if (course) {
      fetchTAData();
    }
  }, [course, fetchTAData]);

  // eslint-disable-next-line no-unused-vars
  const handleViewAssignment = (assignment) => {
    setSelectedAssignment(assignment);
    setViewingSubmissions(true);
  };

  const handleBackToAssignments = () => {
    setSelectedAssignment(null);
    setViewingSubmissions(false);
  };

  // Return submission viewer if viewing submissions
  if (viewingSubmissions && selectedAssignment) {
    return (
      <GradingViewer 
        assignment={selectedAssignment} 
        onBack={handleBackToAssignments}
        userRole="ta"
      />
    );
  }

  // Return assignment detail view if assignment is selected
  if (selectedAssignment) {
    return (
      <AssignmentDetail 
        assignment={selectedAssignment} 
        onBack={handleBackToAssignments}
        userRole="ta"
      />
    );
  }

  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'high': return 'priority-high';
      case 'medium': return 'priority-medium';
      case 'low': return 'priority-low';
      default: return 'priority-medium';
    }
  };

  const renderOverview = () => (
    <div className="ta-overview">
      {regradingNeeded.length > 0 && (
        <div className="regrading-alert">
          <div className="alert-icon">⚠️</div>
          <div className="alert-content">
            <h4>Rubric Updated - Regrading Required</h4>
            <p>
              The instructor has updated the rubric for some assignments. 
              You need to re-grade <strong>{regradingNeeded.length} submission{regradingNeeded.length > 1 ? 's' : ''}</strong>.
            </p>
            <button 
              className="btn btn-primary"
              onClick={() => setActiveTab('regrading')}
            >
              View Submissions for Regrading
            </button>
          </div>
        </div>
      )}
      
      <div className="overview-stats">
        <div className="stat-card pending-stat">
          <div className="stat-icon">⏳</div>
          <div className="stat-content">
            <div className="stat-number">{pendingGrades.length}</div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
        
        <div className="stat-card regrading-stat">
          <div className="stat-icon">🔄</div>
          <div className="stat-content">
            <div className="stat-number">{regradingNeeded.length}</div>
            <div className="stat-label">Needs Regrading</div>
          </div>
        </div>
        
        <div className="stat-card completed-stat">
          <div className="stat-icon">✅</div>
          <div className="stat-content">
            <div className="stat-number">{completedGrades.length}</div>
            <div className="stat-label">Completed</div>
          </div>
        </div>
        
        <div className="stat-card assignments-stat">
          <div className="stat-icon">📝</div>
          <div className="stat-content">
            <div className="stat-number">{assignments.length}</div>
            <div className="stat-label">Assignments</div>
          </div>
        </div>
      </div>

      <div className="course-info">
        <h3>Course Information</h3>
        <div className="course-details">
          <div className="detail-item">
            <span className="detail-label">Course:</span>
            <span className="detail-value">{course.name}</span>
          </div>
          <div className="detail-item">
            <span className="detail-label">Code:</span>
            <span className="detail-value course-code">{course.code}</span>
          </div>
          <div className="detail-item">
            <span className="detail-label">Your Role:</span>
            <span className="detail-value">
              <span className="role-badge ta">Teaching Assistant</span>
            </span>
          </div>
        </div>
      </div>

      <div className="quick-actions">
        <h3>Quick Actions</h3>
        <div className="action-buttons">
          <button 
            className="action-btn"
            onClick={() => setActiveTab('pending')}
          >
            <div className="action-icon">⏳</div>
            <div className="action-text">
              <div className="action-title">Grade Submissions</div>
              <div className="action-subtitle">{pendingGrades.length} pending</div>
            </div>
          </button>
          
          <button 
            className="action-btn"
            onClick={() => setActiveTab('completed')}
          >
            <div className="action-icon">📊</div>
            <div className="action-text">
              <div className="action-title">Review Grades</div>
              <div className="action-subtitle">{completedGrades.length} completed</div>
            </div>
          </button>
        </div>
      </div>
    </div>
  );

  const renderAssignments = () => (
    <div className="assignments-tab">
      <div className="tab-header">
        <h3>Course Assignments</h3>
        <span className="assignment-count">{assignments.length} assignments</span>
      </div>

      {assignments.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📝</div>
          <h4>No assignments yet</h4>
          <p>Assignments created by the instructor will appear here.</p>
        </div>
      ) : (
        <div className="assignments-grid">
          {assignments.map(assignment => (
            <div key={assignment.id} className="assignment-card">
              <div className="assignment-header">
                <h4>{assignment.name}</h4>
                <span className="assignment-status active">Active</span>
              </div>
              
              <div className="assignment-details">
                <div className="detail">
                  <span className="detail-label">Submissions:</span>
                  <span className="detail-value">{assignment.submission_count || 0}</span>
                </div>
                {assignment.rubric && (
                  <div className="detail">
                    <span className="detail-label">Criteria:</span>
                    <span className="detail-value">{assignment.rubric.criteria?.length || 0}</span>
                  </div>
                )}
              </div>
              
              <div className="assignment-actions">
                <button 
                  className="btn btn-outline"
                  onClick={() => handleViewAssignment(assignment)}
                >
                  👁️ View Submissions
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const renderPendingGrades = () => (
    <div className="pending-grades">
      <div className="tab-header">
        <h3>Pending Grades</h3>
        <span className="grade-count">{pendingGrades.length} submissions</span>
      </div>

      {pendingGrades.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">✅</div>
          <h4>All caught up!</h4>
          <p>No pending submissions to grade at the moment.</p>
        </div>
      ) : (
        <div className="grades-list">
          {pendingGrades.map(grade => (
            <div key={grade.id} className="grade-card pending">
              <div className="grade-header">
                <div className="student-info">
                  <div className="student-avatar">
                    {grade.student_name.charAt(0)}
                  </div>
                  <div className="student-details">
                    <div className="student-name">{grade.student_name}</div>
                    <div className="assignment-name">{grade.assignment_name}</div>
                  </div>
                </div>
                <div className={`priority-badge ${getPriorityColor(grade.priority)}`}>
                  {grade.priority}
                </div>
              </div>
              
              <div className="grade-details">
                <div className="submission-info">
                  <span className="detail-label">Submitted:</span>
                  <span className="detail-value">
                    {new Date(grade.submitted_at).toLocaleDateString()}
                  </span>
                </div>
                <div className="pdf-info">
                  <span className="detail-label">File:</span>
                  <span className="detail-value">{grade.pdf_filename}</span>
                </div>
              </div>
              
              <div className="grade-actions">
                <button className="btn btn-outline">View PDF</button>
                <button className="btn btn-primary">Start Grading</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const renderCompletedGrades = () => (
    <div className="completed-grades">
      <div className="tab-header">
        <h3>Completed Grades</h3>
        <span className="grade-count">{completedGrades.length} graded</span>
      </div>

      {completedGrades.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📝</div>
          <h4>No completed grades yet</h4>
          <p>Your graded submissions will appear here.</p>
        </div>
      ) : (
        <div className="grades-list">
          {completedGrades.map(grade => (
            <div key={grade.id} className="grade-card completed">
              <div className="grade-header">
                <div className="student-info">
                  <div className="student-avatar completed">
                    {grade.student_name.charAt(0)}
                  </div>
                  <div className="student-details">
                    <div className="student-name">{grade.student_name}</div>
                    <div className="assignment-name">{grade.assignment_name}</div>
                  </div>
                </div>
                <div className="score-badge">
                  {grade.score}/100
                </div>
              </div>
              
              <div className="grade-details">
                <div className="completion-info">
                  <span className="detail-label">Graded:</span>
                  <span className="detail-value">
                    {new Date(grade.graded_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
              
              <div className="grade-actions">
                <button className="btn btn-outline">Review Grade</button>
                <button className="btn btn-secondary">View Feedback</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const renderRegradingNeeded = () => (
    <div className="regrading-grades">
      <div className="tab-header">
        <h3>Submissions Needing Regrading</h3>
        <span className="grade-count">{regradingNeeded.length} submissions</span>
      </div>

      <div className="regrading-info-banner">
        <div className="info-icon">ℹ️</div>
        <div className="info-text">
          <h4>Why do I need to regrade?</h4>
          <p>
            The instructor has modified the rubric criteria or weights for these assignments. 
            Your previous grades are no longer valid and need to be updated according to the new rubric.
          </p>
        </div>
      </div>

      {regradingNeeded.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">✅</div>
          <h4>All caught up!</h4>
          <p>No submissions require regrading at this time.</p>
        </div>
      ) : (
        <div className="grades-list">
          {regradingNeeded.map(grade => (
            <div key={grade.id} className="grade-card regrading">
              <div className="grade-header">
                <div className="student-info">
                  <div className="student-avatar regrading">
                    {grade.student_name?.charAt(0) || '?'}
                  </div>
                  <div className="student-details">
                    <div className="student-name">{grade.student_name}</div>
                    <div className="assignment-name">{grade.assignment_name}</div>
                  </div>
                </div>
                <div className="priority-badge priority-high">
                  REGRADE NEEDED
                </div>
              </div>
              
              <div className="grade-details">
                <div className="previous-grade">
                  <span className="detail-label">Previous Score:</span>
                  <span className="detail-value">{grade.total_score || 'N/A'}</span>
                </div>
                <div className="rubric-updated">
                  <span className="detail-label">Rubric Updated:</span>
                  <span className="detail-value">
                    {new Date(grade.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
              
              <div className="grade-actions">
                <button className="btn btn-primary">Start Regrading</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const tabs = [
    { id: 'overview', label: 'Overview', icon: '📊' },
    { id: 'assignments', label: 'Assignments', icon: '📝', count: assignments.length },
    { id: 'regrading', label: 'Regrading', icon: '🔄', count: regradingNeeded.length, highlight: regradingNeeded.length > 0 }
  ];

  return (
    <div className="ta-dashboard">
      <div className="dashboard-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Courses
        </button>
        <div className="course-title">
          <h2>{course.name}</h2>
          <span className="ta-badge">TA Dashboard</span>
        </div>
      </div>

      <div className="dashboard-tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            className={`tab-button ${activeTab === tab.id ? 'active' : ''} ${tab.highlight ? 'highlight' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
            {tab.count !== undefined && tab.count > 0 && (
              <span className={`tab-count ${tab.highlight ? 'alert' : ''}`}>{tab.count}</span>
            )}
          </button>
        ))}
      </div>

      <div className="dashboard-content">
        {loading ? (
          <div className="loading-spinner">
            <div className="spinner"></div>
            <p>Loading grading data...</p>
          </div>
        ) : (
          <>
            {activeTab === 'overview' && renderOverview()}
            {activeTab === 'assignments' && renderAssignments()}
            {activeTab === 'regrading' && renderRegradingNeeded()}
          </>
        )}
      </div>
    </div>
  );
};

export default TADashboard;
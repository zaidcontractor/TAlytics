import React, { useState } from 'react';
import axios from 'axios';
import './AssignmentDetail.css';

const AssignmentDetail = ({ assignment, onBack, userRole }) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [submissions, setSubmissions] = useState([
    {
      id: 1,
      student_name: 'John Doe',
      filename: 'john_doe_hw1.pdf',
      uploaded_at: new Date().toISOString(),
      graded: false
    },
    {
      id: 2,
      student_name: 'Jane Smith',
      filename: 'jane_smith_hw1.pdf',
      uploaded_at: new Date().toISOString(),
      graded: true,
      grade: 85
    }
  ]);
  const [rubric, setRubric] = useState({
    criteria: [
      {
        id: 1,
        name: 'Problem Understanding',
        description: 'Student demonstrates understanding of the problem',
        max_points: 25
      },
      {
        id: 2,
        name: 'Solution Quality',
        description: 'Quality and correctness of the solution',
        max_points: 50
      },
      {
        id: 3,
        name: 'Code Style',
        description: 'Code readability and style',
        max_points: 25
      }
    ],
    total_points: 100
  });

  const handleFileUpload = async (event) => {
    const files = Array.from(event.target.files);
    const formData = new FormData();
    
    files.forEach(file => {
      formData.append('submissions', file);
    });
    formData.append('assignment_id', assignment.id);

    try {
      const token = localStorage.getItem('token');
      const response = await axios.post(
        `http://localhost:5000/api/assignments/${assignment.id}/submissions`,
        formData,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'multipart/form-data'
          }
        }
      );
      
      // Refresh submissions list
      console.log('Files uploaded successfully:', response.data);
      // TODO: Refresh submissions from server
    } catch (error) {
      console.error('Upload failed:', error);
      alert('Upload failed. Please try again.');
    }
  };

  const viewPDF = (submission) => {
    // TODO: Implement PDF viewer
    alert(`Viewing ${submission.filename} for ${submission.student_name}`);
  };

  const renderOverview = () => (
    <div className="assignment-overview">
      <div className="overview-stats">
        <div className="stat-card">
          <div className="stat-icon">📄</div>
          <div className="stat-content">
            <div className="stat-number">{submissions.length}</div>
            <div className="stat-label">Submissions</div>
          </div>
        </div>
        
        <div className="stat-card">
          <div className="stat-icon">✅</div>
          <div className="stat-content">
            <div className="stat-number">{submissions.filter(s => s.graded).length}</div>
            <div className="stat-label">Graded</div>
          </div>
        </div>
        
        <div className="stat-card">
          <div className="stat-icon">⏳</div>
          <div className="stat-content">
            <div className="stat-number">{submissions.filter(s => !s.graded).length}</div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
      </div>

      <div className="assignment-details">
        <h3>Assignment Details</h3>
        <div className="detail-item">
          <span className="detail-label">Name:</span>
          <span className="detail-value">{assignment.name}</span>
        </div>
        <div className="detail-item">
          <span className="detail-label">Description:</span>
          <span className="detail-value">{assignment.description}</span>
        </div>
        <div className="detail-item">
          <span className="detail-label">Due Date:</span>
          <span className="detail-value">{new Date(assignment.due_date).toLocaleDateString()}</span>
        </div>
        <div className="detail-item">
          <span className="detail-label">Total Points:</span>
          <span className="detail-value">{rubric.total_points}</span>
        </div>
      </div>
    </div>
  );

  const renderSubmissions = () => (
    <div className="submissions-tab">
      <div className="tab-header">
        <h3>Submissions</h3>
        {userRole === 'instructor' && (
          <div className="upload-section">
            <input
              type="file"
              id="pdf-upload"
              multiple
              accept=".pdf"
              onChange={handleFileUpload}
              style={{ display: 'none' }}
            />
            <label htmlFor="pdf-upload" className="btn btn-success">
              📄 Upload PDFs
            </label>
          </div>
        )}
      </div>

      <div className="submissions-list">
        {submissions.map(submission => (
          <div key={submission.id} className="submission-card">
            <div className="submission-info">
              <div className="student-name">{submission.student_name}</div>
              <div className="file-info">
                <span className="filename">📄 {submission.filename}</span>
                <span className="upload-time">
                  Uploaded: {new Date(submission.uploaded_at).toLocaleDateString()}
                </span>
              </div>
            </div>
            
            <div className="submission-status">
              {submission.graded ? (
                <span className="grade-badge graded">
                  Graded: {submission.grade}/100
                </span>
              ) : (
                <span className="grade-badge pending">
                  Pending
                </span>
              )}
            </div>
            
            <div className="submission-actions">
              <button 
                className="btn btn-outline"
                onClick={() => viewPDF(submission)}
              >
                View PDF
              </button>
              {!submission.graded && (
                <button className="btn btn-primary">
                  Grade
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  const renderRubric = () => (
    <div className="rubric-tab">
      <div className="tab-header">
        <h3>Grading Rubric</h3>
        {userRole === 'instructor' && (
          <button className="btn btn-primary">
            Edit Rubric
          </button>
        )}
      </div>

      <div className="rubric-container">
        <div className="rubric-summary">
          <h4>Total Points: {rubric.total_points}</h4>
        </div>
        
        <div className="criteria-list">
          {rubric.criteria.map(criterion => (
            <div key={criterion.id} className="criterion-card">
              <div className="criterion-header">
                <h4>{criterion.name}</h4>
                <span className="points">{criterion.max_points} pts</span>
              </div>
              <p className="criterion-description">{criterion.description}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );

  const tabs = [
    { id: 'overview', label: 'Overview', icon: '📊' },
    { id: 'submissions', label: 'Submissions', icon: '📄' },
    { id: 'rubric', label: 'Rubric', icon: '📋' }
  ];

  return (
    <div className="assignment-detail">
      <div className="detail-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Assignments
        </button>
        <div className="assignment-title">
          <h2>{assignment.name}</h2>
          <span className="assignment-type">Assignment Detail</span>
        </div>
      </div>

      <div className="detail-tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
          </button>
        ))}
      </div>

      <div className="detail-content">
        {activeTab === 'overview' && renderOverview()}
        {activeTab === 'submissions' && renderSubmissions()}
        {activeTab === 'rubric' && renderRubric()}
      </div>
    </div>
  );
};

export default AssignmentDetail;
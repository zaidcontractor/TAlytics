import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './GradedSubmissionsView.css';

const GradedSubmissionsView = ({ assignment, onBack, onViewSubmission }) => {
  const [gradedSubmissions, setGradedSubmissions] = useState([]);
  const [ungradedSubmissions, setUngradedSubmissions] = useState([]);
  const [activeTab, setActiveTab] = useState('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchSubmissionsData();
  }, [assignment]);

  const fetchSubmissionsData = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      // Fetch all submissions
      const submissionsResponse = await axios.get(
        `http://localhost:5000/api/submissions/assignment/${assignment.id}`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      const allSubmissions = submissionsResponse.data.submissions || [];
      
      // Fetch all grades
      const gradesResponse = await axios.get(
        `http://localhost:5000/api/grades/assignment/${assignment.id}`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      const grades = gradesResponse.data.grades || [];
      
      // Map grades to submissions
      const gradeMap = {};
      grades.forEach(grade => {
        gradeMap[grade.submission_id] = grade;
      });
      
      // Separate graded and ungraded
      const graded = [];
      const ungraded = [];
      
      allSubmissions.forEach(sub => {
        if (gradeMap[sub.id]) {
          graded.push({
            ...sub,
            grade: gradeMap[sub.id]
          });
        } else {
          ungraded.push(sub);
        }
      });
      
      setGradedSubmissions(graded);
      setUngradedSubmissions(ungraded);
    } catch (error) {
      console.error('Error fetching submissions data:', error);
    } finally {
      setLoading(false);
    }
  };

  const calculateMaxScore = () => {
    return assignment.rubric?.weights?.reduce((sum, weight) => sum + weight, 0) || 100;
  };

  const getScoreColor = (score) => {
    const maxScore = calculateMaxScore();
    const percentage = (score / maxScore) * 100;
    if (percentage >= 90) return 'score-excellent';
    if (percentage >= 80) return 'score-good';
    if (percentage >= 70) return 'score-average';
    return 'score-poor';
  };

  if (loading) {
    return (
      <div className="graded-submissions-view">
        <div className="view-header">
          <button className="back-button" onClick={onBack}>
            ← Back
          </button>
          <h2>{assignment.name} - Submissions</h2>
        </div>
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>Loading submissions...</p>
        </div>
      </div>
    );
  }

  const totalSubmissions = gradedSubmissions.length + ungradedSubmissions.length;

  return (
    <div className="graded-submissions-view">
      <div className="view-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Assignments
        </button>
        <div className="header-info">
          <h2>{assignment.name}</h2>
          <p className="submission-stats">
            {totalSubmissions} total • {gradedSubmissions.length} graded • {ungradedSubmissions.length} pending
          </p>
        </div>
      </div>

      <div className="submissions-tabs">
        <button
          className={`tab-btn ${activeTab === 'all' ? 'active' : ''}`}
          onClick={() => setActiveTab('all')}
        >
          All ({totalSubmissions})
        </button>
        <button
          className={`tab-btn ${activeTab === 'graded' ? 'active' : ''}`}
          onClick={() => setActiveTab('graded')}
        >
          Graded ({gradedSubmissions.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'ungraded' ? 'active' : ''}`}
          onClick={() => setActiveTab('ungraded')}
        >
          Pending ({ungradedSubmissions.length})
        </button>
      </div>

      <div className="submissions-content">
        {activeTab === 'all' && (
          <>
            {totalSubmissions === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">📄</div>
                <h3>No Submissions Yet</h3>
              </div>
            ) : (
              <div className="submissions-list">
                {gradedSubmissions.map(sub => (
                  <div key={sub.id} className="submission-card graded">
                    <div className="submission-header">
                      <div className="student-info">
                        <div className="student-avatar">
                          {sub.student_name.charAt(0).toUpperCase()}
                        </div>
                        <div className="student-details">
                          <div className="student-name">{sub.student_name}</div>
                          <div className="student-id">{sub.student_id}</div>
                        </div>
                      </div>
                      <div className={`score-badge ${getScoreColor(sub.grade.total_score)}`}>
                        {sub.grade.total_score.toFixed(1)} / {calculateMaxScore()}
                      </div>
                    </div>
                    <div className="submission-meta">
                      <span>Graded by {sub.grade.grader_name}</span>
                      <span>•</span>
                      <span>{new Date(sub.grade.graded_at).toLocaleDateString()}</span>
                    </div>
                    <button 
                      className="btn btn-primary"
                      onClick={() => onViewSubmission(sub)}
                    >
                      View Details
                    </button>
                  </div>
                ))}
                {ungradedSubmissions.map(sub => (
                  <div key={sub.id} className="submission-card ungraded">
                    <div className="submission-header">
                      <div className="student-info">
                        <div className="student-avatar pending">
                          {sub.student_name.charAt(0).toUpperCase()}
                        </div>
                        <div className="student-details">
                          <div className="student-name">{sub.student_name}</div>
                          <div className="student-id">{sub.student_id}</div>
                        </div>
                      </div>
                      <span className="status-badge">Pending</span>
                    </div>
                    <div className="submission-meta">
                      <span>Submitted {new Date(sub.uploaded_at).toLocaleDateString()}</span>
                    </div>
                    <button 
                      className="btn btn-outline"
                      onClick={() => onViewSubmission(sub)}
                    >
                      View PDF
                    </button>
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {activeTab === 'graded' && (
          <>
            {gradedSubmissions.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">✅</div>
                <h3>No Graded Submissions</h3>
              </div>
            ) : (
              <div className="submissions-list">
                {gradedSubmissions.map(sub => (
                  <div key={sub.id} className="submission-card graded">
                    <div className="submission-header">
                      <div className="student-info">
                        <div className="student-avatar">
                          {sub.student_name.charAt(0).toUpperCase()}
                        </div>
                        <div className="student-details">
                          <div className="student-name">{sub.student_name}</div>
                          <div className="student-id">{sub.student_id}</div>
                        </div>
                      </div>
                      <div className={`score-badge ${getScoreColor(sub.grade.total_score)}`}>
                        {sub.grade.total_score.toFixed(1)} / {calculateMaxScore()}
                      </div>
                    </div>
                    <div className="submission-meta">
                      <span>Graded by {sub.grade.grader_name}</span>
                      <span>•</span>
                      <span>{new Date(sub.grade.graded_at).toLocaleDateString()}</span>
                    </div>
                    <button 
                      className="btn btn-primary"
                      onClick={() => onViewSubmission(sub)}
                    >
                      View Details
                    </button>
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {activeTab === 'ungraded' && (
          <>
            {ungradedSubmissions.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">⏳</div>
                <h3>No Pending Submissions</h3>
                <p>All submissions have been graded!</p>
              </div>
            ) : (
              <div className="submissions-list">
                {ungradedSubmissions.map(sub => (
                  <div key={sub.id} className="submission-card ungraded">
                    <div className="submission-header">
                      <div className="student-info">
                        <div className="student-avatar pending">
                          {sub.student_name.charAt(0).toUpperCase()}
                        </div>
                        <div className="student-details">
                          <div className="student-name">{sub.student_name}</div>
                          <div className="student-id">{sub.student_id}</div>
                        </div>
                      </div>
                      <span className="status-badge">Pending</span>
                    </div>
                    <div className="submission-meta">
                      <span>Submitted {new Date(sub.uploaded_at).toLocaleDateString()}</span>
                    </div>
                    <button 
                      className="btn btn-outline"
                      onClick={() => onViewSubmission(sub)}
                    >
                      View PDF
                    </button>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default GradedSubmissionsView;

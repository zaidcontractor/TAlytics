import React, { useState, useEffect } from 'react';
import axios from 'axios';
import CommentThread from './CommentThread';
import './GradingViewer.css';

const GradingViewer = ({ assignment, onBack, userRole }) => {
  const [submissions, setSubmissions] = useState([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [pdfUrl, setPdfUrl] = useState(null);
  const [rubricScores, setRubricScores] = useState({});
  const [gradeId, setGradeId] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchSubmissions();
  }, [assignment]);

  useEffect(() => {
    if (submissions.length > 0) {
      const submission = submissions[currentIndex];
      loadPdfForSubmission(submission);
      loadExistingGrade(submission);
    }
  }, [currentIndex, submissions]);

  const fetchSubmissions = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      const response = await axios.get(
        `http://localhost:5000/api/submissions/assignment/${assignment.id}`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      setSubmissions(response.data.submissions || []);
    } catch (error) {
      console.error('Error fetching submissions:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadExistingGrade = async (submission) => {
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(
        `http://localhost:5000/api/grades/submission/${submission.id}`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      if (response.data.grade) {
        setRubricScores(response.data.grade.rubric_scores || {});
        setGradeId(response.data.grade.id);
      } else {
        // Initialize empty scores
        const initialScores = {};
        assignment.rubric?.criteria?.forEach((criterion, index) => {
          initialScores[index] = 0;
        });
        setRubricScores(initialScores);
        setGradeId(null);
      }
    } catch (error) {
      // No grade exists yet
      const initialScores = {};
      assignment.rubric?.criteria?.forEach((criterion, index) => {
        initialScores[index] = 0;
      });
      setRubricScores(initialScores);
      setGradeId(null);
    }
  };

  const loadPdfForSubmission = async (submission) => {
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(
        `http://localhost:5000/api/submissions/${submission.id}/file`,
        { 
          headers: { Authorization: `Bearer ${token}` },
          responseType: 'blob'
        }
      );
      const url = URL.createObjectURL(response.data);
      setPdfUrl(url);
    } catch (error) {
      console.error('Error loading PDF:', error);
    }
  };

  const handleScoreChange = (criterionIndex, value) => {
    const maxScore = assignment.rubric.weights[criterionIndex];
    const numValue = parseFloat(value) || 0;
    const clampedValue = Math.min(Math.max(0, numValue), maxScore);
    setRubricScores({
      ...rubricScores,
      [criterionIndex]: clampedValue
    });
  };

  const calculateTotalScore = () => {
    return Object.values(rubricScores).reduce((sum, score) => sum + (parseFloat(score) || 0), 0);
  };

  const calculateMaxScore = () => {
    return assignment.rubric?.weights?.reduce((sum, weight) => sum + weight, 0) || 100;
  };

  const handleSubmitGrade = async () => {
    try {
      setSaving(true);
      const token = localStorage.getItem('token');
      const currentSubmission = submissions[currentIndex];
      
      const response = await axios.post(
        `http://localhost:5000/api/grades`,
        {
          assignment_id: assignment.id,
          submission_id: currentSubmission.id,
          student_id: currentSubmission.student_id,
          rubric_scores: rubricScores,
          total_score: calculateTotalScore()
        },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      // Update gradeId if this is a new grade
      if (response.data.grade) {
        setGradeId(response.data.grade.id);
      }
      
      alert('Grade submitted successfully!');
      
      // Move to next submission if available
      if (currentIndex < submissions.length - 1) {
        setCurrentIndex(currentIndex + 1);
      }
    } catch (error) {
      console.error('Error submitting grade:', error);
      alert('Failed to submit grade: ' + (error.response?.data || error.message));
    } finally {
      setSaving(false);
    }
  };

  const goToPrevious = () => {
    if (currentIndex > 0) {
      setCurrentIndex(currentIndex - 1);
    }
  };

  const goToNext = () => {
    if (currentIndex < submissions.length - 1) {
      setCurrentIndex(currentIndex + 1);
    }
  };

  if (loading) {
    return (
      <div className="grading-viewer">
        <div className="viewer-header">
          <button className="back-button" onClick={onBack}>
            ← Back
          </button>
        </div>
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>Loading submissions...</p>
        </div>
      </div>
    );
  }

  if (submissions.length === 0) {
    return (
      <div className="grading-viewer">
        <div className="viewer-header">
          <button className="back-button" onClick={onBack}>
            ← Back
          </button>
        </div>
        <div className="empty-state">
          <div className="empty-icon">📄</div>
          <h3>No Submissions Yet</h3>
          <p>No submissions available for grading.</p>
        </div>
      </div>
    );
  }

  const currentSubmission = submissions[currentIndex];

  return (
    <div className="grading-viewer">
      <div className="viewer-header">
        <button className="back-button" onClick={onBack}>
          ← Back
        </button>
        <div className="submission-info">
          <h2>{assignment.name}</h2>
          <div className="student-info">
            <span className="student-name">{currentSubmission.student_name}</span>
            <span className="student-id">({currentSubmission.student_id})</span>
          </div>
        </div>
      </div>

      <div className="viewer-controls">
        <div className="nav-buttons">
          <button 
            className="nav-button"
            onClick={goToPrevious}
            disabled={currentIndex === 0}
          >
            <span className="nav-button-icon">←</span> Previous
          </button>
          <button 
            className="nav-button"
            onClick={goToNext}
            disabled={currentIndex === submissions.length - 1}
          >
            Next <span className="nav-button-icon">→</span>
          </button>
        </div>
        <span className="submission-counter">
          {currentIndex + 1} / {submissions.length}
        </span>
      </div>

      <div className="grading-content">
        <div className="pdf-section">
          {pdfUrl ? (
            <iframe
              src={pdfUrl}
              title={`Submission ${currentSubmission.student_name}`}
              className="pdf-viewer"
            />
          ) : (
            <div className="loading-pdf">
              <div className="spinner"></div>
              <p>Loading PDF...</p>
            </div>
          )}
        </div>

        <div className="grading-section">
          <div className="rubric-card">
            <h3>Grading Rubric</h3>
            
            {assignment.rubric?.criteria?.map((criterion, index) => (
              <div key={index} className="rubric-item">
                <div className="criterion-header">
                  <label className="criterion-name">{criterion}</label>
                  <span className="criterion-weight">Max: {assignment.rubric.weights[index]} pts</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max={assignment.rubric.weights[index]}
                  step="0.5"
                  value={rubricScores[index] || 0}
                  onChange={(e) => handleScoreChange(index, e.target.value)}
                  className="score-input"
                  disabled={userRole === 'instructor'}
                />
              </div>
            ))}

            <div className="total-score">
              <span className="total-label">Total Score:</span>
              <span className="total-value">
                {calculateTotalScore().toFixed(1)} / {calculateMaxScore()}
              </span>
            </div>
          </div>

          <div className="comments-card">
            <CommentThread gradeId={gradeId} />
          </div>

          {userRole !== 'instructor' && (
            <button
              onClick={handleSubmitGrade}
              disabled={saving}
              className="btn btn-primary submit-grade-btn"
            >
              {saving ? 'Saving...' : 'Submit Grade'}
            </button>
          )}
        </div>
      </div>

      <div className="submissions-list">
        <h3>All Submissions</h3>
        <div className="submissions-grid">
          {submissions.map((sub, index) => (
            <div 
              key={sub.id}
              className={`submission-item ${index === currentIndex ? 'active' : ''}`}
              onClick={() => setCurrentIndex(index)}
            >
              <div className="submission-icon">📄</div>
              <div className="submission-details">
                <div className="submission-name">{sub.student_name}</div>
                <div className="submission-id">{sub.student_id}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default GradingViewer;

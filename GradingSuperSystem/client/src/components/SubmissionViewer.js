import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './SubmissionViewer.css';

const SubmissionViewer = ({ assignment, onBack }) => {
  const [submissions, setSubmissions] = useState([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [pdfUrl, setPdfUrl] = useState(null);

  useEffect(() => {
    fetchSubmissions();
  }, [assignment]);

  useEffect(() => {
    if (submissions.length > 0) {
      loadPdfForSubmission(submissions[currentIndex]);
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
      <div className="submission-viewer">
        <div className="viewer-header">
          <button className="back-button" onClick={onBack}>
            ← Back to Assignment
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
      <div className="submission-viewer">
        <div className="viewer-header">
          <button className="back-button" onClick={onBack}>
            ← Back to Assignment
          </button>
        </div>
        <div className="empty-state">
          <div className="empty-icon">📄</div>
          <h3>No Submissions Yet</h3>
          <p>Upload student submission PDFs to start grading.</p>
        </div>
      </div>
    );
  }

  const currentSubmission = submissions[currentIndex];

  return (
    <div className="submission-viewer">
      <div className="viewer-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Assignment
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
        <button 
          className="nav-btn"
          onClick={goToPrevious}
          disabled={currentIndex === 0}
        >
          ← Previous
        </button>
        <span className="submission-counter">
          {currentIndex + 1} / {submissions.length}
        </span>
        <button 
          className="nav-btn"
          onClick={goToNext}
          disabled={currentIndex === submissions.length - 1}
        >
          Next →
        </button>
      </div>

      <div className="pdf-container">
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

export default SubmissionViewer;

import React, { useState } from 'react';
import axios from 'axios';
import './Modal.css';

const JoinCourseModal = ({ onClose, onCourseJoined }) => {
  const [courseCode, setCourseCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const token = localStorage.getItem('token');
      const response = await axios.post('http://localhost:5000/api/courses/join', 
        { course_code: courseCode.toUpperCase() },
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      );

      onCourseJoined(response.data);
    } catch (err) {
      console.error('Error joining course:', err);
      // Backend returns plain text errors, not JSON
      setError(err.response?.data || err.message || 'Failed to join course. Please check the code and try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCodeChange = (e) => {
    // Auto-uppercase and limit to 6 characters
    const value = e.target.value.toUpperCase().slice(0, 6);
    setCourseCode(value);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Join Course</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <form onSubmit={handleSubmit} className="modal-form">
          {error && (
            <div className="alert alert-error">
              {error}
            </div>
          )}

          <div className="join-course-info">
            <div className="info-icon">🎓</div>
            <p>Enter the 6-character course code provided by your instructor to join their course.</p>
          </div>

          <div className="form-group">
            <label htmlFor="courseCode">Course Code *</label>
            <input
              type="text"
              id="courseCode"
              value={courseCode}
              onChange={handleCodeChange}
              placeholder="ABC123"
              className="course-code-input"
              pattern="[A-Z0-9]{6}"
              title="6-character alphanumeric code"
              required
              disabled={loading}
              autoFocus
            />
            <small className="form-help">
              Course codes are 6 characters long (letters and numbers)
            </small>
          </div>

          <div className="modal-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-success"
              disabled={loading || courseCode.length !== 6}
            >
              {loading ? (
                <>
                  <span className="spinner-small"></span>
                  Joining...
                </>
              ) : (
                'Join Course'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default JoinCourseModal;
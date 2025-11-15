import React, { useState } from 'react';
import axios from 'axios';
import './Modal.css';

const CreateCourseModal = ({ onClose, onCourseCreated }) => {
  const [formData, setFormData] = useState({
    name: '',
    code: '',
    description: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const generateCourseCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let code = '';
    for (let i = 0; i < 6; i++) {
      code += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setFormData(prev => ({ ...prev, code }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const token = localStorage.getItem('token');
      const response = await axios.post('http://localhost:5000/api/courses', formData, {
        headers: { Authorization: `Bearer ${token}` }
      });

      onCourseCreated(response.data);
    } catch (err) {
      console.error('Error creating course:', err);
      setError(err.response?.data?.error || 'Failed to create course. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Create New Course</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <form onSubmit={handleSubmit} className="modal-form">
          {error && (
            <div className="alert alert-error">
              {error}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="name">Course Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="e.g., Data Structures and Algorithms"
              required
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="code">Course Code *</label>
            <div className="code-input-group">
              <input
                type="text"
                id="code"
                name="code"
                value={formData.code}
                onChange={handleChange}
                placeholder="6-character code"
                maxLength={6}
                pattern="[A-Z0-9]{6}"
                title="6-character alphanumeric code"
                required
                disabled={loading}
              />
              <button
                type="button"
                className="btn btn-outline-small"
                onClick={generateCourseCode}
                disabled={loading}
              >
                Generate
              </button>
            </div>
            <small className="form-help">
              Students will use this code to join your course
            </small>
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              placeholder="Brief description of the course..."
              rows={3}
              disabled={loading}
            />
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
              className="btn btn-primary"
              disabled={loading || !formData.name || !formData.code}
            >
              {loading ? (
                <>
                  <span className="spinner-small"></span>
                  Creating...
                </>
              ) : (
                'Create Course'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateCourseModal;
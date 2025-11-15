import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import axios from 'axios';
import CreateCourseModal from './CreateCourseModal';
import JoinCourseModal from './JoinCourseModal';
import './CourseList.css';

const CourseList = ({ onCourseSelect, canCreateCourse }) => {
  // eslint-disable-next-line no-unused-vars
  const { user } = useAuth();
  const [courses, setCourses] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);

  useEffect(() => {
    fetchCourses();
  }, []);

  const fetchCourses = async () => {
    try {
      setLoading(true);
      setError('');
      
      const token = localStorage.getItem('token');
      const response = await axios.get('http://localhost:5000/api/courses/my-courses', {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      setCourses(response.data.courses || []);
    } catch (err) {
      console.error('Error fetching courses:', err);
      setError('Failed to load courses. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCourseCreated = (newCourse) => {
    setCourses(prev => [...prev, newCourse]);
    setShowCreateModal(false);
  };

  const handleCourseJoined = (joinedCourse) => {
    setCourses(prev => [...prev, joinedCourse]);
    setShowJoinModal(false);
  };

  const getRoleInCourse = (course) => {
    return course.role || 'member';
  };

  const getRoleBadgeColor = (role) => {
    switch (role) {
      case 'instructor':
        return 'instructor';
      case 'ta':
        return 'ta';
      default:
        return 'member';
    }
  };

  if (loading) {
    return (
      <div className="course-list">
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>Loading courses...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="course-list">
      {/* Header Actions */}
      <div className="course-list-header">
        <div className="header-actions">
          {canCreateCourse && (
            <button 
              className="btn btn-primary"
              onClick={() => setShowCreateModal(true)}
            >
              <span className="btn-icon">➕</span>
              Create Course
            </button>
          )}
          <button 
            className="btn btn-success"
            onClick={() => setShowJoinModal(true)}
          >
            <span className="btn-icon">🔗</span>
            Join Course
          </button>
        </div>
      </div>

      {error && (
        <div className="alert alert-error">
          {error}
          <button onClick={fetchCourses} className="retry-button">
            Retry
          </button>
        </div>
      )}

      {courses.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📚</div>
          <h3>No courses yet</h3>
          <p>
            {canCreateCourse 
              ? "Create your first course or join an existing one to get started."
              : "Join a course using the course code provided by your instructor."
            }
          </p>
          <div className="empty-actions">
            {canCreateCourse && (
              <button 
                className="btn btn-primary"
                onClick={() => setShowCreateModal(true)}
              >
                Create Course
              </button>
            )}
            <button 
              className="btn btn-outline"
              onClick={() => setShowJoinModal(true)}
            >
              Join Course
            </button>
          </div>
        </div>
      ) : (
        <div className="courses-grid">
          {courses.map(course => (
            <div 
              key={course.id} 
              className="course-card"
              onClick={() => onCourseSelect(course)}
            >
              <div className="course-header">
                <h3 className="course-name">{course.name}</h3>
                <span className={`role-badge ${getRoleBadgeColor(getRoleInCourse(course))}`}>
                  {getRoleInCourse(course) === 'instructor' ? 'Instructor' : 'TA'}
                </span>
              </div>
              
              <div className="course-details">
                <div className="course-code">{course.code}</div>
                <div className="course-description">{course.description}</div>
              </div>
              
              <div className="course-stats">
                <div className="stat">
                  <span className="stat-label">Assignments:</span>
                  <span className="stat-value">{course.assignment_count || 0}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Members:</span>
                  <span className="stat-value">{course.member_count || 1}</span>
                </div>
              </div>
              
              <div className="course-footer">
                <span className="course-date">
                  Created {new Date(course.created_at).toLocaleDateString()}
                </span>
                <button className="course-action-btn">
                  {getRoleInCourse(course) === 'instructor' ? 'Manage →' : 'Grade →'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modals */}
      {showCreateModal && (
        <CreateCourseModal 
          onClose={() => setShowCreateModal(false)}
          onCourseCreated={handleCourseCreated}
        />
      )}
      
      {showJoinModal && (
        <JoinCourseModal 
          onClose={() => setShowJoinModal(false)}
          onCourseJoined={handleCourseJoined}
        />
      )}
    </div>
  );
};

export default CourseList;
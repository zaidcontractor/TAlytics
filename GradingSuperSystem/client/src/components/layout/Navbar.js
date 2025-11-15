import React from 'react';
import { useAuth } from '../../contexts/AuthContext';
import './Navbar.css';

const Navbar = ({ activeView, setActiveView, selectedCourse, onBackToCourses }) => {
  const { user, logout } = useAuth();
  // eslint-disable-next-line no-unused-vars
  const { isInstructor, isTA } = useAuth();

  const menuItems = [
    {
      id: 'courses',
      label: 'My Courses',
      icon: '📚',
      show: true
    },
    {
      id: 'profile',
      label: 'Profile',
      icon: '👤',
      show: true
    }
  ];

  return (
    <nav className="navbar">
      <div className="navbar-header">
        <div className="navbar-brand">
          <h2>TAlytics</h2>
          <span className="navbar-subtitle">Grading Analytics</span>
        </div>
        
        <div className="user-info">
          <div className="user-avatar">
            {user?.name?.charAt(0)?.toUpperCase() || 'U'}
          </div>
          <div className="user-details">
            <div className="user-name">{user?.name}</div>
            <div className={`user-role ${user?.role}`}>
              {user?.role === 'instructor' ? 'Instructor' : 'TA'}
            </div>
          </div>
        </div>
      </div>

      {selectedCourse && (
        <div className="course-context">
          <button 
            className="back-button"
            onClick={onBackToCourses}
          >
            ← Back to Courses
          </button>
          <div className="current-course">
            <div className="course-name">{selectedCourse.name}</div>
            <div className="course-code">{selectedCourse.code}</div>
          </div>
        </div>
      )}

      <div className="navbar-menu">
        {menuItems
          .filter(item => item.show)
          .map(item => (
            <button
              key={item.id}
              className={`menu-item ${activeView === item.id ? 'active' : ''}`}
              onClick={() => setActiveView(item.id)}
            >
              <span className="menu-icon">{item.icon}</span>
              <span className="menu-label">{item.label}</span>
            </button>
          ))
        }
      </div>

      <div className="navbar-footer">
        <button 
          className="logout-button"
          onClick={logout}
        >
          <span className="menu-icon">🚪</span>
          <span className="menu-label">Sign Out</span>
        </button>
      </div>
    </nav>
  );
};

export default Navbar;
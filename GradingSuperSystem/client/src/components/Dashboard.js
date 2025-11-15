import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import CourseList from './courses/CourseList';
import InstructorDashboard from './InstructorDashboard';
import TADashboard from './TADashboard';
import Navbar from './layout/Navbar';
import './Dashboard.css';

const Dashboard = () => {
  const { user, isInstructor, isTA, logout } = useAuth();
  const [activeView, setActiveView] = useState('courses');
  const [selectedCourse, setSelectedCourse] = useState(null);
  const [courseListKey, setCourseListKey] = useState(0);

  useEffect(() => {
    // Set default view based on role
    if (isInstructor) {
      setActiveView('courses');
    } else if (isTA) {
      setActiveView('courses');
    }
  }, [isInstructor, isTA]);

  const handleCourseSelect = (course) => {
    setSelectedCourse(course);
    if (isInstructor) {
      setActiveView('manage-course');
    } else {
      setActiveView('grade-assignments');
    }
  };

  const handleBackToCourses = () => {
    setSelectedCourse(null);
    setActiveView('courses');
    // Force refresh of course list
    setCourseListKey(prev => prev + 1);
  };

  const renderContent = () => {
    switch (activeView) {
      case 'courses':
        return (
          <CourseList 
            key={courseListKey}
            onCourseSelect={handleCourseSelect}
            canCreateCourse={isInstructor}
          />
        );
      case 'manage-course':
        return isInstructor ? (
          <InstructorDashboard 
            course={selectedCourse}
            onBack={handleBackToCourses}
          />
        ) : null;
      case 'grade-assignments':
        return isTA ? (
          <TADashboard 
            course={selectedCourse}
            onBack={handleBackToCourses}
          />
        ) : null;
      case 'profile':
        return (
          <div className="profile-view">
            <h2>Profile</h2>
            <div className="profile-card">
              <div className="profile-info">
                <h3>{user?.name}</h3>
                <p>{user?.email}</p>
                <span className={`role-badge ${user?.role}`}>
                  {user?.role === 'instructor' ? 'Instructor' : 'Teaching Assistant'}
                </span>
              </div>
              <div className="profile-actions">
                <button onClick={logout} className="btn btn-outline">
                  Sign Out
                </button>
              </div>
            </div>
          </div>
        );
      default:
        return (
          <div className="welcome-view">
            <h2>Welcome to TAlytics</h2>
            <p>Select a course from the sidebar to get started.</p>
          </div>
        );
    }
  };

  return (
    <div className="dashboard">
      <Navbar 
        user={user}
        activeView={activeView}
        setActiveView={setActiveView}
        selectedCourse={selectedCourse}
        onBackToCourses={handleBackToCourses}
      />
      
      <main className="dashboard-content">
        <div className="content-header">
          <h1>
            {activeView === 'courses' && 'My Courses'}
            {activeView === 'manage-course' && `Managing: ${selectedCourse?.name}`}
            {activeView === 'grade-assignments' && `Grading: ${selectedCourse?.name}`}
            {activeView === 'profile' && 'Profile'}
          </h1>
        </div>
        
        <div className="content-body">
          {renderContent()}
        </div>
      </main>
    </div>
  );
};

export default Dashboard;
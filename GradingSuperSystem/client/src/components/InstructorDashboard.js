import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import AssignmentDetail from './AssignmentDetail';
import RubricBuilder from './RubricBuilder';
import RubricAIEditor from './RubricAIEditor';
import GradingViewer from './GradingViewer';
import GradedSubmissionsView from './GradedSubmissionsView';
import AssignmentAnalytics from './AssignmentAnalytics';
import './Dashboard.css';
import './InstructorDashboard.css';

const InstructorDashboard = ({ course, onBack }) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [assignments, setAssignments] = useState([]);
  const [members, setMembers] = useState([]);
  const [selectedAssignment, setSelectedAssignment] = useState(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [showEditForm, setShowEditForm] = useState(false);
  const [showAIRubricEditor, setShowAIRubricEditor] = useState(false);
  const [editingAssignment, setEditingAssignment] = useState(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState('');
  const [loading, setLoading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [rubricData, setRubricData] = useState(null);
  const [uploadedFiles, setUploadedFiles] = useState([]);
  const [viewingSubmissions, setViewingSubmissions] = useState(false);
  const [showGradedView, setShowGradedView] = useState(false);
  const [showAnalytics, setShowAnalytics] = useState(false);
  const [stats, setStats] = useState({
    totalAssignments: 0,
    totalSubmissions: 0,
    courseMembers: 0
  });

  const fetchCourseData = useCallback(async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      // Fetch course assignments
      const assignmentsResponse = await axios.get(
        `http://localhost:5000/api/courses/${course.id}/assignments`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      const assignmentsData = assignmentsResponse.data.assignments || [];
      setAssignments(assignmentsData);

      // Fetch course members  
      const membersResponse = await axios.get(
        `http://localhost:5000/api/courses/${course.id}/members`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      setMembers(membersResponse.data.members || []);

      // Calculate stats
      const totalSubmissions = assignmentsData.reduce((sum, assignment) => 
        sum + (assignment.submission_count || 0), 0
      );

      setStats({
        totalAssignments: assignmentsData.length,
        totalSubmissions: totalSubmissions,
        courseMembers: membersResponse.data.members?.length || 0
      });
    } catch (error) {
      console.error('Error fetching course data:', error);
    } finally {
      setLoading(false);
    }
  }, [course]);

  useEffect(() => {
    if (course) {
      fetchCourseData();
    }
  }, [course, fetchCourseData]);

  const handleCreateAssignment = async (assignmentData) => {
    try {
      const token = localStorage.getItem('token');
      
      // First create the rubric
      const rubricResponse = await axios.post(
        `http://localhost:5000/api/rubrics`,
        {
          course_id: course.id,
          name: rubricData.name,
          criteria: rubricData.criteria,
          weights: rubricData.weights
        },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      const rubricId = rubricResponse.data.rubric.id;
      
      // Then create the assignment with the rubric
      const assignmentResponse = await axios.post(
        `http://localhost:5000/api/assignments`,
        {
          ...assignmentData,
          rubric_id: rubricId
        },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      const assignmentId = assignmentResponse.data.assignment.id;
      
      // Upload all submitted files
      if (uploadedFiles.length > 0) {
        for (const file of uploadedFiles) {
          const formData = new FormData();
          formData.append('student_id', file.studentId);
          formData.append('student_name', file.studentName);
          formData.append('file', file.file);
          
          await axios.post(
            `http://localhost:5000/api/assignments/${assignmentId}/submissions`,
            formData,
            { 
              headers: { 
                Authorization: `Bearer ${token}`,
                'Content-Type': 'multipart/form-data'
              }
            }
          );
        }
      }
      
      setShowCreateForm(false);
      setRubricData(null);
      setUploadedFiles([]);
      fetchCourseData(); // Refresh data
    } catch (error) {
      console.error('Error creating assignment:', error);
      alert('Failed to create assignment: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleEditAssignment = (assignment) => {
    setEditingAssignment(assignment);
    setShowAIRubricEditor(true); // Use AI editor instead of regular edit form
  };

  const handleUpdateAssignment = async (assignmentData) => {
    try {
      const token = localStorage.getItem('token');
      
      // Note: Rubric update now happens in RubricAIEditor
      // This is kept for backward compatibility if needed
      
      // Update the assignment metadata
      await axios.put(
        `http://localhost:5000/api/assignments/${editingAssignment.id}`,
        {
          ...assignmentData,
          rubric_id: editingAssignment.rubric_id
        },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      setShowEditForm(false);
      setEditingAssignment(null);
      setRubricData(null);
      fetchCourseData(); // Refresh data
    } catch (error) {
      console.error('Error updating assignment:', error);
      alert('Failed to update assignment: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleAIRubricSave = () => {
    setShowAIRubricEditor(false);
    setEditingAssignment(null);
    fetchCourseData(); // Refresh to show updated rubric
  };

  const handleAIRubricCancel = () => {
    setShowAIRubricEditor(false);
    setEditingAssignment(null);
  };

  const handleViewAssignment = (assignment) => {
    setSelectedAssignment(assignment);
    setShowGradedView(true);
  };

  const handleViewSubmissionFromGradedView = (submission) => {
    setViewingSubmissions(true);
    setShowGradedView(false);
  };

  const handleBackToAssignments = () => {
    setSelectedAssignment(null);
    setViewingSubmissions(false);
    setShowGradedView(false);
    setShowAnalytics(false);
  };

  const handleViewAnalytics = (assignment) => {
    setSelectedAssignment(assignment);
    setShowAnalytics(true);
  };

  const handleDeleteCourse = async () => {
    try {
      setDeleting(true);
      const token = localStorage.getItem('token');
      await axios.delete(
        `http://localhost:5000/api/courses/${course.id}`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      setShowDeleteModal(false);
      // Navigate back to course list
      onBack();
    } catch (error) {
      console.error('Error deleting course:', error);
      alert('Failed to delete course. Please try again.');
    } finally {
      setDeleting(false);
    }
  };

  const copyJoinCode = async (joinCode) => {
    try {
      await navigator.clipboard.writeText(joinCode);
      alert('Join code copied to clipboard!');
    } catch (err) {
      console.error('Failed to copy:', err);
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = joinCode;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      alert('Join code copied to clipboard!');
    }
  };

  // Return analytics view if showing that
  if (showAnalytics && selectedAssignment) {
    return (
      <AssignmentAnalytics
        assignment={selectedAssignment}
        onBack={handleBackToAssignments}
      />
    );
  }

  // Return graded submissions view if showing that
  if (showGradedView && selectedAssignment) {
    return (
      <GradedSubmissionsView
        assignment={selectedAssignment}
        onBack={handleBackToAssignments}
        onViewSubmission={handleViewSubmissionFromGradedView}
      />
    );
  }

  // Return submission viewer if viewing submissions
  if (viewingSubmissions && selectedAssignment) {
    return (
      <GradingViewer 
        assignment={selectedAssignment} 
        onBack={() => {
          setViewingSubmissions(false);
          setShowGradedView(true);
        }}
        userRole="instructor"
      />
    );
  }

  // Return assignment detail view if assignment is selected
  if (selectedAssignment) {
    return (
      <AssignmentDetail 
        assignment={selectedAssignment} 
        onBack={handleBackToAssignments}
        userRole="instructor"
      />
    );
  }

  const renderOverview = () => (
    <div className="overview-tab">
      <div className="stats-cards">
        <div className="stat-card">
          <div className="stat-icon">📝</div>
          <div className="stat-content">
            <div className="stat-number">{stats.totalAssignments}</div>
            <div className="stat-label">Assignments</div>
          </div>
        </div>
        
        <div className="stat-card">
          <div className="stat-icon">📄</div>
          <div className="stat-content">
            <div className="stat-number">{stats.totalSubmissions}</div>
            <div className="stat-label">Submissions</div>
          </div>
        </div>
        
        <div className="stat-card">
          <div className="stat-icon">👥</div>
          <div className="stat-content">
            <div className="stat-number">{stats.courseMembers}</div>
            <div className="stat-label">Members</div>
          </div>
        </div>
      </div>

      <div className="course-info-card">
        <h3>Course Information</h3>
        <div className="course-details">
          <div className="detail-item">
            <span className="detail-label">Course Code:</span>
            <span className="detail-value course-code">{course.code}</span>
          </div>
          <div className="detail-item">
            <span className="detail-label">Description:</span>
            <span className="detail-value">{course.description || 'No description provided'}</span>
          </div>
          <div className="detail-item">
            <span className="detail-label">Created:</span>
            <span className="detail-value">{new Date(course.created_at).toLocaleDateString()}</span>
          </div>
        </div>
        
        <div className="join-code-section">
          <h4>Join Code</h4>
          <div className="join-code-display">
            <span className="join-code">{course.join_code || course.code}</span>
            <button 
              className="copy-btn"
              onClick={() => copyJoinCode(course.join_code || course.code)}
            >
              📋 Copy
            </button>
          </div>
          <p className="join-code-help">
            Share this code with TAs to let them join your course
          </p>
        </div>
        
        <div className="danger-zone">
          <h4>Danger Zone</h4>
          <p className="danger-zone-description">
            Deleting a course is permanent and cannot be undone. All assignments, grades, and member data will be lost.
          </p>
          <button 
            className="btn btn-danger"
            onClick={() => setShowDeleteModal(true)}
          >
            🗑️ Delete Course
          </button>
        </div>
      </div>
    </div>
  );

  const renderAssignments = () => (
    <div className="assignments-tab">
      <div className="tab-header">
        <h3>Course Assignments</h3>
        <button 
          className="btn btn-primary"
          onClick={() => setShowCreateForm(true)}
        >
          ➕ Create Assignment
        </button>
      </div>

      {assignments.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📝</div>
          <h4>No assignments yet</h4>
          <p>Create your first assignment to get started with grading.</p>
          <button 
            className="btn btn-primary"
            onClick={() => setShowCreateForm(true)}
          >
            Create Assignment
          </button>
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
                  className="btn btn-primary"
                  onClick={() => handleViewAnalytics(assignment)}
                >
                  📊 Analytics
                </button>
                <button 
                  className="btn btn-outline"
                  onClick={() => handleViewAssignment(assignment)}
                >
                  👁️ View Submissions
                </button>
                <button 
                  className="btn btn-secondary"
                  onClick={() => handleEditAssignment(assignment)}
                >
                  🤖 AI Edit Rubric
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  const renderMembers = () => (
    <div className="members-tab">
      <div className="tab-header">
        <h3>Course Members</h3>
        <span className="member-count">{members.length} members</span>
      </div>

      <div className="members-list">
        {members.map(member => (
          <div key={member.id || member.user_id} className="member-card">
            <div className="member-info">
              <div className="member-avatar">
                {member.name.charAt(0).toUpperCase()}
              </div>
              <div className="member-details">
                <div className="member-name">{member.name}</div>
                <div className="member-email">{member.email}</div>
              </div>
            </div>
            <div className="member-role">
              <span className={`role-badge ${member.role}`}>
                {member.role === 'instructor' ? 'Instructor' : 'TA'}
              </span>
            </div>
            <div className="member-joined">
              Joined {new Date(member.joined_at).toLocaleDateString()}
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  const renderCreateForm = () => (
    <div className="modal-overlay">
      <div className="modal-content modal-large">
        <div className="modal-header">
          <h3>Create New Assignment</h3>
            <button 
            className="modal-close"
            onClick={() => {
              setShowCreateForm(false);
              setRubricData(null);
              setUploadedFiles([]);
            }}
          >
            ×
          </button>
        </div>
        
        <form 
          onSubmit={(e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const assignmentData = {
              course_id: course.id,
              name: formData.get('name'),
              description: formData.get('description')
            };
            handleCreateAssignment(assignmentData);
          }}
          className="modal-form"
        >
          <div className="form-group">
            <label htmlFor="name">Assignment Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              placeholder="e.g., Homework 1"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              name="description"
              placeholder="Assignment description..."
              rows={3}
            />
          </div>

          <RubricBuilder 
            onRubricChange={setRubricData}
          />

          <div className="upload-submissions-section">
            <h4>Upload Student Submissions (PDFs)</h4>
            <p className="upload-help">
              Upload PDF submissions for grading. Each file should be named or you can specify student information.
            </p>
            
            <div className="file-upload-area">
              <input
                type="file"
                id="pdf-upload"
                accept=".pdf"
                multiple
                onChange={(e) => {
                  const files = Array.from(e.target.files);
                  const newFiles = files.map((file, index) => {
                    // Try to extract student info from filename
                    const nameMatch = file.name.match(/^([^_]+)_(.+)\.pdf$/);
                    return {
                      id: Date.now() + index,
                      file: file,
                      studentId: nameMatch ? nameMatch[1] : `student_${uploadedFiles.length + index + 1}`,
                      studentName: nameMatch ? nameMatch[2].replace(/_/g, ' ') : `Student ${uploadedFiles.length + index + 1}`
                    };
                  });
                  setUploadedFiles([...uploadedFiles, ...newFiles]);
                }}
                style={{ display: 'none' }}
              />
              <label htmlFor="pdf-upload" className="upload-button">
                📄 Choose PDF Files
              </label>
              <span className="file-count">
                {uploadedFiles.length} file{uploadedFiles.length !== 1 ? 's' : ''} selected
              </span>
            </div>

            {uploadedFiles.length > 0 && (
              <div className="uploaded-files-list">
                {uploadedFiles.map((file, index) => (
                  <div key={file.id} className="uploaded-file-item">
                    <span className="file-icon">📄</span>
                    <div className="file-info">
                      <input
                        type="text"
                        placeholder="Student ID"
                        value={file.studentId}
                        onChange={(e) => {
                          const updated = [...uploadedFiles];
                          updated[index].studentId = e.target.value;
                          setUploadedFiles(updated);
                        }}
                        className="student-input"
                      />
                      <input
                        type="text"
                        placeholder="Student Name"
                        value={file.studentName}
                        onChange={(e) => {
                          const updated = [...uploadedFiles];
                          updated[index].studentName = e.target.value;
                          setUploadedFiles(updated);
                        }}
                        className="student-input"
                      />
                      <span className="file-name">{file.file.name}</span>
                    </div>
                    <button
                      type="button"
                      className="remove-file-btn"
                      onClick={() => {
                        setUploadedFiles(uploadedFiles.filter(f => f.id !== file.id));
                      }}
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="modal-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => {
                setShowCreateForm(false);
                setRubricData(null);
                setUploadedFiles([]);
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={!rubricData || !rubricData.isValid}
            >
              Create Assignment
            </button>
          </div>
        </form>
      </div>
    </div>
  );

  const renderEditForm = () => (
    <div className="modal-overlay">
      <div className="modal-content modal-large">
        <div className="modal-header">
          <h3>Edit Assignment</h3>
          <button 
            className="modal-close"
            onClick={() => {
              setShowEditForm(false);
              setEditingAssignment(null);
              setRubricData(null);
            }}
          >
            ×
          </button>
        </div>
        
        <form 
          onSubmit={(e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const assignmentData = {
              name: formData.get('name'),
              description: formData.get('description')
            };
            handleUpdateAssignment(assignmentData);
          }}
          className="modal-form"
        >
          <div className="form-group">
            <label htmlFor="edit-name">Assignment Name *</label>
            <input
              type="text"
              id="edit-name"
              name="name"
              placeholder="e.g., Homework 1"
              defaultValue={editingAssignment?.name}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="edit-description">Description</label>
            <textarea
              id="edit-description"
              name="description"
              placeholder="Assignment description..."
              rows={3}
              defaultValue={editingAssignment?.description}
            />
          </div>

          <RubricBuilder 
            initialRubric={editingAssignment?.rubric}
            onRubricChange={setRubricData}
          />

          <div className="modal-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => {
                setShowEditForm(false);
                setEditingAssignment(null);
                setRubricData(null);
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={!rubricData || !rubricData.isValid}
            >
              Update Assignment
            </button>
          </div>
        </form>
      </div>
    </div>
  );

  const renderDeleteModal = () => (
    <div className="modal-overlay">
      <div className="modal-content delete-modal">
        <div className="modal-header">
          <h3>Delete Course</h3>
          <button 
            className="modal-close"
            onClick={() => {
              setShowDeleteModal(false);
              setDeleteConfirmText('');
            }}
            disabled={deleting}
          >
            ×
          </button>
        </div>
        
        <div className="modal-body">
          <div className="warning-icon">⚠️</div>
          <h4>Are you sure you want to delete "{course.name}"?</h4>
          <p>
            This action is <strong>permanent</strong> and cannot be undone. All data will be lost including:
          </p>
          <ul>
            <li>All assignments and their settings</li>
            <li>All grades and submissions</li>
            <li>All course members</li>
            <li>All rubrics and analysis data</li>
          </ul>
          <p>
            Type <strong>{course.code}</strong> below to confirm deletion:
          </p>
          <input
            type="text"
            placeholder={`Type "${course.code}" to confirm`}
            value={deleteConfirmText}
            onChange={(e) => setDeleteConfirmText(e.target.value)}
            className="confirm-input"
            disabled={deleting}
            autoFocus
          />
        </div>

        <div className="modal-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => {
              setShowDeleteModal(false);
              setDeleteConfirmText('');
            }}
            disabled={deleting}
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleDeleteCourse}
            disabled={deleteConfirmText !== course.code || deleting}
          >
            {deleting ? 'Deleting...' : 'Delete Course'}
          </button>
        </div>
      </div>
    </div>
  );

  const tabs = [
    { id: 'overview', label: 'Overview', icon: '📊' },
    { id: 'assignments', label: 'Assignments', icon: '📝', count: assignments.length },
    { id: 'members', label: 'Members', icon: '👥', count: members.length }
  ];

  return (
    <div className="instructor-dashboard">
      <div className="dashboard-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Courses
        </button>
        <div className="course-title">
          <h2>{course.name}</h2>
          <span className="instructor-badge">Instructor Dashboard</span>
        </div>
      </div>

      <div className="dashboard-tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
            {tab.count !== undefined && tab.count > 0 && (
              <span className="tab-count">{tab.count}</span>
            )}
          </button>
        ))}
      </div>

      <div className="dashboard-content">
        {loading ? (
          <div className="loading-spinner">
            <div className="spinner"></div>
            <p>Loading course data...</p>
          </div>
        ) : (
          <>
            {activeTab === 'overview' && renderOverview()}
            {activeTab === 'assignments' && renderAssignments()}
            {activeTab === 'members' && renderMembers()}
          </>
        )}
      </div>

      {showCreateForm && renderCreateForm()}
      {showEditForm && renderEditForm()}
      {showAIRubricEditor && editingAssignment && (
        <div className="modal-overlay">
          <div className="modal-content modal-extra-large">
            <RubricAIEditor
              rubric={editingAssignment.rubric}
              onSave={handleAIRubricSave}
              onCancel={handleAIRubricCancel}
            />
          </div>
        </div>
      )}
      {showDeleteModal && renderDeleteModal()}
    </div>
  );
};

export default InstructorDashboard;
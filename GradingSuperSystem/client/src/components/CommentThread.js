import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useAuth } from '../contexts/AuthContext';
import './CommentThread.css';

const CommentThread = ({ gradeId }) => {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [replyingTo, setReplyingTo] = useState(null);
  const [replyText, setReplyText] = useState('');
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  useEffect(() => {
    if (gradeId) {
      fetchComments();
    }
  }, [gradeId]);

  const fetchComments = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(
        `http://localhost:5000/api/grades/${gradeId}/comments`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setComments(response.data.comments || []);
    } catch (error) {
      console.error('Error fetching comments:', error);
      setComments([]);
    } finally {
      setLoading(false);
    }
  };

  const handlePostComment = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    try {
      const token = localStorage.getItem('token');
      const response = await axios.post(
        `http://localhost:5000/api/grades/${gradeId}/comments`,
        { message: newComment },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      setComments([...comments, response.data.comment]);
      setNewComment('');
    } catch (error) {
      console.error('Error posting comment:', error);
      alert('Failed to post comment');
    }
  };

  const handlePostReply = async (e, parentId) => {
    e.preventDefault();
    if (!replyText.trim()) return;

    try {
      const token = localStorage.getItem('token');
      const response = await axios.post(
        `http://localhost:5000/api/grades/${gradeId}/comments`,
        { message: replyText, parent_id: parentId },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      setComments([...comments, response.data.comment]);
      setReplyText('');
      setReplyingTo(null);
    } catch (error) {
      console.error('Error posting reply:', error);
      alert('Failed to post reply');
    }
  };

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);

    if (diffInSeconds < 60) return 'just now';
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} minutes ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
    if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)} days ago`;
    
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getAvatarColor = (role) => {
    return role === 'instructor' ? '#667eea' : '#48bb78';
  };

  const getInitials = (name) => {
    if (!name) return '?';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  };

  // Organize comments into threads
  const organizeComments = () => {
    const topLevel = comments.filter(c => !c.parent_id);
    const replies = comments.filter(c => c.parent_id);
    
    return topLevel.map(comment => ({
      ...comment,
      replies: replies.filter(r => r.parent_id === comment.id)
    }));
  };

  const renderComment = (comment, isReply = false) => {
    const isCurrentUser = user && user.id === comment.user_id;
    
    return (
      <div key={comment.id} className={`comment ${isReply ? 'comment-reply' : ''}`}>
        <div className="comment-avatar" style={{ backgroundColor: getAvatarColor(comment.user_role) }}>
          {getInitials(comment.user_name)}
        </div>
        <div className="comment-content">
          <div className="comment-header">
            <span className="comment-author">
              {comment.user_name}
              <span className={`role-badge ${comment.user_role}`}>
                {comment.user_role === 'instructor' ? '👨‍🏫 Instructor' : '👨‍💼 TA'}
              </span>
            </span>
            <span className="comment-timestamp">{formatTimestamp(comment.created_at)}</span>
          </div>
          <div className="comment-text">{comment.message}</div>
          {!isReply && (
            <button 
              className="reply-button"
              onClick={() => setReplyingTo(comment.id)}
            >
              Reply
            </button>
          )}
          
          {replyingTo === comment.id && (
            <form className="reply-form" onSubmit={(e) => handlePostReply(e, comment.id)}>
              <div className="comment-input-wrapper">
                <div className="comment-avatar small" style={{ backgroundColor: getAvatarColor(user?.role) }}>
                  {getInitials(user?.name)}
                </div>
                <input
                  type="text"
                  className="reply-input"
                  placeholder="Add a reply..."
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="reply-actions">
                <button type="button" className="cancel-button" onClick={() => setReplyingTo(null)}>
                  Cancel
                </button>
                <button type="submit" className="submit-button" disabled={!replyText.trim()}>
                  Reply
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    );
  };

  if (!gradeId) {
    return (
      <div className="comments-section">
        <div className="comments-placeholder">
          <p>💬 Submit a grade to start a conversation</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="comments-section">
        <div className="comments-loading">Loading comments...</div>
      </div>
    );
  }

  const threadedComments = organizeComments();

  return (
    <div className="comments-section">
      <div className="comments-header">
        <h3>💬 Discussion ({comments.length})</h3>
        <p className="comments-subtitle">Collaborate with instructors and TAs about this grade</p>
      </div>

      <form className="new-comment-form" onSubmit={handlePostComment}>
        <div className="comment-input-wrapper">
          <div className="comment-avatar" style={{ backgroundColor: getAvatarColor(user?.role) }}>
            {getInitials(user?.name)}
          </div>
          <input
            type="text"
            className="comment-input"
            placeholder="Add a comment..."
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
          />
        </div>
        <div className="comment-actions">
          <button 
            type="submit" 
            className="submit-button"
            disabled={!newComment.trim()}
          >
            Comment
          </button>
        </div>
      </form>

      <div className="comments-list">
        {threadedComments.length === 0 ? (
          <div className="no-comments">
            <p>No comments yet. Start the conversation!</p>
          </div>
        ) : (
          threadedComments.map(comment => (
            <div key={comment.id} className="comment-thread">
              {renderComment(comment)}
              {comment.replies && comment.replies.length > 0 && (
                <div className="comment-replies">
                  {comment.replies.map(reply => renderComment(reply, true))}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default CommentThread;

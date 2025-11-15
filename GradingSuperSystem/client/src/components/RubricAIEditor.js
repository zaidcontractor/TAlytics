import React, { useState } from 'react';
import axios from 'axios';
import './RubricAIEditor.css';

const RubricAIEditor = ({ rubric, onSave, onCancel }) => {
  const [loading, setLoading] = useState(false);
  const [suggestedRubric, setSuggestedRubric] = useState(null);
  const [selectedChanges, setSelectedChanges] = useState({});
  const [showComparison, setShowComparison] = useState(false);

  const fetchAISuggestions = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      const response = await axios.post(
        `http://localhost:5000/api/rubrics/${rubric.id}/ai-suggest`,
        {},
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      setSuggestedRubric(response.data.suggested_rubric);
      
      // Initialize all suggestions as selected by default
      const initialSelection = {};
      response.data.suggested_rubric.forEach((criterion, index) => {
        initialSelection[index] = true;
      });
      setSelectedChanges(initialSelection);
      setShowComparison(true);
    } catch (error) {
      console.error('Error fetching AI suggestions:', error);
      alert('Failed to get AI suggestions: ' + (error.response?.data || error.message));
    } finally {
      setLoading(false);
    }
  };

  const toggleCriterionSelection = (index) => {
    setSelectedChanges(prev => ({
      ...prev,
      [index]: !prev[index]
    }));
  };

  const handleApplyChanges = async () => {
    // Build the final rubric from selected changes
    // If a criterion is selected, use the AI suggestion; otherwise keep the original
    const finalCriteria = [];
    const finalWeights = [];
    
    suggestedRubric.forEach((criterion, index) => {
      if (selectedChanges[index]) {
        // Use AI suggestion
        finalCriteria.push(criterion.name);
        finalWeights.push(criterion.max_points);
      } else if (rubric.criteria?.[index]) {
        // Keep original criterion
        finalCriteria.push(rubric.criteria[index]);
        finalWeights.push(rubric.weights[index]);
      }
    });
    
    // Check if weights sum to 100
    const totalWeight = finalWeights.reduce((sum, w) => sum + w, 0);
    if (Math.abs(totalWeight - 100) > 0.1) {
      alert(`Criteria weights sum to ${totalWeight.toFixed(1)}%. They must sum to exactly 100%. Please adjust your selection.`);
      return;
    }
    
    if (finalCriteria.length === 0) {
      alert('You must select at least one criterion.');
      return;
    }
    
    // Confirm with instructor
    const confirmMessage = `This will update the rubric and mark ${rubric.usage_count || 0} existing grades for re-grading. All TAs will need to re-grade their submissions. Continue?`;
    if (!window.confirm(confirmMessage)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      
      // Update rubric with force_regrading flag
      await axios.put(
        `http://localhost:5000/api/rubrics/${rubric.id}`,
        {
          name: rubric.name,
          criteria: finalCriteria,
          weights: finalWeights,
          force_regrading: true
        },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      alert('Rubric updated successfully! All existing grades have been marked for re-grading.');
      onSave();
    } catch (error) {
      console.error('Error updating rubric:', error);
      alert('Failed to update rubric: ' + (error.response?.data || error.message));
    }
  };

  const renderCriterionComparison = (suggestedCriterion, index) => {
    const isSelected = selectedChanges[index];
    const originalCriterion = rubric.criteria?.[index];
    const originalWeight = rubric.weights?.[index];
    const isModified = !originalCriterion || 
                       originalCriterion !== suggestedCriterion.name || 
                       originalWeight !== suggestedCriterion.max_points;
    
    return (
      <div 
        key={index} 
        className={`criterion-comparison ${isSelected ? 'selected' : 'unselected'} ${isModified ? 'modified' : ''}`}
        onClick={() => toggleCriterionSelection(index)}
      >
        <div className="selection-checkbox">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={() => {}}
            onClick={(e) => e.stopPropagation()}
          />
        </div>
        
        <div className="comparison-content">
          <div className="criterion-header">
            <div className="criterion-index">#{index + 1}</div>
            <div className="criterion-change-badge">
              {!originalCriterion ? (
                <span className="badge badge-new">NEW</span>
              ) : isModified ? (
                <span className="badge badge-modified">MODIFIED</span>
              ) : (
                <span className="badge badge-unchanged">UNCHANGED</span>
              )}
            </div>
          </div>
          
          <div className="comparison-details">
            <div className="comparison-side original">
              <div className="side-label">Original:</div>
              {originalCriterion ? (
                <>
                  <div className="criterion-name">{originalCriterion}</div>
                  <div className="criterion-weight">{originalWeight} points</div>
                </>
              ) : (
                <div className="no-original">N/A (New criterion)</div>
              )}
            </div>
            
            <div className="comparison-arrow">→</div>
            
            <div className="comparison-side suggested">
              <div className="side-label">AI Suggested:</div>
              <div className="criterion-name suggested-text">{suggestedCriterion.name}</div>
              <div className="criterion-weight suggested-text">{suggestedCriterion.max_points} points</div>
            </div>
          </div>
          
          {suggestedCriterion.change_reason && (
            <div className="change-reason">
              <div className="reason-label">💡 AI Reasoning:</div>
              <div className="reason-text">{suggestedCriterion.change_reason}</div>
            </div>
          )}
        </div>
      </div>
    );
  };

  const calculateSelectedTotal = () => {
    if (!suggestedRubric) return 0;
    return suggestedRubric.reduce((sum, criterion, index) => {
      if (selectedChanges[index]) {
        // Use AI suggested weight
        return sum + criterion.max_points;
      } else if (rubric.criteria?.[index]) {
        // Use original weight
        return sum + rubric.weights[index];
      }
      return sum;
    }, 0);
  };

  if (!showComparison) {
    return (
      <div className="rubric-ai-editor">
        <div className="editor-header">
          <h3>🤖 AI-Powered Rubric Editor</h3>
          <p className="editor-description">
            Let Claude AI analyze your grading data and TA feedback to suggest improvements to this rubric.
          </p>
        </div>
        
        <div className="current-rubric-preview">
          <h4>Current Rubric: {rubric.name}</h4>
          <div className="rubric-criteria-list">
            {rubric.criteria?.map((criterion, index) => (
              <div key={index} className="rubric-criterion">
                <span className="criterion-index">{index + 1}.</span>
                <span className="criterion-name">{criterion}</span>
                <span className="criterion-weight">{rubric.weights[index]} pts</span>
              </div>
            ))}
          </div>
        </div>
        
        <div className="editor-actions">
          <button
            className="btn btn-secondary"
            onClick={onCancel}
          >
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={fetchAISuggestions}
            disabled={loading}
          >
            {loading ? (
              <>
                <span className="spinner-small"></span> Analyzing with AI...
              </>
            ) : (
              '✨ Get AI Suggestions'
            )}
          </button>
        </div>
      </div>
    );
  }

  const selectedTotal = calculateSelectedTotal();
  const isValidTotal = Math.abs(selectedTotal - 100) < 0.1;

  return (
    <div className="rubric-ai-editor comparison-mode">
      <div className="editor-header">
        <h3>🔄 Review AI Suggestions</h3>
        <p className="editor-description">
          Select the changes you want to apply. Click on any criterion to toggle selection.
        </p>
      </div>
      
      <div className="selection-summary">
        <div className="summary-stats">
          <div className="stat">
            <span className="stat-label">Selected:</span>
            <span className="stat-value">
              {Object.values(selectedChanges).filter(Boolean).length} / {suggestedRubric.length}
            </span>
          </div>
          <div className="stat">
            <span className="stat-label">Total Weight:</span>
            <span className={`stat-value ${isValidTotal ? 'valid' : 'invalid'}`}>
              {selectedTotal.toFixed(1)}%
              {isValidTotal ? ' ✓' : ' ⚠'}
            </span>
          </div>
        </div>
        {!isValidTotal && (
          <div className="weight-warning">
            ⚠ Selected criteria must sum to exactly 100%. Current total: {selectedTotal.toFixed(1)}%
          </div>
        )}
      </div>
      
      <div className="comparison-list">
        {suggestedRubric.map((criterion, index) => renderCriterionComparison(criterion, index))}
      </div>
      
      <div className="editor-actions sticky">
        <div className="action-info">
          <span className="info-icon">ℹ️</span>
          <span>Applying changes will require all TAs to re-grade existing submissions</span>
        </div>
        <div className="action-buttons">
          <button
            className="btn btn-secondary"
            onClick={() => setShowComparison(false)}
          >
            ← Back to Edit
          </button>
          <button
            className="btn btn-primary"
            onClick={handleApplyChanges}
            disabled={!isValidTotal || Object.values(selectedChanges).every(v => !v)}
          >
            Apply Selected Changes
          </button>
        </div>
      </div>
    </div>
  );
};

export default RubricAIEditor;

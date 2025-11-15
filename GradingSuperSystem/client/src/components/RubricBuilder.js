import React, { useState, useEffect } from 'react';
import './RubricBuilder.css';

const RubricBuilder = ({ initialRubric = null, onRubricChange, disabled = false }) => {
  const [rubricName, setRubricName] = useState('');
  const [criteria, setCriteria] = useState([
    { id: 1, name: '', weight: '' }
  ]);

  useEffect(() => {
    if (initialRubric) {
      setRubricName(initialRubric.name || '');
      if (initialRubric.criteria && initialRubric.weights) {
        const loadedCriteria = initialRubric.criteria.map((criterion, index) => ({
          id: index + 1,
          name: criterion,
          weight: initialRubric.weights[index]?.toString() || ''
        }));
        setCriteria(loadedCriteria);
      }
    }
  }, [initialRubric]);

  useEffect(() => {
    // Calculate if the rubric is valid
    const criteriaNames = criteria.map(c => c.name.trim()).filter(n => n !== '');
    const weights = criteria.map(c => parseFloat(c.weight) || 0);
    const totalWeight = weights.reduce((sum, w) => sum + w, 0);
    const isValid = 
      rubricName.trim() !== '' &&
      criteriaNames.length > 0 &&
      criteria.every(c => c.name.trim() !== '' && c.weight !== '' && parseFloat(c.weight) > 0) &&
      Math.abs(totalWeight - 100) < 0.01;

    if (onRubricChange) {
      onRubricChange({
        name: rubricName,
        criteria: criteriaNames,
        weights: weights,
        isValid: isValid,
        totalWeight: totalWeight
      });
    }
  }, [rubricName, criteria, onRubricChange]);

  const addCriterion = () => {
    const newId = Math.max(...criteria.map(c => c.id), 0) + 1;
    setCriteria([...criteria, { id: newId, name: '', weight: '' }]);
  };

  const removeCriterion = (id) => {
    if (criteria.length > 1) {
      setCriteria(criteria.filter(c => c.id !== id));
    }
  };

  const updateCriterion = (id, field, value) => {
    setCriteria(criteria.map(c => 
      c.id === id ? { ...c, [field]: value } : c
    ));
  };

  const calculateTotalWeight = () => {
    return criteria.reduce((sum, c) => sum + (parseFloat(c.weight) || 0), 0);
  };

  const totalWeight = calculateTotalWeight();
  const isValidTotal = Math.abs(totalWeight - 100) < 0.01;

  return (
    <div className="rubric-builder">
      <div className="rubric-header">
        <h4>Rubric Configuration</h4>
        <p className="rubric-help">
          Define grading criteria and their weights. All weights must sum to exactly 100%.
        </p>
      </div>

      <div className="form-group">
        <label htmlFor="rubric-name">Rubric Name *</label>
        <input
          type="text"
          id="rubric-name"
          value={rubricName}
          onChange={(e) => setRubricName(e.target.value)}
          placeholder="e.g., Homework 1 Rubric"
          disabled={disabled}
          required
        />
      </div>

      <div className="criteria-section">
        <div className="criteria-header">
          <h5>Grading Criteria</h5>
          <button
            type="button"
            className="btn btn-sm btn-secondary"
            onClick={addCriterion}
            disabled={disabled}
          >
            ➕ Add Criterion
          </button>
        </div>

        <div className="criteria-list">
          {criteria.map((criterion, index) => (
            <div key={criterion.id} className="criterion-item">
              <div className="criterion-number">{index + 1}</div>
              <div className="criterion-fields">
                <input
                  type="text"
                  placeholder="Criterion name (e.g., Problem 1, Code Quality)"
                  value={criterion.name}
                  onChange={(e) => updateCriterion(criterion.id, 'name', e.target.value)}
                  disabled={disabled}
                  required
                />
                <div className="weight-input-group">
                  <input
                    type="number"
                    placeholder="Weight"
                    value={criterion.weight}
                    onChange={(e) => updateCriterion(criterion.id, 'weight', e.target.value)}
                    min="0"
                    max="100"
                    step="0.1"
                    disabled={disabled}
                    required
                  />
                  <span className="weight-unit">%</span>
                </div>
              </div>
              {criteria.length > 1 && (
                <button
                  type="button"
                  className="btn-remove"
                  onClick={() => removeCriterion(criterion.id)}
                  disabled={disabled}
                  title="Remove criterion"
                >
                  🗑️
                </button>
              )}
            </div>
          ))}
        </div>

        <div className={`weight-total ${isValidTotal ? 'valid' : 'invalid'}`}>
          <span className="total-label">Total Weight:</span>
          <span className="total-value">
            {totalWeight.toFixed(1)}%
          </span>
          {isValidTotal ? (
            <span className="valid-badge">✓ Valid</span>
          ) : (
            <span className="invalid-badge">
              ⚠ Must equal 100%
            </span>
          )}
        </div>

        {!isValidTotal && totalWeight > 0 && (
          <div className="weight-warning">
            <p>
              ⚠️ The total weight is currently <strong>{totalWeight.toFixed(1)}%</strong>.
              Please adjust the weights so they sum to exactly 100%.
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default RubricBuilder;

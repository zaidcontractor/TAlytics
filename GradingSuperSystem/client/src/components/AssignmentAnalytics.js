import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './AssignmentAnalytics.css';

const AssignmentAnalytics = ({ assignment, onBack }) => {
  const [analytics, setAnalytics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [selectedCriteria, setSelectedCriteria] = useState([]);
  const [activeView, setActiveView] = useState('overview'); // overview, criteria, graders, distribution, insights
  const [aiInsights, setAiInsights] = useState(null);
  const [loadingInsights, setLoadingInsights] = useState(false);

  useEffect(() => {
    fetchAnalytics();
  }, [assignment]);

  useEffect(() => {
    // Initialize all criteria as selected
    if (assignment.rubric?.criteria) {
      setSelectedCriteria(assignment.rubric.criteria.map((_, index) => index));
    }
  }, [assignment]);

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      const response = await axios.get(
        `http://localhost:5000/api/assignments/${assignment.id}/analytics`,
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      setAnalytics(response.data);
    } catch (error) {
      console.error('Error fetching analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchAIInsights = async () => {
    if (!analytics) return;
    
    try {
      setLoadingInsights(true);
      const token = localStorage.getItem('token');
      
      const response = await axios.post(
        `http://localhost:5000/api/assignments/${assignment.id}/ai-insights`,
        { analytics_data: analytics },
        { headers: { Authorization: `Bearer ${token}` }}
      );
      
      setAiInsights(response.data.insights);
    } catch (error) {
      console.error('Error fetching AI insights:', error);
      setAiInsights('Failed to load AI insights. Please make sure CLAUDE_API_KEY is set in the server environment.');
    } finally {
      setLoadingInsights(false);
    }
  };

  const toggleCriterion = (index) => {
    setSelectedCriteria(prev => {
      if (prev.includes(index)) {
        return prev.filter(i => i !== index);
      } else {
        return [...prev, index];
      }
    });
  };

  const calculateFilteredDistribution = () => {
    if (!analytics || !analytics.grades || !assignment.rubric?.weights) return [];
    
    const scores = analytics.grades.map(grade => {
      let total = 0;
      selectedCriteria.forEach(criterionIndex => {
        total += grade.RubricScores?.[criterionIndex] || 0;
      });
      return total;
    });

    // Create histogram bins
    const maxScore = selectedCriteria.reduce((sum, index) => 
      sum + (assignment.rubric.weights[index] || 0), 0);
    
    if (maxScore === 0) return [];
    
    const binSize = Math.max(5, Math.ceil(maxScore / 10));
    const bins = {};
    
    for (let i = 0; i <= maxScore; i += binSize) {
      bins[i] = 0;
    }
    
    scores.forEach(score => {
      const bin = Math.floor(score / binSize) * binSize;
      bins[bin] = (bins[bin] || 0) + 1;
    });
    
    return Object.entries(bins).map(([score, count]) => ({
      range: `${score}-${parseInt(score) + binSize}`,
      count
    }));
  };

  const getGraderStats = () => {
    if (!analytics || !analytics.grader_stats) return [];
    
    return analytics.grader_stats.map(grader => {
      const criteriaAvgs = {};
      selectedCriteria.forEach(index => {
        criteriaAvgs[index] = grader.criteria_averages[index] || 0;
      });
      
      const filteredTotal = Object.values(criteriaAvgs).reduce((sum, val) => sum + val, 0);
      
      return {
        ...grader,
        filteredAverage: filteredTotal,
        criteriaAvgs
      };
    });
  };

  const getCriteriaStats = () => {
    if (!analytics || !analytics.criteria_stats || !assignment.rubric?.criteria) return [];
    
    return analytics.criteria_stats
      .map((stat, index) => ({
        index,
        name: assignment.rubric.criteria[index] || `Criterion ${index}`,
        maxScore: assignment.rubric.weights?.[index] || 0,
        ...stat
      }))
      .filter((_, index) => selectedCriteria.includes(index));
  };

  if (loading) {
    return (
      <div className="analytics-container">
        <div className="analytics-header">
          <button className="back-button" onClick={onBack}>
            ← Back to Assignment
          </button>
          <h2>Loading Analytics...</h2>
        </div>
      </div>
    );
  }

  if (!assignment.rubric || !assignment.rubric.criteria || !assignment.rubric.weights) {
    return (
      <div className="analytics-container">
        <div className="analytics-header">
          <button className="back-button" onClick={onBack}>
            ← Back to Assignment
          </button>
          <h2>📊 Assignment Analytics</h2>
        </div>
        <div className="no-data">
          <p>This assignment does not have a rubric configured. Analytics require a rubric to function.</p>
        </div>
      </div>
    );
  }

  if (!analytics || !analytics.grades || analytics.grades.length === 0) {
    return (
      <div className="analytics-container">
        <div className="analytics-header">
          <button className="back-button" onClick={onBack}>
            ← Back to Assignment
          </button>
          <h2>📊 Assignment Analytics</h2>
        </div>
        <div className="no-data">
          <p>No grades submitted yet. Analytics will appear once TAs start grading.</p>
        </div>
      </div>
    );
  }

  const distribution = calculateFilteredDistribution();
  const maxCount = Math.max(...distribution.map(d => d.count), 1);
  const graderStats = getGraderStats();
  const criteriaStats = getCriteriaStats();

  return (
    <div className="analytics-container">
      <div className="analytics-header">
        <button className="back-button" onClick={onBack}>
          ← Back to Assignment
        </button>
        <div>
          <h2>📊 {assignment.name} - Analytics</h2>
          <p className="analytics-subtitle">{analytics.total_grades} submissions graded</p>
        </div>
      </div>

      <div className="analytics-tabs">
        <button 
          className={`tab ${activeView === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveView('overview')}
        >
          📈 Overview
        </button>
        <button 
          className={`tab ${activeView === 'distribution' ? 'active' : ''}`}
          onClick={() => setActiveView('distribution')}
        >
          📊 Score Distribution
        </button>
        <button 
          className={`tab ${activeView === 'criteria' ? 'active' : ''}`}
          onClick={() => setActiveView('criteria')}
        >
          🎯 Criteria Analysis
        </button>
        <button 
          className={`tab ${activeView === 'graders' ? 'active' : ''}`}
          onClick={() => setActiveView('graders')}
        >
          👥 Grader Comparison
        </button>
        <button 
          className={`tab ${activeView === 'insights' ? 'active' : ''}`}
          onClick={() => {
            setActiveView('insights');
            if (!aiInsights && !loadingInsights) {
              fetchAIInsights();
            }
          }}
        >
          🤖 AI Insights
        </button>
      </div>

      {/* Criteria Filter */}
      <div className="criteria-filter">
        <h3>Filter by Criteria:</h3>
        <div className="filter-chips">
          {assignment.rubric.criteria.map((criterion, index) => (
            <button
              key={index}
              className={`filter-chip ${selectedCriteria.includes(index) ? 'selected' : ''}`}
              onClick={() => toggleCriterion(index)}
            >
              {criterion} ({assignment.rubric.weights[index]} pts)
              {selectedCriteria.includes(index) && ' ✓'}
            </button>
          ))}
        </div>
        <p className="filter-info">
          Selected: {selectedCriteria.length} / {assignment.rubric.criteria.length} criteria
          {' - '}
          Max Score: {selectedCriteria.reduce((sum, i) => sum + assignment.rubric.weights[i], 0)} pts
        </p>
      </div>

      {/* Overview Tab */}
      {activeView === 'overview' && (
        <div className="analytics-content">
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-icon">📊</div>
              <div className="stat-value">{analytics.overall_average.toFixed(2)}</div>
              <div className="stat-label">Overall Average</div>
              <div className="stat-sub">out of {analytics.max_score}</div>
            </div>
            
            <div className="stat-card">
              <div className="stat-icon">📈</div>
              <div className="stat-value">{analytics.highest_score.toFixed(2)}</div>
              <div className="stat-label">Highest Score</div>
              <div className="stat-sub">{((analytics.highest_score / analytics.max_score) * 100).toFixed(1)}%</div>
            </div>
            
            <div className="stat-card">
              <div className="stat-icon">📉</div>
              <div className="stat-value">{analytics.lowest_score.toFixed(2)}</div>
              <div className="stat-label">Lowest Score</div>
              <div className="stat-sub">{((analytics.lowest_score / analytics.max_score) * 100).toFixed(1)}%</div>
            </div>
            
            <div className="stat-card">
              <div className="stat-icon">📏</div>
              <div className="stat-value">{analytics.std_deviation.toFixed(2)}</div>
              <div className="stat-label">Std Deviation</div>
              <div className="stat-sub">Score spread</div>
            </div>
          </div>

          <div className="quick-distribution">
            <h3>Quick Score Distribution</h3>
            <div className="histogram">
              {distribution.map((bin, index) => (
                <div key={index} className="histogram-bar">
                  <div 
                    className="bar-fill"
                    style={{ height: `${(bin.count / maxCount) * 200}px` }}
                  >
                    <span className="bar-count">{bin.count}</span>
                  </div>
                  <div className="bar-label">{bin.range}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Distribution Tab */}
      {activeView === 'distribution' && (
        <div className="analytics-content">
          <div className="distribution-section">
            <h3>Score Distribution Histogram</h3>
            <p className="section-desc">
              Distribution based on {selectedCriteria.length} selected criteria
            </p>
            
            <div className="histogram-large">
              {distribution.map((bin, index) => (
                <div key={index} className="histogram-bar-large">
                  <div 
                    className="bar-fill-large"
                    style={{ height: `${(bin.count / maxCount) * 300}px` }}
                  >
                    <span className="bar-count-large">{bin.count}</span>
                  </div>
                  <div className="bar-label-large">{bin.range}</div>
                </div>
              ))}
            </div>
            
            <div className="distribution-stats">
              <div className="dist-stat">
                <span className="dist-label">Mean:</span>
                <span className="dist-value">
                  {(analytics.grades.reduce((sum, g) => 
                    sum + selectedCriteria.reduce((s, i) => s + (g.RubricScores?.[i] || 0), 0), 0
                  ) / analytics.grades.length).toFixed(2)}
                </span>
              </div>
              <div className="dist-stat">
                <span className="dist-label">Median:</span>
                <span className="dist-value">
                  {(() => {
                    const scores = analytics.grades.map(g => 
                      selectedCriteria.reduce((s, i) => s + (g.RubricScores?.[i] || 0), 0)
                    ).sort((a, b) => a - b);
                    const mid = Math.floor(scores.length / 2);
                    return scores.length % 2 === 0 
                      ? ((scores[mid - 1] + scores[mid]) / 2).toFixed(2)
                      : scores[mid].toFixed(2);
                  })()}
                </span>
              </div>
              <div className="dist-stat">
                <span className="dist-label">Mode:</span>
                <span className="dist-value">
                  {(() => {
                    const scores = analytics.grades.map(g => 
                      selectedCriteria.reduce((s, i) => s + (g.RubricScores?.[i] || 0), 0)
                    );
                    const freq = {};
                    scores.forEach(s => freq[s] = (freq[s] || 0) + 1);
                    const maxFreq = Math.max(...Object.values(freq));
                    const modes = Object.keys(freq).filter(k => freq[k] === maxFreq);
                    return modes[0];
                  })()}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Criteria Analysis Tab */}
      {activeView === 'criteria' && (
        <div className="analytics-content">
          <div className="criteria-analysis">
            <h3>Criteria Performance Analysis</h3>
            <p className="section-desc">Average scores for each selected criterion</p>
            
            {criteriaStats.map(stat => (
              <div key={stat.index} className="criterion-card">
                <div className="criterion-header">
                  <h4>{stat.name}</h4>
                  <span className="max-score">Max: {stat.maxScore} pts</span>
                </div>
                
                <div className="criterion-stats">
                  <div className="stat-row">
                    <span className="stat-label">Average Score:</span>
                    <span className="stat-value">
                      {stat.average.toFixed(2)} / {stat.maxScore}
                      <span className="percentage">
                        ({((stat.average / stat.maxScore) * 100).toFixed(1)}%)
                      </span>
                    </span>
                  </div>
                  
                  <div className="progress-bar">
                    <div 
                      className="progress-fill"
                      style={{ width: `${(stat.average / stat.maxScore) * 100}%` }}
                    />
                  </div>
                  
                  <div className="stat-row">
                    <span className="stat-label">Range:</span>
                    <span className="stat-value">
                      {stat.min.toFixed(2)} - {stat.max.toFixed(2)}
                    </span>
                  </div>
                  
                  <div className="stat-row">
                    <span className="stat-label">Std Deviation:</span>
                    <span className="stat-value">{stat.std_dev.toFixed(2)}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Grader Comparison Tab */}
      {activeView === 'graders' && (
        <div className="analytics-content">
          <div className="grader-analysis">
            <h3>Grader Comparison</h3>
            <p className="section-desc">
              Average scores given by each TA across selected criteria
            </p>
            
            {graderStats.map(grader => (
              <div key={grader.grader_id} className="grader-card">
                <div className="grader-header">
                  <div className="grader-info">
                    <div className="grader-avatar">{grader.grader_name.split(' ').map(n => n[0]).join('')}</div>
                    <div>
                      <h4>{grader.grader_name}</h4>
                      <p className="grader-meta">{grader.grade_count} submissions graded</p>
                    </div>
                  </div>
                  <div className="grader-overall">
                    <div className="overall-avg">{grader.filteredAverage.toFixed(2)}</div>
                    <div className="overall-label">Avg (selected criteria)</div>
                  </div>
                </div>
                
                <div className="criteria-breakdown">
                  <h5>Criteria Breakdown:</h5>
                  <div className="criteria-grid">
                    {selectedCriteria.map(index => (
                      <div key={index} className="criteria-stat">
                        <div className="criteria-name">{assignment.rubric.criteria[index]}</div>
                        <div className="criteria-score">
                          {grader.criteriaAvgs[index].toFixed(2)} / {assignment.rubric.weights[index]}
                        </div>
                        <div className="mini-progress">
                          <div 
                            className="mini-progress-fill"
                            style={{ 
                              width: `${(grader.criteriaAvgs[index] / assignment.rubric.weights[index]) * 100}%` 
                            }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* AI Insights Tab */}
      {activeView === 'insights' && (
        <div className="analytics-content">
          <div className="ai-insights-section">
            <div className="insights-header">
              <h3>🤖 AI-Powered Insights</h3>
              <p className="section-desc">
                Claude AI analyzes your grading data to provide actionable recommendations
              </p>
              {!loadingInsights && aiInsights && (
                <button 
                  className="refresh-insights-btn"
                  onClick={fetchAIInsights}
                >
                  🔄 Refresh Insights
                </button>
              )}
            </div>
            
            {loadingInsights ? (
              <div className="insights-loading">
                <div className="loading-spinner-large"></div>
                <p>Analyzing grading patterns with Claude AI...</p>
                <p className="loading-subtext">This may take 10-20 seconds</p>
              </div>
            ) : aiInsights ? (
              <div className="insights-content">
                <div className="ai-badge">
                  <span className="ai-icon">✨</span>
                  <span>Powered by Claude 3.5 Sonnet</span>
                </div>
                <div 
                  className="insights-text"
                  dangerouslySetInnerHTML={{ 
                    __html: aiInsights
                      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                      .replace(/\*(.*?)\*/g, '<em>$1</em>')
                      .replace(/^### (.*$)/gm, '<h3>$1</h3>')
                      .replace(/^## (.*$)/gm, '<h2>$1</h2>')
                      .replace(/^# (.*$)/gm, '<h1>$1</h1>')
                      .replace(/^- (.*$)/gm, '<li>$1</li>')
                      .replace(/\n\n/g, '<br><br>')
                      .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
                  }}
                />
              </div>
            ) : (
              <div className="insights-empty">
                <p>Click the button above to generate AI insights</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default AssignmentAnalytics;

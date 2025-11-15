package services

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"talytics/internal/database"
	"talytics/internal/models"
)

// AnomalyService handles anomaly detection and analysis
type AnomalyService struct{}

// NewAnomalyService creates a new anomaly service instance
func NewAnomalyService() *AnomalyService {
	return &AnomalyService{}
}

// GenerateAnomalyReport runs all detection algorithms and creates comprehensive report
func (s *AnomalyService) GenerateAnomalyReport(assignmentID int) (string, error) {
	// Get all grades for the assignment
	grades, err := database.GetAllGradesForAssignment(assignmentID)
	if err != nil {
		return "", fmt.Errorf("failed to get grades: %w", err)
	}

	if len(grades) == 0 {
		return "", fmt.Errorf("no grades found for assignment")
	}

	// Run all detection algorithms
	taSeverityIssues, err := s.DetectTASeverityDeviation(grades)
	if err != nil {
		return "", fmt.Errorf("failed to detect TA severity: %w", err)
	}

	outlierGrades, err := s.DetectOutlierGrades(grades)
	if err != nil {
		return "", fmt.Errorf("failed to detect outliers: %w", err)
	}

	criterionIssues, err := s.DetectCriterionInconsistency(grades)
	if err != nil {
		return "", fmt.Errorf("failed to detect criterion issues: %w", err)
	}

	regradeRisks, err := s.PredictRegradeRisk(grades, taSeverityIssues, outlierGrades)
	if err != nil {
		return "", fmt.Errorf("failed to predict regrade risks: %w", err)
	}

	// Calculate summary statistics
	avgScore, stdDev := s.calculateStatistics(grades)

	// Create summary report
	summary := models.AnomalySummary{
		AssignmentID:      assignmentID,
		TotalGrades:       len(grades),
		AverageScore:      avgScore,
		StandardDeviation: stdDev,
		TASeverityIssues:  taSeverityIssues,
		OutlierGrades:     outlierGrades,
		CriterionIssues:   criterionIssues,
		RegradeRisks:      regradeRisks,
		GeneratedAt:       time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(summary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	return string(jsonData), nil
}

// DetectTASeverityDeviation identifies TAs grading significantly harsher/more lenient than average
func (s *AnomalyService) DetectTASeverityDeviation(grades []map[string]interface{}) ([]models.TASeverityAnomaly, error) {
	// Calculate overall average
	totalScore := 0.0
	for _, grade := range grades {
		totalScore += grade["score"].(float64)
	}
	overallAvg := totalScore / float64(len(grades))

	// Calculate overall standard deviation
	variance := 0.0
	for _, grade := range grades {
		diff := grade["score"].(float64) - overallAvg
		variance += diff * diff
	}
	overallStd := math.Sqrt(variance / float64(len(grades)))

	// Group grades by TA
	taGrades := make(map[int][]float64)
	taEmails := make(map[int]string)

	for _, grade := range grades {
		taID := grade["graded_by"].(int)
		score := grade["score"].(float64)
		taGrades[taID] = append(taGrades[taID], score)
		taEmails[taID] = grade["grader_email"].(string)
	}

	// Analyze each TA
	var anomalies []models.TASeverityAnomaly

	for taID, scores := range taGrades {
		// Calculate TA average
		taTotal := 0.0
		for _, score := range scores {
			taTotal += score
		}
		taAvg := taTotal / float64(len(scores))

		// Calculate deviation in standard deviations
		deviation := (taAvg - overallAvg) / overallStd

		// Flag if deviation exceeds threshold (1.5 standard deviations)
		if math.Abs(deviation) > 1.5 {
			severity := "too_lenient"
			if deviation < 0 {
				severity = "too_harsh"
			}

			anomalies = append(anomalies, models.TASeverityAnomaly{
				TAID:         taID,
				TAEmail:      taEmails[taID],
				AverageScore: taAvg,
				GradesCount:  len(scores),
				Deviation:    deviation,
				Severity:     severity,
			})
		}
	}

	return anomalies, nil
}

// DetectOutlierGrades identifies grades that are statistical outliers (z-score > 2)
func (s *AnomalyService) DetectOutlierGrades(grades []map[string]interface{}) ([]models.OutlierAnomaly, error) {
	// Calculate mean and standard deviation
	mean, stdDev := s.calculateStatistics(grades)

	var outliers []models.OutlierAnomaly

	for _, grade := range grades {
		score := grade["score"].(float64)

		// Calculate z-score
		zScore := (score - mean) / stdDev

		// Flag if |z-score| > 2 (outside 2 standard deviations)
		if math.Abs(zScore) > 2.0 {
			outliers = append(outliers, models.OutlierAnomaly{
				SubmissionID:      grade["submission_id"].(int),
				StudentIdentifier: grade["student_identifier"].(string),
				Score:             score,
				ZScore:            zScore,
				GradedBy:          grade["graded_by"].(int),
				GraderEmail:       grade["grader_email"].(string),
			})
		}
	}

	return outliers, nil
}

// DetectCriterionInconsistency identifies rubric criteria with high variance across TAs
func (s *AnomalyService) DetectCriterionInconsistency(grades []map[string]interface{}) ([]models.CriterionAnomaly, error) {
	// Map to store criterion scores: criterion_name -> []scores
	criterionScores := make(map[string][]float64)
	criterionSubmissions := make(map[string][]int)

	// Parse rubric breakdowns
	for _, grade := range grades {
		rubricBreakdown, ok := grade["rubric_breakdown"].(string)
		if !ok || rubricBreakdown == "" {
			continue
		}

		// Parse JSON breakdown
		var breakdown map[string]float64
		if err := json.Unmarshal([]byte(rubricBreakdown), &breakdown); err != nil {
			continue // Skip invalid JSON
		}

		submissionID := grade["submission_id"].(int)

		// Add scores to criterion map
		for criterion, score := range breakdown {
			criterionScores[criterion] = append(criterionScores[criterion], score)
			criterionSubmissions[criterion] = append(criterionSubmissions[criterion], submissionID)
		}
	}

	var inconsistencies []models.CriterionAnomaly

	// Analyze each criterion
	for criterion, scores := range criterionScores {
		if len(scores) < 3 {
			continue // Need at least 3 data points
		}

		// Calculate mean and std deviation
		total := 0.0
		for _, score := range scores {
			total += score
		}
		mean := total / float64(len(scores))

		variance := 0.0
		for _, score := range scores {
			diff := score - mean
			variance += diff * diff
		}
		stdDev := math.Sqrt(variance / float64(len(scores)))

		// Flag if coefficient of variation is high (> 0.3)
		// CV = std_dev / mean
		cv := stdDev / mean
		if cv > 0.3 && !math.IsNaN(cv) && !math.IsInf(cv, 0) {
			// Find submissions with scores outside 1.5 std deviations
			var inconsistentSubmissions []int
			for i, score := range scores {
				zScore := (score - mean) / stdDev
				if math.Abs(zScore) > 1.5 {
					inconsistentSubmissions = append(inconsistentSubmissions, criterionSubmissions[criterion][i])
				}
			}

			inconsistencies = append(inconsistencies, models.CriterionAnomaly{
				CriterionName:      criterion,
				AverageScore:       mean,
				StandardDeviation:  stdDev,
				InconsistentGrades: inconsistentSubmissions,
			})
		}
	}

	return inconsistencies, nil
}

// PredictRegradeRisk calculates regrade risk score for each submission
func (s *AnomalyService) PredictRegradeRisk(grades []map[string]interface{}, taSeverity []models.TASeverityAnomaly, outliers []models.OutlierAnomaly) ([]models.RegradeRisk, error) {
	// Create lookup maps
	harshTAs := make(map[int]bool)
	lenientTAs := make(map[int]bool)
	for _, ta := range taSeverity {
		if ta.Severity == "too_harsh" {
			harshTAs[ta.TAID] = true
		} else {
			lenientTAs[ta.TAID] = true
		}
	}

	outlierSubmissions := make(map[int]float64) // submission_id -> z_score
	for _, outlier := range outliers {
		outlierSubmissions[outlier.SubmissionID] = outlier.ZScore
	}

	var risks []models.RegradeRisk

	for _, grade := range grades {
		submissionID := grade["submission_id"].(int)
		score := grade["score"].(float64)
		gradedBy := grade["graded_by"].(int)
		studentID := grade["student_identifier"].(string)
		graderEmail := grade["grader_email"].(string)

		riskScore := 0
		var riskFactors []string

		// Factor 1: Is outlier (30 points)
		if zScore, isOutlier := outlierSubmissions[submissionID]; isOutlier {
			riskScore += 30
			if zScore < 0 {
				riskFactors = append(riskFactors, "unusually_low_score")
			} else {
				riskFactors = append(riskFactors, "unusually_high_score")
			}
		}

		// Factor 2: Graded by harsh TA (25 points)
		if harshTAs[gradedBy] {
			riskScore += 25
			riskFactors = append(riskFactors, "harsh_grader")
		}

		// Factor 3: Graded by lenient TA (15 points)
		if lenientTAs[gradedBy] {
			riskScore += 15
			riskFactors = append(riskFactors, "lenient_grader")
		}

		// Factor 4: Near boundary grades (30 points)
		// Common boundaries: 60 (D/F), 70 (C/D), 80 (B/C), 90 (A/B)
		boundaries := []float64{60, 70, 80, 90}
		for _, boundary := range boundaries {
			if math.Abs(score-boundary) < 2.0 { // Within 2 points of boundary
				riskScore += 30
				riskFactors = append(riskFactors, fmt.Sprintf("near_boundary_%.0f", boundary))
				break
			}
		}

		// Factor 5: Very low score (< 50) adds risk
		if score < 50 {
			riskScore += 20
			riskFactors = append(riskFactors, "very_low_score")
		}

		// Only include submissions with some risk
		if riskScore > 0 {
			// Cap at 100
			if riskScore > 100 {
				riskScore = 100
			}

			risks = append(risks, models.RegradeRisk{
				SubmissionID:      submissionID,
				StudentIdentifier: studentID,
				Score:             score,
				RiskScore:         riskScore,
				RiskFactors:       riskFactors,
				GradedBy:          gradedBy,
				GraderEmail:       graderEmail,
			})
		}
	}

	// Sort by risk score (highest first)
	// Simple bubble sort for small datasets
	for i := 0; i < len(risks)-1; i++ {
		for j := 0; j < len(risks)-i-1; j++ {
			if risks[j].RiskScore < risks[j+1].RiskScore {
				risks[j], risks[j+1] = risks[j+1], risks[j]
			}
		}
	}

	return risks, nil
}

// calculateStatistics computes mean and standard deviation
func (s *AnomalyService) calculateStatistics(grades []map[string]interface{}) (float64, float64) {
	if len(grades) == 0 {
		return 0, 0
	}

	// Calculate mean
	total := 0.0
	for _, grade := range grades {
		total += grade["score"].(float64)
	}
	mean := total / float64(len(grades))

	// Calculate standard deviation
	variance := 0.0
	for _, grade := range grades {
		diff := grade["score"].(float64) - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(grades)))

	return mean, stdDev
}

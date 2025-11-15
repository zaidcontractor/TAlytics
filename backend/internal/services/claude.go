package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeModel      = "claude-3-5-sonnet-20241022"
	claudeAPIVersion = "2023-06-01"
)

// ClaudeService handles all interactions with Claude API
type ClaudeService struct {
	APIKey     string
	HTTPClient *http.Client
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

// ClaudeMessage represents a message in the Claude conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewClaudeService creates a new Claude service instance
func NewClaudeService(apiKey string) *ClaudeService {
	return &ClaudeService{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// callClaudeAPI makes a request to Claude API
func (s *ClaudeService) callClaudeAPI(systemPrompt, userPrompt string, maxTokens int) (string, error) {
	if s.APIKey == "" {
		return "", fmt.Errorf("Claude API key is not configured")
	}

	req := ClaudeRequest{
		Model:     claudeModel,
		MaxTokens: maxTokens,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	if systemPrompt != "" {
		req.System = systemPrompt
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", s.APIKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return claudeResp.Content[0].Text, nil
}

// ParseRubricFromText converts extracted PDF text into structured rubric JSON
func (s *ClaudeService) ParseRubricFromText(rubricText string) (string, error) {
	systemPrompt := `You are a rubric parsing expert. Convert the provided rubric text into a structured JSON format.

The JSON should follow this exact structure:
{
  "title": "Rubric Name",
  "max_points": 100,
  "criteria": [
    {
      "name": "Criterion Name",
      "max_points": 20,
      "description": "Description of what this criterion evaluates",
      "levels": [
        {
          "name": "Excellent",
          "points": 20,
          "description": "Full points description"
        },
        {
          "name": "Good",
          "points": 15,
          "description": "Partial points description"
        }
      ]
    }
  ]
}

IMPORTANT: Return ONLY the JSON object, with no additional text, explanation, or markdown formatting.`

	userPrompt := fmt.Sprintf("Parse this rubric text into the specified JSON format:\n\n%s", rubricText)

	response, err := s.callClaudeAPI(systemPrompt, userPrompt, 4000)
	if err != nil {
		return "", fmt.Errorf("Claude API call failed: %w", err)
	}

	// Clean up response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Validate it's valid JSON
	var jsonCheck interface{}
	if err := json.Unmarshal([]byte(response), &jsonCheck); err != nil {
		return "", fmt.Errorf("Claude returned invalid JSON: %w", err)
	}

	return response, nil
}

// GenerateGradingRecommendation provides Claude-assisted grading recommendation
func (s *ClaudeService) GenerateGradingRecommendation(submissionText string, rubricJSON string) (string, error) {
	systemPrompt := `You are a grading assistant. Apply the provided rubric strictly to the student submission and provide a detailed grading recommendation.

Return your response as a JSON object with this structure:
{
  "total_points": 85,
  "max_points": 100,
  "criteria_scores": [
    {
      "criterion_name": "Criterion 1",
      "points_awarded": 18,
      "max_points": 20,
      "justification": "Detailed explanation of why these points were awarded"
    }
  ],
  "overall_feedback": "Summary feedback for the student",
  "strengths": ["Strength 1", "Strength 2"],
  "areas_for_improvement": ["Area 1", "Area 2"]
}

IMPORTANT: Be fair, consistent, and provide detailed justifications. Return ONLY the JSON object.`

	userPrompt := fmt.Sprintf("Rubric:\n%s\n\nStudent Submission:\n%s\n\nProvide grading recommendation:", rubricJSON, submissionText)

	response, err := s.callClaudeAPI(systemPrompt, userPrompt, 4000)
	if err != nil {
		return "", fmt.Errorf("Claude API call failed: %w", err)
	}

	// Clean up response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	return response, nil
}

// AnswerRubricQuestion handles TA questions about rubric interpretation
func (s *ClaudeService) AnswerRubricQuestion(question string, rubricJSON string, courseMaterials []string) (string, error) {
	systemPrompt := `You are a helpful teaching assistant advisor. Answer questions about rubric interpretation using the rubric as the ground truth. Be clear, concise, and helpful.`

	materialsContext := ""
	if len(courseMaterials) > 0 {
		materialsContext = fmt.Sprintf("\n\nCourse Materials:\n%s", strings.Join(courseMaterials, "\n\n"))
	}

	userPrompt := fmt.Sprintf("Rubric:\n%s%s\n\nQuestion: %s", rubricJSON, materialsContext, question)

	response, err := s.callClaudeAPI(systemPrompt, userPrompt, 2000)
	if err != nil {
		return "", fmt.Errorf("Claude API call failed: %w", err)
	}

	return response, nil
}

// GenerateAnomalyInsights provides AI analysis of detected anomalies
func (s *ClaudeService) GenerateAnomalyInsights(anomalyData map[string]interface{}) (string, error) {
	systemPrompt := `You are an educational data analyst. Analyze the provided anomaly detection results and provide actionable insights for professors and head TAs.

Focus on:
1. What the anomalies indicate about grading consistency
2. Whether rubric refinement is needed
3. Whether TA training/guidance is needed
4. Specific recommendations for improvement

Provide clear, actionable recommendations.`

	anomalyJSON, err := json.MarshalIndent(anomalyData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal anomaly data: %w", err)
	}

	userPrompt := fmt.Sprintf("Analyze these grading anomalies and provide insights:\n\n%s", string(anomalyJSON))

	response, err := s.callClaudeAPI(systemPrompt, userPrompt, 3000)
	if err != nil {
		return "", fmt.Errorf("Claude API call failed: %w", err)
	}

	return response, nil
}

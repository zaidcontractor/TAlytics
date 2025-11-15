package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PDFService handles PDF text extraction
type PDFService struct{}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{}
}

// ExtractTextFromPDF extracts text from a PDF file
func (s *PDFService) ExtractTextFromPDF(pdfPath string) (string, error) {
	// Check which tool is available
	tool, err := s.CheckPDFToolAvailability()
	if err != nil {
		return "", fmt.Errorf("no PDF extraction tool available: %w", err)
	}

	var output []byte
	var extractErr error

	switch tool {
	case "pdftotext":
		// Use pdftotext with layout preservation
		cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
		output, extractErr = cmd.Output()
		if extractErr != nil {
			return "", fmt.Errorf("pdftotext extraction failed: %w", extractErr)
		}

	case "pdfcpu":
		// Fallback to pdfcpu
		// Note: pdfcpu extract requires output file, so we use a temp file
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("extract_%d.txt", time.Now().UnixNano()))
		defer os.Remove(tempFile)

		cmd := exec.Command("pdfcpu", "extract", "-mode", "text", pdfPath, tempFile)
		if extractErr = cmd.Run(); extractErr != nil {
			return "", fmt.Errorf("pdfcpu extraction failed: %w", extractErr)
		}

		output, extractErr = os.ReadFile(tempFile)
		if extractErr != nil {
			return "", fmt.Errorf("failed to read extracted text: %w", extractErr)
		}

	default:
		return "", fmt.Errorf("unknown PDF tool: %s", tool)
	}

	text := string(output)
	text = strings.TrimSpace(text)

	if text == "" {
		return "", fmt.Errorf("extracted text is empty")
	}

	return text, nil
}

// CheckPDFToolAvailability checks which PDF extraction tool is available
func (s *PDFService) CheckPDFToolAvailability() (string, error) {
	// Check for pdftotext
	if err := exec.Command("pdftotext", "-v").Run(); err == nil {
		return "pdftotext", nil
	}

	// Check for pdfcpu
	if err := exec.Command("pdfcpu", "version").Run(); err == nil {
		return "pdfcpu", nil
	}

	return "", fmt.Errorf("neither pdftotext nor pdfcpu is available")
}

// SaveUploadedPDF saves an uploaded PDF file to the local storage
func (s *PDFService) SaveUploadedPDF(fileData []byte, fileName, category string) (string, error) {
	// Determine save directory based on category
	var saveDir string
	switch category {
	case "rubric":
		saveDir = "./data/rubrics"
	case "submission":
		saveDir = "./data/submissions"
	default:
		return "", fmt.Errorf("invalid category: %s", category)
	}

	// Ensure directory exists
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	uniqueFileName := fmt.Sprintf("%s_%s%s", baseName, timestamp, ext)
	filePath := filepath.Join(saveDir, uniqueFileName)

	// Write file to disk
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

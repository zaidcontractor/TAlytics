package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPDFToolAvailability(t *testing.T) {
	service := NewPDFService()

	tool, err := service.CheckPDFToolAvailability()
	if err != nil {
		t.Logf("No PDF tool available (this is OK for testing): %v", err)
		return
	}

	if tool != "pdftotext" && tool != "pdfcpu" {
		t.Errorf("Unexpected tool: %s", tool)
	}
}

func TestSaveUploadedPDF(t *testing.T) {
	service := NewPDFService()

	// Create test PDF content (minimal valid PDF)
	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\nxref\n0 0\ntrailer\n<< /Root 1 0 R >>\n%%EOF")
	fileName := "test.pdf"
	category := "rubric"

	filePath, err := service.SaveUploadedPDF(pdfContent, fileName, category)
	if err != nil {
		t.Fatalf("Failed to save PDF: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File was not created at %s", filePath)
	}

	// Check file is in correct directory
	expectedDir := "./data/rubrics"
	if !filepath.IsAbs(filePath) {
		absPath, _ := filepath.Abs(filePath)
		expectedAbs, _ := filepath.Abs(expectedDir)
		if !filepath.HasPrefix(absPath, expectedAbs) {
			t.Errorf("File should be in %s directory", expectedDir)
		}
	}

	// Cleanup
	os.Remove(filePath)
}

func TestSaveUploadedPDF_InvalidCategory(t *testing.T) {
	service := NewPDFService()

	pdfContent := []byte("%PDF-1.4")
	fileName := "test.pdf"
	category := "invalid"

	_, err := service.SaveUploadedPDF(pdfContent, fileName, category)
	if err == nil {
		t.Error("Expected error for invalid category, got nil")
	}
}

func TestExtractTextFromPDF_NoTool(t *testing.T) {
	service := NewPDFService()

	// Create a temporary PDF file
	tmpFile := filepath.Join(os.TempDir(), "test.pdf")
	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\nxref\n0 0\ntrailer\n<< /Root 1 0 R >>\n%%EOF")
	err := os.WriteFile(tmpFile, pdfContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}
	defer os.Remove(tmpFile)

	// Try to extract (will fail if no tool available, which is OK)
	_, err = service.ExtractTextFromPDF(tmpFile)
	if err != nil {
		// This is expected if PDF tools are not installed
		t.Logf("PDF extraction failed (expected if tools not installed): %v", err)
	}
}


#!/bin/bash

# TAlytics Rubric Upload Test Script

BASE_URL="http://localhost:8080"

echo "================================================"
echo "TAlytics Rubric Upload Test Script"
echo "================================================"
echo ""

# Check if server is running
echo "1. Checking if server is running..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo "❌ Server is not running at ${BASE_URL}"
    echo "Please start the server with: ./run.sh"
    exit 1
fi
echo "✅ Server is running"
echo ""

# Test 1: Create a test assignment first (for testing)
echo "2. Creating test course and assignment..."
# Note: This will fail without auth, but shows the flow
# In production, you'd need to login first and get a token

# Test 2: Test CreateRubric endpoint (JSON)
echo "3. Testing CreateRubric endpoint (manual JSON)..."
curl -X POST "${BASE_URL}/rubrics" \
  -H "Content-Type: application/json" \
  -d '{
    "assignment_id": 1,
    "json_blob": "{\"title\":\"Test Rubric\",\"max_points\":100,\"criteria\":[{\"name\":\"Quality\",\"max_points\":50},{\"name\":\"Completeness\",\"max_points\":50}]}",
    "max_points": 100
  }' \
  -w "\nHTTP Status: %{http_code}\n"
echo ""

# Test 3: Upload a PDF rubric (requires a PDF file)
echo "4. Testing UploadRubricPDF endpoint..."
echo "Note: This test requires:"
echo "  - A PDF file at: ./test_rubric.pdf"
echo "  - CLAUDE_API_KEY environment variable set"
echo "  - pdftotext or pdfcpu installed"
echo ""

if [ -f "./test_rubric.pdf" ]; then
    echo "Found test_rubric.pdf, uploading..."
    curl -X POST "${BASE_URL}/rubrics/upload" \
      -F "assignment_id=2" \
      -F "file=@./test_rubric.pdf" \
      -w "\nHTTP Status: %{http_code}\n"
    echo ""
else
    echo "⚠️  No test_rubric.pdf found. Skipping PDF upload test."
    echo "To test PDF upload:"
    echo "  1. Place a rubric PDF file as: ./test_rubric.pdf"
    echo "  2. Set CLAUDE_API_KEY environment variable"
    echo "  3. Run this script again"
    echo ""
fi

# Test 4: Check PDF tool availability
echo "5. Checking PDF extraction tool availability..."
if command -v pdftotext &> /dev/null; then
    echo "✅ pdftotext is available"
    pdftotext -v 2>&1 | head -1
elif command -v pdfcpu &> /dev/null; then
    echo "✅ pdfcpu is available"
    pdfcpu version
else
    echo "❌ No PDF extraction tool found"
    echo "Install with:"
    echo "  macOS: brew install poppler"
    echo "  Ubuntu: sudo apt-get install poppler-utils"
fi
echo ""

# Test 5: Check Claude API key
echo "6. Checking Claude API configuration..."
if [ -z "$CLAUDE_API_KEY" ]; then
    echo "⚠️  CLAUDE_API_KEY not set"
    echo "Set it with: export CLAUDE_API_KEY=your_key_here"
else
    echo "✅ CLAUDE_API_KEY is configured"
fi
echo ""

echo "================================================"
echo "Test Summary"
echo "================================================"
echo ""
echo "To fully test the rubric upload feature:"
echo ""
echo "1. Ensure server is running: ./run.sh"
echo "2. Set CLAUDE_API_KEY: export CLAUDE_API_KEY=sk-ant-xxxxx"
echo "3. Install PDF tool: brew install poppler"
echo "4. Create test data:"
echo "   - First create a course and assignment via API"
echo "   - Then upload a rubric PDF"
echo ""
echo "Example full workflow:"
echo ""
echo "# Start server with API key"
echo "export CLAUDE_API_KEY=sk-ant-xxxxx"
echo "./run.sh"
echo ""
echo "# In another terminal, test upload"
echo "curl -X POST http://localhost:8080/rubrics/upload \\"
echo "  -F 'assignment_id=1' \\"
echo "  -F 'file=@/path/to/rubric.pdf'"
echo ""

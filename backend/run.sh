#!/bin/bash

# TAlytics Backend Quick Start Script

echo "🚀 Starting TAlytics Backend..."
echo ""

# Set default environment variables
export PORT="${PORT:-8080}"
export DB_PATH="${DB_PATH:-../data/talytics.db}"

# Build the application
echo "📦 Building application..."
go build -o bin/talytics cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"
echo ""
echo "🌐 Starting server on port $PORT..."
echo "📊 Database: $DB_PATH"
echo "🏥 Health check: http://localhost:$PORT/health"
echo ""
echo "Press Ctrl+C to stop the server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Run the server
./bin/talytics

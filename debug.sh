#!/bin/bash

echo "🔧 Debug Mode - Stock Panel Server"
echo "=================================="

# Kill any existing process on port 8080
echo "🔄 Killing existing processes on port 8080..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No existing processes found"

# Wait a moment for port to be freed
sleep 1

echo "🚀 Starting server with debugging..."
echo "📝 Server will restart automatically on file changes"
echo "🌐 Access the application at: http://localhost:8080"
echo ""

# Start the server
go run cmd/stock-panel/main.go

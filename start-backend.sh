#!/bin/bash
# Altivion Backend Startup Script

echo "Starting Altivion Backend API..."
echo "================================"

# Check if .env exists
if [ ! -f "api/.env" ]; then
    echo "ERROR: api/.env not found!"
    echo "Please copy api/.env.example to api/.env and configure it."
    exit 1
fi

# Navigate to api directory
cd api || exit 1

# Check if virtual environment exists
if [ -d "venv" ]; then
    echo "Activating virtual environment..."
    source venv/bin/activate
else
    echo "WARNING: Virtual environment not found."
    echo "Consider creating one with: python -m venv venv"
fi

# Check if dependencies are installed
if ! python -c "import fastapi" 2>/dev/null; then
    echo "Installing dependencies..."
    pip install -r requirements.txt
fi

echo ""
echo "Starting FastAPI server on http://127.0.0.1:8000"
echo "Press CTRL+C to stop"
echo ""

# Start the server
uvicorn main:app --host 127.0.0.1 --port 8000 --reload

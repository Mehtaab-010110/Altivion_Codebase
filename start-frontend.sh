#!/bin/bash
# Altivion Frontend Startup Script

echo "Starting Altivion Frontend..."
echo "============================="

# Check if .env.local exists
if [ ! -f "altivion-web/.env.local" ]; then
    echo "WARNING: altivion-web/.env.local not found!"
    echo "Please copy altivion-web/.env.local.example to altivion-web/.env.local"
    echo "and configure your Google Maps API key."
    echo ""
    echo "Continuing anyway (map may not work without API key)..."
    echo ""
fi

# Navigate to frontend directory
cd altivion-web || exit 1

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

echo ""
echo "Starting Next.js development server on http://localhost:3000"
echo "Press CTRL+C to stop"
echo ""

# Start the development server
npm run dev

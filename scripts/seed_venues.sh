#!/bin/bash

# Venue Collection Seeding Script Wrapper
# This script activates the Python virtual environment and runs the venue seeding script

# Set variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_DIR="${SCRIPT_DIR}/python"
VENV_DIR="${SCRIPT_DIR}/../scraper-env"
SEED_SCRIPT="${PYTHON_DIR}/seed_venues.py"

# Check if virtual environment exists
if [ ! -d "${VENV_DIR}" ]; then
    echo "❌ Python virtual environment not found at: ${VENV_DIR}"
    echo "   Please run the project setup first to create the Python environment."
    exit 1
fi

# Check if the seed script exists
if [ ! -f "${SEED_SCRIPT}" ]; then
    echo "❌ Seeding script not found at: ${SEED_SCRIPT}"
    exit 1
fi

# Make the Python script executable
chmod +x "${SEED_SCRIPT}"

# Activate the virtual environment and run the script
echo "🔄 Activating Python virtual environment..."
source "${VENV_DIR}/bin/activate"

echo "🚀 Running venue seeding script..."
python "${SEED_SCRIPT}"

# Deactivate the virtual environment
deactivate

echo "✅ Venue seeding completed" 
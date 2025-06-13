.PHONY: setup run test clean

VENV_DIR = venv
PYTHON = $(VENV_DIR)/bin/python
PIP = $(VENV_DIR)/bin/pip

# Default target
all: setup

# Set up virtual environment and install dependencies
setup:
	@echo "Setting up virtual environment..."
	python3 -m venv $(VENV_DIR)
	$(PIP) install --upgrade pip
	$(PIP) install -r requirements.txt
	$(PYTHON) -m playwright install

# Run the scraper
run:
	@echo "Running scraper..."
	$(PYTHON) src/scrapers/scraper_orchestrator.py

# Run tests
test:
	@echo "Running tests..."
	$(PYTHON) -m pytest tests/

# Clean up
clean:
	@echo "Cleaning up..."
	rm -rf $(VENV_DIR)
	rm -rf __pycache__
	rm -rf src/__pycache__
	rm -rf src/scrapers/__pycache__
	rm -rf .pytest_cache 
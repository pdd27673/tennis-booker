# Python scraping service with Playwright
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies for Playwright (as root)
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Copy Python package files
COPY pyproject.toml .
COPY MANIFEST.in .
COPY src/ ./src/

# Install Python dependencies
RUN pip install --no-cache-dir -e .

# Install Playwright browsers with system dependencies (as root)
RUN playwright install --with-deps chromium

# Source code already copied above

# Create logs directory
RUN mkdir -p logs

# Add non-root user and change ownership
RUN useradd -m -u 1001 appuser && chown -R appuser:appuser /app

# Create playwright cache directory for appuser and copy browsers
RUN mkdir -p /home/appuser/.cache && \
    cp -r /root/.cache/ms-playwright /home/appuser/.cache/ && \
    chown -R appuser:appuser /home/appuser/.cache

USER appuser

# Set playwright environment variables
ENV PLAYWRIGHT_BROWSERS_PATH=/home/appuser/.cache/ms-playwright

# Default command - can be overridden
CMD ["python", "src/playwright_scraper.py", "--all"] 
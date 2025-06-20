# Python scraping service with Playwright
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies for Playwright
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Copy Python requirements
COPY apps/scraper/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Install Playwright and browsers
RUN playwright install --with-deps chromium

# Copy source code
COPY apps/scraper/src/ ./src/

# Copy config files from project root
COPY config/ ./config/

# Create logs directory
RUN mkdir -p logs

# Add non-root user
RUN useradd -m -u 1001 appuser && chown -R appuser:appuser /app
USER appuser

# Default command - can be overridden
CMD ["python", "src/playwright_scraper.py", "--all"] 
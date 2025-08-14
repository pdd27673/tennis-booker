# Multi-stage build for React frontend with Vite optimized for Railway deployment
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including dev dependencies for build)
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage with Nginx
FROM nginx:1.25-alpine AS production

# Install curl for health checks
RUN apk add --no-cache curl

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built application from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy PWA offline page
COPY --from=builder /app/public/offline.html /usr/share/nginx/html/

# Create nginx user and set permissions
RUN addgroup -g 1001 -S nginx && \
    adduser -S nginx -u 1001 -G nginx && \
    chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d

# Railway specific: Create a directory for runtime environment variable substitution
RUN mkdir -p /app/env-config
COPY --from=builder /app/public/env-config.js /app/env-config/env-config.js.template

# Script to replace environment variables at runtime
RUN echo '#!/bin/sh\n\
# Generate env-config.js with runtime environment variables\n\
envsubst < /app/env-config/env-config.js.template > /usr/share/nginx/html/env-config.js\n\
\n\
# Start nginx\n\
exec nginx -g "daemon off;"\n\
' > /docker-entrypoint.sh && \
    chmod +x /docker-entrypoint.sh

# Switch to non-root user
USER nginx

# Railway uses PORT environment variable - expose both 80 and $PORT
EXPOSE 80 $PORT

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:80/health || exit 1

# Use custom entrypoint script
CMD ["/docker-entrypoint.sh"]

#!/bin/bash

# Setup Vault secrets for tennis booker application - PRODUCTION TEMPLATE
# Copy this file and replace all REPLACE_WITH_* values with real production credentials
# NEVER commit the real production script to git!

set -e

VAULT_ADDR=${VAULT_ADDR:-"https://your-production-vault.com"}
VAULT_TOKEN=${VAULT_TOKEN:-"REPLACE_WITH_PRODUCTION_TOKEN"}

echo "🔐 Setting up Vault secrets for Tennis Booker"
echo "Vault Address: $VAULT_ADDR"

# Enable KV secrets engine if not already enabled
echo "📝 Enabling KV secrets engine..."
vault secrets enable -path=kv kv-v2 2>/dev/null || echo "KV engine already enabled"

# JWT Secret
echo "🔑 Setting up JWT secret..."
vault kv put kv/tennisapp/prod/jwt \
    secret="REPLACE_WITH_STRONG_PRODUCTION_JWT_SECRET_AT_LEAST_32_CHARS"

# Database credentials
echo "🗄️ Setting up database credentials..."
vault kv put kv/tennisapp/prod/db \
    username="admin" \
    password="password" \
    host="localhost:27017" \
    database="tennis_booker"

# Email credentials
echo "📧 Setting up email credentials..."
vault kv put kv/tennisapp/prod/email \
    email="REPLACE_WITH_PRODUCTION_EMAIL" \
    password="REPLACE_WITH_PRODUCTION_APP_PASSWORD" \
    smtp_host="smtp.gmail.com" \
    smtp_port="587"

# Redis credentials
echo "🔴 Setting up Redis credentials..."
vault kv put kv/tennisapp/prod/redis \
    host="localhost:6379" \
    password="password"

# API secrets
echo "🔌 Setting up API secrets..."
vault kv put kv/tennisapp/prod/api \
    firecrawl_key="fc-test-key" \
    google_api_key="google-test-key"

# Platform credentials
echo "🎾 Setting up platform credentials..."
vault kv put kv/tennisapp/prod/platforms/lta \
    username="lta-user" \
    password="lta-pass" \
    api_key="lta-api-key"

vault kv put kv/tennisapp/prod/platforms/courtsides \
    username="courtsides-user" \
    password="courtsides-pass" \
    api_key="courtsides-api-key"

# Notification services
echo "📱 Setting up notification service credentials..."
vault kv put kv/tennisapp/prod/notifications/twilio \
    account_sid="twilio-account-sid" \
    auth_token="twilio-auth-token" \
    phone_number="+1234567890"

vault kv put kv/tennisapp/prod/notifications/sendgrid \
    api_key="sendgrid-api-key" \
    from_email="noreply@tennisbooker.com"

echo "✅ All secrets have been set up successfully!"
echo ""
echo "🧪 You can now test the authentication system:"
echo "  cd apps/backend"
echo "  go run cmd/test-auth-server/main.go"
echo ""
echo "📋 Test commands:"
echo "  # Health check"
echo "  curl http://localhost:8080/health"
echo ""
echo "  # Login"
echo "  curl -X POST http://localhost:8080/login -H 'Content-Type: application/json' -d '{\"username\":\"testuser\",\"password\":\"testpass\"}'"
echo ""
echo "  # Access protected route (use token from login response)"
echo "  curl -H 'Authorization: Bearer <your-token>' http://localhost:8080/protected" 
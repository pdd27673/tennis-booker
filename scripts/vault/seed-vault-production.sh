#!/bin/bash

# Production Vault Seeding Script
# This script seeds Vault with production credentials
# NEVER commit this file with real production credentials
# Copy this template and fill in your production values

set -e

# Production Vault configuration
VAULT_ADDR=${VAULT_ADDR:-"https://your-production-vault.com"}
VAULT_TOKEN=${VAULT_TOKEN:-""}

if [ -z "$VAULT_TOKEN" ]; then
    echo "❌ VAULT_TOKEN environment variable is required for production"
    echo "Set it with: export VAULT_TOKEN=your-production-token"
    exit 1
fi

echo "🔐 Seeding Vault with Production Credentials"
echo "Vault Address: $VAULT_ADDR"

# Enable KV secrets engine if not already enabled
echo "📝 Enabling KV secrets engine..."
vault secrets enable -path=kv kv-v2 2>/dev/null || echo "KV engine already enabled"

# Email credentials for notifications
echo "📧 Setting up email credentials..."
vault kv put kv/tennisapp/prod/email \
    email="REPLACE_WITH_PRODUCTION_EMAIL" \
    password="REPLACE_WITH_PRODUCTION_APP_PASSWORD" \
    smtp_host="smtp.gmail.com" \
    smtp_port="587"

# JWT Secret - MUST be different from development
echo "🔑 Setting up JWT secret..."
vault kv put kv/tennisapp/prod/jwt \
    secret="REPLACE_WITH_STRONG_PRODUCTION_JWT_SECRET_AT_LEAST_32_CHARS"

# Database credentials
echo "🗄️ Setting up database credentials..."
vault kv put kv/tennisapp/prod/db \
    host="REPLACE_WITH_PRODUCTION_DB_HOST" \
    port="27017" \
    username="REPLACE_WITH_PRODUCTION_DB_USERNAME" \
    password="REPLACE_WITH_PRODUCTION_DB_PASSWORD" \
    database="tennis_booking"

# Redis credentials
echo "🔴 Setting up Redis credentials..."
vault kv put kv/tennisapp/prod/redis \
    host="REPLACE_WITH_PRODUCTION_REDIS_HOST:6379" \
    password="REPLACE_WITH_PRODUCTION_REDIS_PASSWORD"

# API credentials for external services
echo "🌐 Setting up API credentials..."
vault kv put kv/tennisapp/prod/api \
    playtomic_api_key="REPLACE_WITH_PRODUCTION_PLAYTOMIC_API_KEY" \
    clubspark_api_key="REPLACE_WITH_PRODUCTION_CLUBSPARK_API_KEY" \
    better_api_key="REPLACE_WITH_PRODUCTION_BETTER_API_KEY"

# Platform credentials
echo "🏢 Setting up platform credentials..."
vault kv put kv/tennisapp/prod/platform \
    playtomic_username="REPLACE_WITH_PRODUCTION_PLAYTOMIC_USERNAME" \
    playtomic_password="REPLACE_WITH_PRODUCTION_PLAYTOMIC_PASSWORD" \
    clubspark_username="REPLACE_WITH_PRODUCTION_CLUBSPARK_USERNAME" \
    clubspark_password="REPLACE_WITH_PRODUCTION_CLUBSPARK_PASSWORD"

# Notification service credentials
echo "🔔 Setting up notification credentials..."
vault kv put kv/tennisapp/prod/notifications \
    webhook_secret="REPLACE_WITH_PRODUCTION_WEBHOOK_SECRET" \
    slack_token="REPLACE_WITH_PRODUCTION_SLACK_TOKEN" \
    discord_webhook="REPLACE_WITH_PRODUCTION_DISCORD_WEBHOOK"

echo "✅ Production Vault seeding completed!"
echo ""
echo "📋 Verify secrets with:"
echo "  vault kv get kv/tennisapp/prod/email"
echo "  vault kv get kv/tennisapp/prod/jwt"
echo "  vault kv get kv/tennisapp/prod/db"
echo ""
echo "⚠️  IMPORTANT: Delete this script after use or store it securely!"
echo "⚠️  Never commit production credentials to git!" 
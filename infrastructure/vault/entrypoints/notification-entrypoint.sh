#!/bin/sh
set -e

SERVICE_NAME="notification"
SECRETS_FILE="/vault/secrets/${SERVICE_NAME}.env"
CONFIG_FILE="/vault/secrets/${SERVICE_NAME}-config.json"

echo "🔄 Waiting for Vault Agent to generate secrets for $SERVICE_NAME..."

# Wait for secrets file to be generated
while [ ! -f "$SECRETS_FILE" ]; do
    echo "⏳ Waiting for $SECRETS_FILE..."
    sleep 2
done

# Wait for config file to be generated
while [ ! -f "$CONFIG_FILE" ]; do
    echo "⏳ Waiting for $CONFIG_FILE..."
    sleep 2
done

echo "✅ Secrets files found, loading environment variables..."

# Load environment variables
set -a
source "$SECRETS_FILE"
set +a

echo "🚀 Starting $SERVICE_NAME service..."
exec npm start 
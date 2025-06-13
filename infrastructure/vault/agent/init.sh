#!/bin/sh
set -e

echo "Setting up Vault Agent environment..."

# Fix ownership of secrets directory for vault user
chown vault:vault /vault/secrets
chmod 755 /vault/secrets

# Start Vault Agent
echo "Starting Vault Agent..."
exec vault agent -config=/vault/config/backend-agent.hcl 
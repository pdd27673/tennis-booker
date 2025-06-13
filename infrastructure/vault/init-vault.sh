#!/bin/bash
set -e

# Check if VAULT_ADDR is set, otherwise use default
VAULT_ADDR=${VAULT_ADDR:-http://127.0.0.1:8200}
VAULT_TOKEN=${VAULT_TOKEN:-dev-token}

echo "Initializing Vault at $VAULT_ADDR with token $VAULT_TOKEN"

# Wait for Vault to be ready
until curl -fs $VAULT_ADDR/v1/sys/health > /dev/null; do
  echo "Waiting for Vault to start..."
  sleep 1
done

# Login with the root token
vault login -address=$VAULT_ADDR $VAULT_TOKEN

# Check if KV v2 secrets engine is already enabled
if vault secrets list -address=$VAULT_ADDR | grep -q "^kv/"; then
  echo "KV secrets engine is already enabled at 'kv/' path"
else
  # Enable KV v2 secrets engine
  echo "Enabling KV v2 secrets engine at 'kv/' path"
  vault secrets enable -address=$VAULT_ADDR -version=2 -path=kv kv
fi

# Verify KV v2 is properly configured
echo "Verifying KV v2 configuration"
vault secrets list -address=$VAULT_ADDR -detailed | grep kv

# Create secret paths
echo "Creating secret paths for critical services"

# Database secrets
echo "Creating database secrets"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/db \
  username="admin" \
  password="replace_with_secure_password" \
  host="mongodb" \
  port="27017" \
  dbname="tennis_booking"

# JWT secrets
echo "Creating JWT secrets"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/jwt \
  secret_key="replace_with_secure_jwt_key" \
  expiration="24h"

# Email secrets
echo "Creating email secrets"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/email \
  smtp_host="smtp.gmail.com" \
  smtp_port="587" \
  username="replace_with_email@gmail.com" \
  password="replace_with_app_password"

# Redis secrets
echo "Creating Redis secrets"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/redis \
  host="redis" \
  port="6379" \
  password="replace_with_secure_redis_password"

# API secrets
echo "Creating API secrets"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/api \
  api_key="replace_with_secure_api_key" \
  rate_limit="100"

# LTA/Clubspark credentials
echo "Creating LTA/Clubspark credentials"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/platforms/lta \
  username="replace_with_lta_username" \
  password="replace_with_lta_password"

# Courtsides/Tennis Tower Hamlets credentials
echo "Creating Courtsides/Tennis Tower Hamlets credentials"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/platforms/courtsides \
  username="replace_with_courtsides_username" \
  password="replace_with_courtsides_password"

# Twilio (SMS) credentials
echo "Creating Twilio credentials"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/notifications/twilio \
  account_sid="replace_with_twilio_account_sid" \
  auth_token="replace_with_twilio_auth_token" \
  from_number="replace_with_twilio_from_number"

# SendGrid (Email) credentials
echo "Creating SendGrid credentials"
vault kv put -address=$VAULT_ADDR kv/tennisapp/prod/notifications/sendgrid \
  api_key="replace_with_sendgrid_api_key" \
  from_email="replace_with_sendgrid_from_email"

# Create policies from files
echo "Creating access policies from policy files"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POLICIES_DIR="$SCRIPT_DIR/policies"

# Write the policies to Vault using the policy files
vault policy write -address=$VAULT_ADDR db-policy "$POLICIES_DIR/db-policy.hcl"
vault policy write -address=$VAULT_ADDR jwt-policy "$POLICIES_DIR/jwt-policy.hcl"
vault policy write -address=$VAULT_ADDR email-policy "$POLICIES_DIR/email-policy.hcl"
vault policy write -address=$VAULT_ADDR redis-policy "$POLICIES_DIR/redis-policy.hcl"
vault policy write -address=$VAULT_ADDR api-policy "$POLICIES_DIR/api-policy.hcl"
vault policy write -address=$VAULT_ADDR platforms-policy "$POLICIES_DIR/platforms-policy.hcl"
vault policy write -address=$VAULT_ADDR notifications-policy "$POLICIES_DIR/notifications-policy.hcl"
vault policy write -address=$VAULT_ADDR admin-policy "$POLICIES_DIR/admin-policy.hcl"

echo "Vault initialization complete!"
echo "Use the following commands to read secrets:"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/db"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/jwt"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/email"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/redis"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/api"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/platforms/lta"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/platforms/courtsides"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/notifications/twilio"
echo "vault kv get -address=$VAULT_ADDR kv/tennisapp/prod/notifications/sendgrid" 
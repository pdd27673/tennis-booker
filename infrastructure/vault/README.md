# Vault Infrastructure

This directory contains HashiCorp Vault configuration and setup scripts for the Tennis Booker application.

## Quick Start

### Development/Testing Setup
```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-token
./infrastructure/vault/setup-vault-secrets.sh
```

### Production Setup
1. Copy the template: `cp setup-vault-secrets-template.sh setup-vault-secrets-production.sh`
2. Replace all `REPLACE_WITH_*` placeholders with real production values
3. **Never commit the production script to git** (it's git-ignored)
4. Run: `./setup-vault-secrets-production.sh`

## Directory Structure

```
infrastructure/vault/
├── setup-vault-secrets.sh          # Real credentials (git-ignored)
├── setup-vault-secrets-template.sh # Safe template for production
├── scripts-README.md               # Detailed script documentation
├── policies/                       # Vault access policies
│   └── tennis-app-policies.hcl
└── config/                         # Vault server configuration
```

## Security Notes

- The real `setup-vault-secrets.sh` contains actual credentials and is git-ignored
- Only the template `setup-vault-secrets-template.sh` is committed to git
- Always use strong, unique credentials in production
- Use proper Vault authentication (not dev tokens) in production

## Secret Paths

The setup script creates the following secret paths:

- `kv/tennisapp/prod/jwt` - JWT signing secrets
- `kv/tennisapp/prod/db` - Database credentials
- `kv/tennisapp/prod/email` - Email service credentials
- `kv/tennisapp/prod/redis` - Redis connection details
- `kv/tennisapp/prod/api` - External API keys (Firecrawl, Google, etc.)
- `kv/tennisapp/prod/platforms/lta` - LTA/Clubspark credentials
- `kv/tennisapp/prod/platforms/courtsides` - Courtsides credentials
- `kv/tennisapp/prod/notifications/twilio` - Twilio SMS credentials
- `kv/tennisapp/prod/notifications/sendgrid` - SendGrid email credentials

## Testing

After running the setup script, you can test the authentication system:

```bash
cd apps/backend
go run cmd/test-auth-server/main.go

# In another terminal:
curl http://localhost:8080/health
curl -X POST http://localhost:8080/login -H 'Content-Type: application/json' -d '{"username":"testuser","password":"testpass"}'
```

## Policies

Vault access policies are defined in `policies/tennis-app-policies.hcl` for production use with proper authentication and authorization. 
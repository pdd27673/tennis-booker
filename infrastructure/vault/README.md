# 🔐 Vault Infrastructure & Integration

This directory contains HashiCorp Vault configuration, setup scripts, and Vault Agent integration for the Tennis Booker application.

## 🏗️ Architecture Overview

The integration uses a **DRY (Don't Repeat Yourself)** approach with:

- **Universal Vault Agent Configuration**: Reusable templates for all services
- **Service-Specific Templates**: Customized secret injection per service
- **Consolidated Docker Compose**: Single configuration for all Vault Agents
- **Shared Volume Strategy**: Efficient secret sharing between containers

## 🚀 Quick Start

### Development/Testing Setup
```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=your-vault-token-here
./infrastructure/vault/setup-vault-secrets.sh
```

### Production Setup
1. Copy the template: `cp setup-vault-secrets-template.sh setup-vault-secrets-production.sh`
2. Replace all `REPLACE_WITH_*` placeholders with real production values
3. **Never commit the production script to git** (it's git-ignored)
4. Run: `./setup-vault-secrets-production.sh`

### Vault Agent Integration
```bash
# Start everything with Vault integration
docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml up -d

# Or use the consolidated approach
make vault-up    # Start all services
make vault-status # Check status
make vault-logs   # View logs
make vault-test   # Test integration
```

## 📁 Directory Structure

```
infrastructure/vault/
├── agent/
│   ├── universal-init.sh          # Universal Vault Agent init script
│   ├── generate-config.sh         # Config generation utility
│   └── token                      # Vault token file
├── templates/
│   ├── backend-env.tpl            # Backend environment variables
│   ├── backend-config.json.tpl    # Backend JSON configuration
│   ├── scraper-env.tpl            # Scraper environment variables
│   ├── scraper-config.json.tpl    # Scraper JSON configuration
│   ├── notification-env.tpl       # Notification environment variables
│   └── notification-config.json.tpl # Notification JSON configuration
├── entrypoints/
│   ├── backend-entrypoint.sh      # Backend service entrypoint
│   ├── scraper-entrypoint.sh      # Scraper service entrypoint
│   └── notification-entrypoint.sh # Notification service entrypoint
├── setup-vault-secrets.sh          # Real credentials (git-ignored)
├── setup-vault-secrets-template.sh # Safe template for production
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

## 🔧 How Vault Agent Integration Works

### 1. Universal Vault Agent Pattern

Each service gets its own Vault Agent container that:
- Uses the same base configuration (`universal-init.sh`)
- Generates service-specific config based on `SERVICE_NAME` environment variable
- Renders secrets using service-specific templates
- Shares secrets via Docker volumes

### 2. Template System

Templates use Vault's template syntax to fetch secrets:

```hcl
{{- with secret "kv/data/tennisapp/prod/database" }}
MONGO_URI="mongodb://{{ .Data.data.username }}:{{ .Data.data.password }}@mongodb:27017/{{ .Data.data.database }}?authSource=admin"
{{- end }}
```

### 3. Service Integration

Each service:
- Waits for secrets to be generated
- Sources environment variables from generated files
- Starts the application with injected secrets

## 🛡️ Security Features

- **No Hardcoded Secrets**: All secrets managed by Vault
- **Non-Root Containers**: All application containers run as non-root users
- **File-Based Injection**: Secrets written to files with proper permissions
- **Automatic Rotation**: Vault Agent handles secret renewal
- **Least Privilege**: Each service only accesses required secrets

## 📋 Service-Specific Secrets

### Backend Service
- Database credentials (MongoDB)
- Redis credentials
- JWT secret
- Email credentials (Gmail)

### Scraper Service
- Database credentials (MongoDB)
- Redis credentials

### Notification Service
- Database credentials (MongoDB)
- Redis credentials
- JWT secret
- Email credentials (Gmail)

## 🚨 Troubleshooting

### Common Issues

1. **Vault Agent Not Starting**
   - Check Vault server is running: `docker logs tennis-vault`
   - Verify token file exists: `ls -la infrastructure/vault/agent/token`

2. **Secrets Not Generated**
   - Check Vault Agent logs for authentication errors
   - Verify templates exist and are valid
   - Ensure Vault policies allow secret access

3. **Service Won't Start**
   - Check if secrets files exist in `/vault/secrets/`
   - Verify entrypoint script permissions
   - Check service logs for specific errors

### Debug Commands

```bash
# Check all container status
make vault-status

# View all logs
make vault-logs

# Test secret generation
make vault-test

# View generated secrets (for debugging)
docker exec tennis-vault-agent-backend cat /vault/secrets/backend.env
```

## Policies

Vault access policies are defined in `policies/tennis-app-policies.hcl` for production use with proper authentication and authorization. 
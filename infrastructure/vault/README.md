# Vault Setup and Configuration

This document describes the HashiCorp Vault setup for the Tennis Booking application.

## Development Setup

For local development, Vault is configured to run in development mode within Docker:

```bash
# Start the development environment
docker-compose up -d

# Vault will be available at http://localhost:8200
# Default root token: dev-token
```

## Production Setup

For production, Vault is configured with file storage and proper policies:

```bash
# Start the production environment
docker-compose -f docker-compose.prod.yml up -d

# Initialize Vault (first time only)
# This creates unseal keys and a root token
docker exec -it tennis-vault vault operator init

# Unseal Vault (required after each restart)
# You'll need to run this 3 times with different unseal keys
docker exec -it tennis-vault vault operator unseal <unseal-key>

# Login with root token
docker exec -it tennis-vault vault login <root-token>
```

## Secret Paths

The following secret paths are defined:

### Core Application Secrets
- `kv/tennisapp/prod/db`: Database credentials
- `kv/tennisapp/prod/jwt`: JWT signing keys
- `kv/tennisapp/prod/email`: Email service credentials
- `kv/tennisapp/prod/redis`: Redis credentials
- `kv/tennisapp/prod/api`: API configuration

### Platform-specific Credentials
- `kv/tennisapp/prod/platforms/lta`: LTA/Clubspark credentials
- `kv/tennisapp/prod/platforms/courtsides`: Courtsides/Tennis Tower Hamlets credentials

### Notification Service Credentials
- `kv/tennisapp/prod/notifications/twilio`: Twilio (SMS) credentials
- `kv/tennisapp/prod/notifications/sendgrid`: SendGrid (Email) credentials

## Access Policies

The following policies are defined:

### Service-specific Policies
- `db-policy`: Read-only access to database credentials
- `jwt-policy`: Read-only access to JWT signing keys
- `email-policy`: Read-only access to email service credentials
- `redis-policy`: Read-only access to Redis credentials
- `api-policy`: Read-only access to API configuration

### Group Policies
- `platforms-policy`: Read-only access to all platform credentials
- `notifications-policy`: Read-only access to all notification service credentials

### Administrative Policies
- `admin-policy`: Full access to all Tennis Booking application secrets

## Initial Setup Script

The `init-vault.sh` script automates the initial setup of Vault:

```bash
# For development
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-token
./infrastructure/vault/init-vault.sh

# For production (after unsealing)
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=<your-root-token>
./infrastructure/vault/init-vault.sh
```

## Accessing Secrets

### Command Line

```bash
# Set environment variables
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=<your-token>

# Read secrets
vault kv get kv/tennisapp/prod/db
vault kv get kv/tennisapp/prod/jwt
vault kv get kv/tennisapp/prod/email
vault kv get kv/tennisapp/prod/redis
vault kv get kv/tennisapp/prod/api
vault kv get kv/tennisapp/prod/platforms/lta
vault kv get kv/tennisapp/prod/platforms/courtsides
vault kv get kv/tennisapp/prod/notifications/twilio
vault kv get kv/tennisapp/prod/notifications/sendgrid
```

### From Application

The application services are configured with the `VAULT_ADDR` environment variable. They should use the Vault client library for their respective language to authenticate and retrieve secrets.

Example (Node.js):

```javascript
const vault = require('node-vault')({
  apiVersion: 'v1',
  endpoint: process.env.VAULT_ADDR
});

// Authenticate with token or other method
await vault.tokenLogin({ token: process.env.VAULT_TOKEN });

// Read secrets
const dbSecrets = await vault.read('kv/data/tennisapp/prod/db');
const { username, password } = dbSecrets.data.data;
```

Example (Python):

```python
import hvac

# Initialize the client
client = hvac.Client(url=os.environ['VAULT_ADDR'])

# Authenticate
client.token = os.environ['VAULT_TOKEN']

# Read secrets
db_secrets = client.secrets.kv.v2.read_secret_version(
    path='tennisapp/prod/db',
    mount_point='kv'
)
username = db_secrets['data']['data']['username']
password = db_secrets['data']['data']['password']
```

## Security Considerations

- In production, enable TLS by configuring certificates in `vault.hcl`
- Store unseal keys securely and distribute them to different team members
- Rotate the root token regularly
- Use least-privilege policies for application services
- Consider using auto-unseal with a cloud KMS for production 
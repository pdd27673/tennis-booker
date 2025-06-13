# Vault Seeding Scripts

This directory contains scripts for seeding HashiCorp Vault with credentials for the Tennis Booker application.

## 🔐 Scripts Overview

### `seed-vault-dev.sh`
- **Purpose**: Seeds Vault with development credentials
- **Environment**: Local development only
- **Credentials**: Contains real development credentials
- **Git Status**: ⚠️ **GIT-IGNORED** - Contains sensitive data

### `seed-vault-production.sh`
- **Purpose**: Template for production Vault seeding
- **Environment**: Production deployment
- **Credentials**: Contains placeholder values that must be replaced
- **Git Status**: ✅ **TRACKED** - Safe template with no real credentials

## 🚀 Usage

### Development Setup

1. **Start Vault in dev mode:**
   ```bash
   docker-compose up vault
   ```

2. **Run the development seeding script:**
   ```bash
   cd scripts/vault
   chmod +x seed-vault-dev.sh
   ./seed-vault-dev.sh
   ```

3. **Verify secrets are loaded:**
   ```bash
   export VAULT_ADDR=http://localhost:8200
   export VAULT_TOKEN=dev-token
   vault kv get kv/tennisapp/prod/email
   ```

### Production Setup

1. **Copy the production template:**
   ```bash
   cp seed-vault-production.sh seed-vault-production-real.sh
   ```

2. **Edit the copied script with real production values:**
   ```bash
   # Replace all REPLACE_WITH_* placeholders with real values
   nano seed-vault-production-real.sh
   ```

3. **Set production Vault environment:**
   ```bash
   export VAULT_ADDR=https://your-production-vault.com
   export VAULT_TOKEN=your-production-token
   ```

4. **Run the production seeding script:**
   ```bash
   chmod +x seed-vault-production-real.sh
   ./seed-vault-production-real.sh
   ```

5. **🔥 IMPORTANT: Delete the script after use:**
   ```bash
   rm seed-vault-production-real.sh
   ```

## 🔑 Secrets Structure

The scripts create the following secret paths in Vault:

```
kv/tennisapp/prod/
├── email/          # Gmail SMTP credentials
├── jwt/            # JWT signing secret
├── db/             # MongoDB connection details
├── redis/          # Redis connection details
├── api/            # External API keys
├── platform/       # Platform login credentials
└── notifications/  # Notification service tokens
```

## 🛡️ Security Best Practices

### Development
- ✅ Use the provided `seed-vault-dev.sh` script
- ✅ Keep development credentials separate from production
- ✅ Never commit real credentials to git

### Production
- ⚠️ **NEVER** commit production credentials to git
- ⚠️ **ALWAYS** delete production seeding scripts after use
- ⚠️ **USE** strong, unique passwords for production
- ⚠️ **ROTATE** credentials regularly
- ⚠️ **AUDIT** Vault access logs

## 🔄 Credential Rotation

To rotate credentials:

1. **Update the credential in Vault:**
   ```bash
   vault kv put kv/tennisapp/prod/email \
       email="new-email@domain.com" \
       password="new-app-password"
   ```

2. **Restart affected services** to pick up new credentials

3. **Verify** the rotation worked by checking service logs

## 🆘 Troubleshooting

### Common Issues

**"Permission denied" errors:**
```bash
chmod +x seed-vault-*.sh
```

**"Vault not found" errors:**
```bash
# Install Vault CLI
brew install vault  # macOS
# or
sudo apt-get install vault  # Ubuntu
```

**"Connection refused" errors:**
```bash
# Check Vault is running
docker-compose ps vault
# Check Vault address
echo $VAULT_ADDR
```

### Verification Commands

```bash
# List all secrets
vault kv list kv/tennisapp/prod/

# Get specific secret
vault kv get kv/tennisapp/prod/email

# Check Vault status
vault status
``` 
# Backend Scripts

## Vault Setup

### Development/Testing
Use the existing script with your real credentials:
```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-token
./setup-vault-secrets.sh
```

### Production
1. Copy the template: `cp setup-vault-secrets-template.sh setup-vault-secrets-production.sh`
2. Replace all `REPLACE_WITH_*` placeholders with real production values
3. **Never commit the production script to git** (it's git-ignored)
4. Run: `./setup-vault-secrets-production.sh`

### Security Notes
- The real `setup-vault-secrets.sh` contains actual credentials and is git-ignored
- Only the template `setup-vault-secrets-template.sh` is committed to git
- Always use strong, unique credentials in production
- Use proper Vault authentication (not dev tokens) in production 
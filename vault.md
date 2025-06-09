# HashiCorp Vault Configuration

## Setup

This project uses HashiCorp Vault for secure credential management. Vault is configured in the Docker Compose stack.

### Development Environment

- **Vault Address**: `http://localhost:8200`
- **Development Token**: `dev-token`
- **Secret Path**: `secret/tennis-bot/`

### Stored Secrets

#### LTA Platform
**Path**: `secret/tennis-bot/lta`
- `username`: LTA username for authentication
- `password`: LTA password for authentication  
- `api_key`: LTA API key (if applicable)
- `base_url`: LTA booking platform base URL

#### Courtsides Platform
**Path**: `secret/tennis-bot/courtsides`
- `username`: Courtsides username for authentication
- `password`: Courtsides password for authentication
- `login_url`: Courtsides login URL
- `booking_url`: Courtsides booking API URL

## Usage

### Accessing Vault UI
Visit: http://localhost:8200/ui
Token: `dev-token`

### Environment Variables for Applications
```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-token  # Use AppRole in production
```

### API Access
```bash
# Read LTA credentials
curl -H "X-Vault-Token: dev-token" \
  http://localhost:8200/v1/secret/data/tennis-bot/lta

# Read Courtsides credentials  
curl -H "X-Vault-Token: dev-token" \
  http://localhost:8200/v1/secret/data/tennis-bot/courtsides
```

## Security Notes

- The current setup uses a development token for simplicity
- Production should use AppRole authentication or other secure methods
- Never commit vault tokens to version control
- Credentials shown here are test values, not real platform credentials 
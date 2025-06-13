#!/bin/sh
set -e

SERVICE_NAME=${SERVICE_NAME:-backend}

echo "🔧 Setting up Vault Agent environment for service: $SERVICE_NAME"

# Fix ownership of secrets directory for vault user
chown vault:vault /vault/secrets
chmod 755 /vault/secrets

# Generate service-specific configuration
echo "📝 Generating Vault Agent configuration for $SERVICE_NAME..."
cat > "/vault/config/$SERVICE_NAME-agent.hcl" << EOF
pid_file = "/tmp/pidfile"

vault {
  address = "http://vault:8200"
}

auto_auth {
  method "token_file" {
    config = {
      token_file_path = "/vault/token"
    }
  }

  sink "file" {
    config = {
      path = "/tmp/vault-token"
    }
  }
}

template {
  source      = "/vault/templates/$SERVICE_NAME-env.tpl"
  destination = "/vault/secrets/$SERVICE_NAME.env"
  perms       = 0644
}

template {
  source      = "/vault/templates/$SERVICE_NAME-config.json.tpl"
  destination = "/vault/secrets/$SERVICE_NAME-config.json"
  perms       = 0644
}
EOF

echo "✅ Configuration generated: /vault/config/$SERVICE_NAME-agent.hcl"

# Start Vault Agent
echo "🚀 Starting Vault Agent for $SERVICE_NAME..."
exec vault agent -config="/vault/config/$SERVICE_NAME-agent.hcl" 
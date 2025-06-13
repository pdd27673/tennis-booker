#!/bin/bash
set -e

SERVICE_NAME=${1:-backend}
CONFIG_DIR="/vault/config"
TEMPLATE_DIR="/vault/templates"
SECRETS_DIR="/vault/secrets"

echo "Generating Vault Agent config for service: $SERVICE_NAME"

cat > "$CONFIG_DIR/$SERVICE_NAME-agent.hcl" << EOF
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
  source      = "$TEMPLATE_DIR/$SERVICE_NAME-env.tpl"
  destination = "$SECRETS_DIR/$SERVICE_NAME.env"
  perms       = 0644
}

template {
  source      = "$TEMPLATE_DIR/$SERVICE_NAME-config.json.tpl"
  destination = "$SECRETS_DIR/$SERVICE_NAME-config.json"
  perms       = 0644
}
EOF

echo "Generated config: $CONFIG_DIR/$SERVICE_NAME-agent.hcl" 
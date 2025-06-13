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
  source      = "/vault/templates/notification-env.tpl"
  destination = "/vault/secrets/notification.env"
  perms       = 0644
}

template {
  source      = "/vault/templates/notification-config.json.tpl"
  destination = "/vault/secrets/notification-config.json"
  perms       = 0644
}

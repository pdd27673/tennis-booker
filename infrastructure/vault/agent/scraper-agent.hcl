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
  source      = "/vault/templates/scraper-env.tpl"
  destination = "/vault/secrets/scraper.env"
  perms       = 0644
}

template {
  source      = "/vault/templates/scraper-config.json.tpl"
  destination = "/vault/secrets/scraper-config.json"
  perms       = 0644
}

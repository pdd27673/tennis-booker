storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1  # Disable TLS for development/testing; enable for production with proper certificates
}

api_addr = "http://0.0.0.0:8200"
cluster_addr = "https://0.0.0.0:8201"

ui = true

# Disable mlock for containers as it's not needed when using Docker volumes
disable_mlock = true 
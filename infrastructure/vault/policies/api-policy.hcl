# API policy for Tennis Booking application
# Grants read-only access to API configuration

path "kv/data/tennisapp/prod/api" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/api" {
  capabilities = ["list"]
} 
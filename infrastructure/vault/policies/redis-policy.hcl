# Redis policy for Tennis Booking application
# Grants read-only access to Redis credentials

path "kv/data/tennisapp/prod/redis" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/redis" {
  capabilities = ["list"]
} 
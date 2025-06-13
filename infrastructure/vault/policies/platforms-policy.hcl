# Platforms policy for Tennis Booking application
# Grants read-only access to all platform credentials (LTA, Courtsides, etc.)

path "kv/data/tennisapp/prod/platforms/*" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/platforms/*" {
  capabilities = ["list"]
} 
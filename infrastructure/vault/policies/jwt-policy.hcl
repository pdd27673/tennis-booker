# JWT policy for Tennis Booking application
# Grants read-only access to JWT signing keys

path "kv/data/tennisapp/prod/jwt" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/jwt" {
  capabilities = ["list"]
} 
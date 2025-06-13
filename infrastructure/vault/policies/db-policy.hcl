# Database policy for Tennis Booking application
# Grants read-only access to database credentials

path "kv/data/tennisapp/prod/db" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/db" {
  capabilities = ["list"]
} 
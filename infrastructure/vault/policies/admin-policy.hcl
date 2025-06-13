# Admin policy for Tennis Booking application
# Grants full access to all Tennis Booking application secrets

path "kv/data/tennisapp/prod/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv/metadata/tennisapp/prod/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Allow listing of secret engines
path "sys/mounts" {
  capabilities = ["read"]
}

# Allow listing of policies
path "sys/policies/acl" {
  capabilities = ["list"]
}

# Allow reading specific policies
path "sys/policies/acl/*" {
  capabilities = ["read"]
} 
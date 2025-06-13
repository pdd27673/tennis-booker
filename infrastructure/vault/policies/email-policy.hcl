# Email policy for Tennis Booking application
# Grants read-only access to email service credentials

path "kv/data/tennisapp/prod/email" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/email" {
  capabilities = ["list"]
} 
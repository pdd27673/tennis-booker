# Notifications policy for Tennis Booking application
# Grants read-only access to all notification service credentials (Twilio, SendGrid, etc.)

path "kv/data/tennisapp/prod/notifications/*" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/notifications/*" {
  capabilities = ["list"]
} 
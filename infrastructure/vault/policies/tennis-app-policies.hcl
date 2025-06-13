# Tennis Booking Application - Comprehensive Vault Policies
# This file contains all access policies for the Tennis Booking application

# =============================================================================
# DATABASE POLICY
# Grants read-only access to database credentials
# =============================================================================
path "kv/data/tennisapp/prod/db" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/db" {
  capabilities = ["list"]
}

# =============================================================================
# JWT POLICY
# Grants read-only access to JWT signing keys
# =============================================================================
path "kv/data/tennisapp/prod/jwt" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/jwt" {
  capabilities = ["list"]
}

# =============================================================================
# EMAIL POLICY
# Grants read-only access to email service credentials
# =============================================================================
path "kv/data/tennisapp/prod/email" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/email" {
  capabilities = ["list"]
}

# =============================================================================
# REDIS POLICY
# Grants read-only access to Redis credentials
# =============================================================================
path "kv/data/tennisapp/prod/redis" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/redis" {
  capabilities = ["list"]
}

# =============================================================================
# API POLICY
# Grants read-only access to API configuration
# =============================================================================
path "kv/data/tennisapp/prod/api" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/api" {
  capabilities = ["list"]
}

# =============================================================================
# PLATFORMS POLICY
# Grants read-only access to all platform credentials (LTA, Courtsides, etc.)
# =============================================================================
path "kv/data/tennisapp/prod/platforms/*" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/platforms/*" {
  capabilities = ["list"]
}

# =============================================================================
# NOTIFICATIONS POLICY
# Grants read-only access to all notification service credentials (Twilio, SendGrid, etc.)
# =============================================================================
path "kv/data/tennisapp/prod/notifications/*" {
  capabilities = ["read"]
}

path "kv/metadata/tennisapp/prod/notifications/*" {
  capabilities = ["list"]
}

# =============================================================================
# ADMIN POLICY
# Grants full access to all Tennis Booking application secrets
# =============================================================================
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
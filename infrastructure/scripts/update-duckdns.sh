#!/bin/bash

# DuckDNS IP Update Script
# This script updates your DuckDNS subdomain with the current public IP

set -e

# Load environment variables
if [ -f ".env.prod" ]; then
    source .env.prod
elif [ -f ".env" ]; then
    source .env
else
    echo "Error: No environment file found (.env.prod or .env)"
    exit 1
fi

# Check required variables
if [ -z "$DUCKDNS_TOKEN" ]; then
    echo "Error: DUCKDNS_TOKEN not set in environment file"
    exit 1
fi

if [ -z "$DOMAIN_NAME" ]; then
    echo "Error: DOMAIN_NAME not set in environment file"
    exit 1
fi

# Extract subdomain from DOMAIN_NAME (remove .duckdns.org if present)
SUBDOMAIN=$(echo "$DOMAIN_NAME" | sed 's/\.duckdns\.org$//')

echo "Updating DuckDNS for subdomain: $SUBDOMAIN"

# Get current public IP
echo "Getting current public IP..."
PUBLIC_IP=$(curl -s https://ipv4.icanhazip.com/ || curl -s https://api.ipify.org || curl -s https://checkip.amazonaws.com)

if [ -z "$PUBLIC_IP" ]; then
    echo "Error: Could not determine public IP address"
    exit 1
fi

echo "Current public IP: $PUBLIC_IP"

# Update DuckDNS
echo "Updating DuckDNS..."
RESPONSE=$(curl -s "https://www.duckdns.org/update?domains=$SUBDOMAIN&token=$DUCKDNS_TOKEN&ip=$PUBLIC_IP")

if [ "$RESPONSE" = "OK" ]; then
    echo "✅ DuckDNS updated successfully!"
    echo "Domain: $SUBDOMAIN.duckdns.org -> $PUBLIC_IP"
else
    echo "❌ DuckDNS update failed. Response: $RESPONSE"
    exit 1
fi

# Verify DNS resolution
echo "Verifying DNS resolution..."
sleep 5
RESOLVED_IP=$(nslookup "$SUBDOMAIN.duckdns.org" | grep -A1 "Name:" | tail -1 | awk '{print $2}' || echo "")

if [ "$RESOLVED_IP" = "$PUBLIC_IP" ]; then
    echo "✅ DNS resolution verified: $SUBDOMAIN.duckdns.org resolves to $PUBLIC_IP"
else
    echo "⚠️  DNS resolution may take a few minutes to propagate"
    echo "Expected: $PUBLIC_IP, Got: $RESOLVED_IP"
fi

echo "DuckDNS update complete!" 
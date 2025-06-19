#!/bin/bash

# Generate password hash for Traefik dashboard
# Usage: ./generate-password.sh username password

if [ $# -ne 2 ]; then
    echo "Usage: $0 <username> <password>"
    echo "Example: $0 admin mypassword"
    exit 1
fi

USERNAME=$1
PASSWORD=$2

# Check if htpasswd is available
if command -v htpasswd >/dev/null 2>&1; then
    # Use htpasswd if available
    HASH=$(htpasswd -nbB "$USERNAME" "$PASSWORD" | cut -d: -f2)
elif command -v openssl >/dev/null 2>&1; then
    # Use openssl as fallback
    HASH=$(openssl passwd -apr1 "$PASSWORD")
elif command -v python3 >/dev/null 2>&1; then
    # Use Python as fallback
    HASH=$(python3 -c "
import crypt
import getpass
print(crypt.crypt('$PASSWORD', crypt.mksalt(crypt.METHOD_SHA512)))
")
else
    echo "Error: No password hashing utility found (htpasswd, openssl, or python3)"
    exit 1
fi

echo "Generated hash for user '$USERNAME':"
echo "$USERNAME:$HASH"
echo ""
echo "Add this to your .env file:"
echo "TRAEFIK_USER=$USERNAME"
echo "TRAEFIK_PASSWORD_HASH=$HASH" 
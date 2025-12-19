#!/bin/bash

# Tennis Booker Production Deployment Script
# Deploys code from local machine to OCI production host

set -e  # Exit on any error

# Configuration
HOST="opc@79.72.94.79"
SSH_KEY="$HOME/.ssh/tennis_booker_key"
REMOTE_DIR="/opt/tennis-booker"
LOCAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🚀 Tennis Booker Production Deployment"
echo "========================================"
echo "Local:  $LOCAL_DIR"
echo "Remote: $HOST:$REMOTE_DIR"
echo ""

# Verify SSH key exists
if [ ! -f "$SSH_KEY" ]; then
    echo "❌ SSH key not found: $SSH_KEY"
    exit 1
fi

# Confirm deployment
read -p "⚠️  Deploy to PRODUCTION? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Deployment cancelled"
    exit 0
fi

echo ""
echo "📦 Step 1: Creating remote backup..."
ssh -i "$SSH_KEY" "$HOST" "sudo cp -r $REMOTE_DIR ${REMOTE_DIR}.backup-${TIMESTAMP}"
echo "✅ Backup created: ${REMOTE_DIR}.backup-${TIMESTAMP}"

echo ""
echo "📤 Step 2: Uploading code to temporary location..."
ssh -i "$SSH_KEY" "$HOST" "rm -rf /tmp/tennis-booker-deploy"
scp -i "$SSH_KEY" -r "$LOCAL_DIR" "$HOST:/tmp/tennis-booker-deploy"
echo "✅ Code uploaded"

echo ""
echo "🔄 Step 3: Replacing code on server..."
ssh -i "$SSH_KEY" "$HOST" "
    cd $REMOTE_DIR && \
    docker-compose -f docker-compose.prod.yml down && \
    sudo rsync -av --exclude='.env' --exclude='data/' /tmp/tennis-booker-deploy/ $REMOTE_DIR/ && \
    sudo chown -R opc:opc $REMOTE_DIR
"
echo "✅ Code replaced"

echo ""
echo "🔨 Step 4: Rebuilding Docker images..."
ssh -i "$SSH_KEY" "$HOST" "
    cd $REMOTE_DIR && \
    docker-compose -f docker-compose.prod.yml build --no-cache scraper-service notification-service backend
"
echo "✅ Images rebuilt"

echo ""
echo "🚀 Step 5: Starting services..."
ssh -i "$SSH_KEY" "$HOST" "
    cd $REMOTE_DIR && \
    docker-compose -f docker-compose.prod.yml up -d
"
echo "✅ Services started"

echo ""
echo "🔍 Step 6: Verifying deployment..."
sleep 5
ssh -i "$SSH_KEY" "$HOST" "cd $REMOTE_DIR && docker-compose -f docker-compose.prod.yml ps"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Next steps:"
echo "   - Monitor logs: ssh -i $SSH_KEY $HOST 'docker logs tennis-scraper -f'"
echo "   - Test scraper:  ssh -i $SSH_KEY $HOST 'docker exec tennis-scraper python -m src.main --venue \"Victoria Park\"'"
echo "   - Rollback if needed: ssh -i $SSH_KEY $HOST 'sudo mv $REMOTE_DIR ${REMOTE_DIR}.failed-${TIMESTAMP} && sudo mv ${REMOTE_DIR}.backup-${TIMESTAMP} $REMOTE_DIR && cd $REMOTE_DIR && docker-compose -f docker-compose.prod.yml up -d'"

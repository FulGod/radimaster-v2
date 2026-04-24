#!/bin/bash
set -euo pipefail

# ============================================================
# RadiMaster V2 — Deployment Script
# Usage:
#   ./deploy.sh stage       → deploy branch 'stage' to staging
#   ./deploy.sh production  → deploy branch 'master' to production
# ============================================================

SERVER="radimaster@canthorivertour.com"
REPO="git@github.com:FulGod/radimaster.git"
BASE_DIR="/home/deploy/projects"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${CYAN}[deploy]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1" >&2; exit 1; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }

ENV="${1:-}"
if [[ -z "$ENV" ]]; then
  echo "Usage: ./deploy.sh <stage|production> [branch]"
  echo ""
  echo "  stage       → branch 'stage'  → $BASE_DIR/stage/radimaster      (port 3001)"
  echo "  production  → branch 'master' → $BASE_DIR/production/radimaster  (port 3000)"
  exit 1
fi

case "$ENV" in
  stage)
    REMOTE_DIR="$BASE_DIR/stage/radimaster"
    BRANCH="${2:-stage}"
    COMPOSE_FILE="docker-compose.yml -f docker-compose.stage.yml"
    ENV_FILE=".env.stage"
    ;;
  production)
    REMOTE_DIR="$BASE_DIR/production/radimaster"
    BRANCH="${2:-master}"
    COMPOSE_FILE="docker-compose.yml -f docker-compose.production.yml"
    ENV_FILE=".env.production"
    ;;
  *)
    err "Unknown environment: $ENV (use stage or production)"
    ;;
esac

log "Deploying V2 ($ENV) to $REMOTE_DIR..."
log "Branch: $BRANCH"
echo ""

# Step 1: Sync code
log "Step 1/4 — Syncing code to server..."
ssh "$SERVER" bash -s <<REMOTE_SETUP
  set -e
  if [ ! -d "$REMOTE_DIR/.git" ]; then
    echo "→ Cloning repository..."
    rm -rf $REMOTE_DIR 2>/dev/null || true
    mkdir -p \$(dirname $REMOTE_DIR)
    git clone -b $BRANCH $REPO $REMOTE_DIR
  else
    cd $REMOTE_DIR
    echo "→ Fetching latest..."
    git fetch origin
    git checkout $BRANCH
    git pull origin $BRANCH
  fi
REMOTE_SETUP
ok "Code synced (branch: $BRANCH)"

# Step 2: Copy env file
log "Step 2/4 — Configuring environment..."
scp "v2/$ENV_FILE" "$SERVER:$REMOTE_DIR/v2/.env"
ok "Environment configured ($ENV_FILE → .env)"

# Step 3: Check medical data
MED_DATA="$BASE_DIR/radimaster/v1/public/med-data-bak"
log "Step 3/4 — Checking medical data..."
ssh "$SERVER" bash -s <<REMOTE_DATA
  set -e
  if [ -d "$MED_DATA" ]; then
    COUNT=\$(ls "$MED_DATA" | wc -l | tr -d ' ')
    echo "→ Medical data found: \$COUNT case directories"
  else
    echo "⚠️  Medical data NOT found at $MED_DATA"
    echo "→ Run: rsync -avz med-data/ $SERVER:$MED_DATA/"
  fi
REMOTE_DATA
ok "Data check complete"

# Step 4: Build & deploy
log "Step 4/4 — Building and deploying containers..."
ssh "$SERVER" bash -s <<REMOTE_DEPLOY
  set -e
  cd $REMOTE_DIR/v2

  echo "→ Stopping existing containers..."
  docker compose -f $COMPOSE_FILE --env-file .env down --remove-orphans 2>/dev/null || true

  echo "→ Building images..."
  docker compose -f $COMPOSE_FILE --env-file .env build --no-cache

  echo "→ Starting services..."
  docker compose -f $COMPOSE_FILE --env-file .env up -d

  echo ""
  echo "→ Container status:"
  docker compose -f $COMPOSE_FILE --env-file .env ps

  echo ""
  echo "→ Waiting for health checks..."
  sleep 5
  docker compose -f $COMPOSE_FILE --env-file .env ps
REMOTE_DEPLOY

echo ""
ok "V2 ($ENV) deployed — branch: $BRANCH"
echo ""

case "$ENV" in
  stage)
    log "🌐 Access: http://canthorivertour.com:3001"
    log "🔌 API:    http://canthorivertour.com:8081/api/"
    log "📡 WS:     ws://canthorivertour.com:8081/ws/"
    ;;
  production)
    log "🌐 Access: http://canthorivertour.com:3000"
    log "🔌 API:    http://canthorivertour.com:8080/api/"
    log "📡 WS:     ws://canthorivertour.com:8080/ws/"
    ;;
esac

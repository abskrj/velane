#!/usr/bin/env bash
# setup-nango.sh — one-time Nango environment bootstrap
#
# Creates the Nango account and "dev" environment so Nango's DB has the
# required row. After this, NANGO_SECRET_KEY_DEV in the Nango container
# and NANGO_SECRET_KEY on the control-plane control auth — no DB reads needed.
#
# Idempotent: safe to re-run, exits 0 if already set up.
#
# Usage:
#   NANGO_SECRET_KEY=<uuid> make setup-nango
#
# The key you pass becomes the stable API key. Store it in your secrets manager.

set -euo pipefail

log() { echo "$(date -u +%T) [setup-nango] $*" >&2; }

ADMIN_EMAIL="${NANGO_ADMIN_EMAIL:-admin@velane.internal}"
ADMIN_PASSWORD="${NANGO_ADMIN_PASSWORD:-Velane-Setup-$(openssl rand -hex 4)!}"
ADMIN_NAME="${NANGO_ADMIN_NAME:-velane}"

psql_nango() {
  docker compose exec -T postgres psql -U velane -d nango "$@"
}

nango_api() {
  local method="$1" path="$2" body="${3:-}"
  docker compose exec -T nango node -e "
    const http = require('http');
    const body = $([ -n "$body" ] && echo "JSON.stringify($body)" || echo "null");
    const opts = {
      host: 'localhost', port: 3003, path: '$path', method: '$method',
      headers: { 'Content-Type': 'application/json' }
    };
    if (body) opts.headers['Content-Length'] = Buffer.byteLength(body);
    const req = http.request(opts, res => {
      let b = ''; res.on('data', d => b += d);
      res.on('end', () => { process.stdout.write(res.statusCode + ' ' + b); });
    });
    req.on('error', e => { process.stderr.write(e.message); process.exit(1); });
    if (body) req.write(body);
    req.end();
  "
}

# ---- Wait for Nango ----
log "waiting for Nango ..."
for i in $(seq 1 30); do
  result=$(docker compose exec -T nango node -e \
    "require('http').get('http://localhost:3003/', r => { process.stdout.write(String(r.statusCode)); r.resume(); }).on('error', () => process.stdout.write('0'));" 2>/dev/null || echo "0")
  if [ "$result" = "200" ] || [ "$result" = "302" ]; then
    log "Nango is ready"
    break
  fi
  if [ "$i" -eq 30 ]; then
    log "ERROR: Nango did not become ready in 60s"
    exit 1
  fi
  sleep 2
done

# ---- Check if already set up ----
existing=$(psql_nango -tAc \
  "SELECT COUNT(*) FROM nango._nango_environments WHERE name = 'dev';" 2>/dev/null || echo "0")
existing=$(echo "$existing" | tr -d '[:space:]')

if [ "$existing" != "0" ] && [ "$existing" != "" ]; then
  log "environment already exists — nothing to do"
  log "ensure NANGO_SECRET_KEY is set in your environment and matches NANGO_SECRET_KEY_DEV on the nango container"
  exit 0
fi

# ---- Create account ----
log "creating Nango account ($ADMIN_EMAIL)"
response=$(nango_api POST /api/v1/account/signup \
  "{name:'$ADMIN_NAME',email:'$ADMIN_EMAIL',password:'$ADMIN_PASSWORD'}")
status=$(echo "$response" | cut -d' ' -f1)
if [ "$status" != "200" ] && [ "$status" != "409" ]; then
  log "ERROR: signup failed — $response"
  exit 1
fi
log "account created"

# ---- Verify email (no SMTP configured) ----
psql_nango -c \
  "UPDATE nango._nango_users SET email_verified = true WHERE email = '$ADMIN_EMAIL';" \
  > /dev/null
log "email verified"

log "setup complete"
log ""
log "Add this to your environment (or .env file) and restart:"
log "  NANGO_SECRET_KEY=<choose-a-uuid>   # e.g. $(python3 -c 'import uuid; print(uuid.uuid4())')"
log ""
log "This value feeds both:"
log "  - NANGO_SECRET_KEY_DEV on the nango container  (authenticates API calls)"
log "  - NANGO_SECRET_KEY on the control-plane         (used by Go code to call Nango)"

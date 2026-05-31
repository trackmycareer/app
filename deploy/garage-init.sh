#!/bin/sh
set -e

ADMIN="http://garage:3903"
TOKEN="${GARAGE_ADMIN_TOKEN:-trackmy-admin-token}"
AUTH="Authorization: Bearer $TOKEN"
ACCESS_KEY="${S3_ACCESS_KEY_ID:-trackmy-dev-key}"
SECRET_KEY="${S3_SECRET_ACCESS_KEY:-trackmy-dev-secret}"
BUCKET="${S3_BUCKET:-avatars}"

# ── Wait for Garage to become healthy ─────────────────────────────────────────
echo "Waiting for Garage to become healthy..."
until curl -sf "$ADMIN/health" > /dev/null 2>&1; do
  sleep 1
done
echo "Garage is healthy."

# ── Get node ID ───────────────────────────────────────────────────────────────
NODE_ID=$(curl -sf -H "$AUTH" "$ADMIN/v1/status" | sed -n 's/.*"node":"\([^"]*\)".*/\1/p')
if [ -z "$NODE_ID" ]; then
  echo "ERROR: Could not retrieve node ID from Garage status."
  exit 1
fi
echo "Node ID: $NODE_ID"

# ── Assign layout ────────────────────────────────────────────────────────────
echo "Assigning cluster layout..."
curl -sf -X POST -H "$AUTH" -H "Content-Type: application/json" \
  "$ADMIN/v1/layout" \
  -d "[{\"id\": \"$NODE_ID\", \"zone\": \"dc1\", \"capacity\": 1073741824, \"tags\": [\"node1\"]}]" || true

# ── Apply layout ─────────────────────────────────────────────────────────────
LAYOUT_VERSION=$(curl -sf -H "$AUTH" "$ADMIN/v1/layout" | sed -n 's/.*"version":\([0-9]*\).*/\1/p')
NEXT_VERSION=$((LAYOUT_VERSION + 1))
echo "Applying layout version $NEXT_VERSION..."
curl -sf -X POST -H "$AUTH" -H "Content-Type: application/json" \
  "$ADMIN/v1/layout/apply" \
  -d "{\"version\": $NEXT_VERSION}" || true

# ── Import API key (409 if it already exists) ─────────────────────────────────
echo "Importing API key..."
curl -sf -X POST -H "$AUTH" -H "Content-Type: application/json" \
  "$ADMIN/v1/key/import" \
  -d "{\"accessKeyId\":\"$ACCESS_KEY\",\"secretAccessKey\":\"$SECRET_KEY\",\"name\":\"trackmy-career\"}" || true

# ── Create bucket (409 if it already exists) ──────────────────────────────────
echo "Creating $BUCKET bucket..."
BUCKET_RESP=$(curl -s -X POST -H "$AUTH" -H "Content-Type: application/json" \
  "$ADMIN/v1/bucket" \
  -d "{\"globalAlias\":\"$BUCKET\"}")
BUCKET_ID=$(echo "$BUCKET_RESP" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

# If the bucket already existed (409), look up its ID instead.
if [ -z "$BUCKET_ID" ]; then
  BUCKET_ID=$(curl -sf -H "$AUTH" "$ADMIN/v1/bucket?alias=$BUCKET" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
fi

if [ -z "$BUCKET_ID" ]; then
  echo "ERROR: Could not determine bucket ID for '$BUCKET'."
  exit 1
fi
echo "Bucket ID: $BUCKET_ID"

# ── Grant key permissions on the bucket ───────────────────────────────────────
echo "Granting key permissions on $BUCKET bucket..."
curl -sf -X POST -H "$AUTH" -H "Content-Type: application/json" \
  "$ADMIN/v1/bucket/allow" \
  -d "{\"bucketId\":\"$BUCKET_ID\",\"accessKeyId\":\"$ACCESS_KEY\",\"permissions\":{\"read\":true,\"write\":true,\"owner\":true}}" || true

# ── Set anonymous (public) read access via mc ─────────────────────────────────
echo "Configuring anonymous read access on $BUCKET bucket..."
mc alias set garage http://garage:3900 "$ACCESS_KEY" "$SECRET_KEY"
mc anonymous set download "garage/$BUCKET"

echo "Garage initialisation complete."

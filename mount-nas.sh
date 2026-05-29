#!/bin/sh

set -e

NAS_SHARE_PATH="${NAS_SHARE_PATH:-//10.3.1.161/TzuChiApp}"
NAS_MOUNT_PATH="${NAS_MOUNT_PATH:-/mnt/nas}"
NAS_SHARE_USER="${NAS_SHARE_USER:-}"
NAS_SHARE_PASSWORD="${NAS_SHARE_PASSWORD:-}"
NAS_SHARE_DOMAIN="${NAS_SHARE_DOMAIN:-}"
NAS_MOUNT_OPTIONS="${NAS_MOUNT_OPTIONS:-rw,dir_mode=0777,file_mode=0777,vers=3.0}"

# If credentials not provided, skip mount
if [ -z "$NAS_SHARE_USER" ] || [ -z "$NAS_SHARE_PASSWORD" ]; then
  echo "⚠️  NAS_SHARE_USER and/or NAS_SHARE_PASSWORD not provided, cannot mount NAS"
  exit 1
fi

mkdir -p "$NAS_MOUNT_PATH"

if mountpoint -q "$NAS_MOUNT_PATH" 2>/dev/null; then
  echo "✓ NAS already mounted on $NAS_MOUNT_PATH"
  exit 0
fi

MOUNT_OPTS="$NAS_MOUNT_OPTIONS,username=$NAS_SHARE_USER,password=$NAS_SHARE_PASSWORD"
if [ -n "$NAS_SHARE_DOMAIN" ]; then
  MOUNT_OPTS="$MOUNT_OPTS,domain=$NAS_SHARE_DOMAIN"
fi

echo "📌 Attempting to mount NAS share $NAS_SHARE_PATH to $NAS_MOUNT_PATH"
if mount -t cifs "$NAS_SHARE_PATH" "$NAS_MOUNT_PATH" -o "$MOUNT_OPTS" 2>&1; then
  echo "✓ NAS mounted successfully"
  exit 0
else
  echo "❌ Failed to mount NAS (I/O error or network issue)"
  echo "   Make sure NAS server is reachable and credentials are correct"
  echo "   Stopping because NAS mount is required for uploads."
  exit 1
fi

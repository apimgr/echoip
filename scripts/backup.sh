#!/bin/bash
# echoip backup script

set -e

PROJECTNAME="echoip"
DATA_DIR="/var/lib/echoip"
CONFIG_DIR="/etc/echoip"
LOG_DIR="/var/log/echoip"
BACKUP_DIR="${BACKUP_DIR:-/var/backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/${PROJECTNAME}_${TIMESTAMP}.tar.gz"

echo "📦 Backing up echoip..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create backup
echo "📁 Creating backup archive..."
tar -czf "$BACKUP_FILE" \
    -C / \
    --exclude='*.log' \
    var/lib/echoip \
    etc/echoip \
    2>/dev/null || true

# Check if backup was created
if [ -f "$BACKUP_FILE" ]; then
    SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✅ Backup created: $BACKUP_FILE ($SIZE)"

    # Keep only last 7 backups
    echo "🧹 Cleaning old backups (keeping last 7)..."
    ls -t "$BACKUP_DIR/${PROJECTNAME}"_*.tar.gz | tail -n +8 | xargs -r rm -f

    echo "✅ Backup complete!"
    ls -lh "$BACKUP_DIR/${PROJECTNAME}"_*.tar.gz 2>/dev/null || true
else
    echo "❌ Backup failed"
    exit 1
fi

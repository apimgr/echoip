#!/bin/bash
# echoip uninstallation script

set -e

PROJECTNAME="echoip"
BINARY_PATH="/usr/local/bin/echoip"
SERVICE_USER="echoip"
DATA_DIR="/var/lib/echoip"
LOG_DIR="/var/log/echoip"
CONFIG_DIR="/etc/echoip"

echo "🗑️  Uninstalling echoip..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Stop service
if systemctl is-active --quiet $PROJECTNAME; then
    echo "⏸️  Stopping service..."
    systemctl stop $PROJECTNAME
fi

# Disable service
if systemctl is-enabled --quiet $PROJECTNAME; then
    echo "🔴 Disabling service..."
    systemctl disable $PROJECTNAME
fi

# Remove service file
if [ -f "/etc/systemd/system/$PROJECTNAME.service" ]; then
    echo "🗑️  Removing systemd service..."
    rm /etc/systemd/system/$PROJECTNAME.service
    systemctl daemon-reload
fi

# Remove binary
if [ -f "$BINARY_PATH" ]; then
    echo "🗑️  Removing binary..."
    rm $BINARY_PATH
fi

# Ask about data removal
echo ""
read -p "Remove data directories? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  Removing data directories..."
    rm -rf $DATA_DIR
    rm -rf $LOG_DIR
    rm -rf $CONFIG_DIR

    # Remove user
    if id "$SERVICE_USER" &>/dev/null; then
        echo "👤 Removing user $SERVICE_USER..."
        userdel $SERVICE_USER
    fi

    echo "✅ Data removed"
else
    echo "ℹ️  Data preserved in:"
    echo "   - $DATA_DIR"
    echo "   - $LOG_DIR"
    echo "   - $CONFIG_DIR"
    echo "   To remove manually: sudo rm -rf $DATA_DIR $LOG_DIR $CONFIG_DIR"
fi

echo ""
echo "✅ echoip uninstalled successfully!"

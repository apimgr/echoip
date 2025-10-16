#!/bin/bash
# echoip installation script
# Supports Linux (systemd)

set -e

PROJECTNAME="echoip"
BINARY_URL="https://github.com/apimgr/echoip/releases/latest/download"
INSTALL_DIR="/usr/local/bin"
SERVICE_USER="echoip"
DATA_DIR="/var/lib/echoip"
LOG_DIR="/var/log/echoip"
CONFIG_DIR="/etc/echoip"

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY="echoip-linux-amd64"
        ;;
    aarch64|arm64)
        BINARY="echoip-linux-arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "🚀 Installing echoip..."
echo "   Architecture: $ARCH"
echo "   Binary: $BINARY"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Download binary
echo "📥 Downloading binary..."
curl -L -o /tmp/$PROJECTNAME "$BINARY_URL/$BINARY"
chmod +x /tmp/$PROJECTNAME

# Install binary
echo "📦 Installing binary to $INSTALL_DIR..."
mv /tmp/$PROJECTNAME $INSTALL_DIR/$PROJECTNAME

# Create user
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "👤 Creating user $SERVICE_USER..."
    useradd -r -s /bin/false $SERVICE_USER
fi

# Create directories
echo "📁 Creating directories..."
mkdir -p $DATA_DIR $LOG_DIR $CONFIG_DIR
chown $SERVICE_USER:$SERVICE_USER $DATA_DIR $LOG_DIR $CONFIG_DIR

# Create systemd service
echo "⚙️  Creating systemd service..."
cat > /etc/systemd/system/$PROJECTNAME.service <<EOF
[Unit]
Description=echoip - IP address lookup service
After=network.target
Documentation=https://github.com/apimgr/echoip

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=$INSTALL_DIR/$PROJECTNAME -l :8080 -d $DATA_DIR -r -s
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Environment
Environment="CONFIG_DIR=$CONFIG_DIR"
Environment="DATA_DIR=$DATA_DIR"
Environment="LOGS_DIR=$LOG_DIR"

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "🔄 Reloading systemd..."
systemctl daemon-reload

# Enable service
echo "✅ Enabling service..."
systemctl enable $PROJECTNAME

# Start service
echo "▶️  Starting service..."
systemctl start $PROJECTNAME

# Check status
sleep 2
if systemctl is-active --quiet $PROJECTNAME; then
    echo ""
    echo "✅ echoip installed successfully!"
    echo ""
    echo "Service status:"
    systemctl status $PROJECTNAME --no-pager
    echo ""
    echo "Access: curl http://localhost:8080/"
    echo "Logs:   sudo journalctl -u $PROJECTNAME -f"
else
    echo "❌ Service failed to start"
    echo "Check logs: sudo journalctl -u $PROJECTNAME -n 50"
    exit 1
fi

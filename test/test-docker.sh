#!/bin/bash
# Docker testing script for echoip
# Follows SPEC Section 13 & 14 guidelines

set -e

PROJECTNAME="echoip"
TESTPORT=$(shuf -i 64000-64999 -n 1)
CONTAINER_NAME="${PROJECTNAME}-test-${TESTPORT}"

echo "🧪 Testing ${PROJECTNAME} using Docker"
echo "📡 Test Port: ${TESTPORT}"
echo ""

# Clean up function
cleanup() {
    echo "🧹 Cleaning up..."
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
    rm -rf /tmp/${PROJECTNAME}-test
    echo "✓ Cleanup complete"
}

# Set trap for cleanup on exit
trap cleanup EXIT

# Build dev image
echo "🔨 Building development image..."
make docker-dev

# Run container
echo "🚀 Starting container on port ${TESTPORT}..."
docker run -d \
    --name ${CONTAINER_NAME} \
    -p ${TESTPORT}:80 \
    -v /tmp/${PROJECTNAME}-test:/data \
    -e DATA_DIR=/data \
    ${PROJECTNAME}:dev

# Wait for startup
echo "⏳ Waiting for service to start..."
sleep 5

# Test health endpoint
echo "🏥 Testing health endpoint..."
if curl -sf http://localhost:${TESTPORT}/health > /dev/null; then
    echo "✓ Health check passed"
else
    echo "❌ Health check failed"
    docker logs ${CONTAINER_NAME}
    exit 1
fi

# Test basic IP endpoint
echo "🌐 Testing IP endpoint..."
IP_RESULT=$(curl -sf http://localhost:${TESTPORT}/ip)
if [ -n "$IP_RESULT" ]; then
    echo "✓ IP endpoint returned: $IP_RESULT"
else
    echo "❌ IP endpoint failed"
    exit 1
fi

# Test JSON endpoint
echo "📋 Testing JSON endpoint..."
if curl -sf http://localhost:${TESTPORT}/json | grep -q '"ip"'; then
    echo "✓ JSON endpoint passed"
else
    echo "❌ JSON endpoint failed"
    exit 1
fi

# Test API v1
echo "🔌 Testing API v1..."
if curl -sf http://localhost:${TESTPORT}/api/v1 | grep -q '"ip"'; then
    echo "✓ API v1 endpoint passed"
else
    echo "❌ API v1 endpoint failed"
    exit 1
fi

# Test IP lookup
echo "🔍 Testing IP lookup (8.8.8.8)..."
if curl -sf http://localhost:${TESTPORT}/8.8.8.8 | grep -q '"ip"'; then
    echo "✓ IP lookup passed"
else
    echo "❌ IP lookup failed"
    exit 1
fi

# Test API v1 IP lookup
echo "🔍 Testing API v1 IP lookup..."
if curl -sf http://localhost:${TESTPORT}/api/v1/ip/1.1.1.1 | grep -q '"ip"'; then
    echo "✓ API v1 IP lookup passed"
else
    echo "❌ API v1 IP lookup failed"
    exit 1
fi

# Show logs
echo ""
echo "📝 Container logs:"
docker logs ${CONTAINER_NAME} | tail -20

echo ""
echo "✅ All tests passed!"
echo "   Port: ${TESTPORT}"
echo "   Container: ${CONTAINER_NAME}"

#!/bin/bash

# Test Docker build for cross-platform compatibility
set -e

echo "Testing Docker build for cross-platform compatibility..."

# Test single platform build (current architecture)
echo "1. Testing single platform build..."
docker build -t arda-poc-test:single .
echo "✓ Single platform build successful"

# Test multi-platform build (if buildx is available)
if docker buildx version &> /dev/null; then
    echo "2. Testing multi-platform build..."
    docker buildx build --platform linux/amd64,linux/arm64 -t arda-poc-test:multi .
    echo "✓ Multi-platform build successful"
else
    echo "2. Skipping multi-platform build (buildx not available)"
fi

# Test running the container
echo "3. Testing container execution..."
docker run --rm arda-poc-test:single arda-pocd version
echo "✓ Container execution successful"

# Cleanup
echo "4. Cleaning up..."
docker rmi arda-poc-test:single
if docker buildx version &> /dev/null; then
    # Note: multi-platform images aren't automatically loaded into local docker
    # so we don't need to clean them up
    echo "✓ Cleanup complete"
else
    echo "✓ Cleanup complete"
fi

echo "All tests passed! Docker build is working correctly for cross-platform compatibility."
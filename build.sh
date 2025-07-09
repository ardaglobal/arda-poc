#!/bin/bash

# Build script for Railway deployment
echo "Building arda-poc application..."

# Ensure we're in the right directory
cd /app

# Install dependencies
echo "Installing dependencies..."
go mod download
go mod verify

# Build the application
echo "Building arda-pocd..."
go build -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=arda-poc -X github.com/cosmos/cosmos-sdk/version.AppName=arda-pocd -X github.com/cosmos/cosmos-sdk/version.Version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format='%H') -X github.com/cosmos/cosmos-sdk/version.Commit=$(git log -1 --format='%H')" -o arda-pocd ./cmd/arda-pocd

echo "Build completed successfully!"
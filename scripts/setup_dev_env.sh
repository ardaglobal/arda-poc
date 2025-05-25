#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'
# These versions can be overridden by setting environment variables when running the script:
# IGNITE_VERSION=v28.11.0 GO_VERSION=1.24.1 ./scripts/setup_dev_env.sh

IGNITE_VERSION="${IGNITE_VERSION:-v28.10.0}"
GO_VERSION="${GO_VERSION:-1.24.0}"

# Determine repository root and switch to it
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}/.."
cd "$REPO_ROOT"

check_cmd() {
    command -v "$1" >/dev/null 2>&1
}

install_go() {
    echo -e "${BLUE}Installing Go ${GO_VERSION}${NC}"
    curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    export PATH=/usr/local/go/bin:$PATH
}

install_docker() {
    echo -e "${BLUE}Installing Docker${NC}"
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sudo sh /tmp/get-docker.sh
    rm /tmp/get-docker.sh
}

install_ignite() {
    echo -e "${BLUE}Installing Ignite CLI ${IGNITE_VERSION}${NC}"
    curl https://get.ignite.com/cli@${IGNITE_VERSION}! | bash
}

# Ensure required tools
if ! check_cmd go; then
    install_go
else
    echo -e "${GREEN}Go detected:${NC} $(go version)"
fi

if ! check_cmd docker; then
    install_docker
else
    echo -e "${GREEN}Docker detected:${NC} $(docker --version)"
fi

if ! check_cmd ignite; then
    install_ignite
else
    echo -e "${GREEN}Ignite detected:${NC} $(ignite version)"
fi

# Install additional Go tools and proto deps via Makefile

echo -e "${GREEN}Development environment setup complete!${NC}"

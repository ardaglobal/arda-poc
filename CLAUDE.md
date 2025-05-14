# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Arda is a blockchain application built using the Cosmos SDK and CometBFT (formerly Tendermint) consensus engine. It was created with Ignite CLI as a foundation for a sovereign blockchain with inter-blockchain communication capabilities.

## Development Commands

### Build
```bash
# Install the application binary
make install

# Verify dependencies haven't been modified
go mod verify
```

### Run Development Environment
```bash
# Start a development blockchain
make dev
# OR
ignite chain serve --home ./.arda_data
```

### Testing
```bash
# Run all tests
make test

# Run just unit tests
make test-unit

# Run tests with race condition detection
make test-race

# Run tests with coverage reporting
make test-cover

# Run benchmarks
make bench
```

### Linting and Code Quality
```bash
# Run linter
make lint

# Fix linting issues automatically
make lint-fix

# Run go vet for common errors
make govet

# Run vulnerability checks
make govulncheck
```

### Protobuf Generation
```bash
# Install protobuf dependencies
make proto-deps

# Generate protobuf files
make proto-gen
# OR
ignite generate proto-go --yes
```

## Project Architecture

### Core Components

1. **App Module (`app/`)**: Contains the main application definition and configuration:
   - `app.go`: Main application struct definition and initialization
   - `app_config.go`: Application configuration
   - `genesis.go`: Genesis state handling
   - `export.go`: Chain export functionality
   - `ibc.go`: Inter-Blockchain Communication setup

2. **Command Line Interface (`cmd/ardad/`)**: Entry point for the blockchain node:
   - `main.go`: Application entry point
   - `cmd/`: Command definitions

3. **Configuration**: 
   - `config.yml`: Chain configuration including accounts, validators, and client settings

4. **Module Structure**:
   The application follows the Cosmos SDK modular architecture with these standard modules:
   - Auth: Account handling
   - Bank: Token transfers
   - Staking: Validator operations
   - Distribution: Reward distribution
   - Governance: On-chain governance
   - Params: Parameter management
   - IBC: Cross-chain communication

### Key Files

- `/app/app.go`: Core application definition with module registration and keeper setup
- `/cmd/ardad/main.go`: Entry point for the application
- `/Makefile`: Build and development commands
- `/config.yml`: Chain configuration for development

## Common Development Workflows

1. **Making Changes to the Application**:
   - Modify code in the appropriate module
   - Run `make lint` to ensure code quality
   - Run `make test` to validate changes
   - Use `make dev` to test in a local development environment

2. **Running a Local Development Network**:
   ```bash
   make dev
   ```
   This starts a blockchain node with the configuration defined in `config.yml`

3. **Interacting with the Blockchain**:
   After building the application with `make install`, you can use:
   ```bash
   # Query blockchain state
   ardad query [module] [command]
   
   # Submit transactions
   ardad tx [module] [command]
   
   # Manage keys
   ardad keys [add|list|show|delete]
   ```

4. **Generating Protobuf Files**:
   When modifying proto definitions, regenerate code with:
   ```bash
   make proto-gen
   ```
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
APPNAME := arda-poc

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# Update the ldflags with the app, client & server names
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME)d \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

##############
###  Test  ###
##############

test-unit:
	@echo Running unit tests...
	@go test -mod=readonly -v -timeout 30m ./...

test-race:
	@echo Running unit tests with race condition reporting...
	@go test -mod=readonly -v -race -timeout 30m ./...

test-cover:
	@echo Running unit tests and creating coverage report...
	@go test -mod=readonly -v -timeout 30m -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVER_FILE) -o $(COVER_HTML_FILE)
	@rm $(COVER_FILE)

bench:
	@echo Running unit tests with benchmarking...
	@go test -mod=readonly -v -timeout 30m -bench=. ./...

test: govet govulncheck test-unit

.PHONY: test test-unit test-race test-cover bench

#################
###  Install  ###
#################

all: install

install:
	@echo "--> ensure dependencies have not been modified"
	@go mod verify
	@echo "--> installing $(APPNAME)d"
	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/$(APPNAME)d

.PHONY: all install

##################
###  Protobuf  ###
##################

# Use this target if you do not want to use Ignite for generating proto files
GOLANG_PROTOBUF_VERSION=1.28.1
GRPC_GATEWAY_VERSION=1.16.0
GRPC_GATEWAY_PROTOC_GEN_OPENAPIV2_VERSION=2.20.0

proto-deps:
	@echo "Installing proto deps"
	@go install github.com/bufbuild/buf/cmd/buf@v1.50.0
	@go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest
	@go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(GOLANG_PROTOBUF_VERSION)
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v$(GRPC_GATEWAY_VERSION)
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v$(GRPC_GATEWAY_PROTOC_GEN_OPENAPIV2_VERSION)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto-gen:
	@echo "Generating protobuf files..."
	@ignite generate proto-go --yes

.PHONY: proto-deps proto-gen

#################
###  Linting  ###
#################

golangci_lint_cmd=golangci-lint
golangci_version=v1.61.0

lint: lint-fix govet govulncheck
	@echo "--> Running linter"
	@$(golangci_lint_cmd) run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@$(golangci_lint_cmd) run ./... --fix --timeout 15m

.PHONY: lint lint-fix

###################
### Setup Dev   ###
###################

# setup-dev installs tooling required for local development including
# protobuf generators, Ignite CLI and linters.

ignite-version=v28.10.0
go-version=$(shell go list -m -f '{{.GoVersion}}')

setup-script:
	@echo "--> Making setup script executable"
	@chmod +x scripts/setup_dev_env.sh
	@echo "--> Running setup script"
	@IGNITE_VERSION=$(ignite-version) GO_VERSION=$(go-version) ./scripts/setup_dev_env.sh
.PHONY: setup-script


setup-dev: setup-script proto-deps
	@echo "--> Running go mod tidy and go mod download"
	@go mod tidy
	@go mod download
	@echo "--> Installing lint and security tools"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest
.PHONY: setup-dev

###################
### Development ###
###################

dev:
	@echo "--> Running dev"
	@ignite chain serve 
.PHONY: dev

dev-sidecar:
	@echo "--> Running dev-sidecar with Air hot reload"
	@air -c .air.toml
.PHONY: dev-sidecar

sidecar-docs:
	@echo "--> Generating sidecar OpenAPI docs"
	@swag init --dir cmd/tx-sidecar --output cmd/tx-sidecar/docs
.PHONY: sidecar-docs

govet:
	@echo Running go vet...
	@go vet ./...

govulncheck:
	@echo Running govulncheck...
	@govulncheck ./...

.PHONY: govet govulncheck

clean:
	rm -rf ~/.$(APPNAME)
	rm -rf cmd/tx-sidecar/local_data
.PHONY: clean

# prod assumes the chain binary has been built and initialized. For example you might run `make dev` first while testing a feature, and then make prod to do integration testing with the sidecar.
prod:
	cp config.toml ~/.$(APPNAME)/config/config.toml
	$(APPNAME)d start
.PHONY: prod

################
###  Docker  ###
################

GHCR_NAMESPACE ?= ardaglobal
IMAGE_TAG ?= $(VERSION)
MAIN_IMAGE_NAME ?= arda-poc
SIDECAR_IMAGE_NAME ?= arda-poc-tx-sidecar

# Individual build commands
docker-build-main:
	@echo "--> Building $(MAIN_IMAGE_NAME) docker image"
	@docker build -t $(MAIN_IMAGE_NAME) -f Dockerfile .
.PHONY: docker-build-main

docker-build-tx-sidecar:
	@echo "--> Building $(SIDECAR_IMAGE_NAME) docker image"
	@docker build -t $(SIDECAR_IMAGE_NAME) -f cmd/tx-sidecar/Dockerfile .
.PHONY: docker-build-tx-sidecar

# Push to ghcr.io
# Can use the following to login: $ echo $CR_PAT | docker login ghcr.io -u USERNAME --password-stdin
docker-push-main: docker-build-main
	@echo "--> Pushing $(MAIN_IMAGE_NAME):$(IMAGE_TAG) to ghcr.io"
	@docker tag $(MAIN_IMAGE_NAME) ghcr.io/$(GHCR_NAMESPACE)/$(MAIN_IMAGE_NAME):$(IMAGE_TAG)
	@docker tag $(MAIN_IMAGE_NAME) ghcr.io/$(GHCR_NAMESPACE)/$(MAIN_IMAGE_NAME):latest
	@docker push ghcr.io/$(GHCR_NAMESPACE)/$(MAIN_IMAGE_NAME):$(IMAGE_TAG)
	@docker push ghcr.io/$(GHCR_NAMESPACE)/$(MAIN_IMAGE_NAME):latest
.PHONY: docker-push-main

docker-push-tx-sidecar: docker-build-tx-sidecar
	@echo "--> Pushing $(SIDECAR_IMAGE_NAME):$(IMAGE_TAG) to ghcr.io"
	@docker tag $(SIDECAR_IMAGE_NAME) ghcr.io/$(GHCR_NAMESPACE)/$(SIDECAR_IMAGE_NAME):$(IMAGE_TAG)
	@docker tag $(SIDECAR_IMAGE_NAME) ghcr.io/$(GHCR_NAMESPACE)/$(SIDECAR_IMAGE_NAME):latest
	@docker push ghcr.io/$(GHCR_NAMESPACE)/$(SIDECAR_IMAGE_NAME):$(IMAGE_TAG)
	@docker push ghcr.io/$(GHCR_NAMESPACE)/$(SIDECAR_IMAGE_NAME):latest
.PHONY: docker-push-tx-sidecar

# Docker Compose commands (for published images by default)
dc-up:
	@echo "--> Starting docker-compose services with published images"
	@docker-compose up -d
.PHONY: dc-up

dc-down:
	@echo "--> Stopping docker-compose services"
	@docker-compose down
.PHONY: dc-down

dc-logs:
	@echo "--> Tailing logs for all services"
	@docker-compose logs -f
.PHONY: dc-logs

# Docker Compose commands for local development (building from source)
COMPOSE_DEV_FILES := -f docker-compose.yml -f docker-compose.dev.yml

dc-build-dev:
	@echo "--> Building local docker images for development"
	@docker-compose $(COMPOSE_DEV_FILES) build
.PHONY: dc-build-dev

dc-up-dev:
	@echo "--> Starting docker-compose services for local development"
	@docker-compose $(COMPOSE_DEV_FILES) up -d
.PHONY: dc-up-dev

dc-down-dev:
	@echo "--> Stopping docker-compose services for local development"
	@docker-compose $(COMPOSE_DEV_FILES) down
.PHONY: dc-down-dev

dc-logs-dev:
	@echo "--> Tailing logs for local development services"
	@docker-compose $(COMPOSE_DEV_FILES) logs -f
.PHONY: dc-logs-dev

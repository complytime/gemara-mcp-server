# SPDX-License-Identifier: Apache-2.0
.PHONY: help build build-binary install run compose-up compose-down test

# Container image name (using localhost prefix for Podman)
IMAGE_NAME := gemara-mcp-server
IMAGE_TAG := latest

# Installation directory (user's local bin, which should be in PATH)
INSTALL_DIR := $(HOME)/.local/bin
BINARY_NAME := gemara-mcp-server

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binary
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/gemara-mcp-server

install: build ## Build and install the binary to $(INSTALL_DIR)
	@mkdir -p $(INSTALL_DIR)
	@cp bin/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"


container-build: ## Build container image
	podman build -t $(IMAGE_NAME):$(IMAGE_TAG) -f Containerfile .

container-run: container-build ## Run container with StreamableHTTP transport and verbose logging
	@mkdir -p artifacts
	@chmod 755 artifacts
	podman run --rm --userns=keep-id -p 127.0.0.1:8080:8080 \
        -v "$(PWD)/artifacts:/app/artifacts:z" \
        --user $(shell id -u):$(shell id -g) \
		$(IMAGE_NAME):$(IMAGE_TAG) ./gemara-mcp-server --transport=streamable-http --host=0.0.0.0 --port=8080


container-run-readonly: container-build ## Run container with read-only artifacts (query-only mode, cannot store new artifacts)
	@mkdir -p artifacts
	podman run --rm --userns=keep-id -p 127.0.0.1:8080:8080 \
        -v "$(PWD)/artifacts:/app/artifacts:z,ro" \
        --user $(shell id -u):$(shell id -g) \
		$(IMAGE_NAME):$(IMAGE_TAG) ./gemara-mcp-server --transport=streamable-http --host=0.0.0.0 --port=8080 --debug

container-clean: ## Clean container image
	podman rmi $(IMAGE_NAME):$(IMAGE_TAG)


test: ## Run tests
	go test ./...

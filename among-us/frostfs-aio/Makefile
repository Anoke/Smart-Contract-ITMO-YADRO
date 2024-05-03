#!/usr/bin/make -f

include .env
include help.mk

# Common variables
REPO=$(notdir $(shell pwd))
VERSION ?= "$(shell git describe --tags --match "v*" 2>/dev/null || git rev-parse --short HEAD | sed 's/^v//')"

# Variables for docker
AIO_IMAGE ?= "truecloudlab/frostfs-aio"

# Variables for compose
COMPOSE_CMD_UP ?= docker-compose up -d
COMPOSE_CMD_DOWN ?= docker-compose down

COMPOSE_V2 = "$(shell docker compose version --short | grep -q '^2' && echo true)"
ifeq ($(COMPOSE_V2), "true")
	COMPOSE_CMD_UP = docker compose up -d --wait
	COMPOSE_CMD_DOWN = docker compose down
endif

# Variables for S3
S3_BEARER_RULES ?= "/config/bearer-rules.json"
S3_GATE_PUBLIC_KEY ?= "0312fe65b101565de74eedf477afb43417ff5f795732506cfddc8e044c5a030d76"

# Build aio Docker image
image-aio:
	@echo "⇒ Build aio docker image"
	@docker build \
		--rm \
		-f Dockerfile \
		-t $(AIO_IMAGE):$(AIO_VERSION) .

# Build aio Docker image from all local pre-built binaries
image-aio-local:
	@echo "⇒ Build aio docker image from all local pre-built binaries"
	@docker build \
		--rm \
		-f Dockerfile.local \
		-t $(AIO_IMAGE):$(AIO_VERSION) .

# Build aio Docker image with custom binaries
image-aio-custom:
	@echo "⇒ Build aio Docker image with custom binaries"
	@docker build \
		--rm \
		-f Dockerfile.custom \
		-t $(AIO_IMAGE):$(AIO_VERSION) .

# Start AIO
up:
	@$(COMPOSE_CMD_UP)

# Stop AIO
down:
	@$(COMPOSE_CMD_DOWN)

# Clean up
clean: down
	@$(COMPOSE_CMD_DOWN) --volumes

# Generate S3 credentials
s3cred:
	@docker exec aio /usr/bin/frostfs-s3-authmate issue-secret \
		--wallet /config/user-wallet.json \
		--peer localhost:8080 \
		--gate-public-key $(S3_GATE_PUBLIC_KEY) \
		--container-placement-policy "REP 1"

# Generate S3 credentials based on imported wallets
s3cred-custom:
	@docker exec aio /usr/bin/frostfs-s3-authmate issue-secret \
		--wallet /wallets/$(wallet) \
		--peer localhost:8080 \
		--gate-public-key $(S3_GATE_PUBLIC_KEY) \
		--container-placement-policy "REP 1"

enable-oracle:
	@docker exec aio neo-go contract invokefunction -r http://localhost:30333 \
		--wallet-config /config/node-config.yaml \
		--out /config/tx.json 49cf4e5378ffcd4dec034fd98a174c5491e395e2 designateAsRole \
		int:8 [ bytes:02b3622bf4017bdfe317c58aed5f4c753f206b7db896046fa7d774bbc4bf7f8dc2 ] \
		-- NfgHwwTi3wHAS8aFAN243C5vGbkYDpqLHP
	@docker exec aio neo-go wallet sign --in /config/tx.json \
		-r http://localhost:30333 \
		--wallet-config /config/node-config.yaml \
		-a NfgHwwTi3wHAS8aFAN243C5vGbkYDpqLHP

# Tick new epoch in side chain
tick.epoch:
	@docker exec aio /usr/bin/frostfs-adm --config /config/frostfs-adm.yml morph force-new-epoch

# Show current version
version:
	@echo $(VERSION)
	@echo "frostfs-node: $(AIO_VERSION)"
	@echo "neo-go: $(NEOGO_VERSION)"

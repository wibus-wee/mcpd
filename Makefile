GO ?= go
CONFIG ?= docs/catalog.example.yaml

.PHONY: dev

build:
	$(GO) build ./...

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

# Docker Compose development environment
dev:
	docker compose up -d

down:
	docker compose down

reload:
	docker compose restart dev
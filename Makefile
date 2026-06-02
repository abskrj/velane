.PHONY: up down build logs seed tidy copy-platform-libs test-platform-libs help

## tidy: run go mod tidy to generate/update go.sum (required before first build)
tidy:
	cd services/control-plane && go mod tidy

## up: start all services using pre-built GHCR images
##     Run "make tidy" first to ensure go.sum is up to date.
up:
	docker compose up -d

## dev: build images and start all services in detached mode for local development
dev:
	docker compose -f docker-compose.dev.yml up --build -d

## down: stop and remove containers, keeping volumes (data survives)
down:
	docker compose down
	docker compose -f docker-compose.dev.yml down

## down-clean: stop containers AND wipe all volumes (full reset — re-runs Nango setup)
down-clean:
	docker compose down -v
	docker compose -f docker-compose.dev.yml down -v

## logs: stream control-plane logs
logs:
	docker compose logs -f control-plane

## copy-platform-libs: sync platform-libraries/ into the embedded files directory
copy-platform-libs:
	rm -rf services/control-plane/internal/platformlibs/files/bun \
	       services/control-plane/internal/platformlibs/files/python
	cp -r platform-libraries/bun  services/control-plane/internal/platformlibs/files/
	cp -r platform-libraries/python services/control-plane/internal/platformlibs/files/

## build: compile the control-plane binary locally (requires Go 1.22+ and copy-platform-libs)
build: copy-platform-libs
	cd services/control-plane && go build ./...

## test-platform-libs: run unit tests for all platform libraries (no Salesforce credentials needed)
##   Requires: bun (https://bun.sh) and python3 with pytest (pip install pytest)
test-platform-libs:
	@echo "--- Bun tests ---"
	cd platform-libraries/bun && bun test
	@echo "--- Python tests ---"
	cd platform-libraries/python && python3 -m pytest -v

## setup-nango: one-time Nango account setup. Run once after make up on a fresh stack.
##              Prints NANGO_SECRET_KEY and NANGO_PUBLIC_KEY — store in your secrets manager.
##              Idempotent: safe to re-run, returns existing keys if account already exists.
setup-nango:
	@bash scripts/setup-nango.sh

## seed: create a demo tenant (no auth required — bootstrap endpoint)
##
## After seeding, create an API key with:
##   curl -s -X POST http://localhost:8080/v1/tenants/demo/api-keys \
##     -H "Authorization: Bearer <admin-key>" \
##     -H "Content-Type: application/json" \
##     -d '{"name":"default","scopes":["admin","manage","invoke"]}' | jq .
seed:
	@echo "--- Creating demo tenant ---"
	@curl -sf -X POST http://localhost:8080/v1/tenants \
	  -H "Content-Type: application/json" \
	  -d '{"name":"Demo Org","slug":"demo"}' | jq .

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //'

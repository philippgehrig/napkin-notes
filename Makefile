.PHONY: test test-api test-web test-fontforge test-e2e test-all dev

test: test-api test-web test-fontforge

test-all: test test-e2e

test-api:
	cd napkin-notes/services/api && go test ./...

test-web:
	cd napkin-notes/web && npm run test

test-fontforge:
	cd napkin-notes/services/fontforge && python3 -m pytest tests/ -v

test-e2e:
	cd napkin-notes/e2e && npx playwright test

dev:
	docker compose -f napkin-notes/docker/docker-compose.dev.yml up --build

dev-down:
	docker compose -f napkin-notes/docker/docker-compose.dev.yml down

prod:
	docker compose -f napkin-notes/docker/docker-compose.yml up --build -d

prod-down:
	docker compose -f napkin-notes/docker/docker-compose.yml down

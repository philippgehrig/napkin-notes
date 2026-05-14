.PHONY: test test-api test-web test-fontforge dev

test: test-api test-web test-fontforge

test-api:
	cd services/api && go test ./...

test-web:
	cd web && npm run test

test-fontforge:
	cd services/fontforge && python3 -m pytest tests/ -v

dev:
	docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up --build

dev-down:
	docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml down

prod:
	docker compose -f docker/docker-compose.yml up --build -d

prod-down:
	docker compose -f docker/docker-compose.yml down

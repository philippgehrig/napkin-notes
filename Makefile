.PHONY: test test-api test-web test-fontforge dev

test: test-api test-web test-fontforge

test-api:
	cd services/api && go test ./...

test-web:
	cd web && npm run test

test-fontforge:
	cd services/fontforge && python3 -m pytest tests/ -v

dev:
	@echo "Starting development services..."
	@echo "API:       cd services/api && go run main.go"
	@echo "Web:       cd web && npm run dev"
	@echo "Fontforge: cd services/fontforge && python3 worker.py"

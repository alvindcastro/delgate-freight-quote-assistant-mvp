BACKEND_DIR=backend
FRONTEND_DIR=frontend

.PHONY: backend frontend test docker-build docker-up docker-down docker-logs clean

backend:
	cd $(BACKEND_DIR) && go run ./cmd/server

frontend:
	cd $(FRONTEND_DIR) && npm run dev

test:
	cd $(BACKEND_DIR) && go test ./...
	cd $(FRONTEND_DIR) && npm run build

docker-build:
	docker compose build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

clean:
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules $(BACKEND_DIR)/bin

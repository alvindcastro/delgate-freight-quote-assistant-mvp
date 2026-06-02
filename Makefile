BACKEND_DIR=backend
FRONTEND_DIR=frontend

.PHONY: backend frontend test clean

backend:
	cd $(BACKEND_DIR) && go run ./cmd/server

frontend:
	cd $(FRONTEND_DIR) && npm run dev

test:
	cd $(BACKEND_DIR) && go test ./...
	cd $(FRONTEND_DIR) && npm run build

clean:
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules $(BACKEND_DIR)/bin

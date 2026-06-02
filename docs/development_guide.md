# Development Guide

This guide covers the local development path for the DelGate Freight Quote
Assistant MVP.

## Ports

The backend default port is `8181`. This avoids common local collisions with
other development services.

Default local URLs:

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8181`
- Health check: `http://localhost:8181/api/health`

## Local Development

Start the backend in one terminal:

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

Start the frontend in another terminal:

```bash
cd frontend
cp .env.example .env
npm ci
npm run dev
```

For local Vite development, keep `VITE_API_URL` blank. The Vite dev server
proxies same-origin `/api` calls to `http://localhost:8181`.

If `8181` is unavailable, run both sides with matching ports:

```bash
cd backend
PORT=8282 go run ./cmd/server
```

```bash
cd frontend
BACKEND_PORT=8282 npm run dev
```

## Docker Development

Build and recreate the full stack:

```bash
docker compose up --build -d
```

View container status:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs --tail 80 backend
docker compose logs --tail 80 frontend
```

Stop the stack:

```bash
docker compose down
```

In Docker Compose, the backend is internal on `8181/tcp`. The frontend nginx
container publishes `http://localhost:5173` and proxies `/api/*` to
`http://backend:8181`.

The Compose backend requires an API token because it binds to `0.0.0.0`.
Compose defaults to local-only `API_TOKEN=dev-token`, and nginx injects that
token server-side for proxied browser requests. Replace `dev-token` before
exposing or publishing a backend port. Do not put production secrets in
`VITE_API_TOKEN`; Vite variables are compiled into the browser bundle.

## Smoke Tests

Local backend:

```bash
curl http://localhost:8181/api/health
```

Parser:

```bash
curl -X POST http://localhost:8181/api/parse-request \
  -H "Content-Type: application/json" \
  -d '{"text":"Need to ship 2 pallets from Vancouver to Calgary. Each pallet is 48x40x60 and 350 lbs. Customer needs liftgate and residential delivery before Friday."}'
```

Quote:

```bash
curl -X POST http://localhost:8181/api/quote \
  -H "Content-Type: application/json" \
  -d '{
    "customerName": "Daniel",
    "customerEmail": "ops@example.com",
    "origin": {"city":"Vancouver","province":"BC","postalCode":"V6B 1A1"},
    "destination": {"city":"Calgary","province":"AB","postalCode":"T2P 1J9"},
    "shipmentType": "ltl_pallet",
    "pieces": 2,
    "pallets": 2,
    "weightLbs": 700,
    "lengthIn": 48,
    "widthIn": 40,
    "heightIn": 60,
    "serviceLevel": "standard",
    "accessorials": ["liftgate", "residential"],
    "notes": "Customer prefers delivery before Friday."
  }'
```

Docker frontend proxy:

```bash
curl http://localhost:5173/api/health
```

```bash
curl -X POST http://localhost:5173/api/parse-request \
  -H "Content-Type: application/json" \
  -d '{"text":"Need to ship 2 pallets from Vancouver to Calgary."}'
```

For more curl examples, including direct Docker requests with `X-API-Token`,
see `docs/api_examples.md`.

## Quality Gate

Run everything from the repository root:

```bash
make test
```

Equivalent commands:

```bash
cd backend && go test ./...
cd frontend && npm run build
```

## Implementation Notes

- The quote engine is deterministic Go code in `backend/internal/quote`.
- The assistant layer is optional. Without `OPENAI_API_KEY`, it uses the local
  fallback in `backend/internal/assistant`.
- Quote history is in-memory only and resets when the backend restarts.
- The parser is rule-based and intended for assisted intake, not guaranteed
  extraction.
- Browser API calls are same-origin by default. Local Vite and Docker nginx
  handle proxying to the backend.
- Heroku uses a single web dyno image from `Dockerfile.heroku`; it does not run
  `docker-compose.yml`.

See `docs/operations_notes.md` for runtime details that are useful during
debugging and handoff.

## Environment Variables

Backend:

- `PORT`: backend port, default `8181`.
- `BACKEND_BIND_ADDR`: default `127.0.0.1`; set `0.0.0.0` for containers or
  deployed environments.
- `FRONTEND_ORIGIN`: CORS origin for local browser requests.
- `API_TOKEN`: protects quote, history, and parser routes when set.
- `ALLOW_PUBLIC_API`: permits public bind without `API_TOKEN` for the Heroku
  demo path.
- `OPENAI_API_KEY`: optional key for live AI summaries.
- `OPENAI_MODEL`: defaults to `gpt-4o-mini`.
- `QUOTE_HISTORY_LIMIT`: max in-memory quote history entries.
- `MAX_REQUEST_BODY_BYTES`: JSON body size limit.

Frontend:

- `VITE_API_URL`: leave blank for same-origin `/api` calls.
- `VITE_API_TOKEN`: only for local private demos. Do not use it for public
  browser deployments.
- `BACKEND_PORT`: Vite dev server proxy target port, default `8181`.

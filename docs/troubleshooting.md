# Troubleshooting

## Red Alert Says `Request failed`

This usually means the frontend received a non-JSON error response from the
wrong server or proxy. The Go API returns JSON errors like
`{"error":"missing or invalid API token"}`, so the generic text points to a
proxy or port problem.

Check the backend directly:

```bash
curl -i http://localhost:8181/api/health
```

Check the frontend proxy:

```bash
curl -i http://localhost:5173/api/health
```

If `localhost:8181` is not this app, restart the backend or choose another port
and set `BACKEND_PORT` for Vite.

## Backend Port Collision

The backend defaults to `8181`. If it is occupied:

```bash
cd backend
PORT=8282 go run ./cmd/server
```

```bash
cd frontend
BACKEND_PORT=8282 npm run dev
```

Keep `VITE_API_URL` blank so Vite proxies same-origin `/api` requests.

## Docker Uses Old Config

Docker images and containers can keep old baked files until they are rebuilt and
recreated. Rebuild the full stack:

```bash
docker compose down
docker compose up --build -d
```

Confirm the backend is using `8181`:

```bash
docker compose ps
docker compose logs --tail 30 backend
```

Expected log:

```text
DelGate Freight Quote Assistant API listening on http://0.0.0.0:8181
```

## Docker Frontend Works but Direct Backend Curl Fails

In Compose, the backend is not published to the host by default. Use the
frontend nginx proxy:

```bash
curl http://localhost:5173/api/health
```

For direct backend testing, publish the backend port explicitly:

```bash
docker compose run --rm -p 8181:8181 backend
```

Because the Docker backend has local-only `API_TOKEN=dev-token`, direct POST
requests must include a token. Replace `dev-token` before exposing or
publishing a backend port:

```bash
curl -X POST http://localhost:8181/api/parse-request \
  -H "Content-Type: application/json" \
  -H "X-API-Token: dev-token" \
  -d '{"text":"Need to ship 2 pallets from Vancouver to Calgary."}'
```

## API Token Errors

If the alert says `missing or invalid API token`, the request reached the Go API
but did not include the configured token.

For local backend development, leave `API_TOKEN` empty unless you are testing
auth behavior.

For Docker Compose, keep the same token on both services. `dev-token` is for
local demos only:

```bash
API_TOKEN=dev-token docker compose up --build -d
```

The nginx proxy injects `X-API-Token` server-side, so normal browser requests
through `http://localhost:5173` do not need a browser-visible token.

## Vite Proxy Misses the Backend

Symptoms:

- `http://localhost:5173` loads but parser and quote actions fail.
- `curl http://localhost:8181/api/health` fails.
- `curl http://localhost:5173/api/health` returns plain text or an HTML error.

Fix:

```bash
cd backend
go run ./cmd/server
```

If the backend runs on a custom port:

```bash
cd frontend
BACKEND_PORT=8282 npm run dev
```

## Heroku Port Behavior

Heroku sets `PORT` dynamically. The app default is `8181` for local and Docker
development, but Heroku should still use Heroku's provided `PORT`.

Do not hard-code a Heroku public port in docs, curl examples, or browser code.
Use the app URL:

```bash
curl https://your-delgate-demo-name.herokuapp.com/api/health
```

## OpenAI Is Not Configured

The app still works. The backend logs `OpenAI enabled: false` and uses the local
fallback assistant for summaries and customer messages.

To enable live AI summaries:

```bash
OPENAI_API_KEY=your_api_key_here go run ./cmd/server
```

## Quote History Disappeared

Quote history is in memory. It resets when the backend process or container
restarts. This is expected for the MVP.

## Browser Shows an Old Bundle

If a rebuilt frontend still behaves like the old version:

```bash
docker compose down
docker compose up --build -d
```

Then refresh the browser. If needed, hard refresh or clear site data for
`localhost:5173`.

## Quick Known-Good Path

```bash
docker compose down
docker compose up --build -d
docker compose ps
curl http://localhost:5173/api/health
```

Open `http://localhost:5173`, click **Load demo**, then click **Generate quote**
and **Parse request**.

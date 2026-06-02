# Operations Notes

These are useful runtime details for handoff, debugging, and demos.

## Request Routing

The frontend calls `/api/*` by default.

Local Vite:

- Browser loads `http://localhost:5173`.
- Vite proxies `/api` to `http://localhost:8181`.
- `BACKEND_PORT` can override the proxy target.

Docker Compose:

- Browser loads `http://localhost:5173`.
- The frontend nginx container proxies `/api/*` to `http://backend:8181`.
- The backend is not published to the host by default.

Heroku:

- One web dyno serves both `/api/*` and the built frontend.
- Heroku supplies `PORT`; the local fallback port does not matter there.
- Same-origin API calls keep `VITE_API_URL` blank.

## API Tokens

The backend protects quote, quote history, and parser routes when `API_TOKEN` is
set. Health stays public.

Docker Compose sets local-only `API_TOKEN=dev-token` because the backend binds
to `0.0.0.0`. Nginx injects `X-API-Token` server-side, so the browser bundle
does not need a token. Replace `dev-token` before exposing or publishing a
backend port.

Do not put production secrets in `VITE_API_TOKEN`. Vite exposes those variables
in the compiled JavaScript.

## AI Behavior

`OPENAI_API_KEY` is optional. Without it, the backend returns local fallback
summaries and customer messages. This keeps local and evaluator demos reliable.

If the OpenAI call fails, the assistant falls back locally instead of blocking
quote creation.

## Data Persistence

Quote history is in memory only. It resets when the backend process or container
restarts. No database is required for this MVP.

## Health Checks

Backend health:

```bash
curl http://localhost:8181/api/health
```

Docker frontend proxy health:

```bash
curl http://localhost:5173/api/health
```

Compose healthchecks run inside the backend container against
`http://127.0.0.1:8181/api/health`.

## Request Limits And Timeouts

The backend limits request body size with `MAX_REQUEST_BODY_BYTES`, default
`1048576`.

HTTP timeout defaults:

- `HTTP_READ_TIMEOUT=10s`
- `HTTP_READ_HEADER_TIMEOUT=5s`
- `HTTP_WRITE_TIMEOUT=15s`
- `HTTP_IDLE_TIMEOUT=60s`

The assistant call uses its own shorter timeout and falls back locally on
failure.

## Demo Safety

This is a demo estimate, not a binding freight quote. Production work would
need carrier integrations, authenticated users, persistent storage, audit logs,
admin rate management, and quote approval workflow.

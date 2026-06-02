# DelGate Freight Quote Assistant MVP

A simple AI-powered freight quote assistant built for the DelGate Logistics practical assessment.

The project demonstrates a practical logistics workflow: collect shipment details, calculate a deterministic demo quote, flag missing information, and generate an AI-assisted customer-facing explanation.

## Why this approach

Freight quotes should not be invented by an LLM. This MVP keeps the quote calculation deterministic and explainable in the Go backend, then uses the AI layer for work that LLMs are better suited for:

- Summarizing the quote in plain English
- Identifying missing information
- Producing a customer-ready response
- Flagging shipments that should be manually reviewed
- Helping operations staff move faster without losing control of the quote logic

If no `OPENAI_API_KEY` is provided, the app still works with a local rule-based assistant fallback. This makes the demo reliable for reviewers who run it locally.

## Features

- Go REST API with no third-party backend dependencies
- React + Vite frontend
- Structured freight quote form
- Demo pricing engine with lane, weight, dimensional weight, service level, accessorials, and fuel surcharge
- Manual-review flags for missing information, oversize freight, same-day cross-province delivery, and sensitive/hazmat notes
- Optional OpenAI-powered quote explanation
- Rule-based fallback assistant when no API key is configured
- Messy text parser for freight requests pasted from emails or notes
- Quote history in memory
- Clean README, env examples, and Loom walkthrough script

## Architecture

```text
frontend/ React + Vite
    | REST API calls
backend/ Go net/http API
    | deterministic quote engine
    | optional OpenAI assistant
    | local fallback assistant
    | in-memory quote store
```

## Repository structure

```text
delgate-freight-quote-assistant/
├── .github/
│   └── workflows/ci-cd.yml
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/api/handlers.go
│   ├── internal/assistant/assistant.go
│   ├── internal/parser/parser.go
│   ├── internal/quote/engine.go
│   ├── internal/quote/types.go
│   ├── internal/quote/engine_test.go
│   ├── internal/store/store.go
│   ├── Dockerfile
│   ├── go.mod
│   └── .env.example
├── frontend/
│   ├── src/App.jsx
│   ├── src/api.js
│   ├── src/main.jsx
│   ├── src/styles.css
│   ├── Dockerfile
│   ├── nginx.conf.template
│   ├── package.json
│   ├── index.html
│   └── .env.example
├── docs/
│   ├── ci_cd.md
│   ├── demo_walkthrough.md
│   ├── heroku_deployment.md
│   ├── loom_script.md
│   └── screenshots/
├── Makefile
├── Dockerfile.heroku
├── docker-compose.yml
├── heroku.yml
├── .dockerignore
├── .gitignore
└── README.md
```

## Prerequisites

- Go 1.22+
- Node.js 18+
- npm
- Docker and Docker Compose for the containerized run path

## Run locally

Open two terminals.

### 1. Start the backend

```bash
cd backend
cp .env.example .env
# Optional: edit .env and add OPENAI_API_KEY
go run ./cmd/server
```

The backend starts on `http://localhost:8080`.

Local backend startup loads `backend/.env` automatically when present. Shell
environment variables still take precedence over values in the file.

Shortcut from the repository root:

```bash
make backend
```

### 2. Start the frontend

```bash
cd frontend
cp .env.example .env
npm ci
npm run dev
```

The frontend starts on `http://localhost:5173`.

For local Vite development, `VITE_API_URL` can stay blank. The Vite dev server
proxies `/api` to `http://localhost:8080`.

Shortcut from the repository root:

```bash
make frontend
```

## Run with Docker

Docker Compose builds and runs both services:

```bash
docker compose up --build
```

Open the frontend at `http://localhost:5173`. The backend is kept internal to
the Compose network; the frontend nginx container proxies `/api` requests to it.

The compose stack sets `API_TOKEN=dev-token` by default because the backend
binds to `0.0.0.0` inside Docker. Nginx injects that token server-side when
proxying `/api`, so the browser bundle does not expose it. To use a different
token or enable live AI summaries:

```bash
API_TOKEN=your-token OPENAI_API_KEY=your-api-key docker compose up --build
```

Useful shortcuts:

```bash
make docker-build
make docker-up
make docker-down
```

For direct backend API testing, publish the backend port explicitly:

```bash
docker compose run --rm -p 8080:8080 backend
```

Because Compose sets `API_TOKEN=dev-token` by default, direct quote and parse
requests to that backend must include `X-API-Token: dev-token` or
`Authorization: Bearer dev-token`.

## Deploy to Heroku

The recommended Heroku deployment is a single web dyno that serves both the Go
API and the built React frontend. `heroku.yml` uses `Dockerfile.heroku` to build
that combined image.

For an evaluator demo, use a **Basic** dyno so the app stays awake. Use **Eco**
only if lower cost is more important than avoiding cold starts.

Quick path:

```bash
heroku login
heroku create your-delgate-demo-name
heroku stack:set container -a your-delgate-demo-name
heroku config:set ALLOW_PUBLIC_API=true BACKEND_BIND_ADDR=0.0.0.0 STATIC_DIR=/app/public OPENAI_MODEL=gpt-4o-mini -a your-delgate-demo-name
git push heroku main
heroku ps:scale web=1 -a your-delgate-demo-name
heroku ps:type basic -a your-delgate-demo-name
```

Optional live AI summaries:

```bash
heroku config:set OPENAI_API_KEY=your_api_key_here -a your-delgate-demo-name
```

See `docs/heroku_deployment.md` for the full deployment, verification, and cost
control checklist.

## CI/CD

GitHub Actions validates backend tests, frontend builds, and the Heroku
single-dyno Docker image. When `HEROKU_APP_NAME` and `HEROKU_API_KEY` secrets
are configured, pushes to `main` can deploy to Heroku automatically after CI
passes. See `docs/ci_cd.md`.

## Test and build

Run the project quality gate from the repository root:

```bash
make test
```

This runs:

```bash
cd backend && go test ./...
cd frontend && npm run build
```

You can run those commands separately while iterating on one side of the app.

## Optional OpenAI configuration

The app works without an API key. To enable live AI summaries, set:

```bash
OPENAI_API_KEY=your_api_key_here
OPENAI_MODEL=gpt-4o-mini
```

The backend will use the OpenAI API for summary generation. If the call fails, it falls back to the local assistant automatically.

By default the backend binds to `127.0.0.1`. For deployed or shared API-only
environments, set `BACKEND_BIND_ADDR` explicitly and configure `API_TOKEN`.
When `API_TOKEN` is set, quote and parse requests must send
`Authorization: Bearer <token>` or `X-API-Token`.

For the public single-app Heroku demo, leave `API_TOKEN` unset and use the
explicit `ALLOW_PUBLIC_API=true` setting from `heroku.yml`. Do not put a secret
API token into `VITE_API_TOKEN` for a public browser app; Vite variables are
compiled into the frontend bundle.

## API examples

The examples below assume the local no-token backend path from `go run
./cmd/server`. If you are testing a Docker-backed API directly, add
`-H "X-API-Token: dev-token"` to quote and parse requests.

### Health check

```bash
curl http://localhost:8080/api/health
```

### Create quote

```bash
curl -X POST http://localhost:8080/api/quote \
  -H "Content-Type: application/json" \
  -d '{
    "customerName": "Test Customer",
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

### Parse messy request text

```bash
curl -X POST http://localhost:8080/api/parse-request \
  -H "Content-Type: application/json" \
  -d '{
    "text":"Need to ship 2 pallets from Vancouver to Calgary. Each pallet is 48x40x60 and 350 lbs. Needs liftgate and residential delivery before Friday."
  }'
```

## Smoke test

With the local backend running on `http://localhost:8080`:

```bash
curl http://localhost:8080/api/health
```

```bash
curl -X POST http://localhost:8080/api/quote \
  -H "Content-Type: application/json" \
  -d '{
    "customerName": "Demo Customer",
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

```bash
curl -X POST http://localhost:8080/api/parse-request \
  -H "Content-Type: application/json" \
  -d '{
    "text":"Need to ship 2 pallets from Vancouver to Calgary. Each pallet is 48x40x60 and 350 lbs. Needs liftgate and residential delivery before Friday."
  }'
```

For a direct Docker backend test, include the default token header on POST
requests:

```bash
curl -X POST http://localhost:8080/api/parse-request \
  -H "Content-Type: application/json" \
  -H "X-API-Token: dev-token" \
  -d '{"text":"Need to ship 2 pallets from Vancouver to Calgary."}'
```

## Demo workflow

1. Open the frontend.
2. Click **Load demo**.
3. Generate a quote.
4. Review the estimated range, breakdown, confidence, manual-review flags, and AI summary.
5. Copy the customer-ready response.
6. Paste a messy customer request into the parser and show how the form can be prefilled.
7. Confirm the `From parser: N fields` chip appears in the shipment details header.

## Notes for the evaluator

This is intentionally a demo estimate, not a production freight rating engine. A production version would integrate real carrier APIs, postal-code distance logic, authenticated users, persistent quote history, admin-managed rate tables, CRM integration, PDF quote export, and audit logs.

## Future improvements

- Real carrier/rate API integration
- Postal-code distance and zone lookup
- Admin-managed pricing table
- PostgreSQL quote persistence
- Authentication and role-based access
- CRM or HubSpot integration
- PDF quote export
- Quote status workflow: draft, reviewed, sent, accepted
- Operations handoff queue
- RAG over shipping policies and SOPs
- Carrier SLA and accessorial policy citations

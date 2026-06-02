# Submission Notes

## Project

DelGate Freight Quote Assistant MVP

## Submission assets

- Loom walkthrough: add public or unlisted Loom URL before final submission.
- Live demo: https://delgate-mvp-954a8b03876e.herokuapp.com/
- Screenshots: add final screenshots under `docs/screenshots/` if needed.
- CI/CD checklist: `docs/ci_cd.md`
- API examples: `docs/api_examples.md`
- Developer guide: `docs/development_guide.md`
- Demo walkthrough script: `docs/demo_walkthrough.md`
- Heroku deployment checklist: `docs/heroku_deployment.md`
- Operations notes: `docs/operations_notes.md`
- Troubleshooting: `docs/troubleshooting.md`
- Loom recording script: keep local if it contains presenter-specific details.

## Brief explanation

I built a simple AI-powered freight quote assistant using a Go backend and React frontend.

The app accepts structured shipment details such as pickup, delivery, shipment type, weight, dimensions, service level, accessorials, and notes. The backend calculates a deterministic demo freight estimate and returns a quote range, cost breakdown, missing-field checklist, risk flags, confidence level, and next steps.

The AI layer is used for quote summarization, customer-ready communication, and operations triage. I intentionally kept the pricing engine deterministic so the LLM does not invent freight prices. If an OpenAI API key is provided, the backend can generate AI summaries through the OpenAI API. If no key is configured, the app falls back to a local rule-based assistant so the demo remains runnable.

I also added a messy request parser so a user can paste a customer email or note, such as “Need to ship 2 pallets from Vancouver to Calgary…”, and the app extracts fields into the quote form.

## Tools used

- Go standard library backend
- React + Vite frontend
- Optional OpenAI API integration
- Rule-based fallback assistant
- In-memory quote history
- AI-assisted development workflow for scaffolding, code review, and iteration

## How to run

### Fastest path with Docker

```bash
docker compose up --build
```

Open `http://localhost:5173`.

Docker keeps the backend internal to the Compose network and serves API calls
through the frontend nginx proxy. The stack uses local-only
`API_TOKEN=dev-token` by default and injects that token server-side for `/api`
requests. Replace `dev-token` before exposing or publishing a backend port.

If the stack has stale containers after a port or proxy change:

```bash
docker compose down --remove-orphans
docker compose up --build --force-recreate
```

### Backend

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

Backend URL: `http://localhost:8181`

Shortcut from the repository root:

```bash
make backend
```

### Frontend

```bash
cd frontend
cp .env.example .env
npm ci
npm run dev
```

Frontend URL: `http://localhost:5173`

Shortcut from the repository root:

```bash
make frontend
```

For Vite local development, `VITE_API_URL` can stay blank because the dev server
proxies `/api` to `http://localhost:8181`.

If `8181` is unavailable, set the same custom port on the backend and the Vite
proxy:

```bash
PORT=8282 go run ./cmd/server
BACKEND_PORT=8282 npm run dev
```

### Heroku demo deployment

This repo includes `heroku.yml` and `Dockerfile.heroku` for a single-dyno Heroku
deployment where the Go server serves both `/api/*` and the built React
frontend.

Recommended evaluator tier: **Basic**. It is always on and avoids Eco cold
starts. Use **Eco** only when lower cost matters more than first-load
reliability.

```bash
heroku login
heroku create your-delgate-demo-name
heroku stack:set container -a your-delgate-demo-name
heroku config:set ALLOW_PUBLIC_API=true BACKEND_BIND_ADDR=0.0.0.0 STATIC_DIR=/app/public OPENAI_MODEL=gpt-4o-mini -a your-delgate-demo-name
git push heroku main
heroku ps:scale web=1 -a your-delgate-demo-name
heroku ps:type basic -a your-delgate-demo-name
```

Full checklist: `docs/heroku_deployment.md`.

More local commands and troubleshooting notes:

- `docs/development_guide.md`
- `docs/api_examples.md`
- `docs/operations_notes.md`
- `docs/troubleshooting.md`

### CI/CD

GitHub Actions runs backend tests, frontend build verification, and the Heroku
single-dyno image build. If `HEROKU_APP_NAME` and `HEROKU_API_KEY` are set as
GitHub secrets, pushes to `main` deploy to Heroku after CI passes.

## Optional AI setup

The app works without an API key. To enable live AI summaries, set this in `backend/.env` or your shell. Local backend startup loads `backend/.env` automatically, while shell variables take precedence:

```bash
OPENAI_API_KEY=your_api_key_here
OPENAI_MODEL=gpt-4o-mini
```

For local/private demos only, `VITE_API_TOKEN` can send a browser-visible bearer
token. Do not use `VITE_API_TOKEN` for public or shared browser deployments
because Vite embeds it into the compiled JavaScript. For shared environments,
use a server-side proxy, cookie/session auth, or identity-provider auth.

## How to test

Run the project quality gate:

```bash
make test
```

This runs backend tests with `go test ./...` and verifies the frontend production
build with `npm run build`. The frontend currently has a build check rather than
a separate unit-test suite.

Manual test:

1. Start backend and frontend.
2. Click **Load demo**.
3. Click **Generate quote**.
4. Review the estimated range, breakdown, confidence, missing fields, risk flags, and customer-ready response.
5. Paste the sample messy request into the parser and click **Parse request**.
6. Confirm the `Shipment details` form shows the subtle parser-applied chip after parsing.

## Screenshots or demo link

Add final screenshot files under `docs/screenshots/` before submission. If a
hosted demo is available, add the URL in the **Submission assets** section above.
If no hosted demo is available, reviewers can run the Docker command above and
open `http://localhost:5173`.

## Loom walkthrough checklist

The Loom video is required for submission. Add the final public or unlisted Loom
URL in the **Submission assets** section above.

- Explain that the assessment asked for a simple AI-powered freight quote assistant.
- Show the Go backend and React frontend running locally, or show the Docker stack running.
- Explain that pricing is deterministic and AI is used for summary/triage/customer communication.
- Generate a demo Vancouver to Calgary LTL pallet quote.
- Show quote range, breakdown, confidence, status, and risk flags.
- Show the customer-ready response and copy button.
- Show the messy request parser.
- Show that parsing subtly populates the shipment details form.
- Mention future improvements: real carrier APIs, postal-code distance, database, auth, CRM integration, PDF quote export, and admin rate tables.

## Production improvements

- Real carrier/rating API integration
- Postal-code distance/zone calculation
- Persistent PostgreSQL storage
- User authentication and role-based permissions
- Admin rate table management
- CRM/HubSpot integration
- PDF quote export
- Quote approval workflow
- RAG over carrier policies and internal SOPs

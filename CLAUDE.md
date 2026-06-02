# Project Instructions for AI Agents

This file provides instructions and context for AI coding agents working on this
project.

## Build & Test

Run the full quality gate from the repository root:

```bash
make test
```

Equivalent targeted commands:

```bash
cd backend && go test ./...
cd frontend && npm run build
```

## Run Commands

Local two-terminal flow:

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

```bash
cd frontend
cp .env.example .env
npm ci
npm run dev
```

Docker flow:

```bash
docker compose up --build -d
```

## Architecture Overview

The frontend is React + Vite. Browser requests use same-origin `/api` calls.
Local Vite proxies those calls to the Go backend on `localhost:8181`; Docker
nginx proxies them to `backend:8181`.

The backend is Go `net/http`. Quote pricing is deterministic in
`backend/internal/quote`, parser behavior is rule-based in
`backend/internal/parser`, quote history is in memory, and the assistant layer
falls back locally when `OPENAI_API_KEY` is not configured.

## Conventions & Patterns

- Default backend port is `8181`.
- Keep `VITE_API_URL` blank for same-origin local and Docker runs.
- Do not expose production secrets through `VITE_API_TOKEN`; Vite variables are
  compiled into the browser bundle.
- Use non-interactive shell flags for file operations, as described in
  `AGENTS.md`.
- See `docs/development_guide.md`, `docs/api_examples.md`,
  `docs/operations_notes.md`, and `docs/troubleshooting.md` before changing
  runtime behavior.

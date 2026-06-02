# Submission Notes

## Project

DelGate Freight Quote Assistant MVP

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

### Backend

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

Backend URL: `http://localhost:8080`

### Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend URL: `http://localhost:5173`

## Optional AI setup

The app works without an API key. To enable live AI summaries, set this in `backend/.env` or your shell:

```bash
OPENAI_API_KEY=your_api_key_here
OPENAI_MODEL=gpt-4o-mini
```

## How to test

Run backend tests:

```bash
cd backend
go test ./...
```

Manual test:

1. Start backend and frontend.
2. Click **Load demo**.
3. Click **Generate quote**.
4. Review the estimated range, breakdown, confidence, missing fields, risk flags, and customer-ready response.
5. Paste the sample messy request into the parser and click **Parse request**.

## Loom walkthrough checklist

- Explain that the assessment asked for a simple AI-powered freight quote assistant.
- Show the Go backend and React frontend running locally.
- Explain that pricing is deterministic and AI is used for summary/triage/customer communication.
- Generate a demo Vancouver to Calgary LTL pallet quote.
- Show quote range, breakdown, confidence, status, and risk flags.
- Show the customer-ready response and copy button.
- Show the messy request parser.
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

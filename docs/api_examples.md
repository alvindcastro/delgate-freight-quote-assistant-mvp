# API Examples

These examples use the local backend default:

```bash
API_BASE=http://localhost:8181
```

For Docker Compose through nginx, use:

```bash
API_BASE=http://localhost:5173
```

## Health

```bash
curl "${API_BASE}/api/health"
```

Health is public and does not require an API token.

## Quote History

```bash
curl "${API_BASE}/api/quotes"
```

Quote history is stored in memory and resets when the backend restarts.

## Parse Request Text

```bash
curl -X POST "${API_BASE}/api/parse-request" \
  -H "Content-Type: application/json" \
  -d '{
    "text":"Need to ship 2 pallets from Vancouver to Calgary. Each pallet is 48x40x60 and 350 lbs. Customer needs liftgate and residential delivery before Friday."
  }'
```

## Create Quote

```bash
curl -X POST "${API_BASE}/api/quote" \
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

## Direct Docker Backend With Token

Compose keeps the backend internal by default. For direct API testing, publish
the backend port:

```bash
docker compose run --rm -p 8181:8181 backend
```

Direct Docker POST requests need the default local-only token. Replace
`dev-token` before exposing or publishing a backend port:

```bash
curl -X POST http://localhost:8181/api/parse-request \
  -H "Content-Type: application/json" \
  -H "X-API-Token: dev-token" \
  -d '{"text":"Need to ship 2 pallets from Vancouver to Calgary."}'
```

```bash
curl -X POST http://localhost:8181/api/quote \
  -H "Content-Type: application/json" \
  -H "X-API-Token: dev-token" \
  -d '{
    "customerName": "Daniel",
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

## Custom Local Backend Port

If you run the backend on a custom port, set `API_BASE` to match:

```bash
PORT=8282 go run ./cmd/server
API_BASE=http://localhost:8282
curl "${API_BASE}/api/health"
```

For the frontend dev server, set `BACKEND_PORT` to the same value:

```bash
BACKEND_PORT=8282 npm run dev
```

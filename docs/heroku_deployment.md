# Heroku Deployment

This project can deploy to Heroku as one web dyno: the Go backend serves both
the `/api/*` routes and the built React frontend from `frontend/dist`.

## Recommended tier

Use **Basic** for a submission or evaluator demo. Basic is always on and avoids
the cold-start/sleep behavior of Eco. Use **Eco** only when the lower monthly
cost matters more than first-load reliability.

Current low-cost dyno choices:

- Eco: lowest-cost Heroku dyno, sleeps after inactivity.
- Basic: low-cost always-on dyno.

No database add-on is required for this MVP because quote history is in memory.

## Deployment model

Heroku does not run the local `docker-compose.yml` stack. Instead, `heroku.yml`
uses `Dockerfile.heroku` to build one image:

1. Build the React frontend with Vite.
2. Build the Go API binary.
3. Copy `frontend/dist` into `/app/public`.
4. Run one web process: `/app/delgate-api`.

The frontend uses same-origin API calls, so `VITE_API_URL` stays blank.

## Public demo API mode

The local and Docker Compose paths can protect API routes with `API_TOKEN`.
For a public single-app Heroku demo, browser requests cannot safely carry a
secret token. The Heroku image therefore opts into:

```text
ALLOW_PUBLIC_API=true
STATIC_DIR=/app/public
BACKEND_BIND_ADDR=0.0.0.0
```

This is appropriate for a public MVP demo with in-memory data and no private
customer records. For production, add real browser authentication
cookie/session auth or an identity provider before protecting user-specific
quote data.

## First-time setup

Install and log in to the Heroku CLI, then create an app:

```bash
heroku login
heroku create your-delgate-demo-name
heroku stack:set container -a your-delgate-demo-name
```

Set the public-demo runtime config explicitly:

```bash
heroku config:set ALLOW_PUBLIC_API=true BACKEND_BIND_ADDR=0.0.0.0 STATIC_DIR=/app/public OPENAI_MODEL=gpt-4o-mini -a your-delgate-demo-name
```

Optional AI summaries:

```bash
heroku config:set OPENAI_API_KEY=your_api_key_here -a your-delgate-demo-name
```

The `heroku.yml` file also records the intended static serving and public demo
variables. Verify the app config before recording or submitting:

```bash
heroku config -a your-delgate-demo-name
```

## Deploy

Push the current branch to Heroku:

```bash
git push heroku main
```

Scale and choose the low-cost dyno type:

```bash
heroku ps:scale web=1 -a your-delgate-demo-name
heroku ps:type basic -a your-delgate-demo-name
```

For the lower-cost sleeping option:

```bash
heroku ps:type eco -a your-delgate-demo-name
```

Open the app:

```bash
heroku open -a your-delgate-demo-name
```

## Verify

```bash
heroku logs --tail -a your-delgate-demo-name
```

```bash
curl https://your-delgate-demo-name.herokuapp.com/api/health
```

Manual smoke path:

1. Open the Heroku app URL.
2. Click **Load demo**.
3. Click **Generate quote**.
4. Confirm the quote result renders.
5. Paste the messy request sample into the parser.
6. Click **Parse request** and confirm the shipment form shows the parser-applied chip.

## Cost control

To stop charges from the running web dyno:

```bash
heroku ps:scale web=0 -a your-delgate-demo-name
```

To remove the app entirely:

```bash
heroku apps:destroy -a your-delgate-demo-name --confirm your-delgate-demo-name
```

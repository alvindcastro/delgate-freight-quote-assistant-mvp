# CI/CD

GitHub Actions runs the project validation and can optionally deploy the
single-dyno Heroku app.

## Workflow

Workflow file: `.github/workflows/ci-cd.yml`

Triggers:

- Pull requests: run CI only.
- Pushes to `main`: run CI and deploy when Heroku secrets are configured.
- Manual dispatch: run CI and optionally deploy.

Jobs:

1. **Backend tests** - sets up Go from `backend/go.mod` and runs `go test ./...`.
2. **Frontend build** - installs with `npm ci` and runs `npm run build`.
3. **Heroku image build** - builds `Dockerfile.heroku` to verify the single-dyno image.
4. **Deploy to Heroku** - configures and deploys the Heroku app after all checks pass.

## Required GitHub secrets for deploy

Set these in GitHub repository settings:

- `HEROKU_APP_NAME` - the Heroku app name, for example `your-delgate-demo-name`.
- `HEROKU_API_KEY` - a Heroku API key from your Heroku account settings.

If either secret is missing, the deploy job exits cleanly after CI and prints a
skip message.

## Heroku behavior

The deploy job:

1. Ensures the app uses the `container` stack.
2. Sets the public demo config:
   - `ALLOW_PUBLIC_API=true`
   - `BACKEND_BIND_ADDR=0.0.0.0`
   - `STATIC_DIR=/app/public`
   - `OPENAI_MODEL=gpt-4o-mini`
3. Pushes the current commit to Heroku Git.
4. Scales one web dyno.
5. Sets the dyno type to `basic`.
6. Verifies `/api/health` with bounded retries while the dyno starts.

The workflow does not set `OPENAI_API_KEY`. Add it directly in Heroku config if
you want live AI summaries:

```bash
heroku config:set OPENAI_API_KEY=your_api_key_here -a your-delgate-demo-name
```

## Manual deploy

Use **Actions -> CI/CD -> Run workflow**.

Set `deploy` to:

- `true` to deploy after CI passes.
- `false` to run validation only.

## Deployment trigger check

After adding `HEROKU_APP_NAME` and `HEROKU_API_KEY` in GitHub repository
secrets, push any commit to `main` to verify deployment. The workflow does not
filter by changed paths, so a documentation-only commit is enough to run CI,
build the Heroku image, deploy to Heroku, and verify `/api/health` with
bounded retries.

Confirm the run in GitHub Actions:

1. Open **Actions -> CI/CD**.
2. Select the latest push run on `main`.
3. Confirm `Backend tests`, `Frontend build`, and `Heroku image build` pass.
4. Confirm `Deploy to Heroku` does not print the missing-secrets skip message.
5. Confirm `Verify Heroku health` succeeds.

## Cost control

The workflow chooses Heroku Basic for evaluator reliability. To switch the app
to Eco after submission:

```bash
heroku ps:type eco -a your-delgate-demo-name
```

To stop the web dyno:

```bash
heroku ps:scale web=0 -a your-delgate-demo-name
```

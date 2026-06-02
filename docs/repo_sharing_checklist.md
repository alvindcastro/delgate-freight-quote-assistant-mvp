# Repository Sharing Checklist

Use this before pushing or sharing the repository.

## Secret Check

Run a broad search for secret-like strings:

```bash
rg -n -i "(api[_-]?key|secret|token|password|private[_-]?key|credential|authorization|bearer|sk-[A-Za-z0-9])" . \
  --glob "!frontend/node_modules/**" \
  --glob "!frontend/dist/**" \
  --glob "!.git/**" \
  --glob "!.beads/**"
```

Expected matches are placeholders, demo values such as `dev-token`, test values,
and documentation telling users how to configure secrets. Real `.env` files
must not be committed.

## Tracked Files

Check what will be shared:

```bash
git status --short
git ls-files
```

Local tool state should stay untracked:

- `.agents/`
- `.claude/`
- `.codex/`
- `.beads/`
- `.dolt/`
- `.env` files

`docs/loom_script.md` is local submission prep by default. Keep it ignored if it
contains presenter-specific details or private recording notes.

## Git History

The current tree can be clean while old commits still contain removed local
state. Check history before public sharing:

```bash
git log --all --name-status -- .beads .claude .agents .codex
```

If personal workflow metadata exists in history and should not be public, share
a fresh clean repository or rewrite history before pushing.

## Large Files

Find large working-tree files outside ignored generated folders:

```bash
find . -type f -size +1M \
  -not -path "./.git/*" \
  -not -path "./frontend/node_modules/*" \
  -not -path "./frontend/dist/*" \
  -not -path "./.beads/*" \
  -printf "%s %p\n"
```

Tracked source files should stay small. Generated frontend output and
`node_modules` should not be tracked.

## Ignore Rules

Confirm these stay ignored:

```bash
git check-ignore -v frontend/node_modules frontend/dist backend/.env frontend/.env .beads .claude .agents .codex
```

## Docker Context

The Docker context excludes local state, env files, generated frontend output,
and dependencies through `.dockerignore`.

Before sharing a Docker-based repro, rebuild cleanly:

```bash
docker compose down --remove-orphans
docker compose up --build --force-recreate -d
docker compose ps
curl http://localhost:5173/api/health
```

## Docs

Make sure these entry points are current:

- `README.md`
- `SUBMISSION.md`
- `docs/development_guide.md`
- `docs/api_examples.md`
- `docs/operations_notes.md`
- `docs/troubleshooting.md`
- `docs/demo_walkthrough.md`
- `docs/heroku_deployment.md`
- `docs/ci_cd.md`

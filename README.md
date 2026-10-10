# Tanuki — development environment manager (v0.1.0)

One project. One configuration. One command to get started.

```bash
tanuki doctor
tanuki validate
tanuki up
tanuki status
tanuki logs
tanuki down
```

## Install

Requires Linux x86_64 (ARM64 works where Docker images exist), Go 1.27+ to build, Docker Engine to run services.

```bash
git clone https://github.com/RakhaYandra/tanuki
cd tanuki
make install   # builds to ~/.local/bin/tanuki (ensure ~/.local/bin is in PATH)
```

Or download a release binary from the `v0.1.0` tag (`dist/` artifacts) and place it in `~/.local/bin`.

## Quickstart

```bash
mkdir demo && cd demo
tanuki init --name pulse
tanuki doctor      # prerequisites, read-only
tanuki validate    # config check, no side effects
tanuki up          # start postgres + redis + api
tanuki status      # READY?
tanuki logs api    # application log
tanuki down        # stop everything owned; volumes kept
```

## Commands

- `tanuki init [--name X] [--dir D] [--force]` — generate `tanuki.yaml`; never overwrites without `--force`.
- `tanuki validate [--config PATH]` — exit 2 on invalid config, no side effects.
- `tanuki doctor [--config PATH]` — read-only checks; exit 0 ok, 2 config, 3 missing prerequisite, 4 port conflict.
- `tanuki up [--dry-run] [--detach] [--wait=true] [--config PATH]` — start services then app detached (PID + `app.log` in state dir). `--dry-run` prints the plan with zero mutations.
- `tanuki down` — stop app (SIGTERM→SIGKILL) then owned containers; volumes always preserved.
- `tanuki status` — READY / DEGRADED with per-resource state.
- `tanuki ps` — NAME/TYPE/STATE/PORT table of managed resources.
- `tanuki logs [postgres|redis|api] [--follow]` — service or app logs.
- `tanuki config [--format yaml|json]` — effective configuration (no secrets in schema).
- `tanuki version` — build version.

Exit codes: `0` ok, `1` general, `2` config/args, `3` prerequisite, `4` port conflict, `5` startup/health, `6` shutdown, `7` state/lock.

## Configuration

See `tanuki.example.yaml` and `docs/CONFIGURATION.md`. Strict schema: unknown fields rejected, `version: "1"` only, Go/PostgreSQL/Redis in v0.1.

## Limits (v0.1)

- Single application process per project.
- PostgreSQL + Redis only, Docker backend only, Linux only.
- `POSTGRES_PASSWORD` defaults to `tanuki` (local dev; don't expose beyond localhost).
- `up`/`down` lock is per-machine; concurrent runs exit 7.
- No managed runtime installation — Tanuki validates, never installs.

Docs: `docs/CONFIGURATION.md`, `docs/ARCHITECTURE.md`, `docs/DEVELOPMENT.md`, `docs/SECURITY.md`, `docs/TROUBLESHOOTING.md`.

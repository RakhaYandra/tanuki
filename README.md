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

Or grab a release binary (`tanuki-linux-amd64` / `tanuki-linux-arm64`, see `SHA256SUMS`) and place it in `~/.local/bin`.

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

## Configuration

`tanuki.yaml` in the project dir (discovery: cwd → parents → `/`; override `--config PATH`). Only `version: "1"`. Strict: unknown fields rejected.

```yaml
version: "1"
project: {name: pulse}
runtime: {go: "1.27"}              # validated, never auto-installed
services:
  postgres: {type: postgres, version: "16", port: 5432}
  redis: {type: redis, version: "7", port: 6379}
application:
  name: api
  command: [go, run, ./cmd/api]    # argv array, never a shell string
  port: 8080
  env_file: .env                   # optional, relative, no secrets in yaml
  depends_on: [postgres, redis]
  healthcheck: {type: http, path: /health, timeout: 60s}
```

Rules: `project.name` and service names `^[a-z0-9][a-z0-9-_]{1,62}$`; ports 1–65535 unique across services+app; `depends_on` must reference declared services; `env_file` relative without `..`; healthcheck `tcp|http`, timeout `1s`–`5m`. Full example: `tanuki.example.yaml`.

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

## How it works

- Containers carry labels `io.tanuki.managed=true`, `io.tanuki.project=<name>`, `io.tanuki.service=<svc>`; `down` only touches labeled resources. Named volumes `tanuki-<project>-<svc>-data` are never deleted by Tanuki.
- The app runs detached (no daemon): PID + `/proc` starttime recorded so `down` never signals a stranger.
- App env = process environment overridden by `.env` keys; values are never printed by any command.
- State in `$XDG_STATE_HOME/tanuki/` (`~/.local/state/tanuki/`), `0600`, self-healing (stale entries pruned, corrupt files backed up).

## Limits (v0.1)

- Single application process per project.
- PostgreSQL + Redis only, Docker backend only, Linux only.
- `POSTGRES_PASSWORD` defaults to `tanuki` (local dev; don't expose beyond localhost).
- `up`/`down` lock is per-machine; concurrent runs exit 7.
- No managed runtime installation — Tanuki validates, never installs.

## Development

```bash
make build
make vet
make test        # unit, no Docker needed
make race
go test -tags=integration ./internal/docker/ ./internal/services/  # needs daemon
./scripts/e2e.sh # full lifecycle, cleans up
```

See `CHANGELOG.md` for release history.

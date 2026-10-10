# Changelog

## v0.1.0

First release. Full local lifecycle for a Go API backed by PostgreSQL and Redis:

- `init` (safe defaults, no overwrite without `--force`), `validate` (strict schema), `config` (yaml|json with yaml keys).
- `doctor` read-only checks (OS/arch, Docker CLI/daemon, Go semver, ports, env_file); exits 0/2/3/4.
- `up`/`down` with ownership labels, named volumes (never deleted), TCP/HTTP health, `--dry-run`/`--detach`/`--wait`; exit 5/6.
- Detached app (setsid, PID+starttime identity, `app.log`), env merge (file wins), graceful shutdown.
- `status` (READY/DEGRADED, stale pruning, config-drift warn), `ps`, `logs` (+`--follow`).
- State recovery (corrupt backup), per-machine up/down lock (exit 7).

Limits: single app, postgres+redis only, Docker backend, Linux, dev-default postgres password, no managed runtime install.

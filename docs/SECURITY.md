# Security

What Tanuki does with your machine, and what it will never do.

## Execution model

- `application.command` is an argv array executed directly via `exec.Command` — no shell, no expansion, no `sh -c`.
- Subprocesses (`go`, `docker`, `tail`) run with the invoking user's privileges. Tanuki never uses sudo, never changes group membership, never touches system repositories.

## Resource ownership

- Every container Tanuki creates carries `io.tanuki.managed=true`, `io.tanuki.project=<name>`, `io.tanuki.service=<svc>`.
- `down` lists by label and verifies ownership before stopping. Containers without these labels are never touched.
- Name: `tanuki-<project>-<service>`. Names alone never authorize mutation.

## Data preservation

- Each service gets a named volume `tanuki-<project>-<svc>-data`. Normal `down` only stops; it never runs `rm -v` or `volume rm`.
- There is no destructive command in v0.1. Removing volumes requires explicit `docker volume rm`.

## Secrets

- `.env` values are loaded into the app process only. `validate`, `doctor`, `config`, `status`, `ps`, dry-run output print names/paths, never values.
- `logs` shows raw process/container output — if your app prints secrets, they appear there. Don't log secrets.
- `config --format json|yaml` renders the schema struct, which has no secret fields.
- `POSTGRES_PASSWORD` defaults to `tanuki` for local dev only. Do not expose these containers beyond localhost.

## Process safety

- The app PID is recorded with its `/proc` starttime. `down`/`status` verify identity before signaling; a reused PID is treated as stale and never signaled.
- State lives in `$XDG_STATE_HOME/tanuki/` (or `~/.local/state/tanuki/`), files `0600`, written atomically.

## Concurrency

- `up`/`down` hold an exclusive non-blocking flock on the state dir. A second invocation exits 7 instead of racing the first.

# Architecture (v0.1)

Modular monolith, no daemon.

```text
cmd/tanuki/main.go → internal/cli (cobra) → internal/config → internal/errors, internal/reporter
                                          → internal/environment ← internal/runtime, internal/docker, internal/ports, internal/envcheck
```

- `cli`: routing only. No business logic. Stubs for doctor/up/down/status/logs/ps.
- `config`: `Find` (upward search), `Parse`/`ParseFile` (viper read + strict unmarshal), `Validate` (pure), `WriteExample`.
- `errors`: exit codes 0-7 per PRD §14. doctor uses 0/1/2/3/4.
- `reporter`: `✓/!/✗`, respects `NO_COLOR`/`TERM=dumb`.
- `environment`: shared `Check{Group,Name,Status,Message,Hint}` + `Report.Counts`.
- `runtime/docker/ports/envcheck`: leaf checkers, injectable seams, no mutations.
- `docker/containers.go`: run/stop/rm/logs/pull + `List` by label, `Owned` gate.
- `services`: pure `Plan`, `Up` (reuse|start|create + TCP wait, cleanup own-created on fail), `Down` (owned only, volumes kept).
- `health`: `WaitTCP` 500ms poll within timeout.
- `state`: JSON per project under XDG_STATE_HOME, 0600, atomic rename. Holds container ids + app PID/starttime/log + config fingerprint; corrupt files backed up; `AcquireLock` serializes up/down.
- `process`: detached spawn (setsid, log file), `/proc` starttime identity, SIGTERM→SIGKILL.
- `appenv`: environ + `.env` merge (file wins), names-only warnings.
- `health/http`: 2xx poll for app readiness.

Later: planner, docker adapter, process manager, health, state. Boundaries stay behind `internal/`.

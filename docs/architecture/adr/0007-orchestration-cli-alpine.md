# ADR-0007: Orchestration via Docker CLI, alpine pins, dev password

## Status
Accepted

## Context
Phase 3 needs container CRUD without new heavy deps. User chose CLI backend and alpine tags.

## Decision
Reuse `runtime.Runner` + `docker ... --format '{{json .}}'` for ps/run/stop; no Docker SDK. Pin `postgres:<v>-alpine`, `redis:<v>-alpine`; reject `latest`/empty. Postgres gets dev-only `POSTGRES_PASSWORD=tanuki` (warned in output, never logged beyond `***`). Health = TCP dial only. `down` stops label-owned containers, never removes volumes.

## Consequences
No new dependencies. Image change requires `down` first (explicit, no auto-recreate). Production passwords out of scope until secret management lands.

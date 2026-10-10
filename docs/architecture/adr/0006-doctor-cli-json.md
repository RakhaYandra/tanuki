# ADR-0006: Doctor via Docker CLI structured JSON

## Status
Accepted

## Context
PRD prefers structured APIs over human-output parsing, but Docker SDK is heavy for v0.1. User chose CLI JSON path.

## Decision
`docker version/info --format '{{json .}}'` + `encoding/json` for facts; substring match on stderr only to classify daemon-down vs permission-denied for hints. Go compare via `golang.org/x/mod/semver` with `1.27` → `v1.27.0` normalization. Ports via TCP dial, labeled advisory.

## Consequences
No SDK dep. Doctor stays read-only. Daemon-down vs perms give distinct exit-3 hints. Port TOCTOU documented in output.

# Configuration (v0.1)

File: `tanuki.yaml` in project dir. Discovery: cwd → parents → `/`.
Override: `tanuki --config PATH <cmd>`.

Only `version: "1"` accepted.

| Field | Required | Rule |
|---|---|---|
| `project.name` | yes | `^[a-z0-9][a-z0-9-_]{1,62}$` |
| `runtime.go` | no | `^\d+\.\d+(\.\d+)?$`, e.g. `1.27` |
| `services.<name>` | no | name same regex, max 16 |
| `services.<n>.type` | yes | `postgres`\|`redis` |
| `services.<n>.version` | yes | non-empty image tag |
| `services.<n>.port` | no | 1-65535, unique across services+app |
| `application.command` | if app present | `[]string`, len>=1, no empty, no shell |
| `application.port` | no | 1-65535, unique |
| `application.depends_on` | no | must reference declared services |
| `application.env_file` | no | relative, no `..` |
| `healthcheck.type` | no | `tcp`\|`http` |
| `healthcheck.path` | for http | starts with `/` |
| `healthcheck.timeout` | no | `1s`-`5m`, default `30s` |

Strict: unknown fields rejected. `command` is argv array, never `sh -c`.
App env: process environ overridden by `.env` keys on clash; empty values warn; values never printed.
Example: `tanuki.example.yaml`. E2E fixture: `testdata/e2e/` (Go `/health` api + `.env.example`).

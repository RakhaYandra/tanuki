#!/usr/bin/env bash
# End-to-end: build, up, status, logs, reuse, down, volume preserved, cleanup.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin/tanuki"
WORK="$(mktemp -d)"
PROJ="tke2e"
PGPORT=55443
RDPORT=56393
APPPORT=18081

cleanup() {
  "$BIN" down --config "$WORK/tanuki.yaml" >/dev/null 2>&1 || true
  docker rm -f "tanuki-$PROJ-postgres" "tanuki-$PROJ-redis" >/dev/null 2>&1 || true
  docker volume rm "tanuki-$PROJ-postgres-data" "tanuki-$PROJ-redis-data" >/dev/null 2>&1 || true
  rm -rf "$WORK"
}
trap cleanup EXIT

go -C "$ROOT" build -o "$BIN" ./cmd/tanuki
cp "$ROOT/testdata/e2e/main.go" "$WORK/main.go"
printf 'PORT=%s\n' "$APPPORT" > "$WORK/.env"
cat > "$WORK/tanuki.yaml" <<EOF
version: "1"
project: {name: $PROJ}
runtime: {go: "1.27"}
services:
  postgres: {type: postgres, version: "16", port: $PGPORT}
  redis: {type: redis, version: "7", port: $RDPORT}
application:
  name: api
  command: [go, run, ./main.go]
  port: $APPPORT
  env_file: .env
  depends_on: [postgres, redis]
  healthcheck: {type: http, path: /health, timeout: 60s}
EOF

"$BIN" validate --config "$WORK/tanuki.yaml"
"$BIN" up --dry-run --config "$WORK/tanuki.yaml" | grep -q "start app"
"$BIN" up --config "$WORK/tanuki.yaml"
"$BIN" status --config "$WORK/tanuki.yaml" | grep -q "Overall status: READY"
"$BIN" logs api --config "$WORK/tanuki.yaml" | grep -q "listening on"
curl -fsS "http://127.0.0.1:$APPPORT/health" | grep -q ok
"$BIN" up --config "$WORK/tanuki.yaml" | grep -q "already running"
"$BIN" down --config "$WORK/tanuki.yaml"
docker volume ls --format '{{.Name}}' | grep -q "tanuki-$PROJ-postgres-data"

echo "E2E OK"

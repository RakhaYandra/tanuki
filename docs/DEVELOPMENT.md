# Development

Requires Go 1.27+ and (for services) Docker Engine.

```bash
make build     # bin/tanuki
make vet       # go vet ./...
make test      # unit tests (no Docker needed)
make race      # unit tests with -race
make lint      # gofmt check + vet
make install   # install to ~/.local/bin
make release   # dist/ amd64+arm64 + SHA256SUMS
./scripts/e2e.sh   # full lifecycle against real Docker
```

## Test layers

- Unit: `go test ./...` — pure, hermetic, Docker never touched.
- Integration: `go test -tags=integration ./internal/docker/ ./internal/services/` — needs a Docker daemon; skips otherwise.
- E2E: `./scripts/e2e.sh` — sample Go API + postgres + redis through the real CLI; cleans up after itself.

## Style

`gofmt -l` must be empty. Keep dependencies minimal; stdlib first. No shell execution of config values, ever.

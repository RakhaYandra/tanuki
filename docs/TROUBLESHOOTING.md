# Troubleshooting

## Docker daemon unavailable (exit 3)

`docker info` fails. Start the daemon (`systemctl start docker`), check group membership on permission errors. Tanuki never fixes this for you.

## Port already in use (exit 4)

`up` and `doctor` report the port and stop. Tanuki never kills the holder. Either stop it yourself or set another `port` in `tanuki.yaml`.

## Service unhealthy (exit 5)

`up` prints the last 50 log lines. Inspect with `tanuki logs postgres|redis`, or `docker logs <name>` directly. Containers stay running for inspection; `tanuki down` stops them.

## Application unhealthy (exit 5)

Services and app keep running. Check `tanuki logs api` (log path shown at startup), verify the health `path` and `port`, then `tanuki down` and retry.

## Stale state

If a container was removed manually or the app PID was reused, `status` prunes dead entries (`state cleaned`) and `down` clears stale app records without signaling strangers. Corrupt state files are renamed to `*.corrupt.<ts>` and rebuilt.

## Config changed since up

`status` warns when `tanuki.yaml` no longer matches the fingerprint recorded at `up`. Run `tanuki up` again to reconcile.

## Another operation running (exit 7)

`up`/`down` take a lock. Wait for the other command to finish and retry.

## Interrupted startup

`Ctrl-C` during `up` may leave created containers running. This is safe: just run `tanuki up` again (idempotent reuse) or `tanuki down` to stop. Volumes are always preserved.

## Invalid config (exit 2)

`tanuki validate` explains every problem without touching anything. Unknown fields are rejected — check spelling against `docs/CONFIGURATION.md`.

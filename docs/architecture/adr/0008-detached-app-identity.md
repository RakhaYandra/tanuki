# ADR-0008: Detached app via setsid + starttime identity

## Status
Accepted

## Context
App must survive the `up` invocation (detached default) without a daemon, and `down` must never signal a wrong process on PID reuse.

## Decision
Spawn with `exec.Command` direct argv, `Setsid`, stdout/stderr to `<state>/<project>-app.log`, `Process.Release`. Record PID + `/proc` starttime in state. Stop verifies starttime match first; mismatch clears state without signaling. Shutdown: SIGTERM → 10s → SIGKILL. Env = process environ overridden by `.env` file keys; names-only in output.

## Consequences
`up` returns while app runs; logs persist in state dir. PID reuse is safe by construction. Single app only in v0.1.

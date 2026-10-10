# ADR-0003: No shell execution

## Status
Accepted

## Context
PRD security: never execute config through implicit shell.

## Decision
`application.command` is `[]string` (argv). Later phases use `exec.Command(name, args...)` direct, no `sh -c`, no expansion.

## Consequences
Users write argv arrays. Shell pipelines must be explicit scripts, reviewed separately.

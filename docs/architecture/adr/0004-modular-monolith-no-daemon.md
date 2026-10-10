# ADR-0004: Modular monolith, no daemon

## Status
Accepted

## Context
PRD: useful without daemon, no microservices/plugins in v0.1.

## Decision
`cmd/tanuki` + `internal/{cli,config,errors,reporter}` only in Phase 1. Later: planner/docker/process/health/state as internal packages with test boundaries.

## Consequences
Simple build, easy test. Cross-project reuse deferred.

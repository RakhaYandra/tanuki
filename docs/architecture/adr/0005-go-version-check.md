# ADR-0005: Go version check via x/mod/semver (Phase 2)

## Status
Proposed

## Context
PRD forbids custom version comparator; must test against real `go version` output.

## Decision
Phase 1: syntax-check only (`^\d+\.\d+(\.\d+)?$`). Phase 2: compare with `golang.org/x/mod/semver` against `go version` output.

## Consequences
Phase 1 reports shape errors only, no toolchain probing. No auto-install ever in v0.1.

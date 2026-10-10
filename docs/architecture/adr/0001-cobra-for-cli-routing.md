# ADR-0001: Cobra for CLI routing

## Status
Accepted

## Context
11 commands (`init/doctor/validate/up/down/status/logs/ps/config/version`). stdlib-only is lighter but cobra was explicitly chosen by user for routing ergonomics. Viper available for config binding.

## Decision
Use cobra for command tree, persistent `--config` flag. Use viper only to read `tanuki.yaml` (`SetConfigFile` + `UnmarshalExact`); no env auto-magic, no defaults magic in Phase 1.

## Consequences
Easier help/completion; heavier deps. Must keep business logic out of cli package. Viper misuse (silent defaults) is a review blocker.

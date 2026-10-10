# ADR-0002: Strict YAML decoding

## Status
Accepted

## Context
PRD: unsupported fields must not be silently ignored.

## Decision
Dual strict path: `yaml.Decoder.KnownFields(true)` in `Parse`, `viper.UnmarshalExact` in `ParseFile`. Both reject unknown fields with exit 2.

## Consequences
Typos fail fast. Schema additions need code + docs + testdata update.

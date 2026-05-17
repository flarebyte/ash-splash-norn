# ash-splash-norn

![Ash Splash Norn illustration](doc/ash-splash-norn.png)

A Go CLI for validating and generating distributed configuration artifacts from CUE sources.

## Who this is for

Use this CLI if you want one source of truth for configuration and localization keys, then generate outputs for multiple runtimes and repositories.

## What the CLI does

- Validates registry and config inputs against CUE schemas.
- Lints key coverage, translation coverage, schema graph integrity, and artifact patterns.
- Generates artifacts for multiple targets:
  - `arb.json` for Flutter i18n bundles
  - `json` and `yaml` for portable config exchange
  - `cue` for CUE-native downstream composition
  - `go` and `dart` runtime config code
- Previews output plans without writing files.

## Input model

The CLI consumes two main CUE inputs:

- Design registry: key schema and generation capabilities.
- Config key entries: `i18nEntries`, `textEntries`, and `validations`.

Reference examples in this repo:

- `doc/design-meta/examples/input/design-registry.example.cue`
- `doc/design-meta/examples/input/config-key.cue`
- `doc/design-meta/examples/model/design-registry.schema.cue`
- `doc/design-meta/examples/model/config-key.schema.cue`

## Typical user workflow

1. Author or update your design registry and config-key CUE files.
2. Run `validate` to catch shape and schema failures early.
3. Run `lint` to enforce policy checks.
4. Run `dry-run-preview` to inspect planned artifact outputs.
5. Run `generate` for your target(s).

## CLI commands

- `validate`: Validate registry/config CUE inputs.
- `lint`: Run full policy checks (aggregate).
- `lint schema`: Validate schema governance rules.
- `lint config`: Validate config completeness and quality.
- `lint sections`: Check validation command sections.
- `lint keys`: Verify expected vs actual generated keys.
- `lint translations`: Enforce required language coverage.
- `lint patterns`: Validate artifact pattern tokens.
- `lint graph`: Detect cycles and key collisions.
- `list`: List available schemas and target capabilities.
- `explain-key`: Show key derivation trace from label path.
- `dry-run-preview`: Show planned outputs without writing files.
- `generate`: Generate artifacts by target capability.
- `version`: Print version/build metadata.

## Example commands

```bash
# Validate sample inputs
ash-splash-norn validate \
  --registry doc/design-meta/examples/input/design-registry.example.cue \
  --registry-schema doc/design-meta/examples/model/design-registry.schema.cue \
  --config doc/design-meta/examples/input/config-key.cue \
  --config-schema doc/design-meta/examples/model/config-key.schema.cue

# Lint all checks
ash-splash-norn lint \
  --registry doc/design-meta/examples/input/design-registry.example.cue \
  --config doc/design-meta/examples/input/config-key.cue

# Lint translation coverage only
ash-splash-norn lint translations \
  --registry doc/design-meta/examples/input/design-registry.example.cue \
  --config doc/design-meta/examples/input/config-key.cue

# Preview output plan
ash-splash-norn dry-run-preview \
  --registry doc/design-meta/examples/input/design-registry.example.cue \
  --config doc/design-meta/examples/input/config-key.cue \
  --format json

# Generate JSON artifacts
ash-splash-norn generate json \
  --registry doc/design-meta/examples/input/design-registry.example.cue \
  --config doc/design-meta/examples/input/config-key.cue
```

## Expected behavior

- Hard failure on schema and generation-policy errors.
- Non-zero exit on lint violations.
- Deterministic output ordering for generated files and machine-readable diagnostics.

## Diagnostics contract

Diagnostics can be emitted as JSON (`--format json`) and follow a stable shape:

- `stage`
- `id`
- `severity`
- `message`
- `location` (when available)
- `suggestion` (when available)

Current ID families:

- `SCH-*`: schema/input validation and artifact-pattern preflight.
- `LNT-*`: lint checks (`schema`, `config`, `sections`, `keys`, `translations`, `patterns`, `graph`).
- `GEN-*`: generation pipeline and artifact writing policy.
- `PRV-*`: dry-run preview registry parsing and pattern policy.
- `KEY-*`: explain-key command validation.
- `LST-*`: list command registry decoding.

## Repository helpers

Common local tasks:

- `make doc-design`: regenerate design docs from `doc/design-meta/*.cue`.
- `make cue-input`: validate CUE example inputs against schema contracts.
- `make ghf-config-validate`: validate `.gh-flarebyte.cue`.
- `make ghf-repo-audit`: audit GitHub settings drift against `.gh-flarebyte.cue`.

## Current project status

This repository currently defines the CLI design and implementation backlog.
See:

- `doc/design/distributed-config-cli.md`
- `scratch/prompt.md`

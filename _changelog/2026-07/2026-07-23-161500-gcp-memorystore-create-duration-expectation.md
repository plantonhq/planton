# GCP catalog: the Memorystore instance create-duration expectation, set where authors read

**Date**: 2026-07-23
**Scope**: `gcpmemorystoreinstance` (spec header behavioral note + regenerated Go stub + docs). Comments and documentation only — zero behavior change.

## What changed

New-generation Memorystore (Valkey / Redis-cluster) instance creation is a
long-running operation: the Terraform provider's default create timeout is
60 minutes (verified in the provider source), and multi-shard instances
commonly run tens of minutes. Neither the spec nor the docs set that
expectation, so a first deploy sitting at 20 minutes reads as a hang — and
the natural reaction, aborting or re-running, is worse than waiting. The
spec header's behavioral-notes list and the docs introduction now carry the
expectation.

The legacy `gcpredisinstance` deliberately does not carry the note: its
provider create budget is 20 minutes, below the threshold where a deploy
reads as broken.

## Why

The same contract as the catalog's other duration notes (Cloud Composer's
25–45 minutes, and the equivalents on other providers' slow services): the
spec and its docs are enough to operate a kind correctly, so duration
expectations belong on the kind surfaces, not in tribal memory.

## Validation

- `buf lint` + `buf format --diff` clean; stub regenerated with the comment
  content verified in the generated file before staging.
- Spec tests and the Pulumi release-entrypoint build green; repo-wide
  `make build-go` green.
- No module, CEL, preset, or chart changed; existing E2E results stand by
  construction.

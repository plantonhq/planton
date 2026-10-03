# Analyzer Invocation and Targeted-Build Guidance in Kind Workflow Rules

**Date**: July 3, 2026
**Type**: Enhancement
**Components**: Build System, Provider Framework

## Summary

Two small guidance fixes in the catalog-kind workflow rules, closing gaps
that cost real agent time: the audit rule now explains how to invoke the repo's
analyzer commands when a different `planton` binary shadows them on PATH, and the
update rule's rename flow no longer prescribes a repo-root `go build ./...`.

## Problem Statement

1. The audit rule prescribes `planton secret-coverage` / `planton validate-refs`
   without saying where that CLI comes from. A developer or agent with a different
   `planton` binary installed on PATH (one that does not carry the analyzer
   commands) hits `unknown command "secret-coverage"` and has to reverse-engineer
   that this repo's root module is itself the CLI.
2. The update rule's kind-rename flow listed `go build ./...` as a validation
   step. A repo-root `go build ./...` compiles 70+ kind modules (each with
   its own Pulumi SDK surface) in one shot and brings the machine to a halt; the
   repo's own build documentation already bans it everywhere else.

## What Changed

- `_rules/catalog-kind/audit/audit-catalog-kind.mdc` — the Category
  11 evaluation now notes that a shadowing `planton` binary may lack the analyzer
  commands and gives the canonical fallback: run from the repo root as
  `go run . secret-coverage --output json` (a targeted single-binary compile).
- `_rules/catalog-kind/update/update-catalog-kind.mdc` — the rename
  flow's validation step now prescribes targeted package builds
  (`go build ./apis/dev/planton/provider/<provider>/<kind>/v1/...`) plus
  `make build-go` for the repo-wide gate, and states explicitly why the repo-root
  form is banned.

## Impact

Every future kind session that runs the audit gate or the rename flow gets
the correct invocation on first read — no re-discovery, no accidental
machine-halting build.

---

**Status**: ✅ Production Ready

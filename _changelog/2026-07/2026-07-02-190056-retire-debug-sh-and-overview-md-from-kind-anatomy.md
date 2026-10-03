# Retire debug.sh and overview.md from the Pulumi Module Anatomy

**Date**: July 2, 2026
**Type**: Refactoring
**Components**: Provider Framework, Build & Tools, API Definitions

## Summary

Removed `iac/pulumi/debug.sh` and `iac/pulumi/overview.md` from the
catalog-kind anatomy, the audit checklists, and the forge docs flow. Zero
of the 407 Pulumi modules in the catalog carry either file — the catalog's
convention had moved on, but the doctrine and workflow rules still demanded the
artifacts and pointed agents at reference files that do not exist. The docs
tooling and rules are now aligned with what the catalog actually ships:
`iac/pulumi/README.md` as the single Pulumi supporting doc.

## Problem Statement / Motivation

The catalog-kind doctrine (`architecture/catalog-kind.md`) listed
`debug.sh` and `overview.md` in the canonical component tree and in two scoring
checklists; the audit rule scored `iac/pulumi/debug.sh exists` under Helper Files;
and forge flow rule 012 instructed agents to generate `debug.sh` modeled on
canonical reference files (`.../awsvpc/v1/iac/pulumi/debug.sh`).

### Pain Points

- No kind in the catalog has `debug.sh` or `overview.md` (0 of 407 Pulumi
  module directories), so every kind silently failed a scored audit check.
- Rule 012's "live ground truth" references pointed at files that do not exist —
  a dead end for any agent following the forge flow.
- The docs-writer script (`pulumi_docs_write.py`) required a `--debug-file`
  argument, forcing agents to author an artifact no rule consumer keeps.

## Solution / What's New

Aligned all durable surfaces with the catalog's real anatomy:

- `architecture/catalog-kind.md` — removed `debug.sh` and `overview.md`
  from the component tree and from both Pulumi supporting-file checklists;
  `README.md` remains the Pulumi supporting doc.
- `_rules/catalog-kind/audit/audit-catalog-kind.mdc` (+ its README) —
  Helper Files now checks only `iac/hack/manifest.yaml`.
- `_rules/catalog-kind/forge/flow/012-pulumi-docs.mdc` — retitled to
  README-only; the canonical reference is now
  `apis/dev/planton/provider/aws/awsalb/v1/iac/pulumi/README.md` (a file that
  exists).
- `_rules/catalog-kind/forge/forge-catalog-kind.mdc`, forge
  `README.md`, `FORGE_ANALYSIS.md`, delete-rule README — step descriptions and
  example trees updated to match.
- `_rules/catalog-kind/_scripts/pulumi_docs_write.py` — now writes
  README.md only; the `--debug-file` argument is gone (docstring and JSON output
  updated; `terraform_docs_write.py` docstring reference updated).

## Impact

Agents forging or auditing any kind no longer chase artifacts the catalog
does not ship, and audit scores no longer carry a universal false negative. No
module code, protos, or generated stubs changed; no user-facing CLI behavior
changed.

## Related Work

Part of the ongoing per-provider catalog quality work; the audit rule's scoring
categories are otherwise unchanged.

---

**Status**: ✅ Production Ready

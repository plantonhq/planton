# Licensing Hygiene in Authoring Rules + License Footer by Default

**Date**: July 22, 2026
**Type**: Enhancement
**Components**: Forge Rules, Chart Authoring, Documentation Tooling, Legal/Licensing

## Summary

The kind and chart authoring workflows now teach the inbound licensing
rule they previously assumed: provider source is read for facts, its code
text is never copied. And the canonical license footer became a
creation-time guarantee — the deterministic README writer appends it when a
draft lacks it, and the chart authoring rule carries it as the README's
closing structural element — so the CI footer lint should never fire on
correctly forged work.

## Problem Statement / Motivation

Two gaps between what the repo enforces and what its rules teach:

1. The forge, update, and architecture documents direct agents to derive
   kind design from cloned Terraform provider source — the authoritative
   reference. Those repos are **MPL-2.0** (file-level copyleft that must not
   be mixed into this Apache-2.0 work), and HashiCorp's Terraform core is
   **BUSL** (not open source). Nothing said what must NOT be taken from
   them. One pasted snippet by a well-meaning agent would contaminate an
   Apache-2.0 module.
2. The license-footer lint enforces the footer on every kind and chart
   README, but nothing *created* the footer: a freshly forged kind or
   chart would fail CI unless the author remembered to add it by hand.

## Solution / What's New

### Read for facts, never copy (three homes, one message)

- `_rules/catalog-kind/forge/forge-catalog-kind.mdc` — new
  Design Philosophy bullet directly after the one that sends agents into the
  provider clones.
- `_rules/catalog-kind/update/update-catalog-kind.mdc` — the
  same rule where its Design Philosophy names the clone.
- `architecture/catalog-kind.md` — one sentence extending "Schema as
  the Floor".

The line is drawn precisely: field sets, defaults, validation bounds, and
ForceNew behavior are facts — extract them freely. Module code is original
authorship informed by those facts. And explicitly NOT restricted: consuming
the Apache-2.0 Pulumi provider SDKs as go.mod dependencies is exactly what
they are for — the rule prohibits transplanting code text from
incompatible-licensed repos, nothing more.

### Footer by default

- `_rules/catalog-kind/_scripts/docs_write.py` — the deterministic
  kind-README writer appends the canonical footer when the drafted
  content does not already end with it (exactly one footer in all cases).
  The tool that writes the file owns the invariant; `lint.license-footers`
  remains the independent CI check.
- `_rules/catalog-kind/forge/flow/007-docs.mdc` — documents the
  guarantee.
- `_rules/charts/forge-planton-infra-chart.mdc` — the footer is now item 6
  of the README structure: charts are in the lint's scope but have no
  deterministic writer, so the authoring rule carries the invariant.

The Pulumi/Terraform module README writers are deliberately unchanged —
those inner docs are outside the footer lint's scope by design (their
shipped artifacts carry LICENSE and NOTICE files instead).

## Validation

- `python3 -m py_compile` on the writer; smoke test of the footer logic with
  a footer-less draft and a footer-carrying draft — both outputs end with
  exactly one canonical footer.
- Full license-footer guard re-run over all 624 in-scope READMEs: green (no
  regression from the writer change).
- MPL-2.0 verified on the LICENSE files of all three workspace provider
  clones before writing the claim into the rules.
- Read-through of each edited rule section for tone and flow with the
  surrounding text.

## Impact

- **Agents forging or updating components/charts** get the licensing
  boundary as part of the workflow that already governs their research —
  no separate legal document to remember.
- **The catalog's Apache-2.0 cleanliness** is protected at the point of
  authorship, which is the only place it can be.
- **Footer compliance is now automatic** — the lint becomes a true backstop
  instead of the first line of defense.

## Related Work

- `lint.license-footers.yaml` + `hack/guards/ensure_license_footers.sh` —
  the CI enforcement this change makes routine.
- NOTICE, TRADEMARKS.md, CLA.md — the outbound legal posture; this change is
  its inbound counterpart at authoring time.

---

**Status**: ✅ Production Ready

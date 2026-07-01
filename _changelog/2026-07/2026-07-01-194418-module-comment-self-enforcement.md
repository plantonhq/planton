# Richly-Commented IaC Modules Become a Scored, Self-Enforcing Requirement

**Date**: July 1, 2026
**Type**: Enhancement
**Components**: Deployment-Component Doctrine, Forge Rules, Update Rule, Fix Rule, Audit Rule

## Summary

The Terraform and Pulumi module code behind every resource kind now carries a
first-class, *scored* requirement to be richly commented. Previously the rule set
graded IaC modules on file presence, size, feature parity, and build success only,
so a working-but-uncommented module could score 100%. This change defines a single
"module-comment bar," plants it at every point where module code is authored, and
makes the audit score it — so module readability can no longer silently erode.

## Problem Statement / Motivation

These modules render on the public catalog; people read them, learn from them, and
copy their patterns into their own infrastructure. Yet the only in-module comment
requirement anywhere was the `PARITY-EXCEPTION:` marker. The doctrine's completeness
model and the audit's Category 4/5 scoring said nothing about authoring-comment
quality. A mandate declared in the forge rule alone would have been toothless: the
next audit-driven pass would still grade a comment-stripped module as complete, and
quality would decay component by component.

## Solution / What's New

A single **module-comment bar**, defined once and referenced everywhere:

> Module code carries authoring comments that explain *why* / trade-offs / provider
> quirks / non-obvious ordering — not line-by-line narration — to the same density
> and intent as the `spec.proto` field-comment standard.

- **Defined once** in `architecture/deployment-component.md` §4.1 (Pulumi Code
  Quality), referenced by the Terraform Code Quality list, the Completeness /
  Incompleteness Indicators, the "IaC Implementation Quality" Quality Multiplier,
  and the "For Auditing" spec list.
- **Planted at every authoring surface**: the forge orchestrator
  (`forge-planton-component.mdc`), the two flow rules that actually write the code
  (`009-pulumi-module`, `013-terraform-module`), the update rule (Scenario 4 update
  IaC + Scenario 3 docs refresh), and the fix rule (Step 2 "Fix IaC Modules" +
  success criteria). The convenience `complete` rule inherits it through audit +
  update, so it needs no edit.
- **Scored, not just declared**: the audit rule adds a "Module Comment Quality
  (scored as part of ...)" sub-block to Category 4 (Pulumi) and Category 5
  (Terraform), plus an appendix checklist item — mirroring the existing
  "Validation Message Quality (scored as part of Proto Files)" precedent. It folds
  into the existing category weights, so the rubric still sums to 100% with no
  re-normalization.

## Design Decisions

- **Scored, not gated.** Release-blocking gates (cross-engine parity, secret
  coverage) are reserved for divergences that would misdeploy. Missing comments do
  not misdeploy — they lower quality — so comment quality reduces the score rather
  than blocking a release, consistent with the doctrine's Quality-Multiplier model.
- **No new percentage slice.** Folding the check into the existing Module Files
  (Category 4) and Terraform (Category 5) scores avoids re-weighting the entire
  rubric and reuses the pattern the audit already uses for proto-message quality.

## Also In This Change (stale-reference reconciliation)

While in these files, three drifted references were corrected:

- The doctrine's registry location `apis/project/planton/...` → `apis/dev/planton/...`.
- The doctrine's stack-input example referenced a non-existent `AwsCredential`
  credential type; corrected to the provider-config pattern
  (`<Provider>ProviderConfig` from `provider/<provider>/provider.proto`, e.g.
  `AwsProviderConfig`), consistent with the correct example directly above it.
- The provider numeric-range list (present in the doctrine, the audit rule, and the
  `016-cloud-resource-kind` forge rule) stopped at Cloudflare (1800-2099). All three
  now carry the full authoritative range map derived from `cloud_resource_kind.proto`
  (through Auth0, OpenFGA, OpenStack, Scaleway, Alibaba Cloud, OCI, and Hetzner Cloud
  3500-3699).

## Impact

- Every future forge/update/fix session is instructed to comment modules to the bar
  at the moment it writes them, and every audit scores modules on it — the standard
  holds without a human reviewer.
- Applies uniformly across all providers (the rules are provider-agnostic); no
  per-provider special case.
- Documentation-only change to the rule set; no proto, Go, or module code changed,
  so no build or stub regeneration was required.

---

**Status**: ✅ Production Ready

# Self-Improving Deployment-Component Workflows

**Date**: July 2, 2026
**Type**: Enhancement
**Components**: Provider Framework, Deployment-Component Rules

## Summary

Added a standing "improve this workflow as you use it" duty to the two OSS
deployment-component orchestrator rules — the forge rule (`forge-planton-component.mdc`)
and the update rule (`update-planton-component.mdc`). Any agent running these rules that
had to reverse-engineer an undocumented mechanism, correct a wrong implication, supply a
missing step, or fix a stale path is now expected to fold that learning back into the
rule, its flow rules, or `architecture/deployment-component.md` in the same session — as
timeless guidance — so the next agent forging or updating the next component never has to
re-learn it.

## Problem Statement / Motivation

The deployment-component forge/update workflow is the primary way components are built
across every provider, and coding agents are its primary consumers. Real knowledge about
how the framework behaves (harness wiring, registry mechanics, resolver limitations,
build ordering) has historically surfaced only when an agent hit a dead end and dug
through the code — and then stayed in that one session's context. The next agent started
from the same incomplete rules and repeated the same research and the same wrong
assumptions.

## Solution / What's New

A short, non-negotiable clause in each orchestrator rule:

- **Forge** (`_rules/deployment-component/forge/forge-planton-component.mdc`): added as a
  new bullet in the "Design Philosophy (NON-NEGOTIABLE)" section, beside the existing
  "docs are part of the experience" and "mirror STRUCTURE not DEPTH" principles.
- **Update** (`_rules/deployment-component/update/update-planton-component.mdc`): added as
  a concise "Improve this workflow as you use it (NON-NEGOTIABLE)" section after the Role.

The clause is placed only at the orchestrator level; the nested flow rules are invoked
through the orchestrators and inherit the duty, keeping the edit surface minimal and the
duty un-missable. Guidance is timeless — it never names a project, task, or design
decision — consistent with the framework's no-breadcrumbs standard.

## Impact

Every future forge/update session, on every provider, is now told to leave the workflow
rules sharper than it found them. Over many component sessions this compounds: the rules
become a progressively more complete teacher, reducing repeated research and preventing
recurring wrong assumptions.

## Related Work

Follows the precedent of making a quality duty self-enforcing where the work happens
(the richly-commented-module mandate) rather than declaring it only in external
scaffolding.

---

**Status**: ✅ Production Ready

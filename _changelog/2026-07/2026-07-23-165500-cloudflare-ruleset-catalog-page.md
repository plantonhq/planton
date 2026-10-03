# Cloudflare catalog: a proper Ruleset catalog page replaces the technical deep-dive on the public site

**Date**: 2026-07-23
**Scope**: `cloudflareruleset/v1/catalog-page.md` (new) + the propagated site pages (`site/public/docs/catalog/cloudflare/ruleset/`, provider index link, stale `cloudflareruleset-technical-deep-dive/` pages removed). Documentation only — zero behavior change.

## What changed

CloudflareRuleset was the only Cloudflare kind (of 30) without a
`catalog-page.md`, so the public catalog's page for the kind was its
internal engineering deep-dive (pipeline diagrams, provider schema
analysis, design rationale) served under a
`cloudflareruleset-technical-deep-dive` slug — the wrong register for the
kind-selection moment, and a divergence from every sibling.

The kind now has a proper catalog page in the family shape: what a
Ruleset is (the unified engine behind WAF, rate limiting, cache, origin,
redirect, transform, and configuration rules), what gets created, scope
and List prerequisites, the configuration reference at a glance, stack
outputs, and related kinds. The site now serves it at the clean
`ruleset` slug; the deep-dive remains where it belongs, as the kind's
in-repo technical reference (`docs/README.md`).

## Validation

- Site copy run: the new page renders at
  `site/public/docs/catalog/cloudflare/ruleset/`, the provider index
  links to it, and the stale deep-dive slug pages were removed by the
  copy script's slug replacement.
- No proto, module, preset, or test files touched (documentation only).

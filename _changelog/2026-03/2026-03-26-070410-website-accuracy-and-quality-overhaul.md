# Website Accuracy & Quality Overhaul: Cross-Referencing Every Page Against Source of Truth

**Date**: March 26, 2026
**Type**: Content | Refactoring
**Components**: Features Pages, Solutions Pages, Landing Page, Navigation, UI Components

## Summary

Systematic audit and correction of all 7 product pages, 11 solutions pages, and the homepage hero against the actual codebase, infra-charts repository, CLI docs, OpenMCF repo, and product documentation. Fixed 23 distinct issues across factual errors, visible problems, content quality, and structural improvements. Every code example, CLI command, YAML manifest, and external link on the website now matches the real product.

## Problem Statement / Motivation

The previous two sessions (copywriting overhaul + structural overhaul) updated content and added visual variety, but introduced factual inaccuracies. The AI agent that wrote the content made speculative assumptions about CLI commands, YAML formats, template syntax, and API versions that were provably wrong when checked against the actual codebase.

### Pain Points

- InfraHub page showed Go/Helm template syntax (`{{ .Values.x }}`) but real infra-charts use Jinjava (`{{ values.x }}`)
- InfraHub showed a flat Helm-style values.yaml, but real infra-charts use a `params:` list structure
- InfraHub Chart.yaml showed `name/version/description` but real format is `apiVersion: infra-hub.planton.ai/v1` with `kind: InfraChart`
- Security page used `kubernetes.openmcf.org/v1` for GCP and AWS resources (should be `gcp.openmcf.org/v1` and `aws.openmcf.org/v1`)
- ServiceHub page used `kubernetes.openmcf.org/v1` / `MicroserviceInstance` but services use `service-hub.planton.ai/v1` / `Service`
- CLI page showed fabricated `curl -fsSL https://get.planton.ai | sh` install URL (doesn't exist)
- CLI page showed `planton self-update` but real command is `planton upgrade`
- CLI page showed version `v0.15.0` but real version is `v0.0.3`
- Open Source page showed `openmcf forge init` as a CLI command (forge is a 20-step manual workflow, not a CLI subcommand)
- Runner page had malformed AWS ARN with 3-digit account ID
- Runner CTA linked to `https://docs.planton.ai/runner` instead of `/docs/runner`
- ServiceHub showed fabricated promotion config YAML and Kustomize interpolation syntax
- Agent Fleet showed fabricated `runbook.skill.yaml` workflow format
- Overview page said "150+" resource types while detail pages said "350+"
- Three pages rendered empty `ScreenshotPlaceholder` kinds visible to users
- Agent Fleet hero and deep-dive showed nearly identical terminal animations
- Vanity metrics ("0 Generic Chatbots", "0 Config Required") communicated nothing
- Enterprise page used "Start Free Trial" as primary CTA instead of "Book a Demo"
- Solutions hub linked Platform Engineers to `/solutions/by-role/devops` instead of `/platform-engineers`
- Hardcoded `$19/seat/month` pricing in solution pages
- Homepage hero was cluttered with redundant "Planton" heading, trust badges, sign-in providers text, and competing CTAs
- "Book a Demo" links pointed to `/demo` (interactive tour) instead of Google Forms booking link

## Solution / What's New

### Tier 1: Factual Errors Fixed

1. **InfraHub template syntax**: `{{ .Values.x }}` replaced with `{{ values.x }}` (Jinjava)
2. **InfraHub values.yaml**: Flat YAML replaced with real `params:` list format from infra-charts repo
3. **InfraHub Chart.yaml**: Updated to real `apiVersion: infra-hub.planton.ai/v1` / `kind: InfraChart` format
4. **Security apiVersions**: `kubernetes.openmcf.org/v1` corrected to `gcp.openmcf.org/v1` and `aws.openmcf.org/v1`
5. **ServiceHub manifests**: `kubernetes.openmcf.org/v1` / `MicroserviceInstance` replaced with `service-hub.planton.ai/v1` / `Service` using real spec structure
6. **CLI install URL**: Fabricated `get.planton.ai` replaced with real `brew install plantonhq/tap/planton`
7. **CLI version**: `v0.15.0` removed, version-agnostic display
8. **CLI upgrade command**: `self-update` corrected to `upgrade`
9. **Open Source forge**: `openmcf forge init` replaced with real kind directory structure and 20-step workflow description
10. **Runner AWS ARN**: 3-digit account ID corrected to 12-digit
11. **Runner docs link**: `https://docs.planton.ai/runner` corrected to `/docs/runner`
12. **Resource count**: Standardized to "350+" across overview and detail pages
13. **ServiceHub promotion config**: Fabricated YAML replaced with realistic pipeline output
14. **Agent Fleet skill YAML**: Fabricated workflow YAML replaced with content-based skill representation

### Tier 2: Visible Problems Fixed

15. **Screenshot placeholders removed**: 3 empty `ScreenshotPlaceholder` kinds removed from InfraHub, ServiceHub, CLI pages. Replaced with rich JSX comments describing ideal screenshots for future AI agents
16. **Duplicate terminal**: Agent Fleet deep-dive terminal replaced with distinct security-auditor scenario
17. **Security messaging reframed**: "No plaintext secrets" changed to "No plaintext in production" to align with real dot-env feature

### Tier 3: Content Quality Improved

18. **Vanity metrics replaced**: "0 Generic Chatbots" → "100% Real Execution"; "0 Config Required" → "Git Push to Production"
19. **ServiceHub CTA**: "Explore OpenMCF" → "Read the Docs" (linking to `/docs/ci-cd`)
20. **Enterprise CTA**: "Start Free Trial" demoted to secondary; "Book a Demo" promoted to primary
21. **Hub.tsx link**: Platform Engineers now links to `/solutions/by-role/platform-engineers`
22. **Hardcoded pricing removed**: `$19/seat/month` replaced with "transparent per-seat pricing"
23. **Overview grid**: Changed from 4-column (unbalanced with 7 items) to 3-column layout
24. **Dead files deleted**: Orphaned `roles.tsx` and `page.css` in multi-cloud route

### Tier 4: Screenshot Comments Added

10 strategic `{/* SCREENSHOT OPPORTUNITY */}` comments placed across all product pages, each describing what to capture, why it adds value, and suggested format.

### Homepage Hero Decluttered

Removed from hero:
- Redundant "Planton" standalone heading (already in nav logo)
- "Trusted by Tech Teams - 100% customer retention" badge (unverifiable)
- "Sign in with Google, GitHub, or Microsoft. Privacy Policy" (login page info)
- "See Pricing" link (available in nav)
- Trust indicator pills ("Open Source Foundation", "Zero Vendor Lock-in")
- Visual separator above provider logos

Result: Headline → description → CTA + demo link → terminal → provider logos.

### Book a Demo Links Standardized

All 24 "Book a Demo" CTA links across the website now point to the Google Forms booking link and open in a new tab. The nav menu's "Demo" link under Resources still points to `/demo` (interactive product tour).

## Implementation Details

### Verification Method

Every code example was cross-referenced against:
- `infra-charts/` repo: Template syntax, Chart.yaml format, values.yaml structure
- `planton/ops/organizations/`: Real service manifests for ServiceHub
- `planton/apis/ai/planton/servicehub/service/v1/api.proto`: Service API definition
- `planton.ai/public/docs/cli.md`: Real CLI commands, install methods, version
- `openmcf/` repo: Forge workflow, apiVersion patterns, kind structure

### Files Changed

~50 files across:
- `src/components/product/{infra-hub,service-hub,runner,security,agent-fleet,cli,open-source}/` — capabilities, hero, cta files
- `src/components/product/overview/` — modules-grid, architecture, hero, cta
- `src/components/product/solutions/` — 11 solution page components, hub
- `src/components/landing-page/v3-2026-01-02-1000/HeroSection.tsx`
- `src/components/pricing/ready-to-try.tsx`
- Deleted: `src/app/(root)/solutions/by-use-case/multi-cloud/{roles.tsx,page.css}`

## Benefits

- Every YAML example on the website matches real product manifests
- Every CLI command matches real CLI docs
- Every apiVersion matches real OpenMCF conventions
- No visible placeholder content — site looks complete
- Distinct content on every page section — no duplicated animations
- Enterprise-appropriate CTAs for enterprise audience
- Consistent "Book a Demo" experience across the entire site
- Clean, focused homepage hero

## Impact

### Pages Affected

- **7 product pages** corrected (InfraHub, ServiceHub, Runner, Security, Agent Fleet, CLI, Open Source)
- **1 overview page** corrected (resource count, grid layout)
- **11 solutions pages** corrected (pricing, CTA ordering, links)
- **1 homepage hero** decluttered
- **24 CTA links** standardized to Google Forms

## Related Work

- Website copywriting overhaul (March 25) — updated content; this change corrected factual errors
- Feature pages structural overhaul (March 25) — added visual variety; this change removed placeholders and fixed duplicate content
- Black-and-white theme redesign (March 25) — visual identity; unchanged by this work

---

**Status**: ✅ Live
**Timeline**: Single session

# Connect Section and Documentation Philosophy Evolution

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation

## Summary

Created the complete Connect documentation section (8 pages) covering credential and integration management across cloud providers, git providers, container registries, state backends, Kubernetes clusters, environment authorization, and default connection resolution. Simultaneously evolved the project's documentation philosophy from "source code is the only authority" to a 3-tier model incorporating ADRs and changelogs, and established rules against exposing internal implementation details in user-facing documentation.

## Problem Statement / Motivation

Connect is a cross-cutting concern referenced by both Infra Hub and Service Hub, yet it had no dedicated documentation section. The existing `platform/connections.md` page was heavily marketing-toned ("Your Bridge to the Cloud Ecosystem"), contained fabricated information (incorrect GCP credential fields, unverified features like "Connection Health Checks" and "Usage Analytics"), and failed to document the platform's actual 17 credential types, 3 authentication modes, or the authorization and default connection resolution system.

Additionally, Sessions 1-3 of the docs overhaul had produced technically accurate but implementation-heavy documentation that exposed internal protobuf field names, enum values, and system architecture details that users don't interact with. The documentation needed a philosophical shift toward user-centric writing.

### Pain Points

- Zero documentation for the authorization model (which credentials can be used in which environments)
- Zero documentation for the default connection resolution system
- Existing connections page listed incorrect credential requirements for GCP (claimed "Project ID" was required; actual API only requires a service account key)
- No mention of cross-account trust authentication for AWS or runner-delegated authentication
- 15 out of 15 existing docs pages contained protobuf field name leakage
- 14 out of 15 existing pages lacked "why" context explaining the rationale behind features

## Solution / What's New

### Documentation Guidelines Evolution (3 files)

Established a 3-tier source truth model:
1. **Tier 1 — Current behavior**: Source code (protobuf, backend, CLI, web console)
2. **Tier 2 — Design rationale**: ADRs and co-located micro-documentation
3. **Tier 3 — Evolution context**: Changelogs

Added three new rules:
- "Documentation is for users, not implementors" — no protobuf leakage
- "Co-located micro-documentation is a primary context source" — check README files alongside source code
- Role division — project owner provides platform context, documentation writer owns craft decisions

Captured a retrospective audit of all 15 existing pages with severity rankings for a future remediation session.

### Connect Section (8 pages, ~1,800 lines)

| Page | What It Covers |
|------|---------------|
| `connect/index.md` | Overview, four integration categories, three authentication modes, authorization and defaults concept, managed services taxonomy |
| `connect/cloud-providers.md` | AWS (3 auth methods including CloudFormation Quick Create), GCP, Azure, DigitalOcean, Civo, Cloudflare |
| `connect/git-providers.md` | GitHub OAuth App Install flow, GitLab access tokens, self-hosted support |
| `connect/container-registries.md` | Docker (5 registry providers), NPM (3 providers), Maven (3 providers), Cloudflare Wrangler R2 |
| `connect/state-backends.md` | Pulumi backends (4 types), Terraform backends (3 types), R2 state backend story |
| `connect/kubernetes-clusters.md` | GKE (complete), DOKS (complete), EKS and AKS (placeholder status documented honestly) |
| `connect/environment-mappings.md` | Authorization scopes, Provider Matrix, CLI commands, common patterns |
| `connect/default-connections.md` | Resolution algorithm, two-level defaults, resolve command, troubleshooting |

### Old Page Removed

Deleted `platform/connections.md` (11.5KB of marketing-heavy content) and updated the cross-reference in `platform/resource-hierarchy.md`.

## Implementation Details

All content was sourced from three tiers:

- **17 protobuf credential APIs** in `apis/ai/planton/connect/` — used for accuracy, never exposed to users
- **5 ADRs** — DefaultProviderConnection, ProviderConnectionAuthorization, CloudFormation AWS, auth modes, default runner binding
- **20+ changelogs** — AWS wizard phases 1-6, connection environments, authorization API
- **Co-located micro-docs** — `backend/services/connect/docs/` (security, quick-reference, downstream-services)
- **Web console** — Mission Control layout categories from `utils.ts`, wizard components, Provider Matrix
- **CLI commands** — `planton connect aws`, `planton connection authorization *`, `planton connection default *`

The Connect pages are the first exemplar of the revised documentation philosophy. They contain zero protobuf field name leakage, lead with "why" context from ADRs, and describe workflows from the user's perspective.

## Benefits

- **8 new pages** covering a previously undocumented cross-cutting concern
- **Complete CLI reference** for connection authorization and default management
- **Honest documentation** — EKS/AKS placeholder status, GitLab "coming soon" status documented accurately instead of speculatively
- **Quality exemplar** — Connect pages serve as the reference for the retrospective audit of Sessions 1-3 pages
- **Evolved guidelines** prevent future documentation from exposing internal implementation details

## Impact

- Connect section fills the largest gap in the documentation (8 pages covering 17 credential types, authorization, and defaults)
- Old marketing-heavy connections page replaced with source-verified, user-centric content
- Guidelines evolution affects all future documentation work and queues a retrospective for existing pages
- Total docs site now has ~40 pages across 5 sections (Root, Platform, Connect, Infra Hub, Service Hub)

## Related Work

- Session 1 (`CP01`): Quality fixes across 25 existing pages
- Session 2 (`CP02`): 6 new Infra Hub pages
- Session 3 (`CP03`): 2 rewritten + 6 new Service Hub pages
- Next: Phase 5 (Cloud Ops), Phase 6 (Runner), or Phase 4.5 (retrospective audit)

---

**Status**: Live
**Timeline**: Single session

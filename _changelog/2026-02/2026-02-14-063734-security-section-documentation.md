# Security Section Documentation

**Date**: February 14, 2026
**Type**: Feature
**Components**: Documentation

## Summary

Created the Security documentation section — 3 new pages providing an enterprise security overview, authentication and authorization deep dive, and audit trails documentation. This is the first documentation of the audit system (a shipped feature with zero prior coverage) and the first unified security narrative tying together credential isolation, encryption, authorization, and change tracking across the platform.

## Problem Statement / Motivation

The Planton documentation covered individual security features across scattered sections — credential isolation in Runner, encryption in Secrets and Config, role-based access in Teams and Access, authentication modes in Connect — but lacked a unified security narrative. An enterprise security reviewer or CISO had no single page to evaluate the platform's security posture. Additionally, the audit system (immutable version records, diffs, web console UI) was a fully shipped feature with zero documentation.

### Pain Points

- No centralized security overview for enterprise evaluation
- Audit trails (immutable version records, state snapshots, diffs) completely undocumented
- Authentication model (OAuth, PKCE, API keys) documented only in Teams and Access as a list of CLI commands, without explaining how the system works
- Authorization model (OpenFGA, relationship-based access, permission inheritance) mentioned but never explained from the user's perspective
- Security-related pages in Runner, Connect, Secrets, and Teams sections had no cross-references to each other

## Solution / What's New

### 3 New Security Pages

**`security/index.md`** — Enterprise security overview. The "one page a CISO reads." Covers credential isolation, envelope encryption, authentication, authorization, audit trails, and network security at a summary level with links to detailed pages in every section.

**`security/authentication-and-authorization.md`** — How authentication and authorization actually work. Explains OAuth + PKCE for interactive login, API key scoping and hashed storage, OpenFGA-based relationship access control, five standard roles, hierarchical permission inheritance (organization → environment → resource), team-based access propagation, resource sharing model, and context switching. Includes full CLI reference for IAM commands.

**`security/audit-trails.md`** — First documentation of the audit system. Covers what gets captured (full YAML snapshots, unified diffs, who/when, event types, Infra Job linkage), immutability guarantees, infrastructure manifest diffs for Infra Components, web console UX (versions list, detail view, side-by-side diff), and practical uses (debugging, compliance, rollback investigation).

### 6 Cross-Reference Updates

- `docs/index.md` — Added Security section between Runner and Teams and Access
- `platform/index.md` — Added Security to Platform Sections list
- `teams-and-access/index.md` — Added Security links to Related Documentation
- `runner/security-model.md` — Added Security Overview to Related Documentation
- `secrets-and-config/index.md` — Added Related Documentation section with Security link
- `connect/index.md` — Added Security Overview to Related Documentation

## Implementation Details

### Scope Decisions

Three craft decisions shaped the final scope:

1. **Deployment security tiers skipped** — The original information architecture included `deployment-security-tiers.md` (Basic, Trusted, Customer-hosted). Source code exploration confirmed this concept does not exist in any protobuf, backend, CLI, or web console code. Per the no-speculation rule, the page was not created.

2. **No standalone credential-isolation page** — Credential isolation is thoroughly documented in `runner/security-model.md`, `connect/cloud-providers.md`, and `secrets-and-config/secret-backends.md`. Instead of duplicating this content, the Security index page covers it as a summary section with cross-references.

3. **3 pages instead of 4** — The original plan called for 4 pages. With deployment security tiers removed and credential isolation consolidated into the index, the section is 3 pages focused on genuine, source-verified content.

### Source Verification

All content verified against:

- **19 OpenFGA model files** in `backend/services/iam/src/main/resources/fga/model/` — organization, environment, infra_component, service, credential_resource, team, identity_account, and more
- **IAM role SQL seed** (`V6__insert_roles_initial_data.sql`) — verified role names: owner, admin, iam_admin, viewer, member
- **Audit protobuf APIs** (`apis/ai/planton/audit/apiresourceversion/v1/`) — ApiResourceVersion, CatalogObjectVersion, query RPCs
- **Audit service architecture** (`backend/services/audit/docs/architecture.md`) — event-driven via NATS, MongoDB storage, append-only
- **11 CLI command files** — authentication (login, who, list, use), authorization (iampolicy add/get/remove, iamrole list), API keys (new, list)
- **Web console audit pages** — version list, detail, diff kinds with side-by-side/line-by-line toggle
- **Web console IAM kinds** — grant permission, edit permission, API key management
- **7 ADRs** — envelope encryption, CMEK providers, secret backend abstraction, provider connection auth modes, CLI authentication redesign, runner deployment, permission denied handling
- **Runner IP preservation guidelines** applied throughout — no named technologies, no internal endpoints, security guarantees stated as outcomes

### Runner IP Compliance

The Security index page describes Runner's security properties at a conceptual level:
- No mention of Konnectivity, Temporal, or Go
- No internal endpoints, ports, or channel IDs
- No kind names (Tunnel Agent, gRPC Server)
- mTLS described as an outcome ("both sides verify identity"), not a mechanism

## Benefits

- **Enterprise evaluation enabled** — Security reviewers have a single starting point for evaluating the platform
- **Audit trails documented** — A shipped feature now has user-facing documentation for the first time
- **Authorization model explained** — Users can now understand *how* permissions work, not just what roles exist
- **Cross-cutting security narrative** — The security story is unified across all product areas with consistent cross-references
- **Documentation total: 53 pages across 9 sections** (up from 50 across 8)

## Impact

- 3 new documentation pages (security/index, authentication-and-authorization, audit-trails)
- 6 existing pages updated with cross-references
- 7 screenshot placeholders added for future capture
- 0 broken links (verified all cross-references)
- 0 Runner IP violations (verified against preservation guidelines)

## Related Work

- Phase 1-16 of the Planton docs overhaul (sessions 1-16)
- Phase 15/15b mobile responsiveness (sessions 17-18)
- This is the final content section (Phase 7b) — all planned documentation sections now exist

---

**Status**: Live
**Timeline**: Single session

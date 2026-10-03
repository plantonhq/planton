# Platform Section Deep Content Rewrite

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation, Platform Pages

## Summary

Deep content pass on the platform documentation section and root getting-started page. Rewrote 4 pages (3 clean-slate, 1 targeted) to match the quality bar established by the Connect section exemplar. Removed all "Planton" references, marketing language, emoji, unverified claims, false field names, and hardcoded legacy images. All content verified against source code: protobuf APIs, web console routes, CLI commands, and IAM role database seed.

## Problem Statement / Motivation

The platform section was the least-touched since the Phase 1 quality fixes at the start of Week 1. While Sessions 2-12 brought all other sections to a high quality bar, the platform pages — the first thing new users read — still contained:

### Pain Points

- 15+ "Planton" references across 3 pages (platform name is "Planton")
- Marketing language: "really running, really fast", "infrastructure shopping mall", "navigate like a pro", "cloud chaos", "you can't break anything"
- Emoji checkmarks in getting-started summary
- Beta messaging ("Welcome to Beta", "platform is in beta") that is outdated
- Hardcoded `assets.planton.ai` image URLs instead of standard screenshot placeholders
- False claim: environment creation form has a "Type" field (EnvironmentSpec only has `description`)
- Unverified time claims: "10 minutes", "2-3 minutes for a VPC"
- Incorrect permission role names: "Developer" (does not exist), missing "IAM Admin" role
- Author blocks with marketing bios on all 4 pages
- Root getting-started.md had duplicated "Key Concepts at a Glance" section
- Platform tour referenced features that don't exist: "audit logs" and "security policies" in Settings

## Solution / What's New

### platform/getting-started.md (Clean-Slate Rewrite)

Restructured around the actual 8-task onboarding checklist from `getting-started-constants.ts`. Each step verified against source code: organization creation wizard route (`/organizations/new/setup`), `EnvironmentSpec` proto (no type field), Connections page layout, and Catalog Kind Store route (`/platform/deployment-store`). Added CLI commands from verified Go source.

### platform/platform-tour.md (Clean-Slate Rewrite)

Mapped directly from `routes/index.ts` — all 7 sidebar sections documented with verified sub-pages and tabs. Added Agent Fleet (omitted from old docs) and IaC Module Registry (verified in header-actions.tsx). Documented Service Hub's 7 tabs, Settings' 3 tabs (General, Manage Members, Teams), and the correct sidebar structure. Removed false claims about audit logs and security policies in Settings.

### platform/resource-hierarchy.md (Targeted Rewrite)

Corrected role names from the IAM role database seed (`V6__insert_roles_initial_data.sql`): 5 organization roles (owner, admin, iam_admin, member, viewer) and 3 environment roles (admin, iam_admin, viewer). The old page claimed "Developer" as an environment role — this does not exist. Added OpenFGA mention for open-source credibility. Removed marketing language and emoji while preserving the good structural bones.

### Root getting-started.md (Clean-Slate Rewrite)

Transformed from a 117-line marketing-heavy docs hub (with duplicated sections, author block, and "Planton" title) into a focused 56-line navigation page organized by user goal: deploy infrastructure, deploy applications, or understand the platform.

## Implementation Details

### Source Verification

| Page | Sources Verified |
|------|-----------------|
| getting-started.md | `OrganizationSpec` proto, `EnvironmentSpec` proto, `getting-started-constants.ts`, `routes/index.ts`, CLI `context` and `env` commands |
| platform-tour.md | `routes/index.ts` (7 sidebar routes), `header-actions.tsx` (IaC Module Registry tooltip), `settings.tsx` (3 tabs), `agent-fleet/page.tsx`, `search/v1/resourcemanager/query.proto` |
| resource-hierarchy.md | `EnvironmentSpec` proto (no type field), `V6__insert_roles_initial_data.sql` (actual role names), CLI `env` and `context` commands |
| root getting-started.md | All section index pages (verified link targets) |

### Key Corrections

- Environment creation: removed false "Type" field (EnvironmentSpec only has `description`)
- Organization creation route: corrected to `/organizations/new/setup` (was `/orgs/create`)
- Settings tabs: corrected to General, Manage Members, Teams (was claiming audit logs, security policies)
- Organization roles: corrected to owner, admin, iam_admin, member, viewer
- Environment roles: corrected to admin, iam_admin, viewer (removed false "Developer" role)
- Agent Fleet: now documented (was omitted entirely from platform tour)
- IaC Module Registry: now documented (confirmed via `header-actions.tsx`)

## Benefits

- All 4 pages now meet the quality bar established by the Connect section
- Zero "Planton" references remain in the platform section
- All role names match the actual IAM database seed
- All console routes and features are verified against source code
- New users get accurate, non-marketing onboarding documentation

## Impact

- **4 files changed**: +411 / -905 lines (net reduction of 494 lines — cutting filler)
- **1 commit**: `6d03099`
- **0 broken cross-references** verified across all docs
- Platform section is now consistent with the quality of Connect, Cloud Ops, Runner, and Secrets sections

## Related Work

- Session 1 (Phase 1) quality fixes established the initial cleanup
- Session 4 (Phase 4) Connect section established the quality exemplar
- Session 9-10 (Phase 4.5) retrospective audit of Infra Hub and Service Hub pages
- Session 11 (Phase 8+9) Teams and Access extraction verified actual role names

---

**Status**: Live
**Timeline**: Session 13 of the Planton Docs Overhaul project (Week 2, Day 1)

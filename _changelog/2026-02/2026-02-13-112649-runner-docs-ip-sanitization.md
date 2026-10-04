# Runner Documentation IP Sanitization

**Date**: February 13, 2026
**Type**: Refactoring
**Components**: Documentation

## Summary

Rewrote the entire Runner documentation section to remove intellectual property details while strengthening the enterprise trust narrative. Consolidated from 5 pages to 3 pages. All references to internal technologies (Konnectivity, Temporal), endpoints, ports, protocol details, component names, and architectural internals have been removed. The docs now focus exclusively on what enterprise security-conscious customers need to evaluate and trust the Runner model.

## Problem Statement / Motivation

The original Runner documentation (5 pages, ~860 lines) was written with deep source-code verification, which inadvertently exposed implementation details that constitute intellectual property:

### Details Exposed

- **Konnectivity** named 3 times with direct links to Kubernetes documentation
- **Temporal** named as the job execution engine with polling and task queue details
- Internal tunnel endpoint: `runner-tunnel.planton.live:443`
- Internal gRPC port: `50051` (6 references across 3 files)
- Channel identifier format: `org.{org-slug}.runner.{runner-slug}`
- Virtual hostname routing: `{channel-id}.tunnel.local:50051`
- HTTP CONNECT protocol details
- Three-component architecture breakdown (Tunnel Agent, gRPC Server, Temporal Worker)
- Execution modes (`grpc`, `temporal`, `dual`) with `--mode` flag
- Local credential storage paths
- Certificate CN validation implementation

None of this helps an enterprise customer decide to trust Runner. It helps a competitor understand how to replicate it.

## Solution / What's New

### Pages Rewritten (3)

- **`runner/index.md`** — Expanded overview absorbing conceptual content from deleted pages. Covers: why Runner exists, what it does (cloud operations + infrastructure deployments), how connectivity works (outbound-only, mTLS, auto-reconnect), network requirements, getting started. High-level Mermaid diagram shows only User → Planton → Secure Connection → Runner → Cloud API.

- **`runner/deployment.md`** — Registration, deployment to 4 targets, default runner binding, management. Removed: local credential storage paths, `--mode` flag, `--temporal-address` flag, `config.yaml` internal fields.

- **`runner/security-model.md`** — Credential isolation, mTLS (high-level), 3 authentication modes, organization-scoped identity (conceptual), trust boundaries. Removed: port numbers, channel ID format, CN validation implementation, `127.0.0.1:50051`.

### Pages Deleted (2)

- **`runner/architecture.md`** — Entirely internal implementation detail (component names, execution modes, Go binary design, request flow diagrams showing internal routing, channel identification system).

- **`runner/runner-tunnel.md`** — Konnectivity references, tunnel endpoint URLs, HTTP CONNECT protocol, virtual hostname format, "Why Temporal Uses a Separate Path" section.

### Cross-References Updated

- `cloud-ops/index.md` — Updated 2 links from `/docs/runner/runner-tunnel` to `/docs/runner`

### IP Preservation Rule Created

- Cursor rule at `planton/docs/_rules/sanitize-planton-runners-implementation-in-public-docs.mdc`
- Project guideline at `coding-guidelines/runner-ip-preservation.md`
- Includes two decision tests: "Competitor test" (would this help replicate?) and "Enterprise test" (would a security reviewer need this?)
- Pre-publish checklist for all future Runner content

## Content Transformations

| Original | Sanitized |
|----------|-----------|
| "Konnectivity" with K8s docs link | "Planton's proprietary reverse tunnel" |
| Port 50051 / endpoint URLs | Removed entirely |
| Channel ID format | "unique cryptographic identity bound to your organization" |
| Tunnel Agent / gRPC Server / Temporal Worker | "cloud operations" / "infrastructure deployments" |
| `--mode grpc\|temporal\|dual` | "can be configured for cloud operations, infrastructure execution, or both" |
| Temporal | "Planton's job execution system" |
| Mermaid diagrams with internal routing | High-level flow: User → Planton → Secure Connection → Runner → Cloud API |

## Benefits

- Zero IP-sensitive terms remain in any published documentation
- Enterprise security reviewers get everything they need: credential isolation guarantees, mTLS, auth modes, trust boundaries, network requirements
- Simpler documentation structure (3 pages vs 5) — easier for customers to evaluate
- Permanent IP preservation rule prevents future regressions

## Impact

- Runner docs reduced from ~860 lines / 5 pages to ~490 lines / 3 pages
- All internal technology names, endpoints, ports, and protocol details removed
- Cross-references updated across cloud-ops section
- Future documentation work protected by cursor rule

## Related Work

- Previous: [Runner Documentation Section](2026-02-13-110603-runner-documentation-section.md) (original 5-page creation)
- Previous: [Cloud Ops Documentation Section](2026-02-13-103224-cloud-ops-documentation-section.md) (cross-reference source)

---

**Status**: Live
**Timeline**: 1 session

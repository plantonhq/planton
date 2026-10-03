# Cloud Ops Documentation Section

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation

## Summary

Created the Cloud Ops documentation section (3 new pages) and rewrote the Service Hub Kubernetes Dashboard page. Cloud Ops — Planton's real-time operations gateway for Kubernetes, AWS, GCP, and Azure — previously had zero documentation on the public docs site. This phase brings the section to full coverage.

## Problem Statement / Motivation

Cloud Ops is a shipped, daily-use feature that lets developers and operators inspect pods, stream logs, exec into containers, and browse infra components — all without distributing credentials. Despite being one of Planton's most distinctive capabilities (dual-mode access, credential-free operations via Runner Tunnel), it had no documentation whatsoever.

Additionally, the existing `service-hub/kubernetes-dashboard.md` page described Cloud Ops features but was marketing-heavy, referenced "Planton" inconsistently, included unverified "Coming Soon" sections for ECS and Cloud Run dashboards, and contained fabricated troubleshooting scenarios.

### Pain Points

- Zero documentation for a daily-use feature (Cloud Ops)
- Existing related page (kubernetes-dashboard.md) was low quality with marketing language and unverified claims
- The Connect section's kubernetes-clusters page already linked to `/docs/cloud-ops` which didn't exist yet
- No CLI reference for `planton kubectl`, `planton aws`, `planton gcp`, or `planton azure` commands

## Solution / What's New

### 3 New Cloud Ops Pages

**`cloud-ops/index.md`** — What Cloud Ops is, why it exists, how it works. Covers the day-2 operations gap, dual access modes (developer mode and admin mode), supported providers, and the tunnel architecture. Written as the entry point for the entire Cloud Ops section.

**`cloud-ops/kubernetes-operations.md`** — Comprehensive Kubernetes operations reference. Covers pod viewing, log streaming with filters, browser-based container exec, resource browsing with DAG visualization, resource editing and deletion. Includes full CLI reference with all flags documented from Go source.

**`cloud-ops/resource-browser.md`** — Multi-infra component browsing for AWS (EC2, S3), GCP (Compute Engine, Cloud Storage), and Azure (VMs, Blob Storage). Complete CLI reference for all 8 provider commands with exact flags, filter syntax, and connection resolution.

### 1 Page Rewritten

**`service-hub/kubernetes-dashboard.md`** — Rewritten from scratch as a focused developer-experience page: "you deployed a service, here's how to check its pods, logs, and get a shell." Cross-references Cloud Ops for the full operations reference. Removed marketing language, "Planton" references, unverified "Coming Soon" sections, author attribution, and transcript citation.

## Implementation Details

### Source Verification

All content verified against four source code layers:

- **Protobuf APIs**: 38 files in `apis/ai/planton/cloudops/` — 7 Kubernetes service definitions, 3 cloud provider modules
- **CLI commands**: 16 commands verified from Go source in `client-apps/cli/cmd/planton/root/kubectl/` and `client-apps/cli/cmd/planton/root/domain/cloudops/`
- **Web console**: Pod list, exec drawer, log viewer, terminal kinds in `client-apps/web/console/src/components/shared/kubernetes-resources/` and `src/services/cloud-ops/kubernetes/`
- **ADRs**: 4 architectural decision records covering dual access mode, API segregation, tunnel routing, and runner architecture
- **Co-located docs**: `apis/ai/planton/cloudops/README.md`, `apis/ai/planton/cloudops/docs/routing-architecture.md`, `apis/ai/planton/cloudops/docs/runner-authentication.md`

### Quality Standards

- Zero protobuf field names, message types, or RPC names in user-facing prose
- Zero marketing language or emoji
- Zero unverified claims or "coming soon" sections
- All CLI examples use exact command names and flag names from Go source
- Content follows the revised documentation philosophy (why/what/user-how ordering)
- Connect section pages serve as the quality exemplar

### Cross-References

- `connect/kubernetes-clusters.md` already linked to `/docs/cloud-ops` — now resolves correctly
- `service-hub/kubernetes-dashboard.md` cross-references `cloud-ops/kubernetes-operations.md`
- All new pages cross-reference each other and related sections (Connect, Runner)

## Benefits

- Cloud Ops goes from 0 to 3 pages of comprehensive documentation
- Previously broken link from Connect section now resolves
- kubernetes-dashboard.md quality raised from low (marketing-heavy) to high (source-verified)
- Complete CLI reference for 16 Cloud Ops commands published for the first time
- Dual access mode concept clearly documented for both developers and admins

## Impact

- **Docs site**: 3 new pages added, 1 page rewritten. Cloud Ops section visible in sidebar at position 50.
- **Total docs pages**: Approximately 36 (up from 33 after Phase 4)
- **Cross-references**: 1 previously broken link now resolves, 8+ new cross-references added

## Related Work

- [Phase 4: Connect section and guidelines evolution](2026-02-13-100000-connect-section-and-docs-philosophy-evolution.md)
- [Phase 3: Service Hub documentation overhaul](2026-02-13-090147-service-hub-documentation-overhaul.md)
- Planton docs overhaul project: `planton/_projects/20260212.02.planton-docs-overhaul/`

---

**Status**: Live
**Timeline**: 1 session

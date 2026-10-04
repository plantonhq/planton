# Runner Documentation Section

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation

## Summary

Created the complete Runner documentation section — 5 new pages covering the secure execution agent that enables Planton to operate in customer infrastructure without credential sharing or inbound firewall rules. All content verified against protobuf APIs, 13 ADRs, CLI source code, backend service implementations, and web console components.

## Problem Statement / Motivation

Runner is one of Planton's most architecturally distinctive features — the component that makes the platform's security story credible for enterprise adoption. It enables infrastructure-as-code execution and real-time cloud operations without credentials leaving customer infrastructure, using outbound-only mTLS tunnel connectivity. Despite being critical for enterprise sales conversations and central to the platform's security model, Runner had zero documentation.

### Pain Points

- No public documentation for the Runner product area
- Enterprise prospects and security teams had no reference material for evaluating Runner's security model
- The mTLS tunnel architecture, credential isolation model, and deployment options were undocumented
- Cloud Ops and Connect sections referenced Runner but had nowhere to link for details
- Existing Cloud Ops and Connect pages already mentioned Runner-delegated authentication but readers had no way to understand what that meant

## Solution / What's New

### 5 New Documentation Pages

1. **Runner Overview** (`runner/index.md`) — What Runner is, why it exists, the fundamental problem it solves (credentials + network access), the three components at a high level, and navigation to detailed pages.

2. **Architecture** (`runner/architecture.md`) — Single-binary design with three components (Tunnel Agent, gRPC Server, Temporal Worker), execution modes (`grpc`, `temporal`, `dual`), request flow diagrams for both Cloud Ops and IaC paths, and channel identification system.

3. **Deployment** (`runner/deployment.md`) — Complete lifecycle from registration through production deployment. Covers CLI registration with credential output formats, local credential storage, binary installation, local development mode, four deployment targets (Kubernetes/Helm, AWS ECS, GCP Cloud Run, Azure Container Apps), web console wizard, default runner binding with resolution chain, and runner management commands.

4. **Runner Tunnel** (`runner/runner-tunnel.md`) — How the reverse tunnel enables outbound-only connectivity. Covers the networking problem, Konnectivity foundation, connection establishment with mTLS, request routing flow, what flows through the tunnel vs. what doesn't (and why Temporal uses a separate path), network requirements, and resilience behavior.

5. **Security Model** (`runner/security-model.md`) — Credential isolation principle, mTLS certificate chain, credential lifecycle (generation, storage, rotation), three authentication modes (inline, runner-delegated, cross-account trust), organization-scoped identity with CN enforcement, and explicit documentation of what the runner cannot do.

### Cross-Reference Updates

- Updated Cloud Ops index to link to `/docs/runner/runner-tunnel` for the specific tunnel page (previously linked to `/docs/runner` generically)
- Added Runner Tunnel to Cloud Ops Related Documentation section

## Implementation Details

### Source Verification

Every claim in all 5 pages was verified against the actual source code:

- **Protobuf APIs**: `RunnerRegistration` (5 files), `DefaultRunnerBinding` (5 files), `RunnerCredentialsBundle`, `ProviderConnectionAuthMode` enum (4 values), 24 Runner CloudAPI proto files
- **ADRs**: 13 architectural decision records from Jan-Feb 2026
- **CLI source** (Go): 12 subcommands under `planton runner` with all flags, the `runnerdeploy` package with 4 deployment targets, the `runnerconfig` package with local credential storage layout
- **Backend** (Java): `RunnerRegistrationGenerateCredentialsHandler` (channel ID format, private key handling), `DefaultRunnerBindingResolver` (resolution chain with Redis cache)
- **Konnectivity patches**: `validateRunnerIdentifierCN()` in `pkg/server/server.go` (CN validation)
- **Runner service** (Go): `config.go` (default port 50051, execution modes), `operator_reference.go` (all flags and defaults)

### Key Findings During Verification

- **gRPC port is 50051, not 8080**: Product docs were incorrect. Source code explicitly states "MUST be 50051 -- CloudOps routes to runners via the tunnel targeting this port."
- **Channel ID format confirmed**: `org.{org-slug}.runner.{runner-slug}` (not the `runner/{org}/{slug}` format from an earlier ADR)
- **Private key not stored**: Confirmed in both proto comments and Java handler — generated server-side, returned once, never persisted
- **CN validation exists**: Konnectivity patch adds `validateRunnerIdentifierCN()` that cryptographically enforces identity

### Content Architecture Decisions

- **Security boundary**: Runner section covers runner-specific security in depth (mTLS, credential isolation, auth modes). The future Security section (Phase 7) will provide the cross-cutting enterprise view and reference these pages for details.
- **Deployment depth**: Full per-target treatment rather than overview-only. A DevOps engineer reading the deployment page gets enough to actually deploy — without the page becoming a 20-page tutorial.
- **Tunnel abstraction level**: Explains the tunnel at the conceptual and operational level. Konnectivity internals (patching strategy, port assignments, proxy protocols) are implementation details omitted from user-facing documentation.

### Files Changed

- `public/docs/runner/index.md` — New (Runner overview)
- `public/docs/runner/architecture.md` — New (Architecture deep dive)
- `public/docs/runner/deployment.md` — New (Deployment lifecycle)
- `public/docs/runner/runner-tunnel.md` — New (Tunnel networking)
- `public/docs/runner/security-model.md` — New (Security model)
- `public/docs/cloud-ops/index.md` — Updated cross-references

## Benefits

- **Enterprise readiness**: Security teams and architects now have comprehensive reference material for evaluating Runner's security model
- **Self-service deployment**: DevOps engineers can register and deploy runners without internal support, using the deployment page as a complete guide
- **Cross-reference integrity**: Cloud Ops and Connect pages that reference Runner now have detailed landing pages to link to
- **Documentation completeness**: Runner section fills the largest remaining gap in the platform documentation

## Impact

- 5 new documentation pages added to the Runner section
- 1 existing page updated (Cloud Ops index cross-references)
- Total Runner section: ~1,100 lines of verified, source-grounded documentation
- Runner section is now the 7th of 9 planned sections to be completed

## Related Work

- [Cloud Ops Documentation Section](2026-02-13-103224-cloud-ops-documentation-section.md) — Session 5, which created the Cloud Ops section that Runner references
- [Connect Section and Docs Philosophy Evolution](2026-02-13-100000-connect-section-and-docs-philosophy-evolution.md) — Session 4, which established the quality exemplar and revised documentation philosophy
- Phase 7 (Security section) and Phase 4.5 (Retrospective audit) are the next planned phases

---

**Status**: Live
**Timeline**: Single session (~2 hours)

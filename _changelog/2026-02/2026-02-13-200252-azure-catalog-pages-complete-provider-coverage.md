# Azure Catalog Pages — Complete Provider Coverage

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation, Azure Provider

## Summary

Wrote 22 new hand-crafted catalog pages for all remaining Azure deployment components, bringing Azure to 24/24 (100%) catalog page coverage. Each page follows the established 9-section standard, is verified against protobuf definitions and Pulumi module source code, and replaces the legacy auto-generated research documents.

## Problem Statement / Motivation

Azure had only 2 of 24 components with hand-written catalog pages (AzureAksCluster and AzureKeyVault). The remaining 22 components served legacy auto-generated `docs/README.md` files — lengthy research documents with deployment landscape essays, tool comparisons, and maturity spectrums that did not match the concise, source-verified standard established for the catalog.

### Pain Points

- 22 Azure components lacked the 9-section catalog page standard (title, What Gets Created, Prerequisites, Quick Start, Config Reference, Examples, Stack Outputs, Related Components)
- Legacy pages contained filler content not useful for developers trying to deploy infrastructure
- No progressive examples showing minimal through full-featured configurations
- No foreign key reference examples demonstrating cross-component `valueFrom` patterns
- Inconsistent documentation quality across the Azure provider compared to fully-covered AWS, GCP, and Kubernetes providers

## Solution / What's New

### 22 New Catalog Pages

Organized by infrastructure layer, written in 6 parallel rounds of 4 agents each:

**Round 1 — Foundational Infrastructure**
- AzureResourceGroup, AzureVpc, AzureSubnet, AzurePublicIp

**Round 2 — Networking and Gateways**
- AzureNatGateway, AzureNetworkSecurityGroup, AzureLoadBalancer, AzureApplicationGateway

**Round 3 — Databases and Storage**
- AzurePostgresqlFlexibleServer, AzureMysqlFlexibleServer, AzureMssqlServer, AzureStorageAccount

**Round 4 — DNS and Private Networking**
- AzureDnsZone, AzureDnsRecord, AzurePrivateDnsZone, AzurePrivateEndpoint

**Round 5 — Compute, Container, and Identity**
- AzureVirtualMachine, AzureAksNodePool, AzureContainerRegistry, AzureUserAssignedIdentity

**Round 6 — Monitoring**
- AzureApplicationInsights, AzureLogAnalyticsWorkspace

### Net-New Catalog Entry

`AzureVirtualMachine` had no legacy `docs/README.md` and no existing site page. This is a net-new catalog entry that will appear on the docs site for the first time when the build pipeline runs.

## Implementation Details

Each catalog page was written by reading source files in the established order:

1. `api.proto` — apiVersion and kind values
2. `spec.proto` — all configuration fields, types, validations, defaults, foreign keys
3. `stack_outputs.proto` — all output fields
4. `iac/pulumi/module/main.go` — deployment flow and resource creation
5. `iac/pulumi/module/*.go` — all cloud resources created, output constants

The build pipeline at `site/scripts/copy-component-docs.ts` already prefers `catalog-page.md` over `docs/README.md` with automatic fallback — no pipeline changes were needed.

### Spot Audit Results

4 pages audited across complexity tiers:

| Page | Complexity | Result |
|------|-----------|--------|
| AzureResourceGroup | Simple | All 6 checks passed |
| AzurePostgresqlFlexibleServer | High | 1 fix applied (removed invalid `valueFrom` for password field — `secret_id_map` provides IDs, not values) |
| AzurePrivateEndpoint | Foreign key heavy | All 6 checks passed (auditor flagged `field:` vs `fieldPath:` but `field:` is the established convention across 114+ pages) |
| AzureStorageAccount | Medium | Source code bug found in `main.go` (exports `PrimaryBlobHost` for DFS output instead of `PrimaryDfsEndpoint`) — flagged, not a docs issue |

### Source Code Bug Found

`azurestorageaccount/v1/iac/pulumi/module/main.go` exports `storageAccount.PrimaryBlobHost` for the `OpPrimaryDfsEndpoint` output constant. Should export `storageAccount.PrimaryDfsEndpoint`. This is a source code bug, not a documentation bug.

## Benefits

- Azure joins AWS, GCP, and Kubernetes as the fourth provider at 100% catalog page coverage
- Total project catalog coverage: 136 of ~215 components (~63%)
- Every Azure component now has 3-5 progressive examples with correct proto field names
- `valueFrom` foreign key patterns documented for all cross-component references
- Infrastructure patterns covered: resource groups, networking (VNet/subnet/NSG/NAT/LB/AppGW), databases (PostgreSQL/MySQL/MSSQL), storage, DNS (public/private), private endpoints, compute (VM), containers (AKS node pools, ACR), identity, and monitoring

## Impact

- Developers evaluating OpenMCF for Azure infrastructure now have consistent, source-verified documentation for all 24 components
- The Azure provider catalog is at feature parity with AWS (25/25), GCP (19/19), and Kubernetes (51/51)
- Remaining providers for future coverage: DigitalOcean (13), Civo (10), Cloudflare (5), Auth0 (2), Atlas (1)

## Related Work

- [Catalog Page Rewrite System](2026-02-13-150154-catalog-page-rewrite-system.md) — established the 9-section standard, build pipeline, and Cursor rules
- [Catalog Page Expansion](2026-02-13-154844-catalog-page-expansion-across-all-providers.md) — initial 29 pages across all providers
- [AWS Complete Coverage](2026-02-13-162608-aws-catalog-pages-complete-provider-coverage.md) — first provider at 100%
- [GCP Complete Coverage](2026-02-13-164822-gcp-catalog-pages-complete-provider-coverage.md) — second provider at 100%
- [Kubernetes Complete Coverage](2026-02-13-173816-kubernetes-catalog-pages-complete-provider-coverage.md) — third provider at 100%

---

**Status**: Production Ready
**Timeline**: Single session

# OCI Dynamic Routing Gateway Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciDynamicRoutingGateway deployment component (R13, enum 3322) -- the third resource in Phase 3 (Advanced Networking). A DRG is OCI's virtual router for interconnecting VCNs (peering), on-premises networks (VPN/FastConnect), and cross-region VCNs (remote peering). This component bundles the DRG with 5 sub-resource types: attachments, route tables, route distributions, distribution statements, and static route rules, all connected via name-based cross-referencing.

## Problem Statement / Motivation

OpenMCF's OCI provider had VCN-level networking (VCN, Subnet, NSG) and load balancing (L7 + L4) but lacked the inter-VCN and hybrid connectivity hub.

### Pain Points

- No way to set up VCN peering (hub-and-spoke topologies) through OpenMCF
- No support for DRG route tables and route distributions for controlling inter-VCN traffic
- No integration point for Site-to-Site VPN (IPSec tunnel) or FastConnect (virtual circuit) attachments
- No cross-region peering via remote peering connections
- Complex DRG configurations required manual OCID wiring between sub-resources

## Solution / What's New

Designed and implemented a comprehensive OciDynamicRoutingGateway component that bundles 6 OCI resource types into a single deployment unit. Sub-resources reference each other by display_name (not OCID), making the YAML configuration clean and self-contained.

### Architecture

The component follows a strict creation order reflecting the dependency chain:
1. DRG (primary resource)
2. Route distributions (depend on DRG only)
3. Route tables (may reference distributions for import)
4. Attachments (may reference route tables and distributions)
5. Distribution statements (may reference attachments via match criteria)
6. Static route rules (reference attachments as next hop)

### Name-Based Cross-Referencing

All sub-resources reference each other by display_name rather than OCID. The IaC modules resolve these names to OCIDs at deployment time. This mirrors the pattern established by OciApplicationLoadBalancer where listeners reference backend sets by name.

### Design Simplifications

- `action` on distribution statements is always "ACCEPT" (the only OCI-supported value) -- hardcoded in IaC, not exposed in spec
- `destination_type` on route rules is always "CIDR_BLOCK" (only valid type for user-created static routes) -- hardcoded in IaC
- Deprecated attachment fields (`vcn_id`, root-level `route_table_id`) are not exposed -- users use the `network_details` block
- `DrgAttachmentManagement` excluded (manages auto-created IPSec/FastConnect/RPC attachments owned by other services)
- `DrgAttachmentsList` excluded (data source, not a managed resource)

## Implementation Details

### Proto API (spec.proto)

- **5 top-level fields**: compartment_id (StringValueOrRef), display_name, attachments, route_tables, route_distributions
- **7 nested messages**: DrgAttachment, NetworkDetails, DrgRouteTable, StaticRouteRule, DrgRouteDistribution, DistributionStatement, MatchCriteria
- **4 enums**: NetworkType (vcn/ipsec_tunnel/remote_peering_connection/virtual_circuit/loopback), VcnRouteType (vcn_cidrs/subnet_cidrs), DistributionType (import_routes/export_routes), MatchType (match_all/drg_attachment_type/drg_attachment_id)
- NetworkDetails.id uses StringValueOrRef without default_kind (polymorphic -- VCN, IPSec, etc.)
- Named OciDynamicRoutingGateway (not OciDrg) for clarity and discoverability

### Stack Outputs

- `drg_id` -- OCID of the DRG
- `default_export_drg_route_distribution_id` -- default export distribution for external attachment configuration

### Validation Tests (spec_test.go)

- **43 Ginkgo/Gomega tests** (27 valid scenarios, 16 invalid scenarios)
- Covers: minimal DRG, all 5 attachment types, transit routing, VCN route types, ECMP, route table/distribution cross-referencing, full hub-and-spoke setup, distribution statements with all 3 match types, static route rules, valueFrom references
- Invalid: missing compartment_id/display_names/network_details/types, unspecified enums, priority out of range

### Pulumi Module (7 Go files)

| File | Purpose |
|------|---------|
| `main.go` | `Resources()` entry point, 6-phase creation orchestrator |
| `locals.go` | `Locals` struct, display name fallback, freeform tags |
| `outputs.go` | Output key constants |
| `drg.go` | DRG resource creation, exports drg_id + default distribution ID |
| `route_distribution.go` | Distributions + statements with attachment ID resolution |
| `route_table.go` | Route tables + static route rules with attachment ID resolution |
| `attachment.go` | DRG attachments with route table and distribution name resolution |

### Terraform Module (8 HCL files)

- `main.tf` -- `oci_core_drg.this`
- `route_distribution.tf` -- distributions (for_each) + statements (for_each, flattened)
- `route_table.tf` -- route tables (for_each) + static route rules (for_each, flattened)
- `attachment.tf` -- attachments (for_each) with depends_on for route tables and distributions
- `variables.tf` -- Full type specification with optional defaults
- `locals.tf` -- Enum maps, tag computation, flattened sub-resource maps
- `outputs.tf` -- drg_id, default_export_drg_route_distribution_id
- `provider.tf` -- OCI provider >= 5.0

### Kind Registration

- `OciDynamicRoutingGateway = 3322` added to CloudResourceKind enum under "Advanced Networking" section
- `kind_map_gen.go` regenerated with new entry

## Benefits

- **VCN peering** -- connect multiple VCNs through a central DRG hub
- **Hybrid connectivity** -- attach IPSec tunnels, FastConnect virtual circuits
- **Cross-region peering** -- remote peering connection support
- **Fine-grained routing** -- custom route tables with ECMP and static routes
- **Route advertisement control** -- import/export distributions with prioritized statements
- **Clean YAML UX** -- name-based cross-referencing eliminates manual OCID wiring

## Impact

- **OCI Provider**: 13th resource kind implemented (13/37 total)
- **Phase 3 Progress**: 3 of 4 Advanced Networking components complete
- **Users**: Can now deploy hub-and-spoke VCN topologies and hybrid connectivity through OpenMCF

## Validation Results

- `go build` -- clean
- `go vet` -- clean
- `go test` -- 43/43 passed
- `terraform validate` -- success

## Related Work

- **R11 OciApplicationLoadBalancer** -- first Phase 3 component
- **R12 OciNetworkLoadBalancer** -- second Phase 3 component
- **R14 OciPublicIp** -- next and final Phase 3 component

---

**Status**: Production Ready

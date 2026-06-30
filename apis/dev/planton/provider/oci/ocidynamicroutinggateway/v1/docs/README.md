# OCI Dynamic Routing Gateway: Design Rationale and Research

## Introduction

The OciDynamicRoutingGateway component manages OCI's virtual router — the central connectivity hub for inter-VCN peering, hybrid cloud (VPN/FastConnect), and cross-region networking. The DRG is one of the most complex networking components in OCI, with a rich sub-resource hierarchy: the DRG itself, attachments, route tables, route distributions, distribution statements, and static route rules. This document explains the design decisions that shaped the bundled component.

## Why Bundle Everything into One Component

The DRG and its sub-resources form a tightly coupled routing domain. The bundling rationale:

1. **Sub-resources are scoped to the DRG.** Route tables, distributions, and attachments all require the DRG OCID. None can exist independently.

2. **Cross-references are internal.** Route tables reference distributions by name. Attachments reference route tables and distributions by name. Static route rules reference attachments by name. Distribution statements can reference specific attachments by name. This web of internal references is natural within a single manifest but would be fragile across separate resources.

3. **Name-based references eliminate OCID juggling.** By keeping everything in one manifest, the Pulumi module resolves `displayName` references to OCIDs during creation. If sub-resources were separate components, users would need `valueFrom` references for every cross-reference — 6+ foreign keys for a typical hub-and-spoke setup.

4. **Deployment ordering is complex.** The creation sequence is: DRG → distributions → route tables → attachments → distribution statements → static route rules. Distributions must exist before route tables (for import references). Route tables must exist before attachments (for route table assignment). Attachments must exist before statements and rules (for next-hop and match references). This ordering is an implementation detail best handled by a single module.

**Trade-off:** The spec is moderately complex (8 nested message types, 3 enums). This is mitigated by the progressive disclosure in documentation — the Quick Start uses a minimal DRG with one VCN attachment and no custom routing.

## Why Sub-Resources Reference by Display Name

OCI sub-resources require OCIDs for cross-referencing (e.g., a DRG attachment's `drgRouteTableId` needs the route table OCID). But OCIDs are generated at creation time — they cannot be known when writing the manifest.

Two approaches were considered:

1. **Two-phase deployment.** Create the DRG and sub-resources first, then update references with the generated OCIDs. This requires multiple apply cycles and is error-prone.

2. **Chosen: Name-based references.** The proto uses `display_name` fields (strings) for intra-component references. The Pulumi module builds lookup maps (`map[string]*Resource`) and resolves names to OCIDs during resource creation. This makes the YAML authoring experience clean:

```yaml
attachments:
  - displayName: "spoke-a"
    drgRouteTableName: "spoke-rt"    # resolved to OCID by the module

routeTables:
  - displayName: "spoke-rt"          # the module creates this first
```

The names must be unique within their scope (all attachment names unique, all route table names unique, all distribution names unique). This is enforced by the map-based lookup — duplicate names would overwrite entries.

## Creation Order and Dependency Chain

The Pulumi module creates resources in a strict order that respects the dependency graph:

```
1. DRG (primary resource)
2. Route Distributions (depend on DRG only)
3. Route Tables (may reference distributions for import)
4. Attachments (may reference route tables and distributions)
5. Distribution Statements (may reference attachments via match criteria)
6. Static Route Rules (reference attachments as next hop)
```

This order solves two key dependency cycles:

- **Distributions before route tables:** A route table's `importDrgRouteDistributionName` needs the distribution OCID. Creating distributions first ensures the OCID is available.
- **Attachments before statements/rules:** Distribution statements can match by `drg_attachment_id`, and static route rules use `nextHopAttachmentName`. Both need the attachment OCID.

The separation of distribution creation (step 2) from statement creation (step 5) is deliberate. Distributions are created empty, then statements are added after attachments exist. This avoids a circular dependency: statements may reference attachments, but attachments may reference distributions.

## Route Distribution Action: Always ACCEPT

The proto does not include an `action` field on distribution statements because OCI only supports one action: `ACCEPT`. The Pulumi module hardcodes `Action: "ACCEPT"` in the statement creation:

```go
_, err := core.NewDrgRouteDistributionStatement(ctx, resourceName, &core.DrgRouteDistributionStatementArgs{
    Action: pulumi.String("ACCEPT"),
    // ...
})
```

Including an `action` field in the proto would add a single-valued enum that provides no user value. If OCI adds additional actions in the future (e.g., REJECT), the proto can be extended with a new field without breaking existing manifests.

## NetworkDetails.id: Untyped StringValueOrRef

Unlike most foreign key fields in Planton, `networkDetails.id` does not have a `default_kind` annotation. This is because the referenced resource type depends on the `networkDetails.type` value:

| type | id references |
|------|---------------|
| `vcn` | VCN OCID (OciVcn) |
| `ipsec_tunnel` | IPSec connection OCID (not an Planton resource) |
| `virtual_circuit` | Virtual circuit OCID (not an Planton resource) |
| `remote_peering_connection` | Remote peering OCID (not an Planton resource) |
| `loopback` | Loopback OCID |

Only VCN attachments can use `valueFrom` to reference an Planton resource. IPSec connections, virtual circuits, and remote peering connections are external resources not managed by Planton. Adding a `default_kind` would be misleading for non-VCN attachment types.

## What's Excluded and Why

### DrgAttachmentManagement

OCI auto-creates attachments for IPSec tunnels, virtual circuits, and remote peering connections. The `oci_core_drg_attachment_management` Terraform resource allows modifying these auto-created attachments (e.g., changing their route table). This is excluded because:

1. Auto-created attachments are owned by their parent resources (IPSec connection, virtual circuit).
2. Managing them requires a separate reconciliation loop that conflicts with the declarative model.
3. The DRG component's attachments list handles explicitly created attachments; auto-created ones are managed by their parent resources.

### DrgAttachmentsList

This is a read-only data source in Terraform. It has no deployment semantics and is excluded from the component.

### Per-Attachment Freeform Tags

Route tables and distributions receive freeform tags. Attachments also receive tags. However, individual route rules and distribution statements do not support tags in the OCI API.

## What's Deferred

- **Defined Tags** — Requires pre-created tag namespace. Freeform tags cover the majority of use cases.
- **Remote Peering Connection Management** — Creating the remote peering connection itself requires cross-region coordination. The DRG component can attach to an existing remote peering connection but does not create one.
- **FastConnect Virtual Circuit Management** — Virtual circuits involve partner provisioning workflows. The DRG component attaches to existing virtual circuits.

## Research Notes

### DRG Limits

| Resource | Limit | Notes |
|----------|-------|-------|
| DRGs per region per tenancy | 1000 (default) | Adjustable via service limit request |
| Attachments per DRG | 300 | Across all network types |
| VCN attachments per DRG | 300 | A VCN can attach to only 1 DRG |
| Route tables per DRG | 200 | Plus the auto-created defaults |
| Route distributions per DRG | 200 | Plus the auto-created default export |
| Static route rules per route table | 200 | |
| Statements per distribution | 200 | |

### Default Resources Created by OCI

When a DRG is created, OCI automatically provisions:

- **Default VCN route table** — routes traffic between VCN attachments
- **Default IPSec/Virtual Circuit route table** — routes traffic from VPN/FastConnect
- **Default export route distribution** — exports all VCN routes to non-VCN attachments

The default export distribution OCID is exported as a stack output because external resources (IPSec connections, virtual circuits created outside Planton) may need to reference it.

### VCN Route Type: VCN CIDRs vs Subnet CIDRs

The `vcnRouteType` field on VCN attachments controls route import granularity:

- **vcn_cidrs** (default): The DRG learns the VCN's CIDR blocks as aggregate routes. Simpler but coarser — all subnet traffic routes to the same attachment.
- **subnet_cidrs**: The DRG learns individual subnet CIDRs. Enables finer-grained routing — traffic to specific subnets can be routed through different paths (e.g., application subnet traffic goes direct, database subnet traffic goes through a firewall).

Subnet CIDRs is recommended for transit routing scenarios where a hub firewall needs to inspect traffic to specific subnets.

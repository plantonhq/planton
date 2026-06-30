---
title: "Dynamic Routing Gateway"
description: "Dynamic Routing Gateway deployment documentation"
icon: "package"
order: 100
componentName: "ocidynamicroutinggateway"
---

# OCI Dynamic Routing Gateway

Deploys an Oracle Cloud Infrastructure Dynamic Routing Gateway (DRG) with VCN attachments, custom route tables, route distributions, and static route rules in a single deployment unit. The DRG is OCI's virtual router for connectivity between VCNs (peering), on-premises networks (Site-to-Site VPN, FastConnect), and cross-region VCNs (remote peering). Sub-resources reference each other by display name, making the YAML experience self-contained.

## What Gets Created

When you deploy an OciDynamicRoutingGateway resource, Planton provisions:

- **Dynamic Routing Gateway** — an `oci_core_drg` resource in the specified compartment. OCI automatically creates default route tables (one per network type) and a default export route distribution. Standard Planton freeform tags are applied.
- **Route Distributions** — one `oci_core_drg_route_distribution` per entry in `routeDistributions`. Controls which routes are advertised to route tables (import) or to attachments (export).
- **Distribution Statements** — one `oci_core_drg_route_distribution_statement` per entry in each distribution's `statements` list. Prioritized rules that match routes by attachment type or specific attachment.
- **Route Tables** — one `oci_core_drg_route_table` per entry in `routeTables`. Controls traffic forwarding between DRG attachments. May import routes from a distribution and contain static route rules.
- **Static Route Rules** — one `oci_core_drg_route_table_route_rule` per entry in each route table's `staticRouteRules` list. Directs traffic for a specific CIDR to a named attachment.
- **Attachments** — one `oci_core_drg_attachment` per entry in `attachments`. Connects a VCN, IPSec tunnel, virtual circuit, remote peering connection, or loopback to the DRG.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the DRG will be created — literal value or reference to an OciCompartment resource
- **VCN OCIDs** for each VCN being attached — literal values or references to OciVcn resources
- **IPSec connection or virtual circuit OCIDs** if attaching on-premises networks (these are created outside this component)

## Quick Start

Create a file `drg.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDynamicRoutingGateway
metadata:
  name: my-drg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciDynamicRoutingGateway.my-drg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  attachments:
    - displayName: "app-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.example"
```

Deploy:

```shell
planton apply -f drg.yaml
```

This creates a DRG with a single VCN attachment. The DRG uses its default route tables and default export distribution. The DRG OCID and default export distribution OCID are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the DRG will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the DRG shown in the OCI Console. |
| `attachments` | `DrgAttachment[]` | — | Network attachments connecting VCNs or other resources to this DRG. See [attachment fields](#attachment-fields). |
| `routeTables` | `DrgRouteTable[]` | — | Custom DRG route tables for controlling traffic routing within the DRG. See [routeTable fields](#routetable-fields). |
| `routeDistributions` | `DrgRouteDistribution[]` | — | Custom route distributions controlling route advertisement. See [routeDistribution fields](#routedistribution-fields). |

### attachment Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `displayName` | `string` | Unique name for this attachment. Used by route rules and distribution statements to reference this attachment. | Minimum 1 character |
| `networkDetails` | `NetworkDetails` | Details of the network being attached. See [networkDetails fields](#networkdetails-fields). | Required |
| `drgRouteTableName` | `string` | Name of a route table defined in `routeTables`. When set, the attachment uses this custom route table instead of the default. | Optional |
| `exportDrgRouteDistributionName` | `string` | Name of a distribution defined in `routeDistributions`. When set, the attachment uses this distribution for exporting routes. | Optional |

### networkDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `type` | `enum` | Type of network resource. Values: `vcn`, `ipsec_tunnel`, `remote_peering_connection`, `virtual_circuit`, `loopback`. | Required (cannot be unspecified) |
| `id` | `StringValueOrRef` | OCID of the network resource (VCN, IPSec connection, virtual circuit, or remote peering connection). | Required |
| `routeTableId` | `string` | OCID of a VCN route table for ingress routing (transit routing). Only applicable for VCN attachments. | Optional |
| `vcnRouteType` | `enum` | Controls whether VCN CIDRs or subnet CIDRs are imported into the DRG route table. Values: `vcn_cidrs`, `subnet_cidrs`. Only applicable for VCN attachments. | Optional |

### routeTable Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `displayName` | `string` | Unique name for this route table. Attachments reference route tables by this name. | Minimum 1 character |
| `importDrgRouteDistributionName` | `string` | Name of a distribution defined in `routeDistributions`. Routes from matching attachments are automatically imported. | Optional |
| `isEcmpEnabled` | `bool` | When `true`, enables Equal-Cost Multi-Path routing across multiple IPSec tunnels or virtual circuits. | Optional |
| `staticRouteRules` | `StaticRouteRule[]` | Static routes for this table. Static routes take precedence over dynamically imported routes. See [staticRouteRule fields](#staticrouterule-fields). | Optional |

### staticRouteRule Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `destination` | `string` | Destination CIDR block (IPv4 or IPv6). Example: `"10.0.0.0/8"`. | Minimum 1 character |
| `nextHopAttachmentName` | `string` | Name of a DRG attachment (defined in `attachments`) that serves as the next hop. | Minimum 1 character |

### routeDistribution Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `displayName` | `string` | Unique name for this distribution. Route tables and attachments reference distributions by this name. | Minimum 1 character |
| `distributionType` | `enum` | Direction of route distribution. Values: `import_routes` (controls import into route tables), `export_routes` (controls export to attachments). | Required (cannot be unspecified) |
| `statements` | `DistributionStatement[]` | Prioritized rules that define which routes are accepted. See [statement fields](#statement-fields). | Optional |

### statement Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `priority` | `int32` | Priority (1-65535). Lower numbers are evaluated first. Must be unique within a distribution. | 1–65535 |
| `matchCriteria` | `MatchCriteria` | Criteria for selecting routes. See [matchCriteria fields](#matchcriteria-fields). | Required |

### matchCriteria Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `matchType` | `enum` | How to match routes. Values: `match_all` (all routes), `drg_attachment_type` (by network type), `drg_attachment_id` (by specific attachment). | Required (cannot be unspecified) |
| `attachmentType` | `string` | Network type to match. Values: `"VCN"`, `"IPSEC_TUNNEL"`, `"VIRTUAL_CIRCUIT"`, `"REMOTE_PEERING_CONNECTION"`. Required when `matchType` is `drg_attachment_type`. | Optional |
| `drgAttachmentName` | `string` | Name of a DRG attachment (defined in `attachments`) to match. Required when `matchType` is `drg_attachment_id`. | Optional |

## Examples

### Simple VCN Peering

Two VCNs attached to a DRG for local peering within the same region. Traffic between VCNs routes through the DRG using the default route tables:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDynamicRoutingGateway
metadata:
  name: peering-drg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciDynamicRoutingGateway.peering-drg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  attachments:
    - displayName: "app-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.appvcn"
    - displayName: "db-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.dbvcn"
```

### Hub-and-Spoke with Custom Route Tables

A hub VCN routing traffic between spoke VCNs through the DRG. Custom route tables and an import distribution control which routes are visible to each spoke:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDynamicRoutingGateway
metadata:
  name: hub-drg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: networking
    pulumi.planton.dev/stack.name: prod.OciDynamicRoutingGateway.hub-drg
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: network-compartment
      fieldPath: status.outputs.compartmentId
  displayName: "Hub DRG"
  routeDistributions:
    - displayName: "import-all-vcns"
      distributionType: import_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: drg_attachment_type
            attachmentType: "VCN"
  routeTables:
    - displayName: "spoke-route-table"
      importDrgRouteDistributionName: "import-all-vcns"
  attachments:
    - displayName: "hub-vcn"
      networkDetails:
        type: vcn
        id:
          valueFrom:
            kind: OciVcn
            name: hub-vcn
            fieldPath: status.outputs.vcnId
    - displayName: "spoke-a"
      networkDetails:
        type: vcn
        id:
          valueFrom:
            kind: OciVcn
            name: spoke-a-vcn
            fieldPath: status.outputs.vcnId
      drgRouteTableName: "spoke-route-table"
    - displayName: "spoke-b"
      networkDetails:
        type: vcn
        id:
          valueFrom:
            kind: OciVcn
            name: spoke-b-vcn
            fieldPath: status.outputs.vcnId
      drgRouteTableName: "spoke-route-table"
```

### Transit Routing with On-Premises VPN

A DRG connecting a VCN to an on-premises network via IPSec VPN. A static route in a custom route table directs on-premises traffic (10.100.0.0/16) to the IPSec tunnel attachment, with ECMP enabled across multiple tunnels:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDynamicRoutingGateway
metadata:
  name: transit-drg
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: networking
    pulumi.planton.dev/stack.name: prod.OciDynamicRoutingGateway.transit-drg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  displayName: "Transit DRG"
  routeDistributions:
    - displayName: "import-vpn-routes"
      distributionType: import_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: drg_attachment_type
            attachmentType: "IPSEC_TUNNEL"
        - priority: 2
          matchCriteria:
            matchType: drg_attachment_type
            attachmentType: "VCN"
  routeTables:
    - displayName: "vcn-to-onprem"
      importDrgRouteDistributionName: "import-vpn-routes"
      isEcmpEnabled: true
      staticRouteRules:
        - destination: "10.100.0.0/16"
          nextHopAttachmentName: "vpn-tunnel"
  attachments:
    - displayName: "prod-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.prodvcn"
        vcnRouteType: subnet_cidrs
      drgRouteTableName: "vcn-to-onprem"
    - displayName: "vpn-tunnel"
      networkDetails:
        type: ipsec_tunnel
        id:
          value: "ocid1.ipsecconnection.oc1.iad.example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `drg_id` | `string` | OCID of the created DRG. |
| `default_export_drg_route_distribution_id` | `string` | OCID of the default export route distribution that OCI automatically creates. Useful for configuring external DRG attachments managed outside this component. |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/vcn) — provides VCN OCIDs for VCN attachments via `valueFrom`
- [OciSubnet](/docs/catalog/oci/subnet) — subnets within attached VCNs route traffic through the DRG for cross-VCN and on-premises connectivity

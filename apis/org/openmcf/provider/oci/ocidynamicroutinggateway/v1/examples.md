# OCI Dynamic Routing Gateway Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Dynamic Routing Gateways using the OpenMCF API. Each example demonstrates different network topologies from simple VCN peering to complex hub-and-spoke architectures with transit routing.

## Table of Contents

- [Example 1: Simple VCN Peering](#example-1-simple-vcn-peering)
- [Example 2: Hub-and-Spoke with Route Control](#example-2-hub-and-spoke-with-route-control)
- [Example 3: Transit Routing with On-Premises VPN](#example-3-transit-routing-with-on-premises-vpn)
- [Example 4: Multi-VCN with Selective Route Distribution](#example-4-multi-vcn-with-selective-route-distribution)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Simple VCN Peering

**Use Case:** Two VCNs need to communicate. The DRG connects both VCNs, and OCI's default route tables handle routing automatically.

**Configuration:**
- **Attachments:** 2 VCN attachments
- **Route Tables:** Default (auto-created by OCI)
- **Distributions:** Default

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicRoutingGateway
metadata:
  name: peering-drg
  org: my-org
  env: dev
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

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f peering-drg.yaml
```

**What happens:**
- A DRG is created in the specified compartment.
- Two VCN attachments are created, connecting both VCNs to the DRG.
- OCI's default VCN route table automatically imports VCN CIDR routes, so both VCNs can reach each other through the DRG.
- VCN route tables in each VCN need a route rule pointing their peer's CIDR to the DRG (configured in OciSubnet).

---

## Example 2: Hub-and-Spoke with Route Control

**Use Case:** Three VCNs in a hub-and-spoke topology. Spoke VCNs communicate through the hub DRG with controlled routing. An import distribution ensures spoke route tables only see VCN routes.

**Configuration:**
- **Attachments:** 3 VCN attachments (1 hub, 2 spokes)
- **Route Tables:** 1 custom table for spokes
- **Distributions:** 1 import distribution accepting all VCN routes

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicRoutingGateway
metadata:
  name: hub-drg
  org: acme
  env: prod
  labels:
    team: networking
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: network-compartment
      fieldPath: status.outputs.compartmentId
  displayName: "Hub DRG"
  routeDistributions:
    - displayName: "import-vcn-routes"
      distributionType: import_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: drg_attachment_type
            attachmentType: "VCN"
  routeTables:
    - displayName: "spoke-rt"
      importDrgRouteDistributionName: "import-vcn-routes"
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
      drgRouteTableName: "spoke-rt"
    - displayName: "spoke-b"
      networkDetails:
        type: vcn
        id:
          valueFrom:
            kind: OciVcn
            name: spoke-b-vcn
            fieldPath: status.outputs.vcnId
      drgRouteTableName: "spoke-rt"
```

**What happens:**
- A custom route table `spoke-rt` imports all VCN routes via the `import-vcn-routes` distribution.
- Both spoke attachments use `spoke-rt`, so they see routes from all three VCNs.
- The hub VCN uses the default route table (which also sees all VCN routes by default).
- Spoke-to-spoke traffic transits through the DRG, not through the hub VCN.

---

## Example 3: Transit Routing with On-Premises VPN

**Use Case:** A VCN connected to an on-premises network via IPSec VPN. Static routes direct on-premises traffic to the VPN tunnel. ECMP is enabled for failover across multiple tunnels.

**Configuration:**
- **Attachments:** 1 VCN + 1 IPSec tunnel
- **Route Tables:** Custom table with static route and ECMP
- **Distributions:** Import distribution for both VPN and VCN routes

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicRoutingGateway
metadata:
  name: transit-drg
  org: acme
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  displayName: "Transit DRG"
  routeDistributions:
    - displayName: "import-all"
      distributionType: import_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: match_all
  routeTables:
    - displayName: "vcn-rt"
      importDrgRouteDistributionName: "import-all"
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
      drgRouteTableName: "vcn-rt"
    - displayName: "vpn-tunnel"
      networkDetails:
        type: ipsec_tunnel
        id:
          value: "ocid1.ipsecconnection.oc1.iad.example"
```

**What happens:**
- The VCN attachment imports individual subnet CIDRs (`vcnRouteType: subnet_cidrs`) for finer-grained on-premises routing.
- The `vcn-rt` route table imports routes from all attachments and adds a static route for the on-premises CIDR (10.100.0.0/16) pointing to the VPN tunnel.
- ECMP is enabled for load-balancing across multiple IPSec tunnels if the VPN connection has more than one tunnel.
- Static routes override dynamic imports for the same CIDR, ensuring on-premises traffic always goes through the VPN.

---

## Example 4: Multi-VCN with Selective Route Distribution

**Use Case:** Isolation between application tiers. The database VCN can reach the app VCN, but the app VCN cannot initiate connections to the database VCN. A custom export distribution on the DB attachment controls which routes are advertised.

**Configuration:**
- **Attachments:** 2 VCN attachments with different distributions
- **Distributions:** Selective export for DB VCN, import for specific attachment

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicRoutingGateway
metadata:
  name: isolated-drg
  org: acme
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  displayName: "Isolated Tier DRG"
  routeDistributions:
    - displayName: "export-app-only"
      distributionType: export_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: drg_attachment_id
            drgAttachmentName: "app-vcn"
    - displayName: "import-from-db"
      distributionType: import_routes
      statements:
        - priority: 1
          matchCriteria:
            matchType: drg_attachment_id
            drgAttachmentName: "db-vcn"
  routeTables:
    - displayName: "app-rt"
      importDrgRouteDistributionName: "import-from-db"
  attachments:
    - displayName: "app-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.appvcn"
      drgRouteTableName: "app-rt"
    - displayName: "db-vcn"
      networkDetails:
        type: vcn
        id:
          value: "ocid1.vcn.oc1.iad.dbvcn"
      exportDrgRouteDistributionName: "export-app-only"
```

**What happens:**
- The `app-rt` route table only imports routes from the `db-vcn` attachment, so the app VCN knows how to reach the database VCN.
- The DB VCN's export distribution only exports routes from the `app-vcn` attachment, so the DB VCN only learns about the app VCN's routes.
- This creates a controlled, bidirectional path between app and DB tiers while preventing other VCNs from being visible.

---

## Common Operations

### Get DRG ID After Deployment

```bash
# Pulumi
pulumi stack output drg_id

# Terraform
terraform output drg_id
```

### Add a VCN Route Rule Pointing to the DRG

For VCNs to use the DRG, their subnets need route rules directing traffic to the DRG. Configure this in the OciSubnet resource:

```yaml
spec:
  routeRules:
    - destinationCidr: "10.1.0.0/16"
      networkEntityId:
        valueFrom:
          kind: OciDynamicRoutingGateway
          name: hub-drg
          fieldPath: status.outputs.drgId
```

---

## Best Practices

### Start with Default Route Tables

OCI creates default route tables per network type (VCN, IPSec, virtual circuit). These defaults automatically import routes from all attachments of the same type. Only create custom route tables when you need:
- Selective route import (not all routes)
- Static route overrides
- ECMP for VPN failover
- Different routing behavior per attachment

### Use Name-Based References

All sub-resources in the manifest reference each other by `displayName`. This keeps the YAML self-contained — you do not need to know OCIDs ahead of time. The Pulumi module resolves names to OCIDs during deployment.

### Prefer Import Distributions Over Static Routes

Static routes are useful for overrides and on-premises CIDRs, but they require manual updates when networks change. Import distributions automatically learn routes as VCNs and attachments are added or modified.

### Plan for Scale

| Network Topology | Recommended DRG Configuration |
|-----------------|------------------------------|
| 2-3 VCN peering | Single DRG, default route tables |
| Hub-and-spoke (5+ VCNs) | Custom route tables for spokes, import distribution for VCN routes |
| Hybrid cloud (VPN/FastConnect) | ECMP-enabled route table, static routes for on-premises CIDRs |
| Multi-tier isolation | Per-tier route tables with selective import distributions |

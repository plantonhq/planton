# Overview

The **OCI Dynamic Routing Gateway API Resource** provides a consistent interface for deploying and managing DRGs on Oracle Cloud Infrastructure. A DRG is OCI's virtual router enabling connectivity between VCNs (local peering), on-premises networks (Site-to-Site VPN via IPSec, FastConnect via virtual circuits), cross-region VCNs (remote peering), and loopback interfaces. In hub-and-spoke network topologies, the DRG acts as the central hub.

## Purpose

This API resource bundles the DRG with its attachments, route tables, route distributions, distribution statements, and static route rules into a single deployment unit. Sub-resources reference each other by display name rather than OCID, making the YAML authoring experience clean and self-contained:

- **VCN Peering**: Attach multiple VCNs to a single DRG for cross-VCN communication without requiring separate Local Peering Gateways for each VCN pair.
- **Hybrid Cloud Connectivity**: Connect VCNs to on-premises networks via IPSec VPN tunnels or FastConnect virtual circuits.
- **Hub-and-Spoke Topologies**: Configure custom route tables and distributions to control exactly which routes are visible to each spoke VCN.
- **Transit Routing**: Route traffic between VCNs and on-premises through a hub VCN with firewall appliances for inspection.
- **Infra-Chart Composability**: The DRG OCID and default export distribution OCID are exported for use by external network resources.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Self-Referencing Sub-Resources**: Attachments, route tables, distributions, and route rules reference each other by `displayName` within the same manifest. No need to manage OCIDs across separate resource definitions.
- **Five Network Types**: Supports VCN, IPSec tunnel, virtual circuit, remote peering connection, and loopback attachment types.
- **Custom Route Tables**: Define route tables with static routes and dynamic import from distributions. Attachments can use custom or default route tables.
- **Route Distributions with Priority Statements**: Fine-grained control over which routes are imported into route tables or exported to attachments, with prioritized match criteria.
- **ECMP Support**: Equal-Cost Multi-Path routing across multiple IPSec tunnels or virtual circuits for bandwidth aggregation and high availability.
- **VCN Route Type Control**: Choose between importing aggregate VCN CIDRs or individual subnet CIDRs for finer-grained routing decisions.
- **Automatic Tagging**: Standard OpenMCF freeform tags applied to the DRG and all attachments for resource tracking.

## How OCI DRG Differs from AWS Transit Gateway

| Aspect | OCI DRG | AWS Transit Gateway |
|--------|---------|-------------------|
| **Peering model** | Single DRG per region with multiple attachments | Single TGW per region with multiple attachments |
| **Attachment types** | VCN, IPSec, FastConnect, Remote Peering, Loopback | VPC, VPN, Direct Connect, Peering, Connect |
| **Route tables** | Per-attachment (via DRG route tables) | Per-attachment (via TGW route tables) |
| **Route propagation** | Via route distributions with priority statements | Via association and propagation tables |
| **ECMP** | Per-route-table toggle | Automatic across VPN attachments |
| **Default behavior** | Default route tables per network type, default export distribution | Default association and propagation tables |

The key difference is OCI's route distribution model. AWS TGW uses a simpler association/propagation model where attachments are associated with one route table and propagate to others. OCI's distribution statements offer more granular control — you can match by attachment type, specific attachment, or all attachments, with explicit priority ordering.

## Critical Constraints

- **One DRG per VCN**: A VCN can be attached to only one DRG at a time. However, a single DRG can have multiple VCN attachments.
- **Name-Based References**: Sub-resources within the manifest reference each other by `displayName`. These names must be unique within their scope (attachments, route tables, distributions).
- **Action Is Always ACCEPT**: Distribution statements in OCI only support the ACCEPT action. The match criteria determine which routes are accepted; there is no explicit reject.
- **Static Route Precedence**: Static routes in DRG route tables take precedence over dynamically imported routes. This allows override of specific CIDRs while still importing the bulk of routes dynamically.
- **IPSec and Virtual Circuit Attachments**: The IPSec connection or virtual circuit must exist before creating the DRG attachment. These resources are managed outside this component.

## Use Cases

- **Multi-VCN Peering**: Replace point-to-point Local Peering Gateways with a central DRG. With N VCNs, a single DRG replaces N*(N-1)/2 LPGs.
- **Site-to-Site VPN**: Connect OCI VCNs to an on-premises data center via IPSec tunnels. ECMP across multiple tunnels provides bandwidth aggregation.
- **FastConnect**: High-bandwidth, low-latency private connectivity between OCI and on-premises via Oracle's partner network or co-location.
- **Cross-Region Peering**: Connect VCNs in different OCI regions via remote peering connections attached to regional DRGs.
- **Transit Routing with Firewall Inspection**: Route all inter-VCN and VPN traffic through a hub VCN containing firewall appliances. Custom DRG route tables and VCN ingress routing steer traffic through the firewall before reaching the destination.

## Production Features

- **Custom Route Tables**: Fine-grained control over traffic forwarding between attachments, with both static routes and dynamic imports.
- **Route Distributions**: Prioritized statements controlling which routes are visible to which route tables and attachments.
- **ECMP**: Load-balance traffic across multiple IPSec tunnels or virtual circuits for higher aggregate bandwidth and failover.
- **Subnet-Level Route Import**: Import individual subnet CIDRs instead of aggregate VCN CIDRs for finer routing granularity.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags on the DRG and all attachments.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical resource topology and outputs.

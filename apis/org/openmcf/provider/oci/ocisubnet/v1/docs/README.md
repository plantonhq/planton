# OCI Subnet: Design Rationale and Research

## Introduction

The OCI Subnet component is the Layer 1 building block in the OCI networking stack. While the VCN (Layer 0) establishes the network boundary, gateways, and address space, it is the subnet that determines where workloads land, how they are routed, and whether they can communicate with the internet. Every compute instance, OKE node pool, load balancer, and database in OCI attaches to a subnet.

This document explains the design decisions behind the OciSubnet component, the route table ownership model, differences from AWS subnets, and the rationale for public/private controls.

## Why Subnets Are Separate from the VCN

The OciVcn component bundles gateways (Internet, NAT, Service) because they are tightly coupled to the VCN lifecycle — a gateway cannot exist without a VCN, and gateways are not independently configured or delegated.

Subnets are deliberately separate for three reasons:

**1. Subnets are the delegation boundary.** A platform team creates the VCN and gateways, then application teams create their own subnets with team-specific CIDRs, routing, and access controls. Bundling subnets into the VCN would force all subnet configuration into a single resource manifest managed by a single team.

**2. Subnets have meaningful independent configuration.** Each subnet has its own CIDR, route table, security lists, DNS label, and public/private designation. Two subnets in the same VCN can have completely different routing and access characteristics. This level of per-subnet configuration makes the subnet a natural unit of individual resource management.

**3. Subnet count varies.** A VCN might have 3 subnets or 30. Bundling a variable-length list of complex nested objects into the VCN spec would make the VCN manifest unwieldy and would complicate incremental changes (adding a subnet shouldn't require re-applying the entire VCN).

This separation mirrors how OCI itself models the relationship — subnets are child resources of a VCN, referenced by `vcn_id`, not embedded within it.

## The Route Table Ownership Model

The most distinctive design feature of OciSubnet is its three-mode route table association. This was the primary design challenge and the area where we made the most deliberate trade-offs.

### The Three Modes

1. **Inline route rules** (`routeRules`): The subnet spec includes route rules, and the module creates a dedicated `oci_core_route_table` named `{displayName}-rt` in the same compartment and VCN. The route table's lifecycle is tied to the subnet — created with it, updated with it, deleted with it.

2. **External reference** (`routeTableId`): The subnet references an existing route table by OCID. The route table is managed externally (by another team, another tool, or the OCI Console). The subnet does not create or modify it.

3. **VCN default** (neither provided): The subnet uses the VCN's default route table, which OCI creates automatically with every VCN. This is the simplest option and suitable for development or flat networks where all subnets share the same routing.

### Why Mutual Exclusivity

The `routeTableId` and `routeRules` fields are mutually exclusive, enforced by a CEL validation rule:

```
!has(this.route_table_id) || this.route_rules.size() == 0
```

This prevents ambiguity. If both were allowed, the module would need to decide which one wins — a silent precedence rule that would confuse users. By making them mutually exclusive, the manifest is always explicit about where routing comes from.

### Why Inline Route Rules Exist

In most production OCI environments, each subnet needs its own route table:
- Public subnets route to the Internet Gateway
- Private subnets route to the NAT Gateway
- Service subnets route to the Service Gateway
- Database subnets may have no internet route at all

Without inline route rules, creating a private subnet with proper routing requires three separate resource definitions: the subnet, a route table, and a route table attachment. The inline model collapses this into a single manifest where the route rules are declared directly on the subnet.

The trade-off: inline route tables cannot be shared across subnets. If two subnets need identical routing, you either duplicate the rules (inline) or create a shared route table externally and reference it via `routeTableId`. Both patterns are supported.

### Route Table Naming

Inline route tables are named `{displayName}-rt`. Since `displayName` falls back to `metadata.name`, a subnet named `app-tier` produces a route table named `app-tier-rt`. This convention ensures the route table is identifiable in the OCI Console and linked to its owning subnet.

## OCI Subnet vs AWS Subnet: A Detailed Comparison

### Scope and Placement

| Aspect | OCI Subnet | AWS Subnet |
|--------|------------|------------|
| **Default scope** | Regional (spans all ADs) | Single AZ (always) |
| **AD/AZ-specific** | Optional — set `availabilityDomain` | Always — `availability_zone` is required |
| **HA implication** | One regional subnet serves all ADs | One subnet per AZ needed for cross-AZ HA |
| **Typical count** | 3-6 per VCN (public, private, service tiers) | 6-12 per VPC (1 per tier per AZ, for 2-3 AZs) |

The practical consequence: OCI environments need fewer subnets. A typical OKE deployment might have 3 subnets (public LB, private workers, private API endpoint), while an equivalent EKS deployment needs 6 (2 AZs x 3 tiers) or 9 (3 AZs x 3 tiers).

### Route Table Association

| Aspect | OCI | AWS |
|--------|-----|-----|
| **Default** | VCN's default route table | VPC's main route table |
| **Explicit** | Set at creation via `route_table_id` | Associated post-creation via `aws_route_table_association` |
| **Inline creation** | Supported by OpenMCF (routeRules) | Not available in native Terraform |
| **Shared route tables** | Supported via `routeTableId` reference | Supported via explicit association |

AWS requires a separate `aws_route_table_association` resource to link a route table to a subnet. OCI sets the association at creation time. OpenMCF goes further by supporting inline route rule declarations that create and associate a route table in a single operation.

### Public/Private Model

| Aspect | OCI | AWS |
|--------|-----|-----|
| **How determined** | Explicit: `prohibitPublicIpOnVnic` + `prohibitInternetIngress` | Implicit: presence of IGW route in route table |
| **Public IP control** | `prohibitPublicIpOnVnic: true` blocks public IPs | Subnet `map_public_ip_on_launch` + instance-level setting |
| **Ingress control** | `prohibitInternetIngress: true` blocks inbound internet at subnet level | NACLs and Security Groups |

OCI's model is more explicit. In AWS, a "private subnet" is simply a subnet whose route table has no route to an Internet Gateway — there's no flag that says "this is private." In OCI, `prohibitPublicIpOnVnic` and `prohibitInternetIngress` are direct controls that enforce privacy regardless of the route table configuration.

### Security Model

| Layer | OCI | AWS |
|-------|-----|-----|
| **Subnet-level** | Security Lists (stateful, per-subnet) | NACLs (stateless, per-subnet) |
| **Instance-level** | Network Security Groups (stateful, per-VNIC) | Security Groups (stateful, per-ENI) |
| **Recommendation** | NSGs for new deployments | Security Groups for most use cases |

OCI security lists are stateful (unlike AWS NACLs, which are stateless). This means return traffic is automatically allowed in OCI security lists, similar to AWS Security Groups. However, OCI still recommends NSGs over security lists for new deployments because NSGs are per-VNIC and more granular.

## Regional vs AD-Specific Subnets

### Why Regional Is the Default

OCI subnets are regional by default — they span all availability domains in a region. This was an intentional OCI platform decision, and OpenMCF follows it by making `availabilityDomain` optional.

Regional subnets are preferred because:
- **Simplified HA**: A single subnet can host workloads across all ADs. No need to create per-AD subnets and manage multiple subnet references.
- **Fewer resources**: A 3-tier architecture needs 3 subnets (public, private app, private DB) instead of 9 (3 tiers x 3 ADs).
- **OKE compatibility**: OKE node pools work with regional subnets and distribute nodes across ADs automatically.
- **Load balancer flexibility**: OCI load balancers in regional subnets can span all ADs.

### When to Use AD-Specific Subnets

AD-specific subnets are appropriate in narrow scenarios:
- **Bare metal instances**: Some bare metal shapes are available only in specific ADs.
- **Legacy architectures**: Migrating from environments that assumed AD-scoped networking.
- **Compliance**: Regulations requiring data residency within a specific physical location (AD).

When using AD-specific subnets, you need one subnet per AD per tier, which increases the total subnet count significantly.

## Public vs Private: How the Two Booleans Interact

OCI provides two independent controls for subnet access:

| `prohibitPublicIpOnVnic` | `prohibitInternetIngress` | Effect |
|--------------------------|---------------------------|--------|
| `false` | `false` | **Public**: VNICs can have public IPs, inbound internet traffic allowed (subject to security rules) |
| `true` | `true` | **Private**: No public IPs, no inbound internet traffic — the strongest isolation |
| `false` | `true` | **Hybrid**: VNICs can have public IPs, but inbound internet traffic is blocked at the subnet level |
| `true` | `false` | **Semi-private**: No public IPs on VNICs, but the subnet does not explicitly block inbound internet traffic (traffic would still fail without a public IP to route to) |

The recommended production pattern is `prohibitPublicIpOnVnic: true` + `prohibitInternetIngress: true` for all subnets that don't need direct internet access. This provides defense in depth — even if a security rule or NSG accidentally allows inbound traffic, the subnet-level control blocks it.

For load balancer subnets, use the default (`false`/`false`) to allow public IP assignment and inbound traffic.

## Design Decisions

### DisplayName Falls Back to metadata.name

The `displayName` field is optional. When omitted, both the Pulumi and Terraform modules fall back to `metadata.name`. Minimal manifests that only set `metadata.name` still produce human-readable resources in the OCI Console. The inline route table name also derives from the display name: `{displayName}-rt`.

### Freeform Tags Over Defined Tags

Consistent with the OciVcn component, OpenMCF uses freeform tags for both the subnet and any created route table:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciSubnet"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

Freeform tags require no pre-configuration and work across all compartments without additional IAM policies.

### Why Security Lists Are Referenced, Not Bundled

Security lists are referenced by OCID (`securityListIds`) rather than defined inline for two reasons:

1. **OCI recommends NSGs**: For new deployments, OCI recommends Network Security Groups over security lists. Bundling security list creation into the subnet would invest design effort in a declining pattern.
2. **Security lists are shared**: A single security list can be associated with multiple subnets. Embedding security list definitions inside the subnet spec would prevent sharing and force duplication.

The `securityListIds` field accepts up to 5 `StringValueOrRef` entries, matching OCI's per-subnet limit.

### Why DHCP Options Are Referenced, Not Bundled

Custom DHCP options are uncommon. The VCN's default DHCP options (OCI internal DNS resolver) are sufficient for the vast majority of deployments. The `dhcpOptionsId` field is available for the rare case where custom DNS resolvers or search domains are needed, but there is no justification for bundling DHCP option creation into the subnet component.

## Downstream Dependencies

The OciSubnet component is consumed by virtually every OCI compute and service component:

```
OciSubnet
├── OciComputeInstance (VNIC attached to this subnet)
├── OciContainerEngineCluster (API endpoint subnet)
├── OciContainerEngineNodePool (worker node subnet)
├── OciApplicationLoadBalancer (placed in this subnet)
├── OciNetworkLoadBalancer (placed in this subnet)
├── OciDbSystem (database system in private subnet)
├── OciAutonomousDatabase (private endpoint in this subnet)
├── OciMysqlDbSystem (placed in private subnet)
├── OciPostgresqlDbSystem (placed in private subnet)
├── OciContainerInstance (placed in this subnet)
└── OciFunctionsApplication (functions in this subnet for VCN access)
```

This makes the subnet the most heavily referenced Layer 1 component. Its `subnet_id` output is the primary input for every workload component in the OCI provider.

## What OpenMCF Supports

### Current Implementation

The OciSubnet component covers the complete subnet use case:

- **Subnet creation** with configurable CIDR, DNS, IPv6, public/private controls
- **Route table creation** via inline route rules with NAT, Internet, Service, and DRG gateway support
- **External route table reference** for shared or externally managed route tables
- **Security list binding** with up to 5 security lists per subnet
- **Custom DHCP options** via external reference
- **Both Pulumi and Terraform** implementations producing identical resource topology and outputs
- **5 stack outputs** for downstream composability (subnet ID, domain name, virtual router IP/MAC, route table ID)

### What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Inline security list creation** — OCI recommends NSGs; security lists are referenced by OCID
- **Inline DHCP options creation** — uncommon; the VCN default covers most use cases
- **IPv6 route rules** — IPv6 routing is supported by OCI but deferred until IPv6 adoption in OCI matures
- **Cross-AD subnet groups** — creating matched sets of AD-specific subnets is an infra-chart concern, not a single-resource concern

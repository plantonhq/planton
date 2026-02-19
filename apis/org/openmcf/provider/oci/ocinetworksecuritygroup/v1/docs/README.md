# OCI Network Security Group: Design Rationale and Research

## Introduction

The OCI Network Security Group component is the per-VNIC security boundary in the OCI networking stack. While the VCN (Layer 0) establishes the network perimeter and the subnet (Layer 1) determines IP addressing and routing, the NSG controls which traffic can reach individual compute instances, databases, load balancers, and other VNIC-attached resources. NSGs are OCI's recommended firewall mechanism for new deployments, superseding the older security list model.

This document explains the design decisions behind the OciNetworkSecurityGroup component, compares OCI's security model with AWS and Azure, and documents the rationale for the direction-split rule design, inline rule management, and protocol abstraction.

## Why NSGs Are Separate from VCNs and Subnets

The OciVcn component bundles gateways because they are tightly coupled to the VCN lifecycle. The OciSubnet component bundles route tables (optionally) because a subnet's routing is intrinsic to its identity. NSGs are deliberately separate from both for four reasons:

**1. NSGs operate at a different granularity.** VCNs and subnets are network topology constructs — they define address space and routing. NSGs are security policy constructs — they define who can talk to whom. A single subnet can have resources attached to many different NSGs with different security postures. Bundling NSGs into the subnet would conflate topology with policy.

**2. NSGs change more frequently than topology.** Network topology (VCNs, subnets, route tables) is typically established once and updated rarely. Security rules change as applications evolve — new ports, new source services, new compliance requirements. Separating NSGs from topology ensures that security policy updates don't require re-applying network infrastructure.

**3. NSGs are the delegation boundary for security.** A platform team creates the VCN and subnets. Application teams define their own NSGs with team-specific security rules. Bundling NSGs into the subnet would prevent this delegation model.

**4. Multiple NSGs per VNIC.** OCI allows attaching up to 5 NSGs to a single VNIC. This composability model — where different security concerns are expressed as different NSGs — requires NSGs to be independent resources. A VNIC might have a "base connectivity" NSG managed by the platform team and an "application-specific" NSG managed by the app team.

## The Direction-Split Design Decision

The most significant design choice in OciNetworkSecurityGroupSpec is splitting rules into `ingress_rules` and `egress_rules` rather than having a single `rules` list with a `direction` field.

### The Problem with the Raw API

In the OCI provider API (both Terraform and Pulumi), a security rule has a `direction` field (`INGRESS` or `EGRESS`) alongside `source`, `source_type`, `destination`, and `destination_type` fields. The semantics are:

- When `direction` is `INGRESS`: `source` and `source_type` are set; `destination` and `destination_type` are null.
- When `direction` is `EGRESS`: `destination` and `destination_type` are set; `source` and `source_type` are null.

This creates a conditional schema where 2 of 4 fields are always null depending on the direction. Users must remember which fields to set for which direction, and a mistake (setting `destination` on an ingress rule) silently produces a broken rule.

### The OpenMCF Solution

By splitting into `ingress_rules` and `egress_rules`, the direction is implicit from the field name:

- `ingress_rules[].source` — always valid, always set
- `egress_rules[].destination` — always valid, always set

There is no `direction` field, no conditional null fields, and no way to accidentally set the wrong fields. The proto schema enforces correctness at the type level rather than relying on runtime validation.

### The Trade-Off

The trade-off is that common fields (protocol, description, stateless, tcp_options, udp_options, icmp_options) are duplicated between IngressRule and EgressRule messages. This was an intentional choice — type safety and user clarity outweigh the minor proto duplication. The IaC modules handle the mapping back to OCI's unified rule model internally.

## OCI NSG vs AWS Security Groups: A Detailed Comparison

### Scope and Attachment

| Aspect | OCI NSG | AWS Security Group |
|--------|---------|-------------------|
| **Attachment target** | VNIC | ENI (Elastic Network Interface) |
| **Attachment limit** | 5 NSGs per VNIC | 5 SGs per ENI (default, configurable to 16) |
| **VCN/VPC scope** | Belongs to one VCN | Belongs to one VPC |
| **Cross-VCN/VPC** | Not shareable | Not shareable |
| **Default rules** | None (empty = deny all) | Allow all outbound, deny all inbound |

The most operationally significant difference is the default behavior. AWS Security Groups allow all outbound traffic by default, so attaching a new SG to an instance does not break outbound connectivity. OCI NSGs have no defaults — an empty NSG blocks everything. This is why OpenMCF presets always include at least one egress rule.

### Rule Model

| Aspect | OCI NSG | AWS Security Group |
|--------|---------|-------------------|
| **Rule limit** | 120 total (ingress + egress) | 60 inbound + 60 outbound (120 total) |
| **Stateful** | Yes (default) | Yes (always) |
| **Stateless option** | Yes (`stateless: true`) | No — SGs are always stateful |
| **Protocol specification** | Numeric string ("6", "17") or "all" | Protocol name or number |
| **Source/destination types** | CIDR, Service CIDR, NSG OCID | CIDR, Prefix List, SG ID |
| **ICMP filtering** | Type + optional code | Type + optional code |
| **Port ranges** | min/max (source + destination) | From port / to port |

OCI's stateless option is unique among major cloud providers. It allows per-rule opt-out of connection tracking, which can improve throughput for high-volume workloads (e.g., NFS, streaming data) at the cost of requiring explicit return-traffic rules.

### Service CIDR Labels vs Prefix Lists

OCI uses **service CIDR labels** (e.g., `all-iad-services-in-oracle-services-network`) to represent the IP ranges of OCI services. These labels are region-specific and automatically updated as Oracle adds new service endpoints.

AWS uses **prefix lists** (e.g., `pl-63a5400a` for S3 in us-east-1) for a similar purpose. AWS prefix lists can be AWS-managed (for services) or customer-managed.

The key difference: OCI's service CIDR labels are human-readable and self-documenting in the manifest. AWS prefix list IDs are opaque and require a lookup to understand what they represent.

### NSG-to-NSG vs SG-to-SG References

Both OCI and AWS support referencing another security group/NSG as a traffic source or destination:

```yaml
# OCI NSG: sourceType = network_security_group
source: "ocid1.networksecuritygroup.oc1.iad.example"
sourceType: network_security_group

# AWS SG equivalent concept:
# source_security_group_id = "sg-0123456789abcdef0"
```

The models are functionally equivalent. Both enable micro-segmentation where rules reference group membership instead of IP addresses.

## OCI NSG vs Azure NSG

Azure Network Security Groups have a different model worth noting:

| Aspect | OCI NSG | Azure NSG |
|--------|---------|-----------|
| **Attachment** | Per-VNIC | Per-subnet or per-NIC |
| **Priority** | No priority — all rules evaluated | Priority-based (100–4096) |
| **Default rules** | None | 3 default inbound + 3 default outbound rules |
| **Rule limit** | 120 total | 1000 per NSG |
| **Deny rules** | Not supported (implicit deny only) | Explicit deny rules supported |

The most significant difference is that OCI NSGs do not support explicit deny rules. Traffic is either allowed by a matching rule or implicitly denied. Azure NSGs support both allow and deny rules with priority ordering, enabling more complex rule evaluation logic at the cost of increased complexity.

Azure's dual-attachment model (per-subnet or per-NIC) is also notable. OCI separates these concerns: security lists handle subnet-level filtering, and NSGs handle VNIC-level filtering.

## Why Security Rules Are Inline (Not Separate Resources)

Security rules are defined as repeated fields within the NSG spec rather than as separate OpenMCF resources. This matches the lifecycle reality:

**Rules have no independent identity.** A security rule does not have a stable OCID in the OCI data model — it exists only as a sub-resource of an NSG. While Terraform's `oci_core_network_security_group_security_rule` resource provides per-rule management, this is a provider convenience, not a reflection of OCI's object model.

**Rules change together.** When updating an NSG's security posture, you typically modify multiple rules simultaneously — adding a new port, removing an old source, updating a description. Managing each rule as a separate resource would require N separate manifests and N separate apply operations for what is conceptually a single security policy change.

**The proto model is clearer.** A single OciNetworkSecurityGroup manifest with inline rules provides a complete, readable picture of the security policy. A directory of individual rule manifests would scatter the security posture across many files, making review and auditing harder.

**Reconciliation is simpler.** The IaC modules manage all rules within a single Pulumi/Terraform operation. If the desired state has 10 rules and the current state has 8, the module adds 2. This would be significantly more complex if rules were independent resources with their own drift detection.

## Protocol Enum Mapping

The spec uses human-readable protocol names (`tcp`, `udp`, `icmp`, `icmpv6`, `all`) instead of OCI's numeric protocol strings (`"6"`, `"17"`, `"1"`, `"58"`, `"all"`).

### Why Not Use OCI's Native Strings

OCI's API represents protocols as string-typed numbers: `"6"` for TCP, `"17"` for UDP, `"1"` for ICMP. These are IANA protocol numbers encoded as strings. Using them directly in the spec would force users to memorize IANA protocol numbers — a usability failure for a developer-facing API.

### The Mapping

| Spec Value | OCI Protocol String | IANA Number |
|------------|--------------------:|-------------|
| `all` | `"all"` | — |
| `tcp` | `"6"` | 6 |
| `udp` | `"17"` | 17 |
| `icmp` | `"1"` | 1 |
| `icmpv6` | `"58"` | 58 |

The Pulumi module's `protocolString()` function and the Terraform module's `protocol_map` local perform this mapping. Both are defined in the IaC code and tested against OCI's API expectations.

### Why Only 5 Protocols

OCI supports specifying any IANA protocol number as a string. However, the vast majority of real-world NSG rules use only TCP, UDP, ICMP, or "all." Supporting arbitrary protocol numbers would complicate the enum without serving a meaningful use case. If a future need arises for protocols like GRE (47) or ESP (50), the enum can be extended.

## The 120-Rule Limit

OCI enforces a hard limit of 120 security rules per NSG (ingress + egress combined). This is an API-level constraint — the OCI API returns an error if you attempt to create a 121st rule.

### Proto-Level Enforcement

The spec includes a CEL validation rule that enforces this constraint at submission time:

```
this.ingress_rules.size() + this.egress_rules.size() <= 120
```

This prevents deployment failures by catching the violation before the manifest reaches the IaC engine. The error message is explicit: "an NSG supports at most 120 rules (ingress + egress combined)."

### Implications for Complex Environments

120 rules is generous for most use cases:

- A typical web tier: 3–5 rules
- A typical backend: 2–4 rules
- A complex multi-service NSG: 20–40 rules

For environments that need more than 120 rules, the workaround is multiple NSGs. OCI allows up to 5 NSGs per VNIC, providing a theoretical maximum of 600 rules per VNIC. The rules from all attached NSGs are combined (union) — a packet is allowed if any attached NSG has a rule that permits it.

## Freeform Tags

Consistent with all OCI components in OpenMCF, the NSG receives freeform tags derived from metadata:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciNetworkSecurityGroup"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

Tags are applied only to the NSG resource itself, not to individual security rules (rules do not support tags in the OCI API).

## Downstream Dependencies

The OciNetworkSecurityGroup component is consumed by every OCI resource that attaches to a VNIC:

```
OciNetworkSecurityGroup
├── OciComputeInstance (VNIC NSG association)
├── OciContainerEngineCluster (API endpoint NSG)
├── OciContainerEngineNodePool (worker node NSG)
├── OciLoadBalancer (LB NSG association)
├── OciNetworkLoadBalancer (NLB NSG association)
├── OciDbSystem (database NSG)
├── OciMysqlDbSystem (MySQL NSG)
├── OciPostgresqlDbSystem (PostgreSQL NSG)
├── OciContainerInstance (container NSG)
└── OciFunctionsApplication (function NSG for VCN access)
```

The `network_security_group_id` output is the single cross-resource reference. Downstream components reference it via `StringValueOrRef`.

## What OpenMCF Supports

### Current Implementation

The OciNetworkSecurityGroup component covers the complete NSG use case:

- **NSG creation** with display name and freeform tags
- **Inline security rules** with full ingress/egress support
- **Five protocol types** with TCP/UDP port ranges and ICMP type/code filtering
- **Three target types** for CIDR, service CIDR, and NSG-to-NSG rules
- **Stateful and stateless** rule support
- **120-rule validation** at the proto level
- **Both Pulumi and Terraform** implementations producing identical resource topology and outputs
- **Proto validation** for required fields, port ranges, and rule limits

### What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Defined tags** — OCI's schema-enforced tags require pre-configuration of tag namespaces and are uncommon in initial deployments. Freeform tags cover the primary use cases.
- **Rule ordering** — OCI evaluates all NSG rules in parallel (no priority ordering like Azure). Rule order in the manifest affects only the Terraform resource naming, not the security behavior.
- **Arbitrary protocol numbers** — Only the 5 most common protocols are supported via enum. Supporting arbitrary IANA protocol numbers can be added if a concrete use case emerges.
- **IPv6 CIDR rules** — IPv6 source/destination CIDRs work with the existing string fields, but dedicated IPv6 examples and validation are deferred until IPv6 adoption in OCI matures.

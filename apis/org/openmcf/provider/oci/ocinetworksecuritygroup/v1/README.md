# Overview

The **OCI Network Security Group API Resource** provides a consistent and standardized interface for deploying and managing Network Security Groups on Oracle Cloud Infrastructure. An NSG acts as a virtual firewall for compute instances and other VNIC-attached resources, providing per-VNIC traffic control through ingress and egress security rules that support TCP, UDP, ICMP, ICMPv6, and protocol-agnostic filtering.

## Purpose

This API resource streamlines the deployment and management of OCI NSGs with inline security rules. By offering a unified interface, it reduces the complexity of network security configuration, enabling users to:

- **Define Per-VNIC Firewall Rules**: Attach NSGs to individual VNICs for fine-grained traffic control, independent of subnet-level security lists.
- **Split Rules by Direction**: Declare `ingressRules` and `egressRules` separately, eliminating the error-prone pattern of specifying direction as a field on each rule alongside conditional source/destination.
- **Target Multiple Source Types**: Write rules that reference CIDR blocks, OCI service CIDR labels, or other NSGs — enabling CIDR-based, service-aware, and NSG-to-NSG micro-segmentation in a single resource.
- **Enable Infra-Chart Composability**: Export the NSG OCID as a `StringValueOrRef` target for downstream resources that need network security group associations.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Direction-Split Rules**: Ingress and egress rules are separate repeated fields. Source is always on ingress, destination is always on egress — no conditional interpretation needed.
- **Five Protocol Types**: `all`, `tcp`, `udp`, `icmp`, and `icmpv6`. The IaC modules map these human-readable names to OCI's numeric protocol strings internally.
- **Three Target Types**: CIDR blocks for traditional IP-based rules, service CIDR labels for OCI-service-aware rules, and NSG OCIDs for micro-segmented zero-trust architectures.
- **Stateful and Stateless**: Rules are stateful by default (return traffic is automatically allowed). Set `stateless: true` for high-throughput scenarios where explicit bidirectional rules are preferred.
- **120-Rule Limit Validation**: OCI enforces a maximum of 120 rules per NSG. The proto schema validates this constraint at submission time with a CEL expression, preventing deployment failures.
- **Port Range and ICMP Options**: TCP and UDP rules support destination and source port ranges. ICMP rules support type and optional code filtering.
- **Automatic Tagging**: Standard OpenMCF freeform tags are applied to the NSG (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Ready**: Exports the NSG OCID as a stack output for downstream `StringValueOrRef` references from compute instances, OKE clusters, load balancers, and databases.

## How OCI NSGs Differ from Other Providers

Understanding these differences is essential when coming from AWS or Azure:

- **Per-VNIC vs Per-ENI**: OCI NSGs attach to individual VNICs, similar to AWS Security Groups which attach to ENIs. The operational model is nearly identical — both are stateful, instance-level firewalls. The key difference is that OCI also has security lists (subnet-level, similar to AWS NACLs but stateful), and recommends NSGs over security lists for new deployments.
- **Rule Limit**: OCI allows 120 rules per NSG (ingress + egress combined). AWS Security Groups allow 60 inbound + 60 outbound rules (120 total, matching OCI). The limit applies per NSG/SG, and multiple NSGs/SGs can be attached to a single VNIC/ENI.
- **Source Types**: OCI NSG rules support three source types: CIDR block, service CIDR label, and another NSG OCID. AWS Security Groups support CIDR blocks, prefix lists, and other Security Group IDs. The models are functionally equivalent.
- **Service CIDR Labels**: Unique to OCI. NSG rules can reference OCI service CIDRs (e.g., `all-iad-services-in-oracle-services-network`) to allow traffic to/from OCI services. AWS achieves similar functionality with prefix lists for AWS services.
- **Default Behavior**: OCI NSGs have no default rules — an NSG with zero rules blocks all traffic. AWS Security Groups have a default "allow all outbound" rule. This means OCI NSGs require explicit egress rules for any outbound communication.
- **Stateless Option**: OCI NSG rules support a per-rule `stateless` flag that disables connection tracking for that rule. AWS Security Groups are always stateful with no opt-out.
- **VCN Scope**: An OCI NSG belongs to a single VCN and cannot be shared across VCNs. AWS Security Groups belong to a single VPC with the same constraint.

## Critical Constraints

- **120-Rule Maximum**: OCI enforces a hard limit of 120 security rules per NSG (ingress + egress combined). For complex environments, use multiple NSGs attached to the same VNIC.
- **VCN Binding**: An NSG belongs to a single VCN. Changing `vcnId` forces recreation of the NSG and all its rules.
- **No Default Rules**: An NSG with no rules blocks all traffic. Always include at least an egress rule for outbound connectivity unless intentional isolation is desired.
- **VNIC Attachment Limit**: OCI allows up to 5 NSGs per VNIC. Plan NSG granularity accordingly.

## Use Cases

- **Web Tier Security**: NSG allowing HTTPS/HTTP inbound from the internet with ICMP Path MTU Discovery. Attached to load balancer and web server VNICs. Combined with a private backend NSG on application servers for layered security.
- **Database Isolation**: NSG allowing only TCP port 1521 (Oracle) or 5432 (PostgreSQL) from application-tier NSGs. NSG-to-NSG source type ensures only authorized application servers can connect, regardless of CIDR changes.
- **OKE Cluster Security**: Separate NSGs for the API endpoint (port 6443 from bastion and CI/CD subnets), worker nodes (all traffic from the API endpoint NSG), and load balancer subnets (HTTP/HTTPS from the internet).
- **Zero-Trust Micro-Segmentation**: NSGs referencing other NSGs as sources/destinations instead of CIDR blocks. Traffic is allowed based on group membership, not IP addresses — surviving subnet changes, scaling events, and IP recycling.

## Production Features

This resource provides complete support for production-grade NSG deployments, including:

- **Direction-Split Security Rules**: Ingress and egress rules with full TCP, UDP, and ICMP protocol options.
- **NSG-to-NSG References**: Source and destination can be another NSG OCID for micro-segmented architectures.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical resource topology and outputs.
- **Proto Validation**: Required fields, port range constraints, and the 120-rule limit are validated at the schema level before deployment.
- **Infra-Chart Composability**: Designed as a security building block that downstream resources reference via `StringValueOrRef` for the NSG OCID.

# OCI Virtual Cloud Network: Design Rationale and Research

## Introduction

The OCI VCN component is the networking foundation for every other OCI resource in OpenMCF. Every subnet, compute instance, database, OKE cluster, and load balancer in Oracle Cloud Infrastructure lives inside a VCN. Getting this component right — its spec shape, gateway bundling decisions, and output surface — determines the ergonomics and composability of the entire OCI provider.

This document explains the design decisions behind the OciVcn component, compares OCI networking with AWS and other providers, and documents the research that informed the implementation.

## VCN as the Foundational Layer

In OCI's network architecture, a VCN is an isolated, software-defined network that operates within a single region and compartment. Unlike AWS where a VPC is the primary network boundary, OCI layers an additional organizational primitive — the compartment — above the VCN. Every OCI resource lives in a compartment, and every network-attached resource lives in a VCN within that compartment.

The practical consequence: `compartmentId` is always field 1 in every OCI component spec, and the VCN is the first resource that must exist before any network-dependent resources can be created.

### What OCI Creates Automatically

When a VCN is created, OCI automatically provisions three companion resources:

1. **Default Route Table** — an empty route table with no rules. All subnets use this unless a custom route table is specified.
2. **Default Security List** — allows all egress traffic and SSH (port 22) ingress from any source. This is intentionally permissive to avoid blocking initial connectivity — production deployments should use Network Security Groups instead.
3. **Default DHCP Options** — configured with OCI's internal DNS resolver. Enables DNS resolution within the VCN when a DNS label is set.

OpenMCF exports the OCIDs of all three defaults as stack outputs. Downstream components (OciSubnet in particular) can reference these directly or use custom replacements.

## Why Gateways Are Bundled with the VCN

### The Coupling Argument

Internet, NAT, and Service Gateways are bundled into the OciVcn component rather than being separate resources. This is a deliberate design decision with a clear rationale:

**Gateways are tightly coupled to the VCN lifecycle.** A gateway cannot exist without a VCN, and a gateway is useless outside the context of the VCN it's attached to. They are not shared across VCNs. They are not independently versioned. They are not independently scaled. Their creation, update, and deletion lifecycle is tied to the VCN.

**Contrast with subnets,** which are deliberately separate. Different teams create different subnets with different CIDR blocks, different route tables, and different security rules. Subnets have meaningful independent configuration and are the natural unit of delegation — a platform team creates the VCN and gateways, then application teams create their own subnets within it.

### The Ergonomics Argument

Without bundling, creating a VCN with internet access requires 4 separate resource manifests: VCN, Internet Gateway, NAT Gateway, and Service Gateway. The user must also wire the gateway OCID references into route table rules manually. With bundling, a single manifest with boolean toggles provisions everything:

```yaml
spec:
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
  isServiceGatewayEnabled: true
```

The gateway OCIDs are automatically exported as stack outputs, ready for route table configuration in OciSubnet resources. This reduces a 4-resource orchestration problem to a single resource with toggles.

### What's NOT Bundled (and Why)

- **Route tables** — left as OCI defaults. Custom route tables are configured per-subnet in the OciSubnet component, because different subnets need different routing (public subnets route to the Internet Gateway, private subnets route to the NAT Gateway).
- **Security lists** — left as OCI defaults. OCI recommends using Network Security Groups (OciSecurityGroup) instead of security lists for new deployments. NSGs are stateful, per-VNIC, and more flexible.
- **DHCP options** — left as OCI defaults. Custom DHCP options are uncommon and can be added as a future enhancement if demand emerges.

## OCI VCN vs AWS VPC: A Detailed Comparison

Understanding the differences between OCI VCN and AWS VPC is essential for platform engineers coming from AWS. The components serve the same fundamental purpose but differ in meaningful ways that affect the API shape.

### CIDR Block Model

| Aspect | OCI VCN | AWS VPC |
|--------|---------|---------|
| **CIDRs at creation** | Multiple CIDRs supported natively (`cidr_blocks` is a repeated field) | Single primary CIDR at creation |
| **Adding CIDRs later** | Not supported after creation | Secondary CIDRs can be added via `aws_vpc_ipv4_cidr_block_association` |
| **CIDR range** | /16 to /30 | /16 to /28 |
| **IPv6** | Oracle-assigned /56 GUA prefix (toggle) | Amazon-assigned /56 or BYOIP |

The practical implication: OCI's approach requires planning all CIDR blocks upfront, while AWS allows incremental expansion. OpenMCF's `cidrBlocks` repeated field reflects OCI's native model.

### Gateway Model

| Gateway Type | OCI | AWS Equivalent |
|-------------|-----|----------------|
| **Internet access** | Internet Gateway (1 per VCN) | Internet Gateway (1 per VPC) |
| **Private outbound** | NAT Gateway (1 per VCN, regional) | NAT Gateway (1 per AZ, zonal) |
| **Private service access** | Service Gateway (all OCI services in one resource) | VPC Endpoints (one per service, Gateway or Interface type) |

The most significant difference is the Service Gateway. In AWS, accessing S3 privately requires a Gateway VPC Endpoint for S3; accessing ECR privately requires an Interface VPC Endpoint for ECR, another for ECR's Docker API, and another for S3. In OCI, a single Service Gateway provides private access to all OCI services in one resource.

This is why the Service Gateway implementation uses a data source lookup (`core.GetServices`) to automatically include all available services — the user should never need to enumerate individual services.

### Default Resources

OCI creates a default route table, default security list, and default DHCP options automatically. AWS creates a default route table (main route table), default NACL, and default security group. The semantics are similar, but OCI's security list is a different abstraction from AWS's security group and NACL.

### Compartment vs Account

AWS uses accounts as the primary isolation boundary. Multi-account architectures (via AWS Organizations) are the standard pattern for enterprise isolation.

OCI uses compartments, which are hierarchical and can be nested. A single OCI tenancy can have a deep compartment tree (e.g., `root > platform > production > networking`). IAM policies are scoped to compartments. This means `compartmentId` is a first-class field on every OCI resource, while AWS resources inherit their account from the credential context.

## Service Gateway: Why Auto-Wire to All Services

The Service Gateway implementation in both the Pulumi and Terraform modules performs a data source lookup to discover all available OCI services and wires the gateway to all of them:

**Pulumi:**
```go
allServices, err := core.GetServices(ctx, nil, pulumi.Provider(provider))
for _, svc := range allServices.Services {
    serviceEntries = append(serviceEntries, &core.ServiceGatewayServiceArgs{
        ServiceId: pulumi.String(svc.Id),
    })
}
```

**Terraform:**
```hcl
data "oci_core_services" "all" {
  count = var.spec.is_service_gateway_enabled ? 1 : 0
}
```

**Why auto-wire?** Three reasons:

1. **Service OCIDs vary by region.** The OCID for "All Services in Oracle Services Network" is different in `us-ashburn-1` than in `eu-frankfurt-1`. Users should never need to look up region-specific service OCIDs.
2. **No downside to including all services.** Unlike AWS VPC Endpoints (which have per-hour costs and require per-service configuration), OCI's Service Gateway has no incremental cost for additional services. Including all services is strictly better than including a subset.
3. **Forward compatibility.** As Oracle adds new services to the Oracle Services Network, existing Service Gateways automatically gain private access to them without any configuration change.

## Design Decisions

### Region Is Not in Spec

The region is not a field on OciVcnSpec (or any OCI component spec). It comes from the `OciProviderConfig.region` field in the stack input. This avoids:

- Duplication between the provider config and every resource spec
- Inconsistency if a user accidentally sets different regions on the VCN and its subnets
- Complexity in the spec for a value that is always the same across all resources in a deployment

### DisplayName Falls Back to metadata.name

The `displayName` field is optional. When omitted, the Pulumi module and Terraform module both fall back to `metadata.name`. This means minimal manifests (which only set `metadata.name`) still produce human-readable resources in the OCI Console.

### Freeform Tags Over Defined Tags

OCI supports two tagging systems: freeform tags (key-value strings, no schema) and defined tags (schema-enforced, namespace-scoped). OpenMCF uses freeform tags because:

- They require no pre-configuration (no tag namespace or schema to create first)
- They work across all compartments without additional IAM policies
- They are sufficient for resource tracking, cost allocation, and compliance metadata
- Defined tags can be added as an enhancement if enterprise customers need schema enforcement

### Tagging Includes OpenMCF Metadata

Every VCN and gateway receives freeform tags derived from metadata:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciVcn"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

This enables filtering and cost tracking in the OCI Console and via the OCI API.

## What OpenMCF Supports

### Current Implementation

The OciVcn component covers the core VCN use case:

- **VCN creation** with multiple CIDRs, DNS, IPv6
- **Gateway provisioning** via boolean toggles (Internet, NAT, Service)
- **Automatic tagging** with OpenMCF metadata
- **7 stack outputs** for downstream composability
- **Both Pulumi and Terraform** implementations producing identical resource topology and outputs
- **Proto validation** via buf-validate (required fields, minimum CIDR count)

### What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Custom route tables** — handled by OciSubnet, which creates per-subnet route tables with appropriate gateway references
- **Custom security lists** — OCI recommends Network Security Groups (OciSecurityGroup) for new deployments
- **Custom DHCP options** — uncommon; the OCI default (internal DNS resolver) covers most use cases
- **Local Peering Gateways** — cross-VCN peering within a region is handled by OciDrg (Dynamic Routing Gateway), which supports both local and remote peering
- **DRG attachment** — managed by the OciDrg component, which creates the attachment to the VCN
- **Per-service Service Gateway filtering** — the auto-wire-all approach is simpler and has no downside; individual service selection can be added if a concrete need emerges

## Downstream Dependencies

The OciVcn component is consumed by virtually every other OCI networking and compute component:

```
OciVcn
├── OciSubnet (references vcn_id, gateway IDs for route rules)
├── OciSecurityGroup (references vcn_id)
├── OciApplicationLoadBalancer (deployed into subnets within this VCN)
├── OciContainerEngineCluster (API endpoint in a subnet within this VCN)
├── OciComputeInstance (VNIC attached to a subnet within this VCN)
└── OciDrg (DRG attachment to this VCN for peering/VPN)
```

This makes VCN the highest-leverage component in the OCI provider. Stability, correctness, and output completeness directly impact every downstream component.

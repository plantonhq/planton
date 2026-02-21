# OCI VCN Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Virtual Cloud Networks using the OpenMCF API. Each example demonstrates different use cases from minimal development setups to production-grade configurations with all gateways.

## Table of Contents

- [Example 1: Minimal Development VCN](#example-1-minimal-development-vcn)
- [Example 2: Public-Facing VCN](#example-2-public-facing-vcn)
- [Example 3: Production VCN with All Gateways](#example-3-production-vcn-with-all-gateways)
- [Example 4: Multi-CIDR VCN](#example-4-multi-cidr-vcn)
- [Example 5: IPv6-Enabled VCN](#example-5-ipv6-enabled-vcn)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Development VCN

**Use Case:** Quick development or experimentation environment. Single CIDR, no gateways, no DNS.

**Configuration:**
- **CIDR:** 10.0.0.0/16
- **Gateways:** None
- **DNS:** Disabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: dev-vcn
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f dev-vcn.yaml
```

**What happens:**
- A VCN named `dev-vcn` is created in the specified compartment.
- OCI automatically creates a default route table, default security list, and default DHCP options.
- No gateways are provisioned — the VCN has no internet connectivity.
- Standard OpenMCF freeform tags are applied automatically.

---

## Example 2: Public-Facing VCN

**Use Case:** Workloads that need internet access. An Internet Gateway allows resources in public subnets to accept inbound traffic and communicate outbound. A NAT Gateway allows private subnets to reach the internet without being reachable from it.

**Configuration:**
- **CIDR:** 10.0.0.0/16
- **Gateways:** Internet Gateway + NAT Gateway
- **DNS:** Enabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: web-vcn
  org: my-org
  env: staging
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
  displayName: "Web Tier VCN"
  dnsLabel: "webvcn"
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
```

**Deploy:**

```bash
openmcf apply -f web-vcn.yaml
```

**What happens:**
- A VCN with DNS label `webvcn` is created, enabling the domain `webvcn.oraclevcn.com` for internal hostname resolution.
- An Internet Gateway (`web-vcn-igw`) is attached to the VCN.
- A NAT Gateway (`web-vcn-ngw`) is attached to the VCN.
- The `internet_gateway_id` and `nat_gateway_id` stack outputs are populated and available for route table configuration in OciSubnet resources.

---

## Example 3: Production VCN with All Gateways

**Use Case:** Full production environment with internet access, NAT for private subnets, and a Service Gateway for private OCI service access. DNS enabled for hostname resolution, deletion protection via organization-level policies.

**Configuration:**
- **CIDRs:** 10.0.0.0/16 + 172.16.0.0/16
- **Gateways:** Internet + NAT + Service
- **DNS:** Enabled
- **IPv6:** Disabled (enable if your workloads require it)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: prod-vcn
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..aaaaaaaaprodcompartment"
  cidrBlocks:
    - "10.0.0.0/16"
    - "172.16.0.0/16"
  displayName: "Production VCN"
  dnsLabel: "prodvcn"
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
  isServiceGatewayEnabled: true
```

**Deploy:**

```bash
openmcf apply -f prod-vcn.yaml
```

**What happens:**
- A VCN with two CIDR blocks is created, providing 10.0.0.0/16 for application workloads and 172.16.0.0/16 for database and middleware tiers.
- All three gateways are provisioned and attached to the VCN.
- The Service Gateway is automatically configured for all services in the Oracle Services Network, enabling private access to Object Storage, Autonomous Database, Container Registry, and other OCI services without internet transit.
- Freeform tags include `team: platform` and `cost-center: infrastructure` from the metadata labels.

---

## Example 4: Multi-CIDR VCN

**Use Case:** Large organizations that need to segment address space across multiple CIDR blocks. Common when merging networks or when a single /16 is insufficient.

**Configuration:**
- **CIDRs:** Three non-overlapping blocks for different tiers
- **Gateways:** NAT + Service (no Internet Gateway — all inbound access via load balancers in a separate VCN or via an OCI DRG)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: segmented-vcn
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
    - "10.1.0.0/16"
    - "10.2.0.0/16"
  displayName: "Segmented Network VCN"
  dnsLabel: "segvcn"
  isNatGatewayEnabled: true
  isServiceGatewayEnabled: true
```

**What happens:**
- A VCN with three /16 blocks is created, providing ~196,000 usable IP addresses.
- Subnets can be created across any of the three CIDR blocks, allowing clear separation (e.g., 10.0.x.x for compute, 10.1.x.x for databases, 10.2.x.x for middleware).
- No Internet Gateway — this VCN is fully private, suitable for backend services that communicate outbound only through the NAT Gateway.

---

## Example 5: IPv6-Enabled VCN

**Use Case:** Dual-stack networking for workloads that need IPv6 connectivity, such as IoT backends or services that must be reachable over IPv6.

**Configuration:**
- **CIDR:** 10.0.0.0/16 (IPv4) + Oracle-assigned /56 (IPv6)
- **Gateways:** Internet + NAT
- **IPv6:** Enabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: dualstack-vcn
  org: my-org
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
  displayName: "Dual-Stack VCN"
  dnsLabel: "dsvcn"
  isIpv6Enabled: true
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
```

**What happens:**
- A VCN is created with both the specified IPv4 CIDR and an Oracle-assigned /56 IPv6 GUA prefix.
- Subnets created within this VCN can be configured for dual-stack, receiving both IPv4 and IPv6 address ranges.
- The Internet Gateway supports both IPv4 and IPv6 traffic.

---

## Common Operations

### Get VCN ID After Deployment

```bash
# Pulumi
pulumi stack output vcn_id

# Terraform
terraform output vcn_id
```

### Use VCN ID in a Downstream OciSubnet

The `vcn_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in downstream resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: app-subnet
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    valueFrom:
      kind: OciVcn
      name: prod-vcn
      fieldPath: status.outputs.vcnId
  cidrBlock: "10.0.1.0/24"
  displayName: "Application Subnet"
```

### Reference Gateway IDs for Route Tables

When configuring route rules in an OciSubnet, reference the gateway IDs from the VCN:

```yaml
spec:
  routeRules:
    - destinationCidr: "0.0.0.0/0"
      networkEntityId:
        valueFrom:
          kind: OciVcn
          name: prod-vcn
          fieldPath: status.outputs.internetGatewayId
```

### Destroy a VCN

A VCN must be empty (no subnets, no gateways beyond those managed by this resource) before it can be deleted. OpenMCF handles gateway deletion automatically since gateways are part of this resource.

```bash
openmcf destroy -f vcn.yaml
```

---

## Best Practices

### Choose Gateways Based on Your Network Topology

| Gateway | Enable When | Typical Use |
|---------|-------------|-------------|
| **Internet Gateway** | You need resources directly reachable from the internet (load balancers, bastion hosts) | Public subnets |
| **NAT Gateway** | Private resources need outbound internet access (pulling images, calling external APIs) | Private subnets with outbound-only access |
| **Service Gateway** | Any production environment using OCI services (Object Storage, Container Registry, databases) | All production VCNs |

**Recommendation:** For production VCNs, enable all three gateways. The Service Gateway has no cost beyond the VCN and prevents OCI service traffic from traversing the internet — a security and performance win.

### Plan CIDRs Before Creation

- Use /16 blocks for production VCNs to allow ample subnet space.
- Avoid overlapping with on-premises networks if you plan to use OCI DRG for VPN or FastConnect peering.
- Use multiple CIDRs for address segmentation across tiers (compute, database, middleware) rather than one large block.

### Set DNS Labels Consistently

- Use short, meaningful labels: `prodvcn`, `devvcn`, `okevcn`.
- DNS labels are immutable — choose carefully at creation time.
- The resulting VCN domain (`<dnsLabel>.oraclevcn.com`) becomes part of the FQDN for all resources in the VCN.

### One VCN Per Environment

Use separate VCNs for dev, staging, and production. This provides:
- Network isolation between environments
- Independent gateway and route table management
- Clear blast radius containment
- Separate security list and NSG policies

### Tag for Cost and Compliance

Metadata labels are applied as OCI freeform tags. Use consistent labels across all resources:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
```

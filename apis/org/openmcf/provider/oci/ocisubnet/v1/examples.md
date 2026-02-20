# OCI Subnet Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure subnets using the OpenMCF API. Each example demonstrates different use cases from minimal private subnets to production-grade configurations with inline routing and cross-resource references.

## Table of Contents

- [Example 1: Minimal Private Subnet](#example-1-minimal-private-subnet)
- [Example 2: Public Subnet with DNS](#example-2-public-subnet-with-dns)
- [Example 3: Private Subnet with NAT Gateway Routing](#example-3-private-subnet-with-nat-gateway-routing)
- [Example 4: AD-Specific Subnet](#example-4-ad-specific-subnet)
- [Example 5: Full-Featured Production Subnet](#example-5-full-featured-production-subnet)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Private Subnet

**Use Case:** Application backend or database tier that should not be reachable from the internet. VNICs in this subnet cannot have public IPs and all inbound internet traffic is blocked.

**Configuration:**
- **CIDR:** 10.0.10.0/24
- **Access:** Private (no public IPs, no internet ingress)
- **Route Table:** VCN default
- **DNS:** Enabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: private-backend
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.10.0/24"
  dnsLabel: "privbe"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f private-backend.yaml
```

**What happens:**
- A regional subnet named `private-backend` is created spanning all availability domains.
- VNICs in this subnet cannot have public IPs assigned.
- Inbound internet traffic is blocked at the subnet level, even if security rules or NSGs would allow it.
- The VCN's default route table is used (no inline rules or external reference provided).
- DNS label `privbe` enables hostname resolution at `privbe.<vcnDnsLabel>.oraclevcn.com`.

---

## Example 2: Public Subnet with DNS

**Use Case:** Load balancers and bastion hosts that need direct internet-facing access. Public IPs can be assigned to VNICs and internet ingress is allowed.

**Configuration:**
- **CIDR:** 10.0.0.0/24
- **Access:** Public (default — both `prohibitPublicIpOnVnic` and `prohibitInternetIngress` are `false`)
- **Route Table:** VCN default
- **DNS:** Enabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: public-lb
  org: my-org
  env: staging
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.0.0/24"
  displayName: "Load Balancer Subnet"
  dnsLabel: "publb"
```

**Deploy:**

```bash
openmcf apply -f public-lb.yaml
```

**What happens:**
- A regional public subnet is created. Both `prohibitPublicIpOnVnic` and `prohibitInternetIngress` default to `false`, so VNICs can have public IPs and inbound traffic is permitted.
- The display name in the OCI Console is "Load Balancer Subnet" rather than the metadata name.
- DNS label `publb` enables hostname resolution at `publb.<vcnDnsLabel>.oraclevcn.com`.
- The VCN's default route table is used. For internet routing, the VCN's default route table should include a rule pointing to the Internet Gateway.

---

## Example 3: Private Subnet with NAT Gateway Routing

**Use Case:** Application tier that needs outbound internet access (e.g., pulling container images, calling external APIs) but must not be directly reachable from the internet. A dedicated route table routes traffic through the NAT Gateway for internet and the Service Gateway for OCI services.

**Configuration:**
- **CIDR:** 10.0.20.0/24
- **Access:** Private
- **Route Table:** Inline (2 rules: NAT Gateway for internet, Service Gateway for OCI services)
- **DNS:** Enabled

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: app-tier
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..aaaaaaaaprodcompartment"
  vcnId:
    value: "ocid1.vcn.oc1.iad.aaaaaaaaprodvcn"
  cidrBlock: "10.0.20.0/24"
  displayName: "Application Tier"
  dnsLabel: "apptier"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
  routeRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      networkEntityId:
        value: "ocid1.natgateway.oc1.iad.examplenatgw"
      description: "Internet traffic via NAT Gateway"
    - destination: "all-iad-services-in-oracle-services-network"
      destinationType: service_cidr_block
      networkEntityId:
        value: "ocid1.servicegateway.oc1.iad.examplesgw"
      description: "OCI services via Service Gateway"
```

**Deploy:**

```bash
openmcf apply -f app-tier.yaml
```

**What happens:**
- A private regional subnet is created with a dedicated route table named `Application Tier-rt`.
- All internet-bound traffic (0.0.0.0/0) is routed through the NAT Gateway — the subnet has outbound internet access but no inbound.
- Traffic to OCI services (Object Storage, Container Registry, Autonomous Database) is routed through the Service Gateway, staying on the Oracle backbone network.
- The `route_table_id` output reflects the ID of the newly created route table.

---

## Example 4: AD-Specific Subnet

**Use Case:** Legacy workloads or services that require placement in a specific availability domain. AD-specific subnets are uncommon in modern OCI architectures but may be necessary for bare metal instances or specific compliance requirements.

**Configuration:**
- **CIDR:** 10.0.40.0/24
- **Access:** Private
- **Availability Domain:** US-ASHBURN-AD-1
- **Route Table:** VCN default

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: ad1-subnet
  org: my-org
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.40.0/24"
  displayName: "AD-1 Subnet"
  dnsLabel: "ad1sub"
  availabilityDomain: "Iocq:US-ASHBURN-AD-1"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
```

**What happens:**
- A subnet scoped to a single availability domain (AD-1) is created. Resources in this subnet can only be placed in AD-1.
- Unlike regional subnets, AD-specific subnets do not provide cross-AD high availability. You would need one subnet per AD if you need workloads in multiple ADs.
- Regional subnets are recommended for most use cases. Use AD-specific subnets only when a specific AD constraint exists.

---

## Example 5: Full-Featured Production Subnet

**Use Case:** Production private subnet with all available configuration options. Includes DNS, inline routing, custom DHCP options, security lists, and IPv6 for dual-stack networking.

**Configuration:**
- **CIDR:** 10.0.50.0/24 (IPv4) + 2001:0db8:0123:4500::/64 (IPv6)
- **Access:** Private
- **Route Table:** Inline (NAT + Service Gateway)
- **DNS:** Enabled
- **Security Lists:** Custom
- **DHCP Options:** Custom

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: prod-app
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..aaaaaaaaprodcompartment"
  vcnId:
    value: "ocid1.vcn.oc1.iad.aaaaaaaaprodvcn"
  cidrBlock: "10.0.50.0/24"
  displayName: "Production Application Subnet"
  dnsLabel: "prodapp"
  ipv6CidrBlock: "2001:0db8:0123:4500::/64"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
  dhcpOptionsId:
    value: "ocid1.dhcpoptions.oc1.iad.exampledhcp"
  securityListIds:
    - value: "ocid1.securitylist.oc1.iad.examplesl1"
    - value: "ocid1.securitylist.oc1.iad.examplesl2"
  routeRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      networkEntityId:
        value: "ocid1.natgateway.oc1.iad.examplenatgw"
      description: "Internet traffic via NAT Gateway"
    - destination: "all-iad-services-in-oracle-services-network"
      destinationType: service_cidr_block
      networkEntityId:
        value: "ocid1.servicegateway.oc1.iad.examplesgw"
      description: "OCI services via Service Gateway"
```

**What happens:**
- A production-grade private subnet is created with dual-stack networking (IPv4 + IPv6).
- Custom DHCP options override the VCN defaults (e.g., for custom DNS resolvers).
- Two security lists are associated with the subnet (within the 5-list maximum).
- A dedicated route table (`Production Application Subnet-rt`) is created with NAT and Service Gateway routes.
- Freeform tags include `team: platform` and `cost-center: infrastructure` from the metadata labels.

---

## Common Operations

### Get Subnet ID After Deployment

```bash
# Pulumi
pulumi stack output subnet_id

# Terraform
terraform output subnet_id
```

### Use Subnet ID in a Downstream OciComputeInstance

The `subnet_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in downstream resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: app-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: private-backend
      fieldPath: status.outputs.subnetId
  shape: "VM.Standard.E4.Flex"
```

### Reference Gateway IDs from VCN for Route Rules

When configuring inline route rules, reference gateway IDs from the parent VCN using `valueFrom`:

```yaml
spec:
  routeRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      networkEntityId:
        valueFrom:
          kind: OciVcn
          name: prod-vcn
          fieldPath: status.outputs.natGatewayId
      description: "Internet traffic via NAT Gateway"
```

### Destroy a Subnet

A subnet must be empty (no compute instances, load balancers, or other resources attached) before deletion:

```bash
openmcf destroy -f subnet.yaml
```

If the subnet has an inline route table, the route table is deleted along with the subnet.

---

## Best Practices

### Choose Public vs Private Based on Access Requirements

| Subnet Type | `prohibitPublicIpOnVnic` | `prohibitInternetIngress` | Use For |
|-------------|--------------------------|---------------------------|---------|
| **Public** | `false` (default) | `false` (default) | Load balancers, bastion hosts, internet-facing services |
| **Private** | `true` | `true` | Application backends, databases, OKE worker nodes |
| **Hybrid** | `false` | `true` | VNICs can have public IPs but inbound internet traffic is still blocked at the subnet level |

**Recommendation:** Set both `prohibitPublicIpOnVnic: true` and `prohibitInternetIngress: true` for any subnet that does not need direct internet access. Defense in depth — even if a security rule accidentally allows inbound traffic, the subnet-level block prevents it.

### Plan CIDR Blocks Within the VCN

- Use /24 blocks for most subnets (251 usable IPs — sufficient for most workloads).
- Use /28 for small subnets (11 usable IPs — load balancers, bastion hosts).
- Use /20 or /19 for large subnets (4,091 or 8,187 usable IPs — OKE node pools that may scale significantly).
- All subnet CIDRs must fall within one of the parent VCN's CIDR blocks and must not overlap with each other.

### Use Regional Subnets Unless Constrained

Regional subnets span all availability domains in a region. This is the recommended default because:
- Resources deployed into a regional subnet can be placed in any AD.
- High availability across ADs requires fewer subnets.
- OKE node pools, load balancers, and most compute workloads work well with regional subnets.

Use AD-specific subnets only when a workload requires placement in a specific AD (e.g., bare metal instances, compliance requirements).

### Set DNS Labels Consistently

- Use short, meaningful labels: `publb`, `privapp`, `dbtier`, `okenode`.
- DNS labels are immutable — choose carefully at creation time.
- The resulting subnet domain (`<dnsLabel>.<vcnDnsLabel>.oraclevcn.com`) becomes part of the FQDN for resources in the subnet.
- Both the subnet and the parent VCN must have DNS labels for subnet DNS to function.

### Prefer Inline Route Rules for Dedicated Routing

Three options for route table association:

1. **Inline `routeRules`**: Best for subnets that own their routing logic (most production subnets). A dedicated route table is created and managed as part of the subnet lifecycle.
2. **External `routeTableId`**: Use when multiple subnets share the same route table, or when the route table is managed by a separate team or process.
3. **Neither (VCN default)**: Suitable for development environments or subnets that use the same routing as every other subnet in the VCN.

### Use NSGs Instead of Security Lists for New Deployments

Security lists (`securityListIds`) are supported for backward compatibility, but OCI recommends Network Security Groups (OciSecurityGroup) for new deployments. NSGs are:
- Stateful (security lists are also stateful, but NSGs are per-VNIC rather than per-subnet)
- More granular (attach to specific VNICs, not entire subnets)
- Easier to manage in multi-tier architectures

Use security lists only when you need subnet-wide rules that apply to all VNICs regardless of their NSG membership.

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

# Scaleway Private Network Examples

This document provides practical examples for deploying Scaleway Private Networks using the OpenMCF API. Each example demonstrates different use cases from minimal development setups to production multi-tier architectures.

## Table of Contents

- [Example 1: Minimal Development Network](#example-1-minimal-development-network)
- [Example 2: Network with Explicit Subnet](#example-2-network-with-explicit-subnet)
- [Example 3: Production Network with valueFrom VPC Reference](#example-3-production-network-with-valuefrom-vpc-reference)
- [Example 4: Multi-Tier Architecture](#example-4-multi-tier-architecture)
- [Example 5: Dual-Stack Network with IPv6](#example-5-dual-stack-network-with-ipv6)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Development Network

**Use Case:** Quick development environment. A single Private Network inside a VPC, with automatic CIDR allocation.

**Configuration:**
- **VPC:** Referenced by literal ID
- **Region:** fr-par (Paris)
- **Subnet:** Auto-allocated by Scaleway IPAM

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: dev-network
  org: my-org
  env: dev
spec:
  vpcId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  region: fr-par
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f dev-network.yaml
```

**What happens:**
- A Private Network named `dev-network` is created in the `fr-par` region inside the specified VPC.
- Scaleway IPAM automatically allocates an IPv4 subnet.
- Standard OpenMCF tags are applied automatically.

---

## Example 2: Network with Explicit Subnet

**Use Case:** Production network where you need to control the IP address range. Essential when multiple Private Networks in the same VPC need non-overlapping CIDRs for routing.

**Configuration:**
- **VPC:** Referenced by literal ID
- **Region:** fr-par (Paris)
- **Subnet:** Explicitly set to 10.0.1.0/24

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
  org: my-org
  env: prod
spec:
  vpcId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
  enableDefaultRoutePropagation: true
```

**What happens:**
- A Private Network is created with the specific 10.0.1.0/24 subnet.
- Default route propagation is enabled, allowing resources to communicate with other networks in the VPC.

---

## Example 3: Production Network with valueFrom VPC Reference

**Use Case:** The most common production pattern. The Private Network references a ScalewayVpc resource using `valueFrom`, enabling the platform to build a dependency DAG and deploy in the correct order.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: prod-vpc
  org: my-org
  env: prod
spec:
  region: fr-par
  enableRouting: true
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
  enableDefaultRoutePropagation: true
```

**What happens:**
- The platform resolves the VPC reference, ensures the VPC is created first, then creates the Private Network.
- The Private Network's `vpc_id` is automatically populated from the VPC's stack output.

**Using the Private Network ID in downstream resources:**

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: app-cluster
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  region: fr-par
  # ... other cluster config
```

---

## Example 4: Multi-Tier Architecture

**Use Case:** A production environment with separate Private Networks for application, database, and cache tiers. Each has its own non-overlapping subnet. All share a VPC with routing enabled so they can communicate.

```yaml
# Foundation: VPC with routing enabled
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: prod-vpc
  org: my-org
  env: prod
spec:
  region: fr-par
  enableRouting: true
---
# Tier 1: Application network (Kapsule cluster)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
  enableDefaultRoutePropagation: true
---
# Tier 2: Database network (RDB, Redis)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: db-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.2.0/24"
  enableDefaultRoutePropagation: true
---
# Tier 3: Cache network (Redis)
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: cache-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.3.0/24"
  enableDefaultRoutePropagation: true
```

**What happens:**
- Three Private Networks are created inside the same VPC.
- Each has a unique, non-overlapping subnet (10.0.1.0/24, 10.0.2.0/24, 10.0.3.0/24).
- With VPC routing enabled and default route propagation on each network, resources in any tier can communicate with resources in other tiers.

---

## Example 5: Dual-Stack Network with IPv6

**Use Case:** A network that supports both IPv4 and IPv6 for workloads requiring IPv6 connectivity.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: dual-stack-network
  org: my-org
  env: prod
spec:
  vpcId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  region: fr-par
  ipv4Subnet: "10.0.10.0/24"
  ipv6Subnets:
    - "fd46:78ab:30b8:177c::/64"
```

---

## Common Operations

### Get Private Network ID After Deployment

```bash
# Pulumi
pulumi stack output private_network_id

# Terraform
terraform output private_network_id
```

### Get Allocated Subnet

```bash
# Pulumi
pulumi stack output ipv4_subnet_cidr

# Terraform
terraform output ipv4_subnet_cidr
```

### Use Private Network ID in Another Resource

The `private_network_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in downstream resources:

```yaml
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
```

### Destroy a Private Network

A Private Network must have no resources attached before it can be deleted. Detach all Kapsule clusters, databases, instances, etc. first.

```bash
# Pulumi
pulumi destroy

# Terraform
terraform destroy
```

---

## Best Practices

### Plan CIDRs for Multi-Tier Architectures

When multiple Private Networks share a VPC with routing enabled, their IPv4 CIDRs **must not overlap**. Use a consistent addressing scheme:

| Tier | Suggested CIDR | Purpose |
|------|---------------|---------|
| Application | 10.0.1.0/24 | Kapsule clusters, instances |
| Database | 10.0.2.0/24 | RDB instances, MongoDB |
| Cache | 10.0.3.0/24 | Redis clusters |
| Gateway | 10.0.4.0/24 | Public Gateway, NAT |

### Enable Route Propagation for Multi-Network VPCs

If you have more than one Private Network in a VPC and resources need to communicate across networks, enable both:
1. `enableRouting: true` on the VPC.
2. `enableDefaultRoutePropagation: true` on each Private Network.

### Use valueFrom for Infra Chart Composability

Always use `valueFrom` references (not hardcoded IDs) when deploying as part of an infra chart. This enables:
- Automatic dependency ordering.
- Clean environment promotion (dev/staging/prod use same template, different names).
- Impact analysis in the platform's dependency graph.

### One Network Per Tier

Separate concerns into distinct Private Networks. This provides:
- **Network isolation**: A compromised application tier can't directly access the database network.
- **Independent scaling**: Each tier's address space can be sized independently.
- **Clear ownership**: Teams can own and manage their tier's network configuration.

### Match Regions

The Private Network region **must** match its parent VPC region. This is enforced by the Scaleway API. Double-check region consistency when writing infra chart templates.

# Scaleway VPC Examples

This document provides practical examples for deploying Scaleway Virtual Private Cloud (VPC) networks using the OpenMCF API. Each example demonstrates different use cases from minimal development setups to production-grade multi-tier architectures.

## Table of Contents

- [Example 1: Minimal Development VPC](#example-1-minimal-development-vpc)
- [Example 2: Production VPC with Routing](#example-2-production-vpc-with-routing)
- [Example 3: Advanced VPC with Custom Routes](#example-3-advanced-vpc-with-custom-routes)
- [Example 4: Multi-Region Setup](#example-4-multi-region-setup)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Development VPC

**Use Case:** Quick development environment. Just a logical container for a single Private Network.

**Configuration:**
- **Region:** fr-par (Paris)
- **Routing:** Disabled (single Private Network, no need)
- **Cost:** Free (Scaleway VPCs have no cost)

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: dev-vpc
  org: my-org
  env: dev
spec:
  region: fr-par
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f dev-vpc.yaml
```

**What happens:**
- A VPC named `dev-vpc` is created in the `fr-par` region.
- Routing is disabled (default). This VPC is suitable for a single Private Network.
- Standard OpenMCF tags are applied automatically.

---

## Example 2: Production VPC with Routing

**Use Case:** Production environment with multiple Private Networks that need to communicate. A Kapsule cluster in one Private Network talks to an RDB instance in another.

**Configuration:**
- **Region:** fr-par (Paris)
- **Routing:** Enabled (Private Networks can communicate)

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
```

**Deploy:**

```bash
openmcf apply -f prod-vpc.yaml
```

**What happens:**
- A VPC with inter-Private-Network routing is created.
- Resources in different Private Networks attached to this VPC can communicate.
- **WARNING:** Routing cannot be disabled after creation. This is a one-way toggle.

**Using the VPC ID in downstream resources (ScalewayPrivateNetwork):**

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
```

---

## Example 3: Advanced VPC with Custom Routes

**Use Case:** VPC with custom route propagation for advanced networking scenarios, such as a VPN gateway that advertises routes to other Private Networks.

**Configuration:**
- **Region:** nl-ams (Amsterdam)
- **Routing:** Enabled
- **Custom Routes Propagation:** Enabled

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: gateway-vpc
  org: my-org
  env: prod
spec:
  region: nl-ams
  enableRouting: true
  enableCustomRoutesPropagation: true
```

**Deploy:**

```bash
openmcf apply -f gateway-vpc.yaml
```

**What happens:**
- A VPC with both routing and custom route propagation is created.
- Custom routes from any Private Network are advertised to all other Private Networks in this VPC.
- **WARNING:** Both flags are one-way toggles. Neither can be disabled after creation.

---

## Example 4: Multi-Region Setup

**Use Case:** Separate VPCs for different regions in a global deployment.

```yaml
# Paris region
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: eu-west-vpc
  org: my-org
  env: prod
spec:
  region: fr-par
  enableRouting: true
---
# Amsterdam region
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: eu-central-vpc
  org: my-org
  env: prod
spec:
  region: nl-ams
  enableRouting: true
---
# Warsaw region
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: eu-east-vpc
  org: my-org
  env: prod
spec:
  region: pl-waw
  enableRouting: true
```

---

## Common Operations

### Get VPC ID After Deployment

```bash
# Pulumi
pulumi stack output vpc_id

# Terraform
terraform output vpc_id
```

### Use VPC ID in Another Resource

The `vpc_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in downstream resources:

```yaml
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
```

### Destroy a VPC

A VPC must be empty (no Private Networks attached) before it can be deleted.

```bash
# Pulumi
pulumi destroy

# Terraform
terraform destroy
```

---

## Best Practices

### Plan Routing Before Creation

Routing is a one-way toggle. If you think you might need multi-tier networking in the future, enable routing at creation time. It's free and has no performance penalty.

**Recommendation:** For production VPCs, always enable routing. The cost is zero and it prevents having to recreate the VPC later.

### One VPC Per Environment

Use separate VPCs for dev, staging, and production environments. This provides network isolation and makes it impossible for development resources to accidentally communicate with production databases.

### Regional Architecture

Scaleway VPCs are regional. For multi-region deployments:
- Create one VPC per region.
- Use consistent naming: `{env}-{region}-vpc` (e.g., `prod-fr-par-vpc`).
- Cross-region communication requires external mechanisms (public endpoints, VPN, etc.).

### Enable Routing for Infra Charts

The `kapsule-environment` infra chart creates a VPC with routing enabled by default, because Kapsule clusters need to communicate with databases and other services in separate Private Networks. When building custom infra charts, follow this pattern.

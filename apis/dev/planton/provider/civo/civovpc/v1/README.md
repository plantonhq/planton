# CivoVpc

## Overview

**API Version:** `civo.planton.dev/v1`  
**Kind:** `CivoVpc`  
**ID Prefix:** `civpc`

`CivoVpc` is a Planton Cloud Resource that creates and manages isolated private networks (VPCs) on the Civo cloud platform. Civo's private networks provide strict Layer 3 network isolation for Kubernetes clusters, compute instances, and load balancers.

## Key Features

- **True Network Isolation**: OpenStack-based networking ensures complete tenant isolation at Layer 3
- **Regional Networks**: Each VPC is region-specific (e.g., LON1, NYC1, FRA1) with no cross-region peering
- **Simplified Architecture**: No subnets, no availability zones—just a flat /24 network
- **Zero Egress Fees**: No data transfer costs, enabling best-practice network segmentation without financial penalty
- **Auto-Allocation**: Optional CIDR auto-allocation for quick dev/test environments
- **Permanent Assignment**: Resources (clusters, instances) cannot be moved between networks after creation

## Architecture

### The /24 Constraint

Civo networks are limited to a maximum prefix size of **/24** (256 IP addresses). This differs from AWS VPCs (up to /16, 65,536 IPs) or GCP VPCs (custom sizes). However, this constraint is viable because:

- **Kubernetes clusters use overlay networking** (Flannel or Cilium)
- Pods get IPs from a separate virtual network (e.g., `10.42.0.0/16`)
- The VPC's /24 address space is only consumed by:
  - Kubernetes cluster nodes (typically 3-10)
  - Standalone compute instances
  - Civo Load Balancers

### Design Pattern: Many Small Networks

Embrace **many small, isolated networks** instead of one monolithic network:

```
prod-lon1-network:    10.10.1.0/24  (production in London)
stage-lon1-network:   10.10.2.0/24  (staging in London)
dev-nyc1-network:     10.20.1.0/24  (dev in New York)
```

**CIDR Planning Schema:**
```
10.A.B.0/24
  A = Region ID (e.g., 10 for LON1, 20 for NYC1, 30 for FRA1)
  B = Environment ID (e.g., 1 for prod, 2 for staging, 3 for dev)
```

### Two-Tier Security Model

1. **Layer 1: Platform Firewall (`civo_firewall`)**
   - Controls North-South traffic (entering/leaving the network)
   - Stateful firewall attached to the VPC
   - Default-deny: all traffic blocked unless explicitly allowed
   - Use case: Open port 6443 for K8s API, 80/443 for web traffic

2. **Layer 2: Kubernetes CNI Network Policy**
   - Controls East-West traffic (pod-to-pod communication)
   - Standard Kubernetes `NetworkPolicy` objects
   - Use case: Zero-trust security within the cluster

**Best Practice:** Use both layers. Civo firewalls protect the network perimeter; CNI policies enforce intra-cluster security.

## Critical Constraints

### 1. Permanent Network Assignment

**Once a resource (Kubernetes cluster or instance) is assigned to a network, it cannot be moved.**

- The only migration path is **destroy and recreate**
- Changing `network_id` in IaC triggers a destructive replacement
- Always run `terraform plan` or `pulumi preview` before applying network changes

### 2. Regional Isolation

**Networks are region-specific and not globally routable.**

- A network in `LON1` is not visible in `NYC1`
- No native inter-region peering
- Multi-region workloads communicate via public IPs (secured with TLS) or self-managed VPNs

### 3. No Default Network for Production

Every Civo region has a "Default" network that auto-provisions resources without explicit network configuration. **Never use this for production.** Always create dedicated networks for each environment.

## Specification

### CivoVpcSpec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `civo_credential_id` | string | Yes | Reference to the Civo API credential |
| `network_name` | string | Yes | DNS-friendly label for the network (lowercase alphanumeric + hyphens) |
| `region` | string | Yes | Civo region code (e.g., `LON1`, `NYC1`, `FRA1`) |
| `ip_range_cidr` | string | No | IPv4 CIDR block (max /24). If omitted, Civo auto-allocates from available address pools |
| `is_default_for_region` | bool | No | Whether this network should be the default for the region. Only one default per region. Default: `false` |
| `description` | string | No | Human-readable description (max 100 characters) |

### Stack Outputs

After creating a Civo VPC, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `network_id` | string | The unique ID of the created network (used as `network_id` in clusters, instances, firewalls) |
| `cidr_block` | string | The actual IPv4 CIDR block (auto-allocated if not specified) |
| `created_at_rfc3339` | string | Timestamp when the network was created (RFC 3339 format) |

## Usage

### Basic Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: dev-test-network
spec:
  civoCredentialId: civo-cred-123
  networkName: dev-test-network
  region: LON1
  description: "Development test network"
```

**Rationale:**
- No `ip_range_cidr` specified—Civo auto-allocates from available address pools
- Simplifies dev workflows (no CIDR planning required)
- Not default network (`is_default_for_region` defaults to `false`)

### Production Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: prod-main-network
spec:
  civoCredentialId: civo-cred-123
  networkName: prod-main-network
  region: NYC1
  ipRangeCidr: "10.20.1.0/24"
  description: "Production network (NYC1)"
```

**Rationale:**
- Explicit CIDR following hierarchical schema (`10.20.1.0/24` = NYC1 region, prod environment)
- Prevents conflicts if adding staging (`10.20.2.0/24`) or dev (`10.20.3.0/24`) later
- Safe for future VPN connectivity without overlapping address spaces

### Multi-Region Example

```yaml
# Production - London
---
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: prod-lon1-network
spec:
  civoCredentialId: civo-cred-123
  networkName: prod-lon1-network
  region: LON1
  ipRangeCidr: "10.10.1.0/24"
  description: "Production network (London)"

# Production - Frankfurt
---
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: prod-fra1-network
spec:
  civoCredentialId: civo-cred-123
  networkName: prod-fra1-network
  region: FRA1
  ipRangeCidr: "10.30.1.0/24"
  description: "Production network (Frankfurt)"
```

**Rationale:**
- Separate networks per region (Civo's constraint)
- Non-overlapping CIDRs using hierarchical schema (LON1: 10.10.x, FRA1: 10.30.x)
- Both use environment ID `1` for production
- Enables future multi-region connectivity via self-managed VPN

For more examples, see [examples.md](./examples.md).

## Best Practices

### ✅ Do

1. **Plan CIDR Blocks Upfront**
   - Use a hierarchical schema (`10.A.B.0/24`) to prevent future conflicts
   - Document your CIDR allocation in version control

2. **Create Dedicated Networks per Environment**
   - Separate networks for dev, staging, and prod in each region
   - Never mix environments in the same network

3. **Use IaC from Day One**
   - Even for dev environments
   - Manual networks lead to drift and misconfiguration

4. **Implement Two-Tier Security**
   - Define Civo firewalls for perimeter security (North-South)
   - Use Kubernetes Network Policies for intra-cluster security (East-West)

5. **Preview Changes Before Applying**
   - Always run `terraform plan` or `pulumi preview`
   - Watch for resources marked as "replacement"—these are destructive

### ❌ Don't

1. **Deploy Production to the Default Network**
   - No isolation, no predictability, no security boundary

2. **Copy-Paste CIDR Blocks**
   - Forgetting to change `cidr_v4` when reusing configs for different environments
   - Leads to overlapping address spaces that break routing

3. **Assume Network Changes Are Non-Destructive**
   - Changing `network_id` on a cluster or instance triggers replacement
   - Always preview changes and plan migration windows

4. **Mix Security Layers**
   - Don't assume the Civo firewall controls pod-to-pod traffic
   - Don't forget to configure CNI policies

5. **Ignore the /24 Limit**
   - Don't try to build a monolithic network with thousands of instances
   - Embrace the "many small networks" pattern

## Implementation Details

### Under the Hood: Pulumi (Go)

Planton uses **Pulumi (Go)** for Civo VPC provisioning:

- **Language Consistency**: Our multi-cloud orchestration is Go-based
- **Equivalent Coverage**: Pulumi's Civo provider (bridged from Terraform) supports all network operations
- **Programming Model**: Code-based approach simplifies conditional logic and custom integrations

**Note:** The choice of Pulumi vs. Terraform is an implementation detail. The protobuf API remains the same regardless of the underlying IaC tool.

### Default Choices

- **CIDR Auto-Allocation**: If `ip_range_cidr` is empty, Civo auto-allocates. Simplifies dev/test workflows. Always specify explicit CIDR for production.
- **Not Default Network**: We default `is_default_for_region` to `false` to prevent accidental misuse.
- **No VLAN Fields**: Civo's API includes VLAN-related fields for private cloud deployments (CivoStack Enterprise). We omit these from the public cloud API to avoid confusion.

## 80/20 Design Philosophy

This component follows the **80/20 principle**: 80% of users need only the essential fields we expose. The remaining 20% (advanced use cases like VLAN integration for private cloud) are intentionally omitted to avoid API clutter and confusion.

## Migration Considerations

- **Migration to Civo**: Trivial. Define a new `CivoVpc` and deploy workloads into it.
- **Migration within Civo**: No "move" pattern exists. Resources must be **destroyed and recreated** in the new network. This is Civo's platform constraint, not a Planton limitation.

## Civo's Philosophy: Simplicity as a Feature

Civo deliberately omits complex features found in hyperscalers:

- No VPC Peering
- No Transit Gateways
- No PrivateLink/VPC Endpoints
- No dynamic CIDR expansion

**This is not a limitation—it's a design choice.** Civo focuses on the 80% of workloads that need:

- Strict isolation
- Transparent pricing
- Zero egress fees
- Operational simplicity

For teams building cloud-native applications on Kubernetes, Civo's model is refreshingly straightforward.

## Further Reading

- **Comprehensive Deployment Guide**: [docs/README.md](./docs/README.md) - Deep research on deployment methods, IaC comparison, and production patterns
- **Configuration Examples**: [examples.md](./examples.md) - Practical configurations for dev, staging, production, and multi-tenant scenarios
- **Civo VPC Documentation**: [Civo Docs - Private Networks](https://www.civo.com/docs/networking/private-networks)
- **Why Create Multiple Networks?**: [Civo Learn - Multi-Network Strategies](https://www.civo.com/learn/why-create-multiple-networks)
- **Civo Firewalls**: [Civo Docs - Firewalls](https://www.civo.com/docs/networking/firewalls)
- **Kubernetes Network Policies**: [Civo Learn - Network Policies with Cilium](https://www.civo.com/learn/network-policies-with-cilium)

## Related Resources

- **CivoFirewall**: Define firewall rules for North-South traffic
- **CivoKubernetesCluster**: Create Kubernetes clusters within a VPC
- **CivoInstance**: Deploy compute instances within a VPC

---

**Bottom Line**: CivoVpc provides simple, regional, /24-limited networks with true isolation for cloud-native workloads. Manage them with IaC (never manually), plan your CIDR scheme upfront, and embrace the "many small networks" pattern. Planton abstracts Civo's networking into a protobuf API that works consistently across clouds while respecting Civo's deliberate simplicity.


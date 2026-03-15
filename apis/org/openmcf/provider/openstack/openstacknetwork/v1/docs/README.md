# OpenStack Neutron Network — Research Documentation

## Introduction

OpenStack Neutron provides Software Defined Networking (SDN) as a service. At its core, a **network** is the most fundamental abstraction: an isolated Layer 2 broadcast domain that provides connectivity to virtual machines, containers, bare-metal servers, and other services running on an OpenStack cloud.

Every meaningful OpenStack workload requires at least one network. Subnets carve IP address space out of a network. Ports attach to a network and receive an IP address from a subnet. Routers connect networks together and provide external gateway access. Security groups filter traffic on ports. Floating IPs are allocated from external networks and associated with ports. Compute instances boot with one or more network interfaces.

The network is the root of the OpenStack networking dependency tree, and OpenMCF's `OpenStackNetwork` component reflects this: it has no foreign key dependencies, but nearly every other OpenStack component references it.

## Historical Context

OpenStack networking has evolved significantly since the project's inception:

**Nova-network era (2010–2014):** The original OpenStack compute service (Nova) included its own basic networking implementation. Networks were flat or VLAN-based, managed directly by Nova. This was simple but limited: no overlapping IP spaces, no advanced routing, no multi-tenant network isolation.

**Neutron emergence (2012–2015):** Originally called "Quantum," Neutron was introduced as a standalone networking service to replace Nova-network. Neutron introduced the concept of virtual networks backed by pluggable backends (ML2 plugin architecture). This enabled overlay networks (VXLAN, GRE), distributed virtual routers (DVR), and rich security group models.

**ML2 and mechanism drivers (2013–present):** The Modular Layer 2 (ML2) plugin became Neutron's standard plugin, delegating to mechanism drivers for specific network implementations: Open vSwitch (OVS), Linux Bridge, SR-IOV, and vendor-specific backends (Cisco ACI, VMware NSX, Juniper Contrail). This pluggable architecture means the same Neutron API works across radically different network fabrics.

**OVN era (2019–present):** Open Virtual Network (OVN) is now the recommended Neutron backend for new deployments. OVN provides native distributed routing, DHCP, and security groups without the agent-per-node architecture of earlier OVS deployments. This improves scale and reduces operational complexity.

The implication for OpenMCF is that the `openstack_networking_network_v2` Terraform resource and `openstack.networking.Network` Pulumi resource abstract over all of these backends. A network created through our component works identically regardless of whether the underlying cloud uses OVS, OVN, Linux Bridge, or a vendor plugin.

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

OpenStack's web UI (Horizon) provides a "Networks" panel under the "Network" tab:

1. Navigate to **Project > Network > Networks**
2. Click **Create Network**
3. Fill in name, admin state, shared/external flags
4. Optionally add a subnet in the same wizard
5. Click **Create**

**Pros:**
- Visual, intuitive for small-scale operations
- Immediate feedback on network topology

**Cons:**
- Not reproducible or auditable
- Manual errors are common (typos in CIDR, wrong MTU)
- Cannot be version-controlled
- Does not scale beyond a handful of networks

**Verdict:** Useful for learning and one-off exploration. Not suitable for production or repeated deployments.

### Level 1: CLI (openstack client)

The `openstack` unified CLI provides network management:

```bash
# Create a basic network
openstack network create dev-network

# Create with options
openstack network create prod-network \
  --description "Production application network" \
  --mtu 1450 \
  --dns-domain "prod.internal." \
  --tag "env:production" \
  --tag "team:platform"

# Create external network (admin)
openstack network create external-net \
  --external \
  --share

# List networks
openstack network list

# Show details
openstack network show dev-network
```

**Pros:**
- Scriptable, can be put in shell scripts
- Full access to all network attributes
- Good for quick operations and debugging

**Cons:**
- Shell scripts are fragile and hard to maintain
- No state tracking (can't detect drift)
- No dependency management (must manually order operations)
- Error handling is ad-hoc

**Verdict:** Good for ad-hoc operations and debugging. Not recommended for managing infrastructure at scale.

### Level 2: IaC — Terraform

The `terraform-provider-openstack` provides comprehensive Neutron network management:

```hcl
terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 3.0"
    }
  }
}

resource "openstack_networking_network_v2" "app_network" {
  name                  = "app-network"
  description           = "Application network"
  admin_state_up        = true
  mtu                   = 1450
  dns_domain            = "app.internal."
  port_security_enabled = true

  tags = ["env:production", "managed-by:terraform"]
}

output "network_id" {
  value = openstack_networking_network_v2.app_network.id
}
```

**Pros:**
- Declarative infrastructure-as-code
- State tracking and drift detection
- Dependency graph for ordered operations
- Plan/apply workflow for safe changes
- Mature ecosystem and community

**Cons:**
- HCL syntax has a learning curve
- State management requires backend configuration (S3, Consul, etc.)
- No built-in multi-environment management
- Sensitive values in state files need careful handling
- No higher-level abstractions for common patterns

**Verdict:** Industry standard for infrastructure automation. Good for teams comfortable with HCL. Lacks the unified API and multi-cloud abstraction that OpenMCF provides.

### Level 3: IaC — Pulumi

The `pulumi-openstack` SDK provides the same capability in general-purpose languages:

```go
package main

import (
    "github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/networking"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        network, err := networking.NewNetwork(ctx, "app-network", &networking.NetworkArgs{
            Name:                pulumi.String("app-network"),
            Description:         pulumi.StringPtr("Application network"),
            AdminStateUp:        pulumi.BoolPtr(true),
            Mtu:                 pulumi.IntPtr(1450),
            DnsDomain:           pulumi.StringPtr("app.internal."),
            PortSecurityEnabled: pulumi.BoolPtr(true),
            Tags:                pulumi.StringArray{pulumi.String("env:production")},
        })
        if err != nil {
            return err
        }

        ctx.Export("networkId", network.ID())
        return nil
    })
}
```

**Pros:**
- Real programming language (Go, Python, TypeScript, etc.)
- Strong typing catches errors at compile time
- Built-in secret management
- Rich testing capabilities
- Reusable components via packages

**Cons:**
- Requires programming knowledge
- More verbose than HCL for simple resources
- Smaller community than Terraform for OpenStack
- State management still needed

**Verdict:** Excellent for teams with software engineering backgrounds. OpenMCF uses Pulumi internally for its Go-native IaC modules.

### Other Methods

**Ansible (`openstack.cloud` collection):** Procedural automation with `os_network` module. Good for configuration management workflows but lacks state tracking and dependency graphs.

**Heat (OpenStack Orchestration):** OpenStack's native orchestration service. Templates define stacks of resources. Tightly coupled to OpenStack (no multi-cloud). Heat is considered legacy; many deployments don't enable it.

**Crossplane:** Kubernetes-native infrastructure management. Has an OpenStack provider but it is community-maintained with limited coverage.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Dependency Graph | Multi-Cloud | Learning Curve |
|--------|---------------|-----------------|------------------|-------------|----------------|
| Horizon | No | No | No | No | Low |
| CLI | No | No | No | No | Low |
| Terraform | Yes | Yes | Yes | Yes | Medium |
| Pulumi | Yes | Yes | Yes | Yes | Medium-High |
| Ansible | No | Limited | Manual | Yes | Medium |
| Heat | Yes | Limited | Yes | No | Medium |
| **OpenMCF** | **Yes** | **Yes** | **Yes** | **Yes** | **Low** |

## The OpenMCF Approach

OpenMCF wraps both Terraform and Pulumi behind a unified, Kubernetes-style API:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: my-network
spec:
  mtu: 1450
  dns_domain: "my-domain.internal."
```

### What OpenMCF Automates

1. **Provider configuration:** Credentials are resolved from the platform's credential store, not hardcoded in files
2. **State management:** Backend configuration (local, S3, etc.) is handled by the CLI
3. **Naming conventions:** Network name is derived from `metadata.name`, labels are auto-applied
4. **Dual IaC engines:** The same manifest works with both Pulumi and Terraform backends
5. **DAG orchestration:** When used in InfraCharts, the `network_id` output feeds into downstream components (Subnet, Router, Port) via `StringValueOrRef` foreign keys

### The 80/20 Principle

The Terraform `openstack_networking_network_v2` resource exposes 19 arguments. OpenMCF's `OpenStackNetworkSpec` exposes 9 fields — the ones that cover 95%+ of real-world use cases:

**Included:**
- `description` — Human-readable description
- `admin_state_up` — Administrative state (default: true)
- `shared` — Tenant-sharing (admin-only)
- `external` — External/provider network (admin-only)
- `mtu` — Maximum Transmission Unit
- `dns_domain` — DNS integration
- `port_security_enabled` — Port security enforcement
- `tags` — Resource tagging
- `region` — Region override

**Excluded:**
- `tenant_id` — Admin-only; the authenticated tenant is used by default
- `segments` — Multi-segment provider networks; extremely niche
- `value_specs` — Vendor-specific escape hatch; adds complexity without clear benefit
- `availability_zone_hints` — Relevant only for deployments with AZ-aware networking
- `transparent_vlan` — VLAN transparency; niche feature for specific network designs
- `qos_policy_id` — QoS policies are an advanced feature; no QoS component in scope

### API Design Decisions

**`optional bool admin_state_up` with default `true`:** Proto3 `bool` defaults to `false`. Without `optional` + explicit default, users who omit `admin_state_up` would accidentally create admin-down networks — a dangerous foot-gun. The `optional` keyword enables presence tracking, and the OpenMCF middleware applies the default before the IaC module runs.

**`optional bool port_security_enabled` without default:** The correct default depends on the OpenStack deployment's ML2 plugin configuration. By using `optional` without a default, we let OpenStack decide when the user doesn't specify — respecting the deployment's own policy.

**`shared` and `external` as plain `bool`:** Their proto3 default (`false`) matches the correct default for tenant users. No need for `optional` + default.

**Network name from `metadata.name`:** The OpenMCF resource name IS the OpenStack network name. This is consistent across all OpenStack components and mirrors how users think about resources.

## Implementation Landscape

### Resources Created

| Resource | Count | Description |
|----------|-------|-------------|
| `openstack_networking_network_v2` (TF) / `networking.Network` (Pulumi) | 1 | The Neutron network |

This is a single-resource component. Its simplicity is by design — networks are building blocks, and keeping them atomic enables maximum composability in InfraCharts.

### Dependency Role

OpenStackNetwork is a **root resource** — it has no foreign key dependencies. However, it is referenced by nearly every other OpenStack component:

- `OpenStackSubnet.spec.network_id` → `OpenStackNetwork.status.outputs.network_id`
- `OpenStackRouter.spec.external_network_id` → `OpenStackNetwork.status.outputs.network_id`
- `OpenStackNetworkPort.spec.network_id` → `OpenStackNetwork.status.outputs.network_id`
- `OpenStackFloatingIp.spec.floating_network_id` → `OpenStackNetwork.status.outputs.network_id`
- `OpenStackInstance.spec.networks[].uuid` → `OpenStackNetwork.status.outputs.network_id`

This makes `network_id` the single most important output in the entire OpenStack component family.

## Production Best Practices

### Network Segmentation

- Create separate networks for different tiers (frontend, backend, data) rather than one flat network
- Use security groups and port-level rules for micro-segmentation within networks
- External networks should be created by admins and referenced by tenants — tenants should not create external networks

### MTU Configuration

- For VXLAN overlay networks: MTU 1450 (50 bytes overhead from VXLAN encapsulation)
- For GRE overlay networks: MTU 1458 (42 bytes overhead)
- For flat/VLAN networks: MTU 1500 (standard Ethernet)
- For jumbo frame deployments: MTU 9000 (verify switch infrastructure supports it)
- When in doubt, omit MTU and let OpenStack auto-configure based on the network type

### Port Security

- Keep port security enabled (the default) unless you have a specific reason to disable it
- Disabling port security is needed for: network monitoring/sniffing, DHCP servers, load balancers that use virtual IPs
- When disabled, all traffic is allowed regardless of security group rules

### DNS Integration

- Enable DNS domains when you want automatic DNS name assignment for ports
- Requires the `dns-integration` Neutron extension (check with `openstack extension list`)
- DNS domain must end with a dot (`.`) — it's a Fully Qualified Domain Name

### Tags

- Use tags for organizational purposes: team ownership, environment, cost allocation
- Tags are stored on the OpenStack resource and queryable via the API
- Use `metadata.labels` for OpenMCF-level metadata; use `spec.tags` for OpenStack-level tags

## References

- [Terraform openstack_networking_network_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_network_v2)
- [Pulumi openstack.networking.Network](https://www.pulumi.com/registry/packages/openstack/api-docs/networking/network/)
- [OpenStack Neutron API — Networks](https://docs.openstack.org/api-ref/network/v2/#networks)
- [OpenStack Networking Guide](https://docs.openstack.org/neutron/latest/admin/)
- [OVN Architecture](https://docs.openstack.org/neutron/latest/admin/ovn/index.html)

---
title: "Floating IP"
description: "Floating IP deployment documentation"
icon: "package"
order: 100
componentName: "openstackfloatingip"
---

# OpenStack Neutron Floating IP -- Research Documentation

## Introduction

OpenStack Neutron floating IPs provide the mechanism for external (public) connectivity to tenant resources. A floating IP is allocated from an external provider network and mapped via 1:1 NAT (DNAT/SNAT) to a fixed IP address on a tenant port. This enables instances on private tenant networks to be reachable from outside the OpenStack cloud.

Floating IPs are the primary mechanism for exposing services to the internet in OpenStack deployments, analogous to AWS Elastic IPs or GCP external IP addresses.

## Historical Context

**Early Nova floating IPs (2011-2013):** Floating IPs were originally managed through the Nova API (`os-floating-ips`). This legacy API is deprecated and should not be used. OpenMCF exclusively uses the Neutron API.

**Neutron floating IPs (2013-present):** The Neutron `floatingips` API (`/v2.0/floatingips`) provides the modern, stable implementation. The core API has remained stable since Icehouse (2014), with additive features like DNS integration, tagging, and subnet selection added in later releases.

**DNS integration (Queens, 2018):** The `dns_name` and `dns_domain` fields enable automatic DNS record creation when floating IPs are allocated. This requires the `dns-integration` extension and is deployment-dependent. Excluded from our 80/20 spec.

## Deployment Methods Landscape

### Level 0: Manual (Horizon Dashboard)

1. Navigate to **Project > Network > Floating IPs**
2. Click **Allocate IP to Project**
3. Select the external network (pool)
4. Optionally select a specific subnet
5. Click **Allocate IP**
6. To associate: click **Associate** next to the floating IP, select a port

**Pros:** Visual, simple for one-off allocations.
**Cons:** Not reproducible, no audit trail, manual association.

### Level 1: CLI (OpenStack CLI)

```bash
# Allocate
openstack floating ip create external-network

# Allocate with specific address
openstack floating ip create --floating-ip-address 203.0.113.42 external-network

# Associate with a port
openstack floating ip set --port <port-id> <floating-ip-id>

# Associate with a server (convenience)
openstack server add floating ip <server-name> <floating-ip-address>
```

**Pros:** Scriptable, faster than Horizon.
**Cons:** Imperative, no state management, drift-prone.

### Level 2: Terraform

```hcl
resource "openstack_networking_floatingip_v2" "web" {
  pool = "external-network"
}

resource "openstack_networking_floatingip_associate_v2" "web" {
  floating_ip = openstack_networking_floatingip_v2.web.address
  port_id     = openstack_networking_port_v2.web.id
}
```

**Pros:** Declarative, state-tracked, supports both allocation and association.
**Cons:** HCL syntax, no built-in FK resolution, manual dependency management.

### Level 3: OpenMCF (This Component)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
spec:
  floating_network_id:
    value_from:
      name: external-network
  port_id:
    value_from:
      name: web-server-port
```

**Pros:** Declarative YAML, FK resolution via `value_from`, InfraChart DAG integration, dual IaC (Pulumi + Terraform).
**Cons:** Requires OpenMCF platform.

## Terraform Provider Analysis

### Resource: `openstack_networking_floatingip_v2`

14 schema fields. Our 80/20 spec includes 8.

| TF Field | Our Field | Included | Rationale |
|----------|-----------|----------|-----------|
| `pool` | `floating_network_id` | Yes | Renamed to match Neutron API naming. Required. |
| `port_id` | `port_id` | Yes | Built-in association. Optional StringValueOrRef FK. |
| `fixed_ip` | `fixed_ip` | Yes | Multi-IP port disambiguation. |
| `subnet_id` | `subnet_id` | Yes | Allocation from specific external subnet. Plain string. |
| `address` | `address` | Yes | Specific IP request. ForceNew. |
| `description` | `description` | Yes | Human-readable. |
| `tags` | `tags` | Yes | Filtering and organization. |
| `region` | `region` | Yes | Region override. |
| `subnet_ids` | -- | No | Retry-loop for multi-pool. Niche operational feature. |
| `tenant_id` | -- | No | Admin-only. Excluded across all components. |
| `value_specs` | -- | No | Escape hatch. Not needed in 80/20. |
| `dns_name` | -- | No | Requires `dns-integration` extension. Deployment-dependent. |
| `dns_domain` | -- | No | Requires `dns-integration` extension. Deployment-dependent. |
| `all_tags` | -- | No | Computed-only. Not a spec field. |

### Key Design Decisions

**`pool` renamed to `floating_network_id`:** The Terraform field `pool` accepts a network name or UUID. Our field is a `StringValueOrRef` FK that resolves to a network UUID. The name `floating_network_id` is more descriptive and follows OpenMCF's FK naming conventions. The IaC modules map this to the TF `pool` attribute.

**`subnet_id` as plain string (not FK):** The subnet referenced is on the external (admin-managed) network, not a user-created subnet. Same reasoning as `OpenStackRouter.ExternalFixedIp.subnet_id`. Admin-managed resources should not be FK-wired.

**`address` field added (not in original plan):** Allows requesting a specific IP. Common real-world use case for DNS pre-configuration, firewall whitelisting, and IP reservation. ForceNew semantics mean changing it recreates the resource.

**Single-resource IaC (correction from plan):** The original plan suggested creating `floatingip_v2 + optional floatingip_associate_v2`. This was corrected to a single `floatingip_v2` resource. The TF resource natively supports `port_id` for built-in association. The separate `OpenStackFloatingIpAssociate` component handles DAG-visible association.

## Production Considerations

- **IP exhaustion:** External network pools are finite. Monitor allocation to avoid exhaustion.
- **Billing:** Some OpenStack deployments charge for allocated floating IPs, even when unassociated. Clean up unused allocations.
- **Security groups:** A floating IP does not bypass security groups. The associated port's security groups still apply.
- **HA:** Floating IPs follow the L3 agent. In DVR mode, floating IPs are distributed across compute nodes. In non-DVR mode, they are centralized on the network node.
- **ForceNew fields:** `address`, `floating_network_id`, and `region` are ForceNew in Terraform. Changing them destroys and recreates the resource (new IP allocated).

## References

- [Neutron Floating IP API](https://docs.openstack.org/api-ref/network/v2/#floating-ips-floatingips)
- [Terraform openstack_networking_floatingip_v2](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/networking_floatingip_v2)
- [Pulumi openstack.networking.FloatingIp](https://www.pulumi.com/registry/packages/openstack/api-docs/networking/floatingip/)

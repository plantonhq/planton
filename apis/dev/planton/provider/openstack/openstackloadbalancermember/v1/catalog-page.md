# OpenStack Load Balancer Member

Deploys an Octavia pool member in OpenStack, representing a backend server that receives traffic from a load balancer pool. Each member defines an IP address, port, and optional weight for weighted load distribution, with support for cross-subnet routing when the backend resides on a different subnet than the VIP.

## What Gets Created

When you deploy an OpenStackLoadBalancerMember resource, Planton provisions:

- **Octavia Pool Member** — a `loadbalancer.Member` (Pulumi) / `openstack_lb_member_v2` (Terraform) resource that registers a backend server in the specified pool. The member is identified by its address and port, and participates in the pool's load-balancing algorithm. When `weight` is set, the member receives a proportional share of traffic. When `subnetId` is set, Octavia performs L3 routing to reach the backend on a different subnet.

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **An Octavia pool** to add the member to (provide the pool UUID or reference an OpenStackLoadBalancerPool resource)
- **A backend server** with a reachable IP address and listening port
- **A subnet** (optional) if the backend server is on a different subnet than the load balancer VIP

## Quick Start

Create a file `member.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: dev.OpenstackLoadBalancerMember.web-backend-1
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  address: "10.0.0.10"
  protocolPort: 8080
```

Deploy:

```shell
planton apply -f member.yaml
```

This registers a backend server at `10.0.0.10:8080` in the specified Octavia pool with default weight (1).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `poolId` | `StringValueOrRef` | The Octavia pool to add this member to. ForceNew: changing this recreates the member. Can reference OpenStackLoadBalancerPool resource via `valueFrom`. | Required |
| `address` | `string` | IP address of the backend server that receives forwarded traffic. ForceNew: changing this recreates the member. | Required, non-empty |
| `protocolPort` | `int32` | Port on the backend server that accepts traffic. ForceNew: changing this recreates the member. | Required, 1-65535 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subnetId` | `StringValueOrRef` | — | Subnet where the member resides. Used by Octavia for L3 routing when the member is on a different subnet than the VIP. ForceNew: changing this recreates the member. Can reference OpenStackSubnet resource via `valueFrom`. |
| `weight` | `int32` | `1` (set by Octavia) | Weight for weighted load balancing. A member with weight 0 receives no traffic (drain mode). Valid range: 0-256. When omitted, Octavia assigns a default weight of 1. |
| `adminStateUp` | `bool` | `true` | Administrative state of the member. When `false`, the member is removed from the pool's rotation without deleting it. |
| `tags` | `string[]` | `[]` | Tags applied to the member in OpenStack for filtering and organization. Must be unique within this resource. |
| `region` | `string` | provider default | Overrides the region from the provider configuration for this member. |

## Examples

### Basic Member with Direct IP

A minimal member that registers a single backend server in an Octavia pool:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: dev.OpenstackLoadBalancerMember.web-backend-1
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  address: "10.0.0.10"
  protocolPort: 8080
```

### Weighted Members with Cross-Subnet Routing

Two members with different weights on a backend subnet separate from the VIP subnet. The higher-weighted member receives proportionally more traffic:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: app-backend-primary
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: prod.OpenstackLoadBalancerMember.app-backend-primary
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  address: "10.1.0.10"
  protocolPort: 8080
  weight: 10
  subnetId: b2c3d4e5-f6a7-8901-bcde-f12345678901
  tags:
    - production
    - primary
---
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: app-backend-secondary
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: prod.OpenstackLoadBalancerMember.app-backend-secondary
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  address: "10.1.0.11"
  protocolPort: 8080
  weight: 5
  subnetId: b2c3d4e5-f6a7-8901-bcde-f12345678901
  tags:
    - production
    - secondary
```

With these weights, `app-backend-primary` receives approximately twice the traffic of `app-backend-secondary` when the pool uses the ROUND_ROBIN algorithm.

### Draining a Member for Maintenance

Set `weight` to 0 to stop sending new traffic to a member while allowing existing connections to complete. Set `adminStateUp` to `false` to fully remove the member from rotation:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: maintenance-backend
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: prod.OpenstackLoadBalancerMember.maintenance-backend
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  address: "10.0.0.12"
  protocolPort: 8080
  weight: 0
  adminStateUp: false
  tags:
    - maintenance
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding UUIDs:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: ref-backend
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    planton.dev/stack.jobId: prod.OpenstackLoadBalancerMember.ref-backend
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackloadbalancermember/v1/iac/pulumi/module
spec:
  poolId:
    valueFrom:
      kind: OpenStackLoadBalancerPool
      name: web-pool
      field: status.outputs.pool_id
  address: "10.0.0.10"
  protocolPort: 8080
  weight: 10
  subnetId:
    valueFrom:
      kind: OpenStackSubnet
      name: backend-subnet
      field: status.outputs.subnet_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `member_id` | `string` | UUID of the created Octavia pool member |
| `name` | `string` | Name of the member, derived from `metadata.name` |
| `address` | `string` | IP address of the backend server |
| `protocol_port` | `int32` | Port on the backend server |
| `weight` | `int32` | Weight of the member in the load-balancing algorithm |
| `region` | `string` | OpenStack region where the member was created |

## Related Components

- [OpenStackLoadBalancerPool](/docs/catalog/openstack/openstackloadbalancerpool) — the pool this member belongs to; provides the `poolId` foreign key
- [OpenStackSubnet](/docs/catalog/openstack/openstacksubnet) — optional subnet for cross-subnet routing; provides the `subnetId` foreign key
- [OpenStackLoadBalancerListener](/docs/catalog/openstack/openstackloadbalancerlistener) — listener that routes traffic to the pool containing this member
- [OpenStackLoadBalancer](/docs/catalog/openstack/openstackloadbalancer) — top-level load balancer resource that owns the VIP
- [OpenStackLoadBalancerMonitor](/docs/catalog/openstack/openstackloadbalancermonitor) — health monitor attached to the pool that determines member health

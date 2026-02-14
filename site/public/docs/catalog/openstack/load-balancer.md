---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "openstackloadbalancer"
---

# OpenStack Load Balancer

Deploys an Octavia load balancer in OpenStack, provisioning a Virtual IP (VIP) endpoint on a specified subnet. Listeners, pools, members, and health monitors attach to the load balancer to distribute traffic across backend servers.

## What Gets Created

When you deploy an OpenStackLoadBalancer resource, OpenMCF provisions:

- **Octavia Load Balancer** -- a `loadbalancer.LoadBalancer` Pulumi resource (equivalent to `openstack_lb_loadbalancer_v2` in Terraform) with the configured VIP subnet, optional fixed VIP address, description, administrative state, flavor, tags, and region override. The load balancer allocates a VIP address on the target subnet and exposes the VIP port ID for floating IP or security group attachment.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing subnet** where the VIP address will be allocated (can be an OpenMCF-managed OpenStackSubnet)
- **Octavia service** enabled in the target OpenStack project
- **An Octavia flavor** (optional) if you need specific resource limits for the load balancer

## Quick Start

Create a file `loadbalancer.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: my-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackLoadBalancer.my-lb
spec:
  vipSubnetId: e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d
```

Deploy:

```shell
openmcf apply -f loadbalancer.yaml
```

This creates an Octavia load balancer with a VIP auto-allocated from the specified subnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `vipSubnetId` | `StringValueOrRef` | The subnet on which to allocate the VIP address. Determines the network segment where the load balancer's virtual IP lives. The subnet must already exist and have available IP addresses. Can reference an OpenStackSubnet resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vipAddress` | `string` | auto-allocated | A specific IP address to request for the VIP. Must be within the CIDR range of the specified subnet. ForceNew: changing this requires recreating the load balancer. |
| `description` | `string` | `""` | Human-readable description of the load balancer. |
| `adminStateUp` | `bool` | `true` | Administrative state of the load balancer. When false, the load balancer stops accepting traffic. |
| `flavorId` | `string` | -- | The ID of an Octavia flavor to use. Flavors define resource limits such as bandwidth and connections. ForceNew: changing this requires recreating the load balancer. |
| `tags` | `string[]` | `[]` | Tags applied to the load balancer in OpenStack. Must be unique within this resource. |
| `region` | `string` | provider default | Overrides the region from the provider configuration for this load balancer. |

## Examples

### Basic Load Balancer

A minimal load balancer with a VIP auto-allocated from a subnet:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: web-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackLoadBalancer.web-lb
spec:
  vipSubnetId: e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d
  description: Web tier load balancer
```

### Load Balancer with Fixed VIP Address

Pin the VIP to a known IP address on the subnet, useful when DNS records or firewall rules reference a stable address:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: api-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancer.api-lb
spec:
  vipSubnetId: e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d
  vipAddress: "10.0.1.100"
  description: API gateway load balancer with fixed VIP
  tags:
    - production
    - api-tier
```

### Load Balancer with Flavor and Tags

Use an Octavia flavor to control the load balancer's resource allocation (amphora size, topology, provider):

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: premium-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancer.premium-lb
spec:
  vipSubnetId: e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d
  vipAddress: "10.0.1.200"
  description: High-capacity production load balancer
  flavorId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
  adminStateUp: true
  tags:
    - production
    - high-capacity
  region: RegionOne
```

### Using Foreign Key References

Reference an OpenMCF-managed subnet instead of hardcoding its UUID:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancer
metadata:
  name: ref-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackLoadBalancer.ref-lb
spec:
  vipSubnetId:
    valueFrom:
      kind: OpenStackSubnet
      name: app-subnet
      field: status.outputs.subnet_id
  description: Load balancer referencing a managed subnet
  tags:
    - managed
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `loadbalancer_id` | `string` | UUID of the created Octavia load balancer. Primary output used as a foreign key by listeners. |
| `name` | `string` | Name of the load balancer, derived from `metadata.name` |
| `vip_address` | `string` | Virtual IP address allocated to the load balancer. This is the IP that clients connect to. |
| `vip_port_id` | `string` | Neutron port ID of the VIP. Useful for attaching security groups or floating IPs. |
| `vip_subnet_id` | `string` | Subnet where the VIP was allocated |
| `region` | `string` | OpenStack region where the load balancer was created |

## Related Components

- [OpenStackSubnet](/docs/catalog/openstack/openstacksubnet) -- provides the subnet for VIP allocation
- [OpenStackLoadBalancerListener](/docs/catalog/openstack/openstackloadbalancerlistener) -- attaches a protocol/port listener to the load balancer
- [OpenStackLoadBalancerPool](/docs/catalog/openstack/openstackloadbalancerpool) -- defines a backend pool with load balancing algorithm
- [OpenStackLoadBalancerMember](/docs/catalog/openstack/openstackloadbalancermember) -- registers backend servers in a pool
- [OpenStackLoadBalancerMonitor](/docs/catalog/openstack/openstackloadbalancermonitor) -- configures health checks for pool members
- [OpenStackFloatingIpAssociate](/docs/catalog/openstack/openstackfloatingipassociate) -- associates a floating IP with the VIP port for external access

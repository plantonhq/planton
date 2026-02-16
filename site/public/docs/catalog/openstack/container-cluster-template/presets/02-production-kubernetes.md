---
title: "Production Kubernetes Template"
description: "This preset creates a Magnum cluster template for production Kubernetes deployments with explicit network configuration, master load balancing, and floating IPs. Clusters created from this template..."
type: "preset"
rank: "02"
presetSlug: "02-production-kubernetes"
componentSlug: "container-cluster-template"
componentTitle: "Container Cluster Template"
provider: "openstack"
icon: "package"
order: 2
---

# Production Kubernetes Template

This preset creates a Magnum cluster template for production Kubernetes deployments with explicit network configuration, master load balancing, and floating IPs. Clusters created from this template are placed on an existing network/subnet and get a load balancer in front of the master API for HA.

## When to Use

- Production Kubernetes clusters that need HA master access via a load balancer
- Environments where clusters must use an existing network topology (not Magnum-created networks)
- Clusters that need floating IPs for external access to master and worker nodes

## Key Configuration Choices

- **Kubernetes COE** (`coe: kubernetes`)
- **Master LB** (`masterLbEnabled: true`) -- Octavia load balancer in front of the Kubernetes API; required for multi-master HA
- **Floating IPs** (`floatingIpEnabled: true`) -- master and worker nodes get floating IPs for external access
- **Existing network** (`fixedNetwork`, `fixedSubnet`) -- cluster nodes are placed on a pre-existing network instead of Magnum creating one
- **External network** (`externalNetwork`) -- used for floating IP allocation and router gateway
- **50 GB Docker volume** (`dockerVolumeSize: 50`) -- dedicated Cinder volume for container images and layers on each node
- **Flannel networking** -- simple overlay; swap to `calico` for network policy support

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<fedora-coreos-image-id>` | UUID of a Fedora CoreOS image for Kubernetes nodes | `openstack image list` or `OpenStackImage` status outputs |
| `<external-network-id>` | ID of the external (provider) network for floating IPs | OpenStack admin or `OpenStackNetwork` (external) status outputs |
| `<fixed-network-id>` | ID of the existing tenant network for cluster nodes | OpenStack console or `OpenStackNetwork` status outputs |
| `<fixed-subnet-id>` | ID of the existing subnet for cluster nodes | OpenStack console or `OpenStackSubnet` status outputs |

## Related Presets

- **01-standard-kubernetes** -- Use instead for simpler deployments where Magnum manages networking

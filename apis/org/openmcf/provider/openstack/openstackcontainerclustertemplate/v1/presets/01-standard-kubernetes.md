# Standard Kubernetes Template

This preset creates a minimal Magnum cluster template for Kubernetes with Flannel networking and Google DNS. The template defines the base configuration shared by all clusters created from it -- COE type, image, and networking. Clusters reference this template and add their own master/worker counts and flavors.

## When to Use

- Development and small production Kubernetes clusters
- Environments where Flannel's simplicity is preferred over Calico's network policy support
- Getting started with Magnum when no cluster template exists yet

## Key Configuration Choices

- **Kubernetes COE** (`coe: kubernetes`) -- container orchestration engine
- **Flannel networking** (`networkDriver: flannel`) -- simple overlay network; no network policy support
- **Google DNS** (`dnsNameserver: 8.8.8.8`) -- reliable public DNS for cluster nodes
- **No fixed network** -- Magnum creates a new network/subnet per cluster; add `fixedNetwork`/`fixedSubnet` to use an existing network

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<fedora-coreos-image-id>` | UUID of a Fedora CoreOS (or similar) image for Kubernetes nodes | `openstack image list` or `OpenStackImage` status outputs |

## Related Presets

- **02-production-kubernetes** -- Use instead when you need explicit network configuration, master LB, and production-grade settings

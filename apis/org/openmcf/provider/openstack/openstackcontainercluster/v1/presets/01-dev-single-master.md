# Dev Cluster (Single Master)

This preset creates a minimal Magnum container cluster with a single master and one worker node. The cluster configuration (COE, networking, flavors) is defined by the referenced cluster template. This is the cheapest configuration suitable for development and testing.

## When to Use

- Development and testing environments
- Learning and experimentation with Kubernetes on OpenStack
- Workloads that do not require high availability

## Key Configuration Choices

- **Single master** -- default (1 master when `masterCount` is unset); no HA for the control plane
- **1 worker** (`nodeCount: 1`) -- minimal compute; increase for more capacity
- **Template-driven** -- all cluster infrastructure settings (image, networking, flavors) come from the referenced template

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<cluster-template-id>` | ID of the Magnum cluster template | OpenStack console or `OpenStackContainerClusterTemplate` status outputs |

## Related Presets

- **02-ha-multi-master** -- Use instead for production workloads requiring HA control plane

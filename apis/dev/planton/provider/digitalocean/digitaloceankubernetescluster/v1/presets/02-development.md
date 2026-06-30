# Development Kubernetes Cluster

This preset creates a minimal DigitalOcean Kubernetes cluster for development and testing. It uses a non-HA control plane, a fixed-size node pool with smaller instances, and no API server firewall -- keeping costs low while still providing VPC isolation.

## When to Use

- Development, staging, or CI/CD environments
- Learning and experimentation with Kubernetes
- Short-lived clusters for feature branch testing

## Key Configuration Choices

- **Non-HA control plane** -- `highlyAvailable` omitted (defaults to `false`). Sufficient for non-critical workloads and saves cost.
- **Fixed-size node pool** -- 2 nodes with no autoscaling. Predictable cost and sufficient for dev workloads.
- **Smaller instances** (`size: s-2vcpu-4gb`) -- half the resources of the production preset.
- **No API firewall** -- `controlPlaneFirewallAllowedIps` omitted for developer convenience. Add CIDRs if needed.
- **No auto-upgrade or maintenance window** -- manual control over upgrades in dev environments.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-name>` | Name of the target DigitalOcean VPC | `DigitalOceanVpc` resource `metadata.name` |
| `nyc1` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **01-production-ha** -- Use instead for production workloads requiring HA, autoscaling, and API server security

---
title: "Production HA Kubernetes Cluster"
description: "This preset creates a production-grade DigitalOcean Kubernetes (DOKS) cluster with a highly available control plane, autoscaling default node pool, automatic patch upgrades, a scheduled maintenance..."
type: "preset"
rank: "01"
presetSlug: "01-production-ha"
componentSlug: "kubernetes-cluster"
componentTitle: "Kubernetes Cluster"
provider: "digitalocean"
icon: "package"
order: 1
---

# Production HA Kubernetes Cluster

This preset creates a production-grade DigitalOcean Kubernetes (DOKS) cluster with a highly available control plane, autoscaling default node pool, automatic patch upgrades, a scheduled maintenance window, and API server firewall restrictions. This is the recommended starting point for any production Kubernetes workload on DigitalOcean.

## When to Use

- Production applications requiring high availability and automatic recovery
- Workloads that need autoscaling based on demand
- Teams using DigitalOcean Container Registry for private images
- Environments requiring restricted API server access

## Key Configuration Choices

- **HA control plane** (`highlyAvailable: true`) -- multiple control plane replicas for resilience. Eliminates single-point-of-failure for the Kubernetes API.
- **Auto-upgrade** (`autoUpgrade: true`) -- automatically applies patch releases during the maintenance window.
- **Maintenance window** (`maintenanceWindow: "sunday=03:00"`) -- upgrades occur during low-traffic hours. Adjust to your timezone.
- **Registry integration** (`registryIntegration: true`) -- automatically provisions `imagePullSecret` in all namespaces for pulling from DOCR.
- **API server firewall** (`controlPlaneFirewallAllowedIps`) -- restricts `kubectl` and API access to specified CIDRs. Critical for production security.
- **Autoscaling node pool** (`autoScale: true`, 2-5 nodes) -- scales based on pod scheduling pressure. The default pool uses `s-4vcpu-8gb` (general-purpose).
- **VPC reference uses `metadata.name`** -- the system resolves the VPC name to its UUID.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-name>` | Name of the target DigitalOcean VPC | `DigitalOceanVpc` resource `metadata.name` |
| `<your-management-cidr>` | CIDR block for API server access (e.g., `203.0.113.0/24`) | Your network admin or VPN provider |
| `nyc1` | Target DigitalOcean region slug | Must match the VPC's region |
| `1.31` | Kubernetes version | [DOKS supported versions](https://docs.digitalocean.com/products/kubernetes/details/supported-release-policy/) |

## Related Presets

- **02-development** -- Use instead for dev/test clusters where HA and autoscaling are unnecessary

---
title: "Spot VM for Development"
description: "This preset creates a cost-optimized Spot VM with SSH access for development and testing. Spot VMs cost 60-91% less than on-demand but can be preempted. The instance is configured to stop (not..."
type: "preset"
rank: "02"
presetSlug: "02-spot-development"
componentSlug: "compute-instance"
componentTitle: "Compute Instance"
provider: "gcp"
icon: "package"
order: 2
---

# Spot VM for Development

This preset creates a cost-optimized Spot VM with SSH access for development and testing. Spot VMs cost 60-91% less than on-demand but can be preempted. The instance is configured to stop (not delete) on preemption, preserving boot disk data.

## When to Use

- Development and testing environments where cost matters more than uptime
- Temporary workstations for debugging, experimentation, or CI tasks
- Any workload that can tolerate occasional interruption

## Key Configuration Choices

- **Spot VM** (`spot: true`) -- significant cost savings; GCP can reclaim the instance at any time
- **Stop on preemption** (`instanceTerminationAction: STOP`) -- preserves boot disk data; restart manually after preemption
- **External IP** (`accessConfigs` with PREMIUM tier) -- enables direct SSH access
- **e2-medium** -- 2 vCPU, 4 GB RAM; smaller and cheaper than production
- **Standard disk** (`type: pd-standard`, 20 GB) -- lower cost for development
- **SSH keys** -- pre-configured for developer access
- **No deletion protection** -- dev VMs should be easy to tear down

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-zone>` | GCP zone (e.g., `us-central1-a`) | Your deployment zone |
| `<subnet-self-link>` | Self-link of the subnet | `GcpSubnetwork` status outputs |
| `<username>` | SSH username | Your username |
| `<ssh-public-key>` | SSH public key (content of `~/.ssh/id_rsa.pub`) | Your SSH key |

## Related Presets

- **01-standard-production** -- Use for production VMs with SSD, deletion protection, and no external IP

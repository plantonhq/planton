---
title: "Standard Production VM"
description: "This preset creates a production Compute Engine instance with an SSD boot disk, a dedicated service account, deletion protection, and no external IP (private-only networking). It follows GCP security..."
type: "preset"
rank: "01"
presetSlug: "01-standard-production"
componentSlug: "compute-instance"
componentTitle: "Compute Instance"
provider: "gcp"
icon: "package"
order: 1
---

# Standard Production VM

This preset creates a production Compute Engine instance with an SSD boot disk, a dedicated service account, deletion protection, and no external IP (private-only networking). It follows GCP security best practices for VM workloads.

## When to Use

- Production application servers, worker nodes, or database hosts
- VMs that need reliable I/O performance (SSD boot disk)
- Instances that should be protected from accidental deletion

## Key Configuration Choices

- **e2-standard-2** -- 2 vCPU, 8 GB RAM; cost-effective general-purpose machine type
- **SSD boot disk** (`type: pd-ssd`, 50 GB) -- faster I/O than standard disk
- **Debian 12** -- latest stable Debian; change to `ubuntu-os-cloud/ubuntu-2404-lts` for Ubuntu
- **No external IP** -- no `accessConfigs` defined; internet egress via Cloud NAT
- **Dedicated service account** -- follows least-privilege; avoid using the default Compute Engine SA
- **Deletion protection** (`deletionProtection: true`) -- prevents accidental deletion
- **Network tags** -- for firewall rule targeting

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-zone>` | GCP zone (e.g., `us-central1-a`) | Your deployment zone |
| `<subnet-self-link>` | Self-link of the subnet | `GcpSubnetwork` status outputs |
| `<service-account-email>` | Email of the dedicated service account | `GcpServiceAccount` status outputs |
| `<firewall-network-tag>` | Network tag for firewall rule targeting (e.g., `web-server`) | Your firewall rule configuration |

## Related Presets

- **02-spot-development** -- Use for development VMs with SSH access and lower cost

---
title: "Standard Production Project"
description: "This preset creates a GCP project under a resource hierarchy folder with essential APIs pre-enabled, the default network disabled, and deletion protection turned on. It covers the core services..."
type: "preset"
rank: "01"
presetSlug: "01-standard-production"
componentSlug: "project"
componentTitle: "Project"
provider: "gcp"
icon: "package"
order: 1
---

# Standard Production Project

This preset creates a GCP project under a resource hierarchy folder with essential APIs pre-enabled, the default network disabled, and deletion protection turned on. It covers the core services needed by most production workloads: compute, containers, DNS, secrets, IAM, and observability.

## When to Use

- New production workloads that need a dedicated GCP project
- Projects that will host GKE clusters, Cloud Run services, or Compute Engine VMs
- Environments following the recommended folder-based resource hierarchy

## Key Configuration Choices

- **Folder-based hierarchy** (`parentType: folder`) -- projects belong under folders, not directly under the organization
- **Default network disabled** (`disableDefaultNetwork: true`) -- removes the auto-created VPC with overly permissive firewall rules
- **Deletion protection** (`deleteProtection: true`) -- prevents accidental deletion of a production project
- **9 essential APIs enabled** -- compute, container, DNS, secrets, IAM, logging, monitoring, resource manager, and service networking
- **Service networking API** -- required prerequisite for Private Services Access (Cloud SQL, Memorystore private IPs)
- **Labels** -- `environment` and `managed-by` for cost allocation and governance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-project-id>` | Globally unique GCP project ID (6-30 chars, lowercase) | Choose a unique ID for your project |
| `<your-folder-id>` | Numeric folder ID in the GCP resource hierarchy | GCP Resource Manager console or `gcloud resource-manager folders list` |
| `<AAAAAA-BBBBBB-CCCCCC>` | Billing account ID in `XXXXXX-XXXXXX-XXXXXX` format | GCP Billing console or `gcloud billing accounts list` |
| `<owner-email>` | Email of the user, group, or service account to grant Owner role | Your organization's identity provider |

## Related Presets

- **02-development** -- Use for non-production environments with fewer APIs and no deletion protection

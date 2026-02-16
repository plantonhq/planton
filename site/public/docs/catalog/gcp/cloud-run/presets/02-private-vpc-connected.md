---
title: "Private VPC-Connected Cloud Run Service"
description: "This preset deploys a Cloud Run service that is only accessible internally (within the VPC and other GCP services), requires IAM authentication, and has Direct VPC Egress for connecting to private..."
type: "preset"
rank: "02"
presetSlug: "02-private-vpc-connected"
componentSlug: "cloud-run"
componentTitle: "Cloud Run"
provider: "gcp"
icon: "package"
order: 2
---

# Private VPC-Connected Cloud Run Service

This preset deploys a Cloud Run service that is only accessible internally (within the VPC and other GCP services), requires IAM authentication, and has Direct VPC Egress for connecting to private resources like Cloud SQL or Memorystore. It keeps one instance always warm to avoid cold-start latency for internal traffic.

## When to Use

- Backend microservices that communicate with other services inside the VPC
- Services that need to access Cloud SQL, Memorystore, or other private-IP resources
- Internal APIs that should not be exposed to the public internet

## Key Configuration Choices

- **Internal only** (`ingress: INGRESS_TRAFFIC_INTERNAL_ONLY`) -- reachable only from within VPC and GCP services
- **Authenticated** (`allowUnauthenticated: false`) -- requires IAM `run.invoker` role to invoke
- **VPC access** (`vpcAccess`) -- Direct VPC Egress for private IP connectivity to databases and caches
- **Private ranges only** (`egress: PRIVATE_RANGES_ONLY`) -- only RFC 1918 traffic routes through VPC; public traffic uses default path
- **Always-warm** (`replicas.min: 1`) -- avoids cold starts for internal service-to-service calls
- **2 vCPU / 1 GiB** -- more resources for services doing database queries or heavier processing
- **Deletion protection** (`deleteProtection: true`) -- prevents accidental deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<container-image-repo>` | Container image repository | `GcpArtifactRegistryRepo` outputs |
| `<image-tag>` | Image tag | Your CI/CD pipeline |
| `<vpc-network-name>` | VPC network name (not self-link) | `GcpVpc` status outputs |
| `<subnet-name>` | Subnet name (not self-link) | `GcpSubnetwork` status outputs |

## Related Presets

- **01-public-service** -- Use for public-facing services that don't need VPC access

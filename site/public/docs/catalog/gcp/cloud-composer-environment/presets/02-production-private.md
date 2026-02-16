---
title: "Production Private"
description: "A production-grade Cloud Composer environment with private networking, high resilience, and scaled workloads. Designed for production Airflow workloads requiring network isolation and high..."
type: "preset"
rank: "02"
presetSlug: "02-production-private"
componentSlug: "cloud-composer-environment"
componentTitle: "Cloud Composer Environment"
provider: "gcp"
icon: "package"
order: 2
---

# Production Private

A production-grade Cloud Composer environment with private networking, high resilience, and scaled workloads. Designed for production Airflow workloads requiring network isolation and high availability.

## When to Use

- Production data pipelines requiring network isolation
- Workloads that need high availability and resilience
- Environments where private endpoint access is required
- Multi-zone deployments for disaster recovery
- Compliance requirements for private networking

## Key Configuration Choices

- **ENVIRONMENT_SIZE_MEDIUM**: Balanced infrastructure capacity for production workloads
- **HIGH_RESILIENCE**: Multi-zone redundancy for increased availability
- **VPC peering with private endpoint**: Network isolation with private IP access only
- **Scaled workloads**: 2 schedulers, 2 triggerers, 2-6 workers for production capacity
- **Weekend maintenance window**: Scheduled maintenance on weekends to minimize impact
- **Dedicated service account**: Custom service account for fine-grained IAM control

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-vpc-network-name>` | VPC network name (e.g., "projects/PROJECT_ID/global/networks/NETWORK_NAME") | GCP Console > VPC Network > VPC Networks |
| `<your-vpc-subnetwork-name>` | VPC subnetwork name (e.g., "projects/PROJECT_ID/regions/REGION/subnetworks/SUBNET_NAME") | GCP Console > VPC Network > VPC Networks > Subnets |
| `<your-service-account-email>` | Service account email for Composer nodes | GCP Console > IAM & Admin > Service Accounts |
| `<composer-2-latest-airflow-version>` | Latest Composer 2 image version (e.g., "composer-2.9.7-airflow-2.9.3") | GCP Console > Cloud Composer > Environments > Create > Image Version dropdown |

## Prerequisites

1. A VPC network with a subnetwork in the target region
2. VPC peering configured between Composer and your VPC (or use Private Service Connect)
3. Service account with appropriate permissions for Composer nodes
4. Cloud Composer API enabled
5. Required IAM roles for creating Composer environments

## Important Notes

- Private endpoint means the Airflow web UI is only accessible via private IP
- VPC peering requires proper firewall rules to allow traffic
- HIGH_RESILIENCE mode distributes components across multiple zones
- Maintenance window is set to weekends (Saturday-Sunday) - adjust recurrence as needed
- Worker autoscaling (2-6) adjusts based on workload demand
- Consider adding CMEK encryption for additional security (see 03-enterprise-encrypted preset)

## Related Presets

- **01-dev-small**: Minimal development environment
- **03-enterprise-encrypted**: Enterprise setup with CMEK encryption and web server access control

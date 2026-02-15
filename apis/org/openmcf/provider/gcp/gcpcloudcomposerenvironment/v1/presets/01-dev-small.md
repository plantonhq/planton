# Dev Small

A minimal Cloud Composer environment for development and testing workloads. Uses small resource allocations and basic configuration without private networking or advanced features.

## When to Use

- Development and testing of Airflow DAGs
- Learning Cloud Composer basics
- Small-scale data pipelines with low resource requirements
- Proof-of-concept projects
- Cost-sensitive development environments

## Key Configuration Choices

- **ENVIRONMENT_SIZE_SMALL**: Minimal infrastructure footprint for cost efficiency
- **Basic workloads config**: Small resource allocations (0.5 CPU, 2GB memory) for scheduler, web server, and workers
- **Minimal worker scaling**: 1-3 workers for small development workloads
- **No private networking**: Public endpoint for easy access during development
- **No CMEK encryption**: Uses default GCP encryption
- **No maintenance window**: Allows maintenance at any time
- **No recovery config**: No scheduled snapshots for disaster recovery

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<composer-2-latest-airflow-version>` | Latest Composer 2 image version (e.g., "composer-2.9.7-airflow-2.9.3") | GCP Console > Cloud Composer > Environments > Create > Image Version dropdown |

## Prerequisites

1. A GCP project with billing enabled
2. Cloud Composer API enabled
3. Appropriate IAM permissions to create Composer environments

## Important Notes

- This preset is not suitable for production workloads
- Public endpoint means the Airflow web UI is accessible from the internet
- No disaster recovery or backup configuration
- Consider upgrading to a production preset when moving to production

## Related Presets

- **02-production-private**: Production environment with private networking and high resilience
- **03-enterprise-encrypted**: Enterprise-grade setup with CMEK encryption and full security features

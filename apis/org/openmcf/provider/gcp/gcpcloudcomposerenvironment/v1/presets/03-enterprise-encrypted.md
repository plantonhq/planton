# Enterprise Encrypted

An enterprise-grade Cloud Composer environment with full security features including CMEK encryption, private networking, web server access control, and disaster recovery. Designed for organizations with strict security and compliance requirements.

## When to Use

- Enterprise environments requiring customer-managed encryption keys (CMEK)
- Organizations with strict security and compliance requirements
- Workloads requiring IP-based access restrictions for the web UI
- Environments needing scheduled snapshots for disaster recovery
- Large-scale production workloads requiring generous resource allocations

## Key Configuration Choices

- **ENVIRONMENT_SIZE_LARGE**: Maximum infrastructure capacity for large-scale workloads
- **HIGH_RESILIENCE**: Multi-zone redundancy for maximum availability
- **CMEK encryption**: Customer-managed encryption keys for all Composer-managed resources
- **Private endpoint with VPC_PEERING**: Network isolation with private IP access
- **Web server access control**: IP allowlist restricting web UI access to specific networks
- **Recovery config with scheduled snapshots**: Daily snapshots at 4 AM UTC for disaster recovery
- **Generous workloads config**: Large resource allocations (4 CPU, 15GB memory) for high-performance workloads
- **Scaled worker pool**: 3-10 workers with autoscaling based on demand
- **Weekend maintenance window**: Scheduled maintenance on weekends

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-kms-key-resource-id>` | KMS key resource ID (e.g., "projects/PROJECT_ID/locations/LOCATION/keyRings/RING_NAME/cryptoKeys/KEY_NAME") | GCP Console > Security > Key Management |
| `<your-vpc-network-name>` | VPC network name (e.g., "projects/PROJECT_ID/global/networks/NETWORK_NAME") | GCP Console > VPC Network > VPC Networks |
| `<your-vpc-subnetwork-name>` | VPC subnetwork name (e.g., "projects/PROJECT_ID/regions/REGION/subnetworks/SUBNET_NAME") | GCP Console > VPC Network > VPC Networks > Subnets |
| `<your-service-account-email>` | Service account email for Composer nodes | GCP Console > IAM & Admin > Service Accounts |
| `<composer-2-latest-airflow-version>` | Latest Composer 2 image version (e.g., "composer-2.9.7-airflow-2.9.3") | GCP Console > Cloud Composer > Environments > Create > Image Version dropdown |
| `<your-gcs-snapshot-bucket-path>` | GCS bucket path for snapshots (e.g., "gs://my-composer-snapshots/composer-backups") | GCP Console > Cloud Storage > Buckets |
| `<your-allowed-ip-range-1>` | First allowed IP range (CIDR notation, e.g., "10.0.0.0/8") | Your network administrator |
| `<your-allowed-ip-range-2>` | Second allowed IP range (CIDR notation, e.g., "203.0.113.0/24") | Your network administrator |

## Prerequisites

1. A VPC network with a subnetwork in the target region
2. VPC peering configured between Composer and your VPC
3. A Cloud KMS key with appropriate permissions for Composer service accounts
4. Service account with permissions to use the KMS key
5. A GCS bucket for storing Composer snapshots
6. Cloud Composer API enabled
7. Required IAM roles for creating Composer environments

## Important Notes

- **CMEK encryption**: All Composer-managed resources (GKE nodes, Cloud SQL, Cloud Storage) are encrypted with your KMS key
- **Web server access control**: Only IPs in the allowed ranges can access the Airflow web UI. Add all necessary corporate networks and VPN ranges.
- **Scheduled snapshots**: Daily snapshots are created at 4 AM UTC. Adjust the schedule and timezone as needed.
- **Large resource allocations**: This preset uses generous resources suitable for enterprise workloads. Adjust based on your actual requirements.
- **Worker autoscaling**: Workers scale from 3-10 based on workload. Adjust min/max counts based on your needs.
- **Maintenance window**: Set to weekends (Saturday-Sunday). Modify recurrence to match your maintenance schedule.
- **High cost**: ENVIRONMENT_SIZE_LARGE with generous workloads results in higher costs. Monitor usage and adjust as needed.

## Security Considerations

- Ensure the KMS key has appropriate IAM bindings for Composer service accounts
- Regularly rotate KMS keys according to your organization's key rotation policy
- Review and update IP allowlist ranges regularly
- Monitor snapshot storage costs in GCS
- Consider enabling audit logs for compliance requirements

## Related Presets

- **01-dev-small**: Minimal development environment
- **02-production-private**: Production setup with private networking but without CMEK

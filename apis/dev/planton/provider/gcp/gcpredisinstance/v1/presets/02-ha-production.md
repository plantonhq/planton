# HA Production

This preset provisions a production-ready Memorystore for Redis instance with STANDARD_HA tier, authentication, TLS encryption, RDB persistence, a maintenance window, and deletion protection. It is suitable for production workloads that require high availability and security controls.

## When to Use

- Production application caching with 99.9% availability SLA
- Session storage for stateless web applications
- Workloads requiring encrypted connections and AUTH
- Environments where accidental deletion must be prevented

## Key Configuration

- **STANDARD_HA tier** — primary and replica with automatic failover across zones
- **5 GB memory** — moderate capacity; adjust based on workload
- **authEnabled** — Redis AUTH string required; exported in stack outputs
- **transitEncryptionMode: SERVER_AUTHENTICATION** — TLS for client connections
- **maintenanceWindow** — Sunday 3:00 UTC; GCP applies patches during this window
- **persistenceConfig** — RDB snapshots every 12 hours for durability
- **deletionProtection** — prevents Terraform/Pulumi from destroying the instance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Redis instance (2-40 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `prod-cache`) |
| `<gcp-region>` | GCP region for the instance (e.g., `us-central1`) | [GCP regions](https://cloud.google.com/about/locations) |
| `<vpc-network-self-link>` | Full self-link of the VPC network (e.g., `projects/my-project/global/networks/prod-vpc`) | `GcpVpc` status outputs or GCP Console |

## Related Presets

- **01-basic-cache** — Minimal BASIC tier for dev/test
- **03-ha-read-replicas** — STANDARD_HA with read replicas for read scaling

# Standalone Development

This preset creates a single-instance MySQL HeatWave DB System optimized for development and testing. It uses the smallest available shape, minimal storage, short backup retention, and no HA or delete protection, keeping costs low for non-production environments.

## When to Use

- Local development and feature branch testing against a real MySQL HeatWave instance
- CI/CD environments needing a disposable MySQL database for integration tests
- Sandbox environments for experimenting with MySQL features or schema changes
- Cost-sensitive non-production workloads where downtime is acceptable

## Key Configuration Choices

- **No High Availability** (`isHighlyAvailable: false`) -- single instance keeps costs at roughly one-third of the HA configuration. Acceptable for environments where downtime does not impact users.
- **Smallest shape** (`shapeName: MySQL.VM.Standard.E4.1.8GB`) -- 1 OCPU and 8 GB RAM. Sufficient for development workloads with modest concurrency.
- **50 GB storage** -- minimum practical size for development databases. No auto-expansion to control costs.
- **7-day backup retention** -- provides basic recovery capability without accumulating long-term backup storage costs.
- **No delete protection** (`isDeleteProtected: false`) -- allows quick teardown of development environments without manual flag changes.
- **Skip final backup on deletion** -- development databases are disposable and do not require a final backup.
- **Oracle-managed encryption and certificates** -- simplifies the setup by avoiding KMS and certificate management overhead.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the DB System (e.g., `Uocm:PHX-AD-1`) | OCI Console > Compute > Availability Domains |
| `<private-subnet-ocid>` | OCID of the private subnet for the DB System | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<admin-password>` | Admin password (8-32 chars, must include numeric, lowercase, uppercase, and special character) | Generate a password |

## Related Presets

- **01-high-availability** -- Use instead for production workloads requiring automatic failover and data durability

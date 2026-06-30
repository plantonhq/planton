# Standalone Development

This preset creates a single-instance PostgreSQL DB System with AD-local storage for development and testing. It uses the smallest flex shape configuration, short backup retention, and a plain-text password for simplicity, keeping costs low for non-production environments.

## When to Use

- Local development and feature branch testing against a real OCI PostgreSQL instance
- CI/CD environments needing a disposable PostgreSQL database for integration tests
- Learning and experimentation with OCI PostgreSQL features
- Cost-sensitive non-production workloads where regional durability is unnecessary

## Key Configuration Choices

- **Single instance** (`instanceCount: 1`) -- no read replicas. Keeps costs minimal for development.
- **AD-local storage** (`isRegionallyDurable: false`) -- data resides in a single availability domain, which is cheaper than regional replication. Acceptable for non-production where data loss risk is tolerable.
- **1 OCPU, 8 GB RAM** -- the smallest practical flex shape for PostgreSQL development. Sufficient for testing queries and schema migrations.
- **Plain-text password** (`passwordType: plain_text`) -- avoids the overhead of setting up Vault secrets for development environments. Use `vault_secret` for production.
- **Daily backups with 7-day retention** -- provides basic recovery without accumulating long-term backup costs.
- **No NSGs** -- simplifies the development setup. Add NSGs when moving toward production-like configuration.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<private-subnet-ocid>` | OCID of the private subnet for the DB System | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<availability-domain>` | AD for the DB System (e.g., `Uocm:PHX-AD-1`) | OCI Console > Compute > Availability Domains |
| `<admin-password>` | Administrator password for the postgres user | Generate a password |

## Related Presets

- **01-regionally-durable** -- Use instead for production workloads requiring regional storage durability and read replicas

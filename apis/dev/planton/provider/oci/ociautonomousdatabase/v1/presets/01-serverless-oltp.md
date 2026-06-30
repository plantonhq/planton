# Serverless OLTP (Autonomous Transaction Processing)

This preset creates a serverless Autonomous Transaction Processing (ATP) database with the ECPU compute model, private endpoint networking, and auto-scaling for both compute and storage. ATP is optimized for mixed transactional workloads and is the most common Autonomous Database deployment pattern for application backends, microservices, and SaaS platforms.

## When to Use

- Application backends needing a fully managed Oracle database with minimal operational overhead
- Microservices architectures where each service owns its database
- SaaS platforms that need elastic scaling for variable workloads
- Any OLTP workload that benefits from autonomous patching, tuning, and scaling

## Key Configuration Choices

- **OLTP workload** (`dbWorkload: oltp`) -- optimizes the database engine for mixed transactional workloads with low-latency reads and writes.
- **ECPU compute model** (`computeModel: ecpu`) -- Oracle's current recommended billing model. 4 ECPUs provide a solid baseline for production applications.
- **Auto-scaling enabled** for both compute and storage -- compute can burst to 3x (12 ECPUs) during demand spikes; storage expands automatically as data grows. This eliminates manual capacity management.
- **Private endpoint** (`subnetId` + `nsgIds`) -- the database is only accessible from the specified subnet and peered networks. No public endpoint is created, ensuring the database is never exposed to the internet.
- **TLS-only connections** (`isMtlsConnectionRequired: false`) -- allows standard TLS connections from applications without requiring wallet-based mTLS. This simplifies client configuration while maintaining encryption in transit.
- **Standard Edition** (`databaseEdition: standard_edition`) -- covers most OLTP workloads. Upgrade to Enterprise Edition only if advanced features like partitioning or advanced compression are needed.
- **License Included** (`licenseModel: license_included`) -- licensing cost is bundled in the service price. Switch to `bring_your_own_license` if you have existing Oracle Database licenses.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the database | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<admin-password>` | Administrator password (12-30 chars, must include uppercase, lowercase, and numeric) | Generate a strong password; for production use `secretId` instead |
| `<private-subnet-ocid>` | OCID of the private subnet for the database endpoint | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<db-nsg-ocid>` | OCID of the network security group allowing database traffic (port 1522) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |

## Related Presets

- **02-free-tier-development** -- Use instead for zero-cost development and experimentation
- **03-serverless-data-warehouse** -- Use instead for analytic and reporting workloads (ADW)

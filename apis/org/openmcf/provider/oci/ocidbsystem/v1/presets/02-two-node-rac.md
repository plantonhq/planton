# Two-Node RAC

This preset creates a two-node Real Application Clusters (RAC) Oracle Database System for high availability. The nodes are distributed across fault domains for infrastructure-level resilience, with rolling maintenance for zero-downtime patching and a dedicated backup subnet for inter-node communication. Enterprise Edition Extreme Performance is required for RAC.

## When to Use

- Production workloads requiring zero-downtime patching and automatic failover
- Mission-critical databases where single-node failure must not cause an outage
- Applications needing active-active database instances for read/write scalability
- Compliance requirements mandating high-availability database infrastructure

## Key Configuration Choices

- **Two-node RAC** (`nodeCount: 2`) -- Oracle RAC provides active-active clustering where both nodes serve application connections. If one node fails, the other continues serving requests without application changes.
- **Enterprise Edition Extreme Performance** (`databaseEdition: enterprise_edition_extreme_performance`) -- required for RAC. Includes all Enterprise Edition features plus in-memory database and Active Data Guard.
- **Fault domain separation** (`faultDomains: ["FAULT-DOMAIN-1", "FAULT-DOMAIN-2"]`) -- places each RAC node in a different fault domain to survive hardware failures affecting a single fault domain.
- **4 OCPUs per node** (`cpuCoreCount: 4`) on VM.Standard.E4.Flex -- provides meaningful compute capacity per node. Each node gets 4 OCPUs, so the cluster has 8 total.
- **1 TB storage** (`dataStorageSizeInGb: 1024`) -- shared ASM storage accessible by both nodes.
- **Dedicated backup subnet** (`backupSubnetId`) -- required for RAC inter-node communication (cache fusion). Must be a separate subnet from the client network.
- **Rolling maintenance** (`patchingMode: rolling`) -- patches are applied one node at a time while the other continues serving traffic, achieving zero downtime during quarterly patching windows.
- **60-day backup retention** -- extended retention for production databases with stricter recovery requirements.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the DB System (e.g., `Uocm:PHX-AD-1`) | OCI Console > Compute > Availability Domains |
| `<private-subnet-ocid>` | OCID of the private subnet for client connections | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<ssh-public-key>` | SSH public key in OpenSSH format for node access | Your `~/.ssh/id_rsa.pub` or equivalent |
| `<db-nsg-ocid>` | OCID of the NSG allowing database traffic (port 1521) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |
| `<backup-subnet-ocid>` | OCID of the backup subnet for RAC inter-node traffic | OCI Console > Networking > Subnets (separate from client subnet) |
| `<admin-password>` | SYS/SYSTEM password (2-30 chars, uppercase + lowercase + numeric) | Generate a strong password |

## Related Presets

- **01-single-node-vm** -- Use instead for non-HA workloads where single-node cost efficiency is preferred

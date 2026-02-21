# Single-Node VM

This preset creates a single-node Oracle Database System on a VM.Standard.E4.Flex shape with Standard Edition, automatic backups, and a pluggable database. This is the baseline for running a managed Oracle Database when Autonomous Database is not suitable -- for example, when you need RAC-specific features, custom PL/SQL extensions, or full control over the Oracle Database instance.

## When to Use

- Applications requiring a full Oracle Database instance with DBA-level control
- Legacy workloads being migrated to OCI that need specific Oracle Database features unavailable in Autonomous Database
- Development and staging environments matching production Oracle DB configurations
- Workloads that need pluggable databases (PDBs) for multi-tenant database consolidation

## Key Configuration Choices

- **VM.Standard.E4.Flex shape** with 2 OCPUs -- AMD EPYC-based flex shape allowing independent CPU and memory scaling. 2 OCPUs provide a cost-effective baseline; scale up by increasing `cpuCoreCount`.
- **Standard Edition** (`databaseEdition: standard_edition`) -- covers most workloads without the premium of Enterprise Edition. Upgrade to `enterprise_edition` for partitioning, advanced compression, or Data Guard.
- **256 GB storage** (`dataStorageSizeInGb: 256`) -- the minimum supported size for VM DB Systems. Suitable for databases up to ~200 GB after accounting for ASM overhead.
- **Single node** (`nodeCount: 1`) -- no RAC clustering. For HA, consider the 02-two-node-rac preset instead.
- **Oracle 19c** (`dbVersion: "19.0.0.0"`) -- the long-term support release. Change to `"21.0.0.0"` or `"23.0.0.0"` for innovation releases.
- **Automatic backups** with 30-day retention -- provides point-in-time recovery for the last 30 days. OCI manages the backup schedule and Object Storage destination.
- **Pluggable database** (`pdbName: mypdb`) -- creates a PDB within the CDB for application schema isolation, following Oracle's multi-tenant architecture best practice.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the DB System (e.g., `Uocm:PHX-AD-1`) | OCI Console > Compute > Availability Domains, or `oci iam availability-domain list` |
| `<private-subnet-ocid>` | OCID of the private subnet for the DB System | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<ssh-public-key>` | SSH public key in OpenSSH format for node access | Your `~/.ssh/id_rsa.pub` or equivalent |
| `<db-nsg-ocid>` | OCID of the NSG allowing database traffic (port 1521) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |
| `<admin-password>` | SYS/SYSTEM password (2-30 chars, uppercase + lowercase + numeric) | Generate a strong password |

## Related Presets

- **02-two-node-rac** -- Use instead for high-availability production workloads requiring Real Application Clusters

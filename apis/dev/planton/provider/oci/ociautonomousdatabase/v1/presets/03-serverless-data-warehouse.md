# Serverless Data Warehouse (Autonomous Data Warehouse)

This preset creates a serverless Autonomous Data Warehouse (ADW) with the ECPU compute model, Enterprise Edition for advanced analytic features, private endpoint networking, and auto-scaling. ADW is optimized for analytic queries, reporting dashboards, and data lake workloads where large scans and aggregations dominate.

## When to Use

- Business intelligence and reporting platforms querying large datasets
- Data lake architectures combining Object Storage data with managed tables
- ETL/ELT pipelines that load and transform data for downstream analytics
- Machine learning workloads using in-database Oracle Machine Learning (OML)

## Key Configuration Choices

- **DW workload** (`dbWorkload: dw`) -- optimizes the database engine for analytic queries with columnar storage, smart scan, and result caching.
- **Enterprise Edition** (`databaseEdition: enterprise_edition`) -- enables partitioning and advanced compression, which are critical for large analytic datasets. Standard Edition lacks these capabilities.
- **8 ECPUs** (`computeCount: 8`) -- provides substantial query concurrency for analytic dashboards. With auto-scaling, bursts to 24 ECPUs are available during peak report generation.
- **4 TB storage** (`dataStorageSizeInTbs: 4`) -- sized for a mid-scale data warehouse. Storage auto-scaling expands automatically as data volume grows.
- **Private endpoint** (`subnetId` + `nsgIds`) -- analytics databases typically contain sensitive business data and should not be internet-accessible.
- **TLS-only connections** (`isMtlsConnectionRequired: false`) -- simplifies connectivity from BI tools (Oracle Analytics Cloud, Tableau, Power BI) that connect via standard TLS.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the database | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<admin-password>` | Administrator password (12-30 chars, must include uppercase, lowercase, and numeric) | Generate a strong password; for production use `secretId` instead |
| `<private-subnet-ocid>` | OCID of the private subnet for the database endpoint | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<db-nsg-ocid>` | OCID of the network security group allowing database traffic (port 1522) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |

## Related Presets

- **01-serverless-oltp** -- Use instead for transactional application workloads (ATP)
- **02-free-tier-development** -- Use instead for zero-cost development and experimentation

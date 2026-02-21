# AlicloudRdsInstance

Manages an Alibaba Cloud RDS (Relational Database Service) instance with bundled databases, accounts, and account privileges.

## Overview

RDS is Alibaba Cloud's managed relational database service supporting multiple engines (MySQL, PostgreSQL, SQL Server, MariaDB, PPAS) through a single resource type. This component bundles the instance with its databases, accounts, and account privileges into a single deployable unit because an RDS instance without databases and accounts is incomplete for application use.

### What Gets Created

- **RDS Instance** -- a managed database instance with the selected engine and architecture
- **Databases** -- logical databases within the instance
- **Accounts** -- database user accounts with passwords
- **Account Privileges** -- grants linking accounts to databases with specific access levels

### Engine Selection

The `engine` field selects the database engine. All engines share the same component interface, with engine-specific defaults (character sets, ports) handled automatically.

| Engine | Typical Versions | Default Port | Default Charset |
|--------|------------------|-------------|-----------------|
| MySQL | 5.7, 8.0 | 3306 | utf8mb4 |
| PostgreSQL | 14, 15, 16 | 5432 | UTF8 |
| SQLServer | 2019_ent, 2022_ent | 1433 | Chinese_PRC_CI_AS |
| MariaDB | 10.3 | 3306 | utf8mb4 |
| PPAS | 10, 11 | 5432 | UTF8 |

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Go build (Pulumi module)
go build ./apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/iac/pulumi/...

# Go vet
go vet ./apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/iac/pulumi/...

# Spec tests
go test ./apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/...

# Terraform validation
cd apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.

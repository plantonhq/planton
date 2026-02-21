# AliCloudPolardbCluster

Manages an Alibaba Cloud PolarDB cluster with bundled databases, accounts, and account privileges.

## Overview

PolarDB is Alibaba Cloud's cloud-native relational database service built on a shared-storage, compute-storage-separated architecture. It supports MySQL, PostgreSQL, and Oracle compatibility modes through a single component type. This component bundles the cluster with its databases, accounts, and account privileges into a single deployable unit because a PolarDB cluster without databases and accounts is incomplete for application use.

### What Gets Created

- **PolarDB Cluster** -- a cloud-native database cluster with the selected engine, node class, and node count
- **Databases** -- logical databases within the cluster
- **Accounts** -- database user accounts with passwords
- **Account Privileges** -- grants linking accounts to databases with specific access levels

### Engine Selection

The `db_type` field selects the database engine. All engines share the same component interface, with engine-specific defaults (character sets, ports) handled automatically.

| Engine | Typical Versions | Default Port | Default Charset |
|--------|------------------|-------------|-----------------|
| MySQL | 5.6, 5.7, 8.0 | 3306 | utf8 |
| PostgreSQL | 11, 14 | 5432 | UTF8 |
| Oracle | 11 | 1521 | UTF8 |

### PolarDB Editions

PolarDB has two main editions controlled by `creation_category`:

- **Enterprise Edition** (`Normal`) -- shared distributed storage (PSL4/PSL5), auto-scaling storage, compute-storage separation
- **Standard Edition** (`SENormal`) -- local ESSD storage, pre-allocated storage via `storage_space`

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Go build (Pulumi module)
go build ./apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/pulumi/...

# Go vet
go vet ./apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/pulumi/...

# Spec tests
go test ./apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/...

# Terraform validation
cd apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.

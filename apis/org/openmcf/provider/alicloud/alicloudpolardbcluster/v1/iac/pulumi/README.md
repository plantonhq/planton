# AlicloudPolardbCluster -- Pulumi Module

## Overview

This Pulumi module deploys an Alibaba Cloud PolarDB cluster with bundled databases, accounts, and account privileges.

## Module Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point; loads stack input and invokes module |
| `module/main.go` | Cluster creation, orchestrates databases and accounts |
| `module/locals.go` | Tag computation, default helpers |
| `module/outputs.go` | Output key constants |
| `module/databases.go` | Database creation with charset defaults |
| `module/accounts.go` | Account and privilege creation |

## Local Development

```bash
cd /path/to/openmcf-alibaba-cloud

# Build
go build ./apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/pulumi/...

# Vet
go vet ./apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1/iac/pulumi/...
```

## Provider

Uses `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/polardb` for all PolarDB resources.

# AliCloudMongodbInstance

Alibaba Cloud ApsaraDB for MongoDB replica-set instance.

## Overview

This component provisions a managed MongoDB replica-set instance on Alibaba Cloud using the `alicloud_mongodb_instance` Terraform resource. It supports configurable replication factors, multi-zone HA, read-only replicas, TDE and cloud disk encryption, and operational controls such as backup and maintenance windows.

## Architecture

The component creates a single MongoDB replica-set instance. The replica set consists of a primary, secondary, and hidden node (when `replicationFactor` >= 3). Multi-zone HA is achieved by placing each node in a different availability zone via `zoneId`, `secondaryZoneId`, and `hiddenZoneId`.

## Build and Test

All commands are scoped to this component directory:

```bash
cd apis/dev/planton/provider/alicloud/alicloudmongodbinstance/v1/

# Go build and vet
go build ./...
go vet ./...

# Run spec validation tests
go test ./...

# Terraform validation
cd iac/tf/
terraform init
terraform validate
```

## Documentation

- [Catalog Page](catalog-page.md) -- user-facing documentation
- [Examples](examples.md) -- YAML configuration examples

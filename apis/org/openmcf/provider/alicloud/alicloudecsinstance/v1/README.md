# AliCloudEcsInstance

Alibaba Cloud Elastic Compute Service (ECS) instance with inline system and data disks.

## Overview

This component provisions an ECS compute instance on Alibaba Cloud using the `alicloud_instance` Terraform resource. It supports configurable instance types, OS images, system disk tuning, up to 16 additional data disks, SSH key or password authentication, public IP allocation, spot pricing, and PrePaid subscription billing.

Data disks are created inline with the instance (not as separate resources), keeping their lifecycle tied to the instance per DD07 composite bundling.

## Architecture

The component creates a single `alicloud_instance` resource. The instance is placed in a VSwitch (which determines VPC and availability zone) and associated with one or more security groups. An optional public IP is allocated when `internetMaxBandwidthOut` > 0.

## Build and Test

All commands are scoped to this component directory:

```bash
cd apis/org/openmcf/provider/alicloud/alicloudecsinstance/v1/

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

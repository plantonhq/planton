# AliCloudRocketmqInstance

Manages an Alibaba Cloud RocketMQ 5.x instance with bundled topics and consumer groups.

## Overview

RocketMQ is Alibaba Cloud's distributed messaging and streaming platform, supporting normal, FIFO, delayed, and transactional messages. This component targets the RocketMQ 5.x API, which provides VPC-integrated instances with configurable throughput tiers, billing modes, and optional internet access.

### What Gets Created

- **RocketMQ Instance** -- a managed message broker with the selected edition and deployment architecture
- **Topics** -- message channels with configurable message types (NORMAL, FIFO, DELAY, TRANSACTION)
- **Consumer Groups** -- logical consumer identities with retry policies and delivery ordering

### Instance Editions

The edition is controlled by two fields: `series_code` (feature tier) and `sub_series_code` (deployment architecture).

| Series | Sub-series | Use Case |
|--------|-----------|----------|
| standard | single_node | Development and testing |
| standard | cluster_ha | Light production workloads |
| professional | cluster_ha | Production with higher throughput |
| professional | serverless | Auto-scaling production |
| ultimate | cluster_ha | Mission-critical, highest throughput |

### Bundled Resources

Topics and consumer groups are bundled because they are meaningless without a parent instance. ACL accounts and permissions are intentionally excluded -- security configuration has an independent lifecycle and should be managed separately.

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Go build (Pulumi module)
go build ./apis/dev/planton/provider/alicloud/alicloudrocketmqinstance/v1/iac/pulumi/...

# Go vet
go vet ./apis/dev/planton/provider/alicloud/alicloudrocketmqinstance/v1/iac/pulumi/...

# Spec tests
go test ./apis/dev/planton/provider/alicloud/alicloudrocketmqinstance/v1/...

# Terraform validation
cd apis/dev/planton/provider/alicloud/alicloudrocketmqinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.

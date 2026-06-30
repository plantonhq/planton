# AliCloudNatGateway

Manages an Alibaba Cloud Enhanced NAT Gateway with bundled EIP association and SNAT entries.

## Overview

A NAT Gateway enables resources in private VSwitches (with no public IP) to access the internet via Source NAT (SNAT). This component bundles the NAT Gateway, its EIP association, and SNAT entries into a single deployable unit because a NAT Gateway without an EIP and SNAT entries is non-functional.

### What Gets Created

- **NAT Gateway** -- an Enhanced NAT Gateway placed in a VPC/VSwitch
- **EIP Association** -- binds an existing Elastic IP to the NAT Gateway
- **SNAT Entries** -- map private VSwitch or CIDR traffic to the EIP for outbound internet access

### How SNAT Works

Each SNAT entry maps a source (VSwitch ID or CIDR block) to the NAT Gateway's associated EIP. When instances in the specified source send outbound traffic, the NAT Gateway translates the source IP to the EIP's public IP. Multiple SNAT entries can target different VSwitches or CIDRs through the same EIP.

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Proto compilation (from planton repo root, once after proto changes)
make protos

# Go build (Pulumi module)
go build ./apis/dev/planton/provider/alicloud/alicloudnatgateway/v1/iac/pulumi/...

# Go vet
go vet ./apis/dev/planton/provider/alicloud/alicloudnatgateway/v1/iac/pulumi/...

# Spec tests
go test ./apis/dev/planton/provider/alicloud/alicloudnatgateway/v1/...

# Terraform validation
cd apis/dev/planton/provider/alicloud/alicloudnatgateway/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.

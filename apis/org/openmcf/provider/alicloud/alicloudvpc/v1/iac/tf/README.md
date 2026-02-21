# Terraform Module to Deploy AliCloudVpc

This module provisions an Alibaba Cloud VPC with configurable CIDR block, optional IPv6 support, resource group assignment, and automatic tag management. It creates a single `alicloud_vpc` resource and outputs the VPC ID, name, CIDR block, router ID, and route table ID.

Generated `variables.tf` reflects the proto schema for `AliCloudVpc`.

## Usage

Use the OpenMCF CLI (tofu) with the default local backend:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).

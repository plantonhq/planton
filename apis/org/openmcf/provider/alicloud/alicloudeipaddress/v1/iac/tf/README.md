# Terraform Module to Deploy AliCloudEipAddress

This module provisions an Alibaba Cloud Elastic IP Address (EIP) using the
`alicloud_eip_address` resource. The EIP is a standalone public IPv4 address
that can be associated with NAT gateways, load balancers, and other resources.

Generated `variables.tf` reflects the proto schema for `AliCloudEipAddress`.

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

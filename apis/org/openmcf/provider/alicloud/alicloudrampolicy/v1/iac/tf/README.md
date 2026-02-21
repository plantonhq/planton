# Terraform Module to Deploy AlicloudRamPolicy

This module provisions a single Alibaba Cloud RAM custom policy. The
`alicloud_ram_policy` resource is configured with a JSON policy document,
optional version rotation, and tag management.

Generated `variables.tf` reflects the proto schema for `AlicloudRamPolicy`.

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

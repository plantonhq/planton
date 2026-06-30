# Terraform Module to Deploy AliCloudRamPolicy

This module provisions a single Alibaba Cloud RAM custom policy. The
`alicloud_ram_policy` resource is configured with a JSON policy document,
optional version rotation, and tag management.

Generated `variables.tf` reflects the proto schema for `AliCloudRamPolicy`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).

# Terraform Module to Deploy AliCloudLogProject

This module provisions an Alibaba Cloud SLS project with bundled log stores and
full-text indexes. The `for_each` pattern in `main.tf` creates one
`alicloud_log_store` per entry in `spec.log_stores` and one
`alicloud_log_store_index` for each store where `enable_index` is true.

Generated `variables.tf` reflects the proto schema for `AliCloudLogProject`.

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

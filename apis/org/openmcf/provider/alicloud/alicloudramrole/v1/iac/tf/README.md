# Terraform Module to Deploy AliCloudRamRole

This module provisions an Alibaba Cloud RAM role with bundled policy
attachments. Each entry in `policy_attachments` creates an
`alicloud_ram_role_policy_attachment` resource linked to the role via
`for_each`.

Generated `variables.tf` reflects the proto schema for `AliCloudRamRole`.

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

# Terraform Module to Deploy AliCloudKmsKey

This module provisions an Alibaba Cloud KMS customer-managed key (CMK) using
the `alicloud_kms_key` resource. The key can be used for data encryption or
digital signing across Alibaba Cloud services.

Generated `variables.tf` reflects the proto schema for `AliCloudKmsKey`.

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

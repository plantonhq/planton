# Terraform Module to Deploy AliCloudSecurityGroup

This module provisions an Alibaba Cloud Security Group with bundled security
rules. Each entry in `rules` creates an `alicloud_security_group_rule` resource
linked to the security group via `for_each`.

Generated `variables.tf` reflects the proto schema for `AliCloudSecurityGroup`.

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

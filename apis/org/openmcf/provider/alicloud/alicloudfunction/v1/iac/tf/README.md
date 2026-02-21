# Terraform Module to Deploy AliCloudFunction

This module provisions an Alibaba Cloud Function Compute v3 function. The
`main.tf` file creates a single `alicloud_fcv3_function` resource with dynamic
blocks for optional configuration sections (VPC, logging, custom container,
custom runtime, lifecycle hooks, NAS, GPU).

Generated `variables.tf` reflects the proto schema for `AliCloudFunctionSpec`.

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

# Terraform Module to Deploy AliCloudVswitch

This module provisions an Alibaba Cloud VSwitch (subnet) within an existing VPC. It creates a single `alicloud_vswitch` resource bound to a specific Availability Zone with a configured IPv4 CIDR block, optional IPv6 support, and automatic tag management. It outputs the VSwitch ID, name, CIDR block, zone ID, and IPv6 CIDR block.

Generated `variables.tf` reflects the proto schema for `AliCloudVswitch`.

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

## Files

| File | Purpose |
|------|---------|
| `provider.tf` | Alibaba Cloud provider configuration (region from spec) |
| `variables.tf` | Input variables mirroring the proto spec with validations |
| `locals.tf` | Tag computation from metadata and spec.tags |
| `main.tf` | VSwitch resource definition |
| `outputs.tf` | Stack outputs matching `stack_outputs.proto` |

## Outputs

| Output | Description |
|--------|-------------|
| `vswitch_id` | The VSwitch ID assigned by Alibaba Cloud |
| `vswitch_name` | The VSwitch name as created |
| `cidr_block` | The IPv4 CIDR block of the VSwitch |
| `zone_id` | The Availability Zone |
| `ipv6_cidr_block` | The IPv6 CIDR block (empty if IPv6 is not enabled) |

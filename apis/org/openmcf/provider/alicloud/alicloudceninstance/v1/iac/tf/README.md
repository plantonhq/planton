# Terraform Module to Deploy AliCloudCenInstance

This module provisions an Alibaba Cloud CEN (Cloud Enterprise Network) instance
with bundled child-instance attachments using `alicloud_cen_instance` and
`alicloud_cen_instance_attachment` resources.

## Resources Created

- `alicloud_cen_instance.main` — the CEN hub instance
- `alicloud_cen_instance_attachment.attachments` — one per attachment (via `for_each`)

## Usage

Use the OpenMCF CLI (tofu) with the default local backend:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

Credentials are provided via stack input (CLI), not in the manifest `spec`.

## Module Structure

| File | Purpose |
|------|---------|
| `main.tf` | CEN instance and attachment resources |
| `locals.tf` | Tag computation and attachment list-to-map conversion |
| `variables.tf` | Input variables with validation rules |
| `outputs.tf` | CEN ID and instance name outputs |
| `provider.tf` | AliCloud provider configuration scoped to `spec.region` |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cen_id` | `string` | CEN instance ID |
| `cen_instance_name` | `string` | CEN instance name |

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).

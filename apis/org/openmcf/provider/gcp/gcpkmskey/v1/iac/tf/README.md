# GcpKmsKey — Terraform Module

This directory contains the Terraform implementation for the GcpKmsKey component.

## Module Structure

```
provider.tf    — Google provider configuration
variables.tf   — Input variables matching GcpKmsKeySpec
locals.tf      — Derived values from spec
main.tf        — google_kms_crypto_key resource
outputs.tf     — Outputs matching GcpKmsKeyStackOutputs
```

## What It Creates

- `google_kms_crypto_key.this` — A Cloud KMS cryptographic key in the specified key ring

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `spec.key_ring_id.value` | `string` | Fully qualified key ring path |
| `spec.key_name` | `string` | Name of the crypto key |
| `spec.purpose` | `string` | Key purpose (default: ENCRYPT_DECRYPT) |
| `spec.rotation_period` | `string` | Auto-rotation period (e.g., "7776000s") |
| `spec.destroy_scheduled_duration` | `string` | Destroy schedule (default: 30 days) |
| `spec.version_template.algorithm` | `string` | Encryption algorithm |
| `spec.version_template.protection_level` | `string` | SOFTWARE or HSM |
| `spec.skip_initial_version_creation` | `bool` | Skip initial version |
| `provider_config.service_account_key_base64` | `string` | Optional GCP service account key |

## Outputs

| Output | Description |
|--------|-------------|
| `key_id` | Fully qualified resource path |
| `key_name` | Short name of the crypto key |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Important Notes

- **Keys cannot be deleted from GCP.** Running `terraform destroy` destroys all key versions and disables rotation, but the key itself remains in GCP.
- **Most fields are immutable.** Changing name, purpose, key_ring, destroy_scheduled_duration, or protection_level triggers a destroy-and-recreate cycle.
- **Labels are managed by OpenMCF framework.** Framework labels (resource kind, org, env) are applied automatically.

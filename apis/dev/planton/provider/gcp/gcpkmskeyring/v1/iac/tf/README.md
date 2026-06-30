# GcpKmsKeyRing — Terraform Module

This directory contains the Terraform implementation for the GcpKmsKeyRing component.

## Module Structure

```
provider.tf    — Google provider configuration
variables.tf   — Input variables matching GcpKmsKeyRingSpec
locals.tf      — Derived values from spec
main.tf        — google_kms_key_ring resource
outputs.tf     — Outputs matching GcpKmsKeyRingStackOutputs
```

## What It Creates

- `google_kms_key_ring.this` — A Cloud KMS key ring in the specified project and location

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `spec.project_id.value` | `string` | GCP project ID |
| `spec.key_ring_name` | `string` | Name of the key ring |
| `spec.location` | `string` | GCP location (region, multi-region, or `global`) |
| `provider_config.service_account_key_base64` | `string` | Optional GCP service account key |

## Outputs

| Output | Description |
|--------|-------------|
| `key_ring_id` | Fully qualified resource path (`projects/{project}/locations/{location}/keyRings/{name}`) |
| `key_ring_name` | Short name of the key ring |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Important Notes

- **Key rings cannot be deleted from GCP.** Running `terraform destroy` only removes the resource from Terraform state.
- **All fields are ForceNew.** Any change triggers a destroy-and-recreate cycle.
- **No labels support.** GCP KMS key rings do not accept resource labels.

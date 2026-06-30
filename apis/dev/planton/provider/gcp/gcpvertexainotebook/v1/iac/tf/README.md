# Terraform Module: GcpVertexAiNotebook

## Resources

- `google_workbench_instance.this` — Vertex AI Workbench instance

## Provider

Requires the `hashicorp/google` provider version `~> 6.0`.

## Variables

### Required

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton resource metadata (name, id, org, env) |
| `spec` | object | GcpVertexAiNotebook spec (see variables.tf for full schema) |

### Optional

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `provider_config` | object | `{}` | GCP provider configuration (service account key) |

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | Fully qualified instance ID |
| `instance_name` | Short instance name |
| `proxy_uri` | JupyterLab proxy URL |
| `state` | Current instance state |
| `creator` | Email of creator |
| `create_time` | RFC3339 creation timestamp |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

## Feature Parity

This Terraform module has full feature parity with the Pulumi module:

- GCE setup block with machine type, disks, accelerator, networking
- CMEK encryption derived from KMS key presence
- VM image and container image support (mutually exclusive)
- Shielded VM configuration
- Framework GCP labels
- All 6 stack outputs

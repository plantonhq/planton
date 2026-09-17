# Terraform Module: GcpVertexAiNotebook

## Resources

- `google_workbench_instance.this` — Vertex AI Workbench instance
- `google_project_service.notebooks_api` / `google_project_service.compute_api` — API enablement (never disabled on destroy)

## Provider

Requires the `hashicorp/google` provider version `~> 8.3`.

`spec.project_id` is optional: when empty, the instance lands in the provider's
default project. Credentials are never module inputs: the provider block is
empty and the runner injects `GOOGLE_CREDENTIALS` (or the ambient ADC chain
applies) — the catalog-wide contract.

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton resource metadata (name, id, org, env) |
| `spec` | object | GcpVertexAiNotebook spec (see variables.tf for full schema) |

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | Fully qualified instance ID |
| `instance_name` | Short instance name |
| `proxy_uri` | JupyterLab proxy URL |
| `state` | Current instance state |
| `creator` | Email of creator |
| `create_time` | RFC3339 creation timestamp |
| `health_state` | Instance health as reported by the Workbench health service |
| `update_time` | RFC3339 timestamp of the most recent update |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

## Feature Parity

This Terraform module has full feature parity with the Pulumi module:

- GCE setup block with machine type, disks (PD + hyperdisk types), accelerator (through the current H100/H200/B200 generation), networking (incl. static external IP)
- Confidential Computing (SEV) and reservation affinity
- Managed end-user credentials and third-party identity access
- CMEK encryption derived from KMS key presence
- VM image and container image support (mutually exclusive)
- Shielded VM configuration
- Client-side `deletion_policy` (DELETE / PREVENT / ABANDON), mirrored zero-vs-omit
- User labels merged beneath platform attribution labels (identical merge order)
- All 8 stack outputs

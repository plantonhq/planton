# GcpSpannerInstance Terraform Module

This Terraform module provisions a Google Cloud Spanner instance.

## Usage

```bash
# Initialize
terraform init

# Preview changes
terraform plan -var-file=terraform.tfvars

# Apply
terraform apply -var-file=terraform.tfvars
```

## Inputs

See `variables.tf` for the full list of inputs. Key inputs:

- `spec.project_id` - GCP project ID
- `spec.instance_name` - Spanner instance name (6-30 chars)
- `spec.config` - Instance configuration (e.g., "regional-us-central1")
- `spec.display_name` - Human-readable display name (4-30 chars)
- `spec.num_nodes` - Number of nodes (mutually exclusive with processing_units/autoscaling)
- `spec.processing_units` - Processing units (mutually exclusive with num_nodes/autoscaling)
- `spec.autoscaling_config` - Autoscaling configuration (mutually exclusive with num_nodes/processing_units)
- `spec.instance_type` - PROVISIONED or FREE_INSTANCE
- `spec.edition` - STANDARD, ENTERPRISE, or ENTERPRISE_PLUS

## Outputs

- `instance_id` - Fully qualified instance ID
- `instance_name` - Short instance name
- `state` - Instance state (CREATING or READY)

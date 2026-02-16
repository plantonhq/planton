# GcpRedisInstance Terraform Module

This Terraform module provisions a Google Cloud Memorystore for Redis instance.

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
- `spec.instance_name` - Redis instance name
- `spec.region` - GCP region
- `spec.tier` - BASIC or STANDARD_HA
- `spec.memory_size_gb` - Memory size in GiB

## Outputs

- `host` - Primary Redis endpoint IP
- `port` - Primary Redis port
- `read_endpoint` - Read replica endpoint (HA with read replicas)
- `read_endpoint_port` - Read replica port
- `current_location_id` - Zone of the primary
- `auth_string` - Redis AUTH string (sensitive)

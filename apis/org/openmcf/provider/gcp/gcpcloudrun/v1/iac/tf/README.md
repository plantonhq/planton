# Terraform Module for GCP Cloud Run

This Terraform module deploys a Google Cloud Run service with support for:

- Container deployment with configurable CPU and memory
- Environment variables and Secret Manager integration
- Custom DNS domain mapping
- VPC access for private resources
- Auto-scaling configuration
- IAM policy management for public/private access
- Multiple ingress options
- Service account configuration

## Prerequisites

- Terraform >= 1.0
- Google Cloud provider >= 6.19.0
- GCP project with Cloud Run API enabled
- Appropriate IAM permissions

## Usage

### Basic Example

```hcl
module "cloudrun" {
  source = "./iac/tf"

  metadata = {
    name = "my-service"
    env  = "production"
  }

  spec = {
    project_id = "my-gcp-project"
    region     = "us-central1"

    container = {
      image = {
        repo = "gcr.io/my-project/my-app"
        tag  = "v1.0.0"
      }
      cpu    = 1
      memory = 512
      replicas = {
        min = 0
        max = 10
      }
    }
  }
}
```

### With Custom DNS

```hcl
module "cloudrun_with_dns" {
  source = "./iac/tf"

  metadata = {
    name = "api-service"
  }

  spec = {
    project_id = "my-gcp-project"
    region     = "us-central1"

    container = {
      image = {
        repo = "gcr.io/my-project/api"
        tag  = "v2.0.0"
      }
      cpu    = 2
      memory = 1024
      replicas = {
        min = 1
        max = 50
      }
    }

    dns = {
      enabled      = true
      hostnames    = ["api.example.com"]
      managed_zone = "example-com-zone"
    }
  }
}
```

## Deployment Commands

### Using OpenMCF CLI

```shell
# Initialize Terraform with remote state backend
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=openmcf/gcp-stacks/gcp-cloud-run.tfstate"

# Plan deployment
openmcf tofu plan --manifest hack/manifest.yaml

# Apply deployment
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve

# Destroy resources
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

### Using Standard Terraform

```shell
# Initialize
terraform init

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy
```

## Inputs

See [variables.tf](./variables.tf) for complete input specification.

### Required Inputs

| Name | Type | Description |
|------|------|-------------|
| metadata | object | Resource metadata including name, org, env |
| spec.project_id | string | GCP project ID |
| spec.region | string | GCP region (e.g., "us-central1") |
| spec.container.image | object | Container image repo and tag |
| spec.container.cpu | number | CPU units (1, 2, or 4) |
| spec.container.memory | number | Memory in MiB (128-32768) |
| spec.container.replicas | object | Min and max instance counts |

### Optional Inputs

| Name | Type | Default | Description |
|------|------|---------|-------------|
| spec.service_name | string | metadata.name | Custom service name |
| spec.service_account | string | null | Service account email |
| spec.max_concurrency | number | 80 | Max concurrent requests per instance |
| spec.timeout_seconds | number | 300 | Request timeout in seconds |
| spec.ingress | string | "INGRESS_TRAFFIC_ALL" | Ingress traffic source |
| spec.allow_unauthenticated | bool | true | Allow public access |
| spec.execution_environment | string | "EXECUTION_ENVIRONMENT_GEN2" | Execution environment |
| spec.vpc_access | object | null | VPC network access configuration |
| spec.dns | object | null | Custom DNS domain mapping |

## Outputs

| Name | Description |
|------|-------------|
| url | Cloud Run service URL |
| service_name | Service name |
| revision | Latest revision name |

## Features

### Container Configuration

- Configurable CPU (1, 2, or 4 vCPU)
- Memory allocation (128 MiB to 32 GB)
- Custom container port
- Environment variables
- Secret Manager integration

### Scaling

- Minimum instances (including scale-to-zero)
- Maximum instance count
- Concurrent request handling

### Networking

- Public or private ingress
- VPC access for private resources
- Custom DNS domain mapping
- Automatic SSL certificate provisioning

### Security

- IAM-based access control
- Service account configuration
- Secret Manager for sensitive data
- Network isolation options

## Examples

For comprehensive examples, see [examples.md](./examples.md):

- Minimal configuration
- Production service with secrets
- Custom DNS domain mapping
- Private VPC service
- High-traffic configuration
- Multi-region deployment

## Troubleshooting

### Service fails to deploy

1. Verify Cloud Run API is enabled in your GCP project
2. Check IAM permissions for the service account
3. Ensure container image is accessible
4. Review Cloud Run logs in GCP Console

### Domain mapping issues

1. Verify DNS managed zone exists
2. Check domain ownership verification
3. Ensure TXT record is created
4. Allow time for DNS propagation (up to 48 hours)

### VPC access problems

1. Verify VPC network and subnet exist
2. Check VPC connector or Direct VPC Egress configuration
3. Review firewall rules
4. Ensure service account has necessary permissions

## Resources Created

This module creates the following resources:

- `google_cloud_run_v2_service` - Cloud Run service
- `google_cloud_run_v2_service_iam_member` - IAM policy (if allow_unauthenticated is true)
- `google_cloud_run_domain_mapping` - Domain mapping (if DNS is enabled)
- `google_dns_record_set` - TXT verification record (if DNS is enabled)

## Additional Resources

- [GCP Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Terraform Google Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Cloud Run Pricing](https://cloud.google.com/run/pricing)
- [Cloud Run Best Practices](https://cloud.google.com/run/docs/tips)

## Version Compatibility

- Terraform: >= 1.0
- Google Provider: 6.19.0
- Cloud Run API: v2

## License

Part of the OpenMCF platform.

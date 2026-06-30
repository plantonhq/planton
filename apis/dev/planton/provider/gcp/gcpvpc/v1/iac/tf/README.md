# GCP VPC - Terraform Module

## Overview

This directory contains the Terraform implementation for deploying GCP VPC networks using Planton's `GcpVpc` API. The module creates production-ready Virtual Private Cloud networks with best-practice defaults using HashiCorp Terraform.

## Prerequisites

Before deploying, ensure you have:

1. **Terraform** installed (version 1.5.0 or later)
   ```bash
   terraform version  # Should show 1.5.0+
   ```

2. **GCP Project** with billing enabled
   - You need a GCP project where the VPC will be created
   - Ensure you have Owner or Editor role on the project

3. **GCP Credentials** configured
   ```bash
   gcloud auth application-default login
   gcloud config set project <your-project-id>
   ```

4. **Compute Engine API** enabled (the module will enable it automatically, but you can do it manually):
   ```bash
   gcloud services enable compute.googleapis.com --project=<your-project>
   ```

## Directory Structure

```
iac/tf/
├── main.tf        # VPC resource and Compute API enablement
├── variables.tf   # Input variable definitions
├── outputs.tf     # Output value definitions
├── locals.tf      # Local values (labels, routing mode mapping)
├── provider.tf    # Provider configuration
├── backend.tf     # State backend configuration
└── README.md      # This file
```

## Quick Start

### 1. Configure Backend

The module uses S3-compatible backend for state storage. Create a `backend.tfvars` file:

```hcl
bucket = "my-terraform-state-bucket"
key    = "gcp/vpc/dev/terraform.tfstate"
region = "us-west-2"
```

Or use Google Cloud Storage:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state-bucket"
    prefix = "gcp/vpc/dev"
  }
}
```

### 2. Create Variables File

Create a `terraform.tfvars` file with your VPC specification:

```hcl
metadata = {
  name = "my-vpc"
  id   = "gcpvpc-12345"
  org  = "my-org"
  env  = "dev"
}

spec = {
  project_id = {
    value = "my-gcp-project-123"
  }
  auto_create_subnetworks = false
  routing_mode            = 0  # 0=REGIONAL, 1=GLOBAL
}
```

### 3. Initialize Terraform

```bash
cd iac/tf
terraform init -backend-config=backend.tfvars
```

### 4. Plan Deployment

```bash
terraform plan -var-file=terraform.tfvars
```

Review the execution plan to verify what will be created.

### 5. Apply Changes

```bash
terraform apply -var-file=terraform.tfvars
```

Type `yes` when prompted to confirm the deployment.

### 6. View Outputs

After deployment, view the VPC self-link:

```bash
terraform output network_self_link
```

Example output:
```
projects/my-gcp-project-123/global/networks/my-vpc
```

## Module Structure

### Resources

The module creates the following GCP resources:

1. **`google_project_service.compute_api`**: Enables the Compute Engine API in the target project
2. **`google_compute_network.vpc`**: Creates the VPC network with specified configuration

### Variables

#### `metadata` (required)

Metadata for the GCP VPC resource.

```hcl
variable "metadata" {
  description = "Metadata for the GCP VPC resource"
  type = object({
    name = string           # VPC name (must be unique in project)
    id   = string           # Resource ID for tracking
    org  = optional(string) # Organization identifier
    env  = optional(string) # Environment (dev, staging, prod)
  })
}
```

**Example:**
```hcl
metadata = {
  name = "prod-network"
  id   = "gcpvpc-prod-001"
  org  = "platform-team"
  env  = "production"
}
```

#### `spec` (required)

Specification for the GCP VPC.

```hcl
variable "spec" {
  description = "Specification for the GCP VPC"
  type = object({
    project_id = object({
      value = string        # GCP project ID
    })
    auto_create_subnetworks = optional(bool, false)   # Auto-create subnets?
    routing_mode            = optional(number, 0)     # 0=REGIONAL, 1=GLOBAL
  })
}
```

**Example:**
```hcl
spec = {
  project_id = {
    value = "my-prod-project-123"
  }
  auto_create_subnetworks = false
  routing_mode            = 1  # GLOBAL routing for multi-region
}
```

### Outputs

#### `network_self_link`

The full self-link URL of the created VPC network.

```hcl
output "network_self_link" {
  description = "The full self-link URL of the created VPC network"
  value       = google_compute_network.vpc.self_link
}
```

**Format:** `projects/<project>/global/networks/<name>`

**Usage:** Reference this output in other Terraform modules to create subnets, firewall rules, or attach GKE clusters.

### Locals

The module computes local values for:

1. **GCP Labels**: Converts metadata to GCP label format
   ```hcl
   gcp_labels = {
     resource     = var.metadata.name
     resource-id  = var.metadata.id
     resource-org = var.metadata.org
     env          = var.metadata.env
   }
   ```

2. **Routing Mode**: Maps proto enum (0=REGIONAL, 1=GLOBAL) to GCP API strings
   ```hcl
   routing_mode = lookup(local.routing_mode_map, var.spec.routing_mode, "REGIONAL")
   ```

## Deployment Scenarios

### Development Environment

**`dev.tfvars`:**
```hcl
metadata = {
  name = "dev-vpc"
  id   = "gcpvpc-dev-001"
  env  = "development"
}

spec = {
  project_id = {
    value = "dev-project-123"
  }
  auto_create_subnetworks = false
}
```

Deploy:
```bash
terraform apply -var-file=dev.tfvars
```

### Production with Global Routing

**`prod.tfvars`:**
```hcl
metadata = {
  name = "prod-vpc"
  id   = "gcpvpc-prod-001"
  org  = "platform-team"
  env  = "production"
}

spec = {
  project_id = {
    value = "prod-project-456"
  }
  auto_create_subnetworks = false
  routing_mode            = 1  # GLOBAL
}
```

Deploy:
```bash
terraform apply -var-file=prod.tfvars
```

### Shared VPC Host Network

**`shared.tfvars`:**
```hcl
metadata = {
  name = "shared-vpc-host"
  id   = "gcpvpc-shared-001"
  org  = "platform-team"
  env  = "shared"
}

spec = {
  project_id = {
    value = "network-host-project-123"
  }
  auto_create_subnetworks = false
  routing_mode            = 1  # GLOBAL for multi-region service projects
}
```

Deploy:
```bash
terraform apply -var-file=shared.tfvars
```

**Post-deployment:** Enable Shared VPC and attach service projects:
```bash
gcloud compute shared-vpc enable network-host-project-123
gcloud compute shared-vpc associated-projects add service-project-1 \
  --host-project=network-host-project-123
```

## Using Outputs in Other Modules

### Option 1: Terraform Outputs

Reference the output in the same Terraform workspace:

```hcl
# In another module
data "terraform_remote_state" "vpc" {
  backend = "gcs"
  config = {
    bucket = "my-terraform-state-bucket"
    prefix = "gcp/vpc/prod"
  }
}

resource "google_compute_subnetwork" "subnet" {
  network = data.terraform_remote_state.vpc.outputs.network_self_link
  # ...
}
```

### Option 2: Direct Reference

If both resources are in the same Terraform configuration:

```hcl
module "vpc" {
  source = "./path/to/vpc/module"
  # ...
}

resource "google_compute_subnetwork" "subnet" {
  network = module.vpc.network_self_link
  # ...
}
```

## Best Practices

### 1. Use Remote State

**Don't store state locally** (`.tfstate` files in the working directory). Use a remote backend:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state-bucket"
    prefix = "gcp/vpc/prod"
  }
}
```

**Benefits:**
- State is shared across team members
- State is versioned and backed up
- State locking prevents concurrent modifications

### 2. Separate State Per Environment

Use different state files for each environment:

```
gs://my-state-bucket/gcp/vpc/dev/
gs://my-state-bucket/gcp/vpc/staging/
gs://my-state-bucket/gcp/vpc/prod/
```

This prevents accidentally modifying production while working on dev.

### 3. Use Workspaces for Environment Isolation

Alternatively, use Terraform workspaces:

```bash
terraform workspace new dev
terraform workspace new staging
terraform workspace new prod

terraform workspace select dev
terraform apply -var-file=dev.tfvars
```

### 4. Version Lock Providers

Pin provider versions in `provider.tf`:

```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.19.0"  # Pin to specific version
    }
  }
}
```

This prevents unexpected updates from breaking your configuration.

### 5. Plan Before Apply

**Always run `terraform plan`** before `terraform apply`:

```bash
terraform plan -var-file=prod.tfvars -out=plan.tfplan
terraform apply plan.tfplan
```

This ensures you review changes before execution.

### 6. Use Labels for Cost Tracking

Always set `metadata.org` and `metadata.env`:

```hcl
metadata = {
  name = "prod-vpc"
  org  = "platform-team"
  env  = "production"
}
```

These are automatically converted to GCP labels by the module for cost tracking and resource organization.

## Advanced Usage

### Importing Existing VPCs

If a VPC already exists in GCP, import it into Terraform state:

```bash
terraform import google_compute_network.vpc projects/<project>/global/networks/<vpc-name>
```

Then run `terraform plan` to see if any configuration drift exists.

### Targeting Specific Resources

To apply changes to only the VPC (not the API enablement):

```bash
terraform apply -target=google_compute_network.vpc -var-file=prod.tfvars
```

**Warning:** Use targeting sparingly—it can lead to unexpected state issues.

### State Management Commands

**List resources in state:**
```bash
terraform state list
```

**Show resource details:**
```bash
terraform state show google_compute_network.vpc
```

**Move resources:**
```bash
terraform state mv google_compute_network.vpc module.vpc.google_compute_network.vpc
```

**Remove resources:**
```bash
terraform state rm google_compute_network.vpc
```

## Troubleshooting

### "API compute.googleapis.com has not been used"

**Cause**: Compute API is not enabled in the target project.

**Solution**: The module automatically enables it via `google_project_service.compute_api`. If this fails:
```bash
gcloud services enable compute.googleapis.com --project=<your-project>
```

### "Error creating Network: googleapi: Error 409: The resource already exists"

**Cause**: A VPC with the same name already exists in the project.

**Solutions:**

1. **Import existing VPC:**
   ```bash
   terraform import google_compute_network.vpc projects/<project>/global/networks/<name>
   ```

2. **Delete existing VPC:**
   ```bash
   gcloud compute networks delete <vpc-name> --project=<project>
   ```

3. **Change VPC name:**
   Update `metadata.name` in your `.tfvars` file.

### "Error locking state"

**Cause**: Another Terraform process is holding a state lock.

**Solutions:**

1. **Wait for other process to complete** (locks are released automatically)

2. **Force unlock** (only if you're certain no other process is running):
   ```bash
   terraform force-unlock <lock-id>
   ```

### "Permission denied"

**Cause**: GCP credentials lack sufficient permissions.

**Solution**: Ensure your service account or user has the required IAM roles:
```bash
gcloud projects add-iam-policy-binding <project-id> \
  --member=user:<your-email> \
  --role=roles/compute.networkAdmin

gcloud projects add-iam-policy-binding <project-id> \
  --member=user:<your-email> \
  --role=roles/serviceusage.serviceUsageAdmin
```

### "Invalid routing_mode"

**Cause**: `routing_mode` must be 0 (REGIONAL) or 1 (GLOBAL).

**Solution**: Check your `.tfvars` file:
```hcl
spec = {
  routing_mode = 0  # or 1, not "REGIONAL" or "GLOBAL"
}
```

## Validation

### Validate Configuration

Before applying, validate your Terraform configuration:

```bash
terraform validate
```

This checks for syntax errors and invalid configurations.

### Format Configuration

Ensure consistent formatting:

```bash
terraform fmt -recursive
```

### Lint Configuration

Use `tflint` for additional validation:

```bash
tflint --init
tflint
```

## CI/CD Integration

### Example GitHub Actions Workflow

```yaml
name: Terraform Deploy

on:
  push:
    branches: [main]
    paths:
      - 'iac/tf/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
      
      - name: Authenticate to GCP
        uses: google-github-actions/auth@v1
        with:
          credentials_json: ${{ secrets.GCP_CREDENTIALS }}
      
      - name: Terraform Init
        run: terraform init
        working-directory: iac/tf
      
      - name: Terraform Plan
        run: terraform plan -var-file=prod.tfvars
        working-directory: iac/tf
      
      - name: Terraform Apply
        if: github.ref == 'refs/heads/main'
        run: terraform apply -auto-approve -var-file=prod.tfvars
        working-directory: iac/tf
```

### Example GitLab CI Pipeline

```yaml
stages:
  - validate
  - deploy

terraform-validate:
  stage: validate
  image: hashicorp/terraform:1.5.0
  script:
    - cd iac/tf
    - terraform init -backend=false
    - terraform validate

terraform-deploy:
  stage: deploy
  image: hashicorp/terraform:1.5.0
  script:
    - cd iac/tf
    - terraform init
    - terraform plan -var-file=prod.tfvars
    - terraform apply -auto-approve -var-file=prod.tfvars
  only:
    - main
```

## Related Resources

After deploying the VPC, you typically need:

1. **Subnets**: Create via `google_compute_subnetwork` resources
2. **Firewall Rules**: Manage via `google_compute_firewall` resources
3. **Cloud NAT**: For private instance outbound connectivity
4. **Cloud Router**: For dynamic routing (VPN, Interconnect)

## Further Reading

- [Terraform Google Provider Documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [GCP VPC Documentation](https://cloud.google.com/vpc/docs)
- [Planton Component API Reference](../../README.md)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)

## Support

For issues or questions:
1. Check [troubleshooting section](#troubleshooting)
2. Review [examples](../../examples.md)
3. Consult [research documentation](../../docs/README.md)
4. Open an issue in the Planton repository


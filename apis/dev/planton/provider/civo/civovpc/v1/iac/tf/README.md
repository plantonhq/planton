# CivoVpc Terraform Module

This directory contains the Terraform implementation for provisioning Civo VPCs (private networks) using HCL.

## Overview

The Terraform module creates isolated private networks on Civo cloud with the following capabilities:

- **Network Creation**: Provisions Civo networks with configurable CIDR blocks
- **Auto-Allocation**: Supports CIDR auto-allocation when not explicitly specified
- **Regional Deployment**: Creates networks in specific Civo regions (LON1, NYC1, FRA1, etc.)
- **Default Network Control**: Can set a network as the region's default

## Module Structure

```
tf/
├── provider.tf       # Provider configuration and version constraints
├── variables.tf      # Input variable definitions
├── locals.tf         # Local values and label generation
├── main.tf           # Main resource definitions
└── outputs.tf        # Output value definitions
```

## Files Description

### provider.tf

Defines the Civo provider and version constraints:

```hcl
terraform {
  required_providers {
    civo = {
      source  = "civo/civo"
      version = ">= 1.0"
    }
  }
}

provider "civo" {
  region = var.spec.region
}
```

**Authentication:** Civo provider authenticates via `CIVO_TOKEN` environment variable.

### variables.tf

Defines input variables matching the `CivoVpcSpec` protobuf:

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, id, org, env, labels) |
| `spec` | object | CivoVpc specification (network_name, region, cidr, etc.) |

### locals.tf

Extracts values from variables and builds label map:

```hcl
locals {
  network_name = var.spec.network_name
  region       = var.spec.region
  cidr_block   = var.spec.ip_range_cidr
  is_default   = var.spec.is_default_for_region
  description  = var.spec.description
  
  planton_labels = {
    "planton.org/resource"      = "true"
    "planton.org/resource-kind" = "CivoVpc"
    "planton.org/resource-id"   = var.metadata.id
    "planton.org/resource-name" = var.metadata.name
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
  }
}
```

**Note:** Civo Network resources don't currently support labels/tags. These are stored in Planton metadata.

### main.tf

Creates the Civo network resource:

```hcl
resource "civo_network" "main" {
  label   = local.network_name
  region  = local.region
  cidr_v4 = local.cidr_block != "" ? local.cidr_block : null
  default = local.is_default
}
```

**CIDR Handling:**
- If `local.cidr_block` is empty, `cidr_v4 = null` → Civo auto-allocates
- If `local.cidr_block` is specified, uses the explicit CIDR

**Default Network:**
- `default = true` sets this network as the region's default
- Only one default network per region is allowed

### outputs.tf

Exports resource attributes:

```hcl
output "network_id" {
  description = "The unique identifier (ID) of the created Civo network"
  value       = civo_network.main.id
}

output "cidr_block" {
  description = "The IPv4 CIDR block of the created network"
  value       = civo_network.main.cidr_v4
}

output "created_at_rfc3339" {
  description = "Timestamp when the network was created (RFC 3339 format)"
  value       = ""
  # Note: Civo provider doesn't expose created_at timestamp
}
```

## Usage

### Prerequisites

1. **Terraform CLI** installed (v1.0+)
2. **Civo API token** exported as environment variable:
   ```bash
   export CIVO_TOKEN="your-civo-api-token"
   ```
3. Valid `CivoVpcSpec` configuration

### Basic Workflow

#### 1. Initialize Terraform

```bash
cd iac/tf/
terraform init
```

This downloads the Civo provider and initializes the working directory.

#### 2. Create Variables File

Create `terraform.tfvars` with your configuration:

```hcl
metadata = {
  name = "prod-main-network"
  id   = "civpc-abc123"
  org  = "myorg"
  env  = "production"
}

spec = {
  civo_credential_id    = "civo-cred-123"
  network_name          = "prod-main-network"
  region                = "NYC1"
  ip_range_cidr         = "10.20.1.0/24"
  is_default_for_region = false
  description           = "Production network (NYC1)"
}
```

#### 3. Plan Changes

```bash
terraform plan
```

**Output:**
```
Terraform will perform the following actions:

  # civo_network.main will be created
  + resource "civo_network" "main" {
      + cidr_v4 = "10.20.1.0/24"
      + default = false
      + id      = (known after apply)
      + label   = "prod-main-network"
      + region  = "NYC1"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

**Review Carefully:**
- Verify CIDR block is correct (immutable after creation)
- Check region (immutable after creation)
- Confirm network name

#### 4. Apply Changes

```bash
terraform apply
```

Type `yes` to confirm.

**Output:**
```
civo_network.main: Creating...
civo_network.main: Creation complete after 3s [id=abc123-def456]

Outputs:

network_id = "abc123-def456"
cidr_block = "10.20.1.0/24"
```

#### 5. Verify in Civo

```bash
civo network show abc123-def456
```

### Advanced Usage

#### Auto-Allocate CIDR (Dev/Test)

```hcl
spec = {
  network_name = "dev-test-network"
  region       = "LON1"
  ip_range_cidr = ""  # Empty → Civo auto-allocates
}
```

#### Set as Default Network

```hcl
spec = {
  network_name          = "default-network"
  region                = "FRA1"
  ip_range_cidr         = "10.30.1.0/24"
  is_default_for_region = true  # Makes this the default for FRA1
}
```

**Constraint:** Only one default network per region.

#### Multi-Region Deployment

Use Terraform workspaces or separate state files:

```bash
# Create production network in London
terraform workspace new prod-lon1
terraform apply -var-file=prod-lon1.tfvars

# Create production network in New York
terraform workspace new prod-nyc1
terraform apply -var-file=prod-nyc1.tfvars
```

## State Management

### Local State

By default, Terraform stores state in `terraform.tfstate` (local file).

**Not recommended for teams** (no collaboration, no locking).

### Remote State (Recommended)

Use a remote backend for production:

#### Option 1: Civo Object Store (S3-Compatible)

```hcl
terraform {
  backend "s3" {
    bucket                      = "my-terraform-state"
    key                         = "civovpc/prod-nyc1/terraform.tfstate"
    region                      = "us-east-1"
    endpoint                    = "https://objectstore.lon1.civo.com"
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_metadata_api_check     = true
  }
}
```

**Benefits:**
- Fully self-contained (all Civo resources and state in Civo)
- Cost-effective
- No external dependencies

#### Option 2: Terraform Cloud

```hcl
terraform {
  backend "remote" {
    organization = "myorg"
    workspaces {
      name = "civovpc-prod-nyc1"
    }
  }
}
```

**Benefits:**
- Built-in state locking
- Team collaboration
- UI for plan/apply operations

## Outputs and Dependent Resources

### Using Outputs in Other Modules

The `network_id` output can be referenced by dependent resources:

```hcl
# In another module (e.g., Kubernetes cluster)
module "civovpc" {
  source = "../civovpc/iac/tf"
  # ... configuration ...
}

resource "civo_kubernetes_cluster" "main" {
  name       = "prod-cluster"
  region     = "NYC1"
  network_id = module.civovpc.network_id  # Reference output
}
```

Terraform automatically determines dependency order:
1. Create network first
2. Create cluster second (after network exists)

### Querying Outputs

After `terraform apply`:

```bash
# Get all outputs
terraform output

# Get specific output
terraform output network_id
terraform output -json cidr_block
```

## Lifecycle Management

### Update Network

**Immutable Attributes:**
- `region`: Cannot be changed → Forces replacement
- `cidr_v4`: Cannot be changed → Forces replacement

**Mutable Attributes:**
- `label`: Can be renamed in-place
- `default`: Can be toggled (if no other default exists)

**Example - Changing CIDR triggers replacement:**

```bash
# Change CIDR in tfvars
# ip_range_cidr = "10.20.2.0/24"  # Changed from 10.20.1.0/24

terraform plan
```

**Output:**
```
# civo_network.main must be replaced
-/+ resource "civo_network" "main" {
      ~ cidr_v4 = "10.20.1.0/24" -> "10.20.2.0/24" # forces replacement
      ~ id      = "old-id" -> (known after apply)
        label   = "prod-main-network"
        region  = "NYC1"
    }

Plan: 1 to add, 0 to change, 1 to destroy.
```

**Warning:** Replacement is **destructive**. All resources (clusters, instances) attached to the network will be affected.

**Best Practice:** Always run `terraform plan` before applying network changes.

### Destroy Network

```bash
terraform destroy
```

**Prerequisites:**
- All dependent resources (clusters, instances, firewalls) must be destroyed first
- Terraform will error if dependencies exist

**Safe Destruction:**
```bash
# 1. Destroy clusters/instances first
terraform destroy -target=civo_kubernetes_cluster.main

# 2. Then destroy network
terraform destroy -target=civo_network.main
```

## Provider Limitations

### 1. Description Field Not Supported

The Civo Terraform provider's `civo_network` resource doesn't expose a `description` attribute.

**Workaround:**
- Description is stored in Planton metadata
- Use resource comments in Terraform code for documentation

```hcl
# Production network for NYC1 region
# Purpose: Host production Kubernetes clusters
# CIDR: 10.20.1.0/24 (following hierarchical schema)
resource "civo_network" "main" {
  label = "prod-main-network"
  # ...
}
```

### 2. No Labels/Tags Support

Civo Network resources don't support labels or tags via the provider.

**Workaround:**
- Labels defined in `locals.planton_labels` are stored in Planton metadata
- Use naming conventions in `label` field (e.g., `prod-nyc1-network`)

### 3. Created Timestamp Not Exposed

The provider doesn't expose the network's `created_at` timestamp.

**Impact:**
- `created_at_rfc3339` output is defined but always empty
- Included for API consistency with Pulumi module

## Troubleshooting

### Issue: "Error: CIDR block already in use"

**Cause:** Specified CIDR overlaps with an existing network in the same region.

**Solution:**
1. Check existing networks:
   ```bash
   civo network list --region NYC1
   ```
2. Choose a different CIDR or use auto-allocation (omit `ip_range_cidr`)

### Issue: "Error: Only one default network allowed per region"

**Cause:** Trying to set `is_default_for_region = true` when another default exists.

**Solution:**
1. List networks to find current default:
   ```bash
   civo network list --region NYC1
   ```
2. Set other network's `default = false` first, or accept this network as non-default

### Issue: "Error: Provider authentication failed"

**Cause:** Missing or invalid `CIVO_TOKEN` environment variable.

**Solution:**
```bash
export CIVO_TOKEN="your-civo-api-token"
terraform plan
```

### Issue: Terraform wants to replace network unexpectedly

**Cause:** Changed an immutable attribute (`region` or `cidr_v4`).

**Solution:**
1. Review `terraform plan` carefully
2. If replacement is intentional:
   - Back up all data in dependent resources
   - Plan for downtime
   - Run `terraform apply`
3. If unintentional:
   - Revert changes to `region` or `cidr_v4`
   - Run `terraform plan` again to verify no replacement

## Best Practices

### 1. Always Use Remote State

- **Never** commit `terraform.tfstate` to version control
- Use remote backends (Civo Object Store, Terraform Cloud, S3)
- Enable state locking to prevent concurrent modifications

### 2. Version Your Configuration

- Keep `.tf` files in version control (Git)
- Use tags or branches for different environments
- Review all changes via pull requests

### 3. Plan Before Apply

```bash
# Always run plan first
terraform plan -out=tfplan

# Review plan
cat tfplan  # or use `terraform show tfplan`

# Apply only if plan looks correct
terraform apply tfplan
```

### 4. Use tfvars for Environments

Create separate `.tfvars` files for each environment:

```
├── terraform.tfvars         # Default (dev)
├── prod-lon1.tfvars        # Production London
├── prod-nyc1.tfvars        # Production New York
└── stage-nyc1.tfvars       # Staging New York
```

Apply with:
```bash
terraform apply -var-file=prod-nyc1.tfvars
```

### 5. Document CIDR Allocation

Keep a CIDR allocation table in your repository:

```markdown
# CIDR Allocation

| Network | CIDR | Region | Environment |
|---------|------|--------|-------------|
| prod-lon1-network | 10.10.1.0/24 | LON1 | Production |
| prod-nyc1-network | 10.20.1.0/24 | NYC1 | Production |
| stage-nyc1-network | 10.20.2.0/24 | NYC1 | Staging |
```

## Integration with Planton

This Terraform module can be used independently or as part of Planton's orchestration:

1. **Standalone Use**: Run `terraform plan/apply` directly
2. **Planton Integration**: Planton generates `.tfvars` from protobuf spec and invokes Terraform

**Pulumi vs. Terraform:**
- Planton's primary implementation uses Pulumi (Go)
- This Terraform module is provided for teams that prefer Terraform
- Both implementations produce identical Civo networks

## Further Reading

- **User Guide**: [../../README.md](../../README.md) - CivoVpc resource documentation
- **Examples**: [../../examples.md](../../examples.md) - Configuration examples
- **Research**: [../../docs/README.md](../../docs/README.md) - Comprehensive deployment guide
- **Terraform Civo Provider**: [GitHub](https://github.com/civo/terraform-provider-civo)
- **Civo API**: [civo.com/api](https://www.civo.com/api)

---

**Maintained by:** Planton  
**Last Updated:** 2025-11-21


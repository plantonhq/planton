# Cloudflare D1 Database - Terraform Module

Terraform module for provisioning Cloudflare D1 databases using HCL.

## Overview

This module implements the CloudflareD1Database resource using Terraform's HCL syntax and the official Cloudflare provider (v4+). It translates the protobuf-defined `CloudflareD1DatabaseSpec` into Terraform resource configuration.

## Module Structure

```
iac/tf/
├── provider.tf    # Terraform and Cloudflare provider configuration
├── variables.tf   # Input variables (metadata, spec)
├── locals.tf      # Local values and label construction
├── main.tf        # D1 database resource definition
└── outputs.tf     # Module outputs
```

## Inputs

### Variables

The module accepts two input variables:

#### `metadata` (object)

Metadata for the resource, including name and labels:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Resource name |
| `id` | string | No | Resource ID (defaults to name if empty) |
| `org` | string | No | Organization label |
| `env` | string | No | Environment label |
| `labels` | map(string) | No | Additional labels |
| `tags` | list(string) | No | Tags |
| `version` | object | No | Version info (id, message) |

#### `spec` (object)

CloudflareD1DatabaseSpec configuration:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | Cloudflare account ID |
| `database_name` | string | Yes | Database name (max 64 chars) |
| `region` | string | No | Primary location hint (weur, eeur, apac, oc, wnam, enam) |
| `read_replication` | object | No | Read replication config (mode: "auto" or "disabled") |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `database_id` | string | The unique identifier of the created D1 database |
| `database_name` | string | The name of the database (same as input) |
| `connection_string` | string | Connection string (currently empty - D1 uses Worker bindings) |

## Usage

### Via Planton CLI

The typical usage is through the Planton CLI, which handles variable passing and Terraform execution:

```bash
planton apply -f database.yaml
```

### Direct Terraform Execution

For manual execution or debugging:

1. **Set Environment Variables**:
   ```bash
   export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"
   ```

2. **Create Variable Values File** (`terraform.tfvars`):
   ```hcl
   metadata = {
     name = "my-app-prod-db"
     org  = "my-org"
     env  = "production"
   }

   spec = {
     account_id    = "abc123def456..."
     database_name = "my-app-production-db"
     region        = "enam"
     read_replication = {
       mode = "auto"
     }
   }
   ```

3. **Run Terraform**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Resource Mapping

The module maps the spec to Terraform's `cloudflare_d1_database` resource:

| Spec Field | Terraform Argument | Notes |
|------------|-------------------|-------|
| `account_id` | `account_id` | Required |
| `database_name` | `name` | Required |
| `region` | `primary_location_hint` | Optional string |
| `read_replication.mode` | `read_replication.mode` | Optional, dynamic block |

### Region Values

The `region` field accepts the following string values:
- `"weur"` - Western Europe
- `"eeur"` - Eastern Europe
- `"apac"` - Asia Pacific
- `"oc"` - Oceania
- `"wnam"` - Western North America
- `"enam"` - Eastern North America

If omitted, Cloudflare selects a default location.

### Read Replication

The `read_replication` configuration uses a dynamic block:

```hcl
dynamic "read_replication" {
  for_each = var.spec.read_replication != null ? [var.spec.read_replication] : []
  content {
    mode = read_replication.value.mode
  }
}
```

This ensures the `read_replication` block is only added if `spec.read_replication` is non-null.

## Examples

### Minimal Configuration

```hcl
metadata = {
  name = "minimal-db"
}

spec = {
  account_id    = "abc123..."
  database_name = "my-minimal-db"
}
```

### With Region

```hcl
metadata = {
  name = "regional-db"
  env  = "production"
}

spec = {
  account_id    = "abc123..."
  database_name = "my-app-production-db"
  region        = "enam"
}
```

### With Read Replication

```hcl
metadata = {
  name = "global-db"
  env  = "production"
}

spec = {
  account_id    = "abc123..."
  database_name = "my-app-global-db"
  region        = "weur"
  read_replication = {
    mode = "auto"
  }
}
```

## Labels

The module automatically constructs labels for the D1 database resource:

### Base Labels (Always Applied)
- `resource` = `"true"`
- `resource_id` = metadata.id (or metadata.name if id is empty)
- `resource_kind` = `"cloudflare_d1_database"`

### Conditional Labels (Applied if Specified)
- `organization` = metadata.org (if non-empty)
- `environment` = metadata.env (if non-empty)

These labels are merged and applied via `locals.final_labels`.

## State Management

Terraform tracks the D1 database resource state in the configured backend:

- **Local State**: `terraform.tfstate` (default)
- **Remote State**: S3, Azure Blob, Terraform Cloud, etc.

For production, use a remote backend:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "cloudflare-d1/my-app-prod-db.tfstate"
    region = "us-east-1"
  }
}
```

## Provider Configuration

The module uses the official Cloudflare provider:

```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  # Automatically uses CLOUDFLARE_API_TOKEN environment variable
}
```

### Authentication

Set the `CLOUDFLARE_API_TOKEN` environment variable:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
```

Alternatively, configure the provider explicitly:

```hcl
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
```

## Validation

Terraform performs the following validations:

### Built-in Validations
- Required fields: `account_id`, `database_name`
- Type checks: Ensures fields match expected types

### Cloudflare API Validations
- `database_name` must be unique within the account
- `region` must be a valid location hint
- `read_replication.mode` must be "auto" or "disabled"

If validation fails, Terraform reports an error before applying changes.

## Limitations

### What This Module Does

- ✅ Provisions the D1 database resource (the "container")
- ✅ Configures region and read replication
- ✅ Exports database ID and name

### What This Module Does NOT Do

- ❌ Create tables or manage schema (use Wrangler CLI migrations)
- ❌ Configure Worker bindings (use `cloudflare_workers_script` resource or `wrangler.toml`)
- ❌ Manage data migrations or seed data

## Troubleshooting

### "Error: Database already exists"

**Cause**: A database with the specified `database_name` already exists in the account.

**Fix**: Choose a unique name or import the existing database into Terraform state:

```bash
terraform import cloudflare_d1_database.main <database_id>
```

### "Error: Invalid primary_location_hint"

**Cause**: The `region` value is not a valid Cloudflare location hint.

**Fix**: Use one of: `weur`, `eeur`, `apac`, `oc`, `wnam`, `enam`.

### Empty Connection String

**Expected Behavior**: D1 does not use traditional connection strings. Worker bindings are configured separately via `wrangler.toml` or the `cloudflare_workers_script` resource.

## Schema Management

**Important**: This module provisions the database **resource** (the container). It does **not** create tables or manage schema.

Schema management is handled exclusively via Cloudflare's Wrangler CLI:

```bash
# Create a migration
npx wrangler d1 migrations create my-app-db create_users_table

# Apply migrations
npx wrangler d1 migrations apply my-app-db --remote
```

This architectural separation is by design. See [../../docs/README.md](../../docs/README.md) for detailed explanation of the "Orchestration Gap."

## Multi-Environment Pattern

Use Terraform workspaces or separate state files for multi-environment deployments:

### Option 1: Workspaces

```bash
# Create workspaces
terraform workspace new dev
terraform workspace new preview
terraform workspace new prod

# Deploy to dev
terraform workspace select dev
terraform apply -var-file=dev.tfvars

# Deploy to prod
terraform workspace select prod
terraform apply -var-file=prod.tfvars
```

### Option 2: Separate Directories

```
environments/
├── dev/
│   ├── main.tf → symlink to ../../iac/tf/main.tf
│   ├── variables.tf → symlink to ../../iac/tf/variables.tf
│   └── terraform.tfvars
├── preview/
│   └── ...
└── prod/
    └── ...
```

Deploy each environment independently:

```bash
cd environments/dev && terraform apply
cd environments/prod && terraform apply
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy Cloudflare D1 Database

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2

      - name: Terraform Init
        working-directory: ./iac/tf
        run: terraform init

      - name: Terraform Apply
        working-directory: ./iac/tf
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        run: terraform apply -auto-approve -var-file=prod.tfvars

      - name: Apply Migrations
        run: npx wrangler d1 migrations apply my-app-production-db --remote
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
```

## Testing

### Plan Validation

Validate configuration without applying:

```bash
terraform plan
```

### Dry Run

Preview changes before applying:

```bash
terraform plan -out=tfplan
terraform show tfplan
```

### Cleanup

Destroy resources:

```bash
terraform destroy
```

## Further Reading

- **Cloudflare Provider Docs**: [registry.terraform.io/providers/cloudflare/cloudflare](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- **D1 Resource Docs**: [registry.terraform.io/.../resources/d1_database](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/d1_database)
- **Terraform Best Practices**: [developer.hashicorp.com/terraform/tutorials](https://developer.hashicorp.com/terraform/tutorials)
- **Architecture Overview**: [../../docs/README.md](../../docs/README.md)

## Comparison to Pulumi

| Aspect | Terraform (HCL) | Pulumi (Go) |
|--------|----------------|-------------|
| **Language** | HCL (declarative DSL) | Go (strongly typed) |
| **Conditionals** | `dynamic` blocks, `count`, `for_each` | Native Go `if` statements |
| **Type Safety** | Runtime type checks | Compile-time type checks |
| **State** | Terraform state file | Pulumi state file |
| **Community** | Larger (industry standard) | Smaller (growing) |

**Key Insight**: Both produce identical infrastructure. Choose based on team preference.

## Support

For issues specific to this Terraform module:
1. Validate configuration: `terraform validate`
2. Check plan output: `terraform plan`
3. Review Cloudflare provider docs

For general Cloudflare D1 questions, see [../../README.md](../../README.md).


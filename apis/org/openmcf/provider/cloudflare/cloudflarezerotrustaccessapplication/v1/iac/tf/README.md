# Terraform Implementation for Cloudflare Zero Trust Access Application

This directory contains the Terraform-based Infrastructure-as-Code (IaC) implementation for provisioning Cloudflare Zero Trust Access Applications.

## Overview

The Terraform implementation creates:
1. **Cloudflare Access Application**: The protected application resource with hostname and session configuration
2. **Cloudflare Access Policy**: The policy defining who can access the application (email-based or group-based rules)

The implementation uses the official [Cloudflare Terraform provider](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs).

## Prerequisites

1. **Terraform**: Version 1.0 or higher - Install from [terraform.io](https://www.terraform.io/downloads)
2. **Cloudflare Account**: Active Cloudflare account with Zero Trust Access enabled
3. **Cloudflare API Token**: API token with Zero Trust permissions
4. **Cloudflare DNS Zone**: DNS zone for the domain you want to protect
5. **Identity Provider**: Configured identity provider (Google Workspace, Okta, Azure AD, etc.) in Cloudflare Zero Trust dashboard

## Directory Structure

```
iac/tf/
├── README.md       # This file
├── main.tf         # Access Application and Policy resources
├── provider.tf     # Terraform and provider configuration
├── variables.tf    # Input variables
├── locals.tf       # Local computed values
└── outputs.tf      # Output values
```

## Configuration

### Environment Variables

Set these environment variables before running Terraform:

```bash
export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"
```

### Input Variables

The Terraform module expects two input variables:

#### metadata

```hcl
variable "metadata" {
  description = "Metadata for the resource"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}
```

#### spec

```hcl
variable "spec" {
  description = "CloudflareZeroTrustAccessApplicationSpec"
  type = object({
    application_name         = string
    zone_id                  = string
    hostname                 = string
    policy_type              = optional(string, "ALLOW")
    allowed_emails           = optional(list(string), [])
    session_duration_minutes = optional(number, 1440)
    require_mfa              = optional(bool, false)
    allowed_google_groups    = optional(list(string), [])
  })
}
```

## Deployment

### Using OpenMCF CLI (Recommended)

The simplest way to deploy is via OpenMCF CLI:

```bash
# Create a manifest file
cat > access-app.yaml <<EOF
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-app
spec:
  application_name: "My Application"
  zone_id: "your-zone-id"
  hostname: "app.example.com"
  policy_type: ALLOW
  allowed_emails:
    - "@example.com"
  session_duration_minutes: 480
  require_mfa: true
EOF

# Deploy using Terraform backend
planton apply -f access-app.yaml --backend terraform
```

### Using Terraform CLI Directly

For advanced use cases or debugging:

```bash
# Navigate to the terraform directory
cd iac/tf

# Create a terraform.tfvars file
cat > terraform.tfvars <<EOF
metadata = {
  name = "my-app"
  env  = "production"
}

spec = {
  application_name         = "My Application"
  zone_id                  = "your-zone-id"
  hostname                 = "app.example.com"
  policy_type              = "ALLOW"
  allowed_emails           = ["@example.com"]
  session_duration_minutes = 480
  require_mfa              = true
  allowed_google_groups    = ["engineering@example.com"]
}
EOF

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

## Terraform Commands

### Initialize

Download provider plugins and initialize backend:

```bash
terraform init
```

### Plan

Preview changes before applying:

```bash
terraform plan
```

### Apply

Create or update resources:

```bash
terraform apply
```

### Destroy

Delete resources:

```bash
terraform destroy
```

### Output

View output values:

```bash
terraform output application_id
terraform output public_hostname
terraform output policy_id
```

## Outputs

After successful deployment, the following outputs are available:

| Output Name | Description |
|-------------|-------------|
| `application_id` | The unique ID of the Cloudflare Access Application |
| `public_hostname` | The hostname being protected |
| `policy_id` | The ID of the Access Policy |

Access outputs:

```bash
terraform output application_id
terraform output -raw public_hostname
```

## How It Works

### main.tf

The `main.tf` file contains the core resource definitions:

1. **Data Source: Cloudflare Zone**
   ```hcl
   data "cloudflare_zone" "main" {
     zone_id = var.spec.zone_id
   }
   ```
   Looks up the zone to retrieve the account ID (required for Access Policy).

2. **Resource: Access Policy**
   ```hcl
   resource "cloudflare_zero_trust_access_policy" "main" {
     account_id = data.cloudflare_zone.main.account.id
     name       = "default-policy"
     decision   = var.spec.policy_type == "BLOCK" ? "deny" : "allow"

     include = concat(
       [for e in var.spec.allowed_emails : { email = { email = e }, group = null }],
       [for g in var.spec.allowed_google_groups : { email = null, group = { id = g } }],
     )
     require = var.spec.require_mfa ? [{ auth_method = { auth_method = "mfa" } }] : []
   }
   ```
   A policy is a standalone, account-scoped object with Include (who can access) and Require (MFA) rules.

3. **Resource: Access Application**
   ```hcl
   resource "cloudflare_zero_trust_access_application" "main" {
     account_id       = data.cloudflare_zone.main.account.id
     name             = var.spec.application_name
     domain           = var.spec.hostname
     type             = "self_hosted"
     session_duration = var.spec.session_duration_minutes > 0 ? "${var.spec.session_duration_minutes}m" : "24h"

     policies = [{
       id         = cloudflare_zero_trust_access_policy.main.id
       precedence = 1
     }]
   }
   ```
   The application references the policy through its `policies` list (precedence sets evaluation order).

### Rule Construction

Include and Require rules are built with attribute (object) syntax. Each include element carries the
full attribute set (unused keys are `null`) so all elements share one object type:

- Each email in `allowed_emails` becomes `{ email = { email = <value> } }`.
- Each group in `allowed_google_groups` becomes `{ group = { id = <value> } }`.
- When `require_mfa` is `true`, the require list contains `{ auth_method = { auth_method = "mfa" } }`.

## State Management

### Local State (Default)

By default, Terraform stores state locally in `terraform.tfstate`. This is fine for testing but **not recommended for production**.

### Remote State (Recommended)

For production, use a remote backend:

#### Terraform Cloud

```hcl
terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "cloudflare-access"
    }
  }
}
```

#### S3 Backend

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "cloudflare-access/terraform.tfstate"
    region = "us-east-1"
  }
}
```

#### GCS Backend

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "cloudflare-access"
  }
}
```

## Common Issues

### "Zone not found" Error

**Cause**: Invalid or missing `zone_id` in spec.

**Solution**: Verify the zone ID is correct:
```bash
curl -X GET "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN"
```

### "Insufficient permissions" Error

**Cause**: API token doesn't have Zero Trust permissions.

**Solution**: Ensure your API token has the following permissions:
- Account > Zero Trust > Edit
- Zone > DNS > Read

### MFA Not Being Enforced

**Cause**: `require_mfa` is `false`, or identity provider doesn't support MFA prompting.

**Solution**:
1. Set `require_mfa = true` in `terraform.tfvars`
2. Verify your IdP supports MFA (Google Workspace, Okta, Azure AD all support this)
3. Re-apply: `terraform apply`

### Session Duration Format Error

**Cause**: Session duration must be in Cloudflare's format (e.g., "480m", "8h").

**Solution**: The implementation automatically converts `session_duration_minutes` to the correct format. Ensure the value is a positive integer.

### Dynamic Block Not Creating Rules

**Cause**: Empty list variables (e.g., `allowed_emails = []`) don't generate dynamic blocks.

**Solution**: Ensure at least one email or group is specified:
```hcl
allowed_emails = ["@example.com"]
```

## Validation

After deployment, verify your Access Application:

```bash
# View Terraform outputs
terraform output

# Test access to protected hostname
curl https://app.example.com
# Should redirect to identity provider login
```

## Multi-Environment Setup

Use Terraform workspaces or separate directories for different environments:

### Using Workspaces

```bash
# Create workspaces
terraform workspace new dev
terraform workspace new staging
terraform workspace new prod

# Switch to workspace
terraform workspace select prod

# Apply configuration
terraform apply
```

### Using Separate Directories

```
environments/
├── dev/
│   ├── main.tf -> ../../iac/tf/main.tf (symlink)
│   ├── terraform.tfvars
│   └── backend.tf
├── staging/
│   ├── main.tf -> ../../iac/tf/main.tf (symlink)
│   ├── terraform.tfvars
│   └── backend.tf
└── prod/
    ├── main.tf -> ../../iac/tf/main.tf (symlink)
    ├── terraform.tfvars
    └── backend.tf
```

## Best Practices

1. **Use Remote State**: Store state in Terraform Cloud, S3, or GCS for team collaboration

2. **Enable State Locking**: Prevent concurrent modifications
   ```hcl
   backend "s3" {
     bucket         = "my-terraform-state"
     key            = "cloudflare-access/terraform.tfstate"
     region         = "us-east-1"
     dynamodb_table = "terraform-locks"
   }
   ```

3. **Version Control**: Commit `.tf` files to git, but **not** `terraform.tfstate` or `.tfvars` files with secrets

4. **Use Variables**: Store environment-specific values in `.tfvars` files:
   ```bash
   terraform apply -var-file="prod.tfvars"
   ```

5. **Plan Before Apply**: Always run `terraform plan` to preview changes

6. **Automate in CI/CD**: Integrate Terraform with GitHub Actions, GitLab CI, or your CI/CD platform

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy Cloudflare Access Application

on:
  push:
    branches:
      - main

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0

      - name: Terraform Init
        run: terraform init
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}

      - name: Terraform Plan
        run: terraform plan
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main'
        run: terraform apply -auto-approve
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
```

## Importing Existing Resources

If you have existing Access Applications created via the dashboard, import them:

```bash
# Import Access Application
terraform import cloudflare_zero_trust_access_application.main <account_id>/<application_id>

# Import Access Policy
terraform import cloudflare_zero_trust_access_policy.main <account_id>/<policy_id>
```

## Debugging

### Enable Terraform Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply
```

### View Detailed Plan

```bash
terraform plan -out=plan.tfplan
terraform show plan.tfplan
```

### Validate Configuration

```bash
terraform validate
```

### Format Code

```bash
terraform fmt
```

## Further Reading

- **Component Architecture**: See [../../docs/README.md](../../docs/README.md) for architectural overview
- **User Guide**: See [../../README.md](../../README.md) for usage instructions
- **Examples**: See [../../examples.md](../../examples.md) for common use cases
- **Terraform Cloudflare Provider Docs**: [registry.terraform.io/providers/cloudflare/cloudflare](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- **Cloudflare Zero Trust Docs**: [developers.cloudflare.com/cloudflare-one/applications](https://developers.cloudflare.com/cloudflare-one/applications)

## Support

For issues or questions:
- **OpenMCF**: [openmcf.org](https://openmcf.org)
- **Terraform Support**: [discuss.hashicorp.com](https://discuss.hashicorp.com/)
- **Cloudflare Community**: [community.cloudflare.com](https://community.cloudflare.com)


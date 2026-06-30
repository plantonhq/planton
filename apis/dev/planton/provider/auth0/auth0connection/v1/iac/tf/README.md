# Auth0Connection Terraform Module

This directory contains the Terraform implementation for the Auth0Connection deployment component.

## Overview

The Auth0Connection Terraform module creates and manages Auth0 identity connections, including:
- Database connections (Auth0 hosted)
- Social identity providers (Google, Facebook, GitHub, etc.)
- Enterprise SSO (SAML, OIDC, Azure AD)

## Prerequisites

1. **Terraform 1.0+**: Install from https://www.terraform.io/downloads
2. **Auth0 Account**: With a Machine-to-Machine application configured

## Required Providers

```hcl
terraform {
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
  }
}
```

## Usage

### Initialize Terraform

```bash
terraform init
```

### Create terraform.tfvars

```hcl
auth0_credential = {
  domain        = "your-tenant.auth0.com"
  client_id     = "your-m2m-client-id"
  client_secret = "your-m2m-client-secret"
}

metadata = {
  name = "user-database"
  org  = "my-organization"
  env  = "production"
}

spec = {
  strategy     = "auth0"
  display_name = "Email Sign Up"
  enabled_clients = ["client-id-1", "client-id-2"]
  database_options = {
    password_policy        = "good"
    brute_force_protection = true
    password_history_size  = 5
  }
}
```

### Plan and Apply

```bash
# Preview changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy
```

## Module Structure

```
tf/
├── provider.tf   # Auth0 provider configuration
├── variables.tf  # Input variable definitions
├── locals.tf     # Local value computations
├── main.tf       # Auth0 connection resource
├── outputs.tf    # Output definitions
└── README.md     # This file
```

## Input Variables

### auth0_credential

Auth0 API credentials (sensitive):

```hcl
variable "auth0_credential" {
  type = object({
    domain        = string
    client_id     = string
    client_secret = string
  })
  sensitive = true
}
```

### metadata

Resource metadata:

```hcl
variable "metadata" {
  type = object({
    name   = string
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
  })
}
```

### spec

Connection specification with strategy-specific options.

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Auth0 connection ID |
| `name` | Connection name |
| `strategy` | Connection strategy type |
| `is_enabled` | Whether the connection has enabled clients |
| `enabled_client_ids` | List of enabled client IDs |
| `realms` | Connection realms |

## Strategy-Specific Configuration

### Database (auth0)

```hcl
spec = {
  strategy = "auth0"
  database_options = {
    password_policy        = "good"
    brute_force_protection = true
    password_history_size  = 5
  }
}
```

### Social (google-oauth2, github, etc.)

```hcl
spec = {
  strategy = "google-oauth2"
  social_options = {
    client_id     = "google-client-id"
    client_secret = "google-client-secret"
    scopes        = ["openid", "profile", "email"]
  }
}
```

### SAML (samlp)

```hcl
spec = {
  strategy = "samlp"
  saml_options = {
    sign_in_endpoint = "https://idp.example.com/sso"
    signing_cert     = "-----BEGIN CERTIFICATE-----..."
    entity_id        = "https://idp.example.com"
  }
}
```

### OIDC

```hcl
spec = {
  strategy = "oidc"
  oidc_options = {
    issuer        = "https://idp.example.com"
    client_id     = "oidc-client-id"
    client_secret = "oidc-client-secret"
  }
}
```

### Azure AD (waad)

```hcl
spec = {
  strategy = "waad"
  azure_ad_options = {
    client_id     = "azure-app-id"
    client_secret = "azure-secret"
    domain        = "contoso.onmicrosoft.com"
  }
}
```

## Troubleshooting

### "Error: Provider configuration not present"

Ensure `auth0_credential` variable is provided via `terraform.tfvars` or `-var` flags.

### "Error: Connection already exists"

Auth0 connection names must be unique within a tenant. Either:
- Import the existing connection: `terraform import auth0_connection.this con_xxxxx`
- Use a different name in the metadata

### "Error: Invalid strategy"

Ensure the strategy value is one of the supported types:
- auth0, google-oauth2, facebook, github, linkedin, twitter, microsoft-account, apple
- samlp, oidc, waad, ad, adfs

## Related Documentation

- [Auth0Connection spec.proto](../../spec.proto)
- [Examples](../../examples.md)
- [Research Documentation](../../docs/README.md)
- [Auth0 Terraform Provider](https://registry.terraform.io/providers/auth0/auth0/latest/docs)


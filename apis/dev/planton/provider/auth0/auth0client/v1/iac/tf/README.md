# Auth0Client Terraform Module

This directory contains the Terraform implementation for the Auth0Client deployment component.

## Overview

The Auth0Client Terraform module creates and manages Auth0 applications (clients), including:
- Single Page Applications (SPAs)
- Native applications (mobile/desktop)
- Regular web applications (server-side)
- Machine-to-Machine (M2M) applications

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
  name = "my-spa-app"
  org  = "my-organization"
  env  = "production"
}

spec = {
  application_type = "spa"
  description      = "My React SPA Application"
  callbacks = [
    "https://myapp.com/callback",
    "http://localhost:3000/callback"
  ]
  allowed_logout_urls = [
    "https://myapp.com",
    "http://localhost:3000"
  ]
  web_origins = [
    "https://myapp.com",
    "http://localhost:3000"
  ]
  grant_types = ["authorization_code", "refresh_token"]
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
├── main.tf       # Auth0 client resource
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

Client specification with application-type-specific options.

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Auth0 client internal ID |
| `client_id` | OAuth 2.0 client identifier (public) |
| `client_secret` | OAuth 2.0 client secret (confidential, sensitive) |
| `name` | Application name |
| `application_type` | Application type |
| `signing_keys` | JWT signing keys |
| `token_endpoint_auth_method` | Token endpoint auth method |

## Application Type Examples

### Single Page Application (SPA)

```hcl
spec = {
  application_type = "spa"
  callbacks        = ["https://myapp.com/callback"]
  web_origins      = ["https://myapp.com"]
  grant_types      = ["authorization_code", "refresh_token"]
  refresh_token = {
    rotation_type   = "rotating"
    expiration_type = "expiring"
    token_lifetime  = 2592000
  }
}
```

### Native Application (Mobile)

```hcl
spec = {
  application_type = "native"
  callbacks        = ["myapp://callback"]
  grant_types      = ["authorization_code", "refresh_token"]
  mobile = {
    ios = {
      team_id               = "ABCDE12345"
      app_bundle_identifier = "com.example.myapp"
    }
    android = {
      app_package_name         = "com.example.myapp"
      sha256_cert_fingerprints = ["D8:A0:..."]
    }
  }
  native_social_login = {
    apple    = { enabled = true }
    facebook = { enabled = true }
  }
}
```

### Regular Web Application

```hcl
spec = {
  application_type = "regular_web"
  callbacks        = ["https://mywebapp.com/auth/callback"]
  grant_types      = ["authorization_code", "refresh_token"]
  jwt_configuration = {
    lifetime_in_seconds = 36000
    alg                 = "RS256"
  }
}
```

### Machine-to-Machine (M2M)

```hcl
spec = {
  application_type = "non_interactive"
  description      = "Backend API service"
  grant_types      = ["client_credentials"]
  jwt_configuration = {
    lifetime_in_seconds = 86400
    alg                 = "RS256"
  }
}
```

## Troubleshooting

### "Error: Provider configuration not present"

Ensure `auth0_credential` variable is provided via `terraform.tfvars` or `-var` flags.

### "Error: Invalid application type"

Ensure `application_type` is one of:
- `native`: Mobile/desktop applications
- `spa`: Single Page Applications
- `regular_web`: Server-side web applications
- `non_interactive`: Machine-to-Machine applications

### "Error: Invalid grant type"

Ensure grant types match the application type:
- SPAs/Native: `authorization_code`, `refresh_token`
- Regular web: `authorization_code`, `refresh_token`, `client_credentials`
- M2M: `client_credentials`

## Related Documentation

- [Auth0Client spec.proto](../../spec.proto)
- [Examples](../../examples.md)
- [Research Documentation](../../docs/README.md)
- [Auth0 Terraform Provider](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/client)



# Auth0 Resource Server Terraform Module

This Terraform module deploys an Auth0 Resource Server (API).

## Usage

```hcl
module "auth0_resource_server" {
  source = "./path/to/module"

  metadata = {
    name = "my-api"
    org  = "my-organization"
  }

  spec = {
    identifier         = "https://api.example.com/"
    name               = "My Example API"
    signing_alg        = "RS256"
    token_lifetime     = 86400
    allow_offline_access = true
    enforce_policies   = true
    token_dialect      = "access_token_authz"
    scopes = [
      {
        name        = "read:data"
        description = "Read access to data"
      },
      {
        name        = "write:data"
        description = "Write access to data"
      }
    ]
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| auth0 | >= 1.0 |

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name and labels | object | yes |
| spec | Auth0 Resource Server specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| id | The internal Auth0 identifier |
| identifier | The API identifier (audience) |
| name | The friendly display name |
| signing_alg | The token signing algorithm |
| signing_secret | The signing secret (HS256 only, sensitive) |
| token_lifetime | Token validity duration in seconds |
| token_lifetime_for_web | Token validity for implicit/hybrid flows |
| allow_offline_access | Whether refresh tokens can be issued |
| skip_consent_for_verifiable_first_party_clients | Consent skip setting |
| enforce_policies | Whether RBAC is enabled |
| token_dialect | Access token format |

## Provider Configuration

Set the following environment variables:

```bash
export AUTH0_DOMAIN="your-tenant.auth0.com"
export AUTH0_CLIENT_ID="your-client-id"
export AUTH0_CLIENT_SECRET="your-client-secret"
```

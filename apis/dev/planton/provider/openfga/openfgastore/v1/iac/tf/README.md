# OpenFGA Store - Terraform Module

This Terraform module creates an OpenFGA store.

## Overview

A **store** is the top-level container for authorization data in OpenFGA. Each store contains:
- **Authorization models**: Define the types, relations, and permissions
- **Relationship tuples**: The actual authorization data (who has what relation to what object)

## Usage

### Prerequisites

1. OpenFGA server running (self-hosted or cloud-hosted like [Okta FGA](https://fga.dev))
2. OpenFGA credentials configured via environment variables

### Environment Variables

The provider is configured via environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `FGA_API_URL` | Yes | OpenFGA server URL (e.g., `http://localhost:8080` or `https://api.us1.fga.dev`) |
| `FGA_API_TOKEN` | No* | API token for token-based authentication |
| `FGA_CLIENT_ID` | No* | Client ID for client credentials authentication |
| `FGA_CLIENT_SECRET` | No* | Client secret for client credentials authentication |
| `FGA_API_TOKEN_ISSUER` | No* | Token issuer URL for client credentials flow |
| `FGA_API_SCOPES` | No | OAuth scopes (space-separated) |
| `FGA_API_AUDIENCE` | No | OAuth audience |

*Either `FGA_API_TOKEN` OR (`FGA_CLIENT_ID` + `FGA_CLIENT_SECRET` + `FGA_API_TOKEN_ISSUER`) is required.

### Planton CLI

```bash
# Create credentials file
cat > openfga-creds.yaml << EOF
apiUrl: http://localhost:8080
apiToken: your-api-token
EOF

# Deploy using Terraform/Tofu (required - no Pulumi provider)
planton apply --manifest openfga-store.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

### Direct Terraform Usage

```bash
export FGA_API_URL="http://localhost:8080"
export FGA_API_TOKEN="your-api-token"

terraform init
terraform apply -var-file=terraform.tfvars
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env, labels) | `object` | Yes |
| `spec` | Store specification | `object` | Yes |
| `spec.name` | Display name of the store | `string` | Yes |

## Outputs

| Name | Description |
|------|-------------|
| `id` | The unique identifier of the OpenFGA store |
| `name` | The name of the OpenFGA store |

## Example

```hcl
module "openfga_store" {
  source = "./iac/tf"

  metadata = {
    name = "production-authz"
    org  = "my-company"
    env  = "production"
  }

  spec = {
    name = "production-authorization-store"
  }
}
```

## Resources Created

- `openfga_store.this` - The OpenFGA store

## Notes

- Store names are **immutable** - changing the name will replace the store
- Deleting a store will **permanently delete** all authorization models and relationship tuples within it
- Consider using separate stores for different environments (dev/staging/prod) or tenants

## References

- [OpenFGA Documentation](https://openfga.dev/docs)
- [Terraform Provider OpenFGA](https://registry.terraform.io/providers/openfga/openfga/latest/docs)
- [OpenFGA Store Resource](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/store)

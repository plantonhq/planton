# OpenFGA Relationship Tuple - Terraform Module

Deploy OpenFGA relationship tuples using Terraform/OpenTofu.

## Overview

This module creates an OpenFGA relationship tuple in an existing store. Relationship tuples represent the actual authorization data - they define who has what access to which resources.

## Prerequisites

- An existing OpenFGA store (deploy `OpenFgaStore` first)
- An authorization model in the store (deploy `OpenFgaAuthorizationModel` first)
- OpenFGA credentials configured via environment variables

## Usage

This module is typically invoked by Planton CLI:

```bash
planton apply --manifest relationship-tuple.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env, labels) | object | yes |
| `spec.store_id` | ID of the OpenFGA store | string | yes |
| `spec.authorization_model_id` | ID of the authorization model (optional, uses latest) | string | no |
| `spec.user` | The subject being granted access | string | yes |
| `spec.relation` | The relationship type | string | yes |
| `spec.object` | The resource being accessed | string | yes |
| `spec.condition` | Optional condition for dynamic access | object | no |

## Outputs

| Name | Description |
|------|-------------|
| `user` | The user of the relationship tuple |
| `relation` | The relation of the relationship tuple |
| `object` | The object of the relationship tuple |

## Provider Configuration

The OpenFGA provider is configured via environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `FGA_API_URL` | OpenFGA API URL | Yes |
| `FGA_API_TOKEN` | API token (for token auth) | No* |
| `FGA_CLIENT_ID` | Client ID (for OAuth) | No* |
| `FGA_CLIENT_SECRET` | Client secret (for OAuth) | No* |
| `FGA_API_TOKEN_ISSUER` | Token issuer URL (for OAuth) | No* |

*Either token auth or client credentials are required.

## Important Notes

### Immutability

Relationship tuples in OpenFGA are immutable. When you change any field (user, relation, object, or condition), Terraform will:

1. Delete the old tuple
2. Create a new tuple

This is the expected behavior - tuples represent specific relationships and cannot be modified in place.

### Tuple Uniqueness

A tuple is uniquely identified by (store_id, user, relation, object). You cannot create duplicate tuples with the same combination.

### Conditions

Conditions enable dynamic access control. The condition must be defined in the authorization model before it can be used in tuples.

```hcl
spec = {
  store_id = "01HXYZ..."
  user     = "user:anne"
  relation = "viewer"
  object   = "document:budget"
  condition = {
    name         = "in_allowed_ip_range"
    context_json = jsonencode({
      allowed_ips = ["192.168.1.0/24"]
    })
  }
}
```

## References

- [Terraform Provider Documentation](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/relationship_tuple)
- [OpenFGA Relationship Tuple Concepts](https://openfga.dev/docs/concepts#what-is-a-relationship-tuple)
- [OpenFGA Conditions](https://openfga.dev/docs/modeling/conditions)

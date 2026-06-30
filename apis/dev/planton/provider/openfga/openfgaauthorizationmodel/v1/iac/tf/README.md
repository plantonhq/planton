# OpenFGA Authorization Model - Terraform Module

Deploy OpenFGA authorization models using Terraform/OpenTofu.

## Overview

This module creates an OpenFGA authorization model in an existing store. Authorization models define the types, relations, and access rules for fine-grained authorization.

## Prerequisites

- An existing OpenFGA store (deploy `OpenFgaStore` first)
- OpenFGA credentials configured via environment variables

## Usage

This module is typically invoked by Planton CLI:

```bash
planton apply --manifest authorization-model.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env, labels) | object | yes |
| `spec.store_id` | ID of the OpenFGA store | string | yes |
| `spec.model_json` | Authorization model definition in JSON | string | yes |

## Outputs

| Name | Description |
|------|-------------|
| `id` | The unique identifier of the authorization model |

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

Authorization models in OpenFGA are immutable. When you change `model_json`, Terraform will:

1. Create a new model with a new ID
2. The old model remains in the store

This is by design - it allows for safe model evolution and rollback.

### Model Validation

The model JSON must be valid according to the OpenFGA schema. Use the OpenFGA CLI to validate:

```bash
fga model validate --file model.json
```

Or convert from DSL format:

```bash
fga model transform --file model.fga --output-format json
```

## References

- [Terraform Provider Documentation](https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/authorization_model)
- [OpenFGA Authorization Model Concepts](https://openfga.dev/docs/concepts#what-is-an-authorization-model)

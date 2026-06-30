---
title: "Store"
description: "Store deployment documentation"
icon: "package"
order: 100
componentName: "openfgastore"
---

# OpenFGA Store

Deploys an OpenFGA store — the top-level container for authorization models and relationship tuples. Each store provides isolated authorization data, making it suitable for separating environments, applications, or tenants. This component requires Terraform/Tofu as the provisioner; no Pulumi provider is available.

## What Gets Created

When you deploy an OpenFgaStore resource, Planton provisions:

- **OpenFGA Store** — an `openfga_store` resource that creates a named store on the configured OpenFGA server

## Prerequisites

- **OpenFGA server** — a running OpenFGA instance (self-hosted or cloud-hosted)
- **OpenFGA credentials** configured via environment variables: `FGA_API_URL` (required), plus either `FGA_API_TOKEN` for token-based auth or `FGA_CLIENT_ID`, `FGA_CLIENT_SECRET`, and `FGA_API_TOKEN_ISSUER` for client credentials auth
- **Terraform/Tofu** — this component has no Pulumi provider; set the provisioner label to `tofu`

## Quick Start

Create a file `store.yaml`:

```yaml
apiVersion: openfga.planton.dev/v1
kind: OpenFgaStore
metadata:
  name: my-store
  labels:
    planton.dev/provisioner: tofu
spec:
  name: my-authorization-store
```

Deploy:

```shell
planton apply -f store.yaml
```

This creates a single OpenFGA store named `my-authorization-store`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Display name of the store on the OpenFGA server. The name identifies the store in the OpenFGA API. Immutable — changing it requires replacing the store. | Required |

### Optional Fields

This component has no optional fields.

## Examples

### Development Store

A store for local development or CI environments:

```yaml
apiVersion: openfga.planton.dev/v1
kind: OpenFgaStore
metadata:
  name: dev-authz
  labels:
    planton.dev/provisioner: tofu
spec:
  name: dev-authorization-store
```

### Per-Application Store

Separate stores isolate authorization data between applications running in the same OpenFGA server:

```yaml
apiVersion: openfga.planton.dev/v1
kind: OpenFgaStore
metadata:
  name: billing-authz
  labels:
    planton.dev/provisioner: tofu
spec:
  name: billing-service-authz
```

### Production Store

A production environment store with a descriptive name reflecting its scope:

```yaml
apiVersion: openfga.planton.dev/v1
kind: OpenFgaStore
metadata:
  name: prod-authz
  labels:
    planton.dev/provisioner: tofu
spec:
  name: production-authorization-store
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Unique identifier of the OpenFGA store, required for creating authorization models and relationship tuples |
| `name` | `string` | Display name of the store as configured in `spec.name` |

## Related Components

- [OpenFgaAuthorizationModel](/docs/catalog/openfga/authorization-model) — defines the types, relations, and access rules within a store
- [OpenFgaRelationshipTuple](/docs/catalog/openfga/relationship-tuple) — creates authorization data (who has what relation to which object) within a store

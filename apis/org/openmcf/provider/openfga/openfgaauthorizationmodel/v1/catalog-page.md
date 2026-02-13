# OpenFGA Authorization Model

Deploys an authorization model into an existing OpenFGA store. The model defines types, relations, and access rules that govern how permission checks are evaluated. Models can be specified in DSL format (recommended) or JSON format, and are immutable — each change creates a new model version with a new ID. This component requires Terraform/Tofu as the provisioner; no Pulumi provider is available.

## What Gets Created

When you deploy an OpenFgaAuthorizationModel resource, OpenMCF provisions:

- **Authorization Model Document** — an `openfga_authorization_model_document` data source that converts DSL to JSON, created only when `modelDsl` is provided
- **Authorization Model** — an `openfga_authorization_model` resource containing the type definitions, relations, and conditions for fine-grained access control

## Prerequisites

- **OpenFGA server** — a running OpenFGA instance (self-hosted or cloud-hosted)
- **OpenFGA credentials** configured via environment variables: `FGA_API_URL` (required), plus either `FGA_API_TOKEN` for token-based auth or `FGA_CLIENT_ID`, `FGA_CLIENT_SECRET`, and `FGA_API_TOKEN_ISSUER` for client credentials auth
- **An existing OpenFGA store** — provide the store ID directly or reference an OpenFgaStore resource via `valueFrom`
- **Terraform/Tofu** — this component has no Pulumi provider; set the provisioner label to `tofu`

## Quick Start

Create a file `authz-model.yaml`:

```yaml
apiVersion: open-fga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: my-model
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId: "01HXYZ..."
  modelDsl: |
    model
      schema 1.1

    type user

    type document
      relations
        define viewer: [user]
```

Deploy:

```shell
openmcf apply -f authz-model.yaml
```

This creates an authorization model with a `user` type and a `document` type that has a `viewer` relation.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `storeId` | `string` | ID of the OpenFGA store where the model is created. Immutable — changing it requires replacing the model. Can reference an OpenFgaStore resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `modelDsl` | `string` | — | Authorization model in DSL format (recommended). More human-readable than JSON. Automatically converted to JSON during deployment. Exactly one of `modelDsl` or `modelJson` must be specified. |
| `modelJson` | `string` | — | Authorization model in JSON format. Must include `schema_version` and `type_definitions`. Exactly one of `modelDsl` or `modelJson` must be specified. |

## Examples

### Basic Model with DSL

A minimal model defining users and documents with viewer, editor, and owner relations, using the recommended DSL format:

```yaml
apiVersion: open-fga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: basic-model
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId: "01HXYZ..."
  modelDsl: |
    model
      schema 1.1

    type user

    type document
      relations
        define viewer: [user]
        define editor: [user]
        define owner: [user]
```

### Model with JSON Format

The same model defined in JSON, for teams that prefer structured data or are migrating existing JSON definitions:

```yaml
apiVersion: open-fga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: json-model
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId: "01HXYZ..."
  modelJson: |
    {
      "schema_version": "1.1",
      "type_definitions": [
        {
          "type": "user",
          "relations": {}
        },
        {
          "type": "document",
          "relations": {
            "viewer": {"this": {}},
            "editor": {"this": {}},
            "owner": {"this": {}}
          },
          "metadata": {
            "relations": {
              "viewer": {"directly_related_user_types": [{"type": "user"}]},
              "editor": {"directly_related_user_types": [{"type": "user"}]},
              "owner": {"directly_related_user_types": [{"type": "user"}]}
            }
          }
        }
      ]
    }
```

### Hierarchical Model with Foreign Key Reference

A role-based model with groups, folders, and documents using a foreign key reference to resolve the store ID from an OpenFgaStore resource:

```yaml
apiVersion: open-fga.openmcf.org/v1
kind: OpenFgaAuthorizationModel
metadata:
  name: rbac-model
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    valueFrom:
      kind: OpenFgaStore
      name: prod-authz
      field: status.outputs.id
  modelDsl: |
    model
      schema 1.1

    type user

    type group
      relations
        define member: [user]

    type folder
      relations
        define owner: [user]
        define viewer: [user, group#member]

    type document
      relations
        define parent: [folder]
        define owner: [user]
        define editor: [user, group#member]
        define viewer: [user, group#member] or editor or owner or viewer from parent
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Unique identifier of the authorization model version. Each model change produces a new ID. |

## Related Components

- [OpenFgaStore](/docs/catalog/openfga/openfgastore) — provides the store where authorization models are created
- [OpenFgaRelationshipTuple](/docs/catalog/openfga/openfgarelationshiptuple) — creates authorization data (who has what relation to which object) evaluated against the model

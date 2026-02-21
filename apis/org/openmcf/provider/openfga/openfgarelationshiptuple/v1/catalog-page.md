# OpenFGA Relationship Tuple

Deploys a relationship tuple into an existing OpenFGA store. A relationship tuple is the fundamental unit of authorization data in OpenFGA, representing that a specific user (or userset) has a particular relation to an object. Together with an authorization model, tuples determine access decisions at check time. All tuple fields are immutable — changing any field replaces the tuple. This component requires Terraform/Tofu as the provisioner; no Pulumi provider is available.

## What Gets Created

When you deploy an OpenFgaRelationshipTuple resource, OpenMCF provisions:

- **Relationship Tuple** — an `openfga_relationship_tuple` resource that writes a single authorization tuple (user, relation, object) into the specified OpenFGA store

## Prerequisites

- **OpenFGA server** — a running OpenFGA instance (self-hosted or cloud-hosted)
- **OpenFGA credentials** configured via environment variables: `FGA_API_URL` (required), plus either `FGA_API_TOKEN` for token-based auth or `FGA_CLIENT_ID`, `FGA_CLIENT_SECRET`, and `FGA_API_TOKEN_ISSUER` for client credentials auth
- **An existing OpenFGA store** — provide the store ID directly or reference an OpenFgaStore resource via `valueFrom`
- **An authorization model** — the store must contain a model that defines the types and relations used in the tuple
- **Terraform/Tofu** — this component has no Pulumi provider; set the provisioner label to `tofu`

## Quick Start

Create a file `tuple.yaml`:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    value: "01HXYZ..."
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

Deploy:

```shell
openmcf apply -f tuple.yaml
```

This creates a relationship tuple granting user `anne` the `viewer` relation on `document:budget-2024`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `storeId` | `StringValueOrRef` | ID of the OpenFGA store this tuple belongs to. Immutable — changing it requires replacing the tuple. Can reference an OpenFgaStore resource via `valueFrom`. | Required |
| `user` | `object` | The subject of the relationship tuple. Contains `type` (string, required), `id` (StringValueOrRef, required), and `relation` (string, optional — for usersets). The module combines these into `type:id` or `type:id#relation`. | Required |
| `relation` | `string` | The relationship type between the user and object. Must be defined in the authorization model for the object type. Examples: `viewer`, `editor`, `owner`, `member`, `admin`. | Required |
| `object` | `object` | The resource the user is being granted access to. Contains `type` (string, required) and `id` (StringValueOrRef, required). The module combines these into `type:id`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `authorizationModelId` | `StringValueOrRef` | — | ID of the authorization model to validate the tuple against. When omitted, the tuple is associated with the latest model in the store. Can reference an OpenFgaAuthorizationModel resource via `valueFrom`. |
| `condition` | `object` | — | An optional condition that must be satisfied at check time. Contains `name` (string, required — the condition name defined in the authorization model) and `contextJson` (string, optional — partial context in JSON format merged with runtime context). |

## Examples

### Grant a User Access to a Document

A minimal tuple granting a single user viewer access to a specific document, using direct values for all fields:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-views-budget
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    value: "01HXYZ..."
  user:
    type: user
    id:
      value: anne
  relation: viewer
  object:
    type: document
    id:
      value: budget-2024
```

### Group Membership with Userset

Add a user to a group using the userset format (`group:engineering#member`). Other tuples can then reference this group membership to grant indirect access:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: anne-member-engineering
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    value: "01HXYZ..."
  user:
    type: user
    id:
      value: anne
  relation: member
  object:
    type: group
    id:
      value: engineering
```

A second tuple grants all members of the engineering group editor access to a folder, using the userset `relation` field on the user:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: engineering-edits-reports
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    value: "01HXYZ..."
  user:
    type: group
    id:
      value: engineering
    relation: member
  relation: editor
  object:
    type: folder
    id:
      value: reports
```

### Conditional Tuple with Foreign Key References

A tuple that uses foreign key references to resolve the store and model IDs from other OpenMCF resources, and includes a condition that restricts access to a set of allowed IP ranges. The condition must be defined in the authorization model:

```yaml
apiVersion: openfga.openmcf.org/v1
kind: OpenFgaRelationshipTuple
metadata:
  name: bob-edits-roadmap-conditional
  labels:
    openmcf.org/provisioner: tofu
spec:
  storeId:
    valueFrom:
      kind: OpenFgaStore
      name: prod-authz
      field: status.outputs.id
  authorizationModelId:
    valueFrom:
      kind: OpenFgaAuthorizationModel
      name: rbac-model
      field: status.outputs.id
  user:
    type: user
    id:
      value: bob
  relation: editor
  object:
    type: document
    id:
      value: roadmap-2025
  condition:
    name: in_allowed_ip_range
    contextJson: '{"allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"]}'
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `user` | `string` | The subject of the relationship tuple that was created, in `type:id` or `type:id#relation` format |
| `relation` | `string` | The relationship type that was created |
| `object` | `string` | The resource the tuple grants access to, in `type:id` format |

## Related Components

- [OpenFgaStore](/docs/catalog/openfga/openfgastore) — provides the store where relationship tuples are written
- [OpenFgaAuthorizationModel](/docs/catalog/openfga/openfgaauthorizationmodel) — defines the types, relations, and conditions that govern how tuples are evaluated

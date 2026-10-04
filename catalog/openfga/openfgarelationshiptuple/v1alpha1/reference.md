# OpenFgaRelationshipTuple

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `openfga.planton.dev/v1alpha1`

OpenFgaRelationshipTupleSpec defines the configuration for an OpenFGA Relationship Tuple.

A relationship tuple is the core data structure in OpenFGA that represents a relationship
between a user (or userset) and an object through a specific relation. Together with an
authorization model, relationship tuples determine whether a user has access to an object.

The tuple consists of three required parts:
- user: Who or what is being granted access (structured as type + id + optional relation)
- relation: The type of access or relationship (e.g., "viewer", "editor", "owner")
- object: What is being accessed (structured as type + id)

Optionally, a condition can be specified to add dynamic access rules that are evaluated
at check time using the provided context.

IMPORTANT: Relationship tuples are immutable in OpenFGA. Changing any field requires
deleting the old tuple and creating a new one. Terraform handles this automatically.

IMPORTANT: OpenFGA only has a Terraform provider - there is no Pulumi provider available.
This kind supports only Terraform/Tofu as the provisioner.

Reference:
- Terraform: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/relationship_tuple
- OpenFGA Docs: https://openfga.dev/docs/concepts#what-is-a-relationship-tuple

## Example

```yaml
# OpenFgaRelationshipTuple Test Manifest
#
# This manifest is used for testing the OpenFGA Relationship Tuple component.
#
# Prerequisites:
# - OpenFGA server running (locally or cloud-hosted)
# - OpenFGA credentials configured
# - An existing OpenFGA store (store_id from OpenFgaStore deployment)
# - An existing authorization model in the store (from OpenFgaAuthorizationModel deployment)
#
# Usage with Terraform/Tofu (required - no Pulumi provider available):
#   planton apply --manifest manifest.yaml \
#     --openfga-provider-config /path/to/openfga-creds.yaml \
#     --provisioner tofu
#
# Example OpenFGA credentials file (openfga-creds.yaml):
#   apiUrl: http://localhost:8080
#   # For token auth:
#   apiToken: your-api-token
#   # Or for client credentials:
#   # clientId: your-client-id
#   # clientSecret: your-client-secret
#   # apiTokenIssuer: https://your-issuer/oauth/token

apiVersion: openfga.planton.dev/v1alpha1
kind: OpenFgaRelationshipTuple
metadata:
  name: test-tuple
  org: planton
  env: development
spec:
  # Replace with actual store ID from OpenFgaStore deployment
  storeId:
    value: 01HXYZ_REPLACE_WITH_ACTUAL_STORE_ID
  # Optional: specify authorization model ID (uses latest if not specified)
  # authorizationModelId: "01HXYZ_REPLACE_WITH_MODEL_ID"
  # Grant user:anne viewer access to document:budget-2024
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

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.storeId` | `string \| valueFrom` | yes |  | OpenFgaStore (`status.outputs.id`) |
| `spec.authorizationModelId` | `string \| valueFrom` |  |  | OpenFgaAuthorizationModel (`status.outputs.id`) |
| `spec.user` | `OpenFgaRelationshipTupleUser` | yes |  |  |
| `spec.user.type` | `string` | yes |  |  |
| `spec.user.id` | `string \| valueFrom` | yes |  |  |
| `spec.user.relation` | `string` |  |  |  |
| `spec.relation` | `string` | yes |  |  |
| `spec.object` | `OpenFgaRelationshipTupleObject` | yes |  |  |
| `spec.object.type` | `string` | yes |  |  |
| `spec.object.id` | `string \| valueFrom` | yes |  |  |
| `spec.condition` | `OpenFgaRelationshipTupleCondition` |  |  |  |
| `spec.condition.name` | `string` | yes |  |  |
| `spec.condition.contextJson` | `string` |  |  |  |

## Field Details

### spec.storeId

`string | valueFrom` · required

store_id is the unique identifier of the OpenFGA store this tuple belongs to.

This can be either:
- A direct value: {value: "01HXYZ..."}
- A reference to an OpenFgaStore: {value_from: {name: "my-store"}}

When using references, the store ID is automatically resolved from the
OpenFgaStore's status.outputs.id field.

Note: The store_id is immutable - changing it requires replacing the tuple.

- references: OpenFgaStore (`status.outputs.id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: OpenFgaStore, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.authorizationModelId

`string | valueFrom`

authorization_model_id is the unique identifier of the authorization model this tuple
is associated with.

This can be either:
- A direct value: {value: "01HXYZ..."}
- A reference to an OpenFgaAuthorizationModel: {value_from: {name: "my-model"}}

When using references, the model ID is automatically resolved from the
OpenFgaAuthorizationModel's status.outputs.id field.

This field is optional. If not specified, the tuple will be associated with the
latest authorization model in the store at the time of creation.

When specified, the tuple is validated against this specific model version. This is
useful for ensuring tuples are compatible with a known model version in production.

Note: The authorization_model_id is immutable - changing it requires replacing the tuple.

- references: OpenFgaAuthorizationModel (`status.outputs.id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: OpenFgaAuthorizationModel, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.user

`OpenFgaRelationshipTupleUser` · required

user is the subject of the relationship tuple - who is being granted access.

The user is specified as a structured object with:
- type: The user type defined in the authorization model (e.g., "user", "group")
- id: The user identifier (e.g., "anne", "engineering", "*" for wildcard)
- relation: Optional, for usersets (e.g., "member" to create "group:engineering#member")

The IaC module combines these into the OpenFGA format:
- Without relation: "type:id" (e.g., "user:anne")
- With relation: "type:id#relation" (e.g., "group:engineering#member")

Note: The user is immutable - changing it requires replacing the tuple.

- rule: {"required":true}

### spec.user.type

`string` · required

type is the user type as defined in the authorization model.

This must match a type defined in the authorization model that is allowed
as a subject for the target relation.

Examples: "user", "group", "team", "service", "application"

- rule: {"required":true}

### spec.user.id

`string | valueFrom` · required

id is the unique identifier of the user.

This can be either:
- A direct value: {value: "anne"}
- A reference to another resource: {value_from: {name: "my-resource"}}

When using references, the ID is automatically resolved from the referenced
resource's field (default: metadata.id).

Use {value: "*"} for wildcard access (all users of this type).

Examples: {value: "anne"}, {value: "1234"}, {value: "*"}, {value_from: {name: "my-user"}}

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.user.relation

`string`

relation is optional, used to create usersets (type:id#relation format).

When specified, the user represents "all entities that have this relation
to the specified object". For example, "all members of the engineering group".

When omitted, the user is a direct reference to "type:id".
When specified, the user becomes "type:id#relation".

Examples: "member", "admin", "owner"

### spec.relation

`string` · required

relation is the relationship type between the user and object.

The relation must be defined in the authorization model for the object type.
Common relations include: viewer, editor, owner, member, admin, parent.

Note: The relation is immutable - changing it requires replacing the tuple.

Examples: "viewer", "editor", "owner", "member", "admin"

- rule: {"required":true}

### spec.object

`OpenFgaRelationshipTupleObject` · required

object is the resource the user is being granted access to.

The object is specified as a structured object with:
- type: The object type defined in the authorization model (e.g., "document", "folder")
- id: The object identifier (e.g., "budget-2024", "reports")

The IaC module combines these into the OpenFGA format: "type:id"
(e.g., "document:budget-2024")

Note: The object is immutable - changing it requires replacing the tuple.

- rule: {"required":true}

### spec.object.type

`string` · required

type is the object type as defined in the authorization model.

This must match a type defined in the authorization model.

Examples: "document", "folder", "project", "organization", "team"

- rule: {"required":true}

### spec.object.id

`string | valueFrom` · required

id is the unique identifier of the object.

This can be either:
- A direct value: {value: "budget-2024"}
- A reference to another resource: {value_from: {name: "my-resource"}}

When using references, the ID is automatically resolved from the referenced
resource's field (default: metadata.id).

Examples: {value: "budget-2024"}, {value: "reports"}, {value_from: {name: "my-project"}}

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.condition

`OpenFgaRelationshipTupleCondition`

condition optionally specifies a condition that must be satisfied for this tuple
to be considered during access checks.

Conditions enable dynamic access control based on runtime context. The condition
must be defined in the authorization model, and the context must provide values
for the condition's required parameters.

This field is optional. If not specified, the tuple is always considered.

### spec.condition.name

`string` · required

name is the name of the condition as defined in the authorization model.

The condition must be declared in the authorization model's conditions section
before it can be used in tuples.

Example: "in_allowed_ip_range", "during_business_hours"

- rule: {"required":true}

### spec.condition.contextJson

`string`

context_json is the partial context provided with the tuple, in JSON format.

This context is merged with the context provided at check time. The combined
context is then evaluated against the condition defined in the authorization model.

The JSON must be a valid object with keys matching the condition's expected parameters.

Example: {"allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"]}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: OpenFgaRelationshipTuple, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user` | `string` | user is the subject of the relationship tuple that was created. This echoes back the user field from the spec. |
| `status.outputs.relation` | `string` | relation is the relationship type that was created. This echoes back the relation field from the spec. |
| `status.outputs.object` | `string` | object is the resource the tuple grants access to. This echoes back the object field from the spec. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storeId` | OpenFgaStore | `status.outputs.id` |
| `spec.authorizationModelId` | OpenFgaAuthorizationModel | `status.outputs.id` |

## See Also

- [Overview](../README.md)

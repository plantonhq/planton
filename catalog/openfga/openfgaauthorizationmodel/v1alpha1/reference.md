# OpenFgaAuthorizationModel

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `openfga.planton.dev/v1alpha1`

OpenFgaAuthorizationModelSpec defines the configuration for an OpenFGA Authorization Model.

An authorization model defines the types, relations, and conditions that govern access control
within an OpenFGA store. Each store can have multiple authorization models, with each new model
being a complete replacement (not a patch) of the type definitions.

The model can be specified in either DSL format (recommended, more human-readable) or JSON format.
Exactly one of model_dsl or model_json must be specified.

IMPORTANT: Authorization models are immutable in OpenFGA. Creating a new model creates a new
version. Changing the model will trigger a replacement (new model ID).

IMPORTANT: OpenFGA only has a Terraform provider - there is no Pulumi provider available.
This kind supports only Terraform/Tofu as the provisioner.

Reference:
- Terraform: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/authorization_model
- OpenFGA Docs: https://openfga.dev/docs/concepts#what-is-an-authorization-model
- OpenFGA DSL: https://openfga.dev/docs/configuration-language

## Example

```yaml
# OpenFgaAuthorizationModel Test Manifest
#
# This manifest is used for testing the OpenFGA Authorization Model component.
#
# Prerequisites:
# - OpenFGA server running (locally or cloud-hosted)
# - OpenFGA credentials configured
# - An existing OpenFGA store (store_id from OpenFgaStore deployment)
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
kind: OpenFgaAuthorizationModel
metadata:
  name: test-model
  org: planton
  env: development
spec:
  # Replace with actual store ID from OpenFgaStore deployment
  storeId:
    value: 01HXYZ_REPLACE_WITH_ACTUAL_STORE_ID
  # Example authorization model with user and document types
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
            "viewer": {
              "this": {}
            },
            "editor": {
              "this": {}
            },
            "owner": {
              "this": {}
            }
          },
          "metadata": {
            "relations": {
              "viewer": {
                "directly_related_user_types": [
                  {"type": "user"}
                ]
              },
              "editor": {
                "directly_related_user_types": [
                  {"type": "user"}
                ]
              },
              "owner": {
                "directly_related_user_types": [
                  {"type": "user"}
                ]
              }
            }
          }
        }
      ]
    }
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.storeId` | `string \| valueFrom` | yes |  | OpenFgaStore (`status.outputs.id`) |
| `spec.modelDsl` | `string` |  |  |  |
| `spec.modelJson` | `string` |  |  |  |

## Field Details

### spec.storeId

`string | valueFrom` · required

store_id is the unique identifier of the OpenFGA store where this model will be created.

This can be either:
- A direct value: {value: "01HXYZ..."}
- A reference to an OpenFgaStore: {value_from: {name: "my-store"}}

When using references, the store ID is automatically resolved from the
OpenFgaStore's status.outputs.id field.

Note: The store_id is immutable - changing it requires replacing the model.

- references: OpenFgaStore (`status.outputs.id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: OpenFgaStore, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.modelDsl

`string`

model_dsl is the authorization model definition in DSL format (recommended).

The DSL format is more human-readable than JSON and is the preferred format
in OpenFGA documentation. The Terraform module automatically converts DSL to JSON
using the openfga_authorization_model_document data source.

Exactly one of model_dsl or model_json must be specified.

Example:
model
  schema 1.1

type user

type document
  relations
    define viewer: [user]
    define editor: [user]
    define owner: [user]

Reference: https://openfga.dev/docs/configuration-language

### spec.modelJson

`string`

model_json is the authorization model definition in JSON format.

Use this if you prefer JSON over DSL, or if you're migrating from existing JSON models.
The JSON must conform to the OpenFGA authorization model schema and include:
- schema_version: The schema version (e.g., "1.1")
- type_definitions: Array of type definitions with name, relations, and metadata
- conditions (optional): Map of conditions for dynamic access decisions

Exactly one of model_dsl or model_json must be specified.

Note: The model is immutable - changing it requires replacing the model (new ID).

Example:
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

## Outputs

Reference an output from another manifest as `valueFrom: {kind: OpenFgaAuthorizationModel, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the authorization model. This is a version-specific identifier - each new model gets a new ID. The model ID is used to: - Reference this specific model version in API calls - Set as the active authorization model for a store - Import the model state into Terraform Format: Typically a ULID (e.g., "01HXYZ...") |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storeId` | OpenFgaStore | `status.outputs.id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| OpenFgaRelationshipTuple | `spec.authorizationModelId` | `status.outputs.id` |

## See Also

- [Overview](../README.md)

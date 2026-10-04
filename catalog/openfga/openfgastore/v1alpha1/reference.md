# OpenFgaStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `openfga.planton.dev/v1alpha1`

OpenFgaStoreSpec defines the configuration for an OpenFGA Store.

A store is a logical container for authorization data in OpenFGA. Each store contains
one or more versions of an authorization model and can contain various relationship tuples.

Separate stores can be created for separate authorization needs or isolated environments
(e.g., development/staging/production).

IMPORTANT: OpenFGA only has a Terraform provider - there is no Pulumi provider available.
This kind supports only Terraform/Tofu as the provisioner.

Reference:
- Terraform: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/store
- OpenFGA Docs: https://openfga.dev/docs/concepts#what-is-a-store

## Example

```yaml
# OpenFgaStore Test Manifest
#
# This manifest is used for testing the OpenFGA Store component.
#
# Prerequisites:
# - OpenFGA server running (locally or cloud-hosted)
# - OpenFGA credentials configured
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
kind: OpenFgaStore
metadata:
  name: test-store
  org: planton
  env: development
spec:
  name: test-authorization-store
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |

## Field Details

### spec.name

`string` · required

name is the display name of the OpenFGA store.

This name is used to identify the store in the OpenFGA server.
The store name should be descriptive of its purpose or environment.

Note: The store name is immutable - changing it requires replacing the store.

Examples: "production-authz", "staging-acl", "dev-permissions"

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: OpenFgaStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the unique identifier of the OpenFGA store. This is the primary identifier used in OpenFGA APIs to reference the store. The store ID is required for all subsequent operations including: - Creating or updating authorization models - Writing relationship tuples - Checking permissions - Listing objects or users |
| `status.outputs.name` | `string` | name is the display name of the store as configured in the spec. This is the human-readable name used to identify the store. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| OpenFgaAuthorizationModel | `spec.storeId` | `status.outputs.id` |
| OpenFgaRelationshipTuple | `spec.storeId` | `status.outputs.id` |

## See Also

- [Overview](../README.md)

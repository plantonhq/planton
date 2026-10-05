# CloudflareSecretsStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareSecretsStoreSpec creates the account-level Secrets Store: the
vault that holds secrets consumed by Workers bindings, AI Gateway
authentication, and other Cloudflare surfaces. The store itself is just a
named container -- the secrets inside it are managed by
CloudflareSecretsStoreSecret resources referencing this store.

Two hard provider facts shape how this resource behaves:
  - EVERY field is create-only: changing the name (or account) replaces
    the store, destroying every secret inside it. Treat the store as
    permanent infrastructure and name it accordingly.
  - Cloudflare currently allows ONE store per account ("default"-style
    deployments): creating a second store fails at the API. If the
    account already has a store (e.g. created from the dashboard), adopt
    it by import instead of creating a new one.

## Example

```yaml
# Complete example manifest for CloudflareSecretsStore. Creates the
# account-level Secrets Store -- the vault Worker bindings and AI Gateway
# authentication consume secrets from. Both fields are create-only, and
# Cloudflare allows one store per account: adopt an existing store by import
# instead of creating a second.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStore
metadata:
  name: account-secrets
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: account-secrets
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the store belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The store's name. Create-only: renaming replaces the store AND every
secret it holds.

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareSecretsStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.store_id` | `string` | The store's ID -- what CloudflareSecretsStoreSecret resources, Worker secrets-store bindings, and AI Gateway authentication reference. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareAiGateway | `spec.storeId` | `status.outputs.store_id` |
| CloudflareSecretsStoreSecret | `spec.storeId` | `status.outputs.store_id` |

## See Also

- [Overview](../README.md)

# CloudflareSecretsStoreSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareSecretsStoreSecretSpec manages one secret inside the account
Secrets Store. The secret's VALUE is write-only at Cloudflare: the API
never returns it, so a value change is proven by the deployment, never by
reading back. Consumers (Worker secrets-store bindings, AI Gateway
authentication) reference the secret by store and name/ID -- rotating the
value here rotates it for every consumer without touching them.

The name is create-only: renaming replaces the secret (new secret ID),
and consumers referencing the old name must be re-pointed.

## Example

```yaml
# Complete example manifest for CloudflareSecretsStoreSecret. Stores one
# secret in the account Secrets Store, readable by Workers bindings and AI
# Gateway. The value is write-only at Cloudflare (never returned); scopes
# must be listed alphabetically -- the API returns them sorted and any other
# order would drift forever.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStoreSecret
metadata:
  name: openai-api-key
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  store_id:
    value: "7b0a3d5c1e9f42c68d1a2b3c4d5e6f70"
  name: openai-api-key
  value:
    value: REPLACE_WITH_SECRET_VALUE
  scopes:
    - ai_gateway
    - workers
  comment: Provider API key consumed by the AI gateway and Workers.
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.storeId` | `string \| valueFrom` | yes |  | CloudflareSecretsStore (`status.outputs.store_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.scopes` | `[]string` | yes |  |  |
| `spec.comment` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the store belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.storeId

`string | valueFrom` · required

The Secrets Store holding this secret: a literal store ID, or a
reference to a CloudflareSecretsStore resource's store_id output.

- references: CloudflareSecretsStore (`status.outputs.store_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareSecretsStore, name: <that resource's name>, fieldPath: status.outputs.store_id}} -- a bare string does not parse

### spec.name

`string` · required

The secret's name -- what consumers reference. Create-only: renaming
replaces the secret under a new ID.

- rule: {"required":true}

### spec.value

`string | valueFrom` · required · sensitive

The secret value (up to 64 KiB). WRITE-ONLY: Cloudflare never returns
it, so it can never be read back, imported, or drift-detected. Provide a
managed-secret reference; the platform resolves it just-in-time at
deploy and never stores it in plaintext.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.scopes

`[]string` · required

The Cloudflare surfaces allowed to read this secret. Cloudflare accepts
exactly these scopes and always RETURNS them alphabetically sorted --
the provider models the list as ordered, so any other order in
configuration shows as permanent plan drift (a defect documented in the
provider's own tests). This spec therefore requires the canonical
alphabetical order up front, killing the drift class at the source
(live-confirmed 2026-08-27: the refresh-inclusive re-plan is clean
under the CEL-ordered list).

- rule: scopes must be drawn from access, ai_gateway, dex, workers -- listed alphabetically, without duplicates (e.g. [access, workers]); Cloudflare returns them sorted, so any other order would drift forever
- rule: {"repeated":{"minItems":"1"}}

### spec.comment

`string`

A free-form note about the secret (shown in the dashboard).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareSecretsStoreSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.secret_id` | `string` | The secret's ID within its store. |
| `status.outputs.store_id` | `string` | The ID of the store holding the secret (echoed for consumers that need the store/secret pair). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storeId` | CloudflareSecretsStore | `status.outputs.store_id` |

## See Also

- [Overview](../README.md)

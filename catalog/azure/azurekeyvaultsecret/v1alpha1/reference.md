# AzureKeyVaultSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureKeyVaultSecretSpec** defines the configuration for storing a
secret (a password, API key, connection string, or any small
sensitive string) inside an Azure Key Vault.

A Key Vault secret is versioned: every change to the value creates
a NEW version and the versioned identifier moves with it, while the
versionless identifier always resolves to the latest version.
Consumers that should follow updates reference versionless_id;
consumers pinned by a compliance regime reference secret_id (one
frozen version).

Composition seams: the parent vault is a first-class AzureKeyVault
(referenced by key_vault_id); the value is a sensitive
reference-resolved input (reference a managed secret or another
resource's output rather than embedding plaintext in manifests).
The deploying credential needs data-plane secret permissions on the
vault (the "Key Vault Administrator" or "Key Vault Secrets Officer"
RBAC role, or secret permissions in a legacy access policy).

The provider's write-only `value_wo`/`value_wo_version` variant is
deliberately not modeled: it is an ephemeral input that duplicates
`value`, and secret values here are already reference-resolved at
deploy time rather than stored in the manifest.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: the
# sensitive value (a literal here -- offline plans never reach Azure;
# production manifests reference a managed secret instead), content
# type, both RFC 3339 attributes, and user tags merged over the
# derived ones.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureKeyVaultSecret
metadata:
  name: test-key-vault-secret
  org: test-org
  env: dev
spec:
  name: test-db-password
  keyVaultId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.KeyVault/vaults/platform-kv
  value:
    value: test-secret-value
  contentType: text/plain
  notBeforeDate: "2027-01-01T00:00:00Z"
  expirationDate: "2028-01-01T00:00:00Z"
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.contentType` | `string` |  |  |  |
| `spec.notBeforeDate` | `string` |  |  |  |
| `spec.expirationDate` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.name

`string` · required

The secret's name within the vault: 1-127 characters of letters,
digits, and hyphens, unique among the vault's secrets. Changing
the name replaces the secret. A deleted secret's name stays
reserved for the vault's soft-delete retention window unless
purged.

- rule: {"required":true,"string":{"pattern":"^[0-9a-zA-Z-]{1,127}$"}}

### spec.keyVaultId

`string | valueFrom` · required

The vault the secret lives in, by ARM resource ID. Defaults to
referencing an AzureKeyVault's key_vault_id output in composed
environments. Changing the vault replaces the secret -- secret
material never moves between vaults.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.value

`string | valueFrom` · required · sensitive

The secret's value. Changing it creates a NEW version of the
secret (the old version remains readable until purged or
expired). Reference a managed secret or another resource's output
rather than embedding the literal in manifests. Key Vault strips
raw newlines -- encode multi-line values (e.g. base64) before
storing.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.contentType

`string`

A free-form hint for consumers about what the value is (e.g.
"text/plain", "application/x-pkcs12b64"). Key Vault stores it
verbatim and never interprets it. Updatable in place.

### spec.notBeforeDate

`string` · optional (explicit presence)

The UTC instant before which the secret must not be used, RFC
3339 format (e.g. "2027-01-01T00:00:00Z"). Leave unset for a
secret usable immediately. Enforcement is the consumer's job --
Key Vault returns the attribute; RBAC-permitted readers can still
fetch the value.

- rule: {"string":{"pattern":"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?Z$"}}

### spec.expirationDate

`string` · optional (explicit presence)

The UTC instant the secret expires, RFC 3339 format. Expiry is an
attribute consumers should honor (and Azure Policy can audit) --
set it for credentials that rotate on a schedule.

- rule: {"string":{"pattern":"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?Z$"}}

### spec.tags

`map<string, string>`

Free-form tags applied to the secret, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Key Vault caps a
secret at 15 tags (the provider's own bound, mirrored here so the
merged map still fits). Updatable in place.

- rule: {"map":{"maxPairs":"15"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureKeyVaultSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.secret_id` | `string` | The secret's versioned data-plane ID: https://{vault}.vault.azure.net/secrets/{name}/{version}. Pins consumers to THIS version -- value updates do not follow. Reference it only when a compliance regime demands a frozen version. |
| `status.outputs.versionless_id` | `string` | The secret's versionless data-plane ID: https://{vault}.vault.azure.net/secrets/{name}. The reference consumers should use -- Azure resolves it to the latest version, so value updates propagate with zero intervention. |
| `status.outputs.secret_name` | `string` | The secret's name within the vault. |
| `status.outputs.version` | `string` | The current version identifier (the trailing segment of secret_id). |
| `status.outputs.resource_id` | `string` | The secret's versioned ARM resource ID (.../vaults/{vault}/secrets/{name}/versions/{version}) -- the control-plane identity used by ARM-level integrations and RBAC scopes. |
| `status.outputs.resource_versionless_id` | `string` | The secret's versionless ARM resource ID (.../vaults/{vault}/secrets/{name}). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |

## See Also

- [Overview](../README.md)

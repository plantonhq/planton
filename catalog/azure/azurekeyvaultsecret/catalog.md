# Azure Key Vault Secret

Stores a secret (a password, API key, connection string, or any small sensitive string) in an Azure Key Vault as a versioned, access-controlled data-plane object. Every value change creates a NEW version: the versioned identifier freezes on it, while the versionless identifier always resolves to the latest -- the split that makes rotation a zero-intervention event for well-wired consumers. The value itself is a sensitive, reference-resolved input, and it is never an output: readers fetch it from the vault at runtime with their own permissions. Secrets are effectively free -- Key Vault bills fractions of a cent per 10,000 operations.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Key Vault secret** -- the versioned secret with its content-type hint, activation/expiry attributes, and tags (merged over the Planton-derived metadata tags; user values win, capped at 15 total)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The value's source** -- a managed secret (referenced as `$secret/<slug>`) or another resource's sensitive output via ValueFromRef; never a literal in the manifest.

### Azure Subscription

- **An Azure Key Vault** -- the vault the secret lives in, referenced by `keyVaultId`.
- **Data-plane secret permissions** -- the deploying credential needs the "Key Vault Administrator" or "Key Vault Secrets Officer" RBAC role (or secret permissions in a legacy access policy). Subscription ownership alone is NOT enough.

## Deploy

### Console

Open the deployment store, find **Azure Key Vault Secret**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the vault reference, the secret's name, the value reference, and the validity attributes. Start from the **Basic Secret** or **Expiring API Key** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureKeyVaultSecret
metadata:
  name: partner-api-key
  org: acme-corp
  env: prod
spec:
  name: partner-api-key
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: app-vault
      fieldPath: status.outputs.key_vault_id
  value:
    value: $secret/partner-api-key
  contentType: text/plain
```

```shell
planton apply -f secret.yaml
```

This stores one secret in the vault with its value resolved from the managed secret at deploy time -- the manifest records that the secret exists and where it lives, never what it is. An Infra Job tracks the provisioning in real time.

### InfraChart

When the secret's value comes from another resource in the same InfraChart, wire both references with ValueFromRef:

```yaml
spec:
  name: storage-connection-string
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: app-vault
      fieldPath: status.outputs.key_vault_id
  value:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.primary_connection_string
  contentType: text/plain
```

The InfraPipeline resolves the dependency graph, deploys the vault and the storage account first, then stores the account's connection string as the secret -- the handoff point between a credential-emitting resource and its consumers.

## Key Configuration

These are the most important decisions when configuring a Key Vault secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The value never belongs in a manifest** -- `value` is sensitive and reference-resolved: point it at a managed secret (`$secret/<slug>`) or another resource's output. A literal value in a manifest defeats the vault's purpose -- the manifest becomes the secret store, unencrypted and version-controlled. Treat any literal that slips through as leaked and rotate it.

**Reference the versionless ID unless compliance says otherwise** -- every value change creates a NEW version; the versioned ID freezes consumers on the old value while the versionless ID follows the latest. Default every consumer to `versionless_id` -- rotation then propagates with zero intervention. Pin a version only when a compliance regime demands a frozen value, and own the re-pin as part of rotation.

**Expiry is an attribute, not an enforcement** -- `expirationDate` and `notBeforeDate` are advisory: Key Vault returns them, Azure Policy can audit them, and well-behaved consumers honor them -- but an RBAC-permitted reader can still fetch an expired value. Use expiry to make rotation auditable, and pair it with monitoring on near-expiry secrets rather than trusting it as a lock.

**Deleted names linger -- know your vault's posture** -- soft delete reserves a deleted secret's name for the vault's retention window. The module purges on destroy so the name frees immediately -- EXCEPT when the vault has purge protection, where the reservation holds for the full window by design. In purge-protected vaults, treat secret names as append-only: version values, don't recycle names. `name` and `keyVaultId` are both replace-on-change -- secret material never moves between vaults.

**Multi-line values need encoding** -- Key Vault strips raw newlines. Certificates-as-secrets, SSH keys, and JSON blobs should be base64-encoded, with the encoding recorded in `contentType` (e.g. `application/x-pem-file;base64`) so consumers know to decode. Key Vault stores the hint verbatim and never interprets it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureKeyVault** | `keyVaultId` | `status.outputs.key_vault_id` |
| **Any credential-emitting resource** (e.g. AzureStorageAccount) | `value` | `status.outputs.primary_connection_string` (or the source's sensitive output) |

### What This Kind Provides

After provisioning, `status.outputs` contains the identifiers consumers reference to read the secret -- the VALUE is never an output:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `versionless_id` | The versionless data-plane ID (`https://{vault}.vault.azure.net/secrets/{name}`) | The default consumer reference -- app-setting Key Vault references, CSI drivers; value updates propagate automatically |
| `secret_id` | The versioned data-plane ID (one frozen version) | Compliance-pinned consumers that must not follow rotation |
| `resource_id` | The versioned ARM resource ID | ARM-level integrations and RBAC scopes |

`secret_name` and `version` are also exported; the name echoes the manifest's `name`, and the version is the trailing segment of `secret_id`.

## Common Patterns

**The credential handoff** -- a resource emits a sensitive output (a connection string, an admin key), this component stores it in the vault, and applications read it at runtime with their own vault permissions; the credential never touches configuration files or pipeline variables. Start from the **Basic Secret** preset.

**Auditable rotation windows** -- a third-party API key stored with explicit `notBeforeDate` and `expirationDate` makes the credential's validity window auditable infrastructure (Azure Policy has a built-in check for missing expirations); update the dates with each rotation so the new version carries the new window. Start from the **Expiring API Key** preset.

**Runtime consumption via Key Vault references** -- App Service and Functions resolve `@Microsoft.KeyVault(SecretUri=...)` app settings against the versionless ID, so the app's configuration carries a pointer, not a credential.

## Works With

- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the vault the secret lives in
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- a typical value source (connection strings and access keys worth vaulting)
- [**Azure AI Search Service**](/infra-catalog/azure-search-service) -- another value source (admin keys under a rotation policy)
- [**Azure Function App Flex Consumption**](/infra-catalog/azure-function-app-flex-consumption) -- consumes the secret at runtime through Key Vault references in app settings

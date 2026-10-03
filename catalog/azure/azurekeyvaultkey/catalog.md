# Azure Key Vault Key

Deploys a cryptographic key inside an Azure Key Vault -- the customer-managed-key (CMK) building block of an Azure platform. The private part never leaves the vault (or its HSM on the Premium tier); consumers call the vault to encrypt/decrypt, wrap/unwrap, or sign/verify. This is how "bring your own key" works across Azure: storage accounts, disk encryption sets, container registries, and database services all encrypt their data with a data-encryption key that THIS key wraps -- revoking or rotating it revokes access to everything downstream.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Key Vault Key** -- in the referenced vault, with the chosen algorithm family (RSA or EC, optionally HSM-backed on a Premium vault) and strength (modulus size or curve)
- **Capability boundary** -- the permitted-operations list (`keyOpts`): Azure rejects any cryptographic operation not listed
- **Rotation policy** -- when the `rotationPolicy` block is configured: per-version expiry with Event Grid near-expiry notification, and automatic rotation triggers
- **Lifetime instants** -- optional not-before and expiration timestamps on the key

The key lives entirely inside the referenced vault, whose authorization mode, network posture, and purge protection govern it wholesale.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Key Vault** for the key to live in. Reference an AzureKeyVault Infra Component via ValueFromRef, or provide the vault ARM ID directly. HSM key types (RSA_HSM/EC_HSM) require the vault on the PREMIUM SKU.
- **Data-plane key permissions** for the deploying credential -- subscription Owner alone is not enough; the deployer needs data-plane key permissions on the vault (the "Key Vault Administrator" or "Key Vault Crypto Officer" RBAC role, or key permissions in a legacy access policy).

## Deploy

### Console

Open the deployment store, find **Azure Key Vault Key**, and click **Deploy**. The creation wizard walks you through placement (the vault and the key's name), the algorithm family and strength, the permitted operations with persona quick-picks, and the rotation policy. Start from the **RSA Customer-Managed Key** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureKeyVaultKey
metadata:
  name: storage-cmk
  org: acme-corp
  env: prod
spec:
  name: storage-cmk
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: platform-vault
      fieldPath: status.outputs.key_vault_id
  keyType: RSA
  keySize: 2048
  keyOpts:
    - WRAP_KEY
    - UNWRAP_KEY
  rotationPolicy:
    expireAfter: P2Y
    notifyBeforeExpiry: P30D
    automatic:
      timeAfterCreation: P83D
```

```shell
planton apply -f key-vault-key.yaml
```

This creates a rotating RSA 2048 CMK -- the shape every Azure CMK integration accepts, rotating roughly quarterly with versionless consumers following automatically. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, ValueFromRef wires the key to a vault deployed in the same InfraPipeline; downstream CMK consumers then reference this key's outputs:

```yaml
spec:
  keyVaultKeyId:
    valueFrom:
      kind: AzureKeyVaultKey
      name: storage-cmk
      fieldPath: status.outputs.versionless_id
```

The InfraPipeline resolves the dependency graph -- vault first, then the key, then everything the key encrypts.

## Key Configuration

These are the most important decisions when configuring a Key Vault key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Algorithm family** -- fixed at creation. RSA is the general-purpose choice every Azure CMK integration accepts (2048/3072/4096-bit modulus, 2048 the baseline); EC keys (P-256 default, P-256K/P-384/P-521 available) sign and verify but do NOT encrypt, so an EC key can never serve as a CMK. The _HSM variants keep the private key inside FIPS 140-2 Level 3 hardware and require the vault's PREMIUM SKU.

**Permitted operations** -- the key's capability boundary, enforced per call. CMK/envelope encryption needs `WRAP_KEY` + `UNWRAP_KEY`; direct encryption needs `ENCRYPT` + `DECRYPT`; signing needs `SIGN` + `VERIFY`. Grant only what the key's consumers actually perform; the list edits in place.

**Rotation policy** -- each rotation mints a new key VERSION; consumers referencing `versionless_id` follow with zero intervention. The policy pairs per-version expiry with a notification lead (both or neither) and fires on an age trigger (`timeAfterCreation`) or an expiry countdown (`timeBeforeExpiry`, which requires the expiry). Durations are ISO 8601 date durations (P90D, P2Y).

**Expiry** -- a hard cutoff: expired keys refuse all operations, so an expired CMK takes its dependents down with it. Prefer the rotation policy for keys encrypting live data; once set, expiry cannot be fully unset on the underlying key.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureKeyVault** | `keyVaultId` | `status.outputs.key_vault_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `versionless_id` | Versionless data-plane ID (`https://{vault}.vault.azure.net/keys/{name}`) | THE CMK reference -- storage accounts, disk encryption sets, Cosmos DB, container registries, flexible servers, and Service Bus follow rotation through it |
| `key_id` | Versioned data-plane ID of the current version | Compliance regimes that pin a frozen version (MSSQL TDE, managed Redis) |
| `key_name` | The key's name in the vault | Operator tooling |
| `version` | The current version identifier | Audit trails |
| `resource_id` | Versioned ARM resource ID | ARM-level references |
| `resource_versionless_id` | Versionless ARM resource ID | ARM-level references that follow rotation |
| `public_key_pem` | The public key, PEM-encoded | Signature verification outside Azure |
| `public_key_openssh` | The public key, OpenSSH-encoded | SSH-style tooling |

The key outputs no secret material -- the private part never leaves the vault.

## Common Patterns

**RSA customer-managed key** -- RSA 2048 with wrap/unwrap only, the shape every CMK-consuming service accepts. Granting nothing beyond `WRAP_KEY` + `UNWRAP_KEY` keeps the capability boundary tight: the key can enroll in CMK integrations but cannot be misused for direct encryption or signing. Start from the **RSA Customer-Managed Key** preset.

**HSM-backed auto-rotating production CMK** -- RSA-HSM with a full rotation policy (per-version expiry, notification lead, age-based rotation trigger), for compliance regimes mandating FIPS 140-2 Level 3 hardware. The trades: the vault must be on the PREMIUM SKU, and HSM operations carry higher per-operation cost. Start from the **HSM Auto-Rotating Production CMK** preset.

**EC signing key** -- an elliptic-curve key with sign/verify, for JWTs and code signing; smaller signatures and faster verification than RSA. Remember an EC key cannot encrypt -- it can never serve as a CMK, so pick this shape only for signature workloads. Start from the **EC Signing Key** preset.

## Works With

- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the parent vault whose governance this key inherits
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- bridges this key to server-side disk encryption
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- encrypts under this key via its customer-managed-key block
- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- Premium namespaces encrypt messaging data under this key
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants consumers crypto access on the vault

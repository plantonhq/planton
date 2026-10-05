# Azure Key Vault

Deploys an Azure Key Vault -- the tenant-scoped container where an organization's encryption keys, TLS certificates, and application secrets live behind one security boundary. The vault is where governance is set: authorization mode (Azure RBAC vs legacy access policies), network isolation, deletion safety (soft delete and purge protection), and the pricing tier that gates HSM-backed keys. What lives inside is composed, never bundled -- keys and certificates are first-class catalog kinds referencing this vault, and secret VALUES are deliberately out of scope for infrastructure-as-code, so plaintext never enters deployment manifests or state.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Key Vault** -- in the specified region and resource group, on the chosen tier (Standard or Premium), with the globally-unique DNS identity `https://{vault_name}.vault.azure.net/`
- **Authorization mode** -- Azure RBAC (the recommended posture, applied when unspecified) or the legacy access-policy mode with its per-principal grants
- **Deletion safety** -- the soft-delete retention window (7-90 days, fixed at creation) and, when enabled, irreversible purge protection
- **Network rules** -- when the `networkAcls` block is configured: a default action, the trusted-Microsoft-services bypass, an IP allowlist, and a VNet service-endpoint subnet allowlist
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

The vault deliberately contains no objects at creation: encryption keys (AzureKeyVaultKey) and TLS certificates (AzureKeyVaultCertificate) are their own kinds, each referencing this vault's `key_vault_id` output.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the vault will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef. Security foundations usually live in a dedicated group.
- **A globally unique vault name** -- 3-24 characters of letters, digits, and hyphens, starting with a letter, ending with a letter or digit, no consecutive hyphens. It becomes the endpoint `{name}.vault.azure.net`, and a deleted vault's name stays reserved for the soft-delete retention window.
- **For the legacy access-policy mode** -- the object IDs of the Azure AD principals being granted access (for an AzureUserAssignedIdentity, its `principal_id` output -- the directory object, not the client ID).

## Deploy

### Console

Open the deployment store, find **Azure Key Vault**, and click **Deploy**. The creation wizard walks you through placement, the access model (RBAC vs legacy policies with per-principal permission grants), the tier and deletion-safety dials, network access, and governance tags. Start from the **Standard RBAC Vault** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureKeyVault
metadata:
  name: platform-vault
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-security-rg"
  vaultName: acme-platform-kv
  purgeProtectionEnabled: true
  tags:
    purpose: cmk-root
```

```shell
planton apply -f key-vault.yaml
```

This creates a Standard-tier vault in RBAC mode with purge protection on (leaving `sku` and `rbacAuthorizationEnabled` out ships no opinion and applies Standard + RBAC); keys and certificates arrive as their own kinds afterward, referencing this vault. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the vault to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: security-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the vault with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Key Vault. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authorization mode** -- `rbacAuthorizationEnabled` unset (or true) runs Azure RBAC: grants become ordinary AzureRoleAssignment compositions with PIM, access reviews, and fine-grained scopes. An explicit `false` runs the legacy access-policy mode, the only state where the `accessPolicies` list is honored -- ARM stores but IGNORES policies on an RBAC vault. Azure's own guidance is RBAC for new vaults.

**Purge protection** -- off by default, and a ONE-WAY door: once enabled it can never be disabled. Turn it on for production vaults, and ALWAYS when this vault's keys encrypt other resources -- disk encryption sets, storage CMK, and Cosmos DB CMK refuse to enroll keys from a vault without it.

**Soft-delete retention** -- 7-90 days (unspecified applies Azure's 90). The one creation-time-only setting on the vault: changing it later replaces the vault. A deleted vault's name stays reserved for this window.

**Pricing tier** -- STANDARD (Azure's default when unspecified) covers software-protected keys, secrets, and certificates. PREMIUM exists for one feature: HSM-backed keys (FIPS 140-2 Level 3), required before any AzureKeyVaultKey in this vault can use the RSA_HSM/EC_HSM types. The SKU changes in place.

**Network posture** -- `publicNetworkAccessEnabled: false` removes the public endpoint entirely (pair it with an AzurePrivateEndpoint targeting subresource `vault`). For a public vault restricted to known networks, keep it enabled and configure `networkAcls` instead: Deny by default, admit office IP ranges and workload subnets, and decide whether trusted Microsoft services bypass the rules.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `accessPolicies[].objectId` | `status.outputs.principal_id` |
| **AzureSubnet** | `networkAcls.virtualNetworkSubnetIds[]` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `key_vault_id` | Azure Resource Manager ID of the vault | The parent reference for AzureKeyVaultKey and AzureKeyVaultCertificate, the scope for RBAC-mode AzureRoleAssignment grants, and the AzurePrivateEndpoint target (subresource: `vault`) |
| `key_vault_name` | Name of the vault | SDK configuration, monitoring dashboards |
| `vault_uri` | The vault's data-plane endpoint (`https://{name}.vault.azure.net/`) | Key Vault SDK configuration |
| `tenant_id` | The Azure AD tenant the vault was created in | Grant tooling and cross-checking compound-identity policies |
| `resource_group_name` | The resource group hosting the vault | Scoping related deployments |

The vault outputs no secrets -- it stores them. Consuming services reference the vault's CHILD resources: a key's `versionless_id` for customer-managed keys, a certificate's `versionless_secret_id` for TLS.

## Common Patterns

**Standard RBAC vault** -- the recommended baseline: RBAC authorization, purge protection on, grants composed as ordinary AzureRoleAssignment resources with PIM and access reviews. The trade against access policies: grants need Microsoft.Authorization write permission to manage, but they scale and audit like every other Azure grant. Start from the **Standard RBAC Vault** preset.

**Premium network-restricted CMK vault** -- HSM-capable tier with the rule set on DENY and admitted sources only, for customer-managed-key roots under compliance regimes demanding hardware protection and network isolation. Watch the bypass decision: NONE closes the door even to trusted Microsoft services, which breaks first-party integrations like Azure Backup. Start from the **Premium Network-Restricted CMK Vault** preset.

**Legacy access-policy vault** -- explicit per-principal permission lists for workloads and org standards that require the pre-RBAC model. Populating `accessPolicies` on an RBAC vault is almost always a mistake -- ARM stores but ignores them -- so set `rbacAuthorizationEnabled: false` deliberately. Start from the **Legacy Access-Policy Vault** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the vault is created
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- encryption keys living in this vault; the CMK source for the rest of the catalog
- [**Azure Key Vault Certificate**](/infra-catalog/azure-key-vault-certificate) -- TLS certificates living in this vault
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- data-plane grants scoped to the vault ARM ID in RBAC mode
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- admitted to the network rules via service endpoints
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- takes the vault off the public internet (subresource: `vault`)
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- bridges a vault key to server-side disk encryption

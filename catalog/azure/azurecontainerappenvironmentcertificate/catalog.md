# Azure Container App Environment Certificate

Stores a bring-your-own TLS certificate on a Container App Environment -- the certificate store of the Container Apps family. Certificates live on the ENVIRONMENT and are shared by every app in it: upload or reference the certificate once, then bind it to any app's hostname through an Azure Container App Custom Domain. This kind is for certificates you bring -- EV/OV chains, org-mandated CAs, wildcard certificates; for free, domain-validated certificates Azure provisions and renews end to end, use the managed-certificate kind instead. The certificate arrives exactly one way: a Key Vault reference the environment keeps current across renewals, or an inline PFX upload whose rotation is manual.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Environment Certificate** -- on the referenced Container App Environment, sourced exactly one way: a Key Vault reference the environment keeps current across renewals, or an inline PFX upload whose rotation is manual

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureContainerAppEnvironment** to store the certificate on. Reference its `environment_id` output via ValueFromRef.
- **For the Key Vault source**: an AzureKeyVaultCertificate whose `versionless_secret_id` output this certificate references, plus two contracts Azure checks at deploy time -- the reading identity (the environment's system-assigned identity by default) must be ON the environment's identity block, and it must hold read access to the vault's secrets (Key Vault Secrets User under RBAC).
- **For the inline PFX source**: the base64-encoded PKCS#12 bundle and its password stored as managed secrets -- both fields are secret-bearing and hold `$secret/<slug>` references, never pasted material.

## Deploy

### Console

Open the deployment store, find **Azure Container App Environment Certificate**, and click **Deploy**. The creation wizard walks you through the certificate identity (environment + the name every domain binding references), the source (a two-card Key-Vault-or-inline-PFX choice, with versionless-reference teaching and live prerequisite callouts), and tags. Start from the **Key Vault Certificate** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironmentCertificate
metadata:
  name: app-example-com-cert
  org: acme-corp
  env: prod
spec:
  certificateName: app.example.com
  containerAppEnvironmentId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.App/managedEnvironments/prod-apps-env"
  certificateKeyVault:
    keyVaultSecretId:
      value: "https://acme-prod-vault.vault.azure.net/secrets/app-example-com"
```

```shell
planton apply -f certificate.yaml
```

This stores a Key-Vault-sourced certificate named `app.example.com` on the environment, read with the environment's system-assigned identity and following vault renewals automatically. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the certificate to the environment and vault certificate deployed in the same InfraPipeline:

```yaml
spec:
  certificateName: app.example.com
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: apps-env
      fieldPath: status.outputs.environment_id
  certificateKeyVault:
    keyVaultSecretId:
      valueFrom:
        kind: AzureKeyVaultCertificate
        name: app-cert
        fieldPath: status.outputs.versionless_secret_id
```

The InfraPipeline resolves the dependency graph, deploys the environment and vault certificate first, then stores this certificate with the resolved values.

## Key Configuration

These are the most important decisions when configuring a certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Source** -- exactly one per certificate. The Key Vault reference (`certificateKeyVault`) is the recommended path: the environment reads the certificate's SECRET face from the vault and, with a VERSIONLESS reference, follows every renewal automatically. The inline PFX (`certificateBlobBase64` + `certificatePassword`, both secret-bearing) is the fallback for material outside a vault -- its rotation is manual, and Azure never returns the blob on reads, so the spec is the only record of what was uploaded.

**Certificate name** -- a segment of the ARM ID every domain binding references. Commonly the hostname (`app.example.com`) or a hyphenated wildcard form (`wildcard-example-com`). Only `tags` update in place; every other change replaces the certificate -- but the ARM ID rides the name, so replacing the certificate CONTENT under the same name keeps the ID, and every domain binding referencing it re-binds to the new material. That is the designed rotation workflow; renaming mints a different ID and strands the bindings.

**Vault read identity** -- which managed identity reads the vault secret. Unset deploys `System` (the environment's system-assigned identity, Azure's own default); reference an AzureUserAssignedIdentity's `identity_id` output to read with a shared fleet identity. Either way the identity must be on the environment and hold Key Vault Secrets User on the vault.

**PFX password** -- optional by design: passwordless PFX bundles are legal, and both IaC engines send the (empty) password argument Azure expects alongside the blob.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |
| **AzureKeyVaultCertificate** | `certificateKeyVault.keyVaultSecretId` (Key Vault source) | `status.outputs.versionless_secret_id` |
| **AzureUserAssignedIdentity** | `certificateKeyVault.identity` (optional) | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | Azure Resource Manager ID of the certificate | The domain-binding seam: AzureContainerAppCustomDomain references it in `containerAppEnvironmentCertificateId` |
| `subject_name` | The certificate's subject as read back from the material | Confirming the right certificate landed |
| `issuer` | The issuing CA | Compliance checks |
| `issue_date` | When the certificate was issued (RFC 3339) | Audit trails |
| `expiration_date` | When the certificate expires (RFC 3339) | The value to alarm on for inline-PFX certificates, whose rotation is manual |
| `thumbprint` | The SHA-1 fingerprint | Identifying the installed certificate in Azure and operational tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Vault-backed rotation** -- an AzureKeyVaultCertificate renews in the vault; this certificate's versionless reference follows it automatically, and every bound domain serves the current version with no manual step. Start from the **Key Vault Certificate** preset.

**Wildcard for the whole environment** -- one `wildcard-example-com` certificate stored once, referenced by any number of domain bindings (`app.example.com`, `api.example.com`) across the environment's apps.

**Org-mandated CA** -- an EV/OV chain exported as PFX, stored as managed secrets, uploaded inline. Put the `expiration_date` output on a calendar -- nothing else will notice it approaching. Start from the **Inline PFX Certificate** preset.

## Works With

- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- where the certificate is stored
- [**Azure Container App Custom Domain**](/infra-catalog/azure-container-app-custom-domain) -- binds the certificate to an app's hostname
- [**Azure Key Vault Certificate**](/infra-catalog/azure-key-vault-certificate) -- the vault-managed source whose renewals propagate
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- an optional shared identity for the vault read
- [**Azure Container App Environment Managed Certificate**](/infra-catalog/azure-container-app-environment-managed-certificate) -- the free, Azure-renewed alternative for single hostnames

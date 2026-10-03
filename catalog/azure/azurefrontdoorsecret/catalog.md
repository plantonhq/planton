# Azure Front Door Secret

Deploys a Front Door secret -- the bring-your-own TLS certificate node inside an Azure Front Door (Standard/Premium) profile. A secret wraps a Key Vault certificate so custom domains can terminate TLS with it: the domain's `tls.secretId` references this secret, and this secret references the AzureKeyVaultCertificate that actually holds the key material. The secret is a first-class resource rather than a field on the domain because one certificate -- typically a wildcard or multi-SAN cert -- serves many domains, and rotating it must be a single operation, never a per-domain edit.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Secret** -- a named child of the profile wrapping the referenced Key Vault certificate
- **Certificate binding** -- versionless (Front Door follows the certificate's latest version; Key Vault rotation propagates automatically) or version-pinned (one exact certificate ships until the secret is replaced)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Profile** the secret nests under.
- **A CA-issued Key Vault certificate with its complete chain** -- Azure rejects self-signed certificates ("the certificate chain includes an invalid number of certificates"). Use a certificate enrolled through a CA integration or an imported PKCS#12 carrying leaf plus issuer.
- **The one-time vault grant** -- Front Door reads Key Vault with Microsoft's own service principal (the `Microsoft.AzureFrontDoor-Cdn` enterprise application). Grant it read access on the vault -- e.g. the "Key Vault Secrets User" role on an RBAC-mode vault -- before the first secret deploys. Without it, Azure rejects the create with an access-denied error naming the vault.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Secret**, and click **Deploy**. The wizard walks you through the parent profile, the secret name, and the Key Vault certificate -- whose reference form (versionless vs versioned) decides the rotation story. Start from the **Rotating Bring-Your-Own Certificate** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorSecret
metadata:
  name: wildcard-cert-secret
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  secretName: wildcard-example-com
  keyVaultCertificateId:
    valueFrom:
      kind: AzureKeyVaultCertificate
      name: wildcard-example-com-cert
      fieldPath: status.outputs.versionless_id
```

```shell
planton apply -f front-door-secret.yaml
```

This creates a secret that follows the certificate's latest version, ready for custom domains to reference through the `secret_id` output. An Infra Job tracks the provisioning in real time.

### InfraChart

In an InfraChart, wire the secret between its profile and its certificate through `valueFrom`:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  secretName: wildcard-example-com
  keyVaultCertificateId:
    valueFrom:
      kind: AzureKeyVaultCertificate
      name: wildcard-example-com-cert
      fieldPath: status.outputs.versionless_id
```

The InfraPipeline resolves the dependency graph, provisioning the profile and the certificate before the secret that references them.

## Key Configuration

These are the most important decisions when configuring a Front Door secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Secret name** -- 2–260 characters; letters, digits, and hyphens; must start and end with a letter or digit. Unique within the profile. The name is a segment of the secret's ARM ID (the exact string domains reference), so renaming replaces the secret under a new ID.

**Versionless vs versioned reference** -- the single decision that sets the rotation story. The versionless certificate identifier (no trailing version segment, the default reference) tells Front Door to follow the certificate's LATEST version, so Key Vault rotation and auto-renewal propagate to every domain with zero redeploys. A versioned identifier pins one exact certificate: rotation then requires replacing this secret, which is what change-controlled environments and certificate-pinning clients want.

**Which certificate to wrap** -- the wrapped certificate must be CA-issued with a complete chain (leaf plus issuer, at least two certificates); Azure rejects self-signed certificates at deploy time. The reference is a Key Vault data-plane URL, not an ARM ID.

**Immutability** -- Azure exposes no update on Front Door secrets: changing any field replaces the secret. That is safe in practice, because certificate ROTATION happens inside Key Vault (new versions), not by editing the secret -- and domains reference the secret by ARM ID, which survives the replacement.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |
| **AzureKeyVaultCertificate** | `keyVaultCertificateId` | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_id` | ARM resource ID of the secret | AzureFrontDoorCustomDomain.`tls.secretId` |
| `secret_name` | The secret's name within its profile | Operator tooling |
| `subject_alternative_names` | The DNS names the wrapped certificate covers | Confirming a domain's host name is covered before attaching |

## Common Patterns

**Hands-off TLS with a rotating certificate** -- the default for every BYO certificate: reference the certificate's `versionless_id` and pair it with an auto-renewing AzureKeyVaultCertificate (issuer-managed, or self-signed with an AUTO_RENEW lifetime action). Key Vault renews, Front Door follows, every domain rotates -- zero redeploys. Start from the **Rotating Bring-Your-Own Certificate** preset.

**Pinned version for change-controlled rollout** -- reference the versioned `certificate_id` when certificate rollout must be an explicit, auditable deployment, or when clients pin the served certificate (mobile apps, partner allowlists) and an unannounced rotation is an outage. The trade: every rotation is a deliberate secret replacement you schedule. Start from the **Pinned Certificate Version** preset.

**One wildcard secret, many domains** -- point every tenant subdomain's `tls.secretId` at the same secret wrapping a wildcard or multi-SAN certificate. Rotation stays a single Key Vault operation regardless of how many domains serve the certificate; check the `subject_alternative_names` output to confirm coverage before attaching a new hostname.

## Works With

- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container the secret nests under via `profileId`
- [**Azure Key Vault Certificate**](/infra-catalog/azure-key-vault-certificate) -- holds the key material this secret wraps, referenced by `keyVaultCertificateId`
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the vault carrying the certificate; Front Door's service principal needs the one-time read grant on it
- [**Azure Front Door Custom Domain**](/infra-catalog/azure-front-door-custom-domain) -- terminates TLS with this secret via `tls.secretId` when its certificate type is CUSTOMER_CERTIFICATE

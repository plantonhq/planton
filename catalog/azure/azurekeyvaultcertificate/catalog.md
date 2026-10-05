# Azure Key Vault Certificate

Deploys an X.509 certificate inside an Azure Key Vault -- the TLS building block the vault manages end to end: private key custody, enrollment or import, renewal, and expiry notification. A vault certificate is really three linked objects sharing one name: the certificate (public part + policy), its private KEY, and a SECRET whose value is the full bundle. That third face is the composition seam -- Azure services that terminate TLS consume the certificate THROUGH its secret identifier, which is why the `secret_id` outputs are what downstream kinds reference.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Key Vault Certificate** -- in the referenced vault, either vault-GENERATED (self-signed or CA-issued from a policy) or IMPORTED from an existing PFX/PEM bundle
- **Issuance policy** -- when configured: the issuer, the private-key shape (RSA/EC, optionally HSM-backed, exportable or vault-bound), near-expiry actions (auto-renew or contact notification), and the secret encoding (PKCS#12 or PEM)
- **X.509 content** -- for generated certificates: the subject, SANs, key usages, EKU OIDs, and validity window renewals re-issue for

The certificate lives entirely inside the referenced vault, whose authorization mode, network posture, and deletion safety govern it wholesale.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Key Vault** for the certificate to live in. Reference an AzureKeyVault Infra Component via ValueFromRef. HSM-backed private keys require the vault on the PREMIUM SKU.
- **Data-plane certificate permissions** for the deploying credential -- the "Key Vault Administrator" or "Key Vault Certificates Officer" RBAC role, or certificate permissions in a legacy access policy.
- **For imports** -- the base64-encoded PFX/PEM bundle stored as a Planton org secret (it carries the private key; the plaintext never enters the manifest).
- **For CA issuance** -- a CA issuer (DigiCert, GlobalSign) configured on the vault out-of-band.

## Deploy

### Console

Open the deployment store, find **Azure Key Vault Certificate**, and click **Deploy**. The creation wizard walks you through placement, the generate-vs-import source choice (imported bundles ride org-secret references), the issuance policy, and -- for generated certificates -- the X.509 content with live SAN guidance. Start from the **Self-Signed Auto-Renewing Certificate** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureKeyVaultCertificate
metadata:
  name: web-tls
  org: acme-corp
  env: prod
spec:
  name: web-tls
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: platform-vault
      fieldPath: status.outputs.key_vault_id
  certificatePolicy:
    issuerName: Self
    keyProperties:
      exportable: true
      keyType: RSA
      keySize: 2048
    lifetimeActions:
      - actionType: AUTO_RENEW
        trigger:
          lifetimePercentage: 80
    secretProperties:
      contentType: PKCS12
    x509CertificateProperties:
      subject: CN=internal.example.com
      subjectAlternativeNames:
        dnsNames:
          - internal.example.com
      keyUsage:
        - DIGITAL_SIGNATURE
        - KEY_ENCIPHERMENT
      validityInMonths: 12
```

```shell
planton apply -f key-vault-certificate.yaml
```

This creates the fully-hands-off internal-TLS shape: self-signed, renewing itself at 80% lifetime, with consumers following each renewal through the versionless secret ID. An Infra Job tracks the provisioning in real time.

### InfraChart

TLS terminators wire to the certificate through its secret outputs in the same InfraPipeline:

```yaml
spec:
  keyVaultSecretId:
    valueFrom:
      kind: AzureKeyVaultCertificate
      name: web-tls
      fieldPath: status.outputs.versionless_secret_id
```

The InfraPipeline resolves the dependency graph -- vault, then certificate, then the gateway that serves it.

## Key Configuration

These are the most important decisions when configuring a Key Vault certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Generate vs import** -- generate (policy only) keeps the private key inside the vault from birth and enables automatic renewal; import (the `certificate` block) brings an existing PFX/PEM. An import may ALSO carry a policy to govern the imported material; without one, Azure derives it from the bundle.

**The issuer drives renewal** -- "Self" self-signs (internal TLS; auto-renew works); a configured CA issuer automates public issuance end to end; "Unknown" mints a CSR for an out-of-band CA -- renewal is inherently manual there, and only the contact-notification action applies.

**Exportability** -- whether the private key rides inside the secret face consumers read. TLS terminators that load the certificate themselves need `exportable: true`; keep it false when only vault-integrated services consume it.

**Near-expiry actions** -- one action per trigger, each firing on exactly ONE of lifetime-percentage or days-before-expiry. Auto-renew at 80% lifetime is the convention; actions update without minting a new certificate version.

**SANs over subject** -- modern TLS clients validate SANs, not the subject CN. Every hostname the certificate serves must appear in `dnsNames`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureKeyVault** | `keyVaultId` | `status.outputs.key_vault_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `versionless_secret_id` | Versionless SECRET-face ID (`https://{vault}.vault.azure.net/secrets/{name}`) | THE TLS-terminator reference -- Application Gateway and Container Apps environment certificates follow renewals through it |
| `secret_id` | Versioned secret-face ID of the current version | Consumers pinning one issued version (firewall-policy TLS inspection) |
| `certificate_id` | Versioned data-plane certificate ID | Front Door secrets and certificate-object consumers |
| `versionless_id` | Versionless data-plane certificate ID | Certificate-object references that follow renewal |
| `certificate_name` | The certificate's name in the vault | Operator tooling |
| `version` | The current version identifier | Audit trails |
| `thumbprint` | The SHA-1 thumbprint, hex-encoded | Pinning and inventory reconciliation |
| `resource_manager_id` | Versioned ARM resource ID | ARM-level references |
| `resource_manager_versionless_id` | Versionless ARM resource ID | ARM-level references that follow renewal |

## Common Patterns

**Self-signed auto-renew for internal TLS** -- the fully-hands-off shape: the vault creates, signs, and renews forever, and consumers on the versionless secret ID never notice a rotation. The trade: clients must trust the certificate explicitly, so this fits service-to-service TLS, never public hostnames. Start from the **Self-Signed Auto-Renewing Certificate** preset.

**Imported bundle for purchased certificates** -- brings an externally-issued PFX/PEM in without plaintext touching the manifest: `certificate.contents` and `certificate.password` are sensitive fields that take Planton org-secret references (`contents: $secret/web-cert-bundle`). Renewal is on you -- re-importing new contents mints a new version. Start from the **Imported Certificate Bundle** preset.

**Pending CSR for an out-of-band CA** -- the "Unknown"-issuer flow: the vault mints and holds a CSR for a CA that has no vault integration. Renewal is inherently manual, so the lifetime action must be EMAIL_CONTACTS -- pair it with a real renewal runbook, because nothing else will fire. Start from the **CA-Signed via Pending CSR** preset.

## Works With

- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the parent vault whose governance this certificate inherits
- [**Azure Application Gateway**](/infra-catalog/azure-application-gateway) -- terminates TLS with this certificate through its secret ID
- [**Azure Front Door Secret**](/infra-catalog/azure-front-door-secret) -- wraps this certificate (by its versionless ID) to serve Front Door custom domains
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the sibling kind for raw cryptographic keys (CMK)
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants TLS consumers secret-read access on the vault

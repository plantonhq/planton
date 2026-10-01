# GCP KMS Key

Deploys a Cloud KMS cryptographic key within an existing key ring for symmetric encryption (CMEK), asymmetric signing, asymmetric decryption, or MAC generation. The key supports automatic rotation, configurable protection levels (software, HSM, or external key manager), bring-your-own-key import, and scheduled destruction windows. A KMS key can never be deleted from GCP: destroy removes the versions (making data encrypted under them unrecoverable once the recovery window elapses), and the key object remains permanently in the ring — choose names and settings deliberately.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **KMS CryptoKey** -- a `kms.CryptoKey` in the specified key ring, configured with the chosen purpose, algorithm, protection level, and rotation schedule
- **Version Template** -- created only when `versionTemplate` is specified; controls the encryption algorithm and protection level (SOFTWARE, HSM, EXTERNAL, or EXTERNAL_VPC) for new key versions
- **Automatic Rotation** -- created only when `rotationPeriod` is set; generates a new primary CryptoKeyVersion at the specified interval (minimum 24 hours)
- **Cloud KMS API enablement** -- `cloudkms.googleapis.com` enabled in the key ring's project (never disabled on destroy — other keys may be actively encrypting production data)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A KMS key ring** where the key will be created. Provide the fully qualified key ring path (`projects/{project}/locations/{location}/keyRings/{name}`) directly or reference a GcpKmsKeyRing Cloud Resource via ValueFromRef. The module enables the Cloud KMS API itself.
- **An EkmConnection** (only for EXTERNAL_VPC protection) — the `cryptoKeyBackend` field names the external key manager connection reached through your VPC.

## Deploy

### Console

Open the deployment store, find **GCP KMS Key**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Symmetric CMEK Key** preset in the [Presets](#presets) tab to pre-populate a standard CMEK key with 90-day rotation.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKey
metadata:
  name: app-cmek-key
  org: acme-corp
  env: prod
spec:
  keyRingId:
    value: "projects/acme-prod-12345/locations/us-central1/keyRings/prod-encryption"
  keyName: app-data-cmek
  rotationPeriod: "7776000s"
```

```shell
planton apply -f gcp-kms-key.yaml
```

This creates a symmetric encryption key with 90-day automatic rotation using software-level protection. Purpose defaults to ENCRYPT_DECRYPT. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the key to a key ring deployed in the same InfraPipeline:

```yaml
spec:
  keyRingId:
    valueFrom:
      kind: GcpKmsKeyRing
      name: prod-encryption
      fieldPath: status.outputs.key_ring_id
```

The InfraPipeline resolves the dependency graph, deploys the key ring first, then provisions the KMS key with the resolved key ring path.

## Key Configuration

These are the most important decisions when configuring a KMS key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Purpose** -- Set `purpose` to define what cryptographic operations the key supports. ENCRYPT_DECRYPT (default) is used for CMEK across BigQuery, Spanner, Cloud SQL, GCS, and GKE. ASYMMETRIC_SIGN is used for artifact and container image signing. Purpose is immutable after creation.

**Protection level** -- Set `versionTemplate.protectionLevel` to SOFTWARE (the default when unset), HSM, EXTERNAL, or EXTERNAL_VPC. HSM keys are protected by Cloud HSM hardware modules certified to FIPS 140-2 Level 3, required for PCI DSS, HIPAA, and FedRAMP compliance. The EXTERNAL levels keep the material in your own external key manager (EXTERNAL_VPC pairs with `cryptoKeyBackend`, the EkmConnection reached through your VPC). The protection level is immutable after creation.

**Bring your own key (BYOK)** -- Set `importOnly: true` (with `skipInitialVersionCreation: true` -- GCP cannot generate versions for an import-only key) when regulation requires that Google never generates the key material. Import versions after deploy with `gcloud kms keys versions import`. Start from the **Import-Only BYOK Key** preset.

**Labels** -- `labels` attach key-value metadata for inventory filtering and cost attribution across a key fleet; freely mutable.

**Rotation period** -- Set `rotationPeriod` for automatic key version rotation (e.g., `"7776000s"` for 90 days). Only applies to ENCRYPT_DECRYPT keys. Asymmetric keys require manual version management. Omit to disable automatic rotation.

**Destroy scheduled duration** -- `destroyScheduledDuration` controls how long destroyed key versions remain recoverable (default 30 days, minimum 24 hours). Shorter durations reduce the recovery window but limit accidental-destruction protection.

**Deletion policy** -- the key object itself can never be deleted from GCP; `deletionPolicy` governs what destroy does to its VERSIONS. `DELETE` (the default) destroys every version — data encrypted under them becomes unrecoverable once the recovery window elapses. `PREVENT` fails the destroy outright; `ABANDON` removes the key from management with versions intact and data still decryptable. Production CMEK keys warrant `PREVENT` or `ABANDON`.

**Immutability** -- only `rotationPeriod`, `versionTemplate.algorithm`, and `labels` update in place. Every other field — purpose, protection level, `importOnly`, the ring — is immutable, which for an undeletable resource means "abandon and create under a new name".

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpKmsKeyRing** | `keyRingId` | `status.outputs.key_ring_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `key_id` | Fully qualified crypto key path (`projects/{p}/locations/{l}/keyRings/{kr}/cryptoKeys/{name}`) | CMEK references in BigQuery, Spanner, Cloud SQL, GCS, GKE; GcpKmsKeyIamMember `cryptoKeyId` grants |
| `key_name` | Short name of the key | Display, logging, human-readable references |
| `primary_version_name` | The current primary CryptoKeyVersion | Populated only for symmetric encryption keys; empty for asymmetric/MAC/import-only keys |
| `primary_state` | Lifecycle state of the primary version (e.g. `ENABLED`) | Health checks on the encryption path; same population rules as the version name |
| `initial_version_name` | The version created with the key (`.../cryptoKeyVersions/1`) | Asymmetric-sign consumers: `GcpPrivateCaCertificateAuthority.keySpec.cloudKmsKeyVersion`, `GcpBinaryAuthorizationAttestor` PKIX keys; empty with `skip_initial_version_creation` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Symmetric encryption (CMEK)** -- Standard customer-managed encryption key with 90-day rotation and software protection. The most common pattern for encrypting data at rest across GCP services. Start from the **Symmetric CMEK Key** preset.

**HSM-protected symmetric encryption** -- Same CMEK pattern but with HSM protection level for compliance scenarios requiring FIPS 140-2 Level 3 certification (PCI DSS, HIPAA, FedRAMP). Start from the **HSM-Protected Symmetric Key** preset.

**Asymmetric signing** -- ECDSA P-256 key for signing build artifacts, container images, JWTs, or code. No automatic rotation -- key versions are managed manually. Start from the **Asymmetric Signing Key** preset.

## Works With

- [**GCP KMS Key Ring**](/cloud-catalog/gcp-kms-key-ring) -- provides the key ring that contains this cryptographic key
- [**GCP KMS Key IAM Member**](/cloud-catalog/gcp-kms-key-iam-member) -- grants service agents and workloads permission to use this key (CMEK consumers need `cryptoKeyEncrypterDecrypter`)
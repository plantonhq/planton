# GCP KMS Key

Deploys a Cloud KMS cryptographic key (`google_kms_crypto_key`) within an existing key ring. The key performs the actual cryptographic operations — symmetric encryption/decryption (CMEK), asymmetric signing, asymmetric decryption, raw encryption, or MAC generation — and is the resource every CMEK-capable GCP service references for encryption with keys you control.

## Overview

Three properties define the operational model:

- **Undeletable** — keys can never be removed from GCP. Destroy destroys all key *versions* (data encrypted under them becomes unrecoverable once the recovery window elapses) and disables rotation, but the key object remains permanently in its ring, and its name is never reusable there.
- **Version-centric** — the key is a container of CryptoKeyVersions; the *primary* version encrypts new data, old versions keep decrypting existing data. Automatic rotation mints new primaries on a cadence without re-encrypting anything.
- **Mostly immutable** — only `rotationPeriod`, `versionTemplate.algorithm`, and `labels` update in place. Everything else is fixed at creation, which for an undeletable resource means "abandon and create under a new name".

## When to Use

- **Customer-managed encryption (CMEK)** for BigQuery, Spanner, Cloud SQL, AlloyDB, GKE, Cloud Run/Functions, Pub/Sub, GCS, Vertex AI, Filestore, Bigtable, Firestore, Memorystore, Dataproc, Composer
- **Asymmetric signing** for artifact/container signing, JWT issuance, Binary Authorization attestors
- **HSM protection** for FIPS 140-2 Level 3 compliance
- **BYOK / external key managers** via import-only keys and EXTERNAL/EXTERNAL_VPC protection levels

## Prerequisites

- An existing key ring (see [GcpKmsKeyRing](../gcpkmskeyring/)); the key inherits its project and location, and the Cloud KMS API is enabled automatically
- Deployer permissions: the least-privilege permission set the IaC runner's principal needs lives in [`iac/permissions.yaml`](iac/permissions.yaml). Consumers additionally need `roles/cloudkms.cryptoKeyEncrypterDecrypter` granted to their service agents

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKey
metadata:
  name: cmek-data-key
spec:
  keyRingId:
    valueFrom:
      kind: GcpKmsKeyRing
      name: prod-encryption
      fieldPath: status.outputs.key_ring_id
  keyName: cmek-data-key
  rotationPeriod: "7776000s" # 90 days
```

This creates a symmetric encryption key with 90-day automatic rotation — the most common CMEK configuration.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyRingId` | StringValueOrRef | Yes | Fully qualified ring path; reference a `GcpKmsKeyRing`. Immutable |
| `keyName` | string | Yes | Key name (1-63 chars: `[a-zA-Z0-9_-]`). Immutable, never reusable within the ring |
| `purpose` | string | No | `ENCRYPT_DECRYPT` (default), `ASYMMETRIC_SIGN`, `ASYMMETRIC_DECRYPT`, `RAW_ENCRYPT_DECRYPT`, `MAC`, or any newer API purpose. Immutable |
| `rotationPeriod` | string | No | Auto-rotation cadence, ≥ `86400s` (e.g. `"7776000s"`). ENCRYPT_DECRYPT keys only. Mutable |
| `destroyScheduledDuration` | string | No | Recovery window for destroyed versions (default 30 days). Immutable |
| `versionTemplate.algorithm` | string | Within template | Algorithm for new versions (see the [algorithm reference](https://cloud.google.com/kms/docs/reference/rest/v1/CryptoKeyVersionAlgorithm)). Mutable — affects future versions |
| `versionTemplate.protectionLevel` | string | No | `SOFTWARE` (default), `HSM`, `EXTERNAL`, `EXTERNAL_VPC`. Immutable |
| `skipInitialVersionCreation` | bool | No | Create the key empty (required for import-only keys). Create-time only |
| `importOnly` | bool | No | BYOK container: only imported versions ever. Immutable; requires `skipInitialVersionCreation` |
| `cryptoKeyBackend` | string | No | Full name of the EKM connection (made outside the catalog) backing `EXTERNAL_VPC` keys. Immutable |
| `labels` | map | No | User labels, merged beneath Planton's attribution labels. Mutable |
| `deletionPolicy` | string | No | `DELETE` (default: destroy schedules EVERY key version for destruction — data becomes unrecoverable), `PREVENT` (destroy fails — the safe posture for production keys), `ABANDON` (the key leaves management with all versions intact). Mutable |

Cross-field rules enforced before deploy: rotation only on ENCRYPT_DECRYPT keys; `importOnly` requires `skipInitialVersionCreation`; `cryptoKeyBackend` requires `EXTERNAL_VPC` protection.

## Outputs

| Output | Description |
|--------|-------------|
| `key_id` | Fully qualified key path (`projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{name}`) — the CMEK reference every consumer takes |
| `key_name` | The short name of the key |
| `primary_version_name` | Resource name of the current primary version (ENCRYPT_DECRYPT keys; empty otherwise) |
| `primary_state` | Lifecycle state of the primary version (e.g. `ENABLED`) — the quick health probe that the key can encrypt |
| `initial_version_name` | Resource name of the version created with the key (version 1), for every purpose — the version an asymmetric-sign key's consumers (a CA Service authority, a Binary Authorization attestor) name; empty with `skip_initial_version_creation` |

## Important Notes

- **Grant before use**: each consuming service's agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key, or CMEK operations fail at the consumer.
- **Rotation is forward-only**: it limits blast radius for new data; existing data stays under its original version until rewritten.
- **Destroy has teeth**: destroying this resource schedules every version for destruction. Once the recovery window (`destroyScheduledDuration`, default 30 days) elapses, data encrypted under those versions is permanently unrecoverable.

## Related Kinds

- [GcpKmsKeyRing](../gcpkmskeyring/) — the parent container (required)
- [GcpBigQueryDataset](../gcpbigquerydataset/), [GcpSpannerDatabase](../gcpspannerdatabase/), [GcpCloudSql](../gcpcloudsql/), [GcpGkeCluster](../gcpgkecluster/), [GcpPubSubTopic](../gcppubsubtopic/) — CMEK consumers referencing `key_id`
- [GcpProjectIamMember](../gcpprojectiammember/) — models the encrypter/decrypter grants as first-class nodes

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

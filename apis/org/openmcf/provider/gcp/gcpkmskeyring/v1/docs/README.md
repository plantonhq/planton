# GCP KMS Key Ring — Research & Design Documentation

## Overview

Cloud Key Management Service (Cloud KMS) is Google Cloud's centralized key management service that lets you create, use, rotate, and destroy cryptographic keys. Cloud KMS integrates with virtually every GCP service that supports customer-managed encryption keys (CMEK), including BigQuery, Cloud SQL, Spanner, GKE, Compute Engine, Cloud Storage, and more.

A **key ring** is the top-level organizational unit in Cloud KMS. It serves as a logical grouping of cryptographic keys within a specific GCP project and location. Key rings have no cryptographic properties themselves — they exist purely for organization and access control scoping.

## GCP Resource Hierarchy

```
GCP Project
  └── KMS Location (region / multi-region / global)
        └── Key Ring (organizational container)
              ├── CryptoKey (ENCRYPT_DECRYPT)
              ├── CryptoKey (ASYMMETRIC_SIGN)
              ├── CryptoKey (MAC)
              └── CryptoKey (ASYMMETRIC_DECRYPT)
                    └── CryptoKeyVersion (actual key material)
```

## Key Architectural Decisions

### Why Key Ring Is a Separate Resource (Not Bundled with CryptoKey)

Key rings and crypto keys have fundamentally different lifecycles:

1. **One-to-many relationship**: A single key ring typically contains multiple CryptoKeys with different purposes (data encryption, signing, MAC).
2. **IAM scoping**: Key rings are an IAM boundary. You can grant `roles/cloudkms.cryptoKeyEncrypterDecrypter` at the key ring level to allow access to all keys within it, or at the individual key level.
3. **Permanence**: Key rings cannot be deleted. They are permanent fixtures. CryptoKeys can be scheduled for destruction (with recovery windows).
4. **Different creation frequency**: You create a key ring once per project/location/purpose combination, then create many keys within it over time.

Bundling them would force users to declare all keys upfront, which is impractical.

### Immutability

All three fields (`project`, `name`, `location`) are immutable after creation. This is enforced by the GCP API — there is no update endpoint for key rings. In Terraform/Pulumi, all fields are `ForceNew`.

If a user changes any field, the IaC engine will attempt to destroy and recreate. Since key rings cannot be deleted, this creates a new key ring alongside the original (which becomes orphaned from IaC state but still exists in GCP).

### No Deletion Support

This is a critical and unusual property of KMS key rings:

- The GCP API has no `DELETE` endpoint for key rings
- Terraform's `Delete` function only removes the resource from state
- Pulumi's `Delete` function only removes the resource from state
- The key ring continues to exist in GCP permanently

This means:
- Key ring names are consumed permanently within a project+location
- Reusing a name will reference the existing key ring (Terraform import / Pulumi refresh)
- There is no way to "clean up" unused key rings except by deleting the entire GCP project

### No Labels Support

Unlike most GCP resources, KMS key rings do not support resource labels. This is a GCP API limitation. The Pulumi module still computes OpenMCF-standard labels in `locals.go` for internal tracking, but they are not applied to the GCP resource.

## Deployment Landscape

### Methods Compared

| Method | Maturity | OpenMCF Value-Add |
|--------|----------|-------------------|
| `gcloud kms keyrings create` | Stable CLI | OpenMCF provides declarative YAML, cross-resource references, and infra-chart composition |
| Terraform `google_kms_key_ring` | Stable, widely used | OpenMCF adds validation, presets, and dependency-aware deployment |
| Pulumi `kms.KeyRing` | Stable Go SDK | OpenMCF wraps with consistent KRM API and cross-provider patterns |
| GCP Console | Point-and-click | No IaC, no reproducibility |

### When to Use OpenMCF for Key Rings

- You need declarative, version-controlled key ring management
- Your key rings are part of a larger infrastructure setup (infra charts)
- You want consistent cross-resource references (project IDs via `valueFrom`)
- You want validation before deployment (CEL rules catch naming errors early)

## 80/20 Scoping Rationale

KMS key rings are intentionally minimal — only 3 fields. This reflects the GCP API exactly:

**Included (100% of key ring functionality):**
- Project ID
- Key ring name
- Location

**Not applicable / excluded:**
- Labels (not supported by GCP API)
- Access control (managed via IAM, not the key ring resource itself)
- Crypto keys (separate resource with independent lifecycle)
- Key ring metadata/description (not supported by GCP API)

There are no "80/20" cuts here — the spec covers the complete surface area of the `google_kms_key_ring` resource.

## Location Strategy

GCP KMS supports three categories of locations:

### Regional Locations
Keys stored in a single GCP region. Data at rest never leaves that region.

**Best for:** Regulatory compliance (GDPR, HIPAA data residency), co-location with regional workloads, lowest latency.

**Examples:** `us-central1`, `europe-west1`, `asia-east1`, `us-east4`

### Multi-Region Locations
Keys replicated across multiple regions within a geographic boundary.

**Best for:** High availability with continental data residency, disaster recovery.

**Examples:** `us` (all US regions), `europe` (all EU regions), `asia` (all Asia regions)

### Global Location
Keys accessible from any region without geographic restrictions.

**Best for:** Cross-region workloads, global services, applications without data residency requirements.

**Example:** `global`

## Integration with CMEK (Customer-Managed Encryption Keys)

The primary purpose of creating key rings is to eventually create CryptoKeys within them for CMEK. The dependency chain:

```
GcpKmsKeyRing → GcpKmsCryptoKey → CMEK consumers (BigQuery, Spanner, GKE, etc.)
```

Services that support CMEK and would reference CryptoKeys in this key ring:

| Service | Field That Takes CryptoKey |
|---------|---------------------------|
| BigQuery Dataset | `encryption_configuration.kms_key_name` |
| Cloud SQL Instance | `encryption_key_name` |
| Spanner Database | `encryption_config.kms_key_name` |
| GKE Cluster | `database_encryption.key_name` |
| Compute Engine Disk | `disk_encryption_key.kms_key_self_link` |
| Cloud Storage Bucket | `encryption.default_kms_key_name` |
| Pub/Sub Topic | `kms_key_name` |
| Filestore Instance | `kms_key_name` |

## Best Practices

1. **Use a dedicated project for encryption keys** — separation of duties between key management and workload management.
2. **Co-locate key rings with workloads** — use regional locations that match your workload regions for lowest latency.
3. **Use descriptive, permanent names** — since key rings cannot be deleted, choose names that will make sense long-term.
4. **One key ring per environment per region** — e.g., `prod-encryption-us-central1`, `staging-encryption-europe-west1`.
5. **Grant IAM at the key ring level** — simpler to manage than per-key IAM for most use cases.

## References

- [Cloud KMS Documentation](https://cloud.google.com/kms/docs)
- [Creating Key Rings](https://cloud.google.com/kms/docs/create-key-ring)
- [KMS Locations](https://cloud.google.com/kms/docs/locations)
- [CMEK Overview](https://cloud.google.com/kms/docs/cmek)
- [Terraform google_kms_key_ring](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_key_ring)
- [Pulumi gcp.kms.KeyRing](https://www.pulumi.com/registry/packages/gcp/api-docs/kms/keyring/)

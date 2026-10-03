# GCP KMS Key Ring

Deploys a Cloud KMS key ring (`google_kms_key_ring`) — the permanent organizational container for cryptographic keys. Key rings belong to a GCP project and location (region, multi-region, or `global`) and group the `GcpKmsKey` resources that perform the actual encryption, signing, and MAC operations.

## Overview

A key ring carries no encryption policy itself; it exists to group and scope keys. Two properties make it the natural blast-radius boundary:

- **IAM flows down** — a grant on the ring applies to every key inside it, so one ring per environment or per data domain keeps access reviews tractable. Do not create one ring per key.
- **Location anchors CMEK** — most services require their CMEK key in the same location as the data it protects, and keys inherit the ring's location. Place rings where the data lives.

## Critical: Key Rings Cannot Be Deleted

**GCP does not support deletion of KMS key rings.** Once created, a key ring exists permanently (and at no cost) in the project and location. Destroying this resource removes the ring from IaC state only — it does **not** delete it from Google Cloud, and the name can never be reused for a "fresh" ring. Choose names and locations as permanent decisions.

## What Gets Created

- **Cloud KMS API enablement** (`cloudkms.googleapis.com`, with `disable_on_destroy=false` so teardown never disables the API project-wide)
- **KMS Key Ring** — a `google_kms_key_ring` resource in the specified project and location

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** (the Cloud KMS API is enabled automatically)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs

## Quick Start

Create a file `key-ring.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyRing
metadata:
  name: prod-encryption
spec:
  keyRingName: prod-encryption
  location: us-central1
```

Deploy:

```shell
planton apply -f key-ring.yaml
```

With `projectId` omitted, the ring is created in the provider connection's default project. Reference a `GcpProject` (or set a literal ID) to target another project.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | `StringValueOrRef` | No | GCP project for the ring. Omitted rides the provider default; reference a `GcpProject` to compose. Immutable. |
| `keyRingName` | `string` | Yes | The permanent GCP name (1-63 chars: letters, digits, hyphens, underscores). Immutable and never reusable within the project+location. |
| `location` | `string` | Yes | Region (`us-central1`), multi-region (`us`, `europe`, `asia`), or `global`. Immutable. |

### All Fields Are Immutable

Every field is ForceNew. Because rings cannot be deleted, a change abandons the original ring (it stays in GCP) and creates a new one.

## Choosing a Location

| Location Type | Examples | When to Use |
|---------------|----------|-------------|
| Regional | `us-central1`, `europe-west12` | CMEK for regional data services; data residency; lowest latency |
| Multi-region | `us`, `europe`, `asia` | CMEK for multi-region BigQuery/Spanner; continental residency |
| Global | `global` | Keys consumed from many regions (e.g. signing keys); no residency constraint |

Run `gcloud kms locations list` for the full list.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `key_ring_id` | `string` | Fully qualified ring path (`projects/{project}/locations/{location}/keyRings/{name}`) — the exact string a `GcpKmsKey`'s `keyRingId` reference consumes |
| `key_ring_name` | `string` | The short name of the ring |
| `location` | `string` | The location the ring resides in, for consumers that take a bare name + location pair |

## Important Notes

- **Permanent resource**: rings cannot be deleted; names are consumed forever within the project+location.
- **No labels**: the key ring API has no labels surface — no attribution labels are stamped, identically on both engines.
- **Re-creating a same-named ring**: creation against an existing name fails (the ring is still there); import the existing ring instead.

## Related Kinds

- [GcpKmsKey](../gcpkmskey/) — the cryptographic keys grouped by this ring
- [GcpProject](../gcpproject/) — provides the GCP project by reference

## Additional Resources

- [Cloud KMS Overview](https://cloud.google.com/kms/docs/overview)
- [Creating Key Rings](https://cloud.google.com/kms/docs/create-key-ring)
- [KMS Locations](https://cloud.google.com/kms/docs/locations)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

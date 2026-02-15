# GCP KMS Key Ring

Creates an organizational container for cryptographic keys in Google Cloud KMS. A key ring groups CryptoKeys by project and location, providing IAM scoping and logical organization for encryption, signing, and MAC operations across your GCP infrastructure.

**Important:** Key rings cannot be deleted from GCP. Once created, a key ring is permanent. Choose names and locations carefully.

## What Gets Created

When you deploy a GcpKmsKeyRing resource, OpenMCF provisions:

- **KMS Key Ring** — a `google_kms_key_ring` resource in the specified project and location, serving as the container for CryptoKeys

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **Cloud KMS API enabled** (`cloudkms.googleapis.com`) on the target project
- **IAM permissions** — `roles/cloudkms.admin` or `roles/cloudkms.keyRingCreator` on the target project

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: prod-encryption
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.prod-encryption
spec:
  projectId:
    value: my-gcp-project-123
  keyRingName: prod-encryption
  location: us-central1
```

Deploy:

```shell
openmcf apply -f key-ring.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `projectId` | `StringValueOrRef` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. |
| `keyRingName` | `string` | Name of the key ring (1-63 chars: letters, digits, hyphens, underscores). Permanent. |
| `location` | `string` | GCP location: region (`us-central1`), multi-region (`us`, `europe`), or `global`. Permanent. |

All fields are immutable after creation. Any change triggers a destroy-and-recreate cycle.

## Examples

### Regional Key Ring

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: prod-keys
spec:
  projectId:
    value: my-prod-project-123
  keyRingName: prod-encryption
  location: us-central1
```

### Global Key Ring

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: shared-keys
spec:
  projectId:
    value: my-security-project-456
  keyRingName: global-shared-keys
  location: global
```

### Multi-Region Key Ring (EU Compliance)

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: eu-compliance-keys
spec:
  projectId:
    value: my-eu-project-789
  keyRingName: eu-compliance-keys
  location: europe
```

### Cross-Resource Reference

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: data-encryption
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-project
      fieldPath: status.outputs.project_id
  keyRingName: data-encryption
  location: us-central1
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `keyRingId` | `string` | Fully qualified resource path (`projects/{project}/locations/{location}/keyRings/{name}`). Primary reference for GcpKmsCryptoKey. |
| `keyRingName` | `string` | Short name of the key ring. |

## Location Guide

| Type | Examples | Data Residency | Availability |
|------|----------|----------------|--------------|
| Regional | `us-central1`, `europe-west1` | Single region | Single region |
| Multi-region | `us`, `europe`, `asia` | Continental | Cross-region within continent |
| Global | `global` | None | All regions |

## Related Components

- [GcpKmsCryptoKey](/docs/catalog/gcp/gcpkmscryptokey) — creates encryption keys within this key ring for CMEK
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the key ring is created
- [GcpBigQueryDataset](/docs/catalog/gcp/gcpbigquerydataset) — uses CryptoKeys for customer-managed encryption
- [GcpSpannerInstance](/docs/catalog/gcp/gcpspannerinstance) — uses CryptoKeys for CMEK encryption
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — uses CryptoKeys for CMEK encryption

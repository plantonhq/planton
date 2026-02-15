---
title: "KMS Key Ring"
description: "KMS Key Ring deployment documentation"
icon: "package"
order: 100
componentName: "gcpkmskeyring"
---

# GCP KMS Key Ring

Creates an organizational container for cryptographic keys in Google Cloud KMS. A key ring groups CryptoKeys by project and location, providing IAM scoping and logical organization for encryption, signing, and MAC operations. Key rings are permanent resources — once created, they cannot be deleted from GCP.

## What Gets Created

When you deploy a GcpKmsKeyRing resource, OpenMCF provisions:

- **KMS Key Ring** — a `google_kms_key_ring` resource in the specified project and location, serving as the parent container for CryptoKeys

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **Cloud KMS API enabled** (`cloudkms.googleapis.com`) on the target project
- **IAM permissions** — `roles/cloudkms.admin` or `roles/cloudkms.keyRingCreator` on the target project

## Quick Start

Create a file `key-ring.yaml`:

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

This creates a KMS key ring named `prod-encryption` in the `us-central1` region, ready to hold CryptoKeys for encryption, signing, or MAC operations.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the key ring is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `keyRingName` | `string` | Name of the key ring in GCP. Permanent — cannot be changed or reused after creation. | 1-63 chars: letters, digits, hyphens, underscores. Pattern: `^[a-zA-Z0-9_-]{1,63}$` |
| `location` | `string` | GCP location where the key ring resides. Permanent — cannot be changed after creation. Use a region (`us-central1`), multi-region (`us`, `europe`, `asia`), or `global`. | Required |

### Optional Fields

This component has no optional fields. All three fields are required and immutable after creation.

## Examples

### Regional Key Ring

A key ring scoped to a single GCP region, suitable for most production workloads:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: prod-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.prod-keys
spec:
  projectId:
    value: my-prod-project-123
  keyRingName: prod-encryption
  location: us-central1
```

### Global Key Ring

A key ring available in all GCP regions, useful for cross-region encryption or when data residency is not a constraint:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: shared-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.shared-keys
spec:
  projectId:
    value: my-security-project-456
  keyRingName: global-shared-keys
  location: global
```

### Multi-Region Key Ring (EU Compliance)

A key ring scoped to the `europe` multi-region for data residency compliance:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: eu-compliance-keys
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.eu-compliance-keys
spec:
  projectId:
    value: my-eu-project-789
  keyRingName: eu-compliance-keys
  location: europe
```

### Using Foreign Key References

Reference a project ID from a GcpProject resource instead of hardcoding:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKeyRing
metadata:
  name: data-encryption
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpKmsKeyRing.data-encryption
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-project
      field: status.outputs.project_id
  keyRingName: data-encryption
  location: us-central1
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `key_ring_id` | `string` | Fully qualified resource path (e.g., `projects/my-project/locations/us-central1/keyRings/prod-encryption`). This is the primary reference used by GcpKmsKey and other downstream resources. |
| `key_ring_name` | `string` | Short name of the key ring as it exists in GCP. |

## Related Components

- [GcpKmsKey](/docs/catalog/gcp/kms-key) — creates encryption keys within this key ring for CMEK
- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the key ring is created
- [GcpBigQueryDataset](/docs/catalog/gcp/bigquery-dataset) — uses CryptoKeys from this key ring for customer-managed encryption
- [GcpCloudSql](/docs/catalog/gcp/cloud-sql) — uses CryptoKeys from this key ring for CMEK encryption

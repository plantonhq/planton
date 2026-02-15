# GCP KMS Key Ring

Deploys a GCP Cloud KMS key ring (`google_kms_key_ring`) — an organizational container for cryptographic keys. Key rings belong to a specific GCP project and location (region, multi-region, or `global`), and serve as the top-level grouping for CryptoKeys used in encryption, signing, and MAC operations.

## Critical: Key Rings Cannot Be Deleted

**GCP does not support deletion of KMS key rings.** Once created, a key ring exists permanently in the project and location. Destroying this OpenMCF resource only removes the key ring from the IaC state — it does **not** delete the key ring from Google Cloud. Re-creating a key ring with the same name and location will reference the existing one.

Choose your key ring names and locations carefully. They are permanent decisions.

## What Gets Created

When you deploy a GcpKmsKeyRing resource, OpenMCF provisions:

- **KMS Key Ring** — a `google_kms_key_ring` resource in the specified project and location

No additional supporting resources (API enablement, IAM bindings, etc.) are created. The module assumes the Cloud KMS API (`cloudkms.googleapis.com`) is already enabled on the target project.

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

This creates a key ring in the `us-central1` region where you can subsequently create CryptoKeys for encryption at rest, envelope encryption, CMEK, digital signatures, and more.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the key ring is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `keyRingName` | `string` | Name of the key ring in GCP. Permanent — cannot be renamed or deleted. | 1-63 chars: letters (upper/lower), digits, hyphens, underscores |
| `location` | `string` | GCP location where the key ring resides. Permanent — cannot be changed. | Required. Region (e.g., `us-central1`), multi-region (`us`, `europe`, `asia`), or `global` |

### All Fields Are Immutable

Every field in this spec is immutable after creation. Any change triggers a destroy-and-recreate cycle. Since key rings cannot be deleted from GCP, this effectively creates a new key ring alongside the original (which remains orphaned).

## Choosing a Location

| Location Type | Examples | When to Use |
|---------------|----------|-------------|
| Regional | `us-central1`, `europe-west1`, `asia-east1` | Data residency requirements, lowest latency for regional workloads |
| Multi-region | `us`, `europe`, `asia` | High availability across a continent, no single-region constraint |
| Global | `global` | Keys needed from any region, no data residency requirements |

Run `gcloud kms locations list` for a complete list of valid locations.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `keyRingId` | `string` | Fully qualified key ring resource path (`projects/{project}/locations/{location}/keyRings/{name}`). This is the primary reference used by GcpKmsCryptoKey. |
| `keyRingName` | `string` | The short name of the key ring. |

## Deployment Methods

OpenMCF supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Permanent resource**: Key rings cannot be deleted from GCP. Choose names and locations carefully.
- **All fields immutable**: Every field is ForceNew — any change destroys and recreates the resource.
- **No labels**: GCP KMS key rings do not support resource labels.
- **Name uniqueness**: Key ring names must be unique within a project and location. If a previous key ring with the same name exists (even if removed from IaC state), creation may fail or import the existing resource.

## Examples

For comprehensive examples, see [`examples.md`](examples.md), including:

- Regional key ring for production workloads
- Global key ring for cross-region use
- Multi-region key ring for high availability
- Cross-resource reference using GcpProject outputs

## Related Components

- [GcpKmsCryptoKey](/docs/catalog/gcp/gcpkmscryptokey) — creates encryption keys within this key ring
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the key ring is created
- [GcpBigQueryDataset](/docs/catalog/gcp/gcpbigquerydataset) — uses CryptoKeys (inside this key ring) for CMEK encryption
- [GcpSpannerInstance](/docs/catalog/gcp/gcpspannerinstance) — uses CryptoKeys for CMEK encryption
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — uses CryptoKeys for CMEK encryption

## Additional Resources

- [Cloud KMS Overview](https://cloud.google.com/kms/docs/overview)
- [Creating Key Rings](https://cloud.google.com/kms/docs/create-key-ring)
- [KMS Locations](https://cloud.google.com/kms/docs/locations)
- [Key Ring API Reference](https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings)

## Support

For issues, questions, or contributions, please refer to the OpenMCF documentation or open an issue in the repository.

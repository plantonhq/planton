# GCP KMS Key Handle

Requests a customer-managed encryption key from Cloud KMS Autokey for one resource type in one project and location. Autokey creates or reuses an HSM key in the `autokey` key ring, grants the resource type's service agent encrypt and decrypt on it, and returns its name in `kms_key`; the bucket, disk, dataset, topic, or database it protects references that output wherever it accepts a `GcpKmsKey`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `cloudkms.googleapis.com` on the handle's project (never disabled on destroy)
- **Key handle** -- one `kms_key_handle`, and through it the Autokey key

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to create key handles in the project (`roles/cloudkms.autokeyUser`, included in the resource-creation roles Google grants developers).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **Autokey on for the project** -- a **`GcpKmsAutokeyConfig`** on the project or a folder above it.

### Optional Dependencies

- **`GcpProject`** -- the handle's project, by reference (`projectId`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyHandle
metadata:
  name: orders-bucket-key
spec:
  projectId:
    value: orders-prod
  location: us-central1
  resourceTypeSelector: storage.googleapis.com/Bucket
```

```shell
planton apply -f kms-key-handle.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The protected resource's location (a region, a multi-region such as `us`, or `global`). Immutable. |
| `resourceTypeSelector` | `string` | The resource type, e.g. `storage.googleapis.com/Bucket`, `compute.googleapis.com/Disk`, `bigquery.googleapis.com/Dataset`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The protected resource's project (`GcpProject` ref). Immutable. |
| `keyHandleName` | `string` | `metadata.name` | The handle's ID, unique per project and location. Immutable. |

### Validation Rules

- `location` is a region, multi-region, or `global`.
- `resourceTypeSelector` has the form `service.googleapis.com/Type`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/keyHandles/{name}` |
| `kms_key` | `string` | The Autokey key: `projects/{p}/locations/{l}/keyRings/autokey/cryptoKeys/{key}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy keeps the handle.** Google cannot delete a key handle: destroy only forgets it, and the key keeps protecting its resources and billing as an HSM key version until a KMS administrator destroys it. A recreated handle needs a new `keyHandleName`.
- **Same location as the resource.** A CMEK key must be in the resource's location, and Autokey needs Cloud HSM there.
- **One key per resource or per location.** The granularity depends on the service (Google's Autokey page lists it); two handles for the same type and location may share a key.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpKmsAutokeyConfig** -- turns Autokey on for the project or folder
- **GcpGcsBucket**, **GcpComputeDisk**, **GcpBigQueryDataset**, **GcpPubSubTopic**, **GcpCloudSql**, **GcpSecretManagerSecret**, **GcpArtifactRegistryRepo**, **GcpSpannerDatabase**, and the other kinds Autokey serves -- consume `kms_key`
- **GcpKmsKey** -- hand-designed keys, when Autokey's defaults do not fit

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

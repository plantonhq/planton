# GcpKmsKeyRing

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpKmsKeyRingSpec defines the configuration for a GCP Cloud KMS key ring.

A key ring is an organizational grouping of cryptographic keys in Cloud KMS.
Key rings belong to a GCP project and reside in a specific location (region,
multi-region, or "global"). Once created, a key ring cannot be deleted — it
is a permanent container whose name and location are immutable, and whose
name can never be reused within the project+location.

Key rings do not carry any encryption policy themselves; they exist solely to
group and scope CryptoKeys. A single key ring can hold any number of
CryptoKeys with different purposes (ENCRYPT_DECRYPT, ASYMMETRIC_SIGN, etc.),
and IAM granted at the ring level flows down to every key inside it — which
makes the ring the natural blast-radius boundary: one ring per environment
or per data domain, not one ring per key.

All fields in this spec are immutable after creation — any change abandons
the old ring (GCP does not support deletion) and creates a new one.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyRing
metadata:
  name: test-key-ring
spec:
  projectId:
    value: "test-project"
  keyRingName: test-key-ring
  location: us-central1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.keyRingName` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project in which to create this key ring.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Example: "my-prod-project-123"

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.keyRingName

`string` · required

Name of the key ring in GCP. Immutable after creation.
Must be 1-63 characters: letters (upper or lower), digits, hyphens, or underscores.
This is the GCP resource name, distinct from the Planton metadata.name.
Because key rings are permanent, a name can never be reused within its
project and location — pick names that will not need recycling.
Example: "prod-encryption", "data-keys-us-central1"

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_-]{1,63}$"}}

### spec.location

`string` · required

GCP location (region, multi-region, or "global") where the key ring
resides. Immutable after creation. Keys must live in the same location
as the resources they protect for most CMEK integrations, so choose
based on where the encrypted data lives — plus data-residency and
redundancy requirements.

Common values:
  Region:       "us-central1", "europe-west1", "asia-east1"
  Multi-region: "us", "europe", "asia"
  Global:       "global"

Run `gcloud kms locations list` for a full list of valid locations.

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpKmsKeyRing, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.key_ring_id` | `string` | Fully qualified key ring resource path. Format: projects/{project}/locations/{location}/keyRings/{name} This is the primary identifier used by GcpKmsKey and other downstream resources that need to reference this key ring. |
| `status.outputs.key_ring_name` | `string` | The short name of the key ring (the last segment of key_ring_id). Useful for display, logging, and consumers that take the bare ring name alongside a separately supplied project and location. |
| `status.outputs.location` | `string` | The location the key ring resides in (region, multi-region, or "global"), exactly as GCP resolved it. Consumers that take a bare ring name plus a location compose from key_ring_name + this field. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpKmsKey | `spec.keyRingId` | `status.outputs.key_ring_id` |
| KubernetesOpenBao | `spec.autoUnseal.gcpKms.keyRing` | `status.outputs.key_ring_name` |
| KubernetesPlantonPlatform | `spec.vault.autoUnseal.gcpKms.keyRing` | `status.outputs.key_ring_name` |

## See Also

- [Overview](../README.md)

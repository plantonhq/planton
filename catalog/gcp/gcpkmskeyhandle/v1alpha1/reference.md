# GcpKmsKeyHandle

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpKmsKeyHandleSpec requests a customer-managed encryption key from Cloud
KMS Autokey (`google_kms_key_handle`) for one resource type in one
project and location.

The handle is how a team gets CMEK without designing keys: Autokey (on
for the project or its folder, see GcpKmsAutokeyConfig) creates or reuses
an HSM key in the "autokey" key ring, grants the resource type's service
agent encrypt and decrypt on it, and returns its name in the kms_key
output. The resource it protects -- a GcpGcsBucket, GcpComputeDisk,
GcpBigQueryDataset, GcpPubSubTopic, GcpCloudSql, and the other kinds
Autokey serves -- then references that output wherever it accepts a
GcpKmsKey.

Every field is immutable. Destroy only forgets the handle: Google keeps
it (and the key keeps protecting its resources and billing as an HSM key
until a KMS administrator destroys it), so a recreated handle needs a
new key_handle_name.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyHandle
metadata:
  name: orders-bucket-key
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  resourceTypeSelector: storage.googleapis.com/Bucket
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.resourceTypeSelector` | `string` | yes |  |  |
| `spec.keyHandleName` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project of the resource the key will protect (where the handle
lives): a literal project ID or a GcpProject reference. Empty means the
provider's default project. Autokey must be on for this project or a
folder above it.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The location of the resource the key will protect, e.g. "us-central1",
"europe-west4", or a multi-region such as "us". A CMEK key must sit in
the same location as its resource, so this must equal the resource's
location, and Autokey needs Cloud HSM there.

- rule: location must be a region, a multi-region, or global
- rule: {"required":true}

### spec.resourceTypeSelector

`string` · required

The resource type the key is for, as Google names it:
  "storage.googleapis.com/Bucket", "compute.googleapis.com/Disk",
  "bigquery.googleapis.com/Dataset", "pubsub.googleapis.com/Topic",
  "sqladmin.googleapis.com/Instance", "secretmanager.googleapis.com/Secret",
  "artifactregistry.googleapis.com/Repository", "run.googleapis.com/Service",
  "spanner.googleapis.com/Database", "redis.googleapis.com/Instance", ...
Google's Autokey page lists every compatible type and the key
granularity (one key per resource, or per location for some types).

- rule: resource_type_selector must be service.googleapis.com/Type, e.g. storage.googleapis.com/Bucket
- rule: {"required":true}

### spec.keyHandleName

`string`

The handle's ID, unique per project and location. Defaults to
metadata.name. A destroyed handle keeps its ID in Google, so a
recreated handle needs a new one.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpKmsKeyHandle, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/keyHandles/{key_handle_name}. |
| `status.outputs.kms_key` | `string` | The Cloud KMS key Autokey assigned: projects/{p}/locations/{l}/keyRings/autokey/cryptoKeys/{key}. Reference it from the protected resource's customer-managed key field. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAlloydbCluster | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpArtifactRegistryRepo | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpBigQueryDataset | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpBigtableInstance | `spec.clusters[].kmsKeyName` | `status.outputs.kms_key` |
| GcpCloudComposerEnvironment | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpCloudRun | `spec.encryptionKey` | `status.outputs.kms_key` |
| GcpCloudRunJob | `spec.template.encryptionKey` | `status.outputs.kms_key` |
| GcpCloudSql | `spec.encryptionKeyName` | `status.outputs.kms_key` |
| GcpComputeDisk | `spec.kmsKey` | `status.outputs.kms_key` |
| GcpComputeImage | `spec.kmsKey` | `status.outputs.kms_key` |
| GcpDataprocCluster | `spec.clusterConfig.encryptionKmsKeyName` | `status.outputs.kms_key` |
| GcpDatastreamStream | `spec.customerManagedEncryptionKey` | `status.outputs.kms_key` |
| GcpFilestoreInstance | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpGcsBucket | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpMemorystoreInstance | `spec.kmsKey` | `status.outputs.kms_key` |
| GcpPubSubTopic | `spec.kmsKeyName` | `status.outputs.kms_key` |
| GcpRedisCluster | `spec.kmsKey` | `status.outputs.kms_key` |
| GcpRedisInstance | `spec.customerManagedKey` | `status.outputs.kms_key` |
| GcpSecretManagerSecret | `spec.replication.auto.customerManagedEncryption.kmsKey` | `status.outputs.kms_key` |
| GcpSecretManagerSecret | `spec.replication.userManaged.replicas[].customerManagedEncryption.kmsKey` | `status.outputs.kms_key` |
| GcpSecretManagerSecret | `spec.customerManagedEncryption.kmsKey` | `status.outputs.kms_key` |
| GcpSpannerDatabase | `spec.encryptionConfig.kmsKeyName` | `status.outputs.kms_key` |
| GcpWorkflow | `spec.cryptoKey` | `status.outputs.kms_key` |

## See Also

- [Overview](../README.md)

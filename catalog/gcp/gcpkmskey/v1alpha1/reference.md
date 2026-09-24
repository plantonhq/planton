# GcpKmsKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpKmsKeySpec defines the configuration for a GCP Cloud KMS cryptographic key.

A key belongs to a key ring and performs the actual cryptographic operations:
symmetric encryption/decryption (CMEK), asymmetric signing, asymmetric
decryption, raw encryption, or MAC generation. Keys are the primary resource
that downstream services (BigQuery, Spanner, GKE, Cloud SQL, Cloud Run,
Pub/Sub, and others) reference for customer-managed encryption.

Important behavioral notes:

  - Keys cannot be deleted from GCP. On destroy, all key versions are
    destroyed (rendering all data encrypted under them unrecoverable once
    the destroy-scheduled window elapses) and automatic rotation is
    disabled, but the key object itself remains permanently in the key
    ring. A key name can therefore never be reused within its ring.

  - Most fields are immutable after creation: key_ring_id, key_name,
    purpose, destroy_scheduled_duration, import_only, crypto_key_backend,
    and version_template.protection_level. Only rotation_period,
    version_template.algorithm, and labels can be updated in place.

  - Services consuming this key for CMEK need their service agent granted
    roles/cloudkms.cryptoKeyEncrypterDecrypter on the key — model that
    grant as a first-class IAM node alongside the consumer.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKey
metadata:
  name: test-kms-key
spec:
  keyRingId:
    value: "projects/test-project/locations/us-central1/keyRings/test-key-ring"
  keyName: test-encrypt-key
  rotationPeriod: "7776000s"
  labels:
    team: platform
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.keyRingId` | `string \| valueFrom` | yes |  | GcpKmsKeyRing (`status.outputs.key_ring_id`) |
| `spec.keyName` | `string` | yes |  |  |
| `spec.purpose` | `string` |  |  |  |
| `spec.rotationPeriod` | `string` |  |  |  |
| `spec.destroyScheduledDuration` | `string` |  |  |  |
| `spec.versionTemplate` | `GcpKmsKeyVersionTemplate` |  |  |  |
| `spec.versionTemplate.algorithm` | `string` | yes |  |  |
| `spec.versionTemplate.protectionLevel` | `string` |  |  |  |
| `spec.skipInitialVersionCreation` | `bool` |  |  |  |
| `spec.importOnly` | `bool` |  |  |  |
| `spec.cryptoKeyBackend` | `string \| valueFrom` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.keyRingId

`string | valueFrom` · required

The key ring that this key belongs to.
Accepts the fully qualified key ring path
  projects/{project}/locations/{location}/keyRings/{name}
or a reference to a GcpKmsKeyRing resource. The key inherits the ring's
project and location. Immutable after creation.

- references: GcpKmsKeyRing (`status.outputs.key_ring_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKeyRing, name: <that resource's name>, fieldPath: status.outputs.key_ring_id}} -- a bare string does not parse

### spec.keyName

`string` · required

Name of the key in GCP. Immutable after creation.
Must be 1-63 characters: letters (upper or lower), digits, hyphens,
or underscores. This is the GCP resource name, distinct from the
Planton metadata.name. Because keys are permanent (see the message
comment), a name can never be reused within its key ring — pick
versioned names (e.g. "cmek-data-v2") if a key may ever be replaced.
Example: "cmek-encrypt-key", "artifact-signing-key"

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_-]{1,63}$"}}

### spec.purpose

`string`

The immutable purpose of this key. Determines what cryptographic
operations the key supports and which algorithms its version template
may use. Cannot be changed after creation. If not set, GCP defaults
to ENCRYPT_DECRYPT — the purpose every CMEK integration expects.

Known values:
  "ENCRYPT_DECRYPT"      -- symmetric encryption for CMEK (default)
  "ASYMMETRIC_SIGN"      -- digital signatures
  "ASYMMETRIC_DECRYPT"   -- asymmetric decryption
  "RAW_ENCRYPT_DECRYPT"  -- raw AES-GCM/CTR for interoperable encryption
  "MAC"                  -- keyed message authentication codes

GCP adds purposes over time (for example key encapsulation for
post-quantum schemes), so this field deliberately accepts any string
and lets the API validate — see the CryptoKeyPurpose reference for
the authoritative list:
https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys#CryptoKeyPurpose

### spec.rotationPeriod

`string`

How often to auto-generate a new CryptoKeyVersion and set it as primary.
Format: decimal seconds with suffix "s" (e.g., "7776000s" for 90 days).
Must be at least 86400s (24 hours) with at most 9 fractional digits.
Only allowed for ENCRYPT_DECRYPT keys (enforced pre-deploy); other
purposes require manual key version management. Mutable: shortening or
lengthening the period applies in place and reschedules the next
rotation. Old versions remain usable for decryption until destroyed —
rotation limits blast radius going forward; it does not re-encrypt
existing data.

- rule: rotation_period must be a duration in seconds of at least 86400s (24 hours) with at most 9 fractional digits (e.g., '7776000s' for 90 days)

### spec.destroyScheduledDuration

`string`

How long destroyed CryptoKeyVersions spend in DESTROY_SCHEDULED state
before being permanently destroyed — the recovery window during which
a destruction can still be undone. Immutable after creation.
Format: decimal seconds with suffix "s" (e.g., "2592000s" for 30 days).
Defaults to 30 days when not specified. Shorter windows reduce the
time compromised material lingers; longer windows protect against
accidental destruction of data-encrypting keys.

- rule: destroy_scheduled_duration must be a duration in seconds (e.g., '2592000s' for 30 days)

### spec.versionTemplate

`GcpKmsKeyVersionTemplate`

Template describing settings for new CryptoKeyVersions.
Use this to specify the encryption algorithm and protection level.
If omitted, GCP defaults to GOOGLE_SYMMETRIC_ENCRYPTION with SOFTWARE
protection -- which is correct for standard CMEK use cases.

### spec.versionTemplate.algorithm

`string` · required

The algorithm to use when creating a CryptoKeyVersion based on this
template. Required when version_template is specified. The algorithm
must be compatible with the key's purpose; GCP rejects mismatches at
create time. Mutable: changing the algorithm affects only versions
created afterward — existing versions keep the algorithm they were
generated with.

Common values by purpose:
  ENCRYPT_DECRYPT:     "GOOGLE_SYMMETRIC_ENCRYPTION" (default if omitted)
  ASYMMETRIC_SIGN:     "EC_SIGN_P256_SHA256", "RSA_SIGN_PSS_2048_SHA256", ...
  ASYMMETRIC_DECRYPT:  "RSA_DECRYPT_OAEP_2048_SHA256", ...
  RAW_ENCRYPT_DECRYPT: "AES_128_GCM", "AES_256_GCM", ...
  MAC:                 "HMAC_SHA256"

GCP adds algorithms over time (for example post-quantum signature
schemes), so this field deliberately accepts any string and lets the
API validate — see the CryptoKeyVersionAlgorithm reference for the
authoritative list:
https://cloud.google.com/kms/docs/reference/rest/v1/CryptoKeyVersionAlgorithm

- rule: {"required":true}

### spec.versionTemplate.protectionLevel

`string`

The protection level for CryptoKeyVersions created with this template.
Immutable after creation.

Valid values:
  "SOFTWARE" (default) -- keys protected in software
  "HSM"                -- keys protected by Cloud HSM (FIPS 140-2 Level 3)
  "EXTERNAL"           -- key material held in an external key manager,
                          linked per version via an external key URI
  "EXTERNAL_VPC"       -- key material held in an external key manager
                          reached over a VPC; requires the key-level
                          crypto_key_backend to name the EKM connection

- rule: protection_level must be one of: SOFTWARE, HSM, EXTERNAL, EXTERNAL_VPC

### spec.skipInitialVersionCreation

`bool`

If true, the key is created without an initial CryptoKeyVersion.
You must create versions manually afterward (or import material via a
key ring import job). Consumed only at create time. Required (and
enforced pre-deploy) when import_only is true.

### spec.importOnly

`bool`

If true, this key may contain only imported key versions — GCP will
never generate material for it, making the key a bring-your-own-key
(BYOK) container whose material provenance is externally controlled.
Immutable after creation. Requires skip_initial_version_creation
(enforced pre-deploy), since an auto-generated initial version would
violate the import-only guarantee.

### spec.cryptoKeyBackend

`string | valueFrom`

The EKM connection through which an external key manager backs this
key's versions. Applies only when version_template.protection_level is
EXTERNAL_VPC (enforced pre-deploy). Accepts the fully qualified
connection path
  projects/{project}/locations/{location}/ekmConnections/{name}
Immutable after creation.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User-defined labels attached to the key, for cost attribution and
fleet queries. Merged with Planton's platform labels (which win on
key conflicts). Mutable in place.

### spec.deletionPolicy

`string`

What destroying this resource does to the key. KMS keys can NEVER be
deleted from GCP — only their versions can — so the choice here is
about the versions, and it is the most consequential destroy switch
in this spec:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- EVERY key version is scheduled for destruction (after
               destroy_scheduled_duration) and rotation is disabled;
               data encrypted under this key becomes UNRECOVERABLE
               once the versions are destroyed. The key shell remains
               in the project.
  "PREVENT" -- destroy FAILS; the right default posture for any key
               protecting data you cannot afford to lose
  "ABANDON" -- the key leaves management with every version intact
               and enabled — encrypted data stays decryptable

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `rotation_only_for_encrypt_decrypt`: rotation_period can only be set for ENCRYPT_DECRYPT keys — other purposes require manual key version management
- `import_only_requires_skip_initial_version`: import_only keys must also set skip_initial_version_creation — GCP cannot generate an initial version for a key that only accepts imported material
- `crypto_key_backend_requires_external_vpc`: crypto_key_backend applies only to keys with version_template.protection_level EXTERNAL_VPC (EXTERNAL keys link material per version via external key URIs instead)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpKmsKey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.key_id` | `string` | Fully qualified crypto key resource path. Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{name} This is the primary CMEK reference used by BigQuery, Spanner, GKE, CloudSQL, and other downstream resources for customer-managed encryption. |
| `status.outputs.key_name` | `string` | The short name of the key (the last segment of key_id). Useful for display, logging, and consumers that take the bare key name alongside a separately supplied project and location. |
| `status.outputs.primary_version_name` | `string` | Fully qualified resource name of the key's current primary CryptoKeyVersion — the version GCP uses to encrypt new data. Format: projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}/cryptoKeyVersions/{n} Populated by GCP only for ENCRYPT_DECRYPT keys; empty for asymmetric, raw, and MAC keys (which have no primary-version concept) and for keys created with skip_initial_version_creation. |
| `status.outputs.primary_state` | `string` | Lifecycle state of the primary CryptoKeyVersion (e.g. "ENABLED"). Same population rules as primary_version_name — the quick health probe that a CMEK key is actually able to encrypt. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.keyRingId` | GcpKmsKeyRing | `status.outputs.key_ring_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAlloydbCluster | `spec.automatedBackupPolicy.encryptionKmsKeyName` | `status.outputs.key_id` |
| GcpAlloydbCluster | `spec.continuousBackupConfig.encryptionKmsKeyName` | `status.outputs.key_id` |
| GcpAlloydbCluster | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpArtifactRegistryRepo | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpBigQueryConnection | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpBigQueryDataset | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpBigQueryTable | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpBigtableInstance | `spec.clusters[].kmsKeyName` | `status.outputs.key_id` |
| GcpCloudComposerEnvironment | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpCloudFunction | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpCloudRun | `spec.encryptionKey` | `status.outputs.key_id` |
| GcpCloudRunJob | `spec.template.encryptionKey` | `status.outputs.key_id` |
| GcpCloudRunWorkerPool | `spec.encryptionKey` | `status.outputs.key_id` |
| GcpCloudSql | `spec.encryptionKeyName` | `status.outputs.key_id` |
| GcpColabRuntimeTemplate | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpColabSchedule | `spec.notebookExecutionJob.kmsKeyName` | `status.outputs.key_id` |
| GcpColabSchedule | `spec.pipelineJob.kmsKeyName` | `status.outputs.key_id` |
| GcpComputeDisk | `spec.kmsKey` | `status.outputs.key_id` |
| GcpComputeDisk | `spec.sourceImageEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeDisk | `spec.sourceSnapshotEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeInstance | `spec.bootDisk.kmsKey` | `status.outputs.key_id` |
| GcpComputeInstance | `spec.bootDisk.sourceImageEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeInstance | `spec.bootDisk.sourceSnapshotEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeInstance | `spec.attachedDisks[].kmsKey` | `status.outputs.key_id` |
| GcpComputeInstance | `spec.instanceEncryptionKey.kmsKey` | `status.outputs.key_id` |
| GcpComputeMig | `spec.template.disks[].diskEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeMig | `spec.template.disks[].sourceImageEncryption.kmsKey` | `status.outputs.key_id` |
| GcpComputeMig | `spec.template.disks[].sourceSnapshotEncryption.kmsKey` | `status.outputs.key_id` |
| GcpDataprocCluster | `spec.clusterConfig.encryptionKmsKeyName` | `status.outputs.key_id` |
| GcpDataprocCluster | `spec.clusterConfig.securityConfig.kerberosConfig.kmsKeyUri` | `status.outputs.key_id` |
| GcpDatastreamStream | `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.kmsKeyName` | `status.outputs.key_id` |
| GcpDatastreamStream | `spec.customerManagedEncryptionKey` | `status.outputs.key_id` |
| GcpDocumentAiProcessor | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpEventarcMessageBus | `spec.cryptoKey` | `status.outputs.key_id` |
| GcpEventarcMessageBus | `spec.googleApiSources[].cryptoKey` | `status.outputs.key_id` |
| GcpEventarcMessageBus | `spec.pipelines[].cryptoKey` | `status.outputs.key_id` |
| GcpEventarcTrigger | `spec.partnerChannel.cryptoKey` | `status.outputs.key_id` |
| GcpEventarcTrigger | `spec.googleChannelCryptoKey` | `status.outputs.key_id` |
| GcpFilestoreInstance | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpFirestoreDatabase | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpGcsBucket | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpGkeCluster | `spec.clusterAutoscaling.autoProvisioningDefaults.bootDiskKmsKey` | `status.outputs.key_id` |
| GcpGkeCluster | `spec.databaseEncryption.keyName` | `status.outputs.key_id` |
| GcpGkeCluster | `spec.userManagedKeys.controlPlaneDiskEncryptionKey` | `status.outputs.key_id` |
| GcpGkeCluster | `spec.userManagedKeys.gkeopsEtcdBackupEncryptionKey` | `status.outputs.key_id` |
| GcpGkeNodePool | `spec.nodeConfig.bootDiskKmsKey` | `status.outputs.key_id` |
| GcpKmsKeyIamMember | `spec.cryptoKeyId` | `status.outputs.key_id` |
| GcpLogBucket | `spec.cmekKmsKey` | `status.outputs.key_id` |
| GcpLogBucket | `spec.scopeSettings.kmsKey` | `status.outputs.key_id` |
| GcpManagedKafkaCluster | `spec.kmsKey` | `status.outputs.key_id` |
| GcpMemorystoreInstance | `spec.kmsKey` | `status.outputs.key_id` |
| GcpPrivateCaCertificateAuthority | `spec.keySpec.cloudKmsKeyVersion` | `status.outputs.primary_version_name` |
| GcpPrivateCaPool | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpPubSubTopic | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpRedisCluster | `spec.kmsKey` | `status.outputs.key_id` |
| GcpRedisInstance | `spec.customerManagedKey` | `status.outputs.key_id` |
| GcpSecretManagerSecret | `spec.replication.auto.customerManagedEncryption.kmsKey` | `status.outputs.key_id` |
| GcpSecretManagerSecret | `spec.replication.userManaged.replicas[].customerManagedEncryption.kmsKey` | `status.outputs.key_id` |
| GcpSecretManagerSecret | `spec.customerManagedEncryption.kmsKey` | `status.outputs.key_id` |
| GcpSpannerBackupSchedule | `spec.encryptionConfig.kmsKeyName` | `status.outputs.key_id` |
| GcpSpannerBackupSchedule | `spec.encryptionConfig.kmsKeyNames` | `status.outputs.key_id` |
| GcpSpannerDatabase | `spec.encryptionConfig.kmsKeyName` | `status.outputs.key_id` |
| GcpSpannerDatabase | `spec.encryptionConfig.kmsKeyNames` | `status.outputs.key_id` |
| GcpVectorSearchCollection | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiAgentEngine | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiDataset | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiEndpoint | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiFeatureOnlineStore | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiIndex | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiIndexEndpoint | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiNotebook | `spec.bootDisk.kmsKey` | `status.outputs.key_id` |
| GcpVertexAiNotebook | `spec.dataDisk.kmsKey` | `status.outputs.key_id` |
| GcpVertexAiPersistentResource | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiSearchDataConnector | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiSearchDataStore | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiSearchEngine | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpVertexAiTensorboard | `spec.kmsKeyName` | `status.outputs.key_id` |
| GcpWorkflow | `spec.cryptoKey` | `status.outputs.key_id` |
| KubernetesOpenBao | `spec.autoUnseal.gcpKms.cryptoKey` | `status.outputs.key_name` |
| KubernetesPlantonPlatform | `spec.vault.autoUnseal.gcpKms.cryptoKey` | `status.outputs.key_name` |

## See Also

- [Overview](../README.md)

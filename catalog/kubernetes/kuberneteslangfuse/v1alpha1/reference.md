# KubernetesLangfuse

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesLangfuseSpec** installs Langfuse - the open-source LLM
engineering platform for tracing, prompt management, evaluations and
datasets - from the official `langfuse` chart
(https://langfuse.github.io/langfuse-k8s, chart 2.1.x = Langfuse v4).
One release runs two workloads from one configuration: the WEB server
(UI, public API, OTLP ingestion endpoint) and the WORKER (queue
consumers that load events into ClickHouse, run evaluations, exports and
retention).

FOUR STORES, ALL COMPOSED BY REFERENCE (the chart's bundled copies never
ship from this kind):
  - POSTGRES holds users, projects, prompts, API keys and settings. A
    KubernetesPostgres composes naturally; any PostgreSQL 15+ (RDS,
    Cloud SQL, Azure) works by literal host.
  - CLICKHOUSE holds traces, observations and scores. A
    KubernetesClickHouse (25.12 or newer) composes naturally, as do
    ClickHouse Cloud and BYOC by literal host.
  - REDIS (or Valkey) carries the ingestion queues and caches, and MUST
    run with `maxmemory-policy noeviction`: under eviction queued jobs
    are silently lost. A KubernetesValkey composes naturally.
  - BLOB STORAGE receives every ingested event before the worker reads
    it, plus media attachments and batch exports. S3-compatible stores
    (a KubernetesSeaweedFs, Cloudflare R2), AWS S3, Google Cloud Storage
    and Azure Blob Storage each have a native arm.

A Kubernetes Secret reference cannot cross namespaces: every Secret this
spec references by name (store passwords, TLS material, extra
environment) must live in Langfuse's own namespace. Compose the stores
into the same namespace, or copy their Secrets.

SECRETS: every sensitive value the spec carries is written by the module
into its own Secret (the `instance_secret_name` output) and reaches
Langfuse through a secretKeyRef; none is ever rendered into chart values.
The salt, encryption key and NextAuth secret are generated once when not
declared - and they are bound to the Postgres data, not to this resource:
see `secrets`.

NOT MODELED, BY DESIGN (verified at chart 2.1.5 / Langfuse 4.56.0):
- the chart's bundled `postgresql`, `redis`, `clickhouse` and `s3`
  stores: superseded by the first-class kinds above, which own their own
  backups, sizing and lifecycle;
- the chart's `ingress` block: compose a KubernetesHttpRoute or
  KubernetesIngress against the `service` output;
- `allowV1Upgrade`: a guard for upgrading chart 1.x installs;
- `NEXT_PUBLIC_BASE_PATH`: a build-time setting that needs a custom
  image (serve Langfuse at the root of its own host instead);
- OCI native object storage: its auth modes read a mounted config file
  or instance metadata the chart has no seam for; OCI Object Storage
  works through `s3_compatible` with customer secret keys;
- the chart's generic sidecar, init-container, lifecycle-hook and
  `extraManifests` slots: generic pod plumbing with no Langfuse meaning,
  reachable through `helm_values`;
- Langfuse Cloud internals (billing, Stripe, CRM, Sentry, PostHog,
  gateway and metering variables) and undocumented tuning variables:
  `extra_env` carries any of them verbatim.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLangfuse
metadata:
  name: langfuse
spec:
  namespace:
    value: langfuse
  externalUrl: https://langfuse.example.com
  postgres:
    host:
      valueFrom:
        name: langfuse-db
    username: langfuse
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-db
  clickhouse:
    host:
      valueFrom:
        name: langfuse-clickhouse
    username: langfuse
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-clickhouse
    cluster:
      name:
        valueFrom:
          name: langfuse-clickhouse
  redis:
    standalone:
      host:
        valueFrom:
          name: langfuse-cache
    username:
      valueFrom:
        name: langfuse-cache
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-cache
      secretKey:
        valueFrom:
          name: langfuse-cache
  blobStorage:
    target:
      s3Compatible:
        endpoint:
          valueFrom:
            name: langfuse-storage
        bucket: langfuse
        credentialsSecret:
          secretName:
            valueFrom:
              name: langfuse-storage
    mediaUpload:
      publicEndpoint: https://s3.langfuse.example.com
    batchExport:
      publicEndpoint: https://s3.langfuse.example.com
  init:
    organization:
      name: Acme
    project:
      name: Production
    admin:
      email: admin@example.com
  auth:
    disableSignUp: true
  web:
    replicas: 2
  worker:
    replicas: 2
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.1.5` |  |
| `spec.externalUrl` | `string` |  | `http://localhost:3000` |  |
| `spec.postgres` | `KubernetesLangfusePostgres` | yes |  |  |
| `spec.postgres.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.postgres.port` | `int32` |  | `5432` |  |
| `spec.postgres.database` | `string` |  | `langfuse` |  |
| `spec.postgres.username` | `string` | yes |  |  |
| `spec.postgres.passwordSecret` | `KubernetesLangfusePostgresPasswordSecret` | yes |  |  |
| `spec.postgres.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.postgres.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.postgres.args` | `string` |  |  |  |
| `spec.postgres.directUrl` | `string` (sensitive) |  |  |  |
| `spec.postgres.shadowDatabaseUrl` | `string` (sensitive) |  |  |  |
| `spec.postgres.autoMigrate` | `bool` |  | `true` |  |
| `spec.clickhouse` | `KubernetesLangfuseClickHouse` | yes |  |  |
| `spec.clickhouse.host` | `string \| valueFrom` | yes |  | KubernetesClickHouse (`status.outputs.service_name`) |
| `spec.clickhouse.httpPort` | `int32` |  | `8123` |  |
| `spec.clickhouse.nativePort` | `int32` |  | `9000` |  |
| `spec.clickhouse.secure` | `bool` |  |  |  |
| `spec.clickhouse.database` | `string` |  | `default` |  |
| `spec.clickhouse.username` | `string` | yes |  |  |
| `spec.clickhouse.passwordSecret` | `KubernetesLangfuseClickHousePasswordSecret` | yes |  |  |
| `spec.clickhouse.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesClickHouse (`status.outputs.auth_secret_name`) |
| `spec.clickhouse.passwordSecret.secretKey` | `string` |  |  |  |
| `spec.clickhouse.cluster` | `KubernetesLangfuseClickHouseCluster` |  |  |  |
| `spec.clickhouse.cluster.enabled` | `bool` |  | `true` |  |
| `spec.clickhouse.cluster.name` | `string \| valueFrom` |  |  | KubernetesClickHouse (`status.outputs.cluster_name`) |
| `spec.clickhouse.readOnlyUrl` | `string` |  |  |  |
| `spec.clickhouse.autoMigrate` | `bool` |  | `true` |  |
| `spec.clickhouse.deleteMode` | `enum` |  |  |  |
| `spec.clickhouse.useLightweightUpdate` | `bool` |  |  |  |
| `spec.clickhouse.deletionTimeoutMs` | `int32` |  |  |  |
| `spec.redis` | `KubernetesLangfuseRedis` | yes |  |  |
| `spec.redis.standalone` | `KubernetesLangfuseRedisStandalone` |  |  |  |
| `spec.redis.standalone.host` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.service`) |
| `spec.redis.standalone.port` | `int32` |  | `6379` |  |
| `spec.redis.standalone.database` | `int32` |  |  |  |
| `spec.redis.cluster` | `KubernetesLangfuseRedisCluster` |  |  |  |
| `spec.redis.cluster.nodes` | `[]string` | yes |  |  |
| `spec.redis.cluster.slotsRefreshTimeoutMs` | `int32` |  |  |  |
| `spec.redis.sentinel` | `KubernetesLangfuseRedisSentinel` |  |  |  |
| `spec.redis.sentinel.masterName` | `string` | yes |  |  |
| `spec.redis.sentinel.nodes` | `[]string` | yes |  |  |
| `spec.redis.sentinel.username` | `string` |  |  |  |
| `spec.redis.sentinel.passwordSecret` | `KubernetesSecretKey` |  |  |  |
| `spec.redis.sentinel.passwordSecret.name` | `string` |  |  |  |
| `spec.redis.sentinel.passwordSecret.key` | `string` |  |  |  |
| `spec.redis.sentinel.tlsEnabled` | `bool` |  |  |  |
| `spec.redis.username` | `string \| valueFrom` |  |  | KubernetesValkey (`status.outputs.username`) |
| `spec.redis.passwordSecret` | `KubernetesLangfuseRedisPasswordSecret` |  |  |  |
| `spec.redis.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.password_secret.name`) |
| `spec.redis.passwordSecret.secretKey` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.password_secret.key`) |
| `spec.redis.tls` | `KubernetesLangfuseRedisTls` |  |  |  |
| `spec.redis.tls.enabled` | `bool` |  |  |  |
| `spec.redis.tls.ca` | `KubernetesLangfuseCertificateFile` |  |  |  |
| `spec.redis.tls.ca.secretName` | `string` |  |  |  |
| `spec.redis.tls.ca.configMapName` | `string` |  |  |  |
| `spec.redis.tls.ca.key` | `string` |  | `ca.crt` |  |
| `spec.redis.tls.clientCertificate` | `KubernetesLangfuseSecretFile` |  |  |  |
| `spec.redis.tls.clientCertificate.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.redis.tls.clientCertificate.key` | `string` |  |  |  |
| `spec.redis.tls.clientKey` | `KubernetesLangfuseSecretFile` |  |  |  |
| `spec.redis.tls.clientKey.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.redis.tls.clientKey.key` | `string` |  |  |  |
| `spec.redis.tls.clientKeyPassphrase` | `string` (sensitive) |  |  |  |
| `spec.redis.tls.serverName` | `string` |  |  |  |
| `spec.redis.tls.rejectUnauthorized` | `bool` |  |  |  |
| `spec.redis.tls.checkServerIdentity` | `bool` |  |  |  |
| `spec.redis.tls.ciphers` | `string` |  |  |  |
| `spec.redis.tls.secureProtocol` | `string` |  |  |  |
| `spec.redis.tls.honorCipherOrder` | `bool` |  |  |  |
| `spec.redis.keyPrefix` | `string` |  |  |  |
| `spec.redis.socketTimeoutMs` | `int32` |  |  |  |
| `spec.blobStorage` | `KubernetesLangfuseBlobStorage` | yes |  |  |
| `spec.blobStorage.target` | `KubernetesLangfuseBlobTarget` | yes |  |  |
| `spec.blobStorage.target.s3Compatible` | `KubernetesLangfuseS3CompatibleTarget` |  |  |  |
| `spec.blobStorage.target.s3Compatible.endpoint` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.target.s3Compatible.bucket` | `string` | yes |  |  |
| `spec.blobStorage.target.s3Compatible.region` | `string` |  | `auto` |  |
| `spec.blobStorage.target.s3Compatible.forcePathStyle` | `bool` |  | `true` |  |
| `spec.blobStorage.target.s3Compatible.credentialsSecret` | `KubernetesLangfuseS3CredentialsSecret` |  |  |  |
| `spec.blobStorage.target.s3Compatible.credentialsSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`) |
| `spec.blobStorage.target.s3Compatible.credentialsSecret.accessKeyIdKey` | `string` |  | `admin_access_key_id` |  |
| `spec.blobStorage.target.s3Compatible.credentialsSecret.secretAccessKeyKey` | `string` |  | `admin_secret_access_key` |  |
| `spec.blobStorage.target.s3Compatible.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.target.s3Compatible.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.target.s3Compatible.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.target.s3` | `KubernetesLangfuseS3Target` |  |  |  |
| `spec.blobStorage.target.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.blobStorage.target.s3.region` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.region`) |
| `spec.blobStorage.target.s3.endpoint` | `string` |  |  |  |
| `spec.blobStorage.target.s3.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.target.s3.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.target.s3.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.target.r2` | `KubernetesLangfuseR2Target` |  |  |  |
| `spec.blobStorage.target.r2.endpoint` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.target.r2.bucket` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.target.r2.credentials` | `KubernetesLangfuseR2Credentials` | yes |  |  |
| `spec.blobStorage.target.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.blobStorage.target.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.blobStorage.target.gcs` | `KubernetesLangfuseGcsTarget` |  |  |  |
| `spec.blobStorage.target.gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.target.gcs.credentialsJson` | `string` (sensitive) |  |  |  |
| `spec.blobStorage.target.azureBlob` | `KubernetesLangfuseAzureBlobTarget` |  |  |  |
| `spec.blobStorage.target.azureBlob.storageAccount` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.blobStorage.target.azureBlob.container` | `string \| valueFrom` | yes |  | AzureStorageContainer (`status.outputs.container_name`) |
| `spec.blobStorage.target.azureBlob.accountKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.target.azureBlob.endpoint` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.primary_blob_endpoint`) |
| `spec.blobStorage.eventUpload` | `KubernetesLangfuseEventUpload` |  |  |  |
| `spec.blobStorage.eventUpload.prefix` | `string` |  | `events/` |  |
| `spec.blobStorage.eventUpload.sse` | `KubernetesLangfuseServerSideEncryption` |  |  |  |
| `spec.blobStorage.eventUpload.sse.algorithm` | `enum` |  |  |  |
| `spec.blobStorage.eventUpload.sse.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.blobStorage.mediaUpload` | `KubernetesLangfuseMediaUpload` |  |  |  |
| `spec.blobStorage.mediaUpload.target` | `KubernetesLangfuseBlobTarget` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible` | `KubernetesLangfuseS3CompatibleTarget` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.endpoint` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.mediaUpload.target.s3Compatible.bucket` | `string` | yes |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.region` | `string` |  | `auto` |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.forcePathStyle` | `bool` |  | `true` |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret` | `KubernetesLangfuseS3CredentialsSecret` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`) |
| `spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.accessKeyIdKey` | `string` |  | `admin_access_key_id` |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.secretAccessKeyKey` | `string` |  | `admin_secret_access_key` |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.mediaUpload.target.s3Compatible.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.mediaUpload.target.s3` | `KubernetesLangfuseS3Target` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.blobStorage.mediaUpload.target.s3.region` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.region`) |
| `spec.blobStorage.mediaUpload.target.s3.endpoint` | `string` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.mediaUpload.target.s3.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.mediaUpload.target.s3.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.mediaUpload.target.r2` | `KubernetesLangfuseR2Target` |  |  |  |
| `spec.blobStorage.mediaUpload.target.r2.endpoint` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.mediaUpload.target.r2.bucket` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.mediaUpload.target.r2.credentials` | `KubernetesLangfuseR2Credentials` | yes |  |  |
| `spec.blobStorage.mediaUpload.target.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.blobStorage.mediaUpload.target.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.blobStorage.mediaUpload.target.gcs` | `KubernetesLangfuseGcsTarget` |  |  |  |
| `spec.blobStorage.mediaUpload.target.gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.mediaUpload.target.gcs.credentialsJson` | `string` (sensitive) |  |  |  |
| `spec.blobStorage.mediaUpload.target.azureBlob` | `KubernetesLangfuseAzureBlobTarget` |  |  |  |
| `spec.blobStorage.mediaUpload.target.azureBlob.storageAccount` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.blobStorage.mediaUpload.target.azureBlob.container` | `string \| valueFrom` | yes |  | AzureStorageContainer (`status.outputs.container_name`) |
| `spec.blobStorage.mediaUpload.target.azureBlob.accountKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.mediaUpload.target.azureBlob.endpoint` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.primary_blob_endpoint`) |
| `spec.blobStorage.mediaUpload.prefix` | `string` |  | `media/` |  |
| `spec.blobStorage.mediaUpload.publicEndpoint` | `string` |  |  |  |
| `spec.blobStorage.mediaUpload.maxContentLengthBytes` | `int64` |  | `1000000000` |  |
| `spec.blobStorage.mediaUpload.downloadUrlExpirySeconds` | `int32` |  | `3600` |  |
| `spec.blobStorage.mediaUpload.sse` | `KubernetesLangfuseServerSideEncryption` |  |  |  |
| `spec.blobStorage.mediaUpload.sse.algorithm` | `enum` |  |  |  |
| `spec.blobStorage.mediaUpload.sse.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.blobStorage.batchExport` | `KubernetesLangfuseBatchExport` |  |  |  |
| `spec.blobStorage.batchExport.enabled` | `bool` |  | `true` |  |
| `spec.blobStorage.batchExport.target` | `KubernetesLangfuseBlobTarget` |  |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible` | `KubernetesLangfuseS3CompatibleTarget` |  |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible.endpoint` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.batchExport.target.s3Compatible.bucket` | `string` | yes |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible.region` | `string` |  | `auto` |  |
| `spec.blobStorage.batchExport.target.s3Compatible.forcePathStyle` | `bool` |  | `true` |  |
| `spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret` | `KubernetesLangfuseS3CredentialsSecret` |  |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`) |
| `spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.accessKeyIdKey` | `string` |  | `admin_access_key_id` |  |
| `spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.secretAccessKeyKey` | `string` |  | `admin_secret_access_key` |  |
| `spec.blobStorage.batchExport.target.s3Compatible.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.batchExport.target.s3Compatible.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.batchExport.target.s3` | `KubernetesLangfuseS3Target` |  |  |  |
| `spec.blobStorage.batchExport.target.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.blobStorage.batchExport.target.s3.region` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.region`) |
| `spec.blobStorage.batchExport.target.s3.endpoint` | `string` |  |  |  |
| `spec.blobStorage.batchExport.target.s3.accessKey` | `KubernetesLangfuseS3AccessKey` |  |  |  |
| `spec.blobStorage.batchExport.target.s3.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.blobStorage.batchExport.target.s3.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.batchExport.target.r2` | `KubernetesLangfuseR2Target` |  |  |  |
| `spec.blobStorage.batchExport.target.r2.endpoint` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.s3_endpoint`) |
| `spec.blobStorage.batchExport.target.r2.bucket` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.batchExport.target.r2.credentials` | `KubernetesLangfuseR2Credentials` | yes |  |  |
| `spec.blobStorage.batchExport.target.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.blobStorage.batchExport.target.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.blobStorage.batchExport.target.gcs` | `KubernetesLangfuseGcsTarget` |  |  |  |
| `spec.blobStorage.batchExport.target.gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.blobStorage.batchExport.target.gcs.credentialsJson` | `string` (sensitive) |  |  |  |
| `spec.blobStorage.batchExport.target.azureBlob` | `KubernetesLangfuseAzureBlobTarget` |  |  |  |
| `spec.blobStorage.batchExport.target.azureBlob.storageAccount` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.blobStorage.batchExport.target.azureBlob.container` | `string \| valueFrom` | yes |  | AzureStorageContainer (`status.outputs.container_name`) |
| `spec.blobStorage.batchExport.target.azureBlob.accountKey` | `string` (sensitive) | yes |  |  |
| `spec.blobStorage.batchExport.target.azureBlob.endpoint` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.primary_blob_endpoint`) |
| `spec.blobStorage.batchExport.prefix` | `string` |  | `exports/` |  |
| `spec.blobStorage.batchExport.publicEndpoint` | `string` |  |  |  |
| `spec.blobStorage.batchExport.pageSize` | `int32` |  |  |  |
| `spec.blobStorage.batchExport.rowLimit` | `int32` |  |  |  |
| `spec.blobStorage.batchExport.downloadLinkExpirationHours` | `int32` |  |  |  |
| `spec.blobStorage.batchExport.sse` | `KubernetesLangfuseServerSideEncryption` |  |  |  |
| `spec.blobStorage.batchExport.sse.algorithm` | `enum` |  |  |  |
| `spec.blobStorage.batchExport.sse.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.blobStorage.concurrentReads` | `int32` |  | `50` |  |
| `spec.blobStorage.concurrentWrites` | `int32` |  | `50` |  |
| `spec.blobStorage.deleteObjectsChecksumAlgorithm` | `enum` |  |  |  |
| `spec.secrets` | `KubernetesLangfuseInstanceSecrets` |  |  |  |
| `spec.secrets.salt` | `string` (sensitive) |  |  |  |
| `spec.secrets.encryptionKey` | `string` (sensitive) |  |  |  |
| `spec.secrets.nextauthSecret` | `string` (sensitive) |  |  |  |
| `spec.init` | `KubernetesLangfuseInit` |  |  |  |
| `spec.init.enabled` | `bool` |  | `true` |  |
| `spec.init.organization` | `KubernetesLangfuseInitOrganization` |  |  |  |
| `spec.init.organization.id` | `string` |  |  |  |
| `spec.init.organization.name` | `string` |  | `Default Organization` |  |
| `spec.init.project` | `KubernetesLangfuseInitProject` |  |  |  |
| `spec.init.project.id` | `string` |  |  |  |
| `spec.init.project.name` | `string` |  | `Default Project` |  |
| `spec.init.project.retentionDays` | `int32` |  |  |  |
| `spec.init.project.publicKey` | `string` |  |  |  |
| `spec.init.project.secretKey` | `string` (sensitive) |  |  |  |
| `spec.init.admin` | `KubernetesLangfuseInitAdmin` |  |  |  |
| `spec.init.admin.email` | `string` | yes |  |  |
| `spec.init.admin.name` | `string` |  | `Admin` |  |
| `spec.init.admin.password` | `string` (sensitive) |  |  |  |
| `spec.auth` | `KubernetesLangfuseAuth` |  |  |  |
| `spec.auth.disablePasswordSignIn` | `bool` |  |  |  |
| `spec.auth.disableSignUp` | `bool` |  |  |  |
| `spec.auth.emailVerificationRequired` | `bool` |  |  |  |
| `spec.auth.sessionMaxAgeMinutes` | `int32` |  |  |  |
| `spec.auth.ssoEnforcedDomains` | `[]string` |  |  |  |
| `spec.auth.ignoreAccountFields` | `[]string` |  |  |  |
| `spec.auth.httpsProxy` | `string` |  |  |  |
| `spec.auth.httpProxy` | `string` |  |  |  |
| `spec.auth.newUserAccess` | `KubernetesLangfuseNewUserAccess` |  |  |  |
| `spec.auth.newUserAccess.organizationId` | `string` |  |  |  |
| `spec.auth.newUserAccess.organizationRole` | `enum` |  |  |  |
| `spec.auth.newUserAccess.projectIds` | `[]string` |  |  |  |
| `spec.auth.newUserAccess.projectRole` | `enum` |  |  |  |
| `spec.auth.sso` | `KubernetesLangfuseSsoProviders` |  |  |  |
| `spec.auth.sso.google` | `KubernetesLangfuseGoogleSso` |  |  |  |
| `spec.auth.sso.google.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.google.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.google.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.google.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.google.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.google.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.google.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.google.allowedDomains` | `[]string` |  |  |  |
| `spec.auth.sso.github` | `KubernetesLangfuseGitHubSso` |  |  |  |
| `spec.auth.sso.github.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.github.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.github.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.github.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.github.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.github.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.github.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.githubEnterprise` | `KubernetesLangfuseGitHubEnterpriseSso` |  |  |  |
| `spec.auth.sso.githubEnterprise.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.githubEnterprise.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.githubEnterprise.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.githubEnterprise.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.githubEnterprise.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.githubEnterprise.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.githubEnterprise.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.githubEnterprise.baseUrl` | `string` |  |  |  |
| `spec.auth.sso.gitlab` | `KubernetesLangfuseGitLabSso` |  |  |  |
| `spec.auth.sso.gitlab.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.gitlab.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.gitlab.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.gitlab.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.gitlab.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.gitlab.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.gitlab.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.gitlab.url` | `string` |  |  |  |
| `spec.auth.sso.gitlab.issuer` | `string` |  |  |  |
| `spec.auth.sso.azureAd` | `KubernetesLangfuseAzureAdSso` |  |  |  |
| `spec.auth.sso.azureAd.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.azureAd.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.azureAd.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.azureAd.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.azureAd.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.azureAd.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.azureAd.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.azureAd.tenantId` | `string` | yes |  |  |
| `spec.auth.sso.okta` | `KubernetesLangfuseIssuerSso` |  |  |  |
| `spec.auth.sso.okta.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.okta.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.okta.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.okta.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.okta.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.okta.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.okta.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.okta.issuer` | `string` | yes |  |  |
| `spec.auth.sso.authentik` | `KubernetesLangfuseAuthentikSso` |  |  |  |
| `spec.auth.sso.authentik.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.authentik.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.authentik.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.authentik.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.authentik.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.authentik.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.authentik.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.authentik.issuer` | `string` | yes |  |  |
| `spec.auth.sso.authentik.authorizationUrl` | `string` |  |  |  |
| `spec.auth.sso.onelogin` | `KubernetesLangfuseIssuerSso` |  |  |  |
| `spec.auth.sso.onelogin.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.onelogin.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.onelogin.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.onelogin.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.onelogin.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.onelogin.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.onelogin.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.onelogin.issuer` | `string` | yes |  |  |
| `spec.auth.sso.auth0` | `KubernetesLangfuseIssuerSso` |  |  |  |
| `spec.auth.sso.auth0.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.auth0.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.auth0.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.auth0.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.auth0.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.auth0.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.auth0.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.auth0.issuer` | `string` | yes |  |  |
| `spec.auth.sso.cognito` | `KubernetesLangfuseIssuerSso` |  |  |  |
| `spec.auth.sso.cognito.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.cognito.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.cognito.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.cognito.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.cognito.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.cognito.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.cognito.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.cognito.issuer` | `string` | yes |  |  |
| `spec.auth.sso.keycloak` | `KubernetesLangfuseKeycloakSso` |  |  |  |
| `spec.auth.sso.keycloak.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.keycloak.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.keycloak.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.keycloak.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.keycloak.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.keycloak.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.keycloak.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.keycloak.issuer` | `string` | yes |  |  |
| `spec.auth.sso.keycloak.name` | `string` |  |  |  |
| `spec.auth.sso.keycloak.scope` | `string` |  |  |  |
| `spec.auth.sso.keycloak.useIdToken` | `bool` |  |  |  |
| `spec.auth.sso.jumpcloud` | `KubernetesLangfuseJumpCloudSso` |  |  |  |
| `spec.auth.sso.jumpcloud.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.jumpcloud.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.jumpcloud.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.jumpcloud.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.jumpcloud.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.jumpcloud.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.jumpcloud.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.jumpcloud.issuer` | `string` | yes |  |  |
| `spec.auth.sso.jumpcloud.scope` | `string` |  |  |  |
| `spec.auth.sso.custom` | `KubernetesLangfuseCustomSso` |  |  |  |
| `spec.auth.sso.custom.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.custom.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.custom.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.custom.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.custom.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.custom.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.custom.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.custom.issuer` | `string` | yes |  |  |
| `spec.auth.sso.custom.name` | `string` | yes |  |  |
| `spec.auth.sso.custom.scope` | `string` |  |  |  |
| `spec.auth.sso.custom.useIdToken` | `bool` |  |  |  |
| `spec.auth.sso.custom.fetchUserinfo` | `bool` |  |  |  |
| `spec.auth.sso.custom.claims` | `KubernetesLangfuseSsoClaims` |  |  |  |
| `spec.auth.sso.custom.claims.email` | `string` |  |  |  |
| `spec.auth.sso.custom.claims.name` | `string` |  |  |  |
| `spec.auth.sso.custom.claims.subject` | `string` |  |  |  |
| `spec.auth.sso.custom.claims.image` | `string` |  |  |  |
| `spec.auth.sso.workos` | `KubernetesLangfuseWorkOsSso` |  |  |  |
| `spec.auth.sso.workos.client` | `KubernetesLangfuseOAuthClient` | yes |  |  |
| `spec.auth.sso.workos.client.clientId` | `string` | yes |  |  |
| `spec.auth.sso.workos.client.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.workos.client.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.workos.client.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.workos.client.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.workos.client.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.sso.workos.organizationId` | `string` |  |  |  |
| `spec.auth.sso.workos.connectionId` | `string` |  |  |  |
| `spec.auth.sso.wordpress` | `KubernetesLangfuseOAuthClient` |  |  |  |
| `spec.auth.sso.wordpress.clientId` | `string` | yes |  |  |
| `spec.auth.sso.wordpress.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.auth.sso.wordpress.allowAccountLinking` | `bool` |  |  |  |
| `spec.auth.sso.wordpress.clientAuthMethod` | `enum` |  |  |  |
| `spec.auth.sso.wordpress.checks` | `[]enum` |  |  |  |
| `spec.auth.sso.wordpress.idTokenSignedResponseAlg` | `enum` |  |  |  |
| `spec.auth.ssoTimeoutMs` | `int32` |  |  |  |
| `spec.smtp` | `KubernetesLangfuseSmtp` |  |  |  |
| `spec.smtp.connectionUrl` | `string` (sensitive) | yes |  |  |
| `spec.smtp.fromAddress` | `string` | yes |  |  |
| `spec.features` | `KubernetesLangfuseFeatures` |  |  |  |
| `spec.features.telemetryEnabled` | `bool` |  | `true` |  |
| `spec.features.experimentalFeaturesEnabled` | `bool` |  |  |  |
| `spec.features.allowedOrganizationCreators` | `[]string` |  |  |  |
| `spec.features.cspEnforceHttps` | `bool` |  |  |  |
| `spec.features.markdownRenderCharacterLimit` | `int32` |  |  |  |
| `spec.caching` | `KubernetesLangfuseCaching` |  |  |  |
| `spec.caching.apiKeyCacheEnabled` | `bool` |  | `true` |  |
| `spec.caching.apiKeyCacheTtlSeconds` | `int32` |  |  |  |
| `spec.caching.promptCacheEnabled` | `bool` |  | `true` |  |
| `spec.caching.promptCacheTtlSeconds` | `int32` |  |  |  |
| `spec.ingestion` | `KubernetesLangfuseIngestion` |  |  |  |
| `spec.ingestion.queueShardCount` | `int32` |  |  |  |
| `spec.ingestion.queueProcessingConcurrency` | `int32` |  |  |  |
| `spec.ingestion.traceUpsertQueueShardCount` | `int32` |  |  |  |
| `spec.ingestion.traceUpsertWorkerConcurrency` | `int32` |  |  |  |
| `spec.ingestion.skipClickhouseReadMinProjectCreateDate` | `string` |  |  |  |
| `spec.aiFeatures` | `KubernetesLangfuseAiFeatures` |  |  |  |
| `spec.aiFeatures.provider` | `enum` |  |  |  |
| `spec.aiFeatures.model` | `string` |  |  |  |
| `spec.aiFeatures.smallModel` | `string` |  |  |  |
| `spec.aiFeatures.apiKey` | `string` (sensitive) |  |  |  |
| `spec.aiFeatures.baseUrl` | `string` |  |  |  |
| `spec.aiFeatures.extraHeaders` | `map<string, string>` |  |  |  |
| `spec.aiFeatures.useResponsesApi` | `bool` |  |  |  |
| `spec.aiFeatures.bedrockRegion` | `string` |  |  |  |
| `spec.aiFeatures.vertexLocation` | `string` |  |  |  |
| `spec.aiFeatures.traceProjectId` | `string` |  |  |  |
| `spec.aiFeatures.assistant` | `KubernetesLangfuseAssistant` |  |  |  |
| `spec.aiFeatures.assistant.enabled` | `bool` |  |  |  |
| `spec.aiFeatures.assistant.runConcurrency` | `int32` |  |  |  |
| `spec.aiFeatures.assistant.maxActiveRunsPerUser` | `int32` |  |  |  |
| `spec.aiFeatures.assistant.maxActiveRunsPerOrg` | `int32` |  |  |  |
| `spec.aiFeatures.assistant.useInternalWebUrl` | `bool` |  |  |  |
| `spec.aiFeatures.assistant.sandbox` | `KubernetesLangfuseLambdaMicrovmSandbox` |  |  |  |
| `spec.aiFeatures.assistant.sandbox.imageIdentifier` | `string` | yes |  |  |
| `spec.aiFeatures.assistant.sandbox.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.aiFeatures.assistant.sandbox.region` | `string` | yes |  |  |
| `spec.aiFeatures.assistant.sandbox.egressNetworkConnectorArn` | `string` | yes |  |  |
| `spec.evaluation` | `KubernetesLangfuseEvaluation` |  |  |  |
| `spec.evaluation.mediaTransport` | `enum` |  |  |  |
| `spec.evaluation.mediaInlineMaxBytes` | `int32` |  |  |  |
| `spec.evaluation.codeEvaluators` | `KubernetesLangfuseCodeEvaluators` |  |  |  |
| `spec.evaluation.codeEvaluators.awsLambda` | `KubernetesLangfuseLambdaCodeEvaluators` |  |  |  |
| `spec.evaluation.codeEvaluators.awsLambda.nodeFunctionName` | `string` |  |  |  |
| `spec.evaluation.codeEvaluators.awsLambda.pythonFunctionName` | `string` |  |  |  |
| `spec.evaluation.codeEvaluators.awsLambda.endpoint` | `string` |  |  |  |
| `spec.evaluation.codeEvaluators.insecureLocal` | `KubernetesLangfuseLocalCodeEvaluators` |  |  |  |
| `spec.evaluation.codeEvaluators.insecureLocal.timeoutMs` | `int32` |  |  |  |
| `spec.evaluation.codeEvaluators.queueShardCount` | `int32` |  |  |  |
| `spec.evaluation.codeEvaluators.concurrency` | `int32` |  |  |  |
| `spec.slack` | `KubernetesLangfuseSlack` |  |  |  |
| `spec.slack.clientId` | `string` | yes |  |  |
| `spec.slack.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.slack.stateSecret` | `string` (sensitive) |  |  |  |
| `spec.enterprise` | `KubernetesLangfuseEnterprise` |  |  |  |
| `spec.enterprise.licenseKey` | `string` (sensitive) |  |  |  |
| `spec.enterprise.adminApiKey` | `string` (sensitive) |  |  |  |
| `spec.enterprise.ui` | `KubernetesLangfuseUiCustomization` |  |  |  |
| `spec.enterprise.ui.documentationHref` | `string` |  |  |  |
| `spec.enterprise.ui.supportHref` | `string` |  |  |  |
| `spec.enterprise.ui.feedbackHref` | `string` |  |  |  |
| `spec.enterprise.ui.apiHost` | `string` |  |  |  |
| `spec.enterprise.ui.logoLightModeHref` | `string` |  |  |  |
| `spec.enterprise.ui.logoDarkModeHref` | `string` |  |  |  |
| `spec.enterprise.ui.defaultModelAdapter` | `enum` |  |  |  |
| `spec.enterprise.ui.defaultBaseUrlOpenai` | `string` |  |  |  |
| `spec.enterprise.ui.defaultBaseUrlAnthropic` | `string` |  |  |  |
| `spec.enterprise.ui.defaultBaseUrlAzure` | `string` |  |  |  |
| `spec.enterprise.ui.visibleProductModules` | `[]string` |  |  |  |
| `spec.enterprise.ui.hiddenProductModules` | `[]string` |  |  |  |
| `spec.enterprise.ui.instanceLinks` | `[]KubernetesLangfuseInstanceLink` |  |  |  |
| `spec.enterprise.ui.instanceLinks[].name` | `string` | yes |  |  |
| `spec.enterprise.ui.instanceLinks[].url` | `string` | yes |  |  |
| `spec.enterprise.ingestionMasking` | `KubernetesLangfuseIngestionMasking` |  |  |  |
| `spec.enterprise.ingestionMasking.callbackUrl` | `string` | yes |  |  |
| `spec.enterprise.ingestionMasking.timeoutMs` | `int32` |  |  |  |
| `spec.enterprise.ingestionMasking.failClosed` | `bool` |  |  |  |
| `spec.enterprise.ingestionMasking.maxRetries` | `int32` |  |  |  |
| `spec.enterprise.ingestionMasking.propagatedHeaders` | `[]string` |  |  |  |
| `spec.observability` | `KubernetesLangfuseObservability` |  |  |  |
| `spec.observability.logLevel` | `enum` |  |  |  |
| `spec.observability.logFormat` | `enum` |  |  |  |
| `spec.observability.logPropagatedHeaders` | `[]string` |  |  |  |
| `spec.observability.otlpEndpoint` | `string \| valueFrom` |  |  | KubernetesOtelCollector (`status.outputs.otlp_http_endpoint`) |
| `spec.observability.traceSamplingRatio` | `double` |  |  |  |
| `spec.network` | `KubernetesLangfuseNetwork` |  |  |  |
| `spec.network.httpsProxy` | `string` |  |  |  |
| `spec.network.noProxy` | `[]string` |  |  |  |
| `spec.network.trustedCa` | `KubernetesLangfuseCertificateFile` |  |  |  |
| `spec.network.trustedCa.secretName` | `string` |  |  |  |
| `spec.network.trustedCa.configMapName` | `string` |  |  |  |
| `spec.network.trustedCa.key` | `string` |  | `ca.crt` |  |
| `spec.network.allowlists` | `KubernetesLangfuseOutboundAllowlists` |  |  |  |
| `spec.network.allowlists.webhooks` | `KubernetesLangfuseAllowlist` |  |  |  |
| `spec.network.allowlists.webhooks.hosts` | `[]string` |  |  |  |
| `spec.network.allowlists.webhooks.ips` | `[]string` |  |  |  |
| `spec.network.allowlists.webhooks.ipSegments` | `[]string` |  |  |  |
| `spec.network.allowlists.llmConnections` | `KubernetesLangfuseAllowlist` |  |  |  |
| `spec.network.allowlists.llmConnections.hosts` | `[]string` |  |  |  |
| `spec.network.allowlists.llmConnections.ips` | `[]string` |  |  |  |
| `spec.network.allowlists.llmConnections.ipSegments` | `[]string` |  |  |  |
| `spec.network.allowlists.blobStorageEndpoints` | `KubernetesLangfuseAllowlist` |  |  |  |
| `spec.network.allowlists.blobStorageEndpoints.hosts` | `[]string` |  |  |  |
| `spec.network.allowlists.blobStorageEndpoints.ips` | `[]string` |  |  |  |
| `spec.network.allowlists.blobStorageEndpoints.ipSegments` | `[]string` |  |  |  |
| `spec.network.allowlists.externalMediaStorageEndpoints` | `KubernetesLangfuseAllowlist` |  |  |  |
| `spec.network.allowlists.externalMediaStorageEndpoints.hosts` | `[]string` |  |  |  |
| `spec.network.allowlists.externalMediaStorageEndpoints.ips` | `[]string` |  |  |  |
| `spec.network.allowlists.externalMediaStorageEndpoints.ipSegments` | `[]string` |  |  |  |
| `spec.network.allowlists.ssoDiscovery` | `KubernetesLangfuseAllowlist` |  |  |  |
| `spec.network.allowlists.ssoDiscovery.hosts` | `[]string` |  |  |  |
| `spec.network.allowlists.ssoDiscovery.ips` | `[]string` |  |  |  |
| `spec.network.allowlists.ssoDiscovery.ipSegments` | `[]string` |  |  |  |
| `spec.web` | `KubernetesLangfuseWeb` |  |  |  |
| `spec.web.image` | `KubernetesLangfuseImage` |  |  |  |
| `spec.web.image.repository` | `string` |  |  |  |
| `spec.web.image.tag` | `string` |  |  |  |
| `spec.web.image.pullPolicy` | `string` |  |  |  |
| `spec.web.image.pullSecretNames` | `[]string` |  |  |  |
| `spec.web.replicas` | `int32` |  | `1` |  |
| `spec.web.resources` | `ContainerResources` |  |  |  |
| `spec.web.resources.limits` | `CpuMemory` |  |  |  |
| `spec.web.resources.limits.cpu` | `string` |  |  |  |
| `spec.web.resources.limits.memory` | `string` |  |  |  |
| `spec.web.resources.requests` | `CpuMemory` |  |  |  |
| `spec.web.resources.requests.cpu` | `string` |  |  |  |
| `spec.web.resources.requests.memory` | `string` |  |  |  |
| `spec.web.autoscaling` | `KubernetesLangfuseAutoscaling` |  |  |  |
| `spec.web.autoscaling.hpa` | `KubernetesLangfuseHpa` |  |  |  |
| `spec.web.autoscaling.hpa.minReplicas` | `int32` |  | `1` |  |
| `spec.web.autoscaling.hpa.maxReplicas` | `int32` |  | `2` |  |
| `spec.web.autoscaling.hpa.targetCpuUtilizationPercent` | `int32` |  | `50` |  |
| `spec.web.autoscaling.hpa.targetMemoryUtilizationPercent` | `int32` |  |  |  |
| `spec.web.autoscaling.hpa.annotations` | `map<string, string>` |  |  |  |
| `spec.web.autoscaling.keda` | `KubernetesLangfuseKeda` |  |  |  |
| `spec.web.autoscaling.keda.minReplicas` | `int32` |  | `1` |  |
| `spec.web.autoscaling.keda.maxReplicas` | `int32` |  | `2` |  |
| `spec.web.autoscaling.keda.pollingIntervalSeconds` | `int32` |  | `30` |  |
| `spec.web.autoscaling.keda.trigger` | `enum` |  |  |  |
| `spec.web.autoscaling.keda.metricType` | `enum` |  |  |  |
| `spec.web.autoscaling.keda.targetValue` | `string` |  | `50` |  |
| `spec.web.autoscaling.keda.containerName` | `string` |  |  |  |
| `spec.web.autoscaling.vpa` | `KubernetesLangfuseVpa` |  |  |  |
| `spec.web.autoscaling.vpa.controlledResources` | `[]string` |  |  |  |
| `spec.web.autoscaling.vpa.maxAllowed` | `CpuMemory` |  |  |  |
| `spec.web.autoscaling.vpa.maxAllowed.cpu` | `string` |  |  |  |
| `spec.web.autoscaling.vpa.maxAllowed.memory` | `string` |  |  |  |
| `spec.web.autoscaling.vpa.minAllowed` | `CpuMemory` |  |  |  |
| `spec.web.autoscaling.vpa.minAllowed.cpu` | `string` |  |  |  |
| `spec.web.autoscaling.vpa.minAllowed.memory` | `string` |  |  |  |
| `spec.web.autoscaling.vpa.updateMode` | `string` |  | `Auto` |  |
| `spec.web.pdb` | `KubernetesLangfusePdb` |  |  |  |
| `spec.web.pdb.enabled` | `bool` |  | `true` |  |
| `spec.web.pdb.minAvailable` | `string` |  |  |  |
| `spec.web.pdb.maxUnavailable` | `string` |  |  |  |
| `spec.web.livenessProbe` | `KubernetesLangfuseProbe` |  |  |  |
| `spec.web.livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.web.livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.web.livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.web.livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.web.livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.web.readinessProbe` | `KubernetesLangfuseProbe` |  |  |  |
| `spec.web.readinessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.web.readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.web.readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.web.readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.web.readinessProbe.successThreshold` | `int32` |  |  |  |
| `spec.web.service` | `KubernetesLangfuseWebService` |  |  |  |
| `spec.web.service.type` | `string` |  | `ClusterIP` |  |
| `spec.web.service.port` | `int32` |  | `3000` |  |
| `spec.web.service.externalPort` | `int32` |  |  |  |
| `spec.web.service.nodePort` | `int32` |  |  |  |
| `spec.web.service.annotations` | `map<string, string>` |  |  |  |
| `spec.web.service.labels` | `map<string, string>` |  |  |  |
| `spec.web.scheduling` | `KubernetesLangfuseScheduling` |  |  |  |
| `spec.web.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.web.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.web.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.web.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.web.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.web.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.web.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.web.scheduling.nodeAffinity` | `WorkloadNodeAffinity` |  |  |  |
| `spec.web.scheduling.nodeAffinity.required` | `[]WorkloadNodeSelectorTerm` |  |  |  |
| `spec.web.scheduling.nodeAffinity.required[].matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.required[].matchExpressions[].key` | `string` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.required[].matchExpressions[].operator` | `string` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.required[].matchExpressions[].values` | `[]string` |  |  |  |
| `spec.web.scheduling.nodeAffinity.preferred` | `[]WorkloadPreferredNodeSelectorTerm` |  |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].term` | `WorkloadNodeSelectorTerm` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key` | `string` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.web.scheduling.podAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.web.scheduling.podAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.web.scheduling.podAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.web.scheduling.podAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.web.scheduling.podAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.web.scheduling.podAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.web.scheduling.podAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.web.scheduling.podAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.web.scheduling.podAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.web.scheduling.podAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.web.scheduling.podAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.web.scheduling.podAntiAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.web.scheduling.podAntiAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.web.scheduling.podAntiAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.web.scheduling.podAntiAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.web.scheduling.podAntiAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.web.scheduling.podAntiAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.web.scheduling.topologySpreadConstraints` | `[]WorkloadTopologySpreadConstraint` |  |  |  |
| `spec.web.scheduling.topologySpreadConstraints[].maxSkew` | `int32` |  |  |  |
| `spec.web.scheduling.topologySpreadConstraints[].topologyKey` | `string` | yes |  |  |
| `spec.web.scheduling.topologySpreadConstraints[].whenUnsatisfiable` | `string` | yes |  |  |
| `spec.web.scheduling.topologySpreadConstraints[].matchLabels` | `map<string, string>` |  |  |  |
| `spec.web.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.web.workload` | `KubernetesLangfuseWorkloadShape` |  |  |  |
| `spec.web.workload.podAnnotations` | `map<string, string>` |  |  |  |
| `spec.web.workload.podLabels` | `map<string, string>` |  |  |  |
| `spec.web.workload.deploymentAnnotations` | `map<string, string>` |  |  |  |
| `spec.web.workload.deploymentLabels` | `map<string, string>` |  |  |  |
| `spec.web.workload.strategy` | `KubernetesLangfuseRolloutStrategy` |  |  |  |
| `spec.web.workload.strategy.type` | `string` |  |  |  |
| `spec.web.workload.strategy.maxSurge` | `string` |  |  |  |
| `spec.web.workload.strategy.maxUnavailable` | `string` |  |  |  |
| `spec.web.workload.revisionHistoryLimit` | `int32` |  |  |  |
| `spec.web.workload.hostAliases` | `[]WorkloadHostAlias` |  |  |  |
| `spec.web.workload.hostAliases[].ip` | `string` | yes |  |  |
| `spec.web.workload.hostAliases[].hostnames` | `[]string` | yes |  |  |
| `spec.web.extraEnv` | `map<string, string>` (no secrets: use `extraEnvFromSecret`) |  |  |  |
| `spec.web.extraEnvFromSecret` | `map<string, KubernetesSecretKey>` |  |  |  |
| `spec.web.extraEnvFromSecret.*.name` | `string` |  |  |  |
| `spec.web.extraEnvFromSecret.*.key` | `string` |  |  |  |
| `spec.worker` | `KubernetesLangfuseWorker` |  |  |  |
| `spec.worker.image` | `KubernetesLangfuseImage` |  |  |  |
| `spec.worker.image.repository` | `string` |  |  |  |
| `spec.worker.image.tag` | `string` |  |  |  |
| `spec.worker.image.pullPolicy` | `string` |  |  |  |
| `spec.worker.image.pullSecretNames` | `[]string` |  |  |  |
| `spec.worker.replicas` | `int32` |  | `1` |  |
| `spec.worker.resources` | `ContainerResources` |  |  |  |
| `spec.worker.resources.limits` | `CpuMemory` |  |  |  |
| `spec.worker.resources.limits.cpu` | `string` |  |  |  |
| `spec.worker.resources.limits.memory` | `string` |  |  |  |
| `spec.worker.resources.requests` | `CpuMemory` |  |  |  |
| `spec.worker.resources.requests.cpu` | `string` |  |  |  |
| `spec.worker.resources.requests.memory` | `string` |  |  |  |
| `spec.worker.autoscaling` | `KubernetesLangfuseAutoscaling` |  |  |  |
| `spec.worker.autoscaling.hpa` | `KubernetesLangfuseHpa` |  |  |  |
| `spec.worker.autoscaling.hpa.minReplicas` | `int32` |  | `1` |  |
| `spec.worker.autoscaling.hpa.maxReplicas` | `int32` |  | `2` |  |
| `spec.worker.autoscaling.hpa.targetCpuUtilizationPercent` | `int32` |  | `50` |  |
| `spec.worker.autoscaling.hpa.targetMemoryUtilizationPercent` | `int32` |  |  |  |
| `spec.worker.autoscaling.hpa.annotations` | `map<string, string>` |  |  |  |
| `spec.worker.autoscaling.keda` | `KubernetesLangfuseKeda` |  |  |  |
| `spec.worker.autoscaling.keda.minReplicas` | `int32` |  | `1` |  |
| `spec.worker.autoscaling.keda.maxReplicas` | `int32` |  | `2` |  |
| `spec.worker.autoscaling.keda.pollingIntervalSeconds` | `int32` |  | `30` |  |
| `spec.worker.autoscaling.keda.trigger` | `enum` |  |  |  |
| `spec.worker.autoscaling.keda.metricType` | `enum` |  |  |  |
| `spec.worker.autoscaling.keda.targetValue` | `string` |  | `50` |  |
| `spec.worker.autoscaling.keda.containerName` | `string` |  |  |  |
| `spec.worker.autoscaling.vpa` | `KubernetesLangfuseVpa` |  |  |  |
| `spec.worker.autoscaling.vpa.controlledResources` | `[]string` |  |  |  |
| `spec.worker.autoscaling.vpa.maxAllowed` | `CpuMemory` |  |  |  |
| `spec.worker.autoscaling.vpa.maxAllowed.cpu` | `string` |  |  |  |
| `spec.worker.autoscaling.vpa.maxAllowed.memory` | `string` |  |  |  |
| `spec.worker.autoscaling.vpa.minAllowed` | `CpuMemory` |  |  |  |
| `spec.worker.autoscaling.vpa.minAllowed.cpu` | `string` |  |  |  |
| `spec.worker.autoscaling.vpa.minAllowed.memory` | `string` |  |  |  |
| `spec.worker.autoscaling.vpa.updateMode` | `string` |  | `Auto` |  |
| `spec.worker.pdb` | `KubernetesLangfusePdb` |  |  |  |
| `spec.worker.pdb.enabled` | `bool` |  | `true` |  |
| `spec.worker.pdb.minAvailable` | `string` |  |  |  |
| `spec.worker.pdb.maxUnavailable` | `string` |  |  |  |
| `spec.worker.livenessProbe` | `KubernetesLangfuseProbe` |  |  |  |
| `spec.worker.livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.worker.livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.worker.livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.worker.livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.worker.livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.worker.failLivenessIfQueueConsumptionStuck` | `bool` |  |  |  |
| `spec.worker.queueConsumptionStuckThresholdMinutes` | `int32` |  |  |  |
| `spec.worker.scheduling` | `KubernetesLangfuseScheduling` |  |  |  |
| `spec.worker.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.worker.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.worker.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.worker.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.worker.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.worker.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.worker.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.worker.scheduling.nodeAffinity` | `WorkloadNodeAffinity` |  |  |  |
| `spec.worker.scheduling.nodeAffinity.required` | `[]WorkloadNodeSelectorTerm` |  |  |  |
| `spec.worker.scheduling.nodeAffinity.required[].matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].key` | `string` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].operator` | `string` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].values` | `[]string` |  |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred` | `[]WorkloadPreferredNodeSelectorTerm` |  |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].term` | `WorkloadNodeSelectorTerm` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions` | `[]WorkloadNodeSelectorRequirement` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key` | `string` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.worker.scheduling.podAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.worker.scheduling.podAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.worker.scheduling.podAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.worker.scheduling.podAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.worker.scheduling.podAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.worker.scheduling.podAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.worker.scheduling.podAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.worker.scheduling.podAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.worker.scheduling.podAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.worker.scheduling.podAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.worker.scheduling.podAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity` | `WorkloadPodAffinity` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity.required` | `[]WorkloadPodAffinityTerm` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity.required[].matchLabels` | `map<string, string>` | yes |  |  |
| `spec.worker.scheduling.podAntiAffinity.required[].topologyKey` | `string` | yes |  |  |
| `spec.worker.scheduling.podAntiAffinity.required[].namespaces` | `[]string` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred` | `[]WorkloadWeightedPodAffinityTerm` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred[].weight` | `int32` |  |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred[].term` | `WorkloadPodAffinityTerm` | yes |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred[].term.matchLabels` | `map<string, string>` | yes |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred[].term.topologyKey` | `string` | yes |  |  |
| `spec.worker.scheduling.podAntiAffinity.preferred[].term.namespaces` | `[]string` |  |  |  |
| `spec.worker.scheduling.topologySpreadConstraints` | `[]WorkloadTopologySpreadConstraint` |  |  |  |
| `spec.worker.scheduling.topologySpreadConstraints[].maxSkew` | `int32` |  |  |  |
| `spec.worker.scheduling.topologySpreadConstraints[].topologyKey` | `string` | yes |  |  |
| `spec.worker.scheduling.topologySpreadConstraints[].whenUnsatisfiable` | `string` | yes |  |  |
| `spec.worker.scheduling.topologySpreadConstraints[].matchLabels` | `map<string, string>` |  |  |  |
| `spec.worker.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.worker.workload` | `KubernetesLangfuseWorkloadShape` |  |  |  |
| `spec.worker.workload.podAnnotations` | `map<string, string>` |  |  |  |
| `spec.worker.workload.podLabels` | `map<string, string>` |  |  |  |
| `spec.worker.workload.deploymentAnnotations` | `map<string, string>` |  |  |  |
| `spec.worker.workload.deploymentLabels` | `map<string, string>` |  |  |  |
| `spec.worker.workload.strategy` | `KubernetesLangfuseRolloutStrategy` |  |  |  |
| `spec.worker.workload.strategy.type` | `string` |  |  |  |
| `spec.worker.workload.strategy.maxSurge` | `string` |  |  |  |
| `spec.worker.workload.strategy.maxUnavailable` | `string` |  |  |  |
| `spec.worker.workload.revisionHistoryLimit` | `int32` |  |  |  |
| `spec.worker.workload.hostAliases` | `[]WorkloadHostAlias` |  |  |  |
| `spec.worker.workload.hostAliases[].ip` | `string` | yes |  |  |
| `spec.worker.workload.hostAliases[].hostnames` | `[]string` | yes |  |  |
| `spec.worker.extraEnv` | `map<string, string>` (no secrets: use `extraEnvFromSecret`) |  |  |  |
| `spec.worker.extraEnvFromSecret` | `map<string, KubernetesSecretKey>` |  |  |  |
| `spec.worker.extraEnvFromSecret.*.name` | `string` |  |  |  |
| `spec.worker.extraEnvFromSecret.*.key` | `string` |  |  |  |
| `spec.serviceAccount` | `KubernetesLangfuseServiceAccount` |  |  |  |
| `spec.serviceAccount.existingName` | `string` |  |  |  |
| `spec.serviceAccount.annotations` | `map<string, string>` |  |  |  |
| `spec.serviceAccount.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.serviceAccount.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.serviceAccount.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.serviceAccount.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.serviceAccount.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.serviceAccount.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.serviceAccount.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.serviceAccount.workloadIdentity.aks.tenantId` | `string` |  |  |  |
| `spec.serviceAccount.automountToken` | `bool` |  | `true` |  |
| `spec.podSecurityContext` | `WorkloadPodSecurityContext` |  |  |  |
| `spec.podSecurityContext.runAsUser` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podSecurityContext.fsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.fsGroupChangePolicy` | `string` |  |  |  |
| `spec.podSecurityContext.supplementalGroups` | `[]int64` |  |  |  |
| `spec.podSecurityContext.sysctls` | `[]WorkloadSysctl` |  |  |  |
| `spec.podSecurityContext.sysctls[].name` | `string` | yes |  |  |
| `spec.podSecurityContext.sysctls[].value` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile` | `WorkloadSeccompProfile` |  |  |  |
| `spec.podSecurityContext.seccompProfile.type` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.containerSecurityContext` | `WorkloadContainerSecurityContext` |  |  |  |
| `spec.containerSecurityContext.privileged` | `bool` |  |  |  |
| `spec.containerSecurityContext.runAsUser` | `int64` |  |  |  |
| `spec.containerSecurityContext.runAsGroup` | `int64` |  |  |  |
| `spec.containerSecurityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.containerSecurityContext.readOnlyRootFilesystem` | `bool` |  |  |  |
| `spec.containerSecurityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.containerSecurityContext.capabilities` | `WorkloadCapabilities` |  |  |  |
| `spec.containerSecurityContext.capabilities.add` | `[]string` |  |  |  |
| `spec.containerSecurityContext.capabilities.drop` | `[]string` |  |  |  |
| `spec.containerSecurityContext.seccompProfile` | `WorkloadSeccompProfile` |  |  |  |
| `spec.containerSecurityContext.seccompProfile.type` | `string` | yes |  |  |
| `spec.containerSecurityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.dnsConfig` | `WorkloadPodDnsConfig` |  |  |  |
| `spec.dnsConfig.nameservers` | `[]string` |  |  |  |
| `spec.dnsConfig.searches` | `[]string` |  |  |  |
| `spec.dnsConfig.options` | `[]WorkloadPodDnsConfigOption` |  |  |  |
| `spec.dnsConfig.options[].name` | `string` | yes |  |  |
| `spec.dnsConfig.options[].value` | `string` |  |  |  |
| `spec.extraVolumes` | `[]VolumeMount` |  |  |  |
| `spec.extraVolumes[].name` | `string` | yes |  |  |
| `spec.extraVolumes[].mountPath` | `string` | yes |  |  |
| `spec.extraVolumes[].readOnly` | `bool` |  |  |  |
| `spec.extraVolumes[].subPath` | `string` |  |  |  |
| `spec.extraVolumes[].configMap` | `ConfigMapVolumeSource` |  |  |  |
| `spec.extraVolumes[].configMap.name` | `string` | yes |  |  |
| `spec.extraVolumes[].configMap.key` | `string` |  |  |  |
| `spec.extraVolumes[].configMap.path` | `string` |  |  |  |
| `spec.extraVolumes[].configMap.defaultMode` | `int32` |  |  |  |
| `spec.extraVolumes[].secret` | `SecretVolumeSource` |  |  |  |
| `spec.extraVolumes[].secret.name` | `string` | yes |  |  |
| `spec.extraVolumes[].secret.key` | `string` |  |  |  |
| `spec.extraVolumes[].secret.path` | `string` |  |  |  |
| `spec.extraVolumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.extraVolumes[].hostPath` | `HostPathVolumeSource` |  |  |  |
| `spec.extraVolumes[].hostPath.path` | `string` | yes |  |  |
| `spec.extraVolumes[].hostPath.type` | `string` |  |  |  |
| `spec.extraVolumes[].emptyDir` | `EmptyDirVolumeSource` |  |  |  |
| `spec.extraVolumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.extraVolumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.extraVolumes[].pvc` | `PvcVolumeSource` |  |  |  |
| `spec.extraVolumes[].pvc.claimName` | `string` | yes |  |  |
| `spec.extraVolumes[].pvc.readOnly` | `bool` |  |  |  |
| `spec.extraVolumes[].serviceAccountToken` | `ServiceAccountTokenVolumeSource` |  |  |  |
| `spec.extraVolumes[].serviceAccountToken.audience` | `string` | yes |  |  |
| `spec.extraVolumes[].serviceAccountToken.expirationSeconds` | `int64` |  |  |  |
| `spec.extraVolumes[].serviceAccountToken.path` | `string` |  |  |  |
| `spec.extraEnv` | `map<string, string>` (no secrets: use `extraEnvFromSecret`) |  |  |  |
| `spec.extraEnvFromSecret` | `map<string, KubernetesSecretKey>` |  |  |  |
| `spec.extraEnvFromSecret.*.name` | `string` |  |  |  |
| `spec.extraEnvFromSecret.*.key` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. Every Secret the spec
references by name must live in this namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "2.1.5" = Langfuse 4.56.0; each
chart release pins one Langfuse version as its appVersion). Versions
must exist in the SERVED index at https://langfuse.github.io/langfuse-k8s.
Chart 2.x needs Kubernetes 1.28 or newer.

- default: `2.1.5`

### spec.externalUrl

`string` · optional (explicit presence)

The public URL people open Langfuse at (`NEXTAUTH_URL`), e.g.
"https://langfuse.example.com": the base of sign-in redirects and of
every link Langfuse writes into emails, Slack messages and integration
settings. Scheme, host and optional port only - Langfuse serves from
the root of its host. The default suits a port-forward
(`port_forward_command` output); set the URL your KubernetesHttpRoute
or KubernetesIngress serves before enabling single sign-on.

- default: `http://localhost:3000`
- rule: external_url must be scheme://host[:port] with no path or trailing slash, such as https://langfuse.example.com (Langfuse serves from the root of its host).

### spec.postgres

`KubernetesLangfusePostgres` · required

The Postgres database: users, organizations, projects, prompts, API
keys and settings.

- rule: {"required":true}

### spec.postgres.host

`string | valueFrom` · required

Postgres host. Defaults compose a KubernetesPostgres resource's
read-write Service (always the current primary); any reachable
PostgreSQL 15+ host works as a literal (RDS, Aurora, Cloud SQL,
Azure Database for PostgreSQL).

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.postgres.port

`int32` · optional (explicit presence)

Postgres port.

- default: `5432`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.postgres.database

`string` · optional (explicit presence)

Database name. Create it before Langfuse starts (a KubernetesPostgres
bootstrap `database` with this user as owner does).

- default: `langfuse`

### spec.postgres.username

`string` · required

The user Langfuse connects as (a KubernetesPostgres bootstrap owner,
for example).

- rule: {"required":true}

### spec.postgres.passwordSecret

`KubernetesLangfusePostgresPasswordSecret` · required

The Secret holding that user's password, in Langfuse's namespace.
Defaults compose a KubernetesPostgres resource's application
credential Secret, which the operator keeps current across failovers.

- rule: {"required":true}

### spec.postgres.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesPostgres resource's
application credential Secret (`<name>-app`).

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.postgres.passwordSecret.secretKey

`string` · optional (explicit presence)

Key within the Secret.

- default: `password`

### spec.postgres.args

`string`

Extra connection parameters appended to the connection URL (e.g.
"sslmode=require&connection_limit=20").

### spec.postgres.directUrl

`string` · sensitive

A separate connection URL for schema migrations (`DIRECT_URL`), when
the main connection goes through a pooler such as PgBouncer or uses a
user without DDL rights. Contains the credentials, so it is written to
the instance Secret. Give this user long statement timeouts: large
migrations take a while.

### spec.postgres.shadowDatabaseUrl

`string` · sensitive

Connection URL of a shadow database (`SHADOW_DATABASE_URL`), needed
only when the user lacks CREATE DATABASE (common on managed
services) and migrations must diff against a scratch database.
Contains the credentials, so it is written to the instance Secret.

### spec.postgres.autoMigrate

`bool` · optional (explicit presence)

Run Postgres migrations automatically when the web server starts.
Turn off only when migrations run out of band; Langfuse refuses to
serve a schema it does not expect.

- default: `true`

### spec.clickhouse

`KubernetesLangfuseClickHouse` · required

The ClickHouse database: traces, observations, scores and sessions.

- rule: {"required":true}

### spec.clickhouse.host

`string | valueFrom` · required

ClickHouse host. Defaults compose a KubernetesClickHouse resource's
client Service; ClickHouse Cloud or any reachable server works as a
literal host name (no scheme - `secure` selects TLS).

- references: KubernetesClickHouse (`status.outputs.service_name`)
- rule: clickhouse.host is a bare host name; set secure: true for TLS instead of an https:// prefix.
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.clickhouse.httpPort

`int32` · optional (explicit presence)

HTTP interface port (queries). 8443 is the usual TLS port.

- default: `8123`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.clickhouse.nativePort

`int32` · optional (explicit presence)

Native protocol port (migrations). 9440 is the usual TLS port.

- default: `9000`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.clickhouse.secure

`bool`

Connect over TLS on both interfaces (https for queries, TLS for
migrations) - ClickHouse Cloud requires it. Set the TLS ports too.

### spec.clickhouse.database

`string` · optional (explicit presence)

Database name.

- default: `default`

### spec.clickhouse.username

`string` · required

The user Langfuse connects as (a KubernetesClickHouse user, for
example). On a KubernetesClickHouse, declare that user's `networks`
(e.g. "0.0.0.0/0" and "::/0"): a user declared without networks is
fenced to ClickHouse's own pods and Langfuse is refused with what
looks like a wrong-password error.

- rule: {"required":true}

### spec.clickhouse.passwordSecret

`KubernetesLangfuseClickHousePasswordSecret` · required

The Secret holding that user's password, in Langfuse's namespace.

- rule: {"required":true}

### spec.clickhouse.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesClickHouse
resource's auth Secret (`<name>-clickhouse-auth`, one key per user).

- references: KubernetesClickHouse (`status.outputs.auth_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.auth_secret_name}} -- a bare string does not parse

### spec.clickhouse.passwordSecret.secretKey

`string`

Key within the Secret. Empty = `clickhouse.username`, the key a
KubernetesClickHouse auth Secret stores each user's password under.

### spec.clickhouse.cluster

`KubernetesLangfuseClickHouseCluster`

Cluster mode: whether Langfuse runs its DDL `ON CLUSTER` with
replicated table engines.

### spec.clickhouse.cluster.enabled

`bool` · optional (explicit presence)

Run DDL `ON CLUSTER` with replicated table engines
(`CLICKHOUSE_CLUSTER_ENABLED`). Needs ClickHouse Keeper (or ZooKeeper)
coordination - a KubernetesClickHouse with replicas, or with a
managed Keeper. Turn off for a single server without coordination:
cluster-mode migrations fail there. Decide before first start;
switching later needs a data migration.

- default: `true`

### spec.clickhouse.cluster.name

`string | valueFrom`

The cluster's name (`CLICKHOUSE_CLUSTER_NAME`). Defaults compose a
KubernetesClickHouse resource's logical cluster name; empty =
"default". Langfuse renders its migrations for any name.

- references: KubernetesClickHouse (`status.outputs.cluster_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.cluster_name}} -- a bare string does not parse

### spec.clickhouse.readOnlyUrl

`string`

A read-only endpoint for UI and public-API reads
(`CLICKHOUSE_READ_ONLY_URL`, e.g. "https://reader.example.com:8443"):
a separate compute group on ClickHouse Cloud or BYOC. Reuses the user,
password and database. No benefit on a single server.

- rule: read_only_url must be an http(s) URL with host and port, such as https://reader.example.com:8443.

### spec.clickhouse.autoMigrate

`bool` · optional (explicit presence)

Run ClickHouse migrations automatically when the web server starts.

- default: `true`

### spec.clickhouse.deleteMode

`enum`

How Langfuse issues deletes (`CLICKHOUSE_LIGHTWEIGHT_DELETE_MODE`).
Unspecified = Langfuse's default, ALTER TABLE ... DELETE mutations.
The lightweight modes need ClickHouse 25.7+ and cut background merge
work on deletion-heavy workloads (retention, project deletes).

Allowed values (use exactly as shown):

- `unspecified` -- Langfuse's default: ALTER TABLE ... DELETE mutations.
- `alter_update` -- ALTER TABLE ... DELETE mutations, stated explicitly.
- `lightweight_update` -- Lightweight deletes where ClickHouse supports them.
- `lightweight_update_force` -- Lightweight deletes everywhere, even where ClickHouse would fall back.

### spec.clickhouse.useLightweightUpdate

`bool`

Route tracing-table updates through native UPDATE statements instead
of ALTER TABLE ... UPDATE mutations (`CLICKHOUSE_USE_LIGHTWEIGHT_UPDATE`,
ClickHouse 25.7+).

### spec.clickhouse.deletionTimeoutMs

`int32` · optional (explicit presence)

Client-side timeout for delete operations in milliseconds
(`LANGFUSE_CLICKHOUSE_DELETION_TIMEOUT_MS`). Empty = Langfuse's
default (600000, ten minutes). Raise it when retention jobs or
project deletes time out.

- rule: {"int32":{"gt":0}}

### spec.redis

`KubernetesLangfuseRedis` · required

The Redis or Valkey store behind the ingestion queues and caches.

- rule: {"required":true}
- rule: Sentinel TLS applies on top of Redis TLS: set redis.tls.enabled, or turn sentinel.tls_enabled off.

### spec.redis.standalone

`KubernetesLangfuseRedisStandalone`

One primary (with or without replicas behind it) - a
KubernetesValkey, ElastiCache without cluster mode, Memorystore.

### spec.redis.standalone.host

`string | valueFrom` · required

Host. Defaults compose a KubernetesValkey resource's primary Service;
any reachable Redis-protocol host works as a literal.

- references: KubernetesValkey (`status.outputs.service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.service}} -- a bare string does not parse

### spec.redis.standalone.port

`int32` · optional (explicit presence)

Port.

- default: `6379`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.redis.standalone.database

`int32`

Logical database number.

- rule: {"int32":{"lte":15,"gte":0}}

### spec.redis.cluster

`KubernetesLangfuseRedisCluster`

Redis Cluster - ElastiCache with cluster mode, Azure Managed Redis
with the OSS clustering policy.

### spec.redis.cluster.nodes

`[]string` · required

Seed nodes as "host:port" (e.g. "redis-0.redis:6379"); the client
discovers the rest of the cluster from them.

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^[^:\\s]+:[0-9]{1,5}$"}}}}

### spec.redis.cluster.slotsRefreshTimeoutMs

`int32` · optional (explicit presence)

Timeout in milliseconds for refreshing the cluster's slot map
(`REDIS_CLUSTER_SLOTS_REFRESH_TIMEOUT`). Empty = Langfuse's default
(5000). Raise it on timeouts during topology changes.

- rule: {"int32":{"gt":0}}

### spec.redis.sentinel

`KubernetesLangfuseRedisSentinel`

Redis Sentinel - a primary discovered through Sentinel nodes.

### spec.redis.sentinel.masterName

`string` · required

Name of the monitored primary, exactly as the Sentinels know it.

- rule: {"required":true}

### spec.redis.sentinel.nodes

`[]string` · required

Sentinel nodes as "host:port" (e.g. "sentinel-0.sentinel:26379").

- rule: {"repeated":{"minItems":"1","items":{"string":{"pattern":"^[^:\\s]+:[0-9]{1,5}$"}}}}

### spec.redis.sentinel.username

`string`

Username the Sentinels require, when they authenticate.

### spec.redis.sentinel.passwordSecret

`KubernetesSecretKey`

The Secret holding the Sentinels' password, in Langfuse's namespace.
Unset = auth-less Sentinels.

### spec.redis.sentinel.passwordSecret.name

`string`

The name of the Kubernetes Secret.

### spec.redis.sentinel.passwordSecret.key

`string`

The key within the Kubernetes Secret.

### spec.redis.sentinel.tlsEnabled

`bool`

Also use TLS to the Sentinels, not just to the primary and replicas.
Requires `tls.enabled`.

### spec.redis.username

`string | valueFrom`

ACL username, when the store uses named users. Defaults compose a
KubernetesValkey resource's ACL user. Empty = password-only AUTH.

- references: KubernetesValkey (`status.outputs.username`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.username}} -- a bare string does not parse

### spec.redis.passwordSecret

`KubernetesLangfuseRedisPasswordSecret`

The Secret holding the password, in Langfuse's namespace. Unset = an
auth-less store.

### spec.redis.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesValkey resource's
auth Secret.

- references: KubernetesValkey (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.redis.passwordSecret.secretKey

`string | valueFrom` · required

Key within the Secret. Defaults compose a KubernetesValkey resource's
password key (its auth Secret stores each user's password under the
user's name).

- references: KubernetesValkey (`status.outputs.password_secret.key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.password_secret.key}} -- a bare string does not parse

### spec.redis.tls

`KubernetesLangfuseRedisTls`

TLS to Redis (and, for Sentinel, optionally to the Sentinels).

- rule: Mutual TLS needs both client_certificate and client_key, or neither.

### spec.redis.tls.enabled

`bool`

Connect over TLS (`rediss://`).

### spec.redis.tls.ca

`KubernetesLangfuseCertificateFile`

The CA that signed the server certificate, when it is not a public
CA. Mounted into both workloads.

### spec.redis.tls.ca.secretName

`string`

Name of a Secret holding the bundle.

### spec.redis.tls.ca.configMapName

`string`

Name of a ConfigMap holding the bundle.

### spec.redis.tls.ca.key

`string` · optional (explicit presence)

Key holding the PEM bundle.

- default: `ca.crt`

### spec.redis.tls.clientCertificate

`KubernetesLangfuseSecretFile`

Client certificate for mutual TLS. Empty key = "tls.crt".

### spec.redis.tls.clientCertificate.secretName

`string | valueFrom` · required

Name of the Secret. A KubernetesCertificate's Secret composes
naturally.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.redis.tls.clientCertificate.key

`string`

Key within the Secret. Empty = the conventional key for the file
(named on the field that holds this message).

### spec.redis.tls.clientKey

`KubernetesLangfuseSecretFile`

Client private key for mutual TLS. Empty key = "tls.key".

### spec.redis.tls.clientKey.secretName

`string | valueFrom` · required

Name of the Secret. A KubernetesCertificate's Secret composes
naturally.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.redis.tls.clientKey.key

`string`

Key within the Secret. Empty = the conventional key for the file
(named on the field that holds this message).

### spec.redis.tls.clientKeyPassphrase

`string` · sensitive

Passphrase of an encrypted client private key.

### spec.redis.tls.serverName

`string`

Server name for SNI and certificate verification, when it differs
from the host (`REDIS_TLS_SERVERNAME`).

### spec.redis.tls.rejectUnauthorized

`bool` · optional (explicit presence)

Reject a server certificate that does not chain to a trusted CA
(`REDIS_TLS_REJECT_UNAUTHORIZED`). Empty = Node.js's default, true.
Turning it off accepts any certificate - never in production.

### spec.redis.tls.checkServerIdentity

`bool` · optional (explicit presence)

Check that the server certificate names the host
(`REDIS_TLS_CHECK_SERVER_IDENTITY`). Empty = Node.js's default, true.
Some managed services present certificates for another name.

### spec.redis.tls.ciphers

`string`

Allowed TLS cipher suites, OpenSSL format (`REDIS_TLS_CIPHERS`).

### spec.redis.tls.secureProtocol

`string`

TLS protocol method, e.g. "TLSv1_2_method" (`REDIS_TLS_SECURE_PROTOCOL`).

### spec.redis.tls.honorCipherOrder

`bool` · optional (explicit presence)

Prefer the client's cipher order (`REDIS_TLS_HONOR_CIPHER_ORDER`).

### spec.redis.keyPrefix

`string`

Prefix for every key Langfuse writes (`REDIS_KEY_PREFIX`), to share
one store between several Langfuse instances. Changing it on a live
instance orphans every queued job.

### spec.redis.socketTimeoutMs

`int32` · optional (explicit presence)

Socket timeout in milliseconds for every Redis connection
(`REDIS_SOCKET_TIMEOUT_MS`). Empty = Langfuse's default (30000). 0
disables it; a positive value must be at least 10000, because queue
workers block on the socket between jobs.

- rule: socket_timeout_ms must be 0 (disabled) or at least 10000: queue workers block on the socket for seconds between jobs.

### spec.blobStorage

`KubernetesLangfuseBlobStorage` · required

Object storage for ingested events, media attachments and batch
exports.

- rule: {"required":true}

### spec.blobStorage.target

`KubernetesLangfuseBlobTarget` · required

Where events are stored - and media and exports, unless they declare
a target of their own.

- rule: {"required":true}

### spec.blobStorage.target.s3Compatible

`KubernetesLangfuseS3CompatibleTarget`

Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
Storage, Backblaze B2.

### spec.blobStorage.target.s3Compatible.endpoint

`string | valueFrom` · required

The S3 API endpoint (e.g.
"http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
compose a KubernetesSeaweedFs resource's S3 endpoint.

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.target.s3Compatible.bucket

`string` · required

Bucket name. It must exist before Langfuse starts (a
KubernetesSeaweedFs creates the buckets its spec declares).

- rule: {"required":true}

### spec.blobStorage.target.s3Compatible.region

`string` · optional (explicit presence)

Region the store expects in request signatures.

- default: `auto`

### spec.blobStorage.target.s3Compatible.forcePathStyle

`bool` · optional (explicit presence)

Address buckets by path (`endpoint/bucket/key`) rather than by
virtual host - required by SeaweedFS and MinIO.

- default: `true`

### spec.blobStorage.target.s3Compatible.credentialsSecret

`KubernetesLangfuseS3CredentialsSecret`

An existing Secret in Langfuse's namespace holding the key pair.
Defaults compose a KubernetesSeaweedFs resource's S3 credentials
Secret.

### spec.blobStorage.target.s3Compatible.credentialsSecret.secretName

`string | valueFrom` · required

Name of the Secret.

- references: KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_credentials_secret_name}} -- a bare string does not parse

### spec.blobStorage.target.s3Compatible.credentialsSecret.accessKeyIdKey

`string` · optional (explicit presence)

Key holding the access key id.

- default: `admin_access_key_id`

### spec.blobStorage.target.s3Compatible.credentialsSecret.secretAccessKeyKey

`string` · optional (explicit presence)

Key holding the secret access key.

- default: `admin_secret_access_key`

### spec.blobStorage.target.s3Compatible.accessKey

`KubernetesLangfuseS3AccessKey`

The key pair itself, written to the instance Secret.

### spec.blobStorage.target.s3Compatible.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.target.s3Compatible.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.target.s3

`KubernetesLangfuseS3Target`

AWS S3.

### spec.blobStorage.target.s3.bucket

`string | valueFrom` · required

Bucket name. Defaults compose an AwsS3Bucket resource.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.blobStorage.target.s3.region

`string | valueFrom` · required

The bucket's region. Defaults compose an AwsS3Bucket resource's
region.

- references: AwsS3Bucket (`status.outputs.region`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.region}} -- a bare string does not parse

### spec.blobStorage.target.s3.endpoint

`string`

Endpoint override - a FIPS or VPC interface endpoint. Empty = the
regional S3 endpoint.

### spec.blobStorage.target.s3.accessKey

`KubernetesLangfuseS3AccessKey`

A static key pair. Empty = the AWS credential chain: an IAM role
through `service_account.workload_identity` (EKS IRSA) - the
recommended keyless posture - or the node's instance profile.

### spec.blobStorage.target.s3.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.target.s3.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.target.r2

`KubernetesLangfuseR2Target`

Cloudflare R2.

### spec.blobStorage.target.r2.endpoint

`string | valueFrom` · required

The bucket's S3 API endpoint (it encodes the account and the
jurisdiction). Defaults compose a CloudflareR2Bucket resource's
endpoint.

- references: CloudflareR2Bucket (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.target.r2.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.target.r2.credentials

`KubernetesLangfuseR2Credentials` · required

The key pair of an R2 API token scoped to the bucket (Object Read and
Write).

- rule: {"required":true}

### spec.blobStorage.target.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to a
CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
reference for a token minted in the dashboard.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.blobStorage.target.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value.
Reference-only.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.blobStorage.target.gcs

`KubernetesLangfuseGcsTarget`

Google Cloud Storage, through Langfuse's native GCS driver (more
reliable than GCS's S3 interoperability mode, which mishandles
deletes).

### spec.blobStorage.target.gcs.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.target.gcs.credentialsJson

`string` · sensitive

A service account JSON key. Empty = application default credentials:
a Google service account through `service_account.workload_identity`
(GKE Workload Identity) - the recommended keyless posture. Every GCS
target of one Langfuse shares one credential.

### spec.blobStorage.target.azureBlob

`KubernetesLangfuseAzureBlobTarget`

Azure Blob Storage, through Langfuse's native Azure driver.

### spec.blobStorage.target.azureBlob.storageAccount

`string | valueFrom` · required

Storage account name. Defaults compose an AzureStorageAccount
resource.

containment_exempt: Langfuse reads and writes the account - access,
never placement.

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.blobStorage.target.azureBlob.container

`string | valueFrom` · required

Container name. Defaults compose an AzureStorageContainer resource.

- references: AzureStorageContainer (`status.outputs.container_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageContainer, name: <that resource's name>, fieldPath: status.outputs.container_name}} -- a bare string does not parse

### spec.blobStorage.target.azureBlob.accountKey

`string` · required · sensitive

The storage account's access key.

- rule: {"required":true}

### spec.blobStorage.target.azureBlob.endpoint

`string | valueFrom`

Blob endpoint. Defaults compose an AzureStorageAccount resource's
primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.

containment_exempt: an address Langfuse calls - access, never
placement.

- references: AzureStorageAccount (`status.outputs.primary_blob_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_blob_endpoint}} -- a bare string does not parse

### spec.blobStorage.eventUpload

`KubernetesLangfuseEventUpload`

Event upload settings.

### spec.blobStorage.eventUpload.prefix

`string` · optional (explicit presence)

Key prefix for events (must end with "/"). A separate prefix lets a
lifecycle rule expire event files (about 30 days is safe once they
are in ClickHouse) without touching media.

- default: `events/`
- rule: A blob storage prefix must be empty or end with "/", such as events/.

### spec.blobStorage.eventUpload.sse

`KubernetesLangfuseServerSideEncryption`

Server-side encryption (S3 targets).

- rule: kms_key_id applies only to the aws_kms algorithm.

### spec.blobStorage.eventUpload.sse.algorithm

`enum`

Encryption algorithm.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `unspecified` -- Unset.
- `aes256` -- S3-managed keys (AES256).
- `aws_kms` -- AWS KMS keys (aws:kms).

### spec.blobStorage.eventUpload.sse.kmsKeyId

`string | valueFrom`

KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
empty = the bucket's default AWS-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.blobStorage.mediaUpload

`KubernetesLangfuseMediaUpload`

Media attachments (images, audio, documents traced with LLM calls).

### spec.blobStorage.mediaUpload.target

`KubernetesLangfuseBlobTarget`

A target of its own (a publicly reachable bucket, for example). Unset
= the blob storage target.

### spec.blobStorage.mediaUpload.target.s3Compatible

`KubernetesLangfuseS3CompatibleTarget`

Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
Storage, Backblaze B2.

### spec.blobStorage.mediaUpload.target.s3Compatible.endpoint

`string | valueFrom` · required

The S3 API endpoint (e.g.
"http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
compose a KubernetesSeaweedFs resource's S3 endpoint.

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.s3Compatible.bucket

`string` · required

Bucket name. It must exist before Langfuse starts (a
KubernetesSeaweedFs creates the buckets its spec declares).

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.s3Compatible.region

`string` · optional (explicit presence)

Region the store expects in request signatures.

- default: `auto`

### spec.blobStorage.mediaUpload.target.s3Compatible.forcePathStyle

`bool` · optional (explicit presence)

Address buckets by path (`endpoint/bucket/key`) rather than by
virtual host - required by SeaweedFS and MinIO.

- default: `true`

### spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret

`KubernetesLangfuseS3CredentialsSecret`

An existing Secret in Langfuse's namespace holding the key pair.
Defaults compose a KubernetesSeaweedFs resource's S3 credentials
Secret.

### spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.secretName

`string | valueFrom` · required

Name of the Secret.

- references: KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_credentials_secret_name}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.accessKeyIdKey

`string` · optional (explicit presence)

Key holding the access key id.

- default: `admin_access_key_id`

### spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.secretAccessKeyKey

`string` · optional (explicit presence)

Key holding the secret access key.

- default: `admin_secret_access_key`

### spec.blobStorage.mediaUpload.target.s3Compatible.accessKey

`KubernetesLangfuseS3AccessKey`

The key pair itself, written to the instance Secret.

### spec.blobStorage.mediaUpload.target.s3Compatible.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.s3Compatible.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.s3

`KubernetesLangfuseS3Target`

AWS S3.

### spec.blobStorage.mediaUpload.target.s3.bucket

`string | valueFrom` · required

Bucket name. Defaults compose an AwsS3Bucket resource.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.s3.region

`string | valueFrom` · required

The bucket's region. Defaults compose an AwsS3Bucket resource's
region.

- references: AwsS3Bucket (`status.outputs.region`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.region}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.s3.endpoint

`string`

Endpoint override - a FIPS or VPC interface endpoint. Empty = the
regional S3 endpoint.

### spec.blobStorage.mediaUpload.target.s3.accessKey

`KubernetesLangfuseS3AccessKey`

A static key pair. Empty = the AWS credential chain: an IAM role
through `service_account.workload_identity` (EKS IRSA) - the
recommended keyless posture - or the node's instance profile.

### spec.blobStorage.mediaUpload.target.s3.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.s3.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.r2

`KubernetesLangfuseR2Target`

Cloudflare R2.

### spec.blobStorage.mediaUpload.target.r2.endpoint

`string | valueFrom` · required

The bucket's S3 API endpoint (it encodes the account and the
jurisdiction). Defaults compose a CloudflareR2Bucket resource's
endpoint.

- references: CloudflareR2Bucket (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.r2.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.r2.credentials

`KubernetesLangfuseR2Credentials` · required

The key pair of an R2 API token scoped to the bucket (Object Read and
Write).

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to a
CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
reference for a token minted in the dashboard.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value.
Reference-only.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.gcs

`KubernetesLangfuseGcsTarget`

Google Cloud Storage, through Langfuse's native GCS driver (more
reliable than GCS's S3 interoperability mode, which mishandles
deletes).

### spec.blobStorage.mediaUpload.target.gcs.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.gcs.credentialsJson

`string` · sensitive

A service account JSON key. Empty = application default credentials:
a Google service account through `service_account.workload_identity`
(GKE Workload Identity) - the recommended keyless posture. Every GCS
target of one Langfuse shares one credential.

### spec.blobStorage.mediaUpload.target.azureBlob

`KubernetesLangfuseAzureBlobTarget`

Azure Blob Storage, through Langfuse's native Azure driver.

### spec.blobStorage.mediaUpload.target.azureBlob.storageAccount

`string | valueFrom` · required

Storage account name. Defaults compose an AzureStorageAccount
resource.

containment_exempt: Langfuse reads and writes the account - access,
never placement.

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.azureBlob.container

`string | valueFrom` · required

Container name. Defaults compose an AzureStorageContainer resource.

- references: AzureStorageContainer (`status.outputs.container_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageContainer, name: <that resource's name>, fieldPath: status.outputs.container_name}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.target.azureBlob.accountKey

`string` · required · sensitive

The storage account's access key.

- rule: {"required":true}

### spec.blobStorage.mediaUpload.target.azureBlob.endpoint

`string | valueFrom`

Blob endpoint. Defaults compose an AzureStorageAccount resource's
primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.

containment_exempt: an address Langfuse calls - access, never
placement.

- references: AzureStorageAccount (`status.outputs.primary_blob_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_blob_endpoint}} -- a bare string does not parse

### spec.blobStorage.mediaUpload.prefix

`string` · optional (explicit presence)

Key prefix for media (must end with "/"). Put no expiry lifecycle rule
on it: traces keep pointing at their media.

- default: `media/`
- rule: A blob storage prefix must be empty or end with "/", such as media/.

### spec.blobStorage.mediaUpload.publicEndpoint

`string`

The endpoint browsers and SDKs reach the store at, when the target's
endpoint is cluster-internal (e.g. "https://s3.langfuse.example.com"
for an in-cluster SeaweedFS behind a route). Langfuse signs URLs for
this endpoint and reads media itself through the target's endpoint.
Empty = the target's endpoint serves both.

- rule: public_endpoint must be an http(s) URL, such as https://s3.langfuse.example.com.

### spec.blobStorage.mediaUpload.maxContentLengthBytes

`int64` · optional (explicit presence)

Largest file accepted, in bytes (`LANGFUSE_S3_MEDIA_MAX_CONTENT_LENGTH`).

- default: `1000000000`
- rule: {"int64":{"gt":"0"}}

### spec.blobStorage.mediaUpload.downloadUrlExpirySeconds

`int32` · optional (explicit presence)

Lifetime of presigned download URLs in seconds.

- default: `3600`
- rule: {"int32":{"gt":0}}

### spec.blobStorage.mediaUpload.sse

`KubernetesLangfuseServerSideEncryption`

Server-side encryption (S3 targets).

- rule: kms_key_id applies only to the aws_kms algorithm.

### spec.blobStorage.mediaUpload.sse.algorithm

`enum`

Encryption algorithm.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `unspecified` -- Unset.
- `aes256` -- S3-managed keys (AES256).
- `aws_kms` -- AWS KMS keys (aws:kms).

### spec.blobStorage.mediaUpload.sse.kmsKeyId

`string | valueFrom`

KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
empty = the bucket's default AWS-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.blobStorage.batchExport

`KubernetesLangfuseBatchExport`

Batch exports (CSV and JSON downloads of traces, observations and
scores).

### spec.blobStorage.batchExport.enabled

`bool` · optional (explicit presence)

Offer batch exports.

- default: `true`

### spec.blobStorage.batchExport.target

`KubernetesLangfuseBlobTarget`

A target of its own. Unset = the blob storage target.

### spec.blobStorage.batchExport.target.s3Compatible

`KubernetesLangfuseS3CompatibleTarget`

Any S3-compatible store: a KubernetesSeaweedFs, MinIO, OCI Object
Storage, Backblaze B2.

### spec.blobStorage.batchExport.target.s3Compatible.endpoint

`string | valueFrom` · required

The S3 API endpoint (e.g.
"http://seaweedfs-s3.storage.svc.cluster.local:8333"). Defaults
compose a KubernetesSeaweedFs resource's S3 endpoint.

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.s3Compatible.bucket

`string` · required

Bucket name. It must exist before Langfuse starts (a
KubernetesSeaweedFs creates the buckets its spec declares).

- rule: {"required":true}

### spec.blobStorage.batchExport.target.s3Compatible.region

`string` · optional (explicit presence)

Region the store expects in request signatures.

- default: `auto`

### spec.blobStorage.batchExport.target.s3Compatible.forcePathStyle

`bool` · optional (explicit presence)

Address buckets by path (`endpoint/bucket/key`) rather than by
virtual host - required by SeaweedFS and MinIO.

- default: `true`

### spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret

`KubernetesLangfuseS3CredentialsSecret`

An existing Secret in Langfuse's namespace holding the key pair.
Defaults compose a KubernetesSeaweedFs resource's S3 credentials
Secret.

### spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.secretName

`string | valueFrom` · required

Name of the Secret.

- references: KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_credentials_secret_name}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.accessKeyIdKey

`string` · optional (explicit presence)

Key holding the access key id.

- default: `admin_access_key_id`

### spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.secretAccessKeyKey

`string` · optional (explicit presence)

Key holding the secret access key.

- default: `admin_secret_access_key`

### spec.blobStorage.batchExport.target.s3Compatible.accessKey

`KubernetesLangfuseS3AccessKey`

The key pair itself, written to the instance Secret.

### spec.blobStorage.batchExport.target.s3Compatible.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.batchExport.target.s3Compatible.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.batchExport.target.s3

`KubernetesLangfuseS3Target`

AWS S3.

### spec.blobStorage.batchExport.target.s3.bucket

`string | valueFrom` · required

Bucket name. Defaults compose an AwsS3Bucket resource.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.s3.region

`string | valueFrom` · required

The bucket's region. Defaults compose an AwsS3Bucket resource's
region.

- references: AwsS3Bucket (`status.outputs.region`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.region}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.s3.endpoint

`string`

Endpoint override - a FIPS or VPC interface endpoint. Empty = the
regional S3 endpoint.

### spec.blobStorage.batchExport.target.s3.accessKey

`KubernetesLangfuseS3AccessKey`

A static key pair. Empty = the AWS credential chain: an IAM role
through `service_account.workload_identity` (EKS IRSA) - the
recommended keyless posture - or the node's instance profile.

### spec.blobStorage.batchExport.target.s3.accessKey.accessKeyId

`string` · required

Access key id.

- rule: {"required":true}

### spec.blobStorage.batchExport.target.s3.accessKey.secretAccessKey

`string` · required · sensitive

Secret access key.

- rule: {"required":true}

### spec.blobStorage.batchExport.target.r2

`KubernetesLangfuseR2Target`

Cloudflare R2.

### spec.blobStorage.batchExport.target.r2.endpoint

`string | valueFrom` · required

The bucket's S3 API endpoint (it encodes the account and the
jurisdiction). Defaults compose a CloudflareR2Bucket resource's
endpoint.

- references: CloudflareR2Bucket (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.r2.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a CloudflareR2Bucket resource.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.r2.credentials

`KubernetesLangfuseR2Credentials` · required

The key pair of an R2 API token scoped to the bucket (Object Read and
Write).

- rule: {"required":true}

### spec.blobStorage.batchExport.target.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to a
CloudflareAccountApiToken's `r2_access_key_id` output, or a `$secret/`
reference for a token minted in the dashboard.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value.
Reference-only.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.gcs

`KubernetesLangfuseGcsTarget`

Google Cloud Storage, through Langfuse's native GCS driver (more
reliable than GCS's S3 interoperability mode, which mishandles
deletes).

### spec.blobStorage.batchExport.target.gcs.bucket

`string | valueFrom` · required

Bucket name. Defaults compose a GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.gcs.credentialsJson

`string` · sensitive

A service account JSON key. Empty = application default credentials:
a Google service account through `service_account.workload_identity`
(GKE Workload Identity) - the recommended keyless posture. Every GCS
target of one Langfuse shares one credential.

### spec.blobStorage.batchExport.target.azureBlob

`KubernetesLangfuseAzureBlobTarget`

Azure Blob Storage, through Langfuse's native Azure driver.

### spec.blobStorage.batchExport.target.azureBlob.storageAccount

`string | valueFrom` · required

Storage account name. Defaults compose an AzureStorageAccount
resource.

containment_exempt: Langfuse reads and writes the account - access,
never placement.

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.azureBlob.container

`string | valueFrom` · required

Container name. Defaults compose an AzureStorageContainer resource.

- references: AzureStorageContainer (`status.outputs.container_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageContainer, name: <that resource's name>, fieldPath: status.outputs.container_name}} -- a bare string does not parse

### spec.blobStorage.batchExport.target.azureBlob.accountKey

`string` · required · sensitive

The storage account's access key.

- rule: {"required":true}

### spec.blobStorage.batchExport.target.azureBlob.endpoint

`string | valueFrom`

Blob endpoint. Defaults compose an AzureStorageAccount resource's
primary blob endpoint; empty = https://<storage_account>.blob.core.windows.net.

containment_exempt: an address Langfuse calls - access, never
placement.

- references: AzureStorageAccount (`status.outputs.primary_blob_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_blob_endpoint}} -- a bare string does not parse

### spec.blobStorage.batchExport.prefix

`string` · optional (explicit presence)

Key prefix for exports (must end with "/"). Exports are disposable:
an expiry lifecycle rule here is safe.

- default: `exports/`
- rule: A blob storage prefix must be empty or end with "/", such as exports/.

### spec.blobStorage.batchExport.publicEndpoint

`string`

The endpoint download links point at, when the target's endpoint is
cluster-internal (`LANGFUSE_S3_BATCH_EXPORT_EXTERNAL_ENDPOINT`). Empty
= the target's endpoint.

- rule: public_endpoint must be an http(s) URL, such as https://s3.langfuse.example.com.

### spec.blobStorage.batchExport.pageSize

`int32` · optional (explicit presence)

Rows streamed per page while exporting (`BATCH_EXPORT_PAGE_SIZE`).
Empty = Langfuse's default (500). Lower it if exports exhaust worker
memory.

- rule: {"int32":{"gt":0}}

### spec.blobStorage.batchExport.rowLimit

`int32` · optional (explicit presence)

Most rows one export may contain (`BATCH_EXPORT_ROW_LIMIT`). Empty =
Langfuse's default (1500000).

- rule: {"int32":{"gt":0}}

### spec.blobStorage.batchExport.downloadLinkExpirationHours

`int32` · optional (explicit presence)

Hours a download link stays valid
(`BATCH_EXPORT_DOWNLOAD_LINK_EXPIRATION_HOURS`). Empty = Langfuse's
default (24).

- rule: {"int32":{"gt":0}}

### spec.blobStorage.batchExport.sse

`KubernetesLangfuseServerSideEncryption`

Server-side encryption (S3 targets).

- rule: kms_key_id applies only to the aws_kms algorithm.

### spec.blobStorage.batchExport.sse.algorithm

`enum`

Encryption algorithm.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `unspecified` -- Unset.
- `aes256` -- S3-managed keys (AES256).
- `aws_kms` -- AWS KMS keys (aws:kms).

### spec.blobStorage.batchExport.sse.kmsKeyId

`string | valueFrom`

KMS key for `aws_kms`. Defaults compose an AwsKmsKey resource's ARN;
empty = the bucket's default AWS-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.blobStorage.concurrentReads

`int32` · optional (explicit presence)

Maximum concurrent reads from object storage
(`LANGFUSE_S3_CONCURRENT_READS`). Raise it on "socket usage at
capacity" warnings.

- default: `50`
- rule: {"int32":{"gt":0}}

### spec.blobStorage.concurrentWrites

`int32` · optional (explicit presence)

Maximum concurrent writes to object storage
(`LANGFUSE_S3_CONCURRENT_WRITES`).

- default: `50`
- rule: {"int32":{"gt":0}}

### spec.blobStorage.deleteObjectsChecksumAlgorithm

`enum`

Checksum algorithm for S3 DeleteObjects requests
(`LANGFUSE_S3_DELETE_OBJECTS_CHECKSUM_ALGORITHM`). Unspecified = the
SDK's default (CRC32). Some S3-compatible stores reject the default
and need MD5.

Allowed values (use exactly as shown):

- `unspecified` -- The SDK's default (CRC32).
- `md5` -- MD5 - the legacy Content-MD5 header.
- `crc32` -- CRC32.
- `crc32c` -- CRC32C.
- `crc64nvme` -- CRC64NVME.
- `sha1` -- SHA-1.
- `sha256` -- SHA-256.

### spec.secrets

`KubernetesLangfuseInstanceSecrets`

The instance's own secrets: the API-key salt, the encryption key and
the session secret. Generated when not declared.

### spec.secrets.salt

`string` · sensitive

Salt for hashing API keys (`SALT`), at least 256 bits of entropy
(`openssl rand -base64 32`). Empty = generated.

### spec.secrets.encryptionKey

`string` · sensitive

Key for encrypting stored credentials (`ENCRYPTION_KEY`): exactly 64
hex characters, 256 bits (`openssl rand -hex 32`). Langfuse refuses to
start on any other shape. Empty = generated.

### spec.secrets.nextauthSecret

`string` · sensitive

Secret for signing session cookies (`NEXTAUTH_SECRET`). Empty =
generated. Changing it signs everyone out, nothing more.

### spec.init

`KubernetesLangfuseInit`

The organization, project, project key pair and admin user Langfuse
creates at first start (headless initialization). Seeded by default.

### spec.init.enabled

`bool` · optional (explicit presence)

Seed the organization and project at first start. Turn off to start
with an empty instance where the first sign-up creates everything.

- default: `true`

### spec.init.organization

`KubernetesLangfuseInitOrganization`

The seeded organization.

### spec.init.organization.id

`string`

Organization ID. Empty = metadata.name. Fixed at first start.

- rule: Organization IDs use letters, digits, hyphens and underscores, up to 64 characters.

### spec.init.organization.name

`string` · optional (explicit presence)

Organization display name.

- default: `Default Organization`

### spec.init.project

`KubernetesLangfuseInitProject`

The seeded project and its API key pair.

- rule: Declare both public_key and secret_key, or neither (both are then generated).

### spec.init.project.id

`string`

Project ID. Empty = "<metadata.name>-project". Fixed at first start.

- rule: Project IDs use letters, digits, hyphens and underscores, up to 64 characters.

### spec.init.project.name

`string` · optional (explicit presence)

Project display name.

- default: `Default Project`

### spec.init.project.retentionDays

`int32` · optional (explicit presence)

Data retention in days (3 or more). Applied only with an Enterprise
Edition license that includes data retention; otherwise data is kept
indefinitely. Empty = keep indefinitely.

- rule: {"int32":{"gte":3}}

### spec.init.project.publicKey

`string`

Public key of the project's API key pair ("pk-lf-..."). Empty =
generated ("pk-lf-" + 32 random letters and digits).

- rule: public_key must start with pk-lf-.

### spec.init.project.secretKey

`string` · sensitive

Secret key of the project's API key pair. It must start with
"sk-lf-", the prefix Langfuse's SDKs expect. Empty = generated
("sk-lf-" + 32 random letters and digits).

### spec.init.admin

`KubernetesLangfuseInitAdmin`

The seeded admin user, owner of the organization. Unset = no user is
seeded (sign up, or sign in with single sign-on, instead).

### spec.init.admin.email

`string` · required

Email address, the sign-in name.

- rule: {"required":true,"string":{"email":true}}

### spec.init.admin.name

`string` · optional (explicit presence)

Display name.

- default: `Admin`

### spec.init.admin.password

`string` · sensitive

Password. Empty = generated (32 letters and digits) and readable from
the `admin_password_secret` output.

### spec.auth

`KubernetesLangfuseAuth`

Sign-in: email and password, sign-up, sessions, the access new users
receive, and single sign-on providers.

### spec.auth.disablePasswordSignIn

`bool`

Disable email and password sign-in (`AUTH_DISABLE_USERNAME_PASSWORD`)
- single sign-on only. Keep at least one SSO provider configured.

### spec.auth.disableSignUp

`bool`

Disable sign-up for new users, invited ones included
(`AUTH_DISABLE_SIGNUP`). Recommended for any exposed instance once
the admin exists.

### spec.auth.emailVerificationRequired

`bool`

Require new email-and-password users to verify their address
(`AUTH_EMAIL_VERIFICATION_REQUIRED`). Needs `smtp`.

### spec.auth.sessionMaxAgeMinutes

`int32` · optional (explicit presence)

Session lifetime in minutes (`AUTH_SESSION_MAX_AGE`). Empty =
Langfuse's default (20160, fourteen days). Must exceed 5: sessions
refresh every five minutes.

- rule: {"int32":{"gt":5}}

### spec.auth.ssoEnforcedDomains

`[]string`

Email domains that may only sign in through single sign-on
(`AUTH_DOMAINS_WITH_SSO_ENFORCEMENT`), e.g. "example.com".

### spec.auth.ignoreAccountFields

`[]string`

Account fields to ignore from the identity provider when creating an
account (`AUTH_IGNORE_ACCOUNT_FIELDS`) - a fix for providers that
return fields Langfuse cannot store.

### spec.auth.httpsProxy

`string`

Proxy for requests to identity providers (`AUTH_HTTPS_PROXY`).

### spec.auth.httpProxy

`string`

Proxy for plain-HTTP requests to identity providers (`AUTH_HTTP_PROXY`).

### spec.auth.newUserAccess

`KubernetesLangfuseNewUserAccess`

Organization and project memberships every new user receives.

### spec.auth.newUserAccess.organizationId

`string`

Organization every new user joins (`LANGFUSE_DEFAULT_ORG_ID`) - the
seeded organization's ID, for example.

### spec.auth.newUserAccess.organizationRole

`enum`

Their role there (`LANGFUSE_DEFAULT_ORG_ROLE`). Unspecified =
Langfuse's default, VIEWER.

Allowed values (use exactly as shown):

- `unspecified` -- Langfuse's default for the context.
- `OWNER` -- Full control, including billing and deletion.
- `ADMIN` -- Manages members and settings.
- `MEMBER` -- Works with the data.
- `VIEWER` -- Read-only.
- `NONE` -- Organization membership without access to any project.

### spec.auth.newUserAccess.projectIds

`[]string`

Projects every new user joins (`LANGFUSE_DEFAULT_PROJECT_ID`).

### spec.auth.newUserAccess.projectRole

`enum`

Their role there (`LANGFUSE_DEFAULT_PROJECT_ROLE`). Unspecified =
Langfuse's default, VIEWER. NONE is not a project role.

- rule: {"enum":{"definedOnly":true,"notIn":[5]}}

Allowed values (use exactly as shown):

- `unspecified` -- Langfuse's default for the context.
- `OWNER` -- Full control, including billing and deletion.
- `ADMIN` -- Manages members and settings.
- `MEMBER` -- Works with the data.
- `VIEWER` -- Read-only.
- `NONE` -- Organization membership without access to any project.

### spec.auth.sso

`KubernetesLangfuseSsoProviders`

Single sign-on providers. Each one configured becomes a sign-in
button; its redirect URI is `<external_url>/api/auth/callback/<id>`.

### spec.auth.sso.google

`KubernetesLangfuseGoogleSso`

Google.

### spec.auth.sso.google.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.google.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.google.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.google.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.google.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.google.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.google.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.google.allowedDomains

`[]string`

Google Workspace domains allowed to sign in (`AUTH_GOOGLE_ALLOWED_DOMAINS`).
Empty = any Google account.

### spec.auth.sso.github

`KubernetesLangfuseGitHubSso`

GitHub.

- rule: GitHub issues no ID token: leave id_token_signed_response_alg unset.

### spec.auth.sso.github.client

`KubernetesLangfuseOAuthClient` · required

The OAuth app.

- rule: {"required":true}

### spec.auth.sso.github.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.github.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.github.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.github.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.github.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.github.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.githubEnterprise

`KubernetesLangfuseGitHubEnterpriseSso`

GitHub Enterprise Server.

- rule: GitHub Enterprise issues no ID token: leave id_token_signed_response_alg unset.
- rule: GitHub Enterprise needs base_url, the server's address (e.g. https://github.example.com).

### spec.auth.sso.githubEnterprise.client

`KubernetesLangfuseOAuthClient` · required

The OAuth app.

- rule: {"required":true}

### spec.auth.sso.githubEnterprise.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.githubEnterprise.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.githubEnterprise.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.githubEnterprise.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.githubEnterprise.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.githubEnterprise.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.githubEnterprise.baseUrl

`string`

Base URL of the server (`AUTH_GITHUB_ENTERPRISE_BASE_URL`), e.g.
"https://github.example.com".

- rule: {"string":{"uri":true}}

### spec.auth.sso.gitlab

`KubernetesLangfuseGitLabSso`

GitLab (gitlab.com or self-managed).

### spec.auth.sso.gitlab.client

`KubernetesLangfuseOAuthClient` · required

The OAuth application.

- rule: {"required":true}

### spec.auth.sso.gitlab.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.gitlab.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.gitlab.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.gitlab.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.gitlab.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.gitlab.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.gitlab.url

`string`

Base URL of a self-managed GitLab (`AUTH_GITLAB_URL`). Empty =
https://gitlab.com.

### spec.auth.sso.gitlab.issuer

`string`

OIDC issuer, when it differs from `url` (`AUTH_GITLAB_ISSUER`).

### spec.auth.sso.azureAd

`KubernetesLangfuseAzureAdSso`

Microsoft Entra ID (Azure AD).

### spec.auth.sso.azureAd.client

`KubernetesLangfuseOAuthClient` · required

The app registration.

- rule: {"required":true}

### spec.auth.sso.azureAd.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.azureAd.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.azureAd.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.azureAd.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.azureAd.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.azureAd.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.azureAd.tenantId

`string` · required

Directory (tenant) ID (`AUTH_AZURE_AD_TENANT_ID`).

- rule: {"required":true}

### spec.auth.sso.okta

`KubernetesLangfuseIssuerSso`

Okta.

### spec.auth.sso.okta.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.okta.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.okta.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.okta.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.okta.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.okta.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.okta.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.okta.issuer

`string` · required

Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
"https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".

- rule: {"required":true}

### spec.auth.sso.authentik

`KubernetesLangfuseAuthentikSso`

Authentik.

### spec.auth.sso.authentik.client

`KubernetesLangfuseOAuthClient` · required

The OAuth provider.

- rule: {"required":true}

### spec.auth.sso.authentik.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.authentik.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.authentik.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.authentik.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.authentik.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.authentik.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.authentik.issuer

`string` · required

Issuer, "https://<domain>/application/o/<slug>" (no trailing path
beyond the slug).

- rule: {"required":true,"string":{"pattern":"^https://.+/application/o/[^/]+/?$"}}

### spec.auth.sso.authentik.authorizationUrl

`string`

Authorization URL, when it differs from the issuer's
(`AUTH_AUTHENTIK_AUTHORIZATION_URL`).

### spec.auth.sso.onelogin

`KubernetesLangfuseIssuerSso`

OneLogin.

### spec.auth.sso.onelogin.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.onelogin.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.onelogin.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.onelogin.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.onelogin.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.onelogin.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.onelogin.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.onelogin.issuer

`string` · required

Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
"https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".

- rule: {"required":true}

### spec.auth.sso.auth0

`KubernetesLangfuseIssuerSso`

Auth0.

### spec.auth.sso.auth0.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.auth0.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.auth0.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.auth0.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.auth0.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.auth0.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.auth0.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.auth0.issuer

`string` · required

Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
"https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".

- rule: {"required":true}

### spec.auth.sso.cognito

`KubernetesLangfuseIssuerSso`

Amazon Cognito.

### spec.auth.sso.cognito.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.cognito.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.cognito.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.cognito.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.cognito.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.cognito.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.cognito.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.cognito.issuer

`string` · required

Issuer URL, e.g. "https://example.okta.com" or, for Cognito,
"https://cognito-idp.<region>.amazonaws.com/<user-pool-id>".

- rule: {"required":true}

### spec.auth.sso.keycloak

`KubernetesLangfuseKeycloakSso`

Keycloak.

### spec.auth.sso.keycloak.client

`KubernetesLangfuseOAuthClient` · required

The client.

- rule: {"required":true}

### spec.auth.sso.keycloak.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.keycloak.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.keycloak.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.keycloak.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.keycloak.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.keycloak.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.keycloak.issuer

`string` · required

Realm issuer, e.g. "https://keycloak.example.com/realms/langfuse".

- rule: {"required":true}

### spec.auth.sso.keycloak.name

`string`

Button label (`AUTH_KEYCLOAK_NAME`). Empty = "Keycloak".

### spec.auth.sso.keycloak.scope

`string`

OAuth scopes (`AUTH_KEYCLOAK_SCOPE`). Empty = "openid email profile".

### spec.auth.sso.keycloak.useIdToken

`bool` · optional (explicit presence)

Read the user from the ID token (`AUTH_KEYCLOAK_ID_TOKEN`). Empty =
true; false fetches the userinfo endpoint instead.

### spec.auth.sso.jumpcloud

`KubernetesLangfuseJumpCloudSso`

JumpCloud.

### spec.auth.sso.jumpcloud.client

`KubernetesLangfuseOAuthClient` · required

The SSO application.

- rule: {"required":true}

### spec.auth.sso.jumpcloud.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.jumpcloud.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.jumpcloud.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.jumpcloud.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.jumpcloud.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.jumpcloud.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.jumpcloud.issuer

`string` · required

Issuer URL.

- rule: {"required":true}

### spec.auth.sso.jumpcloud.scope

`string`

OAuth scopes (`AUTH_JUMPCLOUD_SCOPE`).

### spec.auth.sso.custom

`KubernetesLangfuseCustomSso`

Any OpenID Connect provider.

### spec.auth.sso.custom.client

`KubernetesLangfuseOAuthClient` · required

The OAuth client.

- rule: {"required":true}

### spec.auth.sso.custom.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.custom.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.custom.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.custom.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.custom.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.custom.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.custom.issuer

`string` · required

Issuer URL (its /.well-known/openid-configuration is discovered).

- rule: {"required":true}

### spec.auth.sso.custom.name

`string` · required

Button label (`AUTH_CUSTOM_NAME`), e.g. "Company SSO".

- rule: {"required":true}

### spec.auth.sso.custom.scope

`string`

OAuth scopes (`AUTH_CUSTOM_SCOPE`). Empty = "openid email profile".

### spec.auth.sso.custom.useIdToken

`bool` · optional (explicit presence)

Read the user from the ID token (`AUTH_CUSTOM_ID_TOKEN`). Empty =
true.

### spec.auth.sso.custom.fetchUserinfo

`bool`

Also fetch the userinfo endpoint (`AUTH_CUSTOM_FETCH_USERINFO`).

### spec.auth.sso.custom.claims

`KubernetesLangfuseSsoClaims`

Claim names, when the provider's differ from the standard ones.

### spec.auth.sso.custom.claims.email

`string`

Email claim (`LANGFUSE_CUSTOM_SSO_EMAIL_CLAIM`). Empty = "email".

### spec.auth.sso.custom.claims.name

`string`

Name claim (`LANGFUSE_CUSTOM_SSO_NAME_CLAIM`). Empty = "name".

### spec.auth.sso.custom.claims.subject

`string`

Subject claim (`LANGFUSE_CUSTOM_SSO_SUB_CLAIM`). Empty = "sub".

### spec.auth.sso.custom.claims.image

`string`

Picture claim (`LANGFUSE_CUSTOM_SSO_IMAGE_CLAIM`). Empty = "picture".

### spec.auth.sso.workos

`KubernetesLangfuseWorkOsSso`

WorkOS.

- rule: WorkOS takes only client_id, client_secret and allow_account_linking; leave client_auth_method, checks and id_token_signed_response_alg unset.

### spec.auth.sso.workos.client

`KubernetesLangfuseOAuthClient` · required

The WorkOS client.

- rule: {"required":true}

### spec.auth.sso.workos.client.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.workos.client.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.workos.client.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.workos.client.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.workos.client.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.workos.client.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.sso.workos.organizationId

`string`

Route sign-in to this organization's connection
(`AUTH_WORKOS_ORGANIZATION_ID`).

### spec.auth.sso.workos.connectionId

`string`

Route sign-in to this connection (`AUTH_WORKOS_CONNECTION_ID`).

### spec.auth.sso.wordpress

`KubernetesLangfuseOAuthClient`

WordPress.com.

### spec.auth.sso.wordpress.clientId

`string` · required

Client ID.

- rule: {"required":true}

### spec.auth.sso.wordpress.clientSecret

`string` · required · sensitive

Client secret.

- rule: {"required":true}

### spec.auth.sso.wordpress.allowAccountLinking

`bool`

Merge a sign-in with an existing account of the same email address
(`ALLOW_ACCOUNT_LINKING`) - how an instance moves from email and
password to SSO. Safe only when the provider verifies emails.

### spec.auth.sso.wordpress.clientAuthMethod

`enum`

Token endpoint authentication method (`CLIENT_AUTH_METHOD`).
Unspecified = client_secret_basic. Use `none` for PKCE.

Allowed values (use exactly as shown):

- `client_auth_method_unspecified` -- client_secret_basic.
- `client_secret_basic` -- HTTP Basic auth with the client secret.
- `client_secret_post` -- Client secret in the request body.
- `client_secret_jwt` -- A JWT signed with the client secret.
- `private_key_jwt` -- A JWT signed with a private key.
- `tls_client_auth` -- Mutual TLS with a CA-issued certificate.
- `self_signed_tls_client_auth` -- Mutual TLS with a self-signed certificate.
- `none` -- No client authentication (PKCE).

### spec.auth.sso.wordpress.checks

`[]enum`

Authentication checks (`CHECKS`), e.g. pkce and state for Okta.
Empty = the provider's defaults.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `check_unspecified` -- Unset.
- `nonce` -- Nonce check.
- `no_checks` -- No checks.
- `pkce` -- PKCE.
- `state` -- State parameter.

### spec.auth.sso.wordpress.idTokenSignedResponseAlg

`enum`

Algorithm the provider signs ID tokens with
(`ID_TOKEN_SIGNED_RESPONSE_ALG`). Unspecified = RS256.

Allowed values (use exactly as shown):

- `id_token_algorithm_unspecified` -- RS256.
- `RS256` -- RSASSA-PKCS1-v1_5 with SHA-256.
- `RS384` -- RSASSA-PKCS1-v1_5 with SHA-384.
- `RS512` -- RSASSA-PKCS1-v1_5 with SHA-512.
- `ES256` -- ECDSA P-256 with SHA-256.
- `ES384` -- ECDSA P-384 with SHA-384.
- `ES512` -- ECDSA P-521 with SHA-512.
- `PS256` -- RSASSA-PSS with SHA-256.
- `PS384` -- RSASSA-PSS with SHA-384.
- `PS512` -- RSASSA-PSS with SHA-512.
- `HS256` -- HMAC with SHA-256.
- `HS384` -- HMAC with SHA-384.
- `HS512` -- HMAC with SHA-512.

### spec.auth.ssoTimeoutMs

`int32` · optional (explicit presence)

Requests to identity providers time out after this many milliseconds
(`AUTH_SSO_TIMEOUT`). Empty = the auth library's default.

- rule: {"int32":{"gt":0}}

### spec.smtp

`KubernetesLangfuseSmtp`

Transactional email (invitations, password resets, notifications).
Unset = Langfuse sends no email.

### spec.smtp.connectionUrl

`string` · required · sensitive

Connection URL (`SMTP_CONNECTION_URL`):
"smtp://user:password@host:587", "smtps://user:password@host:465",
or "ses://<region>" to send through Amazon SES with the AWS credential
chain (an IAM role through workload identity).

- rule: {"required":true}

### spec.smtp.fromAddress

`string` · required

Sender address (`EMAIL_FROM_ADDRESS`).

- rule: {"required":true,"string":{"email":true}}

### spec.features

`KubernetesLangfuseFeatures`

Product switches: telemetry, experimental features, organization
creation, content security policy.

### spec.features.telemetryEnabled

`bool` · optional (explicit presence)

Send anonymous usage statistics to Langfuse (`TELEMETRY_ENABLED`) -
counts, never trace content.

- default: `true`

### spec.features.experimentalFeaturesEnabled

`bool`

Turn on features still in development
(`LANGFUSE_ENABLE_EXPERIMENTAL_FEATURES`).

### spec.features.allowedOrganizationCreators

`[]string`

Emails allowed to create organizations
(`LANGFUSE_ALLOWED_ORGANIZATION_CREATORS`, Enterprise Edition). Empty
= every user may.

- rule: {"repeated":{"items":{"string":{"email":true}}}}

### spec.features.cspEnforceHttps

`bool`

Restrict the Content-Security-Policy to HTTPS connections
(`LANGFUSE_CSP_ENFORCE_HTTPS`). Turn on once `external_url` is https.

### spec.features.markdownRenderCharacterLimit

`int32` · optional (explicit presence)

Inputs and outputs longer than this many characters render as plain
text instead of markdown (`LANGFUSE_MARKDOWN_RENDER_CHARACTER_LIMIT`).
Empty = Langfuse's default (150000).

- rule: {"int32":{"gt":0}}

### spec.caching

`KubernetesLangfuseCaching`

API-key and prompt caching in Redis.

### spec.caching.apiKeyCacheEnabled

`bool` · optional (explicit presence)

Cache API key lookups (`LANGFUSE_CACHE_API_KEY_ENABLED`).

- default: `true`

### spec.caching.apiKeyCacheTtlSeconds

`int32` · optional (explicit presence)

API key cache lifetime in seconds (`LANGFUSE_CACHE_API_KEY_TTL_SECONDS`)
- also the longest a revoked key keeps working. Empty = Langfuse's
default (60).

- rule: {"int32":{"gt":0}}

### spec.caching.promptCacheEnabled

`bool` · optional (explicit presence)

Cache prompts (`LANGFUSE_CACHE_PROMPT_ENABLED`).

- default: `true`

### spec.caching.promptCacheTtlSeconds

`int32` · optional (explicit presence)

Prompt cache lifetime in seconds (`LANGFUSE_CACHE_PROMPT_TTL_SECONDS`).
Empty = Langfuse's default (3600).

- rule: {"int32":{"gt":0}}

### spec.ingestion

`KubernetesLangfuseIngestion`

Queue sharding and worker concurrency for high ingestion volume.

### spec.ingestion.queueShardCount

`int32` · optional (explicit presence)

Shards of the ingestion queue (`LANGFUSE_INGESTION_QUEUE_SHARD_COUNT`),
about 2-3 per Redis Cluster shard. Never reduce it on a live instance:
jobs on removed shards are stranded. Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.ingestion.queueProcessingConcurrency

`int32` · optional (explicit presence)

Ingestion jobs each worker processes concurrently, per shard
(`LANGFUSE_INGESTION_QUEUE_PROCESSING_CONCURRENCY`). Empty = Langfuse's
default (20). With ten shards, 2 keeps the same total.

- rule: {"int32":{"gte":1}}

### spec.ingestion.traceUpsertQueueShardCount

`int32` · optional (explicit presence)

Shards of the trace-upsert queue (`LANGFUSE_TRACE_UPSERT_QUEUE_SHARD_COUNT`).
Never reduce it on a live instance. Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.ingestion.traceUpsertWorkerConcurrency

`int32` · optional (explicit presence)

Trace-upsert jobs each worker processes concurrently, per shard
(`LANGFUSE_TRACE_UPSERT_WORKER_CONCURRENCY`). Empty = Langfuse's
default (25).

- rule: {"int32":{"gte":1}}

### spec.ingestion.skipClickhouseReadMinProjectCreateDate

`string`

Skip the ClickHouse read before each ingestion write for projects
created on or after this date, "YYYY-MM-DD"
(`LANGFUSE_SKIP_INGESTION_CLICKHOUSE_READ_MIN_PROJECT_CREATE_DATE`).
Safe for any instance whose full event history is in blob storage -
a fresh instance: set a date before the first project.

- rule: skip_clickhouse_read_min_project_create_date must be a date, YYYY-MM-DD.

### spec.aiFeatures

`KubernetesLangfuseAiFeatures`

Langfuse AI: the instance-wide model behind the Assistant and Ask AI.

- rule: Set ai_features.provider and ai_features.model together: a model without a provider is ignored, and a provider without a model leaves Langfuse AI unconfigured.
- rule: anthropic and openai need api_key; bedrock and vertex authenticate through the cloud's credential chain and refuse one.

### spec.aiFeatures.provider

`enum`

Model provider (`LANGFUSE_AI_PROVIDER`). Unspecified = Langfuse AI
unconfigured.

Allowed values (use exactly as shown):

- `unspecified` -- Langfuse AI unconfigured.
- `bedrock` -- Amazon Bedrock.
- `anthropic` -- The Anthropic Messages API.
- `openai` -- An OpenAI-compatible endpoint.
- `vertex` -- Google Vertex AI.

### spec.aiFeatures.model

`string`

Model ID for the provider (`LANGFUSE_AI_MODEL`), e.g. a Bedrock geo
inference profile such as "eu.anthropic.claude-opus-5", or
"claude-opus-5" on Vertex.

### spec.aiFeatures.smallModel

`string`

A smaller model for conversation titles and Ask AI
(`LANGFUSE_AI_SMALL_MODEL`). Empty = `model`.

### spec.aiFeatures.apiKey

`string` · sensitive

API key (`LANGFUSE_AI_API_KEY`) - required for anthropic and openai,
refused for bedrock and vertex (they use the cloud's credential
chain: an IAM role or Google service account through workload
identity).

### spec.aiFeatures.baseUrl

`string`

API base URL including "/v1" (`LANGFUSE_AI_BASE_URL`), for a gateway
or an OpenAI-compatible endpoint.

### spec.aiFeatures.extraHeaders

`map<string, string>`

Extra request headers (`LANGFUSE_AI_EXTRA_HEADERS`), anthropic and
openai only.

### spec.aiFeatures.useResponsesApi

`bool`

Call the OpenAI Responses API instead of Chat Completions
(`LANGFUSE_AI_USE_RESPONSES_API`), openai only.

### spec.aiFeatures.bedrockRegion

`string`

Bedrock region (`LANGFUSE_AI_AWS_BEDROCK_REGION`). Empty = the
workload's region.

### spec.aiFeatures.vertexLocation

`string`

Vertex location (`LANGFUSE_AI_VERTEX_LOCATION`). Empty = "global".

### spec.aiFeatures.traceProjectId

`string`

Project on this instance that receives traces of Assistant and Ask AI
runs (`LANGFUSE_AI_FEATURES_PROJECT_ID`). Empty = nothing is traced.

### spec.aiFeatures.assistant

`KubernetesLangfuseAssistant`

The Assistant, the in-app agent.

### spec.aiFeatures.assistant.enabled

`bool`

Show the Assistant (`LANGFUSE_IN_APP_AGENT_ENABLED`, on web and
worker; the worker then consumes the run queue).

### spec.aiFeatures.assistant.runConcurrency

`int32` · optional (explicit presence)

Concurrent Assistant runs per worker
(`LANGFUSE_IN_APP_AGENT_RUN_QUEUE_PROCESSING_CONCURRENCY`). Empty =
Langfuse's default (5).

- rule: {"int32":{"gte":1}}

### spec.aiFeatures.assistant.maxActiveRunsPerUser

`int32` · optional (explicit presence)

Concurrent runs admitted per user
(`LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_USER`). Empty = 5.

- rule: {"int32":{"gte":1}}

### spec.aiFeatures.assistant.maxActiveRunsPerOrg

`int32` · optional (explicit presence)

Concurrent runs admitted per organization
(`LANGFUSE_IN_APP_AGENT_MAX_ACTIVE_RUNS_PER_ORG`). Empty = 20.

- rule: {"int32":{"gte":1}}

### spec.aiFeatures.assistant.useInternalWebUrl

`bool`

Point the worker's MCP calls at the in-cluster web Service
(`LANGFUSE_MCP_BASE_URL`, and the Service host added to the web's
`LANGFUSE_MCP_ALLOWED_HOSTS`) instead of `external_url` - for an
external URL the worker cannot reach.

### spec.aiFeatures.assistant.sandbox

`KubernetesLangfuseLambdaMicrovmSandbox`

The AWS Lambda MicroVM sandbox the Assistant runs code in. Unset =
Langfuse tools only, no code execution.

### spec.aiFeatures.assistant.sandbox.imageIdentifier

`string` · required

MicroVM image ARN.

- rule: {"required":true}

### spec.aiFeatures.assistant.sandbox.executionRoleArn

`string | valueFrom` · required

Execution role ARN. Defaults compose an AwsIamRole resource.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.aiFeatures.assistant.sandbox.region

`string` · required

AWS region of the MicroVM.

- rule: {"required":true}

### spec.aiFeatures.assistant.sandbox.egressNetworkConnectorArn

`string` · required

Egress network connector ARN - a VPC connector whose security group
denies egress. Without it AWS attaches INTERNET_EGRESS and sandboxed
code reaches the public internet.

- rule: {"required":true}

### spec.evaluation

`KubernetesLangfuseEvaluation`

Evaluation runtime: how LLM-as-a-judge evaluators receive media, and
where code evaluators execute.

### spec.evaluation.mediaTransport

`enum`

How LLM-as-a-judge evaluators send media to the model
(`LANGFUSE_EVALUATOR_MEDIA_TRANSPORT`). Unspecified = inline.

Allowed values (use exactly as shown):

- `unspecified` -- inline.
- `inline` -- The worker downloads the attachment and sends its bytes.
- `url` -- The model provider downloads a signed URL (it must reach the store).
- `disabled` -- Media references stay text.

### spec.evaluation.mediaInlineMaxBytes

`int32` · optional (explicit presence)

Largest inline attachment in bytes
(`LANGFUSE_EVALUATOR_MEDIA_INLINE_MAX_BYTES`). Empty = Langfuse's
default (20000000); evaluations with larger attachments fail.

- rule: {"int32":{"gt":0}}

### spec.evaluation.codeEvaluators

`KubernetesLangfuseCodeEvaluators`

Code evaluators: where evaluator code runs. Unset = code evaluators
are unavailable.

### spec.evaluation.codeEvaluators.awsLambda

`KubernetesLangfuseLambdaCodeEvaluators`

AWS Lambda functions - the production dispatcher.

### spec.evaluation.codeEvaluators.awsLambda.nodeFunctionName

`string`

Function for TypeScript and JavaScript evaluators. Empty =
"code-based-eval-executor-node".

### spec.evaluation.codeEvaluators.awsLambda.pythonFunctionName

`string`

Function for Python evaluators. Empty =
"code-based-eval-executor-python".

### spec.evaluation.codeEvaluators.awsLambda.endpoint

`string`

Custom Lambda endpoint (`LANGFUSE_CODE_EVAL_AWS_LAMBDA_ENDPOINT`).

### spec.evaluation.codeEvaluators.insecureLocal

`KubernetesLangfuseLocalCodeEvaluators`

Inside the worker process (`insecure-local`): only when every
evaluator's code is fully trusted.

### spec.evaluation.codeEvaluators.insecureLocal.timeoutMs

`int32` · optional (explicit presence)

Per-evaluation timeout in milliseconds
(`LANGFUSE_CODE_EVAL_LOCAL_TIMEOUT_MS`). Empty = 2000.

- rule: {"int32":{"gt":0}}

### spec.evaluation.codeEvaluators.queueShardCount

`int32` · optional (explicit presence)

Shards of the code-evaluation queue
(`LANGFUSE_CODE_EVAL_EXECUTION_QUEUE_SHARD_COUNT`). Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.evaluation.codeEvaluators.concurrency

`int32` · optional (explicit presence)

Code evaluations each worker runs concurrently
(`LANGFUSE_CODE_EVAL_EXECUTION_WORKER_CONCURRENCY`). Empty = 5.

- rule: {"int32":{"gte":1}}

### spec.slack

`KubernetesLangfuseSlack`

The Slack app that lets projects send notifications to Slack.

### spec.slack.clientId

`string` · required

Slack app client ID.

- rule: {"required":true}

### spec.slack.clientSecret

`string` · required · sensitive

Slack app client secret.

- rule: {"required":true}

### spec.slack.stateSecret

`string` · sensitive

Secret signing the OAuth state parameter. Empty = generated.

### spec.enterprise

`KubernetesLangfuseEnterprise`

Enterprise Edition features: license key, Instance Management API, UI
customization and ingestion masking.

### spec.enterprise.licenseKey

`string` · sensitive

License key (`LANGFUSE_EE_LICENSE_KEY`).

### spec.enterprise.adminApiKey

`string` · sensitive

Key for the Instance Management API (`ADMIN_API_KEY`), which manages
organizations and their API keys - what the Langfuse Terraform
provider uses. Empty = the API is off.

### spec.enterprise.ui

`KubernetesLangfuseUiCustomization`

UI customization.

### spec.enterprise.ui.documentationHref

`string`

Documentation link (`LANGFUSE_UI_DOCUMENTATION_HREF`).

### spec.enterprise.ui.supportHref

`string`

Support link (`LANGFUSE_UI_SUPPORT_HREF`).

### spec.enterprise.ui.feedbackHref

`string`

Feature request and bug report link (`LANGFUSE_UI_FEEDBACK_HREF`).

### spec.enterprise.ui.apiHost

`string`

Host shown in project settings for API requests
(`LANGFUSE_UI_API_HOST`), when API traffic uses another host than
the UI.

### spec.enterprise.ui.logoLightModeHref

`string`

Logo for light mode (`LANGFUSE_UI_LOGO_LIGHT_MODE_HREF`).

### spec.enterprise.ui.logoDarkModeHref

`string`

Logo for dark mode (`LANGFUSE_UI_LOGO_DARK_MODE_HREF`).

### spec.enterprise.ui.defaultModelAdapter

`enum`

Default model adapter in the playground and evals
(`LANGFUSE_UI_DEFAULT_MODEL_ADAPTER`).

Allowed values (use exactly as shown):

- `unspecified` -- Langfuse's default.
- `OpenAI` -- OpenAI.
- `Anthropic` -- Anthropic.
- `Azure` -- Azure OpenAI.

### spec.enterprise.ui.defaultBaseUrlOpenai

`string`

Default OpenAI base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_OPENAI`).

### spec.enterprise.ui.defaultBaseUrlAnthropic

`string`

Default Anthropic base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_ANTHROPIC`).

### spec.enterprise.ui.defaultBaseUrlAzure

`string`

Default Azure OpenAI base URL (`LANGFUSE_UI_DEFAULT_BASE_URL_AZURE`).

### spec.enterprise.ui.visibleProductModules

`[]string`

Show only these product modules (`LANGFUSE_UI_VISIBLE_PRODUCT_MODULES`),
e.g. "tracing", "evaluation", "prompt-management". Wins over
`hidden_product_modules`.

### spec.enterprise.ui.hiddenProductModules

`[]string`

Hide these product modules (`LANGFUSE_UI_HIDDEN_PRODUCT_MODULES`), e.g.
"playground", "datasets".

### spec.enterprise.ui.instanceLinks

`[]KubernetesLangfuseInstanceLink`

Links in the instance switcher (`LANGFUSE_UI_INSTANCE_LINKS`).

### spec.enterprise.ui.instanceLinks[].name

`string` · required

Label, unique among the links.

- rule: {"required":true}

### spec.enterprise.ui.instanceLinks[].url

`string` · required

URL of the other instance.

- rule: {"required":true,"string":{"uri":true}}

### spec.enterprise.ingestionMasking

`KubernetesLangfuseIngestionMasking`

Ingestion masking: a callback that redacts events before they are
stored.

### spec.enterprise.ingestionMasking.callbackUrl

`string` · required

URL of the masking callback (`LANGFUSE_INGESTION_MASKING_CALLBACK_URL`).

- rule: {"required":true,"string":{"uri":true}}

### spec.enterprise.ingestionMasking.timeoutMs

`int32` · optional (explicit presence)

Callback timeout in milliseconds. Empty = 500.

- rule: {"int32":{"gt":0}}

### spec.enterprise.ingestionMasking.failClosed

`bool`

Drop events when the callback fails (fail closed) instead of storing
them unmasked (fail open, the default).

### spec.enterprise.ingestionMasking.maxRetries

`int32` · optional (explicit presence)

Retries of a failed callback. Empty = 1.

- rule: {"int32":{"gte":0}}

### spec.enterprise.ingestionMasking.propagatedHeaders

`[]string`

Headers of the ingestion request forwarded to the callback.

### spec.observability

`KubernetesLangfuseObservability`

Langfuse's own logs and OpenTelemetry traces.

### spec.observability.logLevel

`enum`

Log level.

Allowed values (use exactly as shown):

- `log_level_unspecified` -- info.
- `trace` -- trace.
- `debug` -- debug.
- `info` -- info.
- `warn` -- warn.
- `error` -- error.
- `fatal` -- fatal.

### spec.observability.logFormat

`enum`

Log format: text, or json for log pipelines.

Allowed values (use exactly as shown):

- `log_format_unspecified` -- text.
- `text` -- Human-readable lines.
- `json` -- One JSON object per line.

### spec.observability.logPropagatedHeaders

`[]string`

Request headers carried into log context through OpenTelemetry
baggage (`LANGFUSE_LOG_PROPAGATED_HEADERS`), e.g. "x-request-id".

### spec.observability.otlpEndpoint

`string | valueFrom`

OTLP endpoint Langfuse sends its own traces to
(`OTEL_EXPORTER_OTLP_ENDPOINT`). Defaults compose a
KubernetesOtelCollector's OTLP/HTTP endpoint. Empty = no export.

- references: KubernetesOtelCollector (`status.outputs.otlp_http_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesOtelCollector, name: <that resource's name>, fieldPath: status.outputs.otlp_http_endpoint}} -- a bare string does not parse

### spec.observability.traceSamplingRatio

`double` · optional (explicit presence)

Fraction of requests traced, above 0 and at most 1
(`OTEL_TRACE_SAMPLING_RATIO`). Empty = 1.

- rule: {"double":{"lte":1,"gt":0}}

### spec.network

`KubernetesLangfuseNetwork`

Outbound networking: proxies, private certificate authorities and the
host allowlists for user-configured endpoints.

### spec.network.httpsProxy

`string`

Proxy for outbound HTTPS requests (`HTTPS_PROXY`).

### spec.network.noProxy

`[]string`

Hosts reached directly, bypassing the proxy (`NO_PROXY`), e.g.
".svc.cluster.local".

### spec.network.trustedCa

`KubernetesLangfuseCertificateFile`

Extra certificate authorities Langfuse trusts for outbound TLS
(`NODE_EXTRA_CA_CERTS`): private CAs of ClickHouse, SMTP, identity
providers or LLM endpoints.

### spec.network.trustedCa.secretName

`string`

Name of a Secret holding the bundle.

### spec.network.trustedCa.configMapName

`string`

Name of a ConfigMap holding the bundle.

### spec.network.trustedCa.key

`string` · optional (explicit presence)

Key holding the PEM bundle.

- default: `ca.crt`

### spec.network.allowlists

`KubernetesLangfuseOutboundAllowlists`

Allowlists for endpoints users configure inside Langfuse. By default
Langfuse refuses private and internal addresses for them; an entry
here admits one.

### spec.network.allowlists.webhooks

`KubernetesLangfuseAllowlist`

Webhook targets (`LANGFUSE_WEBHOOK_WHITELISTED_*`).

### spec.network.allowlists.webhooks.hosts

`[]string`

Host names admitted.

### spec.network.allowlists.webhooks.ips

`[]string`

IP addresses admitted.

- rule: Each allowlisted IP must be an IPv4 or IPv6 address.

### spec.network.allowlists.webhooks.ipSegments

`[]string`

CIDR ranges admitted.

- rule: Each allowlisted segment must be a CIDR range, such as 10.0.0.0/8.

### spec.network.allowlists.llmConnections

`KubernetesLangfuseAllowlist`

LLM connection base URLs (`LANGFUSE_LLM_CONNECTION_WHITELISTED_*`).

### spec.network.allowlists.llmConnections.hosts

`[]string`

Host names admitted.

### spec.network.allowlists.llmConnections.ips

`[]string`

IP addresses admitted.

- rule: Each allowlisted IP must be an IPv4 or IPv6 address.

### spec.network.allowlists.llmConnections.ipSegments

`[]string`

CIDR ranges admitted.

- rule: Each allowlisted segment must be a CIDR range, such as 10.0.0.0/8.

### spec.network.allowlists.blobStorageEndpoints

`KubernetesLangfuseAllowlist`

Blob storage export endpoints (`LANGFUSE_BLOB_STORAGE_ENDPOINT_WHITELISTED_*`).

### spec.network.allowlists.blobStorageEndpoints.hosts

`[]string`

Host names admitted.

### spec.network.allowlists.blobStorageEndpoints.ips

`[]string`

IP addresses admitted.

- rule: Each allowlisted IP must be an IPv4 or IPv6 address.

### spec.network.allowlists.blobStorageEndpoints.ipSegments

`[]string`

CIDR ranges admitted.

- rule: Each allowlisted segment must be a CIDR range, such as 10.0.0.0/8.

### spec.network.allowlists.externalMediaStorageEndpoints

`KubernetesLangfuseAllowlist`

External media storage endpoints
(`LANGFUSE_EXTERNAL_MEDIA_STORAGE_ENDPOINT_WHITELISTED_*`).

### spec.network.allowlists.externalMediaStorageEndpoints.hosts

`[]string`

Host names admitted.

### spec.network.allowlists.externalMediaStorageEndpoints.ips

`[]string`

IP addresses admitted.

- rule: Each allowlisted IP must be an IPv4 or IPv6 address.

### spec.network.allowlists.externalMediaStorageEndpoints.ipSegments

`[]string`

CIDR ranges admitted.

- rule: Each allowlisted segment must be a CIDR range, such as 10.0.0.0/8.

### spec.network.allowlists.ssoDiscovery

`KubernetesLangfuseAllowlist`

SSO discovery URLs (`LANGFUSE_SSO_DISCOVERY_WHITELISTED_*`).

### spec.network.allowlists.ssoDiscovery.hosts

`[]string`

Host names admitted.

### spec.network.allowlists.ssoDiscovery.ips

`[]string`

IP addresses admitted.

- rule: Each allowlisted IP must be an IPv4 or IPv6 address.

### spec.network.allowlists.ssoDiscovery.ipSegments

`[]string`

CIDR ranges admitted.

- rule: Each allowlisted segment must be a CIDR range, such as 10.0.0.0/8.

### spec.web

`KubernetesLangfuseWeb`

The web server: UI, public API and ingestion endpoints.

- rule: A variable may be declared in extra_env or extra_env_from_secret, not both.

### spec.web.image

`KubernetesLangfuseImage`

Web image.

### spec.web.image.repository

`string`

Image repository. Empty = the chart's
(docker.langfuse.com/langfuse/langfuse for web,
docker.langfuse.com/langfuse/langfuse-worker for the worker).

### spec.web.image.tag

`string`

Image tag. Empty = the chart's appVersion for `chart_version`. Keep
web and worker on the same version.

### spec.web.image.pullPolicy

`string`

Pull policy: Always, IfNotPresent or Never. Empty = the chart's
default, Always.

- rule: Image pull policy must be one of: Always, IfNotPresent, Never.

### spec.web.image.pullSecretNames

`[]string`

Names of existing image pull Secrets in Langfuse's namespace (for a
private mirror).

### spec.web.replicas

`int32` · optional (explicit presence)

Replicas. Two or more for availability. Ignored while an autoscaler
owns the count.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.web.resources

`ContainerResources`

CPU and memory. Langfuse recommends 2 CPU and 4 GiB per web pod for
production; the defaults suit evaluation.

### spec.web.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.web.resources.limits.cpu

`string`

### spec.web.resources.limits.memory

`string`

### spec.web.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.web.resources.requests.cpu

`string`

### spec.web.resources.requests.memory

`string`

### spec.web.autoscaling

`KubernetesLangfuseAutoscaling`

Autoscaling. One mechanism at most.

- rule: KEDA owns the replica count and the resources it scales on: enable KEDA alone, or HPA and VPA without KEDA.

### spec.web.autoscaling.hpa

`KubernetesLangfuseHpa`

HorizontalPodAutoscaler on CPU.

- rule: The autoscaler's minimum replica count cannot exceed its maximum.

### spec.web.autoscaling.hpa.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.web.autoscaling.hpa.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.web.autoscaling.hpa.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target average CPU utilization percent.

- default: `50`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.web.autoscaling.hpa.targetMemoryUtilizationPercent

`int32` · optional (explicit presence)

Target average memory utilization percent. Empty = memory does not
drive scaling.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.web.autoscaling.hpa.annotations

`map<string, string>`

Annotations on the HorizontalPodAutoscaler.

### spec.web.autoscaling.keda

`KubernetesLangfuseKeda`

KEDA ScaledObject (needs KEDA on the cluster - a KubernetesKeda).

- rule: The autoscaler's minimum replica count cannot exceed its maximum.

### spec.web.autoscaling.keda.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.web.autoscaling.keda.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.web.autoscaling.keda.pollingIntervalSeconds

`int32` · optional (explicit presence)

Seconds between metric checks.

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.web.autoscaling.keda.trigger

`enum`

Resource the trigger scales on.

Allowed values (use exactly as shown):

- `trigger_unspecified` -- cpu.
- `cpu` -- CPU.
- `memory` -- Memory.

### spec.web.autoscaling.keda.metricType

`enum`

How `target_value` reads: a utilization percent or an absolute
average value.

Allowed values (use exactly as shown):

- `metric_type_unspecified` -- Utilization.
- `Utilization` -- A percent of the requested resource.
- `AverageValue` -- An absolute average per pod.

### spec.web.autoscaling.keda.targetValue

`string` · optional (explicit presence)

Target value, e.g. "50" (percent) or "1500m" (average CPU).

- default: `50`

### spec.web.autoscaling.keda.containerName

`string`

Container whose metrics count. Empty = every container in the pod.

### spec.web.autoscaling.vpa

`KubernetesLangfuseVpa`

VerticalPodAutoscaler. The chart renders it only on a cluster that
serves the autoscaling.k8s.io/v1 API (the VPA CRDs); elsewhere it is
skipped silently.

### spec.web.autoscaling.vpa.controlledResources

`[]string`

Resources the VPA controls, e.g. "cpu", "memory". Empty = both.

### spec.web.autoscaling.vpa.maxAllowed

`CpuMemory`

Upper bounds for recommendations.

### spec.web.autoscaling.vpa.maxAllowed.cpu

`string`

### spec.web.autoscaling.vpa.maxAllowed.memory

`string`

### spec.web.autoscaling.vpa.minAllowed

`CpuMemory`

Lower bounds for recommendations.

### spec.web.autoscaling.vpa.minAllowed.cpu

`string`

### spec.web.autoscaling.vpa.minAllowed.memory

`string`

### spec.web.autoscaling.vpa.updateMode

`string` · optional (explicit presence)

How recommendations are applied: Off, Initial, Recreate or Auto.

- default: `Auto`
- rule: VPA update_mode must be one of: Off, Initial, Recreate, Auto.

### spec.web.pdb

`KubernetesLangfusePdb`

PodDisruptionBudget.

- rule: A PodDisruptionBudget takes either min_available or max_unavailable, not both.

### spec.web.pdb.enabled

`bool` · optional (explicit presence)

Create the PodDisruptionBudget.

- default: `true`

### spec.web.pdb.minAvailable

`string`

Minimum pods available during a voluntary disruption: an integer or a
percentage.

- rule: min_available must be a non-negative integer or a percentage up to 100%, such as 50%.

### spec.web.pdb.maxUnavailable

`string`

Maximum pods unavailable during a voluntary disruption. Empty with
min_available also empty = 1.

- rule: max_unavailable must be a non-negative integer or a percentage up to 100%, such as 25%.

### spec.web.livenessProbe

`KubernetesLangfuseProbe`

Liveness probe on `/api/public/health`.

### spec.web.livenessProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after start before the first probe. Empty = 20 (the chart
reads 0 as unset).

- rule: {"int32":{"gte":1}}

### spec.web.livenessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probes. Empty = 10.

- rule: {"int32":{"gte":1}}

### spec.web.livenessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds before a probe times out. Empty = 5.

- rule: {"int32":{"gte":1}}

### spec.web.livenessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe fails. Empty = 3.

- rule: {"int32":{"gte":1}}

### spec.web.livenessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe passes. Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.web.readinessProbe

`KubernetesLangfuseProbe`

Readiness probe on `/api/public/ready`.

### spec.web.readinessProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after start before the first probe. Empty = 20 (the chart
reads 0 as unset).

- rule: {"int32":{"gte":1}}

### spec.web.readinessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probes. Empty = 10.

- rule: {"int32":{"gte":1}}

### spec.web.readinessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds before a probe times out. Empty = 5.

- rule: {"int32":{"gte":1}}

### spec.web.readinessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe fails. Empty = 3.

- rule: {"int32":{"gte":1}}

### spec.web.readinessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe passes. Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.web.service

`KubernetesLangfuseWebService`

The web Service.

### spec.web.service.type

`string` · optional (explicit presence)

Service type: ClusterIP, NodePort or LoadBalancer. Keep ClusterIP and
compose a route or ingress for external access.

- default: `ClusterIP`
- rule: Service type must be one of: ClusterIP, NodePort, LoadBalancer.

### spec.web.service.port

`int32` · optional (explicit presence)

Port the web server listens on, and the Service's target port.

- default: `3000`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.web.service.externalPort

`int32` · optional (explicit presence)

Port the Service exposes, when it differs from `port`.

- rule: {"int32":{"lte":65535,"gt":0}}

### spec.web.service.nodePort

`int32` · optional (explicit presence)

Node port, for type NodePort.

- rule: {"int32":{"lte":32767,"gte":30000}}

### spec.web.service.annotations

`map<string, string>`

Annotations on the Service (internal load balancer recipes, for
example).

### spec.web.service.labels

`map<string, string>`

Extra labels on the Service.

### spec.web.scheduling

`KubernetesLangfuseScheduling`

Pod placement.

### spec.web.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

### spec.web.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.web.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.web.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.web.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.web.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.web.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.web.scheduling.nodeAffinity

`WorkloadNodeAffinity`

Node affinity: hard requirements and weighted preferences.

### spec.web.scheduling.nodeAffinity.required

`[]WorkloadNodeSelectorTerm`

Hard requirement. The outer list ORs its terms; expressions within one term AND.

### spec.web.scheduling.nodeAffinity.required[].matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.web.scheduling.nodeAffinity.required[].matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.web.scheduling.nodeAffinity.required[].matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.web.scheduling.nodeAffinity.required[].matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.web.scheduling.nodeAffinity.preferred

`[]WorkloadPreferredNodeSelectorTerm`

Weighted soft preferences.

### spec.web.scheduling.nodeAffinity.preferred[].weight

`int32`

Preference weight, 1–100. Higher weights dominate placement scoring.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.web.scheduling.nodeAffinity.preferred[].term

`WorkloadNodeSelectorTerm` · required

- rule: {"required":true}

### spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.web.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.web.scheduling.podAffinity

`WorkloadPodAffinity`

Pod affinity toward pods already running elsewhere.

### spec.web.scheduling.podAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.web.scheduling.podAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.web.scheduling.podAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.web.scheduling.podAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.web.scheduling.podAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.web.scheduling.podAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.web.scheduling.podAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.web.scheduling.podAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.web.scheduling.podAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.web.scheduling.podAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.web.scheduling.podAntiAffinity

`WorkloadPodAffinity`

Pod anti-affinity - spread replicas across nodes or zones.

### spec.web.scheduling.podAntiAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.web.scheduling.podAntiAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.web.scheduling.podAntiAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.web.scheduling.podAntiAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.web.scheduling.podAntiAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.web.scheduling.podAntiAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.web.scheduling.podAntiAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.web.scheduling.podAntiAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.web.scheduling.podAntiAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.web.scheduling.podAntiAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.web.scheduling.topologySpreadConstraints

`[]WorkloadTopologySpreadConstraint`

Topology spread constraints.

### spec.web.scheduling.topologySpreadConstraints[].maxSkew

`int32`

Maximum allowed difference in matching-pod counts between any two domains.
1 is the strictest even spread.

- rule: {"int32":{"gte":1}}

### spec.web.scheduling.topologySpreadConstraints[].topologyKey

`string` · required

Node label defining the domains to spread across (e.g.
"topology.kubernetes.io/zone").

- rule: {"required":true}

### spec.web.scheduling.topologySpreadConstraints[].whenUnsatisfiable

`string` · required

What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).

- rule: whenUnsatisfiable must be either "DoNotSchedule" or "ScheduleAnyway"
- rule: {"required":true}

### spec.web.scheduling.topologySpreadConstraints[].matchLabels

`map<string, string>`

Labels selecting the pods counted per domain. Omit to have the module default to
the workload's own selector labels — self-spreading, the overwhelmingly common
intent.

### spec.web.scheduling.priorityClassName

`string`

PriorityClass for the pods.

### spec.web.workload

`KubernetesLangfuseWorkloadShape`

Workload metadata and rollout.

### spec.web.workload.podAnnotations

`map<string, string>`

Annotations on the pods.

### spec.web.workload.podLabels

`map<string, string>`

Labels on the pods.

### spec.web.workload.deploymentAnnotations

`map<string, string>`

Annotations on the Deployment.

### spec.web.workload.deploymentLabels

`map<string, string>`

Extra labels on the Deployment.

### spec.web.workload.strategy

`KubernetesLangfuseRolloutStrategy`

Rollout strategy.

- rule: max_surge and max_unavailable apply only to RollingUpdate.

### spec.web.workload.strategy.type

`string`

Strategy type: RollingUpdate or Recreate. Empty = RollingUpdate.

- rule: Rollout strategy type must be RollingUpdate or Recreate.

### spec.web.workload.strategy.maxSurge

`string`

Extra pods during a rolling update: an integer or a percentage.

### spec.web.workload.strategy.maxUnavailable

`string`

Pods unavailable during a rolling update: an integer or a percentage.

### spec.web.workload.revisionHistoryLimit

`int32` · optional (explicit presence)

Old ReplicaSets kept for rollback. Empty = 10.

- rule: {"int32":{"gte":0}}

### spec.web.workload.hostAliases

`[]WorkloadHostAlias`

Extra /etc/hosts entries.

### spec.web.workload.hostAliases[].ip

`string` · required

The IP the hostnames resolve to.

- rule: {"required":true,"string":{"ip":true}}

### spec.web.workload.hostAliases[].hostnames

`[]string` · required

Hostnames mapped to the IP.

- rule: {"repeated":{"minItems":"1"}}

### spec.web.extraEnv

`map<string, string>` · no secrets

Extra plain environment variables for the web server only.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `extraEnvFromSecret`, which keeps it in a secret store the workload reads by reference
### spec.web.extraEnvFromSecret

`map<string, KubernetesSecretKey>`

Extra environment variables for the web server read from existing
Secrets, keyed by variable name.

### spec.web.extraEnvFromSecret.*.name

`string`

The name of the Kubernetes Secret.

### spec.web.extraEnvFromSecret.*.key

`string`

The key within the Kubernetes Secret.

### spec.worker

`KubernetesLangfuseWorker`

The worker: queue consumers, evaluations, exports and retention.

- rule: A variable may be declared in extra_env or extra_env_from_secret, not both.

### spec.worker.image

`KubernetesLangfuseImage`

Worker image.

### spec.worker.image.repository

`string`

Image repository. Empty = the chart's
(docker.langfuse.com/langfuse/langfuse for web,
docker.langfuse.com/langfuse/langfuse-worker for the worker).

### spec.worker.image.tag

`string`

Image tag. Empty = the chart's appVersion for `chart_version`. Keep
web and worker on the same version.

### spec.worker.image.pullPolicy

`string`

Pull policy: Always, IfNotPresent or Never. Empty = the chart's
default, Always.

- rule: Image pull policy must be one of: Always, IfNotPresent, Never.

### spec.worker.image.pullSecretNames

`[]string`

Names of existing image pull Secrets in Langfuse's namespace (for a
private mirror).

### spec.worker.replicas

`int32` · optional (explicit presence)

Replicas. Scale on CPU: a worker above 50% of 2 CPU is saturated.
Ignored while an autoscaler owns the count.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.worker.resources

`ContainerResources`

CPU and memory. Langfuse recommends 2 CPU and 4 GiB per worker for
production; the defaults suit evaluation.

### spec.worker.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.worker.resources.limits.cpu

`string`

### spec.worker.resources.limits.memory

`string`

### spec.worker.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.worker.resources.requests.cpu

`string`

### spec.worker.resources.requests.memory

`string`

### spec.worker.autoscaling

`KubernetesLangfuseAutoscaling`

Autoscaling. One mechanism at most.

- rule: KEDA owns the replica count and the resources it scales on: enable KEDA alone, or HPA and VPA without KEDA.

### spec.worker.autoscaling.hpa

`KubernetesLangfuseHpa`

HorizontalPodAutoscaler on CPU.

- rule: The autoscaler's minimum replica count cannot exceed its maximum.

### spec.worker.autoscaling.hpa.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.worker.autoscaling.hpa.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.worker.autoscaling.hpa.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target average CPU utilization percent.

- default: `50`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.worker.autoscaling.hpa.targetMemoryUtilizationPercent

`int32` · optional (explicit presence)

Target average memory utilization percent. Empty = memory does not
drive scaling.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.worker.autoscaling.hpa.annotations

`map<string, string>`

Annotations on the HorizontalPodAutoscaler.

### spec.worker.autoscaling.keda

`KubernetesLangfuseKeda`

KEDA ScaledObject (needs KEDA on the cluster - a KubernetesKeda).

- rule: The autoscaler's minimum replica count cannot exceed its maximum.

### spec.worker.autoscaling.keda.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.worker.autoscaling.keda.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.worker.autoscaling.keda.pollingIntervalSeconds

`int32` · optional (explicit presence)

Seconds between metric checks.

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.worker.autoscaling.keda.trigger

`enum`

Resource the trigger scales on.

Allowed values (use exactly as shown):

- `trigger_unspecified` -- cpu.
- `cpu` -- CPU.
- `memory` -- Memory.

### spec.worker.autoscaling.keda.metricType

`enum`

How `target_value` reads: a utilization percent or an absolute
average value.

Allowed values (use exactly as shown):

- `metric_type_unspecified` -- Utilization.
- `Utilization` -- A percent of the requested resource.
- `AverageValue` -- An absolute average per pod.

### spec.worker.autoscaling.keda.targetValue

`string` · optional (explicit presence)

Target value, e.g. "50" (percent) or "1500m" (average CPU).

- default: `50`

### spec.worker.autoscaling.keda.containerName

`string`

Container whose metrics count. Empty = every container in the pod.

### spec.worker.autoscaling.vpa

`KubernetesLangfuseVpa`

VerticalPodAutoscaler. The chart renders it only on a cluster that
serves the autoscaling.k8s.io/v1 API (the VPA CRDs); elsewhere it is
skipped silently.

### spec.worker.autoscaling.vpa.controlledResources

`[]string`

Resources the VPA controls, e.g. "cpu", "memory". Empty = both.

### spec.worker.autoscaling.vpa.maxAllowed

`CpuMemory`

Upper bounds for recommendations.

### spec.worker.autoscaling.vpa.maxAllowed.cpu

`string`

### spec.worker.autoscaling.vpa.maxAllowed.memory

`string`

### spec.worker.autoscaling.vpa.minAllowed

`CpuMemory`

Lower bounds for recommendations.

### spec.worker.autoscaling.vpa.minAllowed.cpu

`string`

### spec.worker.autoscaling.vpa.minAllowed.memory

`string`

### spec.worker.autoscaling.vpa.updateMode

`string` · optional (explicit presence)

How recommendations are applied: Off, Initial, Recreate or Auto.

- default: `Auto`
- rule: VPA update_mode must be one of: Off, Initial, Recreate, Auto.

### spec.worker.pdb

`KubernetesLangfusePdb`

PodDisruptionBudget.

- rule: A PodDisruptionBudget takes either min_available or max_unavailable, not both.

### spec.worker.pdb.enabled

`bool` · optional (explicit presence)

Create the PodDisruptionBudget.

- default: `true`

### spec.worker.pdb.minAvailable

`string`

Minimum pods available during a voluntary disruption: an integer or a
percentage.

- rule: min_available must be a non-negative integer or a percentage up to 100%, such as 50%.

### spec.worker.pdb.maxUnavailable

`string`

Maximum pods unavailable during a voluntary disruption. Empty with
min_available also empty = 1.

- rule: max_unavailable must be a non-negative integer or a percentage up to 100%, such as 25%.

### spec.worker.livenessProbe

`KubernetesLangfuseProbe`

Liveness probe on `/api/health`.

### spec.worker.livenessProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after start before the first probe. Empty = 20 (the chart
reads 0 as unset).

- rule: {"int32":{"gte":1}}

### spec.worker.livenessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probes. Empty = 10.

- rule: {"int32":{"gte":1}}

### spec.worker.livenessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds before a probe times out. Empty = 5.

- rule: {"int32":{"gte":1}}

### spec.worker.livenessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe fails. Empty = 3.

- rule: {"int32":{"gte":1}}

### spec.worker.livenessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe passes. Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.worker.failLivenessIfQueueConsumptionStuck

`bool`

Fail the liveness probe when the worker has consumed no job on any
queue for longer than `queue_consumption_stuck_threshold_minutes`, so
a wedged worker restarts itself. Scheduled jobs keep a healthy worker
busy at least hourly.

### spec.worker.queueConsumptionStuckThresholdMinutes

`int32` · optional (explicit presence)

Minutes without consumption before the probe above fails
(`LANGFUSE_QUEUE_CONSUMPTION_STUCK_THRESHOLD_MINUTES`). Empty =
Langfuse's default (60).

- rule: {"int32":{"gte":1}}

### spec.worker.scheduling

`KubernetesLangfuseScheduling`

Pod placement.

### spec.worker.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

### spec.worker.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.worker.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.worker.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.worker.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.worker.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.worker.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.worker.scheduling.nodeAffinity

`WorkloadNodeAffinity`

Node affinity: hard requirements and weighted preferences.

### spec.worker.scheduling.nodeAffinity.required

`[]WorkloadNodeSelectorTerm`

Hard requirement. The outer list ORs its terms; expressions within one term AND.

### spec.worker.scheduling.nodeAffinity.required[].matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.worker.scheduling.nodeAffinity.required[].matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.worker.scheduling.nodeAffinity.preferred

`[]WorkloadPreferredNodeSelectorTerm`

Weighted soft preferences.

### spec.worker.scheduling.nodeAffinity.preferred[].weight

`int32`

Preference weight, 1–100. Higher weights dominate placement scoring.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.worker.scheduling.nodeAffinity.preferred[].term

`WorkloadNodeSelectorTerm` · required

- rule: {"required":true}

### spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions

`[]WorkloadNodeSelectorRequirement` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: In/NotIn require at least one value, Gt/Lt exactly one, Exists/DoesNotExist none

### spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].key

`string` · required

Node label key, e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].operator

`string` · required

Operator: "In"/"NotIn" (value set), "Exists"/"DoesNotExist" (key presence), or
"Gt"/"Lt" (single integer value, as strings — the Kubernetes API convention).

- rule: Operator must be one of "In", "NotIn", "Exists", "DoesNotExist", "Gt", or "Lt"
- rule: {"required":true}

### spec.worker.scheduling.nodeAffinity.preferred[].term.matchExpressions[].values

`[]string`

Values for the operator: required non-empty for In/NotIn, exactly one integer
string for Gt/Lt, and must be empty for Exists/DoesNotExist.

### spec.worker.scheduling.podAffinity

`WorkloadPodAffinity`

Pod affinity toward pods already running elsewhere.

### spec.worker.scheduling.podAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.worker.scheduling.podAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.worker.scheduling.podAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.worker.scheduling.podAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.worker.scheduling.podAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.worker.scheduling.podAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.worker.scheduling.podAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.worker.scheduling.podAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.worker.scheduling.podAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.worker.scheduling.podAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.worker.scheduling.podAntiAffinity

`WorkloadPodAffinity`

Pod anti-affinity - spread replicas across nodes or zones.

### spec.worker.scheduling.podAntiAffinity.required

`[]WorkloadPodAffinityTerm`

Hard rules — unschedulable until satisfied. Use sparingly; they can deadlock rollouts.

### spec.worker.scheduling.podAntiAffinity.required[].matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.worker.scheduling.podAntiAffinity.required[].topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.worker.scheduling.podAntiAffinity.required[].namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.worker.scheduling.podAntiAffinity.preferred

`[]WorkloadWeightedPodAffinityTerm`

Weighted soft rules — the scheduler's tiebreakers.

### spec.worker.scheduling.podAntiAffinity.preferred[].weight

`int32`

Preference weight, 1–100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.worker.scheduling.podAntiAffinity.preferred[].term

`WorkloadPodAffinityTerm` · required

- rule: {"required":true}

### spec.worker.scheduling.podAntiAffinity.preferred[].term.matchLabels

`map<string, string>` · required

Labels of the pods to match against — for self-anti-affinity, the workload's own
selector labels (exported as the `selector_labels` output).

- rule: {"map":{"minPairs":"1"}}

### spec.worker.scheduling.podAntiAffinity.preferred[].term.topologyKey

`string` · required

Node label defining the domain: "kubernetes.io/hostname" separates by node,
"topology.kubernetes.io/zone" by zone.

- rule: {"required":true}

### spec.worker.scheduling.podAntiAffinity.preferred[].term.namespaces

`[]string`

Namespaces whose pods are considered. Empty means the workload's own namespace.

### spec.worker.scheduling.topologySpreadConstraints

`[]WorkloadTopologySpreadConstraint`

Topology spread constraints.

### spec.worker.scheduling.topologySpreadConstraints[].maxSkew

`int32`

Maximum allowed difference in matching-pod counts between any two domains.
1 is the strictest even spread.

- rule: {"int32":{"gte":1}}

### spec.worker.scheduling.topologySpreadConstraints[].topologyKey

`string` · required

Node label defining the domains to spread across (e.g.
"topology.kubernetes.io/zone").

- rule: {"required":true}

### spec.worker.scheduling.topologySpreadConstraints[].whenUnsatisfiable

`string` · required

What happens when the constraint cannot be met: "DoNotSchedule" (hard — pod
stays Pending) or "ScheduleAnyway" (soft — scheduler minimizes skew).

- rule: whenUnsatisfiable must be either "DoNotSchedule" or "ScheduleAnyway"
- rule: {"required":true}

### spec.worker.scheduling.topologySpreadConstraints[].matchLabels

`map<string, string>`

Labels selecting the pods counted per domain. Omit to have the module default to
the workload's own selector labels — self-spreading, the overwhelmingly common
intent.

### spec.worker.scheduling.priorityClassName

`string`

PriorityClass for the pods.

### spec.worker.workload

`KubernetesLangfuseWorkloadShape`

Workload metadata and rollout.

### spec.worker.workload.podAnnotations

`map<string, string>`

Annotations on the pods.

### spec.worker.workload.podLabels

`map<string, string>`

Labels on the pods.

### spec.worker.workload.deploymentAnnotations

`map<string, string>`

Annotations on the Deployment.

### spec.worker.workload.deploymentLabels

`map<string, string>`

Extra labels on the Deployment.

### spec.worker.workload.strategy

`KubernetesLangfuseRolloutStrategy`

Rollout strategy.

- rule: max_surge and max_unavailable apply only to RollingUpdate.

### spec.worker.workload.strategy.type

`string`

Strategy type: RollingUpdate or Recreate. Empty = RollingUpdate.

- rule: Rollout strategy type must be RollingUpdate or Recreate.

### spec.worker.workload.strategy.maxSurge

`string`

Extra pods during a rolling update: an integer or a percentage.

### spec.worker.workload.strategy.maxUnavailable

`string`

Pods unavailable during a rolling update: an integer or a percentage.

### spec.worker.workload.revisionHistoryLimit

`int32` · optional (explicit presence)

Old ReplicaSets kept for rollback. Empty = 10.

- rule: {"int32":{"gte":0}}

### spec.worker.workload.hostAliases

`[]WorkloadHostAlias`

Extra /etc/hosts entries.

### spec.worker.workload.hostAliases[].ip

`string` · required

The IP the hostnames resolve to.

- rule: {"required":true,"string":{"ip":true}}

### spec.worker.workload.hostAliases[].hostnames

`[]string` · required

Hostnames mapped to the IP.

- rule: {"repeated":{"minItems":"1"}}

### spec.worker.extraEnv

`map<string, string>` · no secrets

Extra plain environment variables for the worker only (queue consumer
switches such as QUEUE_CONSUMER_BATCH_EXPORT_QUEUE_IS_ENABLED, for
example).

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `extraEnvFromSecret`, which keeps it in a secret store the workload reads by reference
### spec.worker.extraEnvFromSecret

`map<string, KubernetesSecretKey>`

Extra environment variables for the worker read from existing
Secrets, keyed by variable name.

### spec.worker.extraEnvFromSecret.*.name

`string`

The name of the Kubernetes Secret.

### spec.worker.extraEnvFromSecret.*.key

`string`

The key within the Kubernetes Secret.

### spec.serviceAccount

`KubernetesLangfuseServiceAccount`

The ServiceAccount both workloads run as, and its cloud identity.

- rule: An existing ServiceAccount carries its own annotations: leave annotations and workload_identity unset when existing_name is set.

### spec.serviceAccount.existingName

`string`

Run as an existing ServiceAccount instead of the one the chart
creates (`<metadata.name>`). Its annotations are then yours to
manage; `annotations` and `workload_identity` are refused.

### spec.serviceAccount.annotations

`map<string, string>`

Annotations on the ServiceAccount the chart creates.

### spec.serviceAccount.workloadIdentity

`KubernetesWorkloadIdentity`

The cloud identity the ServiceAccount federates with - keyless access
to S3, GCS, Bedrock, Vertex and SES. Rendered as the ServiceAccount
annotation each cloud expects. On AKS, also label the pods
`azure.workload.identity/use: "true"` through `workload.pod_labels`.

### spec.serviceAccount.workloadIdentity.gke

`KubernetesWorkloadIdentityGke`

GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.

### spec.serviceAccount.workloadIdentity.gke.serviceAccountEmail

`string | valueFrom` · required

GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
email or a reference to a GcpServiceAccount resource's output.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.serviceAccount.workloadIdentity.eks

`KubernetesWorkloadIdentityEksIrsa`

EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.

### spec.serviceAccount.workloadIdentity.eks.roleArn

`string | valueFrom` · required

AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.serviceAccount.workloadIdentity.aks

`KubernetesWorkloadIdentityAks`

Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
(or user-assigned managed identity) client ID.

### spec.serviceAccount.workloadIdentity.aks.clientId

`string | valueFrom` · required

Client ID (GUID) of the user-assigned managed identity or Entra application.
Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
GUID or a reference to an AzureUserAssignedIdentity resource's output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.serviceAccount.workloadIdentity.aks.tenantId

`string` · optional (explicit presence)

Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
omitted the azure-workload-identity webhook uses its default tenant. Applied as
the `azure.workload.identity/tenant-id` annotation when set.

- rule: {"string":{"uuid":true}}

### spec.serviceAccount.automountToken

`bool` · optional (explicit presence)

Mount the ServiceAccount token into the pods. Workload identity on
EKS and GKE needs it.

- default: `true`

### spec.podSecurityContext

`WorkloadPodSecurityContext`

Pod-level security context for both workloads.

### spec.podSecurityContext.runAsUser

`int64` · optional (explicit presence)

UID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsGroup

`int64` · optional (explicit presence)

Primary GID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsNonRoot

`bool` · optional (explicit presence)

Refuse to start any container whose effective user is root.

### spec.podSecurityContext.fsGroup

`int64` · optional (explicit presence)

GID that owns mounted volumes and is added to every container's supplemental
groups — the standard fix for "permission denied" on persistent volumes written
by non-root apps.

### spec.podSecurityContext.fsGroupChangePolicy

`string`

When volume ownership is re-chowned to fs_group: "Always" (default) or
"OnRootMismatch" (skip the recursive chown when the root already matches —
dramatically faster pod starts on large volumes).

- rule: fsGroupChangePolicy must be either "Always" or "OnRootMismatch"

### spec.podSecurityContext.supplementalGroups

`[]int64`

Additional group IDs applied to all container processes.

### spec.podSecurityContext.sysctls

`[]WorkloadSysctl`

Kernel parameters set for the pod. Only safe sysctls (or those the cluster
administrator has allow-listed on the kubelet) are admitted.

### spec.podSecurityContext.sysctls[].name

`string` · required

Sysctl name, e.g. "net.core.somaxconn".

- rule: {"required":true}

### spec.podSecurityContext.sysctls[].value

`string` · required

Sysctl value, e.g. "1024".

- rule: {"required":true}

### spec.podSecurityContext.seccompProfile

`WorkloadSeccompProfile`

Pod-wide seccomp profile; containers may override with their own.

- rule: localhost_profile is required when type is "Localhost" and must be empty otherwise

### spec.podSecurityContext.seccompProfile.type

`string` · required

Profile type: "RuntimeDefault" (the container runtime's default filter — the
recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
file installed on the node, named via localhost_profile).

- rule: Seccomp profile type must be one of "RuntimeDefault", "Unconfined", or "Localhost"
- rule: {"required":true}

### spec.podSecurityContext.seccompProfile.localhostProfile

`string`

Path of the profile file relative to the node's seccomp profile root. Required
when (and only meaningful when) type is "Localhost".

### spec.containerSecurityContext

`WorkloadContainerSecurityContext`

Container-level security context for both workloads.

### spec.containerSecurityContext.privileged

`bool`

Runs the container with full host access — equivalent to root on the node.
Required by some node-level agents (device managers, network plugins). Never
combine with untrusted images.

### spec.containerSecurityContext.runAsUser

`int64` · optional (explicit presence)

UID the container process runs as. Overrides the image's USER directive.

### spec.containerSecurityContext.runAsGroup

`int64` · optional (explicit presence)

Primary GID the container process runs as.

### spec.containerSecurityContext.runAsNonRoot

`bool` · optional (explicit presence)

Refuses to start the container if its effective user is root. The standard
baseline hardening — it catches images that silently default to UID 0.

### spec.containerSecurityContext.readOnlyRootFilesystem

`bool` · optional (explicit presence)

Mounts the container's root filesystem read-only. Pair with EmptyDir mounts for
paths the app must write (e.g. /tmp).

### spec.containerSecurityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

Whether the process can gain more privileges than its parent (setuid binaries,
file capabilities). The restricted Pod Security Standard requires this to be
false. Always true when `privileged` is set, so leave it unset in that case.

### spec.containerSecurityContext.capabilities

`WorkloadCapabilities`

Linux capabilities to add or drop. The restricted profile drops ALL and adds
back only NET_BIND_SERVICE when needed. Capability names are uppercase without
the CAP_ prefix (e.g. "NET_ADMIN", "SYS_TIME").

### spec.containerSecurityContext.capabilities.add

`[]string`

Capabilities to add (e.g. "NET_BIND_SERVICE").

### spec.containerSecurityContext.capabilities.drop

`[]string`

Capabilities to drop. Use ["ALL"] as the hardened baseline.

### spec.containerSecurityContext.seccompProfile

`WorkloadSeccompProfile`

Seccomp syscall filter for the container. "RuntimeDefault" is the hardened
baseline; "Localhost" selects a node-local profile file via `localhost_profile`.

- rule: localhost_profile is required when type is "Localhost" and must be empty otherwise

### spec.containerSecurityContext.seccompProfile.type

`string` · required

Profile type: "RuntimeDefault" (the container runtime's default filter — the
recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
file installed on the node, named via localhost_profile).

- rule: Seccomp profile type must be one of "RuntimeDefault", "Unconfined", or "Localhost"
- rule: {"required":true}

### spec.containerSecurityContext.seccompProfile.localhostProfile

`string`

Path of the profile file relative to the node's seccomp profile root. Required
when (and only meaningful when) type is "Localhost".

### spec.dnsConfig

`WorkloadPodDnsConfig`

DNS settings for both workloads' pods.

### spec.dnsConfig.nameservers

`[]string`

Nameserver IPs (max 3 total after merging with the policy's servers).

- rule: {"repeated":{"maxItems":"3"}}

### spec.dnsConfig.searches

`[]string`

Search domains for hostname lookup.

### spec.dnsConfig.options

`[]WorkloadPodDnsConfigOption`

resolv.conf options, e.g. name "ndots" value "2". Value may be empty for flags.

### spec.dnsConfig.options[].name

`string` · required

- rule: {"required":true}

### spec.dnsConfig.options[].value

`string`

### spec.extraVolumes

`[]VolumeMount`

Extra volumes mounted into both workloads - a file a variable in
`extra_env` points at, for example.

### spec.extraVolumes[].name

`string` · required

Name of the volume mount. Must be unique within the container.
Used to correlate with the volume definition.

- rule: {"required":true}

### spec.extraVolumes[].mountPath

`string` · required

Path within the container at which the volume should be mounted.
Must be an absolute path.

- rule: {"required":true}

### spec.extraVolumes[].readOnly

`bool`

Whether the volume should be mounted read-only.
Default is false.

### spec.extraVolumes[].subPath

`string`

Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).
Useful for mounting a subdirectory of a volume.

### spec.extraVolumes[].configMap

`ConfigMapVolumeSource`

ConfigMap volume source.
Use this to mount a ConfigMap as a file or directory.

### spec.extraVolumes[].configMap.name

`string` · required

Name of the ConfigMap to mount.
Can reference a ConfigMap defined in spec.config_maps or an existing one in the namespace.

- rule: {"required":true}

### spec.extraVolumes[].configMap.key

`string`

Specific key from the ConfigMap to mount as a single file.
If not specified, all keys are mounted as files in the directory.

### spec.extraVolumes[].configMap.path

`string`

If key is specified, this is the filename to use for the mounted file.
Defaults to the key name if not specified.
Example: key="config" path="app.yaml" mounts the "config" key as "app.yaml"

### spec.extraVolumes[].configMap.defaultMode

`int32`

Mode bits to use on created files. Must be a value between 0 and 0777.
Defaults to 0644.
Use 0755 (493 in decimal) for executable scripts.

### spec.extraVolumes[].secret

`SecretVolumeSource`

Secret volume source.
Use this to mount a Secret as a file or directory.

### spec.extraVolumes[].secret.name

`string` · required

Name of the Secret to mount.

- rule: {"required":true}

### spec.extraVolumes[].secret.key

`string`

Specific key from the Secret to mount as a single file.
If not specified, all keys are mounted as files in the directory.

### spec.extraVolumes[].secret.path

`string`

If key is specified, this is the filename to use for the mounted file.
Defaults to the key name if not specified.

### spec.extraVolumes[].secret.defaultMode

`int32`

Mode bits to use on created files. Must be a value between 0 and 0777.
Defaults to 0644.

### spec.extraVolumes[].hostPath

`HostPathVolumeSource`

HostPath volume source.
Use this to mount a file or directory from the host node's filesystem.
Common for DaemonSets that need access to node-level resources.

### spec.extraVolumes[].hostPath.path

`string` · required

Path on the host to mount.

- rule: {"required":true}

### spec.extraVolumes[].hostPath.type

`string`

Type of the host path.
Valid values:
  "" - Empty string (default) means no check is performed before mounting
  "DirectoryOrCreate" - Create directory if it doesn't exist
  "Directory" - Directory must exist
  "FileOrCreate" - Create file if it doesn't exist
  "File" - File must exist
  "Socket" - UNIX socket must exist
  "CharDevice" - Character device must exist
  "BlockDevice" - Block device must exist

- rule: Type must be one of: "", "DirectoryOrCreate", "Directory", "FileOrCreate", "File", "Socket", "CharDevice", "BlockDevice"

### spec.extraVolumes[].emptyDir

`EmptyDirVolumeSource`

EmptyDir volume source.
Use this for temporary storage that is erased when the pod is removed.
Useful for scratch space, caching, or sharing data between containers.

### spec.extraVolumes[].emptyDir.medium

`string`

Medium for the empty directory.
"" (default) uses the node's default medium (typically disk).
"Memory" uses a tmpfs (RAM-backed filesystem).

Memory-backed volumes are faster but:
- Count against container memory limits
- Are lost on node restart
- Should have sizeLimit set to prevent OOM

- rule: Medium must be either "" or "Memory"

### spec.extraVolumes[].emptyDir.sizeLimit

`string`

Size limit for the empty directory.
Format: Kubernetes quantity (e.g., "1Gi", "500Mi").
Only strictly enforced when medium is "Memory".
For disk-backed volumes, this is a best-effort limit.

### spec.extraVolumes[].pvc

`PvcVolumeSource`

PersistentVolumeClaim volume source.
Use this to mount an existing PVC.
For StatefulSets, this can reference a volumeClaimTemplate.

### spec.extraVolumes[].pvc.claimName

`string` · required

Name of the PersistentVolumeClaim to mount.
For StatefulSets, this can be the name of a volumeClaimTemplate.

- rule: {"required":true}

### spec.extraVolumes[].pvc.readOnly

`bool`

Whether the PVC should be mounted read-only.
Default is false.

### spec.extraVolumes[].serviceAccountToken

`ServiceAccountTokenVolumeSource`

Projected ServiceAccount token volume source.
Use this to mount a short-lived, audience-bound identity token that the
kubelet issues for the pod's ServiceAccount and rotates automatically.

### spec.extraVolumes[].serviceAccountToken.audience

`string` · required

Intended audience of the token. The receiving service must identify
itself with this audience when verifying the token; a token minted for a
different audience is rejected. Required: an audience-less token would be
replayable against any service in the cluster.

- rule: {"required":true}

### spec.extraVolumes[].serviceAccountToken.expirationSeconds

`int64`

Requested lifetime of the token in seconds. The kubelet starts rotating
the token when it passes 80% of its lifetime or 24 hours, whichever is
shorter. Defaults to 3600 (1 hour). The Kubernetes API enforces a
minimum of 600 (10 minutes).

- rule: Expiration must be at least 600 seconds (the Kubernetes API minimum)

### spec.extraVolumes[].serviceAccountToken.path

`string`

Filename for the token relative to the mount path.
Defaults to "token".

### spec.extraEnv

`map<string, string>` · no secrets

Extra plain environment variables for both workloads: any Langfuse
variable this spec does not type, set verbatim (e.g.
LANGFUSE_INGESTION_QUEUE_DELAY_MS). A variable the spec types is set
through its field; naming it here too conflicts with the chart's own
validation. Secret values belong in `extra_env_from_secret`.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `extraEnvFromSecret`, which keeps it in a secret store the workload reads by reference
### spec.extraEnvFromSecret

`map<string, KubernetesSecretKey>`

Extra environment variables for both workloads read from existing
Secrets in Langfuse's namespace, keyed by variable name.

### spec.extraEnvFromSecret.*.name

`string`

The name of the Kubernetes Secret.

### spec.extraEnvFromSecret.*.key

`string`

The key within the Kubernetes Secret.

### spec.helmValues

`string`

Advanced escape hatch: raw Helm values merged LAST (Helm `-f`
semantics) over everything this spec renders - later keys win. Use it
for chart keys this spec does not type (sidecars, init containers,
lifecycle hooks, `extraManifests`). The module re-pins
`fullnameOverride` and keeps the bundled stores off after the merge.
Never put secret material here: it lands in the Helm release in plain
text. YAML document as a string.

## Validation Rules

- `spec.extra_env.unique_names`: A variable may be declared in extra_env or extra_env_from_secret, not both.
- `spec.auth.sso_requires_external_url`: Single sign-on needs external_url, the public address people open Langfuse at: the identity provider sends people back to <external_url>/api/auth/callback/<provider>, and with the localhost default no browser outside a port-forward can complete sign-in. Set external_url to the URL your route or ingress serves (e.g. https://langfuse.example.com).
- `spec.blob_storage.one_family`: Langfuse selects Google Cloud Storage or Azure Blob Storage for every upload type at once: when any blob storage target is gcs (or azure_blob), the event, media and batch-export targets must all be gcs (or all azure_blob). S3, R2 and S3-compatible targets mix freely.
- `spec.ai_features.assistant_requires_model`: The Assistant needs the instance-wide Langfuse AI model: set ai_features.provider and ai_features.model, or turn the Assistant off.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesLangfuse, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Langfuse runs in. |
| `status.outputs.service` | `string` | The web Service name (`<metadata.name>-web`), the backend a KubernetesHttpRoute or KubernetesIngress exposes. |
| `status.outputs.endpoint` | `string` | In-cluster base URL of the web server (e.g. "http://langfuse-web.langfuse.svc.cluster.local:3000") - the `LANGFUSE_BASE_URL` (or `LANGFUSE_HOST`) an in-cluster service's SDK uses. Browsers and sign-in use `external_url` instead. |
| `status.outputs.external_url` | `string` | The public URL Langfuse was configured with (`spec.external_url`) - what people open, and the base of every link Langfuse writes into emails and notifications. |
| `status.outputs.otel_endpoint` | `string` | In-cluster OTLP/HTTP traces endpoint (`<endpoint>/api/public/otel`). OpenTelemetry exporters authenticate with HTTP Basic auth built from the project key pair (public key as user, secret key as password). |
| `status.outputs.port_forward_command` | `string` | Copy-paste command for reaching the web UI from a workstation. |
| `status.outputs.instance_secret_name` | `string` | The module-owned Secret holding every secret value this resource manages: the salt, encryption key and NextAuth secret, the seeded credentials, and every sensitive setting the spec declares. Back up its `salt` and `encryption-key` keys: they are bound to the Postgres data, and recreating Langfuse against that data needs them declared again. |
| `status.outputs.organization_id` | `string` | ID of the organization seeded at first start. Empty when seeding is disabled. |
| `status.outputs.project_id` | `string` | ID of the project seeded at first start. Empty when seeding is disabled. |
| `status.outputs.project_public_key` | `string` | Public key of the seeded project key pair (`pk-lf-...`). Not secret: it identifies the project; the secret key authenticates. |
| `status.outputs.project_secret_key_secret` | `KubernetesSecretKey` | Where the seeded project secret key (`sk-lf-...`) lives - a key of the instance Secret. Services read it through a secretKeyRef. |
| `status.outputs.project_secret_key_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.project_secret_key_secret.key` | `string` | The key within the Kubernetes Secret. |
| `status.outputs.admin_email` | `string` | Email of the seeded admin user. Empty when no admin is seeded. |
| `status.outputs.admin_password_secret` | `KubernetesSecretKey` | Where the seeded admin user's password lives - a key of the instance Secret. Empty when no admin is seeded. |
| `status.outputs.admin_password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.admin_password_secret.key` | `string` | The key within the Kubernetes Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.postgres.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.postgres.passwordSecret.secretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.clickhouse.host` | KubernetesClickHouse | `status.outputs.service_name` |
| `spec.clickhouse.passwordSecret.secretName` | KubernetesClickHouse | `status.outputs.auth_secret_name` |
| `spec.clickhouse.cluster.name` | KubernetesClickHouse | `status.outputs.cluster_name` |
| `spec.redis.standalone.host` | KubernetesValkey | `status.outputs.service` |
| `spec.redis.username` | KubernetesValkey | `status.outputs.username` |
| `spec.redis.passwordSecret.secretName` | KubernetesValkey | `status.outputs.password_secret.name` |
| `spec.redis.passwordSecret.secretKey` | KubernetesValkey | `status.outputs.password_secret.key` |
| `spec.redis.tls.clientCertificate.secretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.redis.tls.clientKey.secretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.blobStorage.target.s3Compatible.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.blobStorage.target.s3Compatible.credentialsSecret.secretName` | KubernetesSeaweedFs | `status.outputs.s3_credentials_secret_name` |
| `spec.blobStorage.target.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.blobStorage.target.s3.region` | AwsS3Bucket | `status.outputs.region` |
| `spec.blobStorage.target.r2.endpoint` | CloudflareR2Bucket | `status.outputs.s3_endpoint` |
| `spec.blobStorage.target.r2.bucket` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.blobStorage.target.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.blobStorage.target.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.blobStorage.target.gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.blobStorage.target.azureBlob.storageAccount` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.blobStorage.target.azureBlob.container` | AzureStorageContainer | `status.outputs.container_name` |
| `spec.blobStorage.target.azureBlob.endpoint` | AzureStorageAccount | `status.outputs.primary_blob_endpoint` |
| `spec.blobStorage.eventUpload.sse.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.blobStorage.mediaUpload.target.s3Compatible.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.blobStorage.mediaUpload.target.s3Compatible.credentialsSecret.secretName` | KubernetesSeaweedFs | `status.outputs.s3_credentials_secret_name` |
| `spec.blobStorage.mediaUpload.target.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.blobStorage.mediaUpload.target.s3.region` | AwsS3Bucket | `status.outputs.region` |
| `spec.blobStorage.mediaUpload.target.r2.endpoint` | CloudflareR2Bucket | `status.outputs.s3_endpoint` |
| `spec.blobStorage.mediaUpload.target.r2.bucket` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.blobStorage.mediaUpload.target.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.blobStorage.mediaUpload.target.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.blobStorage.mediaUpload.target.gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.blobStorage.mediaUpload.target.azureBlob.storageAccount` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.blobStorage.mediaUpload.target.azureBlob.container` | AzureStorageContainer | `status.outputs.container_name` |
| `spec.blobStorage.mediaUpload.target.azureBlob.endpoint` | AzureStorageAccount | `status.outputs.primary_blob_endpoint` |
| `spec.blobStorage.mediaUpload.sse.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.blobStorage.batchExport.target.s3Compatible.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.blobStorage.batchExport.target.s3Compatible.credentialsSecret.secretName` | KubernetesSeaweedFs | `status.outputs.s3_credentials_secret_name` |
| `spec.blobStorage.batchExport.target.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.blobStorage.batchExport.target.s3.region` | AwsS3Bucket | `status.outputs.region` |
| `spec.blobStorage.batchExport.target.r2.endpoint` | CloudflareR2Bucket | `status.outputs.s3_endpoint` |
| `spec.blobStorage.batchExport.target.r2.bucket` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.blobStorage.batchExport.target.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.blobStorage.batchExport.target.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.blobStorage.batchExport.target.gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.blobStorage.batchExport.target.azureBlob.storageAccount` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.blobStorage.batchExport.target.azureBlob.container` | AzureStorageContainer | `status.outputs.container_name` |
| `spec.blobStorage.batchExport.target.azureBlob.endpoint` | AzureStorageAccount | `status.outputs.primary_blob_endpoint` |
| `spec.blobStorage.batchExport.sse.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.aiFeatures.assistant.sandbox.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.observability.otlpEndpoint` | KubernetesOtelCollector | `status.outputs.otlp_http_endpoint` |
| `spec.serviceAccount.workloadIdentity.gke.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.serviceAccount.workloadIdentity.eks.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.serviceAccount.workloadIdentity.aks.clientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)

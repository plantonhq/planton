# KubernetesPlantonPlatform

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesPlantonPlatformSpec** declares a self-hosted Planton
platform — control plane, web console, identity server (Keycloak),
PostgreSQL (CloudNativePG), cache, workflow engine (Temporal), secrets
manager (OpenBAO), and an in-cluster deployment runner — as a
`PlantonPlatform` custom resource that the Planton operator reconciles.

PREREQUISITE: the Planton operator must be installed first (declare a
KubernetesPlantonOperator resource). One operator serves EVERY platform
on the cluster; each platform lives in its own namespace with its own
URL, identity server, and databases.

ZERO-CONFIG BY DESIGN: `version` is the only required choice. A
version-only platform serves console, API, and sign-in on one origin
through a built-in gateway reachable with a single `kubectl
port-forward` (the exact command is this resource's
`port_forward_command` output), and the console's first visitor becomes
the admin using a setup code read from a Secret (the
`setup_code_command` output). Everything else — a real hostname and
TLS, storage classes and sizes, workload identity for the runner —
is opt-in refinement of a working platform.

VERSION IS A DELIBERATE CHOICE, ALWAYS: the field is required and never
defaulted, because a default that moved with catalog updates would
silently upgrade a running platform — databases and all — on an
ordinary re-apply. Upgrades are a one-line edit of this field.

SPEC SURFACE AND OPERATOR VERSION: this spec models the PlantonPlatform
schema as of the catalog release. A newer operator may accept fields
this catalog does not model yet — reach them by updating the catalog,
or through the `planton` umbrella Helm chart, whose values pass the
platform spec through verbatim.

DESTROY BEHAVIOR: deleting the resource deletes the platform — every
workload, Service, Secret, and volume the operator created is
owner-referenced to this declaration and garbage-collected by
Kubernetes (so teardown completes even when the operator itself is
already gone), and the database layer removes its volumes and
credentials together — no orphaned volume holds a password a reinstall
cannot match. Two residues are known: build caches and workflow
volumes can survive in the namespace (deleting the namespace, automatic
when this resource owned it via create_namespace, sweeps them), and the
platform's namespace-qualified token-review ClusterRole/Binding
lingers inert (its subject ServiceAccount is deleted with the
platform) until an operator release adds the janitor. Cluster-shared
sub-operators (CloudNativePG, Tekton) deliberately stay — sibling
platforms may ride them.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs — every
# typed field exercised with realistic values. DELIBERATELY NOT a live
# scenario: this shape declares a real cluster's fittings (a `gp3`
# StorageClass, an nginx IngressClass, a Let's Encrypt ClusterIssuer, a
# license Secret), none of which exist on the lane's single-node kind
# cluster, so a live run here could only fail for reasons that say nothing
# about the platform. The zero-config `scenarios/minimal.yaml` is the live
# install proof; the full surface earns its live proof on a cluster that
# has these fittings.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  create_namespace: true
  version: v0.0.50
  license:
    secret_key_ref:
      name: planton-license
      key: key
  storage:
    storage_class_name: gp3
    size: 20Gi
  database:
    postgresql:
      replicas: 2
      storage_size: 20Gi
      # The platform's own database archives to a Cloudflare R2 bucket. The
      # values here are placeholders in the right shape (an offline plan
      # cannot resolve a reference); the real declaration is BY REFERENCE
      # to the bucket and token resources — see presets/06-backups-to-r2.yaml.
      backup:
        object_store:
          destination_path: s3://planton-backups/planton
          r2:
            account_id:
              value: 0123456789abcdef0123456789abcdef
            jurisdiction:
              value: default
            credentials:
              access_key_id:
                value: replace-with-the-token-id
              secret_access_key:
                value: replace-with-the-token-secret
        retention_policy: 30d
        schedule: "0 0 2 * * *"
    redis:
      storage_size: 2Gi
  ingress:
    enabled: true
    hostname: planton.example.com
    ingress_class_name: nginx
    annotations:
      nginx.ingress.kubernetes.io/proxy-body-size: 64m
    tls:
      issuer:
        name: letsencrypt
        kind: ClusterIssuer
    reachability: public
  gateway:
    local_port: 8080
  email:
    from:
      address: no-reply@planton.example.com
      name: Planton
    reply_to: it-help@example.com
    smtp:
      host: smtp.office365.com
      port: 587
      security: starttls
      credentials_secret_name: planton-email
      ca_bundle_secret_ref:
        name: corp-ca
        key: ca.crt
  identity:
    realm: planton
    admin_email: admin@example.com
  bootstrap:
    organization:
      slug: acme
      name: Acme Corp
    environment:
      slug: production
      name: Production
    admins:
      - platform-team@example.com
    iac_provisioner: tofu
    secret_backend:
      type: platform
  runner:
    enabled: true
    storage_size: 4Gi
    service_account_annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/planton-runner
  build:
    enabled: true
  vault:
    enabled: true
    init_mode: auto
    storage_size: 2Gi
  components:
    graph:
      enabled: true
      storage_size: 10Gi
  prerequisites:
    postgres_operator: auto
    tekton_pipelines: auto
    postgres_backup_plugin: auto
  control_plane:
    replicas: 1
  console:
    replicas: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` | yes |  |  |
| `spec.license` | `KubernetesPlantonPlatformLicense` |  |  |  |
| `spec.license.key` | `string` (sensitive) |  |  |  |
| `spec.license.secretKeyRef` | `KubernetesPlantonPlatformSecretKeyRef` |  |  |  |
| `spec.license.secretKeyRef.name` | `string` | yes |  |  |
| `spec.license.secretKeyRef.key` | `string` | yes |  |  |
| `spec.storage` | `KubernetesPlantonPlatformStorage` |  |  |  |
| `spec.storage.storageClassName` | `string` |  |  |  |
| `spec.storage.size` | `string` |  |  |  |
| `spec.database` | `KubernetesPlantonPlatformDatabase` |  |  |  |
| `spec.database.postgresql` | `KubernetesPlantonPlatformPostgresql` |  |  |  |
| `spec.database.postgresql.replicas` | `int32` |  | `1` |  |
| `spec.database.postgresql.storageSize` | `string` |  |  |  |
| `spec.database.postgresql.storageClassName` | `string` |  |  |  |
| `spec.database.postgresql.backup` | `KubernetesPlantonPlatformPostgresqlBackup` |  |  |  |
| `spec.database.postgresql.backup.objectStore` | `KubernetesPlantonPlatformObjectStore` | yes |  |  |
| `spec.database.postgresql.backup.objectStore.destinationPath` | `string` | yes |  |  |
| `spec.database.postgresql.backup.objectStore.s3` | `KubernetesPlantonPlatformS3ObjectStore` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.region` | `string` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.endpointUrl` | `string` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.endpointCaPem` | `string` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.accessKeys` | `KubernetesPlantonPlatformS3AccessKeys` |  |  |  |
| `spec.database.postgresql.backup.objectStore.s3.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.database.postgresql.backup.objectStore.s3.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.database.postgresql.backup.objectStore.gcs` | `KubernetesPlantonPlatformGcsObjectStore` |  |  |  |
| `spec.database.postgresql.backup.objectStore.gcs.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.backup.objectStore.gcs.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.database.postgresql.backup.objectStore.azureBlob` | `KubernetesPlantonPlatformAzureBlobObjectStore` |  |  |  |
| `spec.database.postgresql.backup.objectStore.azureBlob.storageAccount` | `string` | yes |  |  |
| `spec.database.postgresql.backup.objectStore.azureBlob.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.backup.objectStore.azureBlob.connectionString` | `string` (sensitive) |  |  |  |
| `spec.database.postgresql.backup.objectStore.r2` | `KubernetesPlantonPlatformR2ObjectStore` |  |  |  |
| `spec.database.postgresql.backup.objectStore.r2.accountId` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.account_id`) |
| `spec.database.postgresql.backup.objectStore.r2.jurisdiction` | `string \| valueFrom` |  |  | CloudflareR2Bucket (`status.outputs.jurisdiction`) |
| `spec.database.postgresql.backup.objectStore.r2.credentials` | `KubernetesPlantonPlatformR2Credentials` | yes |  |  |
| `spec.database.postgresql.backup.objectStore.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.database.postgresql.backup.objectStore.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.database.postgresql.backup.retentionPolicy` | `string` |  | `30d` |  |
| `spec.database.postgresql.backup.schedule` | `string` |  | `0 0 2 * * *` |  |
| `spec.database.postgresql.backup.serviceAccountAnnotations` | `map<string, string>` |  |  |  |
| `spec.database.postgresql.recoverFrom` | `KubernetesPlantonPlatformPostgresqlRecoverFrom` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore` | `KubernetesPlantonPlatformObjectStore` | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.destinationPath` | `string` | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3` | `KubernetesPlantonPlatformS3ObjectStore` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.region` | `string` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.endpointUrl` | `string` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.endpointCaPem` | `string` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys` | `KubernetesPlantonPlatformS3AccessKeys` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.gcs` | `KubernetesPlantonPlatformGcsObjectStore` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.gcs.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.gcs.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.azureBlob` | `KubernetesPlantonPlatformAzureBlobObjectStore` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.azureBlob.storageAccount` | `string` | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.azureBlob.keyless` | `bool` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.azureBlob.connectionString` | `string` (sensitive) |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.r2` | `KubernetesPlantonPlatformR2ObjectStore` |  |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.r2.accountId` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.account_id`) |
| `spec.database.postgresql.recoverFrom.objectStore.r2.jurisdiction` | `string \| valueFrom` |  |  | CloudflareR2Bucket (`status.outputs.jurisdiction`) |
| `spec.database.postgresql.recoverFrom.objectStore.r2.credentials` | `KubernetesPlantonPlatformR2Credentials` | yes |  |  |
| `spec.database.postgresql.recoverFrom.objectStore.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.database.postgresql.recoverFrom.objectStore.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.database.postgresql.recoverFrom.serverName` | `string` | yes |  |  |
| `spec.database.postgresql.recoverFrom.targetTime` | `string` |  |  |  |
| `spec.database.redis` | `KubernetesPlantonPlatformRedis` |  |  |  |
| `spec.database.redis.storageSize` | `string` |  |  |  |
| `spec.database.redis.storageClassName` | `string` |  |  |  |
| `spec.ingress` | `KubernetesPlantonPlatformIngress` |  |  |  |
| `spec.ingress.enabled` | `bool` |  |  |  |
| `spec.ingress.hostname` | `string` |  |  |  |
| `spec.ingress.ingressClassName` | `string` |  |  |  |
| `spec.ingress.annotations` | `map<string, string>` |  |  |  |
| `spec.ingress.tls` | `KubernetesPlantonPlatformIngressTls` |  |  |  |
| `spec.ingress.tls.secretName` | `string` |  |  |  |
| `spec.ingress.tls.issuer` | `KubernetesPlantonPlatformCertManagerIssuer` |  |  |  |
| `spec.ingress.tls.issuer.name` | `string` | yes |  |  |
| `spec.ingress.tls.issuer.kind` | `string` |  | `Issuer` |  |
| `spec.ingress.gatewayRef` | `KubernetesPlantonPlatformGatewayRef` |  |  |  |
| `spec.ingress.gatewayRef.name` | `string \| valueFrom` | yes |  | KubernetesGateway (`status.outputs.gateway_name`) |
| `spec.ingress.gatewayRef.namespace` | `string \| valueFrom` |  |  | KubernetesGateway (`status.outputs.namespace`) |
| `spec.ingress.gatewayRef.sectionName` | `string` |  |  |  |
| `spec.ingress.reachability` | `string` |  | `auto` |  |
| `spec.gateway` | `KubernetesPlantonPlatformGateway` |  |  |  |
| `spec.gateway.localPort` | `int32` |  | `8080` |  |
| `spec.identity` | `KubernetesPlantonPlatformIdentity` |  |  |  |
| `spec.identity.realm` | `string` |  | `planton` |  |
| `spec.identity.adminEmail` | `string` |  |  |  |
| `spec.bootstrap` | `KubernetesPlantonPlatformBootstrap` |  |  |  |
| `spec.bootstrap.organization` | `KubernetesPlantonPlatformBootstrapOrg` |  |  |  |
| `spec.bootstrap.organization.slug` | `string` |  | `default` |  |
| `spec.bootstrap.organization.name` | `string` |  |  |  |
| `spec.bootstrap.environment` | `KubernetesPlantonPlatformBootstrapEnv` |  |  |  |
| `spec.bootstrap.environment.slug` | `string` |  | `default` |  |
| `spec.bootstrap.environment.name` | `string` |  |  |  |
| `spec.bootstrap.admins` | `[]string` |  |  |  |
| `spec.bootstrap.iacProvisioner` | `string` |  | `tofu` |  |
| `spec.bootstrap.secretBackend` | `KubernetesPlantonPlatformSecretBackend` |  |  |  |
| `spec.bootstrap.secretBackend.type` | `string` | yes |  |  |
| `spec.bootstrap.secretBackend.awsSecretsManager` | `KubernetesPlantonPlatformAwsSecretsManager` |  |  |  |
| `spec.bootstrap.secretBackend.awsSecretsManager.region` | `string` | yes |  |  |
| `spec.bootstrap.secretBackend.awsSecretsManager.kmsKeyArn` | `string` | yes |  |  |
| `spec.runner` | `KubernetesPlantonPlatformRunner` |  |  |  |
| `spec.runner.enabled` | `bool` |  | `true` |  |
| `spec.runner.storageSize` | `string` |  |  |  |
| `spec.runner.storageClassName` | `string` |  |  |  |
| `spec.runner.serviceAccountAnnotations` | `map<string, string>` |  |  |  |
| `spec.runner.cloudCredentialsSecretName` | `string` |  |  |  |
| `spec.build` | `KubernetesPlantonPlatformBuild` |  |  |  |
| `spec.build.enabled` | `bool` |  | `true` |  |
| `spec.vault` | `KubernetesPlantonPlatformVault` |  |  |  |
| `spec.vault.enabled` | `bool` |  | `true` |  |
| `spec.vault.initMode` | `string` |  | `auto` |  |
| `spec.vault.storageSize` | `string` |  |  |  |
| `spec.vault.storageClassName` | `string` |  |  |  |
| `spec.components` | `KubernetesPlantonPlatformComponents` |  |  |  |
| `spec.components.graph` | `KubernetesPlantonPlatformGraph` |  |  |  |
| `spec.components.graph.enabled` | `bool` |  |  |  |
| `spec.components.graph.storageSize` | `string` |  |  |  |
| `spec.components.graph.storageClassName` | `string` |  |  |  |
| `spec.prerequisites` | `KubernetesPlantonPlatformPrerequisites` |  |  |  |
| `spec.prerequisites.postgresOperator` | `string` |  | `auto` |  |
| `spec.prerequisites.tektonPipelines` | `string` |  | `auto` |  |
| `spec.prerequisites.postgresBackupPlugin` | `string` |  | `auto` |  |
| `spec.controlPlane` | `KubernetesPlantonPlatformControlPlane` |  |  |  |
| `spec.controlPlane.image` | `KubernetesPlantonPlatformImage` |  |  |  |
| `spec.controlPlane.image.repository` | `string` |  |  |  |
| `spec.controlPlane.image.tag` | `string` |  |  |  |
| `spec.controlPlane.replicas` | `int32` |  | `1` |  |
| `spec.controlPlane.externalConfigSecretName` | `string` |  |  |  |
| `spec.controlPlane.serviceAccountAnnotations` | `map<string, string>` |  |  |  |
| `spec.console` | `KubernetesPlantonPlatformConsole` |  |  |  |
| `spec.console.image` | `KubernetesPlantonPlatformImage` |  |  |  |
| `spec.console.image.repository` | `string` |  |  |  |
| `spec.console.image.tag` | `string` |  |  |  |
| `spec.console.replicas` | `int32` |  | `1` |  |
| `spec.console.externalConfigSecretName` | `string` |  |  |  |
| `spec.remoteRunners` | `KubernetesPlantonPlatformRemoteRunners` |  |  |  |
| `spec.remoteRunners.enabled` | `bool` |  | `false` |  |
| `spec.email` | `KubernetesPlantonPlatformEmail` |  |  |  |
| `spec.email.from` | `KubernetesPlantonPlatformEmailFrom` | yes |  |  |
| `spec.email.from.address` | `string` | yes |  |  |
| `spec.email.from.name` | `string` |  | `Planton` |  |
| `spec.email.replyTo` | `string` |  |  |  |
| `spec.email.smtp` | `KubernetesPlantonPlatformEmailSmtp` |  |  |  |
| `spec.email.smtp.host` | `string` | yes |  |  |
| `spec.email.smtp.port` | `int32` |  | `587` |  |
| `spec.email.smtp.security` | `string` |  | `starttls` |  |
| `spec.email.smtp.credentialsSecretName` | `string` |  |  |  |
| `spec.email.smtp.oauth2` | `KubernetesPlantonPlatformEmailSmtpOauth2` |  |  |  |
| `spec.email.smtp.oauth2.user` | `string` | yes |  |  |
| `spec.email.smtp.oauth2.tokenUrl` | `string` | yes |  |  |
| `spec.email.smtp.oauth2.scope` | `string` | yes |  |  |
| `spec.email.smtp.oauth2.clientId` | `string` | yes |  |  |
| `spec.email.smtp.oauth2.clientSecretRef` | `KubernetesPlantonPlatformSecretKeyRef` | yes |  |  |
| `spec.email.smtp.oauth2.clientSecretRef.name` | `string` | yes |  |  |
| `spec.email.smtp.oauth2.clientSecretRef.key` | `string` | yes |  |  |
| `spec.email.smtp.caBundleSecretRef` | `KubernetesPlantonPlatformSecretKeyRef` |  |  |  |
| `spec.email.smtp.caBundleSecretRef.name` | `string` | yes |  |  |
| `spec.email.smtp.caBundleSecretRef.key` | `string` | yes |  |  |
| `spec.email.resend` | `KubernetesPlantonPlatformEmailResend` |  |  |  |
| `spec.email.resend.apiKeySecretRef` | `KubernetesPlantonPlatformSecretKeyRef` | yes |  |  |
| `spec.email.resend.apiKeySecretRef.name` | `string` | yes |  |  |
| `spec.email.resend.apiKeySecretRef.key` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The namespace the platform lives in. Accepts a literal namespace name
or a reference to a KubernetesNamespace resource. Every platform
workload, Service, Secret, and volume is created here, named from
this resource's metadata.name.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the platform is declared, and deleted with
the resource. When false, the namespace must already exist.

### spec.version

`string` · required

The Planton platform version to deploy (e.g.
"v0.0.45") — pins the control plane, console, and
runner images as one coherent line. REQUIRED, never defaulted:
changing it is how a platform upgrades, and upgrades of a system
holding your data must always be a deliberate act.

- rule: {"string":{"minLen":"1"}}

### spec.license

`KubernetesPlantonPlatformLicense`

License activation. Without one the platform runs the Community
feature set. Set AT MOST one of key or secret_key_ref.

- rule: set at most one of key or secret_key_ref — one license, one delivery form

### spec.license.key

`string` · sensitive

The license key, inline. This is a secret: supply it as a
managed-secret reference, never inline plaintext.

### spec.license.secretKeyRef

`KubernetesPlantonPlatformSecretKeyRef`

Read the license key from an existing Kubernetes Secret in the
platform's namespace instead.

### spec.license.secretKeyRef.name

`string` · required

Secret name (in the platform's namespace).

- rule: {"string":{"minLen":"1"}}

### spec.license.secretKeyRef.key

`string` · required

Key within the Secret holding the value.

- rule: {"string":{"minLen":"1"}}

### spec.storage

`KubernetesPlantonPlatformStorage`

Platform-wide storage defaults: every persistent volume the platform
creates (databases, cache, workflow engine, secrets manager, runner
state) uses these unless its component overrides them.
Unset means the cluster's default StorageClass and each component's
built-in size.

### spec.storage.storageClassName

`string`

StorageClass for every platform volume unless a component overrides
it. Unset = the cluster's default class (the operator verifies the
default class can actually provision before deploying).

### spec.storage.size

`string`

Size for every platform volume unless a component overrides it
(e.g. "10Gi"). Useful when the storage backend enforces a minimum
volume size — one value lifts every volume above the floor.

- rule: size must be a Kubernetes quantity like "10Gi" or "800Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.database

`KubernetesPlantonPlatformDatabase`

The platform's own data stores (PostgreSQL and the Redis-protocol
cache). Sizing and placement only — the stores themselves are always
deployed; the platform cannot run without them.

### spec.database.postgresql

`KubernetesPlantonPlatformPostgresql`

PostgreSQL (CloudNativePG-managed).

### spec.database.postgresql.replicas

`int32` · optional (explicit presence)

PostgreSQL instances: 1 is a single primary; 2+ grows the SAME
database into a streaming-replication pair with automatic failover —
a LIVE edit, no reinstall.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.database.postgresql.storageSize

`string`

Volume size per instance (e.g. "10Gi"). Falls back to
spec.storage.size, then the platform default.

- rule: storage_size must be a Kubernetes quantity like "10Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.database.postgresql.storageClassName

`string`

StorageClass override for the database volumes.

### spec.database.postgresql.backup

`KubernetesPlantonPlatformPostgresqlBackup`

Where the platform's own database is backed up, and how. Declaring it
turns on continuous WAL archiving to the store, a base backup on the
schedule (the first one the moment backups are declared — WAL alone
restores nothing), and a retention the store enforces; the operator
installs the Barman Cloud plugin, CloudNativePG's backup engine, if
the cluster does not have it yet (see prerequisites.postgres_backup_plugin).

Absent means no backup: the database lives on one volume in this
cluster and nothing copies it anywhere. The `BACKUP` column of
`kubectl get plantonplatform` says which — `NotConfigured`,
`Deploying`, `Healthy`, `Failing` (in the plugin's own words), or
`Unavailable` — and `status.backup` carries the archive's server name,
the first recoverability point, and the last successful base backup.
A failing backup never takes a working platform out of Ready.

The module creates the credential Secret this store needs BEFORE the
platform resource, in the same apply, so the database is born
archiving whichever order the declaration's parts are read in.

### spec.database.postgresql.backup.objectStore

`KubernetesPlantonPlatformObjectStore` · required

The object store the backups go to and how the database's pods
authenticate to it. WAL archiving into it starts as soon as the
database is healthy; the schedule adds the base backups a
point-in-time recovery replays WAL onto.

- rule: {"required":true}
- rule: the s3 backend stores at an s3:// destination path (also for S3-compatible stores like MinIO)
- rule: the gcs backend stores at a gs:// destination path
- rule: the azure_blob backend stores at an https:// destination path (https://<account>.blob.core.windows.net/<container>/<path>)
- rule: the r2 backend stores at an s3:// destination path (s3://<bucket>/<path> — R2 is addressed through its S3 API; the bucket name is the CloudflareR2Bucket's bucket_name)

### spec.database.postgresql.backup.objectStore.destinationPath

`string` · required

Where in the store the archive lives — the backend's native URI form:
`s3://bucket/path` for S3, Cloudflare R2, and every S3-compatible
store, `gs://bucket/path` for GCS, and
`https://<account>.blob.core.windows.net/<container>/<path>` for
Azure Blob. Base backups and WAL are filed beneath it under this
platform's server name, so several platforms can share one path
without ever touching each other's archive.

- rule: {"required":true}

### spec.database.postgresql.backup.objectStore.s3

`KubernetesPlantonPlatformS3ObjectStore`

AWS S3 — or ANY S3-compatible store (MinIO, Ceph RGW, DigitalOcean
Spaces, ...) via the endpoint_url override. Cloudflare R2 has its own
arm (`r2`) that composes the catalog's Cloudflare kinds.

- rule: keyless and access_keys are alternative credential postures — set exactly one
- rule: an S3-compatible endpoint (endpoint_url) authenticates with access_keys — the keyless posture only mints AWS credentials

### spec.database.postgresql.backup.objectStore.s3.region

`string`

AWS region of the bucket. Required for real S3; for S3-compatible
stores use the store's expected value (MinIO accepts any).

### spec.database.postgresql.backup.objectStore.s3.endpointUrl

`string`

S3-COMPATIBLE ARM: endpoint URL of the store (e.g.
http://minio.minio-system.svc:9000 for in-cluster MinIO). Empty = real
AWS S3. For Cloudflare R2 prefer the `r2` arm, which composes the
endpoint from the bucket's account and jurisdiction.

- rule: endpoint_url must be an http(s) URL (e.g. http://minio.minio-system.svc:9000)

### spec.database.postgresql.backup.objectStore.s3.endpointCaPem

`string`

PEM CA bundle for verifying a self-signed endpoint_url TLS certificate
(materialized as a Secret the plugin reads).

### spec.database.postgresql.backup.objectStore.s3.keyless

`bool`

Keyless posture: the database pods' AWS identity (IRSA through
backup.service_account_annotations, EKS Pod Identity, or the node's
instance profile) authenticates to S3 — no stored keys. Mutually
exclusive with access_keys.

### spec.database.postgresql.backup.objectStore.s3.accessKeys

`KubernetesPlantonPlatformS3AccessKeys`

Static access keys, materialized as a Kubernetes Secret the plugin
reads. The declared-credential arm — for S3-compatible stores and
clusters without IRSA.

### spec.database.postgresql.backup.objectStore.s3.accessKeys.accessKeyId

`string` · required

Access key ID — the public identifier of the key pair, not a secret;
only the paired secret access key is a credential. For MinIO this is
the access key / username.

- rule: {"required":true}

### spec.database.postgresql.backup.objectStore.s3.accessKeys.secretAccessKey

`string` · required · sensitive

Secret access key (for MinIO: the secret key / password).

- rule: {"required":true}

### spec.database.postgresql.backup.objectStore.gcs

`KubernetesPlantonPlatformGcsObjectStore`

Google Cloud Storage.

- rule: keyless and service_account_key_json are alternative credential postures — set exactly one

### spec.database.postgresql.backup.objectStore.gcs.keyless

`bool`

Keyless posture: the database pods' GCP identity (GKE Workload
Identity through backup.service_account_annotations) authenticates to
GCS — no stored key. Mutually exclusive with service_account_key_json.

THE IDENTITY NEEDS TWO ROLES ON THE BUCKET, not one: Barman Cloud
verifies the archive destination with a bucket-level read
(`storage.buckets.get`) before every WAL archive, and
`roles/storage.objectAdmin` does not carry it — an identity granted
objectAdmin alone fails every archive with "does not have
storage.buckets.get access" while the database reports healthy. Grant
`roles/storage.objectAdmin` AND `roles/storage.legacyBucketReader` (a
GcpGcsBucket's `iam_members`, one entry each).

### spec.database.postgresql.backup.objectStore.gcs.serviceAccountKeyJson

`string` · sensitive

GCP service-account key (the JSON key file's content), materialized as
a Kubernetes Secret the plugin reads. The declared-credential arm for
non-GKE clusters backing up to GCS.

### spec.database.postgresql.backup.objectStore.azureBlob

`KubernetesPlantonPlatformAzureBlobObjectStore`

Azure Blob Storage.

- rule: keyless and connection_string are alternative credential postures — set exactly one

### spec.database.postgresql.backup.objectStore.azureBlob.storageAccount

`string` · required

Storage-account name. Always required: it identifies the storage
endpoint under either posture.

- rule: {"required":true}

### spec.database.postgresql.backup.objectStore.azureBlob.keyless

`bool`

Keyless posture: the database pods' Azure identity (AKS Workload
Identity through backup.service_account_annotations, or a managed
identity) authenticates to Blob Storage — no stored secret. Mutually
exclusive with connection_string.

### spec.database.postgresql.backup.objectStore.azureBlob.connectionString

`string` · sensitive

Storage-account connection string — the all-in-one declared
credential, materialized as a Kubernetes Secret the plugin reads.

### spec.database.postgresql.backup.objectStore.r2

`KubernetesPlantonPlatformR2ObjectStore`

Cloudflare R2, in R2's own vocabulary: the owning account, the
bucket's jurisdiction, and a Cloudflare credential — each a reference
onto the catalog's CloudflareR2Bucket and CloudflareAccountApiToken by
default. The operator performs the S3 translation R2 needs (the
jurisdiction's endpoint host, region `auto`); nothing S3-shaped is
typed here.

### spec.database.postgresql.backup.objectStore.r2.accountId

`string | valueFrom` · required

The Cloudflare account that owns the bucket (32 hex characters). By
reference to the bucket resource's `account_id` output, so the arm
follows the bucket; a literal names an account outside the catalog.

- references: CloudflareR2Bucket (`status.outputs.account_id`)
- rule: account_id is the 32-hex-character Cloudflare account id
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.account_id}} -- a bare string does not parse

### spec.database.postgresql.backup.objectStore.r2.jurisdiction

`string | valueFrom`

The bucket's data-residency jurisdiction: `default` (or empty), `eu`,
`fedramp`, or `us`. It selects the S3 host the operator composes — a
bucket created in a jurisdiction is unreachable through any other
host — so it must match the bucket exactly; by reference to the bucket
resource's `jurisdiction` output it cannot drift.

- references: CloudflareR2Bucket (`status.outputs.jurisdiction`)
- rule: jurisdiction must be one of "default", "eu", "fedramp", "us" (or empty for default)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.jurisdiction}} -- a bare string does not parse

### spec.database.postgresql.backup.objectStore.r2.credentials

`KubernetesPlantonPlatformR2Credentials` · required

The Cloudflare credential, as the S3 key pair R2's S3 API
authenticates. Materialized as a Kubernetes Secret the plugin reads;
never plaintext in the rendered resource.

- rule: {"required":true}

### spec.database.postgresql.backup.objectStore.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to the token
resource's `r2_access_key_id` output.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.database.postgresql.backup.objectStore.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value. By
reference to the token resource's `r2_secret_access_key` output.
Rotates with the token: a rotated token is a new key pair, and the
Secret this arm materializes follows the reference on the next apply.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.database.postgresql.backup.retentionPolicy

`string` · optional (explicit presence)

How long the store keeps base backups and the WAL that goes with
them, as `<n>d|w|m` (days, weeks, months). Enforced by the plugin
after each backup.

- default: `30d`
- rule: retention_policy must be a positive number of days, weeks, or months — e.g. '30d', '8w', or '6m'

### spec.database.postgresql.backup.schedule

`string` · optional (explicit presence)

When base backups run: a cron expression WITH SECONDS — six fields,
not the five Kubernetes CronJobs use ("0 0 2 * * *" is daily at
02:00 UTC, the default). The first base backup always runs the moment
backups are declared, whatever this says.

- default: `0 0 2 * * *`
- rule: schedule is a SIX-field cron expression (seconds first) — e.g. '0 0 2 * * *' for daily at 02:00; the five-field Kubernetes form is missing the seconds field

### spec.database.postgresql.backup.serviceAccountAnnotations

`map<string, string>`

Annotations for the ServiceAccount the database's own pods run as
(CloudNativePG names it after the database: `<platform>-postgres`).
This is where a keyless posture binds a cloud identity: EKS IRSA
(`eks.amazonaws.com/role-arn`), GKE Workload Identity
(`iam.gke.io/gcp-service-account`), AKS Workload Identity
(`azure.workload.identity/client-id`). Distinct from
runner.service_account_annotations — the runner deploys infrastructure
with its identity; the database's identity only writes backups. R2 has
no keyless posture and never needs this.

### spec.database.postgresql.recoverFrom

`KubernetesPlantonPlatformPostgresqlRecoverFrom`

Restore this platform's database from another platform's archive
instead of creating it empty. Honored only when the database is first
created — a running platform keeps its database as it is and the
status names the procedure (delete the platform, declare it again
with recover_from). Nothing here ever destroys data to honor a
declaration.

A restored platform archives its OWN backups under a new server name
(every platform files its archive under its name plus a unique
suffix), so it never writes over the source it restored from; declare
`backup` beside `recover_from` and the two can share one bucket and
one path.

WHAT COMES BACK: every record the control plane keeps — organizations,
environments, connections, projects, pipeline history, members, and
the identity realm with its users, so existing passwords sign in.
WHAT DOES NOT: the secrets manager's contents (OpenBAO keeps its data
on its own volume, outside this archive), so every secret value the
source held — the credentials behind connections above all — is
re-entered after a restore.

### spec.database.postgresql.recoverFrom.objectStore

`KubernetesPlantonPlatformObjectStore` · required

The store the source platform archived to. Recovery only READS from
it; this platform's own backups go to `backup`. The same bucket and
path as the source's `backup` are fine — server names keep the two
archives apart.

- rule: {"required":true}
- rule: the s3 backend stores at an s3:// destination path (also for S3-compatible stores like MinIO)
- rule: the gcs backend stores at a gs:// destination path
- rule: the azure_blob backend stores at an https:// destination path (https://<account>.blob.core.windows.net/<container>/<path>)
- rule: the r2 backend stores at an s3:// destination path (s3://<bucket>/<path> — R2 is addressed through its S3 API; the bucket name is the CloudflareR2Bucket's bucket_name)

### spec.database.postgresql.recoverFrom.objectStore.destinationPath

`string` · required

Where in the store the archive lives — the backend's native URI form:
`s3://bucket/path` for S3, Cloudflare R2, and every S3-compatible
store, `gs://bucket/path` for GCS, and
`https://<account>.blob.core.windows.net/<container>/<path>` for
Azure Blob. Base backups and WAL are filed beneath it under this
platform's server name, so several platforms can share one path
without ever touching each other's archive.

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.objectStore.s3

`KubernetesPlantonPlatformS3ObjectStore`

AWS S3 — or ANY S3-compatible store (MinIO, Ceph RGW, DigitalOcean
Spaces, ...) via the endpoint_url override. Cloudflare R2 has its own
arm (`r2`) that composes the catalog's Cloudflare kinds.

- rule: keyless and access_keys are alternative credential postures — set exactly one
- rule: an S3-compatible endpoint (endpoint_url) authenticates with access_keys — the keyless posture only mints AWS credentials

### spec.database.postgresql.recoverFrom.objectStore.s3.region

`string`

AWS region of the bucket. Required for real S3; for S3-compatible
stores use the store's expected value (MinIO accepts any).

### spec.database.postgresql.recoverFrom.objectStore.s3.endpointUrl

`string`

S3-COMPATIBLE ARM: endpoint URL of the store (e.g.
http://minio.minio-system.svc:9000 for in-cluster MinIO). Empty = real
AWS S3. For Cloudflare R2 prefer the `r2` arm, which composes the
endpoint from the bucket's account and jurisdiction.

- rule: endpoint_url must be an http(s) URL (e.g. http://minio.minio-system.svc:9000)

### spec.database.postgresql.recoverFrom.objectStore.s3.endpointCaPem

`string`

PEM CA bundle for verifying a self-signed endpoint_url TLS certificate
(materialized as a Secret the plugin reads).

### spec.database.postgresql.recoverFrom.objectStore.s3.keyless

`bool`

Keyless posture: the database pods' AWS identity (IRSA through
backup.service_account_annotations, EKS Pod Identity, or the node's
instance profile) authenticates to S3 — no stored keys. Mutually
exclusive with access_keys.

### spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys

`KubernetesPlantonPlatformS3AccessKeys`

Static access keys, materialized as a Kubernetes Secret the plugin
reads. The declared-credential arm — for S3-compatible stores and
clusters without IRSA.

### spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys.accessKeyId

`string` · required

Access key ID — the public identifier of the key pair, not a secret;
only the paired secret access key is a credential. For MinIO this is
the access key / username.

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.objectStore.s3.accessKeys.secretAccessKey

`string` · required · sensitive

Secret access key (for MinIO: the secret key / password).

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.objectStore.gcs

`KubernetesPlantonPlatformGcsObjectStore`

Google Cloud Storage.

- rule: keyless and service_account_key_json are alternative credential postures — set exactly one

### spec.database.postgresql.recoverFrom.objectStore.gcs.keyless

`bool`

Keyless posture: the database pods' GCP identity (GKE Workload
Identity through backup.service_account_annotations) authenticates to
GCS — no stored key. Mutually exclusive with service_account_key_json.

THE IDENTITY NEEDS TWO ROLES ON THE BUCKET, not one: Barman Cloud
verifies the archive destination with a bucket-level read
(`storage.buckets.get`) before every WAL archive, and
`roles/storage.objectAdmin` does not carry it — an identity granted
objectAdmin alone fails every archive with "does not have
storage.buckets.get access" while the database reports healthy. Grant
`roles/storage.objectAdmin` AND `roles/storage.legacyBucketReader` (a
GcpGcsBucket's `iam_members`, one entry each).

### spec.database.postgresql.recoverFrom.objectStore.gcs.serviceAccountKeyJson

`string` · sensitive

GCP service-account key (the JSON key file's content), materialized as
a Kubernetes Secret the plugin reads. The declared-credential arm for
non-GKE clusters backing up to GCS.

### spec.database.postgresql.recoverFrom.objectStore.azureBlob

`KubernetesPlantonPlatformAzureBlobObjectStore`

Azure Blob Storage.

- rule: keyless and connection_string are alternative credential postures — set exactly one

### spec.database.postgresql.recoverFrom.objectStore.azureBlob.storageAccount

`string` · required

Storage-account name. Always required: it identifies the storage
endpoint under either posture.

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.objectStore.azureBlob.keyless

`bool`

Keyless posture: the database pods' Azure identity (AKS Workload
Identity through backup.service_account_annotations, or a managed
identity) authenticates to Blob Storage — no stored secret. Mutually
exclusive with connection_string.

### spec.database.postgresql.recoverFrom.objectStore.azureBlob.connectionString

`string` · sensitive

Storage-account connection string — the all-in-one declared
credential, materialized as a Kubernetes Secret the plugin reads.

### spec.database.postgresql.recoverFrom.objectStore.r2

`KubernetesPlantonPlatformR2ObjectStore`

Cloudflare R2, in R2's own vocabulary: the owning account, the
bucket's jurisdiction, and a Cloudflare credential — each a reference
onto the catalog's CloudflareR2Bucket and CloudflareAccountApiToken by
default. The operator performs the S3 translation R2 needs (the
jurisdiction's endpoint host, region `auto`); nothing S3-shaped is
typed here.

### spec.database.postgresql.recoverFrom.objectStore.r2.accountId

`string | valueFrom` · required

The Cloudflare account that owns the bucket (32 hex characters). By
reference to the bucket resource's `account_id` output, so the arm
follows the bucket; a literal names an account outside the catalog.

- references: CloudflareR2Bucket (`status.outputs.account_id`)
- rule: account_id is the 32-hex-character Cloudflare account id
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.account_id}} -- a bare string does not parse

### spec.database.postgresql.recoverFrom.objectStore.r2.jurisdiction

`string | valueFrom`

The bucket's data-residency jurisdiction: `default` (or empty), `eu`,
`fedramp`, or `us`. It selects the S3 host the operator composes — a
bucket created in a jurisdiction is unreachable through any other
host — so it must match the bucket exactly; by reference to the bucket
resource's `jurisdiction` output it cannot drift.

- references: CloudflareR2Bucket (`status.outputs.jurisdiction`)
- rule: jurisdiction must be one of "default", "eu", "fedramp", "us" (or empty for default)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.jurisdiction}} -- a bare string does not parse

### spec.database.postgresql.recoverFrom.objectStore.r2.credentials

`KubernetesPlantonPlatformR2Credentials` · required

The Cloudflare credential, as the S3 key pair R2's S3 API
authenticates. Materialized as a Kubernetes Secret the plugin reads;
never plaintext in the rendered resource.

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.objectStore.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to the token
resource's `r2_access_key_id` output.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.database.postgresql.recoverFrom.objectStore.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value. By
reference to the token resource's `r2_secret_access_key` output.
Rotates with the token: a rotated token is a new key pair, and the
Secret this arm materializes follows the reference on the next apply.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.database.postgresql.recoverFrom.serverName

`string` · required

The name the source's archive is filed under: `status.backup.serverName`
on the source platform, or the folder name in the bucket's own listing
under the destination path when the source is gone. Every platform
archives under its own name plus a unique suffix, so this names one
archive exactly.

- rule: {"required":true}

### spec.database.postgresql.recoverFrom.targetTime

`string`

Recover to a point in time, as an RFC 3339 timestamp
("2026-09-13T20:30:00Z"). Empty recovers to the end of the archive —
every WAL segment the source shipped.

- rule: target_time is an RFC 3339 timestamp — e.g. '2026-09-13T20:30:00Z' or '2026-09-13T20:30:00+05:30'

### spec.database.redis

`KubernetesPlantonPlatformRedis`

The Redis-protocol cache (Valkey).

### spec.database.redis.storageSize

`string`

Volume size (e.g. "1Gi"). Falls back to spec.storage.size, then the
platform default.

- rule: storage_size must be a Kubernetes quantity like "1Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.database.redis.storageClassName

`string`

StorageClass override for the cache volume.

### spec.ingress

`KubernetesPlantonPlatformIngress`

Expose the platform at a real URL through the cluster's existing
front door — an Ingress controller, or a Gateway API Gateway
(`gateway_ref`). Off by default — the built-in gateway plus
port-forward is the zero-config door. `hostname` serves your own
domain; `tls` adds HTTPS (bring a certificate Secret or name a
cert-manager issuer).

- rule: tls requires hostname: a certificate cannot be brought or issued for an auto-derived hostname
- rule: gateway_ref and ingress_class_name name two different front doors; set one — gateway_ref attaches to a Gateway API Gateway, ingress_class_name renders an Ingress
- rule: with gateway_ref the Gateway's HTTPS listener owns the certificate: attach to a listener that already serves the hostname, or set tls.issuer to have a certificate issued for the listener to reference
- rule: reachability: public declares an address the internet reaches, but with enabled: false the platform is reached only through kubectl port-forward from the machine running it; set enabled: true, or leave reachability at auto

### spec.ingress.enabled

`bool`

Expose the platform through the cluster's front door. With no
hostname, the operator derives a working URL from the front door's
published address (magic DNS) — the Ingress controller's, or the
Gateway's.

### spec.ingress.hostname

`string`

The platform's hostname (e.g. "planton.example.com"). One hostname
serves console AND API. The identity server bakes this URL into its
realm at first boot — set it before the first sign-in.

### spec.ingress.ingressClassName

`string`

IngressClass to use. Unset = the cluster's default. Names one front
door; never combined with gateway_ref.

### spec.ingress.annotations

`map<string, string>`

Extra annotations on the Ingress (controller-specific tuning —
ALB schemes, proxy budgets). Ignored with gateway_ref: the Gateway
API expresses behavior in typed fields, not annotations.

### spec.ingress.tls

`KubernetesPlantonPlatformIngressTls`

HTTPS. Requires hostname. Exactly one of secret_name (bring your own
certificate) or issuer (cert-manager issues one) — except with
gateway_ref, where the route attaches to whichever listener matches
the hostname and HTTPS is inferred from that listener; only issuer
applies there.

- rule: exactly one of secret_name or issuer must be set — bring a certificate or have cert-manager issue one, never both

### spec.ingress.tls.secretName

`string`

An existing kubernetes.io/tls Secret in the platform's namespace.

### spec.ingress.tls.issuer

`KubernetesPlantonPlatformCertManagerIssuer`

A cert-manager issuer to obtain the certificate from. Requires
cert-manager on the cluster.

### spec.ingress.tls.issuer.name

`string` · required

Issuer name.

- rule: {"string":{"minLen":"1"}}

### spec.ingress.tls.issuer.kind

`string` · optional (explicit presence)

Issuer kind.

- default: `Issuer`
- rule: {"string":{"in":["","Issuer","ClusterIssuer"]}}

### spec.ingress.gatewayRef

`KubernetesPlantonPlatformGatewayRef`

Serve Planton through a Gateway API Gateway the cluster already runs
(Istio, Envoy Gateway, Cilium, a cloud Gateway) instead of an Ingress
controller: the operator attaches an HTTPRoute for the hostname to
that Gateway. The Gateway is the cluster team's object and is never
modified — its listeners decide which hostnames are admitted, which
namespaces may attach routes, and how HTTPS is terminated; the
operator reads those facts and explains any mismatch in the
platform's status. Requires a planton-operator chart that knows this
field (0.9.0 or newer); an older definition refuses the declaration.

### spec.ingress.gatewayRef.name

`string | valueFrom` · required

Name of the Gateway. Defaults to a KubernetesGateway foreign key
(`status.outputs.gateway_name`): `valueFrom` for a Planton-managed
Gateway, `value:` for one created outside Planton.

- references: KubernetesGateway (`status.outputs.gateway_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_name}} -- a bare string does not parse

### spec.ingress.gatewayRef.namespace

`string | valueFrom`

Namespace of the Gateway. Defaults to the platform's own namespace when
omitted. A Gateway in another namespace must allow routes from this one
(spec.listeners[].allowedRoutes.namespaces). Defaults to the same
KubernetesGateway foreign key as `name` (`status.outputs.namespace`), so
one resource wires both; a Gateway created outside Planton takes the
literal namespace with `value:`.

- references: KubernetesGateway (`status.outputs.namespace`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.namespace}} -- a bare string does not parse

### spec.ingress.gatewayRef.sectionName

`string`

Pins the route to one named listener of the Gateway. When empty, the
route attaches to every listener whose hostname admits the
platform's hostname.

### spec.ingress.reachability

`string` · optional (explicit presence)

Whether the public internet can reach this front door — the one fact
about the door the operator cannot observe from inside the cluster.
The capabilities that need an inbound path from the internet (keyless
cloud connections, where the cloud fetches the issuer's discovery
document; GitHub webhook delivery) are offered only where the door is
public. `auto` (default) resolves from the door's shape: a hostname
served over HTTPS is public, anything else private. Declare `private`
for an HTTPS door only your network reaches (split DNS, a corporate
CA, an internal load balancer); declare `public` to affirm it. Only
`public` is refused when enabled is false — a port-forward door is
never reached from the internet — while `private` there is simply
true. Requires a planton-operator chart that knows this field (0.11.0
or newer); an older definition refuses the declaration.

- default: `auto`
- rule: {"string":{"in":["","auto","public","private"]}}

### spec.gateway

`KubernetesPlantonPlatformGateway`

The built-in front-door gateway: console, API, and sign-in on one
origin over a single `kubectl port-forward`. Always deployed; the
local_port here is baked into the identity server's issuer and the
console's callbacks at first boot, so pick it before the first visit.

### spec.gateway.localPort

`int32` · optional (explicit presence)

The local port the port-forward door advertises
(`kubectl port-forward ... {local_port}:80`). Baked into the identity
server's issuer and the console's callbacks at first boot — pick it
before the first visit; two port-forward platforms on one machine
need distinct ports.

- default: `8080`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.identity

`KubernetesPlantonPlatformIdentity`

The platform's identity server (Keycloak, operator-deployed).
Everything defaults: the realm is "planton" and the first console
visitor becomes the admin through the setup-code page. Set
admin_email only to pre-seed a known admin instead.

### spec.identity.realm

`string` · optional (explicit presence)

The Keycloak realm name.

- default: `planton`

### spec.identity.adminEmail

`string`

Pre-seed a known admin account instead of the first-visitor setup
page: the operator creates this user and writes a one-time password
into the platform's admin-user Secret. Unset (the default) leaves
the setup-code flow — the first console visitor becomes the admin.

- rule: admin_email must be an email address like "admin@example.com"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.bootstrap

`KubernetesPlantonPlatformBootstrap`

First-boot seeding: the initial organization and environment, extra
admins, the IaC provisioner the in-cluster runner uses, and where the
platform's managed secrets live.

### spec.bootstrap.organization

`KubernetesPlantonPlatformBootstrapOrg`

The initial organization.

### spec.bootstrap.organization.slug

`string` · optional (explicit presence)

Organization slug.

- default: `default`

### spec.bootstrap.organization.name

`string`

Display name. Defaults to the slug.

### spec.bootstrap.environment

`KubernetesPlantonPlatformBootstrapEnv`

The initial environment.

### spec.bootstrap.environment.slug

`string` · optional (explicit presence)

Environment slug.

- default: `default`

### spec.bootstrap.environment.name

`string`

Display name. Defaults to the slug.

### spec.bootstrap.admins

`[]string`

Additional admin emails granted org ownership and platform-operator
rights at boot (the identity.admin_email, when set, is always
included).

### spec.bootstrap.iacProvisioner

`string` · optional (explicit presence)

The IaC provisioner the in-cluster runner deploys with.

- default: `tofu`
- rule: {"string":{"in":["","tofu","terraform"]}}

### spec.bootstrap.secretBackend

`KubernetesPlantonPlatformSecretBackend`

Where the platform's managed secrets live: the bundled secrets
manager ("platform", the default) or a cloud backend.

- rule: awsSecretsManager needs its configuration block: aws_secrets_manager.region and aws_secrets_manager.kms_key_arn are required

### spec.bootstrap.secretBackend.type

`string` · required

Backend type: "platform" (the bundled OpenBAO) or
"awsSecretsManager".

- rule: {"required":true,"string":{"in":["platform","awsSecretsManager"]}}

### spec.bootstrap.secretBackend.awsSecretsManager

`KubernetesPlantonPlatformAwsSecretsManager`

AWS Secrets Manager configuration (required when type is
awsSecretsManager). The control plane reaches AWS through its own
workload identity — see control_plane.service_account_annotations.

### spec.bootstrap.secretBackend.awsSecretsManager.region

`string` · required

AWS region (e.g. "us-east-1").

- rule: {"string":{"minLen":"1"}}

### spec.bootstrap.secretBackend.awsSecretsManager.kmsKeyArn

`string` · required

KMS key ARN encrypting the secrets.

- rule: {"string":{"minLen":"1"}}

### spec.runner

`KubernetesPlantonPlatformRunner`

The in-cluster deployment runner — how this platform deploys real
infrastructure. ON by default; disabling it leaves a platform that
can model but not deploy. Cloud identity comes from workload-identity
annotations OR a customer-owned credentials Secret — the platform
stores no cloud credentials itself.

### spec.runner.enabled

`bool` · optional (explicit presence)

Deploy the runner. Platform default: true. Explicit false leaves a
platform that can model infrastructure but not deploy it.

- default: `true`

### spec.runner.storageSize

`string`

Runner state volume size (e.g. "2Gi"). Falls back to
spec.storage.size, then the platform default.

- rule: storage_size must be a Kubernetes quantity like "2Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.runner.storageClassName

`string`

StorageClass override for the runner state volume.

### spec.runner.serviceAccountAnnotations

`map<string, string>`

Workload-identity annotations on the runner's ServiceAccount — the
keyless way the runner reaches your cloud (EKS:
"eks.amazonaws.com/role-arn"; GKE: "iam.gke.io/gcp-service-account";
AKS: "azure.workload.identity/client-id").

### spec.runner.cloudCredentialsSecretName

`string`

Name of a customer-owned Secret (in the platform's namespace) whose
keys are injected into the runner as environment variables — the
static-credentials way the runner reaches your cloud. The platform
stores nothing: rotate by updating YOUR Secret.

### spec.build

`KubernetesPlantonPlatformBuild`

Container-image build pipelines (Tekton). ON by default; follows the
runner off when the runner is disabled. NOTE: Tekton allows exactly
one cluster-wide build-events sink, so builds can feed only ONE
build-enabled platform per cluster — disable this on all but one when
several platforms share a cluster.

### spec.build.enabled

`bool` · optional (explicit presence)

Enable Tekton-backed build pipelines. Platform default: true
(follows the runner off when the runner is disabled). One
build-enabled platform per cluster — Tekton's build-events sink is
cluster-wide.

- default: `true`

### spec.vault

`KubernetesPlantonPlatformVault`

The bundled secrets manager (OpenBAO). ON by default — a
version-only platform stores connection secrets with zero
configuration. Explicit `enabled: false` is the deliberate opt-out
(bring a cloud secret backend through bootstrap.secret_backend
instead).

### spec.vault.enabled

`bool` · optional (explicit presence)

Deploy the bundled secrets manager (OpenBAO). Platform default:
true. Explicit false is the deliberate opt-out — pair it with a
cloud backend in bootstrap.secret_backend or connection secrets have
nowhere to live.

- default: `true`

### spec.vault.initMode

`string` · optional (explicit presence)

Initialization: "auto" (the operator initializes and unseals,
storing the unseal keys in an annotated platform Secret) or "manual"
(you run the init ceremony).

- default: `auto`
- rule: {"string":{"in":["","auto","manual"]}}

### spec.vault.storageSize

`string`

Volume size (e.g. "2Gi"). Falls back to spec.storage.size, then the
platform default.

- rule: storage_size must be a Kubernetes quantity like "2Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.vault.storageClassName

`string`

StorageClass override for the secrets-manager volume.

### spec.components

`KubernetesPlantonPlatformComponents`

Opt-in platform components, off by default: the graph explorer
(Neo4j).

### spec.components.graph

`KubernetesPlantonPlatformGraph`

The graph explorer (Neo4j).

### spec.components.graph.enabled

`bool`

Enable the graph explorer.

### spec.components.graph.storageSize

`string`

Volume size (e.g. "10Gi"). Falls back to spec.storage.size, then the
platform default.

- rule: storage_size must be a Kubernetes quantity like "10Gi"
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.components.graph.storageClassName

`string`

StorageClass override for the graph volume.

### spec.prerequisites

`KubernetesPlantonPlatformPrerequisites`

Cluster-shared sub-operators the platform rides (CloudNativePG,
Tekton Pipelines). "auto" (the default) installs each one only when
no installation exists on the cluster; "skip" declares that something
else manages it.

### spec.prerequisites.postgresOperator

`string` · optional (explicit presence)

CloudNativePG: "auto" installs it only when absent; "skip" declares
it externally managed (a helm/GitOps CloudNativePG is respected
automatically either way).

- default: `auto`
- rule: {"string":{"in":["","auto","skip"]}}

### spec.prerequisites.tektonPipelines

`string` · optional (explicit presence)

Tekton Pipelines (build pipelines). The build-events sink is wired
even on "skip".

- default: `auto`
- rule: {"string":{"in":["","auto","skip"]}}

### spec.prerequisites.postgresBackupPlugin

`string` · optional (explicit presence)

The Barman Cloud plugin, CloudNativePG's backup engine, which serves
every PostgreSQL on the cluster — the platform's own database and any
database deployed through Planton. "auto" installs it whenever the
operator installed CloudNativePG itself and cert-manager is on the
cluster (the plugin needs it for the TLS between operator and plugin),
or whenever database.postgresql.backup is declared; a plugin installed
by any other means is detected and respected. "skip" declares it
externally managed (or unwanted) and installs nothing.

- default: `auto`
- rule: {"string":{"in":["","auto","skip"]}}

### spec.controlPlane

`KubernetesPlantonPlatformControlPlane`

The platform's control-plane deployment (the API monolith). Sizing,
image mirrors, extra environment through a Secret, and the
platform's OWN cloud identity (for cloud secret backends and KMS) —
distinct from the runner's deploy-time identity.

### spec.controlPlane.image

`KubernetesPlantonPlatformImage`

Image override (air-gapped mirrors of the SAME version).

### spec.controlPlane.image.repository

`string`

Full image repository (e.g.
"my-mirror.example.com/planton/control-plane").

### spec.controlPlane.image.tag

`string`

Image tag. Empty = spec.version.

### spec.controlPlane.replicas

`int32` · optional (explicit presence)

Control-plane replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.controlPlane.externalConfigSecretName

`string`

Name of a Secret (in the platform's namespace) whose keys are all
injected into the control plane as environment variables — the
escape hatch for configuration the spec does not model.

### spec.controlPlane.serviceAccountAnnotations

`map<string, string>`

Workload-identity annotations on the control plane's ServiceAccount
— the platform's OWN cloud identity (cloud secret backends, KMS).
Distinct from runner.service_account_annotations, which is the
DEPLOY-TIME identity.

### spec.console

`KubernetesPlantonPlatformConsole`

The web console deployment. Sizing, image mirrors, and extra
environment through a Secret.

### spec.console.image

`KubernetesPlantonPlatformImage`

Image override (air-gapped mirrors of the SAME version).

### spec.console.image.repository

`string`

Full image repository (e.g.
"my-mirror.example.com/planton/control-plane").

### spec.console.image.tag

`string`

Image tag. Empty = spec.version.

### spec.console.replicas

`int32` · optional (explicit presence)

Console replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.console.externalConfigSecretName

`string`

Name of a Secret (in the platform's namespace) whose keys are all
injected into the console as environment variables.

### spec.remoteRunners

`KubernetesPlantonPlatformRemoteRunners`

Runners outside this cluster — a developer's laptop deploying with the
cloud sign-in already on it, an appliance in another network — pulling
this platform's deploy work. OFF by default. Rides the front door:
the deploy queue is routed through the platform hostname beside the
native gRPC API, so it needs a Gateway API front door (ingress with a
gateway_ref); on any other door the capability stays closed and the
platform's status says why. The in-cluster runner is unaffected.

### spec.remoteRunners.enabled

`bool` · optional (explicit presence)

Open the deploy queue to runners outside the cluster and advertise the
front door's address to them. Platform default: false — an install that
has not chosen this keeps its queue in-cluster, and a runner asking to
enroll from outside is refused with the reason, never handed an address
it cannot reach. What opens: the queue's workflow service, over TLS,
without authentication of its own (the posture the hosted platform
carries for its remote runners); the queue's administrative service
never leaves the cluster.

- default: `false`

### spec.email

`KubernetesPlantonPlatformEmail`

The one mail provider every sender on the install uses: the control
plane (invitations, alerts, license mail) and the identity server
(password resets) both send through it, so one declaration is the
whole configuration. Absent, the install sends no email: invitations
are shared as links and the sign-in page offers no "Forgot password?".
Exactly one provider arm is set — an SMTP relay or a Resend account.
Credentials are never inline: they are Secrets in the platform's
namespace, named here, and reach the control plane as mounted files so
a rotated password is live on the next send. The operator delivers the
declaration and preflights every Secret it names; the control plane
checks the relay on demand from the console's Email settings and
reports each failure in the relay's own words. Requires a
planton-operator chart that knows this field (0.14.1 or newer); an
older definition refuses the declaration.

- rule: email declares exactly one provider: set spec.email.smtp for a relay or spec.email.resend for a Resend account, never both, never neither

### spec.email.from

`KubernetesPlantonPlatformEmailFrom` · required

The identity every email carries: the address the install sends as
and the display name beside it. The relay must permit sending as this
address (a mailbox's own address, or one it has Send As rights to);
SPF and DKIM for the domain are the domain owner's job.

- rule: {"required":true}

### spec.email.from.address

`string` · required

The address the install sends as, e.g. no-reply@planton.acme.com.
Required whenever email is declared: there is no default address,
because a default would name somebody else's domain.

- rule: {"string":{"minLen":"1"}}

### spec.email.from.name

`string` · optional (explicit presence)

The name shown beside the address in mail clients. Platform default:
Planton.

- default: `Planton`

### spec.email.replyTo

`string`

Where a person's reply lands — a help desk or a shared mailbox — when
the sending address is a no-reply one. Empty, replies go to
from.address.

### spec.email.smtp

`KubernetesPlantonPlatformEmailSmtp`

Send through any SMTP relay: a workplace mail system (Exchange Online,
Google Workspace, an internal smart host) or a transactional vendor's
SMTP endpoint (SES, SendGrid, Postmark, Mailgun, Resend).

- rule: smtp authenticates one way: set credentials_secret_name for a username and password, or oauth2 for a token, not both
- rule: security: none would send credentials in the clear; keep security at starttls or tls, or drop credentials_secret_name and oauth2 for a relay that admits this cluster's address without them

### spec.email.smtp.host

`string` · required

Host of the relay, e.g. smtp.office365.com or
smtp-relay.corp.acme.com.

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.port

`int32` · optional (explicit presence)

Port the relay listens on. 587 is the submission port most relays use
with STARTTLS; implicit-TLS relays (security: tls) usually listen on
465; an internal plaintext relay on 25. Platform default: 587.

- default: `587`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.email.smtp.security

`string` · optional (explicit presence)

How the connection to the relay is protected: starttls (connect in the
clear and REQUIRE the upgrade before anything is sent — a relay that
does not offer it is a failed connection, never a silent fallback; the
default), tls (implicit TLS from the first byte), or none (plaintext
end to end, for credential-free internal relays only; credentials are
refused on it). Platform default: starttls.

- default: `starttls`
- rule: {"string":{"in":["","starttls","tls","none"]}}

### spec.email.smtp.credentialsSecretName

`string`

Name of a kubernetes.io/basic-auth Secret in the platform's namespace
whose username and password keys sign in to the relay:

  kubectl -n <namespace> create secret generic planton-email \
    --type=kubernetes.io/basic-auth \
    --from-literal=username=... --from-literal=password=...

Omit it for a relay that admits this cluster by network address. The
values reach the control plane as mounted files, so a rotated password
is live on the next send with no restart.

### spec.email.smtp.oauth2

`KubernetesPlantonPlatformEmailSmtpOauth2`

Sign in with a token from an OAuth2 client-credentials grant (SASL
XOAUTH2) instead of a password. Exchange Online: token_url
https://login.microsoftonline.com/<tenant>/oauth2/v2.0/token, scope
https://outlook.office365.com/.default, the app registration's client
id and secret, and user = the mailbox the app may send as.

### spec.email.smtp.oauth2.user

`string` · required

The mailbox the token sends as — the account the app registration has
been permitted to use.

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.oauth2.tokenUrl

`string` · required

The provider's OAuth2 token endpoint. Must be an https:// URL.

- rule: token_url must be an https:// URL
- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.oauth2.scope

`string` · required

The scope requested for the token.

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.oauth2.clientId

`string` · required

The app registration's client id.

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.oauth2.clientSecretRef

`KubernetesPlantonPlatformSecretKeyRef` · required

The app registration's client secret, by reference: the secret is
never inline.

- rule: {"required":true}

### spec.email.smtp.oauth2.clientSecretRef.name

`string` · required

Secret name (in the platform's namespace).

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.oauth2.clientSecretRef.key

`string` · required

Key within the Secret holding the value.

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.caBundleSecretRef

`KubernetesPlantonPlatformSecretKeyRef`

A PEM CA bundle for verifying the relay's TLS certificate — the
private-CA case, the classic enterprise blocker. Omit it when the
relay's certificate chains to a public root.

### spec.email.smtp.caBundleSecretRef.name

`string` · required

Secret name (in the platform's namespace).

- rule: {"string":{"minLen":"1"}}

### spec.email.smtp.caBundleSecretRef.key

`string` · required

Key within the Secret holding the value.

- rule: {"string":{"minLen":"1"}}

### spec.email.resend

`KubernetesPlantonPlatformEmailResend`

Send through Resend's API with an API key.

### spec.email.resend.apiKeySecretRef

`KubernetesPlantonPlatformSecretKeyRef` · required

The Resend API key, by reference: the key is never inline. Reaches the
control plane as a mounted file, so a rotated key is live on the next
send with no restart.

- rule: {"required":true}

### spec.email.resend.apiKeySecretRef.name

`string` · required

Secret name (in the platform's namespace).

- rule: {"string":{"minLen":"1"}}

### spec.email.resend.apiKeySecretRef.key

`string` · required

Key within the Secret holding the value.

- rule: {"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPlantonPlatform, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the platform lives in. |
| `status.outputs.platform_name` | `string` | The platform name (the PlantonPlatform CR name — the prefix of every object the operator creates for this platform). |
| `status.outputs.gateway_service` | `string` | The built-in front-door gateway Service ("{platform_name}-gateway") — console, API, and sign-in on one origin. |
| `status.outputs.setup_code_secret` | `string` | The Secret holding the first-run setup code ("{platform_name}-identity-setup-code") — the console's setup page asks for this code when the first visitor becomes the admin. |
| `status.outputs.port_forward_command` | `string` | The exact command that opens the platform's door on this machine. |
| `status.outputs.setup_code_command` | `string` | The exact command that reads the first-run setup code. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.database.postgresql.backup.objectStore.r2.accountId` | CloudflareR2Bucket | `status.outputs.account_id` |
| `spec.database.postgresql.backup.objectStore.r2.jurisdiction` | CloudflareR2Bucket | `status.outputs.jurisdiction` |
| `spec.database.postgresql.backup.objectStore.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.database.postgresql.backup.objectStore.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.database.postgresql.recoverFrom.objectStore.r2.accountId` | CloudflareR2Bucket | `status.outputs.account_id` |
| `spec.database.postgresql.recoverFrom.objectStore.r2.jurisdiction` | CloudflareR2Bucket | `status.outputs.jurisdiction` |
| `spec.database.postgresql.recoverFrom.objectStore.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.database.postgresql.recoverFrom.objectStore.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.ingress.gatewayRef.name` | KubernetesGateway | `status.outputs.gateway_name` |
| `spec.ingress.gatewayRef.namespace` | KubernetesGateway | `status.outputs.namespace` |

## See Also

- [Overview](../README.md)

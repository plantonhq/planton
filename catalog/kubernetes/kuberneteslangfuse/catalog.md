# Langfuse

Deploys Langfuse, the open-source LLM engineering platform for tracing, prompt management, evaluations and datasets, on Kubernetes from the official `langfuse` Helm chart. One release runs the web server (UI, public API and the OTLP ingestion endpoint) and the worker (queue consumers that load events into ClickHouse and run evaluations, exports and retention) against four stores you compose by reference: Postgres, ClickHouse, a Redis-protocol store and object storage. The chart's bundled copies of those stores never ship from this kind. Every secret reaches Langfuse through a module-owned Secret, never through Helm values, and by default the install comes up with a seeded organization, project and API key pair, so an SDK can send its first trace the moment the release is ready.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise Langfuse installs into an existing namespace, which is the usual shape because the stores' Secrets must live there too
- **The instance Secret** (`<name>-secrets`) -- every sensitive value the spec declares (store connection URLs, the SMTP URL, single sign-on client secrets, blob storage keys, the AI provider key, the license key) plus what the module generates when the spec leaves it empty: the API-key salt, the 64-hex-character encryption key, the NextAuth session secret, the seeded project key pair and the seeded admin's password
- **The Helm release** -- the official `langfuse` chart with its four bundled stores pinned off, rendering:
  - The web Deployment and Service (`<name>-web`, port 3000) -- the UI, the public API, the ingestion API and the OTLP endpoint; Postgres and ClickHouse migrations run when it starts
  - The worker Deployment (`<name>-worker`) -- queue consumers, evaluations, batch exports and data retention
  - The ServiceAccount (`<name>`), unless `serviceAccount.existingName` names one -- annotated for EKS, GKE or AKS workload identity when `serviceAccount.workloadIdentity` is set
  - Per workload, only when declared: a HorizontalPodAutoscaler, a KEDA ScaledObject, a VerticalPodAutoscaler and a PodDisruptionBudget
  - The chart's own `<name>-app` Secret -- created empty, because the module supplies the salt, encryption key and session secret itself
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or name it explicitly when creating the Infra Component.
- **Planton Runner** -- required when credentials are delivered through a Runner; not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Kubernetes 1.28 or newer** -- the floor chart 2.x declares.
- **The four stores, in Langfuse's namespace** -- a Kubernetes Secret reference cannot cross namespaces, and every store password reaches the pods through one. Compose the stores into the same namespace (the `self-hosted-langfuse` InfraChart does), or copy their Secrets there.
- **A Postgres database Langfuse owns** -- PostgreSQL 15 or newer in UTC. On a composed PostgreSQL, declare the `langfuse` database at bootstrap (`initdb.database`) with the Langfuse user as owner: migrations run as that user on every start.
- **A ClickHouse user that can reach the server** -- ClickHouse 25.12 or newer. On a composed ClickHouse, declare the user's `networks` (for example `0.0.0.0/0` and `::/0`): without them the user is fenced to ClickHouse's own pods and Langfuse is refused with what looks like a wrong-password error. Cluster mode (`clickhouse.cluster.enabled`, on by default) needs ClickHouse Keeper.
- **A Redis-protocol store with `maxmemory-policy noeviction`** -- under any eviction policy, queued ingestion jobs are silently dropped.
- **A bucket that already exists** -- every ingested event is written to object storage before the worker reads it.
- **KEDA** (only for `autoscaling.keda`) and **the VerticalPodAutoscaler** (only for `autoscaling.vpa`) installed in the cluster.
- **A name within budget** -- keep `metadata.name` at 52 characters or fewer: the chart's longest child name is `<name>-worker-pdb`. Both modules fail before applying when the name is longer.

## Deploy

### Console

Open the deployment store, find **Langfuse**, and click **Deploy**. The creation wizard walks you through the four stores, sign-in and the workloads. Start from the **Evaluation** preset for a single-server install against in-cluster stores, or **Production In-Cluster** for replicated workloads, durable secrets and a seeded admin, in the [Presets](#presets) tab. The **AWS**, **Google Cloud** and **Azure** presets put events in the cloud's own object storage.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLangfuse
metadata:
  name: langfuse
  org: acme-corp
  env: prod
spec:
  namespace:
    value: langfuse
  postgres:
    host:
      value: langfuse-pg-rw.langfuse.svc.cluster.local
    username: langfuse
    passwordSecret:
      secretName:
        value: langfuse-pg-app
  clickhouse:
    host:
      value: clickhouse-langfuse-ch.langfuse.svc.cluster.local
    username: langfuse
    passwordSecret:
      secretName:
        value: langfuse-ch-clickhouse-auth
    cluster:
      enabled: false
  redis:
    standalone:
      host:
        value: langfuse-valkey.langfuse.svc.cluster.local
    passwordSecret:
      secretName:
        value: langfuse-valkey-auth
      secretKey:
        value: default
  blobStorage:
    target:
      s3Compatible:
        endpoint:
          value: http://langfuse-s3-s3.langfuse.svc.cluster.local:8333
        bucket: langfuse
        credentialsSecret:
          secretName:
            value: langfuse-s3-s3-secret
```

```shell
planton apply -f langfuse.yaml
```

This installs Langfuse into the existing `langfuse` namespace against four stores already running there, with a single ClickHouse server (cluster mode off). The salt, encryption key and session secret are generated, and Langfuse seeds an organization (ID `langfuse`) and a project (ID `langfuse-project`) with a generated key pair. Open the UI with the `port_forward_command` output and read the project's secret key from the `langfuse-secrets` Secret. An Infra Job tracks the provisioning in real time.

### InfraChart

When Langfuse is part of an InfraChart, wire each store by reference. Every reference names only the resource; the kind and field path come from the field's composition default:

```yaml
spec:
  namespace:
    value: langfuse
  postgres:
    host:
      valueFrom:
        name: langfuse-pg
    username: langfuse
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-pg
  clickhouse:
    host:
      valueFrom:
        name: langfuse-ch
    username: langfuse
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-ch
    cluster:
      name:
        valueFrom:
          name: langfuse-ch
  redis:
    standalone:
      host:
        valueFrom:
          name: langfuse-valkey
    username:
      valueFrom:
        name: langfuse-valkey
    passwordSecret:
      secretName:
        valueFrom:
          name: langfuse-valkey
      secretKey:
        valueFrom:
          name: langfuse-valkey
  blobStorage:
    target:
      s3Compatible:
        endpoint:
          valueFrom:
            name: langfuse-s3
        bucket: langfuse
        credentialsSecret:
          secretName:
            valueFrom:
              name: langfuse-s3
```

The InfraPipeline deploys the four stores first and Langfuse after them: each host resolves to the store's Service and each password to the store's own Secret, which Langfuse reads by reference at runtime and never copies.

## Key Configuration

These are the most important decisions when configuring Langfuse. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Declare the salt and encryption key for any install you intend to keep.** Langfuse hashes API keys with `secrets.salt` and encrypts stored LLM credentials with `secrets.encryptionKey`, and both results live in Postgres, not in this resource. Generated values survive every apply, but deleting and recreating the resource against a surviving database generates new ones, and every existing API key and stored credential stops working with no way back. Declare both as `$secret/` references from the first deploy, or copy the generated values out of the instance Secret and declare them before any recreate.

**Decide ClickHouse cluster mode before the first start.** With `clickhouse.cluster.enabled` (the default), Langfuse creates its tables `ON CLUSTER` with replicated engines, which needs ClickHouse Keeper and a user allowed to run `ON CLUSTER` statements. A single server without Keeper needs `enabled: false`, and the migrations fail otherwise. Switching after data exists means migrating that data.

**The Redis-protocol store must never evict.** Langfuse's ingestion queues live there, and under an eviction policy queued events disappear without an error. Run the store with `maxmemory-policy noeviction` and size its memory for the queue backlog you can tolerate.

**Media and batch exports travel through presigned URLs.** Browsers and SDKs upload and download media, and people download exports, straight from object storage. An in-cluster store is unreachable from outside the cluster, so set `blobStorage.mediaUpload.publicEndpoint` (and `blobStorage.batchExport.publicEndpoint`) to the address a route serves the store at, or give media its own publicly reachable target. Event uploads never leave the cluster.

**Set `externalUrl` before single sign-on, and close sign-up once the admin exists.** `externalUrl` is the base of every sign-in redirect and every link Langfuse writes into email and Slack. The spec refuses any SSO provider while it is still the localhost default. On an exposed instance, `auth.disableSignUp` stops strangers from creating accounts, and `auth.newUserAccess` grants the memberships people need when they arrive through SSO.

**Prefer a cloud identity over a static key.** On AWS and Google Cloud, leaving the target's static credentials empty makes Langfuse use the pod's identity: `serviceAccount.workloadIdentity` binds the ServiceAccount to an IAM role (EKS IRSA) or a Google service account (GKE Workload Identity). The same identity covers Bedrock, Vertex and SES. Azure Blob Storage is the exception: Langfuse authenticates to it with the account key only.

**Size the web server and worker for your trace volume.** Both default to requests of 250m CPU and 1 GiB with limits of 2 CPU and 4 GiB, which suits evaluation; Langfuse recommends requesting the full 2 CPU and 4 GiB per pod in production. Web replicas serve the UI and accept ingestion; workers drain the queues into ClickHouse, so scale workers first when ingestion lags. Each workload takes at most one autoscaler (an HPA, a KEDA ScaledObject or a VPA), and a `pdb` belongs on any workload running two or more replicas.

**Changing the seeded project later changes nothing in Langfuse.** Headless initialization creates each object only if it does not exist yet. A renamed organization or project, or a new admin password, is ignored after the first start, and a new key pair is added beside the old one without revoking it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Kubernetes Namespace | `namespace` | `spec.name` |
| PostgreSQL | `postgres.host` | `status.outputs.rw_service` |
| PostgreSQL | `postgres.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| ClickHouse | `clickhouse.host` | `status.outputs.service_name` |
| ClickHouse | `clickhouse.passwordSecret.secretName` | `status.outputs.auth_secret_name` |
| ClickHouse | `clickhouse.cluster.name` | `status.outputs.cluster_name` |
| Valkey | `redis.standalone.host` | `status.outputs.service` |
| Valkey | `redis.username` | `status.outputs.username` |
| Valkey | `redis.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| Valkey | `redis.passwordSecret.secretKey` | `status.outputs.password_secret.key` |
| SeaweedFS | `blobStorage.target.s3Compatible.endpoint` | `status.outputs.s3_endpoint` |
| SeaweedFS | `blobStorage.target.s3Compatible.credentialsSecret.secretName` | `status.outputs.s3_credentials_secret_name` |
| AWS S3 Bucket | `blobStorage.target.s3.bucket` | `status.outputs.bucket_id` |
| AWS S3 Bucket | `blobStorage.target.s3.region` | `status.outputs.region` |
| AWS KMS Key | `blobStorage.eventUpload.sse.kmsKeyId` | `status.outputs.key_arn` |
| Cloudflare R2 Bucket | `blobStorage.target.r2.endpoint` | `status.outputs.s3_endpoint` |
| Cloudflare R2 Bucket | `blobStorage.target.r2.bucket` | `status.outputs.bucket_name` |
| Cloudflare Account API Token | `blobStorage.target.r2.credentials.accessKeyId` | `status.outputs.r2_access_key_id` |
| Cloudflare Account API Token | `blobStorage.target.r2.credentials.secretAccessKey` | `status.outputs.r2_secret_access_key` |
| GCP GCS Bucket | `blobStorage.target.gcs.bucket` | `status.outputs.bucket_name` |
| Azure Storage Account | `blobStorage.target.azureBlob.storageAccount` | `status.outputs.storage_account_name` |
| Azure Storage Account | `blobStorage.target.azureBlob.endpoint` | `status.outputs.primary_blob_endpoint` |
| Azure Storage Container | `blobStorage.target.azureBlob.container` | `status.outputs.container_name` |
| Cert Manager Certificate | `redis.tls.clientCertificate.secretName` | `status.outputs.secret_name` |
| OpenTelemetry Collector | `observability.otlpEndpoint` | `status.outputs.otlp_http_endpoint` |
| AWS IAM Role | `aiFeatures.assistant.sandbox.executionRoleArn` | `status.outputs.role_arn` |

The media and batch-export targets accept the same references as the event target, the TLS client key accepts a Cert Manager Certificate's Secret the same way, and `serviceAccount.workloadIdentity` references an IAM role, a Google service account or an Azure managed identity.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|---|---|---|
| `service` | The web Service name | The backend of a KubernetesHttpRoute or KubernetesIngress that exposes Langfuse |
| `endpoint` | The in-cluster base URL of the web server | `LANGFUSE_HOST` for applications in the same cluster that send traces through a Langfuse SDK |
| `otel_endpoint` | The OTLP endpoint (`<endpoint>/api/public/otel`) | The exporter endpoint of an OpenTelemetry Collector or an OTLP-instrumented application |
| `project_public_key` | The seeded project's public key | `LANGFUSE_PUBLIC_KEY` for applications that send traces |
| `project_secret_key_secret` | The Secret name and key holding the seeded project's secret key | A `secretKeyRef` for `LANGFUSE_SECRET_KEY` in an application's Deployment |
| `admin_password_secret` | The Secret name and key holding the seeded admin's password | Handing the first sign-in to the platform team |
| `port_forward_command` | The `kubectl port-forward` command that opens the UI locally | First sign-in before any route exists |

## Common Patterns

**Evaluation against in-cluster stores.** One web replica, one worker, a single ClickHouse server with cluster mode off, and every secret generated. It is the fastest way to see traces in Langfuse, and nothing in it is shaped for keeping data. Start from the **Evaluation** preset.

**Production on in-cluster stores.** Two replicas of each workload with disruption budgets, a replicated ClickHouse with Keeper, the salt and encryption key declared as managed secrets, a seeded admin, sign-up closed and an `externalUrl` that a route serves. The `self-hosted-langfuse` InfraChart composes this shape with its stores. Start from the **Production In-Cluster** preset.

**Langfuse on a managed cloud.** Keep Langfuse on the cluster and move events, media and exports to the cloud's object storage with keyless access: S3 with an IAM role on EKS, Google Cloud Storage with a Google service account on GKE, or Azure Blob Storage with an account key on AKS. Postgres and ClickHouse can be managed services too, by literal host. Start from the **AWS**, **Google Cloud** or **Azure** preset.

**Many applications, one Langfuse.** Each team's application reads `endpoint`, `project_public_key` and `project_secret_key_secret` by reference, and its traces land in the seeded project. Teams that need their own projects create them in the UI; the seeded project is the bootstrap, not the limit.

## Works With

- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the Postgres store; the `postgres` foreign-key defaults point at its read-write Service and application credential Secret
- [**ClickHouse**](/infra-catalog/kubernetes-click-house) -- the trace store; declare the Langfuse user with `networks`, and the `clickhouse` foreign-key defaults do the rest
- [**Valkey**](/infra-catalog/kubernetes-valkey) -- the queue and cache store, run with `noeviction`
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) -- in-cluster S3-compatible object storage for events, media and exports
- [**CloudNativePG Operator**](/infra-catalog/kubernetes-cloud-native-pg-operator) and [**Altinity ClickHouse Operator**](/infra-catalog/kubernetes-altinity-operator) -- the operators behind the composed Postgres and ClickHouse
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket), [**Cloudflare R2 Bucket**](/infra-catalog/cloudflare-r2-bucket), [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) and [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- cloud object storage targets
- [**KEDA**](/infra-catalog/kubernetes-keda) -- the runtime behind `autoscaling.keda` on either workload
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) and [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- exposure over the `service` output, at the host `externalUrl` names
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace Langfuse and its stores share

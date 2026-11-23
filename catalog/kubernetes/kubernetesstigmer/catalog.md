# Stigmer

Installs Stigmer, the open-source AI agent platform, on your Kubernetes cluster: the Stigmer operator and the one Stigmer install it runs, with your model key, database, bucket, front door, sign-in and per-session sandboxes declared in one block. One Stigmer runs per cluster, the operator's version is Stigmer's version, and a destroy waits for the operator to remove everything it installed while keeping the encryption keys every backup is sealed with.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** for the operator — created only when `operator.create_namespace` is `true` (`stigmer-system` is the convention); otherwise the operator installs into an existing namespace
- **Helm Release** (`stigmer-operator`) — the Stigmer operator at `version`, with its cluster-wide role and the `stigmers.stigmer.ai` definition, which Helm keeps when the release is removed
- **Kubernetes Secrets** in the Stigmer's namespace — one for each credential the spec gives as a value (the model key, the bucket's key pair, the database password, the SMTP login, the license, Temporal's key and certificates, secret extra settings), each holding only the keys declared; none when every credential names a Secret you keep
- **Stigmer** (`stigmer`) — the install itself, declared with only the settings you chose, so the operator's own defaults decide the rest

From the Stigmer, the operator creates the server and the shared runner, Postgres through CloudNativePG and Temporal (unless you bring your own), the front door and the bundled sign-in (when you declare a hostname), and the sandbox namespace (with a per-session driver). It installs CloudNativePG, its backup plugin, agent-sandbox or Agent Substrate when the declaration needs them and the cluster lacks them, and leaves a copy someone else installed alone. It mints `stigmer-keys`, the encryption keys, and keeps them when the Stigmer is deleted.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster. Installing an operator is a cluster-admin act: the chart grants it a cluster-wide role, and the operator installs cluster-level prerequisites. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **kubectl where the deploy runs** — the restore check and the destroy's cleanup wait ask the cluster with `kubectl`, which the Planton Runner carries.

### Kubernetes Cluster

- **Kubernetes 1.30 or newer with a default storage class**, or a storage class named in the volume settings (GKE 1.37 has no default).
- **The Stigmer's namespace exists** — this kind never creates or deletes it, because it holds the encryption keys. Give it a Kubernetes Namespace of its own.
- **No other Stigmer operator** — one runs per cluster; another install is refused before anything is applied.
- **No CloudNativePG left behind without its controller** — the definitions a destroyed CloudNativePG leaves are refused before anything is applied, with the command that clears them, because Stigmer's operator would trust them and wait forever.
- **cert-manager** (only for backups, or a certificate issued by an issuer) — the backup plugin needs it, and nothing here installs it.
- **An ingress controller or a Gateway** (only with a hostname).
- **A CNI that enforces NetworkPolicy** (only for egress fences and sandbox isolation).

## Deploy

### Console

Open the deployment store, find **Stigmer**, and click **Deploy**. The default path asks for the namespace, the version and the model key; the database, the bucket, the front door, sign-in, sandboxes and Enterprise are optional sections. Start from the **Minimal** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStigmer
metadata:
  name: stigmer
  org: acme-corp
  env: prod
spec:
  namespace:
    value: stigmer
  version: 0.3.2
  operator:
    namespace:
      value: stigmer-system
    create_namespace: true
  owner:
    email: owner@acme.example
  runner:
    model:
      keys:
        anthropic_api_key: $secret/anthropic-api-key
```

```shell
planton apply -f stigmer.yaml
```

This installs the operator in `stigmer-system`, writes the model key from your Planton secret into a Secret the runner reads, and declares a Stigmer reachable from inside the cluster with its database and Temporal run for you. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace, your own Postgres and Temporal, and the bucket by reference:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: stigmer
      fieldPath: spec.name
  database:
    external:
      host:
        valueFrom:
          kind: KubernetesPostgres
          name: stigmer-db
          fieldPath: status.outputs.rw_service
      user: stigmer
      password:
        value: $secret/stigmer-db-password
  temporal:
    external:
      host_port:
        valueFrom:
          kind: KubernetesTemporal
          name: stigmer-temporal
          fieldPath: status.outputs.frontend_endpoint
  object_store:
    s3:
      bucket:
        value: stigmer
      endpoint:
        valueFrom:
          kind: KubernetesSeaweedFs
          name: storage
          fieldPath: status.outputs.s3_endpoint
      credentials:
        access_key:
          access_key_id: stigmer
          secret_access_key: $secret/stigmer-bucket-secret-key
```

The InfraPipeline deploys the namespace, the database, Temporal and the store first, then installs Stigmer on them.

## Key Configuration

These are the most important decisions when configuring a Stigmer install. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The version is pinned on purpose, and only moves forward** — `version` is the chart's, the operator's and the server's version, and it has no default, so an ordinary re-apply never upgrades a running install. Moving it below the version the server runs is refused at plan time: Stigmer's server refuses to start on a database a newer release migrated, so going back is a restore into a new install. With backups on, the operator backs the database up before every upgrade and holds the upgrade while that backup fails; `upgrade_without_backup_to` releases that one upgrade.

**Keep the keys outside the cluster** — the operator mints `stigmer-keys` once and keeps it through a destroy, because every stored secret and every backup is sealed with it. Deleting the Stigmer's namespace deletes it. Run the `keys_export_command` output once and keep the file; a restore into a new cluster applies it first, and a restore without it is refused before the Stigmer is declared.

**Bundled or your own database decides how people sign in** — left unset, the operator runs Postgres and Temporal for you. Your own Postgres needs your own Temporal, and cannot back the bundled sign-in, whose accounts live on the bundled database: an exposed install on your own Postgres signs people in through your own OpenID Connect issuer.

**A hostname turns sign-in on** — without `ingress`, every caller acts as the owner, which is safe only while the install is reachable from inside the cluster. With it, the operator runs the bundled sign-in unless you name your own issuer, and the owner's account is created first with a one-time password. `allow_unauthenticated_exposure` exposes the install with sign-in off, deliberately.

**A bucket is the gate to backups and sandboxes** — with `object_store`, files leave the shared volume, the bundled database archives every change from its first minute, and per-session sandboxes become possible. Backups also need cert-manager on the cluster.

**Per-session sandboxes isolate engineers' agents** — `agentSandbox` runs each session in a pod of its own behind a network fence; `substrate` runs each in a gVisor sandbox on a pool of workers and needs Kubernetes 1.37 with the pod-certificate APIs. Stigmer Enterprise requires one of them, and your own issuer.

**Destroy waits, and never blocks** — the Stigmer goes first, then the destroy waits up to ten minutes for the operator to remove the database, the sandboxes and the prerequisites it installed, before the chart goes. If the operator cannot finish, the destroy continues and prints what stays and how to remove it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|----------------|
| **Kubernetes Namespace** | `namespace`, `operator.namespace` | `spec.name` |
| **PostgreSQL** | `database.external.host` | `status.outputs.rw_service` |
| **PostgreSQL** | `database.external.password.existing_secret.name` | `status.outputs.password_secret.name` |
| **Temporal** | `temporal.external.host_port` | `status.outputs.frontend_endpoint` |
| **SeaweedFS** | `object_store.s3.endpoint` | `status.outputs.s3_endpoint` |
| **AWS S3 Bucket** | `object_store.s3.bucket` | `status.outputs.bucket_id` |
| **Cert Manager Cluster Issuer** | `ingress.tls.issuer.name` | `status.outputs.cluster_issuer_name` |
| **Kubernetes Gateway** | `ingress.gateway.name`, `ingress.gateway.namespace` | `status.outputs.gateway_name`, `status.outputs.namespace` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `server_endpoint` | The server's in-cluster address, port 7234 | An in-cluster client, a gateway route or a service mesh entry pointing at Stigmer |
| `url` | The hostname with a front door, or the port-forward address | Links in runbooks and dashboards |
| `keys_export_command` | Saves the encryption keys as a re-appliable file | The disaster-recovery runbook; a restore into a new cluster |
| `owner_password_command` | Reads the owner's one-time password | The owner's first sign-in |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Inside the cluster** — everything bundled, no hostname, reached by port-forward. The right first install and the evaluation path. Start from the **Minimal** preset.

**On databases you already run** — your own Postgres and Temporal by reference, with your own issuer when exposed. Start from the **Bring Your Own Database** preset.

**Recoverable** — files and the database's archive in a bucket, a daily backup and a standby. Start from the **Bucket and Backups** preset.

**Exposed to your team** — a hostname with a certificate and the bundled sign-in, sign-up for your domains verified by email. Start from the **Exposed with Sign-in** preset.

**Isolated sessions** — each session in a sandbox of its own behind a fence. Start from the **Per-session Sandboxes** preset; Stigmer Enterprise builds on it in the **Enterprise** preset.

**Shared cluster operators** — CloudNativePG and its backup plugin from the catalog, deployed first with `depends_on`, so Stigmer uses them and never removes them. Start from the **Composed Prerequisites** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — the Stigmer's namespace and the operator's; the Stigmer's holds the encryption keys
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) — your own database, by its read-write Service
- [**Temporal**](/infra-catalog/kubernetes-temporal) — your own workflow engine, by its frontend endpoint
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) — the S3-compatible bucket for files and backups
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) — required for backups and issued certificates
- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) — issues the front door's certificate
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) — the front door as routes on your Gateway
- [**CloudNativePG Operator**](/infra-catalog/kubernetes-cloud-native-pg-operator) and [**CNPG Barman Cloud Plugin**](/infra-catalog/kubernetes-cnpg-barman-cloud-plugin) — shared database operators, deployed first
- [**Agent Sandbox Controller**](/infra-catalog/kubernetes-agent-sandbox-controller) and [**Agent Substrate**](/infra-catalog/kubernetes-agent-substrate) — the per-session sandbox runtimes, when the cluster already runs them

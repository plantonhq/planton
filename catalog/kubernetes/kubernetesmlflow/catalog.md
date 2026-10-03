# MLflow

Declares one MLflow tracking server and model registry -- every training run's parameters, metrics, artifacts and models logged, compared, versioned and staged from one server your whole team points `MLFLOW_TRACKING_URI` at. MLflow publishes NO Helm chart: the IaC module renders its own typed manifests around the official `ghcr.io/mlflow/mlflow` image (the `-full` variant, which carries the database drivers, object-store clients and auth dependencies the bare image omits), SECURED-BY-DEFAULT -- upstream ships an OPEN server whose documented example is `admin/password1234`, and neither ever ships from here: basic auth is on with a module-generated admin password unless you explicitly turn it off. Two stores shape the design: the BACKEND store (experiments, runs, metrics, the registry -- and with auth on, users and permissions) on a sqlite PVC by default or in a composed PostgreSQL/MySQL, and the ARTIFACT store (models, datasets) on a local PVC by default or in a composed SeaweedFS bucket, AWS S3, GCS or Azure Blob. The server PROXIES all artifact traffic -- clients never hold storage credentials.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The tracking-server Deployment** -- the official image at the pinned tag (default `v3.15.0-full`), `replicas` per your declaration (more than one requires the postgres backend AND an object artifact store -- enforced at validation), with readiness/liveness probes and your resource requests/limits
- **The Service** -- port 5000, `ClusterIP` by default; the annotations surface carries cloud-LB recipes when you choose `LoadBalancer`
- **Store volumes, only on the PVC postures** -- a 5Gi tracking-state PVC when no backend is declared (or your customized sqlite PVC), and a 10Gi artifact PVC when no artifact store is declared (or your customized one)
- **Module-owned Secrets** -- the backend connection URI composed AT APPLY TIME from your referenced credential Secret (the password never renders into any manifest), and with auth on and no BYO Secret, a generated admin password in `<name>-admin-auth`
- **The GC CronJob** -- only when `gc.enabled` is `true`: `mlflow gc` against the same stores on your schedule, hard-deleting runs past the retention window
- **A ServiceMonitor** -- only when `metrics.serviceMonitorEnabled` is `true` (requires the Prometheus Operator CRDs in the cluster)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A backend database that ALREADY EXISTS, when declared** -- the postgres/mysql arms connect to a database that must pre-exist (the server creates tables, never the database). On a composed KubernetesPostgres, declare `mlflow` at bootstrap (`initdb`); the spec's foreign-key defaults then wire the host to its read-write Service and the password to its application-user Secret.
- **A bucket that ALREADY EXISTS, when declared** -- MLflow writes objects, it never creates buckets. On a composed KubernetesSeaweedFs, declare the bucket under `s3.buckets`; on AWS/GCS/Azure, create it first.
- **Credential Secrets in THIS namespace** -- the module reads the database credential Secret at apply time to compose the connection URI, and a Secret is readable only from the workload's own namespace. Co-locate MLflow with its database and object store, or replicate the credential Secrets.
- **Ambient cloud identity for the keyless postures** -- on AWS S3 and GCS, an empty credentials block means the pod's identity (IRSA / EKS Pod Identity, GKE Workload Identity) must grant bucket access.
- **A plan for PVC-backed state** -- with the default stores, tracking state and artifacts live on PVCs in this namespace: DELETING THE NAMESPACE DELETES THEM. Move both stores external before treating an instance as disposable.

## Deploy

### Console

Open the deployment store, find **MLflow**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team tracking server — zero dependencies** preset for the zero-dependency secured server, or the **Production — PostgreSQL backend, S3 artifacts** preset for the durable team-scale shape in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMlflow
metadata:
  name: mlflow
  org: acme-corp
  env: dev
spec:
  namespace:
    value: "mlflow"
  createNamespace: true
```

```shell
planton apply -f mlflow.yaml
```

This declares the secured team default -- and the near-empty manifest IS the point: basic auth ON with the admin password generated into `mlflow-admin-auth` (key `password`), tracking state on a 5Gi sqlite PVC, artifacts on a 10Gi PVC served through the tracking server, everything by absence. Point training code at the tracking endpoint from the outputs and set `MLFLOW_TRACKING_USERNAME` / `MLFLOW_TRACKING_PASSWORD`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the backend to a PostgreSQL and the artifacts to a SeaweedFS managed by other Infra Components:

```yaml
spec:
  namespace:
    value: "mlflow"
  createNamespace: true
  backendStore:
    postgres:
      host:
        valueFrom:
          kind: KubernetesPostgres
          name: mlflow-db
          fieldPath: status.outputs.rw_service
      passwordSecret:
        secretName:
          valueFrom:
            kind: KubernetesPostgres
            name: mlflow-db
            fieldPath: status.outputs.password_secret.name
  artifactStore:
    s3Compatible:
      endpoint:
        valueFrom:
          kind: KubernetesSeaweedFs
          name: artifacts
          fieldPath: status.outputs.s3_endpoint
      bucket: mlflow-artifacts
      credentialsSecret:
        secretName:
          valueFrom:
            kind: KubernetesSeaweedFs
            name: artifacts
            fieldPath: status.outputs.s3_credentials_secret_name
```

The InfraPipeline deploys the database and object store first, then declares MLflow against them -- the module composes the connection URI from the referenced Secrets at deploy time, so nothing password-shaped is ever written into this manifest or any rendered resource. Declare the `mlflow` database at the Postgres kind's bootstrap (`initdb`) and the `mlflow-artifacts` bucket under the SeaweedFs kind's `s3.buckets`.

## Key Configuration

These are the most important decisions when configuring MLflow. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Sign-in, read this first** -- upstream MLflow ships NO authentication; this kind deliberately inverts that. Leave `auth` untouched and sign-in is ON with a generated admin password in `<name>-admin-auth`; turning it off is an explicit, warned decision (anyone reaching the Service can read, write and DELETE everything). `defaultPermission` (READ | EDIT | MANAGE | NO_PERMISSIONS) decides what signed-in users can do on experiments nobody granted them -- `NO_PERMISSIONS` makes experiments private to their creators until shared. Users and grants live in the BACKEND store, managed through the admin account via MLflow's auth API.

**The image variant matters, verified live** -- the bare `vX.Y.Z` image carries NO database drivers (psycopg2/PyMySQL), NO object-store clients (boto3/GCS/Azure) and NO auth dependency: postgres/mysql backends, every remote artifact store and the secured sign-in ALL crash-loop on it (`No module named 'psycopg2'`). The default tag `v3.15.0-full` carries everything; a custom image must ship the same dependency set.

**The backend store, sqlite is right to start** -- no `backendStore` block = sqlite on a 5Gi PVC: zero external dependencies, single replica only. Declare `postgres` (or `mysql`) when tracking state must survive volume loss, snapshot with your database fleet, or serve more than one replica -- the database must PRE-EXIST (declare it at the Postgres kind's `initdb`). Auth state follows this choice: moving the backend moves your user roster with it.

**Artifacts and the proxy seam** -- no `artifactStore` block = a local 10Gi PVC served through the tracking server. Five explicit arms: `pvc` (customized local) | `s3Compatible` (a composed KubernetesSeaweedFs slots in by reference; MinIO-class endpoints work as literals -- addressing style is automatic, boto3 uses path-style for custom endpoints, no knob exists by design) | `awsS3` and `gcs` (KEYLESS when the credentials block is empty -- IRSA / Workload Identity carries access) | `azureBlob` (access key required). Whatever the arm, the server proxies ALL artifact traffic: clients need only the tracking URI and their login; rotating a store key touches one Secret in this namespace.

**Replicas cross both stores** -- the server is stateless once BOTH stores are external, and the spec enforces exactly that: more than one replica requires the postgres backend AND an object artifact store (an ABSENT artifact store counts as the PVC posture). Scale `workers` (1-64, default 4) for one pod's concurrency before scaling pods.

**Garbage collection makes deletion real** -- deleting a run in MLflow only MARKS it; data stays and the run is restorable until `gc` (off by default) collects it. Once collected, a run is UNRECOVERABLE -- `olderThan` (default `30d`) is the team's undo window; `schedule` (default `0 3 * * *`) is when the CronJob sweeps.

**The front door composes** -- `service.type` defaults `ClusterIP`: training jobs in-cluster reach the server by Service DNS, and real exposure composes from first-class kinds over the exported handle. TLS never terminates here. The `extraEnv` / `extraEnvFromSecret` / `extraArgs` escape hatches follow one rule: plain values and arguments render VERBATIM into the pod spec -- anything secret rides the Secret-backed variables (valueFrom references).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesPostgres** | `backendStore.postgres.host` | `status.outputs.rw_service` |
| **KubernetesPostgres** | `backendStore.postgres.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| **KubernetesMysql** | `backendStore.mysql.host` | `status.outputs.primary_service` |
| **KubernetesMysql** | `backendStore.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| **KubernetesSeaweedFs** | `artifactStore.s3Compatible.endpoint` | `status.outputs.s3_endpoint` |
| **KubernetesSeaweedFs** | `artifactStore.s3Compatible.credentialsSecret.secretName` | `status.outputs.s3_credentials_secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace MLflow runs in | Co-locating the database, object store and diagnostics |
| `service` | The tracking server's Service | The handle exposure kinds route to |
| `tracking_endpoint` | In-cluster URL, port 5000 | Set as `MLFLOW_TRACKING_URI` in training jobs and notebooks |
| `admin_password_secret` | Secret + key holding the generated admin password (EMPTY when auth is disabled) | Distributing the admin credential; rotating it |
| `backend_store_uri_secret_name` | The module-owned Secret carrying the composed connection URI (EMPTY on the sqlite posture) | Sidecars needing the same database -- the URI embeds the password; share the NAME, never echo contents |
| `port_forward_command` | kubectl port-forward one-liner for the UI | Reaching the UI from a workstation on the ClusterIP default |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team tracking in minutes** -- The zero-dependency secured server: sqlite tracking state, local artifacts through the proxy seam, sign-in on with a generated admin password. One manifest, nothing else to operate. Start from the **Team tracking server — zero dependencies** preset.

**Production on composed state** -- Tracking state in a composed KubernetesPostgres, artifacts in a composed KubernetesSeaweedFs bucket, two stateless replicas behind one Service, experiments private-by-default (`NO_PERMISSIONS`), scheduled GC keeping storage honest. Start from the **Production — PostgreSQL backend, S3 artifacts** preset.

**CI experiment sandboxes** -- The near-empty default shape per pipeline namespace, with aggressive GC (`olderThan: 24h`) -- every run disposable by design, and the namespace deleted with the pipeline.

## Works With

- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the durable backend store; the `backendStore.postgres` foreign-key defaults point at it
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the MySQL alternative for the backend store
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) -- the natural in-cluster artifact store; its S3 endpoint and credentials Secret slot into `artifactStore.s3Compatible`
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the deployment
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- exposure over the exported `service` handle, where TLS terminates
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the Gateway API alternative for exposing the tracking server

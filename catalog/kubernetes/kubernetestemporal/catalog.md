# Temporal

Deploys Temporal -- the durable workflow engine (long-running business logic, human-in-the-loop flows, saga orchestration, AI-agent pipelines) -- from the official `temporal` Helm chart: the four server services (frontend, history, matching, worker) as separate Deployments, the Web UI (on by default), an admin-tools pod for operational commands, and the schema-setup Jobs that prepare the databases before the server starts. Bring your own database -- nothing is bundled: PostgreSQL (the recommended path; a KubernetesPostgres composes naturally), MySQL 8, or an external Cassandra cluster for the default store (the visibility store must be SQL). Supports per-service replica and resource tuning, declarative Temporal namespaces with retention, runtime limits via dynamic config, archival of closed histories to S3/GCS, ServiceMonitor metrics, container image overrides for air-gapped clusters, and a Helm values escape hatch.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Temporal Helm Release** -- the official Temporal Helm chart, creating:
  - Deployments for the four server services: frontend (the gRPC/HTTP API gateway), history (workflow state and the history shards -- the most resource-intensive), matching (task-queue dispatch), and worker (Temporal's internal system workflows)
  - Schema-setup Jobs that create and migrate the default and visibility store schemas before the server starts (skippable with `database.skipSchemaSetup` for teams with their own schema pipeline)
  - Web UI Deployment (on by default; turn off with `webUi.enabled: false`)
  - Admin-tools pod (on by default) -- a shell with `temporal` and the schema tools pre-installed, for `kubectl exec` operational commands
  - Internal-frontend service when `internalFrontendEnabled` is `true` -- required when you enable authorization on the public frontend
  - A Job that creates each declared Temporal namespace idempotently (`temporal operator namespace create`)
  - ClusterIP Services for the frontend (gRPC 7233, HTTP 7243) and Web UI (HTTP 8080)
  - ServiceMonitor resources (one per server service) when `serviceMonitorEnabled` is `true`; otherwise pods carry `prometheus.io` scrape annotations for annotation-based collection
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A reachable database** -- a composed KubernetesPostgres (recommended) or KubernetesMysql in the same namespace, or any reachable PostgreSQL ≥ 12 / MySQL 8 / Cassandra endpoint. Temporal uses TWO databases: the default store (workflow state, default name `temporal`) and the visibility store (the search index, default `temporal_visibility`). Both must exist before install unless `database.createDatabases` is on (which needs create-database privileges). On a KubernetesPostgres: declare `temporal` as the bootstrap database and create `temporal_visibility` with one line of `post_init_sql`.
- **The database credential Secret in the install namespace** -- the password is read through a secretKeyRef, and a secretKeyRef can only read Secrets in the workload's OWN namespace. Co-locate Temporal with its database (the default composition) or replicate the credential Secret.
- **Prometheus Operator CRDs** (only when `serviceMonitorEnabled` is `true`) -- a KubernetesKubePrometheusStack composes naturally.

## Deploy

### Console

Open the deployment store, find **Temporal**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev preset** for the smallest useful cluster or the **Production preset** for sized services with ServiceMonitor metrics in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTemporal
metadata:
  name: workflow-engine
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "temporal"
  createNamespace: true
  database:
    postgres:
      host:
        valueFrom:
          name: temporal-db
      username: temporal
      passwordSecret:
        secretName:
          valueFrom:
            name: temporal-db
  temporalNamespaces:
    - name: default
      retention: 7d
```

```shell
planton apply -f temporal.yaml
```

This creates a Temporal cluster against a composed KubernetesPostgres named `temporal-db` in the same namespace: the host reference resolves to the Postgres read-write Service (always the current primary) and the credential to the operator-maintained application Secret, so nothing password-shaped ever appears in the manifest. Everything else is defaults -- the pinned chart, one replica of each server service, the Web UI and admin-tools pod on, 512 history shards, and one Temporal namespace (`default`) with 7-day retention. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Temporal deployment to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: temporal-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the Temporal cluster into it.

## Key Configuration

These are the most important decisions when configuring a Temporal deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Database backend** -- Exactly one of `postgres` (recommended -- a KubernetesPostgres composes; any reachable PostgreSQL ≥ 12 works with literal values), `mysql` (a KubernetesMysql composes), or `cassandra` (an EXTERNAL cluster you operate yourself -- the catalog has no Cassandra kind). Cassandra serves the default store ONLY: Temporal removed Cassandra visibility support in v1.21, so the cassandra backend REQUIRES a `database.visibility` SQL block. For SQL backends, `visibility` is optional -- declare it only to place the search index on a different server than workflow state.

**History shards** -- `numHistoryShards` is IMMUTABLE: the value is baked into the default store's schema at first install and cannot be changed afterwards without a full cluster migration. Pick for the cluster you will grow into. Empty = 512, the upstream default, which carries most production workloads.

**Per-service scaling** -- The `services` section sets replicas and container resources independently for each of the four server services. All four scale horizontally; history is the stateful heavy-lifter (shard ownership redistributes across its replicas) and benefits first from more replicas and memory.

**Dynamic configuration** -- The `dynamicConfig` section pushes runtime limits without a server restart: workflow history size and event-count ceilings, and per-payload blob size limits, each with an error (terminate) and warn (log) threshold. Raise `blobSizeLimitError` for workflows passing large payloads -- better yet, pass references to object storage. Other dynamic-config keys ride `helmValues` under `server.dynamicConfig`.

**Archival** -- Closed workflow histories and visibility records can be archived to S3, GCS, or a pod-local filesystem path (dev/test only) so they survive retention-driven deletion. Cloud credentials are ambient (IRSA / Workload Identity) -- nothing credential-bearing is rendered.

**External access** -- The frontend and Web UI Services stay ClusterIP; this kind deploys no exposure itself. Compose first-class kinds over the exported service handles: a KubernetesService for a LoadBalancer, or Gateway API route kinds for hostname-based access. Workers and SDK clients inside the cluster connect through the exported `frontend_endpoint` directly.

**Helm values** -- `helmValues` is the escape hatch for chart values the typed fields do not model (mTLS certificate mounts, JWT authorization, multi-cluster replication, per-service scheduling), merged last with Helm `-f` semantics. Never put secret material in it -- passwords ride existing Secrets through the typed fields. KNOW THIS: the chart's old bundled Cassandra/Elasticsearch/Prometheus/Grafana subcharts were REMOVED upstream, and declaring their legacy keys makes the chart itself fail rendering.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| PostgreSQL | `spec.database.postgres.host` | `status.outputs.rw_service` |
| PostgreSQL | `spec.database.postgres.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| MySQL | `spec.database.mysql.host` | `status.outputs.primary_service` |
| MySQL | `spec.database.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace where the cluster runs | Worker application deployment manifests |
| `frontend_service` | The frontend Service's name (`<name>-frontend`) | The handle exposure kinds route to |
| `frontend_endpoint` | In-cluster frontend gRPC endpoint (`<name>-frontend.<namespace>.svc.cluster.local:7233`) | Temporal SDK worker and client server addresses |
| `frontend_http_endpoint` | In-cluster frontend HTTP API endpoint (port 7243) | Clients that cannot speak gRPC |
| `web_ui_service` | The Web UI Service's name (`<name>-web`); empty when the UI is off | Exposure kinds routing browser traffic |
| `web_ui_endpoint` | In-cluster Web UI endpoint (port 8080); empty when the UI is off | Internal UI access |
| `port_forward_frontend_command` | Ready-to-run `kubectl port-forward` command for the frontend | Local development access |
| `port_forward_web_ui_command` | Ready-to-run `kubectl port-forward` command for the Web UI | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Smallest useful cluster for development** -- All four server services at one replica, the Web UI, and one Temporal namespace (`default`), against a composed KubernetesPostgres in the same Kubernetes namespace. The 512-shard default is kept deliberately -- never lower it "because dev"; it is immutable. Start from the **Dev preset**.

**Sized for production** -- Replicated frontend/matching/worker, three history replicas, a replicated stateless UI, explicit retention on every declared Temporal namespace, and ServiceMonitor metrics for the Prometheus Operator. Start from the **Production preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the Temporal deployment
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the recommended default and visibility store; the host and credential references resolve to its read-write Service and operator-maintained Secret
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the MySQL 8 alternative for both stores
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- provides the Prometheus Operator that scrapes the per-service ServiceMonitors

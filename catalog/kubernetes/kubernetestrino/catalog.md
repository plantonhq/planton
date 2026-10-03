# Trino

Declares one Trino install -- the distributed SQL query engine that queries data WHERE IT LIVES (data lakes, object stores, relational databases) and JOINs across sources in one query. The official Helm chart renders a coordinator Deployment (the brain: parses, plans and schedules; REST API + Web UI on port 8080) and a worker Deployment (the muscle: query splits execute here), with every configuration surface rendered into checksum-annotated ConfigMaps so config changes roll the pods automatically. SECURED BY DEFAULT: upstream ships NO authentication -- anyone who can reach the Service can query every catalog -- but this kind enables PASSWORD (file) authentication with a module-generated admin (`<name>-auth` Secret) and configures the internal-communication shared secret Trino requires once auth is on. And the install is immediately queryable: the in-image `tpch`/`tpcds` sample catalogs answer `SELECT count(*) FROM tpch.tiny.nation` on a fresh deploy, no data source required.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The Helm release** -- the official `trino/trino` chart, rendering:
  - The coordinator (`<name>-trino-coordinator`, port 8080) -- query planning, scheduling, the REST API and Web UI
  - The worker Deployment (`<name>-trino-worker`) -- 2 replicas by default; 0 is the single-node shape paired with `coordinator.includeInScheduling`; optionally HPA- or KEDA-scaled
  - One `<name>.properties` catalog file per declared catalog, plus the samples until disabled
  - With graceful shutdown: a preStop drain, with `terminationGracePeriodSeconds` set to TWICE the grace period automatically (the chart's own requirement)
- **Module-owned Secrets** -- the admin credential (`<name>-auth`, key `password`, bcrypt entry written into the server's password file) and the internal-communication shared secret -- catalog passwords reach Trino as `${ENV:VAR}` references resolved from YOUR referenced Secrets, never rendered literals
- **JMX metrics, when enabled** -- the Prometheus JMX-exporter sidecar on coordinator and worker pods, with optional per-tier ServiceMonitors (requires the Prometheus operator CRDs)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Catalog credential Secrets in THIS namespace** -- catalog passwords are read by the Trino pods at RUNTIME, and a Secret can only be referenced from the workload's own namespace. Co-locate Trino with the database Secrets its catalogs reference (the default composition), or replicate them.
- **A name within budget** -- keep `metadata.name` at 36 characters or fewer (the module renders a `-schemas-volume-coordinator` ConfigMap suffix against the Kubernetes 63-character cap), and at 27 or fewer when declaring `resourceGroupsConfig` (its longer suffix). Both engines fail loudly over budget.
- **The KEDA operator, when event-scaling workers** -- a KubernetesKeda composes naturally. Without it, the ScaledObject the chart renders has nothing to reconcile it.
- **Durable spooling storage, for fault tolerance** -- fault-tolerant execution needs an object-store destination (`s3://bucket`); a KubernetesSeaweedFs bucket works via its S3 endpoint.

## Deploy

### Console

Open the deployment store, find **Trino**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **SQL playground** preset for the two-line, immediately queryable install, or **Federated warehouse** for the production posture with a composed PostgreSQL catalog in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTrino
metadata:
  name: analytics-trino
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "analytics"
  createNamespace: true
  catalogs:
    postgres:
      - name: warehouse
        host:
          value: "warehouse-pg-rw.analytics.svc"
        database: warehouse
        passwordSecret:
          secretName:
            value: "warehouse-pg-app"
```

```shell
planton apply -f analytics-trino.yaml
```

This declares the near-defaults install: a coordinator and two workers, PASSWORD authentication ON with a module-generated admin (the secured default -- the open server never ships), the samples still on beside one real PostgreSQL catalog -- `SELECT * FROM warehouse.public.orders JOIN tpch.tiny.nation ...` works the moment it is up. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire a catalog to a database managed by another Infra Component:

```yaml
spec:
  namespace:
    value: "analytics"
  createNamespace: true
  catalogs:
    postgres:
      - name: warehouse
        host:
          valueFrom:
            kind: KubernetesPostgres
            name: warehouse-pg
            fieldPath: status.outputs.rw_service
        database: warehouse
        passwordSecret:
          secretName:
            valueFrom:
              kind: KubernetesPostgres
              name: warehouse-pg
              fieldPath: status.outputs.password_secret.name
```

The InfraPipeline deploys the PostgreSQL cluster first, then declares Trino against it -- the SAME parent for both fields: the host resolves the cluster's read-write Service (always the current primary) and the password its application-user Secret.

## Key Configuration

These are the most important decisions when configuring Trino. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Catalogs are the star, read this first** -- each catalog becomes one `<name>.properties` file and a queryable prefix (`SELECT ... FROM <catalog>.<schema>.<table>`). `postgres` and `mysql` rows compose naturally from other kinds (the FK defaults wire host and credential to the SAME composed database); the `custom` map takes raw properties (`connector.name=...`) for every other connector -- Iceberg, Hive, Kafka, anything. MySQL catalogs carry NO database segment (MySQL databases become Trino SCHEMAS). Names are unique across all three surfaces; `system` is always reserved, and `tpch`/`tpcds` while `sampleCatalogsEnabled` is on (unset = on).

**Credentials are SECRET-NATIVE** -- catalog passwords and the internal shared secret never appear in rendered ConfigMaps: properties reference environment variables (`${ENV:VAR}`, Trino's own secrets mechanism) and the variables arrive via Secret references. For custom-catalog credentials and exchange-manager S3 keys, declare `extraEnvFromSecret` entries and reference them the same way.

**The auth posture, verified live** -- Trino runs password authentication ONLY on secure requests; the often-suggested `allow-insecure-over-http` flag does NOT extend it to HTTP. The module sets `http-server.process-forwarded=true`: requests through a TLS-terminating proxy (composed exposure kinds) authenticate against the password file, and plain-HTTP data-plane requests are REFUSED outright (health probes unaffected). Terminate TLS at composed exposure kinds, or enable the `https` arm (a JKS keystore Secret). Disabling auth means anyone reaching the Service queries every catalog and impersonates any user.

**THE HEAP TRAP, verified live** -- the server validates `query.max-memory-per-node` against the JVM's ACTUAL max heap at boot and REFUSES TO START ("Heap size cannot be greater than maximum heap size"). The 1GB default already exceeds the heap of containers limited below ~1.7Gi at 60%. Prefer `jvm.maxHeapPercent` (the heap follows the container limit) and size the per-node ceiling comfortably below `limits.memory × maxHeapPercent`. Size the cluster-wide `maxQueryMemory` ("4GB" default) together with the per-node dials.

**Workers scale two ways, or not at all** -- no autoscaling arm = fixed-count workers (blank replicas = 2; an explicit 0 is the single-node shape). `hpa` scales on CPU/memory utilization (needs a metrics server AND worker resource requests). `keda` scales on Prometheus query metrics DOWN TO ZERO between queries -- typically on Trino's own `trino_execution_ClusterSizeMonitor_RequiredWorkers` metric, which exists only while `metrics.enabled` is on; triggers are REQUIRED (KEDA without triggers scales nothing). Enable `gracefulShutdown` with any autoscaler -- otherwise every scale-down kills in-flight queries.

**Fault-tolerant execution survives worker loss** -- `retryPolicy: TASK` (batch/ETL -- failed tasks retry individually, making aggressive scale-down and spot capacity safe) or `QUERY` (interactive). The exchange manager needs at least one durable spooling destination -- object-store URIs for production; a local path dies with its pod.

**Governance is three JSON documents** -- `accessControlRules` (who may access which catalogs/schemas/tables; empty = authenticated users see everything), `resourceGroupsConfig` (per-group concurrency/memory/queue budgets -- declaring it ENGAGES the 27-character name budget), and `sessionPropertiesConfig`. All three follow Trino's own file-provider schemas and ship verbatim; the server validates content at startup.

**The escape hatch, merged LAST** -- `helmValues` is raw chart values with Helm `-f` semantics for what the typed fields don't model (probes, lifecycle hooks, sidecars, configMounts). The module RE-PINS authentication wiring and the internal shared secret AFTER the merge -- the security posture cannot be silently disabled from here. NEVER secret material.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesPostgres** | `catalogs.postgres[].host` | `status.outputs.rw_service` |
| **KubernetesPostgres** | `catalogs.postgres[].passwordSecret.secretName` | `status.outputs.password_secret.name` |
| **KubernetesMysql** | `catalogs.mysql[].host` | `status.outputs.primary_service` |
| **KubernetesMysql** | `catalogs.mysql[].passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| **KubernetesKeda** (runtime prerequisite) | -- | the KEDA operator must run on the cluster when `workers.keda` is declared |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Trino runs in | Co-locating exposure kinds and catalog credential Secrets |
| `coordinator_service` | The coordinator Service (`<name>-trino-coordinator`) | The handle exposure kinds route to |
| `coordinator_endpoint` | In-cluster endpoint (`http://<coordinator_service>.<namespace>.svc.cluster.local:8080`) | What SQL clients, JDBC/ODBC and BI tools (a KubernetesSuperset) connect to |
| `admin_username` | The bootstrap admin username (empty when auth is off or bring-your-own) | Pairing with the credential below for clients |
| `admin_password_secret` | Secret + key holding the bootstrap admin password (empty when auth is off or bring-your-own) | Signing in; wiring BI tools and automation |
| `worker_service` | The worker Service (`<name>-trino-worker`) -- internal | Network-policy composition |
| `port_forward_command` | kubectl port-forward one-liner for the Trino Web UI | Reaching the UI from a workstation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**SQL playground** -- The two-line install: a coordinator and two workers, PASSWORD auth on with the module-generated admin, and the in-image samples answering queries immediately. Start from the **SQL playground** preset.

**Federated warehouse** -- The production posture: a composed KubernetesPostgres queryable through Trino (and JOIN-able against any other catalog), HPA-scaled workers that drain running queries before termination, percent-based JVM heaps that follow the container limits, and JMX metrics flowing to the Prometheus operator. Start from the **Federated warehouse** preset.

**Elastic batch/ETL** -- `faultTolerantExecution.retryPolicy: TASK` with an object-store exchange manager, KEDA scale-to-zero workers triggered on Trino's own RequiredWorkers metric, and graceful shutdown -- aggressive scale-down and spot capacity without failed queries.

## Works With

- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the natural postgres catalog; the `catalogs.postgres` foreign-key defaults point at it
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the natural mysql catalog; the `catalogs.mysql` foreign-key defaults point at it
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the deployment
- [**KEDA**](/infra-catalog/kubernetes-keda) -- the runtime prerequisite for event-driven worker autoscaling
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- the operator CRDs the ServiceMonitors need, and the Prometheus a KEDA trigger queries
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) -- S3-compatible spooling storage for fault-tolerant execution
- [**Apache Superset**](/infra-catalog/kubernetes-superset) -- a BI layer over the exported coordinator endpoint

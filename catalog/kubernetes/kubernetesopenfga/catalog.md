# OpenFGA

Deploys OpenFGA -- the CNCF authorization engine built on Google's Zanzibar model -- from the official `openfga` chart at `openfga.github.io/helm-charts`. The server is stateless: every store, authorization model, and relationship tuple lives in the datastore you declare (PostgreSQL recommended, MySQL supported, in-memory for evaluation only), so replicas scale Check and ListObjects throughput linearly.

Know the grain before you deploy: this component deploys the ENGINE, never the DATA. Stores, authorization models, and relationship tuples are managed through OpenFGA's own API against the exported endpoints -- the `fga` CLI, the SDKs, or the platform's OpenFgaStore / OpenFgaAuthorizationModel / OpenFgaRelationshipTuple resources composed against `api_http_endpoint`. And know the default posture: without an `authn` block the API is OPEN -- anyone who can reach the Service can read and write every store.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- the `openfga` chart, creating:
  - Deployment with the configured replicas (default 3; the in-memory engine pins the count to 1), each pod running the schema-migration init container (`openfga migrate` -- idempotent, gating every rollout on the datastore being reachable and schema-current; the chart's hook-Job migration is deliberately not used)
  - Kubernetes Service for cluster-internal access on ports 8080 (HTTP), 8081 (gRPC), and 2112 (metrics -- served by DEFAULT)
  - ServiceAccount carrying your workload-identity annotations
  - The connection URI rendered WITHOUT credentials -- username and password ride environment variables read from the password Secret via secretKeyRef
- **Authn keys Secret** -- created only when pre-shared keys are DECLARED in the spec (`<name>-authn-keys`, data key `keys`); the declared `$secret/<slug>` references are materialized at deploy time, and only the Secret NAME ever renders into values
- **HorizontalPodAutoscaler** -- created only when `hpa.enabled` is `true` (illegal with the in-memory engine -- the spec's own rule)
- **ServiceMonitor** -- created only when `metrics.serviceMonitorEnabled` is `true`; requires the Prometheus Operator CRDs on the cluster (the install FAILS without them)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

Deliberately absent: the chart's demo playground is ALWAYS disabled (upstream security default -- the server refuses to start with the playground combined with any authentication), and no ingress is created -- public exposure composes from Gateway API kinds.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A PostgreSQL or MySQL database** reachable from the cluster (in-memory needs nothing -- and keeps nothing). A KubernetesPostgres resource composes naturally: reference it and the host resolves to its read-write Service, the password to its operator-maintained `<cluster>-app` Secret. The database itself (e.g. `openfga`) must exist -- declare it at the Postgres bootstrap (`initdb`).
- **The password Secret in the SAME namespace as OpenFGA** -- a secretKeyRef reads only its own namespace's Secrets (a Kubernetes constraint): co-locate OpenFGA with its database or replicate the credential Secret.
- **Prometheus Operator CRDs** (only when enabling the ServiceMonitor) -- the install FAILS without them.
- **An OTLP gRPC collector** (only when enabling tracing) -- a KubernetesOtelCollector or KubernetesSignoz Service is the natural target.

## Deploy

### Console

Open the deployment store, find **OpenFGA**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **PostgreSQL datastore preset** in the [Presets](#presets) tab for the production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenFga
metadata:
  name: openfga
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "openfga"
  createNamespace: true
  replicas: 3
  datastore:
    postgres:
      host:
        value: postgres-rw.openfga.svc.cluster.local
      database: openfga
      username: openfga
      passwordSecret:
        secretName:
          value: my-postgres-app
  authn:
    preshared:
      existingKeysSecretName: openfga-api-keys
```

```shell
planton apply -f openfga.yaml
```

This deploys three stateless server replicas against a PostgreSQL datastore, with schema migrations running as an init container in every pod and the API guarded by pre-shared keys from a Secret you maintain. Metrics serve on port 2112 by default. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire OpenFGA to a KubernetesPostgres managed alongside it:

```yaml
spec:
  namespace:
    value: openfga
  createNamespace: true
  datastore:
    postgres:
      host:
        valueFrom:
          kind: KubernetesPostgres
          name: my-postgres
          fieldPath: status.outputs.rw_service
      database: openfga
      username: openfga
      passwordSecret:
        secretName:
          valueFrom:
            kind: KubernetesPostgres
            name: my-postgres
            fieldPath: status.outputs.password_secret.name
    migrationTimeout: 5m
```

The InfraPipeline deploys the Postgres cluster first, then provisions OpenFGA against its read-write Service and operator-maintained credential Secret. Widen `migrationTimeout` when both provision concurrently -- the init container retries until the database is up.

## Key Configuration

These are the most important decisions when configuring OpenFGA. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Datastore engine** -- `datastore.postgres`, `datastore.mysql`, and `datastore.memory` are one required choice (no default arm). PostgreSQL is the recommended production engine. `memory: {}` is the zero-dependency evaluation sandbox: all data lost on every restart, replicas pinned to 1 -- never for real use. Credentials never render: the URI ships credential-free and the password rides a secretKeyRef. Pool dials (`maxOpenConns`, default 30 per replica) multiply by replicas against the database's connection limit.

**Authentication** -- `authn.preshared` and `authn.oidc` are one choice, and NO choice means NO authentication -- an open API, fine on a lab cluster, never in production. Pre-shared keys go exactly one way: declare them (org-secret references materialized into the managed `<name>-authn-keys` Secret) OR point `existingKeysSecretName` at a Secret you maintain (comma-separated keys under the data key `keys` -- the chart's contract). OIDC requires BOTH `issuer` and `audience` -- the server hard-fails at startup with either empty (audience validation became mandatory as a security fix).

**Replicas and the HPA** -- `replicas` (default 3, range 1-50) scales reads linearly; the ceiling is the DATABASE. `hpa.enabled` hands the count to the autoscaler (min <= max -- the spec's own rule; the platform defaults max to 10 where the chart's own fallback is 100). Autoscaling on the memory engine is ILLEGAL -- pods would each hold their own divergent authorization world.

**Observability** -- Metrics are the unusual default-ON: `/metrics` on port 2112 unless explicitly disabled. `metrics.enableRpcHistograms` trades metric cardinality for the per-RPC latency visibility that sizes CPU against check volume. `tracing.enabled` requires `otlpEndpoint` (the spec's own rule); `sampleRatio` defaults to 0.2 -- 1.0 on a busy authorization path floods the collector. Log level `none` exists on the server but the chart's closed values schema rejects it -- the chart's vocabulary is law.

**Query tuning** -- Every `tuning` field maps to a server flag; empty keeps the server's own defaults. The check-query cache trades up to its TTL (default 10s) of authorization-change propagation delay for large Check throughput gains -- a REVOKED permission may keep answering allowed until its entry expires. `experimentals` accepts exactly four values (the chart's closed schema), and declaring ANY replaces the server's own default set -- which ships `pipeline_list_objects` enabled; re-include it unless you mean to turn it off.

**The escape hatch** -- `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics), but this chart's values schema is CLOSED (`additionalProperties: false`): only keys the chart already defines are legal. The real surface is `extraEnvVars` -- the ~50 `OPENFGA_*` server flags without values paths -- plus TLS file wiring and sidecars. The module re-pins `fullnameOverride` after the merge so the exported Service names stay stable. Never put secret material in this document.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesPostgres** (optional) | `datastore.postgres.host` | `status.outputs.rw_service` |
| **KubernetesPostgres** (optional) | `datastore.postgres.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| **KubernetesMysql** (optional) | `datastore.mysql.host` | `status.outputs.primary_service` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the server runs in | Locating the install for diagnostics |
| `service` | The OpenFGA Service name | General client traffic |
| `api_http_endpoint` | In-cluster HTTP API endpoint (port 8080) | SDKs, the `fga` CLI, and the platform's OpenFGA provider credential -- the seam OpenFgaStore / OpenFgaAuthorizationModel / OpenFgaRelationshipTuple resources compose against |
| `api_grpc_endpoint` | In-cluster gRPC API endpoint host:port (port 8081, plaintext) | gRPC SDKs |
| `authn_keys_secret_name` | The managed keys Secret (`<name>-authn-keys`) -- filled only when keys were DECLARED; empty when authn is unset or rides an existing Secret | Granting workloads read access to the API keys |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Workstation access |

Deliberately absent: stores, models, and tuples are API data the deployment never owns -- create them through the exported endpoints, not through this spec.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production with PostgreSQL** -- Three stateless replicas against a composed KubernetesPostgres (read-write Service + operator-maintained credential Secret), pre-shared-key authentication, metrics on. Start from the **PostgreSQL datastore preset**.

**Evaluation sandbox** -- `datastore.memory: {}`, one replica, no authentication: the zero-dependency shape for evaluating authorization models and API integration. Everything resets on every restart -- that is the point.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the OpenFGA install
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the recommended production datastore; the host and credential Secret both resolve by reference
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the alternative datastore engine
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- provides the Prometheus Operator CRDs the ServiceMonitor needs
- [**OpenTelemetry Collector**](/infra-catalog/kubernetes-otel-collector) -- the OTLP gRPC target for trace export
- [**OpenFGA Store**](/infra-catalog/openfga-store) -- creates stores as first-class resources against the exported HTTP endpoint

# Kubernetes Workloads and Their Cluster

Every resource whose kind starts with `Kubernetes` (`KubernetesIstio`,
`KubernetesCertManager`, `KubernetesDeployment`, `KubernetesGatewayApiCrds`,
…) deploys INTO a cluster, and it reaches that cluster through a **Kubernetes
provider connection** — a platform record carrying the cluster's endpoint and
credentials.

How a resource's connection resolves, in order:

1. Explicit spec value (`spec.provider_info.connection`) — rarely used in charts.
2. The `planton.dev/connection` **annotation** — the same-chart mechanism.
   Value is the org-scoped connection slug.
3. The environment's default Kubernetes connection.
4. The organization's default Kubernetes connection.

The platform keeps the DEFAULT arm loaded: when a cluster deploys through
Planton, its materialized connection is authorized org-wide and becomes the
organization's default Kubernetes connection if none exists yet. The same
happens when a user connects an existing cluster through the desktop. So the
connection question has exactly one decision, made before writing any
Kubernetes manifest: **is the cluster IN this chart, or not?**

## Scenario 1 — the chart creates the cluster too (one-run composition)

The platform automatically creates ("materializes") a Kubernetes provider
connection when a cluster resource (e.g. `AwsEksCluster`) finishes deploying —
before any dependent node is released. Here the workloads DO annotate: the
binding to the chart's own cluster must be deterministic, and the default
binding cannot promise that (the organization may already run another cluster
whose connection holds the default). The chart's job is to make the
connection's name and the workloads' annotation agree, and the pattern that
makes disagreement impossible is ONE values param used on BOTH ends:

```yaml
# values.yaml
params:
  - name: connection_name
    description: Name of the Kubernetes connection published for this cluster
    value: my-cluster

# templates/cluster.yaml — the PRODUCER declares the connection's name
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksCluster
metadata:
  name: "{{ values.env }}-cluster"
  annotations:
    planton.dev/connection-name: "{{ values.env }}-{{ values.connection_name }}"
spec:
  …

# templates/kubernetes/istio.yaml — every CONSUMER references the same expression
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstio
metadata:
  name: "{{ values.env }}-istio"
  annotations:
    planton.dev/connection: "{{ values.env }}-{{ values.connection_name }}"
  relationships:
    - kind: AwsEksCluster
      name: "{{ values.env }}-cluster"
      type: runs_on
spec:
  …
```

Rules of the pattern:

- `planton.dev/connection-name` (on the cluster) sets the published
  connection's name **literally** — nothing is prepended. Always embed
  `{{ values.env }}` in the value: connections are org-scoped, and two
  environments deploying the same chart must not collide. A collision with a
  connection the cluster did not create fails the deploy loudly.
- `planton.dev/connection` (on each workload) must render to the **identical
  string**. Using the same values expression on both ends guarantees it.
- Without the `connection-name` annotation, the platform publishes the
  connection at `<env>-<cluster resource metadata.name>` — you will see this
  legacy formula in older charts, where the workload annotations re-derive it
  (`"{{ values.env }}-{{ values.cluster_name }}"` against a cluster named
  without an env prefix). Prefer the explicit `connection-name` pattern in new
  charts; it composes cleanly with env-prefixed resource names.

## Scenario 2 — the cluster already exists

When the user deploys onto a cluster that is NOT in this chart, **write no
connection wiring at all** — no annotation, no connection param. The cluster's
connection (materialized when it deployed through Planton, or created when the
user connected it) is the organization's default Kubernetes binding, and the
platform resolves it for every Kubernetes resource that carries no annotation.
A `cluster_connection`-style param here is invented ceremony: it makes the
user hand-carry a value the platform already knows — the exact failure class
the references-before-params check exists to prevent, in connection form.

In this scenario the workloads also need no `runs_on` relationship to a
cluster — the cluster is not in this chart's graph.

**The multi-cluster exception, by name.** The default binding points at ONE
cluster. When the organization runs several and the user's words name a
specific one that may not be the default ("deploy this to the staging
cluster"), ground the intended cluster's connection slug with the CLI (list
the org's Kubernetes provider connections) and annotate the workloads with
that literal slug — an explicit, grounded choice, never a param the user must
fill:

```yaml
# only when the org runs multiple clusters AND the target is not the default:
metadata:
  annotations:
    planton.dev/connection: staging-cluster   # grounded via the CLI, never guessed
```

## Ordering — a separate concern from credentials

`metadata.relationships` entries create REAL dependency edges in the deploy
graph: a resource with a `runs_on` or `depends_on` relationship does not start
until its target succeeds. The connection annotation carries the credentials,
and the platform reads it for placement too: a workload whose annotation names
the connection the cluster publishes draws inside that cluster and waits for
the cluster itself. It does not wait for the nodes, so both are still required
in Scenario 1:

- **Every Kubernetes workload gets a `runs_on` relationship** to the in-chart
  cluster resource (`kind` + the cluster's exact `metadata.name` expression).
- **When the chart has a node group** (e.g. `AwsEksNodeGroup`), anything that
  schedules pods (operators, deployments, Helm-backed addons) should `runs_on`
  the NODE GROUP instead — a cluster with zero nodes accepts the install but
  pods never start and the deploy times out. CRD-only kinds (e.g.
  `KubernetesGatewayApiCrds`) may target the cluster directly.
- **Chain the Kubernetes resources among themselves with `depends_on`** in
  their real order: CRDs → the operator that serves them → the instances that
  use them. Example: `KubernetesGatewayApiCrds` ← `KubernetesIstio` ←
  `KubernetesGatewayClass`.

## Never render a connection manifest inside a chart

Charts render cloud resources only. A `KubernetesProviderConnection` document
in templates fails the build ("UNSUPPORTED CLOUD RESOURCE KIND") — connections
are org-scoped, authorization-bearing records the platform materializes or
users create; they never belong in templates.

## Self-check before finishing any chart with Kubernetes kinds

1. The cluster is IN this chart → every `Kubernetes*` resource carries
   `planton.dev/connection`, the cluster carries `planton.dev/connection-name`,
   and both ends use the same values expression.
2. The cluster is NOT in this chart → no connection annotation and no
   connection param anywhere (unless the multi-cluster exception applied, in
   which case the annotation carries a CLI-grounded literal slug).
3. Every workload has a `runs_on` relationship (node group when pods are
   scheduled and one exists; cluster otherwise) — same-chart clusters only.
4. Kubernetes resources are `depends_on`-chained in dependency order
   (CRDs → operators → instances).
5. `planton chart build` is green — but note the build cannot verify
   connection wiring at all; this self-check is what protects the deploy.

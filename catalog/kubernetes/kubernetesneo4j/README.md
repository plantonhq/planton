# Kubernetes Neo4j

## When NOT to Use This

**One resource is ONE Neo4j server.** The chart's own grain is a
StatefulSet of exactly one pod, and this component preserves it. The
community edition (the default) is single-instance by license;
Enterprise clustering is built by installing MULTIPLE KubernetesNeo4j
resources that share the same `cluster_name` — each member is its own
first-class resource, not a replicas knob.

Also not the right component when:

- **You want a replicas field** — there is none, by design. Community
  cannot cluster (the spec rejects `cluster_name` on community), and
  Enterprise members are separate resources sharing `cluster_name`,
  each requiring `accept_license_agreement: true` and a valid Neo4j
  license.
- **You want a managed graph database** — use a managed cloud service;
  this component is for running Neo4j ON the Kubernetes cluster
  itself.
- **You expect a public endpoint out of the box** — the chart's
  default LoadBalancer Service is DELIBERATELY overridden to
  ClusterIP here. Exposure composes from first-class kinds
  (KubernetesIngress, Gateway API kinds) over the exported service
  handle, or set `service.type: LoadBalancer` explicitly.
- **You need resources below the chart's floor** — the chart REJECTS
  installs under 500m CPU / 2Gi memory; the module never defaults
  below it (empty resources = the chart's own defaults).

## Overview

**KubernetesNeo4j** deploys a Neo4j graph database — the standard
engine for knowledge graphs, GraphRAG and agent-memory architectures —
from the official `neo4j` Helm chart (https://helm.neo4j.com/neo4j).
Graph databases carry the relationship-shaped data those AI
architectures depend on: a knowledge graph the retrieval layer walks,
an agent's long-lived memory of entities and their connections —
workloads where the JOIN-heavy relational alternative degrades and a
native graph engine does not.

**The credential contract — minted, never required**: leave `auth`
empty and the modules generate the `neo4j` admin user's password,
materialize it as the `<name>-auth` Kubernetes Secret carrying the
chart's contract (`NEO4J_AUTH`, value `neo4j/<password>`) plus a bare
`password` key, and point the chart at it via `passwordFromSecret`.
The chart looks that Secret up AT TEMPLATE TIME, so the modules create
it BEFORE the release — and the password never lands in rendered Helm
values; `auth_secret_name` and `password_secret` tell workloads where
to read it, and nobody escrows a value the module can mint. Declare
`auth.password` (a managed-secret reference) to bring your own, or
`auth.existing_secret` to reference a Secret you own that already
carries the `NEO4J_AUTH` key.

**Key design points:**

- **ClusterIP by design.** The chart ships its exposure Service
  (`<neo4j.name>-lb-neo4j`; bolt 7687, http 7474, https 7473) as type
  LoadBalancer, which would provision a cloud load balancer (or hang
  Pending) on every install. This component pins it to ClusterIP
  unless `service.type` says otherwise. In-cluster clients use the
  always-created default Service (= the resource name — the endpoints
  in the stack outputs).
- **TLS has a key-name bridge.** The chart mounts `private.key` and
  `public.crt` from each `ssl` scope's Secret (its subPath defaults);
  cert-manager-issued Secrets carry `tls.key`/`tls.crt` instead. The
  module does NOT silently rewrite key names — when wiring a
  KubernetesCertificate's Secret, mirror its data into a Secret with
  the chart's expected keys (or produce one with those keys directly).
- **Memory is typed.** The `memory` block renders the three neo4j.conf
  memory keys (heap initial/max, page cache); typed keys WIN over
  duplicates in the free-form `config` map. Empty = Neo4j
  auto-computes from the container memory (the chart default — usually
  right).
- **Licensing posture.** The chart itself is Apache-2.0; the community
  engine is GPLv3 — referenced as the official image, never
  distributed by this component. Enterprise requires a commercial
  license and `accept_license_agreement: true` (the spec enforces it).
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines); for the chart surface beyond
  the typed fields (log4j XML, extra volumes, LDAP secrets, the
  operations sidecar, probes, PDB), never a substitute for them.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `2026.6.0` — chart
  versions track Neo4j calendar releases)
- **`spec.edition`**: `community` (default, single-instance) or
  `enterprise` (clustering and advanced features; requires
  `accept_license_agreement` and a valid license)
- **`spec.auth`**: optional. Empty = the module generates the admin
  password into the `<name>-auth` Secret; `password` (a managed-secret
  reference, materialized the same way) or `existing_secret` (must
  already carry `NEO4J_AUTH: neo4j/<password>` and exist before the
  install) to bring your own
- **`spec.cluster_name`**: Enterprise members sharing this name form
  one cluster; empty = standalone (and always standalone on community)
- **`spec.resources`**: chart minimum 500m CPU / 2Gi memory — the
  chart rejects installs below its floor; empty = the chart defaults
- **`spec.data_volume`**: size (default 10Gi) and StorageClass
  (literal or a KubernetesStorageClass reference; empty = the cluster
  default)
- **`spec.memory`**: `heap_initial` / `heap_max` (keep them equal in
  production) / `page_cache` (what remains after heap and OS overhead)
- **`spec.config` / `spec.apoc_config`**: extra neo4j.conf and
  apoc.conf entries (memory keys belong in `memory`; auth/TLS keys are
  chart-owned)
- **`spec.service`**: type (ClusterIP default — NOT the chart's
  LoadBalancer default), annotations for cloud LB recipes
- **`spec.ssl`**: per-scope TLS (`bolt`, `https`) from existing
  Secrets carrying `private.key`/`public.crt` — see the key-name
  bridge above
- **`spec.scheduling`**: node selector, tolerations, pod anti-affinity
  (chart default true — meaningful for Enterprise cluster members),
  priority class
- **`spec.service_monitor_enabled`**: Prometheus ServiceMonitor
  (requires the Prometheus Operator CRDs)
- **`spec.additional_jvm_arguments` / `spec.use_default_jvm_arguments`
  / `spec.image` / `spec.helm_values`**: JVM tuning, the air-gap image
  path, and the escape hatch

## Stack Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the server runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service_name` | The main Neo4j Service (bolt/http/https ports; = the release name) |
| `bolt_endpoint` | In-cluster bolt endpoint drivers connect to (`neo4j://<name>.<ns>.svc.cluster.local:7687`) |
| `http_endpoint` | In-cluster HTTP API / Browser endpoint (port 7474) |
| `auth_secret_name` | The Secret holding the admin credentials in the chart's contract (`NEO4J_AUTH`) — the module-materialized `<name>-auth` (declared or generated password), or the referenced existing Secret; always set |
| `port_forward_command` | Port-forward command for bolt access from a workstation |
| `password_secret` | `{name, key}` of the admin user's bare password in the module-materialized `<name>-auth` Secret (key `password`); unset when `existing_secret` is declared, since that Secret's layout is yours |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`);
  **`data_volume.storage_class`** references a KubernetesStorageClass;
  the **`ssl` scope Secrets** accept a KubernetesCertificate reference
  — with the key-name bridge above.
- **Applications consume the outputs**: `bolt_endpoint` as the driver
  URI, `password_secret` for the bare password (or `auth_secret_name`
  when a workload wants the chart's `neo4j/<password>` pair) — the
  password rides the Secret, never the manifest, whether it was
  declared or module-generated.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route over `service_name` for HTTP/Browser; bolt is a TCP
  protocol — a TCP route or an explicit `service.type: LoadBalancer`
  with cloud annotations for external drivers.
- **Enterprise clusters are graphs of resources**: N KubernetesNeo4j
  resources sharing `cluster_name`, each with
  `accept_license_agreement: true` — every member individually
  addressable, sized, and upgraded.

## Examples

### Development (community, module-minted password)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNeo4j
metadata:
  name: dev-graph
spec:
  namespace:
    value: dev-graph
  create_namespace: true
  # no auth: the module mints the admin password into dev-graph-auth
```

### Production single instance (sized, tuned memory)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNeo4j
metadata:
  name: knowledge-graph
spec:
  namespace:
    value: graph
  create_namespace: true
  auth:
    password: $secret/knowledge-graph-admin # bring your own, as a managed-secret reference
  resources:
    requests: { cpu: "2", memory: 8Gi }
    limits: { cpu: "4", memory: 8Gi }
  data_volume:
    size: 100Gi
  memory:
    heap_initial: 3G
    heap_max: 3G
    page_cache: 4G
  service_monitor_enabled: true
```

### Enterprise cluster member (one of three)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNeo4j
metadata:
  name: graph-core-1
spec:
  namespace:
    value: graph
  edition: enterprise
  accept_license_agreement: true
  cluster_name: graph-core
  auth:
    existing_secret: graph-core-auth # NEO4J_AUTH: neo4j/<password>
  resources:
    requests: { cpu: "2", memory: 8Gi }
    limits: { cpu: "4", memory: 8Gi }
  data_volume:
    size: 200Gi
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

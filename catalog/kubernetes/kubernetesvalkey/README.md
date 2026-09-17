# Kubernetes Valkey

## When NOT to Use This

**Automated failover is not part of this component.** The official
Valkey chart does not ship Sentinel or Cluster mode yet — they are
upstream chart milestones — so neither is modeled here. In replication
mode, if the primary pod dies, Kubernetes restarts it and the replicas
re-attach; no replica is promoted. Durability through a primary restart
comes from PERSISTENCE (the append-only file on a volume), not from
promotion.

Also not the right component when:

- **You need automated failover or horizontal sharding** — Sentinel HA
  and Cluster mode are deliberately absent because the upstream chart
  does not ship them. When failover is non-negotiable today, use a
  managed cache from the host cloud's kinds, or compose an
  operator-based topology through KubernetesManifest.
- **You want external exposure baked in** — no ingress block exists.
  The store is in-cluster plumbing reachable at the exported
  `kube_endpoint`; to reach it from outside, compose a first-class
  exposure kind against the exported service names. (The service
  block's type/annotations exist for the LoadBalancer arm of
  managed-cloud recipes — a Service knob, not a hostname/DNS story.)
- **You need Redis itself** — Valkey is the Linux Foundation's
  BSD-licensed fork of Redis 7.2 and a drop-in replacement for Redis
  clients; if you specifically need Redis-branded releases or modules
  beyond that surface, this component does not provide them.

## Overview

**KubernetesValkey** deploys Valkey — the Linux Foundation's
Redis-compatible in-memory data store (the open-source successor every
Redis client library speaks natively) — from the OFFICIAL Valkey Helm
chart (`valkey` at https://valkey.io/valkey-helm/, pinned 0.11.0),
not Bitnami.

**Two topologies**, chosen by the presence of the `replication` block:

- **Standalone** (default): one instance backed by a Deployment, with
  an optional persistent volume. The right shape for caches and
  development.
- **Primary/replica**: one primary plus N replicas from a StatefulSet,
  with streaming replication and a dedicated read Service
  (`<name>-read`). Persistence is REQUIRED in this mode — an ephemeral
  primary would replicate an empty dataset after every restart.

**The naming contract**: the Helm release and chart fullname are pinned
to `metadata.name`, so several Valkey instances coexist in one cluster.
The chart renders `<name>` (the write Service), and — in replication
mode — `<name>-read` (load balances reads across all pods) and
`<name>-headless` (direct pod discovery; the standalone Deployment
renders no headless Service).

**Key design points:**

- **Authentication is declared, never defaulted; passwords are minted,
  never required** — the chart ships with auth OFF (anyone who can reach
  the Service can read and write). Declaring ACL users turns it on; the
  user list MUST include `default` (otherwise unauthenticated clients
  keep full access). A user needs only a name: the module generates a
  password for every user declared without one and keeps it stable
  across re-applies; a declared password (a managed-secret reference)
  is used as given. Either way the credentials materialize as the
  `<name>-auth` Kubernetes Secret (one key per username) — never
  plaintext in rendered chart values — and `password_secret` tells
  clients where to read the `default` user's.
- **Durability is module-owned** — Valkey's persistence and memory
  directives live in valkey.conf, which the chart accepts only as one
  raw string. The typed `config` block (`append_only`,
  `rdb_save_points`, `max_memory`, `max_memory_policy`, ...) renders
  that string deterministically on both engines; `extra_directives`
  appends anything beyond the typed fields.
- **TLS rides an existing certificate** — `tls.certificate_secret`
  names a kubernetes.io/tls Secret, the cert-manager seam: issue a
  KubernetesCertificate and wire its secret here.
- **Exposure is composed, never embedded** — no ingress block exists in
  the spec; external reachability is a separate first-class kind wired
  against the exported service names.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines); a safety valve, never the
  primary interface.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to deploy into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `0.11.0`, which ships
  Valkey 9.1.1 — chart and app versions move separately; the chart pin
  governs)
- **`spec.replication`**: switches to primary/replica — `replicas`
  (default 2; total pods = replicas + 1 primary), REQUIRED
  `persistence`, `min_replicas_to_write` (set 1+ so a partitioned
  primary stops accepting writes replicas would never see),
  `min_replicas_max_lag`, `diskless_sync`, and the `read_service`
  block (enabled by default)
- **`spec.persistence`**: standalone-mode storage — `size`, optional
  `storage_class` (literal or a KubernetesStorageClass reference), and
  `keep_on_uninstall`. Omitted = a pure in-memory cache that starts
  empty after every pod restart. In replication mode, persistence is
  declared INSIDE the replication block instead.
- **`spec.config`**: the valkey.conf surface — `append_only` (the
  durability posture that makes a pod restart lossless when paired
  with a volume), `rdb_save_points` XOR `snapshots_disabled`,
  `max_memory` (always set this for caches — empty means the pod's
  memory limit is the only bound, and hitting it is an OOM kill, not
  an eviction), `max_memory_policy` (`noeviction` default — right for
  durable stores; `allkeys-lru` and friends for caches), and
  `extra_directives`
- **`spec.auth`**: ACL users (name; an optional password — empty means
  the module generates one; `permissions` rule string — empty means
  full access); must include the `default` user
- **`spec.tls`**: `enabled` + `certificate_secret` (a
  KubernetesCertificate reference or a literal kubernetes.io/tls
  Secret name), optional `require_client_certificate` for mutual TLS
- **`spec.service`**: the write Service — type (ClusterIP default),
  port (6379 default), per-cloud LoadBalancer annotations
- **`spec.resources`**: CPU/memory for the Valkey container — size
  memory ABOVE `config.max_memory`; Valkey needs headroom for
  replication buffers and fork-based persistence
- **`spec.metrics`**: the redis_exporter sidecar the chart ships;
  `service_monitor_enabled` additionally renders a ServiceMonitor
  (requires the Prometheus operator CRDs — the release fails to
  install without them)
- **`spec.scheduling`**: node selector, tolerations, priority class
- **`spec.pod_disruption_budget`**: replication mode only (the chart
  renders no PDB for a single standalone pod) — one of
  `max_unavailable` / `min_available`
- **`spec.log_level`** / **`spec.image`** /
  **`spec.image_pull_secrets`** / **`spec.helm_values`**

## Stack Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the instance runs in |
| `service` | The write Service (`<name>`) — targets the primary in replication mode, the one instance standalone |
| `read_service` | The read Service (`<name>-read`) — replication mode with the read Service enabled; empty otherwise |
| `headless_service` | Direct pod discovery (`<name>-headless`) — replication mode only; empty standalone |
| `kube_endpoint` | In-cluster endpoint of the write Service (`<name>.<namespace>.svc.cluster.local:<port>`) |
| `port_forward_command` | Port-forward command for workstation access when no exposure is composed |
| `username` | The ACL username applications authenticate with (`default` when auth is declared; empty when auth is off) |
| `password_secret` | `{name, key}` of that user's password in the module-materialized `<name>-auth` Secret, whether declared or module-generated; unset when auth is off |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`);
  **`persistence.storage_class`** references a KubernetesStorageClass
  (`status.outputs.storage_class_name`);
  **`tls.certificate_secret`** references a KubernetesCertificate
  (`status.outputs.secret_name`) — the cert-manager seam.
- **Applications consume the outputs**: `kube_endpoint` as the
  connection host, `username` / `password_secret` as env-from
  references — the credential rides the module-materialized Secret,
  never the manifest.
- **Exposure composes, never embeds**: a KubernetesService of type
  LoadBalancer (or a TCP route on a Gateway) targets the `service`
  output.
- **Auth off means open access** for anyone with network reach —
  acceptable only behind a KubernetesNetworkPolicy or for development;
  declare users everywhere else.

## Examples

### Development cache (standalone, in-memory)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesValkey
metadata:
  name: dev-cache
spec:
  namespace:
    value: dev-cache
  create_namespace: true
  config:
    snapshots_disabled: true
    max_memory: 256mb
    max_memory_policy: allkeys-lru
```

### Durable standalone (persistence, AOF, auth)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesValkey
metadata:
  name: jobs-store
spec:
  namespace:
    value: jobs
  create_namespace: true
  persistence:
    size: 5Gi
  config:
    append_only: true
    max_memory: 512mb
  auth:
    users:
      - name: default # no password: the module mints one into jobs-store-auth
```

### Production replication (read Service, write safety, metrics)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesValkey
metadata:
  name: sessions
spec:
  namespace:
    value: sessions
  create_namespace: true
  replication:
    replicas: 2
    persistence:
      size: 10Gi
    min_replicas_to_write: 1 # a partitioned primary refuses writes
  config:
    append_only: true
    max_memory: 1gb
  auth:
    users:
      - name: default # module-generated
      - name: app
        password: $secret/sessions-app-password # bring your own, as a managed-secret reference
        permissions: "~* -@all +@read +@write +ping +info"
  resources:
    requests:
      cpu: 500m
      memory: 1536Mi # headroom above max_memory
    limits:
      cpu: "1"
      memory: 2Gi
  metrics:
    enabled: true
  pod_disruption_budget:
    enabled: true
    max_unavailable: 1
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

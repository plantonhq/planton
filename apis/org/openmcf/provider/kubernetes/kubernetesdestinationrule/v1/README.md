# Kubernetes Destination Rule

Provision an Istio `DestinationRule` -- a namespaced resource that configures what
happens to traffic **after** routing has selected a destination service: the load
balancing algorithm, connection-pool limits, circuit breaking / outlier detection, and
the TLS settings the sidecar uses when originating upstream connections. It can also
split a service into named `subsets` (e.g. versions) that route rules target.

A DestinationRule does not, by itself, route traffic -- it tunes how the mesh talks to a
host that is already in the registry (a Kubernetes `Service` or a `ServiceEntry`).

## What Gets Created

- A namespaced `networking.istio.io/v1` `DestinationRule` custom resource.
- `host` (required), plus an optional `traffic_policy`, `subsets`, `export_to`, and
  `workload_selector`, scoped to this resource's namespace.

## How it relates to the other Istio networking resources

- **ServiceEntry** *adds* a destination to the registry (what can be reached).
- **DestinationRule** configures *how* to talk to a destination already in the registry
  (load balancing, TLS origination, connection pools, circuit breaking).
- **VirtualService** routes requests *to* a host/subset; subset policies only take effect
  once a route sends traffic to the subset.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to apply the policy. The
  resource applies with only the CRDs present, but it only affects traffic where istiod
  and a sidecar (or ambient) data plane are running.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDestinationRule
metadata:
  name: reviews-lb
spec:
  namespace:
    value: bookinfo
  host: reviews.bookinfo.svc.cluster.local
  traffic_policy:
    load_balancer:
      simple: LEAST_REQUEST
```

```bash
openmcf apply -f destinationrule.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the DestinationRule is created in. |
| `host` | string | Registry host the rule applies to. For short names, istiod resolves relative to this rule's namespace -- prefer FQDNs. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `traffic_policy` | object | Load balancing, connection pool, outlier detection, TLS, per-port overrides, tunnel, PROXY protocol. |
| `subsets` | list | Named subsets (`name`, `labels`, per-subset `traffic_policy`). |
| `export_to` | list | Namespaces the rule is visible to (`.` = same namespace, `*` = all). Default all. |
| `workload_selector.match_labels` | map | Pods/VMs (by label) the rule applies to. Matched by istiod; not a foreign key. |

### `traffic_policy`

| Field | Type | Description |
|-------|------|-------------|
| `load_balancer` | object | `simple` (`LEAST_CONN`/`RANDOM`/`PASSTHROUGH`/`ROUND_ROBIN`/`LEAST_REQUEST`) **or** `consistent_hash`; plus `locality_lb_setting` and `warmup`. |
| `connection_pool` | object | `tcp` (max connections, timeouts, keepalive) and `http` (pending/active requests, retries, h2 upgrade). |
| `outlier_detection` | object | Circuit breaking: consecutive 5xx / gateway / local-origin errors, sweep `interval`, `base_ejection_time`, ejection/health percentages. |
| `tls` | object | Client TLS the sidecar originates: `mode`, cert paths or `credential_name`, `sni`, SANs, CRL. |
| `port_level_settings` | list | Per-port overrides (`port.number` + the four blocks above). Fully overrides the destination-level settings for that port. |
| `tunnel` | object | Tunnel TCP/TLS over `CONNECT`/`POST` to a `target_host:target_port`. |
| `proxy_protocol` | object | Upstream PROXY protocol `version` (`V1`/`V2`). |

Durations (`connect_timeout`, `interval`, `base_ejection_time`, cookie `ttl`, ...) are
strings like `10s`, `1.5s`, `2h45m`; the ones upstream documents as ">= 1ms" are
validated as such.

## Examples

### Load balancing + circuit breaking

```yaml
spec:
  namespace:
    value: bookinfo
  host: reviews.bookinfo.svc.cluster.local
  traffic_policy:
    load_balancer:
      simple: LEAST_REQUEST
    connection_pool:
      tcp:
        max_connections: 100
      http:
        http2_max_requests: 1000
        max_requests_per_connection: 10
    outlier_detection:
      consecutive_5xx_errors: 7
      interval: 5m
      base_ejection_time: 15m
```

### mTLS origination to an egress host via a credential

```yaml
spec:
  namespace:
    value: egress
  host: external-db.example.com
  workload_selector:
    match_labels:
      app: db-client
  traffic_policy:
    tls:
      mode: MUTUAL
      credential_name: db-client-cert
      sni: external-db.example.com
```

### Subsets for a canary

```yaml
spec:
  namespace:
    value: bookinfo
  host: reviews.bookinfo.svc.cluster.local
  subsets:
    - name: v1
      labels:
        version: v1
    - name: v2
      labels:
        version: v2
      traffic_policy:
        load_balancer:
          simple: ROUND_ROBIN
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace` output
via `valueFrom`. The `host`, `workload_selector.match_labels`, and TLS `credential_name`
are **not** foreign keys -- istiod/Envoy resolve them at runtime, so they create no
automatic DAG edge. To order this DestinationRule relative to the resources it depends on
(the service it configures, the secret holding its client certs, or the workloads it
selects), declare the dependency on `metadata.relationships`:

```yaml
metadata:
  name: db-mtls
  relationships:
    - kind: KubernetesSecret
      name: db-client-cert
      type: uses
    - kind: KubernetesService
      name: external-db
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: egress-ns
      fieldPath: spec.name
  host: external-db.example.com
  traffic_policy:
    tls:
      mode: MUTUAL
      credential_name: db-client-cert
```

See `docs/README.md` for the full composability rationale.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `destination_rule_name` | Name of the created DestinationRule (equals metadata.name). |
| `namespace` | Namespace the DestinationRule was created in. |

## Related Components

- [Kubernetes Service Entry](kubernetesserviceentry)
- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)

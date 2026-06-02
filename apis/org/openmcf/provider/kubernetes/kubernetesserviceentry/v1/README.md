# Kubernetes Service Entry

Provision an Istio `ServiceEntry` -- a namespaced resource that adds an entry into
Istio's internal service registry. Use it to make a service that the mesh does not
otherwise know about (an external API reached over the public internet, a SaaS
endpoint, or a VM/legacy service) a first-class destination: mesh workloads can then
route to it, apply traffic policy and telemetry against it, and (for MESH_INTERNAL
services) treat it as part of the mesh.

## What Gets Created

- A namespaced `networking.istio.io/v1` `ServiceEntry` custom resource.
- `hosts` (required), plus an optional combination of `addresses`, `ports`,
  `location`, `resolution`, and either static `endpoints` **or** a
  `workload_selector`, scoped to this resource's namespace and exported per
  `export_to`.

## How endpoints are resolved

`resolution` controls how the proxy turns a host into a destination IP:

- **NONE** -- forward to the connection's original destination IP (no endpoints).
- **STATIC** -- use the IPs listed in `endpoints`.
- **DNS** -- resolve the hosts (or endpoint addresses) via DNS, asynchronously.
- **DNS_ROUND_ROBIN** -- like DNS, but pins to the first resolved IP per new
  connection (best for large web-scale services); at most one endpoint.

`endpoints` (static IPs/hosts) and `workload_selector` (in-mesh pods/VMs by label)
are mutually exclusive. `endpoints` cannot be used with NONE resolution. CIDR
`addresses` are honored only with NONE or STATIC resolution.

## ServiceEntry vs the other Istio networking resources

- **ServiceEntry** *adds* a destination to the registry (what can be reached).
- **DestinationRule** configures *how* to talk to a destination (load balancing, TLS,
  connection pools) once it is in the registry.
- **Sidecar** limits which registry entries a workload can see.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to program the registry.
  The resource applies with only the CRDs present, but routing requires istiod and a
  sidecar (or ambient) data plane.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesServiceEntry
metadata:
  name: external-payments-api
spec:
  namespace:
    value: payments
  hosts:
    - api.stripe.com
  location: MESH_EXTERNAL
  resolution: DNS
  ports:
    - number: 443
      name: https
      protocol: TLS
```

```bash
openmcf apply -f serviceentry.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the ServiceEntry is created in. |
| `hosts` | list | One or more hosts the entry matches (HTTP authority, TLS SNI, or DNS name). At least one; a bare `*` is not allowed. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `addresses` | list | Virtual IPs or CIDR prefixes. CIDR is allowed only with NONE/STATIC resolution. |
| `ports` | list | Exposed ports (`number`, `name`, `protocol`, `target_port`). `name` and `number` must each be unique. |
| `location` | string | `MESH_EXTERNAL` (default) or `MESH_INTERNAL`. |
| `resolution` | string | `NONE` (default), `STATIC`, `DNS`, or `DNS_ROUND_ROBIN`. |
| `endpoints` | list | Static backing endpoints (see below). Mutually exclusive with `workload_selector`; not allowed with NONE; at most one with DNS_ROUND_ROBIN. |
| `export_to` | list | Namespaces the service is visible to (`.` = same namespace, `*` = all). Default all. |
| `subject_alt_names` | list | SANs verified on the server certificate when originating TLS. |
| `workload_selector.labels` | map | In-mesh pod/VM labels backing the service (MESH_INTERNAL). Mutually exclusive with `endpoints`. |

### Port fields

| Field | Type | Description |
|-------|------|-------------|
| `number` | int | **Required.** Service port, 1-65535. |
| `name` | string | **Required.** Unique port label. |
| `protocol` | string | One of `HTTP`, `HTTPS`, `GRPC`, `HTTP2`, `MONGO`, `TCP`, `TLS`. |
| `target_port` | int | Endpoint receive port, 1-65535 (defaults to `number`). |

### Endpoint fields

| Field | Type | Description |
|-------|------|-------------|
| `address` | string | Endpoint IP, DNS name, or `unix://` socket. Required unless `network` is set; a `unix://` address may not carry ports. |
| `ports` | map | Service-port-name -> endpoint port (1-65535). Not for `unix://` addresses. |
| `labels` | map | Labels on the endpoint. |
| `network` | string | L3 network ID for multi-network meshes; required when `address` is empty. |
| `locality` | string | Failure-domain locality (e.g. `us-west2/us-west2-a`). |
| `weight` | int | Load-balancing weight. |
| `service_account` | string | Workload service account (same namespace). |

## Examples

### External HTTPS API over DNS

```yaml
spec:
  namespace:
    value: payments
  hosts:
    - api.stripe.com
  location: MESH_EXTERNAL
  resolution: DNS
  ports:
    - number: 443
      name: https
      protocol: TLS
```

### Static external endpoints

```yaml
spec:
  namespace:
    value: data
  hosts:
    - legacy-db.internal.example.com
  addresses:
    - 10.10.0.0/24
  location: MESH_INTERNAL
  resolution: STATIC
  ports:
    - number: 5432
      name: tcp-postgres
      protocol: TCP
  endpoints:
    - address: 10.10.0.5
      ports:
        tcp-postgres: 5432
    - address: 10.10.0.6
      ports:
        tcp-postgres: 5432
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace`
output via `valueFrom`. Neither `workload_selector.labels` nor the `hosts` /
`addresses` values are foreign keys -- istiod resolves them at runtime, so they
create no automatic DAG edge to any workload or service. To order this ServiceEntry
relative to the workloads it fronts (MESH_INTERNAL) in an infra chart, declare the
dependency on `metadata.relationships`:

```yaml
metadata:
  name: legacy-vm-service
  relationships:
    - kind: KubernetesDeployment
      name: legacy-app
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: data-ns
      fieldPath: spec.name
  hosts:
    - legacy.mesh.internal
  location: MESH_INTERNAL
  resolution: STATIC
  workload_selector:
    labels:
      app: legacy-app
```

See `docs/README.md` for the full composability rationale.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `service_entry_name` | Name of the created ServiceEntry (equals metadata.name). |
| `namespace` | Namespace the ServiceEntry was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)

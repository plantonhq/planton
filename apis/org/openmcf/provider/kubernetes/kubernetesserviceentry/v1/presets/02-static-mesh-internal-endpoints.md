# Bring Static Endpoints Into the Mesh

Register a service that has a fixed set of backing IPs (a VM-hosted database, a legacy
service, an appliance) as a MESH_INTERNAL destination with STATIC resolution. Mesh
workloads then reach it by hostname, and istiod load-balances across the explicit
endpoint IPs you supply.

## When to Use

- You run a service on VMs or fixed IPs (not Kubernetes pods) and want mesh workloads
  to address it by a stable hostname with explicit endpoints.
- You are expanding the mesh to include unmanaged infrastructure and want it treated as
  internal (so mesh policy and telemetry apply).

## Key Configuration Choices

- **`location: MESH_INTERNAL`** -- the service is part of the mesh; mesh policy and
  telemetry apply to it.
- **`resolution: STATIC`** -- istiod uses the IPs in `endpoints` directly (no DNS).
- **`endpoints`** -- the explicit backing IPs. Each maps the service port name
  (`tcp-postgres`) to the port on that endpoint. `endpoints` and `workload_selector`
  are mutually exclusive; STATIC requires `endpoints` (NONE forbids them).
- **`addresses`** -- optional virtual IP/CIDR the service is reached at. CIDR is allowed
  here because resolution is STATIC.

To instead back the service by in-mesh pods/VMs selected by label, drop `endpoints` and
set `workload_selector.labels` (the two are mutually exclusive).

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the service is registered in (e.g. `data`). |
| `<internal-host>` | The hostname workloads use (e.g. `legacy-db.mesh.internal`). |
| `<cidr-or-ip>` | Virtual IP or CIDR the service is reached at (e.g. `10.10.0.0/24`). |
| `<endpoint-ip-1>` / `<endpoint-ip-2>` | The backing endpoint IPs (e.g. `10.10.0.5`). |

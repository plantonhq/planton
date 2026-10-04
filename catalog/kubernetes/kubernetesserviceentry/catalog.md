# Istio Service Entry

Defines an Istio ServiceEntry: a namespaced resource that adds a host into Istio's internal service registry. Use it to make a service that platform service discovery does not know about -- an external API reached over the public internet, a SaaS endpoint, or a VM/legacy service -- routable, observable, and policy-addressable from inside the mesh by its hostname. A ServiceEntry teaches the mesh how to reach OUT to a service; it is not an ingress.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A ServiceEntry** -- a namespaced Istio registry entry describing the service's hosts, ports (with protocols), location (external to the mesh or part of it), how its endpoints are resolved, the static endpoints or in-mesh workloads that optionally back it, the namespaces it is exported to, and the subject alternate names the proxy checks when originating TLS.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (Istio Base CRDs) must be present and the Istio control plane (istiod) running. The entry is only honored where istiod is active.
- **Target namespace exists** -- the entry is created in a specific namespace; reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Istio Service Entry**, and click **Deploy**. The creation wizard walks you through the namespace, the service identity (hosts, location, ports), and how the service is resolved and backed, with guidance at each step. Start from the **Reach an External HTTPS API** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceEntry
metadata:
  name: external-https-api
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  hosts:
    - api.stripe.com
  location: MESH_EXTERNAL
  resolution: DNS
  ports:
    - number: 443
      name: https
      protocol: TLS
```

```shell
planton apply -f service-entry.yaml
```

This registers `api.stripe.com` as an external HTTPS service that workloads in `prod-apps` can reach by name, with TLS routed by SNI and the host resolved via DNS. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the entry to its Planton-managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  hosts:
    - api.stripe.com
  location: MESH_EXTERNAL
  resolution: DNS
  ports:
    - number: 443
      name: https
      protocol: TLS
```

The InfraPipeline deploys the namespace first, then registers the entry inside it.

## Key Configuration

These are the most important decisions when configuring a ServiceEntry. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Resolution and backing must agree -- the constraint web** -- static endpoints and a workload selector are mutually exclusive (at most one backs the service; both omitted is valid, e.g. an external host resolved by DNS). Static endpoints require Static, DNS, or DNS Round Robin resolution -- never None -- and DNS Round Robin allows at most one endpoint. CIDR `addresses` are honored only with None or Static resolution. Getting this web wrong is the main way ServiceEntries fail, and the spec rejects the invalid combinations at validation time instead of leaving them silently inert.

**Hosts are the match key everywhere** -- the hostname is matched against the HTTP Host/Authority header, the TLS SNI, and (for DNS resolution) the name to resolve. A bare `*` is not allowed; partial wildcards like `*.example.com` are -- the shape behind wildcard-egress registries.

**Location changes the trust story** -- MESH_EXTERNAL (the default posture for SaaS and partner APIs) treats the service as outside the mesh: no mTLS expectation, policy applies at egress. MESH_INTERNAL brings VMs and legacy services INTO the mesh's identity model -- reserve it for endpoints that actually run a proxy or share the mesh trust domain.

**Namespace is fixed at creation, and visibility is a choice** -- the entry registers relative to its namespace; `exportTo` controls which namespaces see it. The default (everywhere) makes an external host reachable mesh-wide -- fence it deliberately in multi-team clusters.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components and operators can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_entry_name` | Name of the created ServiceEntry (equals `metadata.name`) | Ordering resources that depend on the entry being in place |
| `namespace` | The namespace the entry was created in | Confirming where the entry applies |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

- **Reach an external HTTPS API** -- register a SaaS/partner endpoint so mesh workloads can call it by hostname, with TLS routed by SNI and the host resolved via DNS. The canonical starting point. Start from the **Reach an External HTTPS API** preset.
- **Bring static endpoints into the mesh** -- register a VM-hosted database or legacy service as an internal destination with explicit backing IPs and Static resolution. Start from the **Bring Static Endpoints Into the Mesh** preset.

## Works With

ServiceEntry is part of the Istio networking family. It requires the Istio Base CRDs and a running Istio control plane, and it pairs naturally with a Destination Rule -- the ServiceEntry makes the host *reachable and addressable*, while a DestinationRule configures *how* the mesh talks to it (load balancing, outlier detection, TLS origination). To order the entry after the workloads it fronts within an Infra Chart, express the dependency through `metadata.relationships`.

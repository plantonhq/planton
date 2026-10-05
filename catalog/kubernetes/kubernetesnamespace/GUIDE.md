# KubernetesNamespace Guide

The judgment this guide carries: when a namespace deserves to be its own
component in the architecture, versus riding along as a `createNamespace`
flag on whatever workload lands in it first.

## Choosing this component

Reach for a dedicated KubernetesNamespace the moment either is true:

- **More than one component will ever live in the namespace.** Workload
  kinds' `createNamespace` flag makes the FIRST deployed component the
  namespace's owner in IaC state: a second component with the same flag
  fails its deploy (the namespace already exists), and destroying the owner
  deletes the namespace out from under every other tenant. A dedicated
  component gives the namespace exactly one owner that no workload teardown
  touches. The full failure story and the wiring:
  [namespace-ownership pattern](../../_patterns/namespace-ownership.md).
- **The namespace itself needs configuration.** The flag creates a bare
  namespace (governance labels only). This kind's spec opens the
  namespace's real surface: resource quotas (T-shirt presets or custom),
  default-deny network policies with explicit allows, pod security
  standards, and service-mesh sidecar injection — see
  [reference.md](v1alpha1/reference.md) for every field.

Skip it only for a genuinely single-tenant namespace with no configuration
needs — that is the case `createNamespace: true` exists for.

## Platform conventions

- Workload kinds declare `spec.namespace` as a foreign key whose default
  target is this kind, reading `spec.name`. Wire it with `valueFrom` rather
  than repeating the namespace string — the reference makes the dependency
  explicit, orders the deploys, and draws the edge on the diagram:

```yaml
namespace:
  valueFrom:
    kind: KubernetesNamespace
    name: team-data
    fieldPath: spec.name
```

- `spec.name` is the actual Kubernetes namespace name (`metadata.name` is
  the Planton resource's name); the two often match, but only `spec.name`
  reaches the cluster.

## On the diagram

Architectures render as resource graphs users reason about. A dedicated
namespace component is a visible node with reference edges from each tenant
— the namespace boundary becomes part of the picture. A `createNamespace`
flag renders as nothing: the namespace exists in the cluster but is
invisible in the architecture.

## Pairs well with

Every Kubernetes workload kind that carries `spec.namespace` — which is
most of the provider's catalog. Compose one KubernetesNamespace per team or
per application environment, then point tenants at it.

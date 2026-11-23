---
kinds:
  - KubernetesNamespace
  - KubernetesPostgres
  - KubernetesValkey
  - KubernetesAgentSubstrate
  - KubernetesStigmer
---

# Namespace Ownership: Dedicated Component vs the createNamespace Flag

Most Kubernetes workload kinds carry a `createNamespace` convenience flag, and
composing several of them into one namespace with that flag set is the single
most common way agent-composed architectures fail on their second deploy. This
pattern is the judgment call: who owns the namespace.

## The problem

Kubernetes workload kinds (KubernetesPostgres, KubernetesValkey, and their
siblings) declare their target namespace the same way: a required
`spec.namespace` reference, plus a `createNamespace` boolean. When the flag is
true, the kind's IaC module creates the namespace as its OWN resource:
it enters that component's state, and — as the spec documents — it is
**deleted with the resource**.

Two consequences follow, both invisible at validation time and painful at
deploy time:

1. **The second creator fails.** Deploy `orders-db` and `orders-cache` into
   the same namespace, both with `createNamespace: true`: the first deploy
   creates and owns the namespace; the second fails, because its module
   issues a plain create for a namespace that now already exists.
2. **The owner's teardown takes the neighborhood down.** Destroying the
   component that owns the namespace deletes the namespace — and everything
   every other component deployed into it.

KubernetesManifest is the one kind whose anchor is optional: a raw manifest
can hold nothing but cluster-scoped objects, which no namespace distorts.
Without an anchor, a namespaced document that names no namespace of its own
is refused before anything applies, so the omission never places an object
in `default`; `createNamespace` needs an anchor to create.

## The composition

Give the namespace one owner: a dedicated KubernetesNamespace component.
Every workload references it and leaves `createNamespace` unset. The
`spec.namespace` field on workload kinds is a foreign key whose default
target is exactly this kind — the wiring below is what the schema itself
steers toward.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNamespace
metadata:
  name: team-data
spec:
  name: team-data
  resourceProfile:
    preset: medium
  podSecurityStandard: baseline
```

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPostgres
metadata:
  name: orders-db
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-data
      fieldPath: spec.name
  instances: 2
  storage:
    size: 10Gi
```

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesValkey
metadata:
  name: orders-cache
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-data
      fieldPath: spec.name
```

Deploy order follows the reference: the namespace component first, then the
workloads in any order. Adding a third workload later is a reference, not a
race.

## Choices and consequences

| Choice | What it buys | What it costs / risks |
|---|---|---|
| Dedicated KubernetesNamespace + references | One owner in one state; teardown of any workload never deletes the namespace; the namespace's own surface opens up (resource quotas, default-deny network policies, pod security standards, service-mesh injection); **renders as a visible node on the architecture diagram** with reference edges from each workload | One more kind to declare |
| `createNamespace: true` on a single workload | One less manifest for a genuinely single-tenant namespace | The workload owns the namespace: destroying it deletes the namespace; a second component with the flag fails on already-exists; namespace stays bare (no quotas, no network policies — nothing beyond the standard governance labels the module applies); **invisible on the diagram** — the namespace exists in the cluster but appears nowhere in the architecture |
| `createNamespace: true` on several components sharing a namespace | Nothing | The failure in "The problem" — first deploy wins, second fails |

The diagram consequence deserves weight when composing for a user:
architectures on this platform render as resource graphs, and a dedicated
namespace component is part of the picture a user sees and reasons about. A
flag is not.

## When not to use this

A genuinely single-tenant namespace — one component, nothing else ever
planned in it, no quota/network/security requirements on the namespace
itself — is exactly what `createNamespace: true` is for. One manifest,
correct ownership (the component IS the only tenant), one fewer moving part.
Use the dedicated component the moment a second tenant appears or the
namespace itself needs configuration.

The canonical instances of this case are the cluster's shared
Infra Components — an ingress controller in "ingress-nginx",
cert-manager in "cert-manager", ExternalDNS, a shared-chart operator in a
namespace of its own. Each conventionally owns a dedicated namespace no
other tenant will ever join, so the flag is the NORMAL shape there, not a
violation of this pattern.

One sole-tenant case still wants the dedicated component: a kind that reads Secrets someone must put in its namespace before it starts. KubernetesAgentSubstrate is the instance. Its namespace, `ate-system`, is fixed by upstream, and with S3 storage or an external database it reads the store's key and the connection strings from Secrets there. With the flag, those Secrets could only be created after the control plane they are meant for. So a KubernetesNamespace owns `ate-system`, the Secrets follow it, and the control plane comes last with `createNamespace: false`. It is also the shape that keeps the bundled database's volume across a destroy of the control plane, because the namespace it lives in is not the control plane's to delete.

One kind has no flag at all: KubernetesStigmer never creates or deletes the namespace its Stigmer lives in. Stigmer's operator keeps the encryption keys there (`stigmer-keys`), and they must outlive every destroy, because every backup is sealed with them. So a KubernetesNamespace owns that namespace, and deleting it, an environment teardown included, deletes the keys: run the kind's `keys_export_command` first. The operator's own namespace is the ordinary sole-tenant case, where `operator.create_namespace` is the normal shape.

## See also

- [KubernetesNamespace reference](../kubernetes/kubernetesnamespace/v1alpha1/reference.md)
  and [guide](../kubernetes/kubernetesnamespace/GUIDE.md)
- [KubernetesPostgres reference](../kubernetes/kubernetespostgres/v1alpha1/reference.md)
  and [guide](../kubernetes/kubernetespostgres/GUIDE.md)
- [KubernetesValkey reference](../kubernetes/kubernetesvalkey/v1alpha1/reference.md)
- [KubernetesAgentSubstrate reference](../kubernetes/kubernetesagentsubstrate/v1alpha1/reference.md)
  and [guide](../kubernetes/kubernetesagentsubstrate/GUIDE.md)

# Kubernetes Resource Quota

## Overview

**KubernetesResourceQuota** is a catalog kind that governs resource consumption in one namespace. It manages a governance PAIR: a `core/v1` **ResourceQuota** carrying aggregate caps on what the namespace may consume in total, and — when `spec.limit_defaults` is set — a companion `core/v1` **LimitRange** applying per-object defaults and bounds to individual pods, containers, and claims. They are two Kubernetes objects but one governance story — "how much may this namespace consume, and what does a workload get when it doesn't say?" — which is why this kind manages both.

The kind covers the complete `core/v1` ResourceQuotaSpec and LimitRangeSpec surfaces: compute, storage, and object-count caps; coarse scopes and fine-grained scope selectors; and per-container, per-pod, and per-claim defaults, bounds, and burst ratios.

## Purpose

Without quotas, a namespace can consume the entire cluster — one runaway team, one misconfigured HPA, one forgotten batch job. ResourceQuota is the standard multi-tenancy guardrail: the API server rejects any object whose creation would push the namespace's aggregate usage over a cap.

**Key value over raw manifests:**

- **The pair is managed together**: A compute quota without container defaults makes the API reject pods that omit requests/limits — the classic quota footgun. Setting `limit_defaults` alongside `hard` keeps the namespace livable, and both objects share one name and one lifecycle
- **Schema-level validation**: Conflicting scope pairs, the best_effort-caps-only-pods rule, the scope-selector operator contract, and defaults-on-container-only — all caught before anything reaches the cluster (these mirror the API server's own admission rejections)
- **Namespace by value or reference**: `spec.namespace` accepts a literal name or a reference to a `KubernetesNamespace` resource, so an infra chart can create the namespace and its governance in one run
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## The Quota + Defaults Interaction

The single most important thing to understand about ResourceQuota: **once a quota caps a compute resource (`requests.cpu`, `limits.memory`, ...), the API REJECTS pods that omit that request or limit.** A naive `kubectl run nginx` in a compute-governed namespace fails admission.

The fix is per-container defaults, which upstream models as a separate LimitRange object. This component treats the pair as one unit:

- **`spec.hard` alone** — the quota exists; every pod must explicitly declare the capped requests/limits or be rejected
- **`spec.hard` + `spec.limit_defaults`** — the safe pairing; workloads that omit requests/limits inherit the defaults instead of being rejected

The companion LimitRange shares the quota's name and namespace — one governance pair, one identity. It is created only when `limit_defaults` is set.

## The `hard` Vocabulary

Aggregate caps are resource name → quantity, using upstream's vocabulary:

- **Compute**: `requests.cpu`, `requests.memory`, `limits.cpu`, `limits.memory` (also plain `cpu`/`memory`, aliases for requests)
- **Storage**: `requests.storage` (total claimed), `persistentvolumeclaims` (count), and per-class variants (`<class>.storageclass.storage.k8s.io/requests.storage`)
- **Object counts**: `pods`, `services`, `services.loadbalancers`, `services.nodeports`, `secrets`, `configmaps`, `resourcequotas`, and the generic `count/<resource>.<group>` form for any countable object

Object-count caps are the safest to introduce — they constrain nothing that pods must declare, so they never break naive pod creation.

## Scopes and Scope Selectors

- **`spec.scopes`** — coarse filters on which objects the quota tracks; an object must match ALL listed scopes to be counted. `terminating`/`not_terminating` split by active deadline, `best_effort`/`not_best_effort` by QoS class, plus `priority_class`, `cross_namespace_pod_affinity`, and `volume_attributes_class`. When empty, the quota tracks everything its `hard` entries name
- **`spec.scope_selector`** — fine-grained filters for scopes that carry values, most usefully `priority_class` with `In`/`NotIn` (e.g. "quota only pods of priority class critical"). ANDs with `scopes`

Scoped quotas restrict which resources they may cap: a `best_effort` quota may cap only `pods` (BestEffort pods have no requests or limits to meter), and the pod-behavior scopes accept only the `Exists` operator in a scope selector. The schema enforces these rules — and rejects the conflicting pairs (`best_effort` + `not_best_effort`, `terminating` + `not_terminating`) — before deployment.

## `limit_defaults` — the Companion LimitRange

Each item governs one object type:

- **`container`** — the type that carries `default_request` and `default_limit` (containers are where requests/limits actually live), plus `min`/`max`/`max_limit_request_ratio`. When only `default_limit` is set, Kubernetes copies it into the request, producing Guaranteed-QoS containers
- **`pod`** — bounds across all of a pod's containers summed; pods cannot carry defaults, only `min`/`max`/`max_limit_request_ratio`
- **`persistent_volume_claim`** — per-claim storage bounds (`min`/`max` on `storage`)

## When to Use KubernetesNamespace Instead

For simple T-shirt-size governance on namespaces **Planton creates**, prefer the `KubernetesNamespace` kind's resource profiles — they manage a quota and limit range internally with sensible presets. This kind is the full-fidelity instrument: scope-filtered quotas, object-count caps, ratio bounds, and governance for namespaces Planton does not own.

## Quick Start

Create a file `resource-quota.yaml` — compute caps paired with container defaults, the safe pattern:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesResourceQuota
metadata:
  name: team-alpha-quota
spec:
  namespace:
    value: team-alpha
  name: team-alpha-quota
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
  limit_defaults:
    - type: container
      default_request:
        cpu: 100m
        memory: 128Mi
      default_limit:
        cpu: 500m
        memory: 512Mi
```

Deploy:

```shell
planton apply -f resource-quota.yaml
```

The namespace may now consume at most 10 CPUs / 20Gi of requests in aggregate, and any container that omits its requests/limits inherits the defaults instead of being rejected by the quota's admission check.

## Essential Configuration Fields

### Required

- **`spec.name`**: The ResourceQuota name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars); the companion LimitRange shares this name
- **`spec.hard`**: The aggregate caps (at least one entry)

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. When omitted, governs the cluster's `default` namespace
- **`spec.limit_defaults`**: Per-object defaults and bounds — the companion LimitRange. Omit entirely to manage only the quota
- **`spec.scopes`** / **`spec.scope_selector`**: Filters on which objects the quota tracks
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance; applied to both created objects

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`resource_quota_name`**: The name of the ResourceQuota object as created in the cluster
- **`namespace`**: The namespace the quota governs
- **`limit_range_name`**: The name of the companion LimitRange object; empty when the spec set no `limit_defaults` (no LimitRange is created)

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference; `default` when omitted)
2. Merge user labels and annotations with standard Planton tracking labels
3. Create the `core/v1` ResourceQuota with the `hard` caps, scopes (mapped to the API strings, e.g. `best_effort` → `BestEffort`), and scope selector
4. When `limit_defaults` is set, create the companion `core/v1` LimitRange under the same name and namespace, mapping each item's type and quantity maps one-to-one
5. Export the quota name, namespace, and LimitRange name (empty when none exists) for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesResourceQuota** when you need:

- Per-team or per-environment consumption caps in a shared cluster
- Compute governance with safe container defaults — the quota + LimitRange pairing in one resource
- Object-count caps (pods, services, LoadBalancers, PVCs) to contain sprawl and cloud cost
- Scope-filtered quotas — e.g. capping only BestEffort pods, or only pods of a given priority class
- Governance for namespaces Planton did not create

**Do NOT use** when:

- You want simple T-shirt-size governance on a Planton-created namespace — use `KubernetesNamespace`'s resource profiles instead
- You need cluster-wide (cross-namespace) aggregate limits — ResourceQuota is namespaced; hierarchical quota needs an addon
- You need runtime throttling — quotas act at admission time; they reject creations, they do not slow running pods

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (ResourceQuota and LimitRange admission plugins are enabled by default on all mainstream distributions)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the quota (unless deploying to `default`, or creating the namespace in the same chart via a reference)

## Best Practices

1. **Never cap compute without defaults**: A quota on `requests.*`/`limits.*` without `limit_defaults` (or defaults set elsewhere) makes the API reject every pod that omits them. Pair them — that is what this kind is built for
2. **Start with object counts**: `pods`, `services.loadbalancers`, and `persistentvolumeclaims` caps are safe to introduce on a live namespace; compute caps deserve a measured rollout
3. **Size quotas from observed usage, not guesses**: Check `kubectl describe resourcequota` after deployment — it shows used vs hard per resource — and adjust from real numbers
4. **Keep defaults modest and maxima honest**: `default_request` sets what every naive container is billed for against the quota; a generous default silently eats the namespace's budget
5. **Use scoped quotas for targeted policy**: A `best_effort`-scoped pod cap contains unbounded naive pods without touching well-behaved workloads; a `priority_class` scope selector can budget critical vs batch tiers separately

## References

- [Kubernetes Resource Quotas Documentation](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Kubernetes Limit Ranges Documentation](https://kubernetes.io/docs/concepts/policy/limit-range/)
- [ResourceQuota API Reference](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/resource-quota-v1/)
- [LimitRange API Reference](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/limit-range-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

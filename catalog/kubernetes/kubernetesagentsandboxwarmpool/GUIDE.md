# KubernetesAgentSandboxWarmPool Guide

The judgment this guide carries: a warm pool trades money for latency. Every member is a running sandbox nobody is using yet, and the size, the autoscaling and the rollout strategy decide how much latency you buy and when you stop paying for it.

## What warm buys

A cold sandbox costs a full pod start: scheduling, pulling the image, starting the runtime (gVisor or a Kata VM adds to it), and the container's own start-up. A warm member has done all of that already. A claim against the pool takes a ready member, re-labels it as its own, and is Ready in about the time it takes to update an object. The pool starts a replacement at the same moment, so the next claim is warm too, as long as claims don't outrun refills.

Two things always start cold, pool or no pool: a claim that sets environment variables or volume claim templates (a running pod cannot take either), and any claim that arrives when the pool has no ready member.

## Sizing

Size the pool for the burst of claims that arrive within one refill, the time a new member takes to become ready. A pool of three serves three claims in a burst warm; the fourth waits for the controller's warm-candidate window (a member without a pod IP yet) or starts cold. Every member reserves its template's resources the whole time it waits, so the pool's size times the template's reservation is the standing cost of instant sandboxes. A pool of zero is valid: every claim starts cold, and you pay nothing while idle.

## Autoscaling

The pool serves the scale subresource: its size at `.spec.replicas`, its current members at `.status.replicas`, and the selector of its members' pods at `.status.selector`. That is everything a HorizontalPodAutoscaler or KEDA needs to scale it like a Deployment:

- A **KubernetesHorizontalPodAutoscaler** targeting API version `extensions.agents.x-k8s.io/v1beta1`, kind `SandboxWarmPool` and this pool's `sandbox_warm_pool_name` output, scaling on the members' CPU or memory.
- **KEDA** for anything else: a schedule (zero outside working hours), the depth of a task queue, a custom metric.

Leave `replicas` unset when either owns the size. Planton applies the pool with server-side apply and never writes a field the spec leaves out, so the cluster's initial 1 and every later size the autoscaler sets stand. Setting `replicas` hands the field back to Planton, and every apply would reset the autoscaler.

## Rollouts

The controller compares each unclaimed member's blueprint (the template's pod, volume claim templates and `service`, and the template's name) with the current one. A member that no longer matches is stale:

- **`OnReplenish`** (default) leaves stale members in place. Each is replaced, from the current template, when it is claimed or deleted. No dip in ready members, but claims can get the previous shape until the pool turns over.
- **`Recreate`** deletes stale members at once and starts replacements. Every new claim gets the current template, at the cost of a window with fewer ready members.

Labels, annotations and the template's network and claim policies never make a member stale. Network policy changes reach every sandbox at once anyway, through the template's shared NetworkPolicy. Claimed sandboxes are never touched: they belong to their claims.

## Watching it

- The pool's status reports `readyReplicas`; ready below the declared size for long means claims are starting cold.
- The controller exports `agent_sandbox_warmpool_size` and claim start-up latency on its metrics endpoint (KubernetesAgentSandboxController's metrics Service).
- The controller's warm-pool settings (refill batch size and rate, replenish delay, readiness grace period, eviction) are on KubernetesAgentSandboxController, because they apply to every pool on the cluster.

## Parity accounting

Built for 100% parity with the upstream SandboxWarmPool (`extensions.agents.x-k8s.io/v1beta1`) at agent-sandbox v1.0.7. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.namespace`, `metadata.labels`, `metadata.annotations` | `namespace`, `labels`, `annotations` (the envelope) |
| `spec.replicas` | `replicas` |
| `spec.sandboxTemplateRef.name` | `sandboxTemplateRef.name`, a reference to a KubernetesAgentSandboxTemplate |
| `spec.updateStrategy.type` | `updateStrategy.type` |

### What the spec adds beyond upstream's schema

- **The template as a reference**, so the resource graph draws template, then pool, and deploys the template first.
- **The update strategy as a closed set**, refused before the apply instead of by the API server.

### Not modeled

- **The status block** (`replicas`, `readyReplicas`, `selector`, `observedGeneration`) — the controller's report. The pool's name and namespace are the outputs a consumer needs.

## Pairs well with

- **[KubernetesAgentSandboxTemplate](../kubernetesagentsandboxtemplate/README.md)** — the blueprint every member is built from.
- **[KubernetesAgentSandboxClaim](../kubernetesagentsandboxclaim/README.md)** — a declared claim against the pool.
- **[KubernetesHorizontalPodAutoscaler](../kuberneteshorizontalpodautoscaler/README.md)** — scales the pool through its scale endpoint.
- **[KubernetesAgentSandboxController](../kubernetesagentsandboxcontroller/README.md)** — fills the pool, binds claims, and carries the cluster-wide warm-pool settings.

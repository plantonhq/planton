# KubernetesAgentSandboxClaim Guide

The judgment this guide carries: a claim is the moment a sandbox becomes someone's. Whether it arrives warm, carries what it was asked for, and ends when it should is decided by what the claim asks for and what its pool's template allows. Ask for as little as the purpose needs.

## Declare it, or let the product create it

A product hands out sandboxes by creating claims at runtime, usually through agent-sandbox's SDKs: `create_sandbox(warmpool=..., namespace=...)` creates a claim, waits for it to bind, and reaches the sandbox through the router. Declare a claim here only when its sandbox should exist as long as the declaration does: a standing workbench for a named person, a team's shared agent, a demo, or a canary that proves the pool hands out working sandboxes. A declared claim is reconciled like any other resource. Remove it from the declaration and its sandbox goes.

## Warm or cold

| The claim sets | What it gets |
|----------------|--------------|
| nothing beyond the pool (and any pod labels or annotations) | a ready pool member, re-labelled as the claim's, in about the time it takes to update an object |
| environment variables or volume claim templates | a sandbox started for it from the pool's template, named after the claim (a full pod start) |
| anything, when the pool has no ready member | a cold start, after the controller's short wait for a member that is still getting its pod IP |

Pod labels and annotations never cost a warm start: a running member takes them in place. Variables and volumes always do, because a running pod cannot take either.

## What the template must allow

The pool's template decides what a claim may change:

- **Environment variables** need `envVarsInjectionPolicy` `Allowed` (names the template doesn't define) or `Overrides` (also its values). Under `Disallowed` the claim reports `Ready=False` with reason `EnvVarsInjectionRejected`.
- **Volume claim templates** need `volumeClaimTemplatesPolicy` `Allowed` (new ones) or `Overrides` (also replacing the template's by name, such as a larger workspace). Otherwise the reason is `VolumeClaimTemplatesError`.
- **Pod labels** must carry a domain from the controller's allowed label domains (`sandbox.users.io` and its subdomains unless the controller's `extensions.allowed_label_domains` says otherwise), with values of at most 63 characters. Annotations in restricted system domains are refused, except `cluster-autoscaler.kubernetes.io/safe-to-evict`. Either mistake is reason `InvalidMetadata`.

The refusals happen in the controller, after the apply succeeds, so watch the claim's `Ready` condition, not only the deploy.

Claim variables are literals that land in the custom resource as written. Keep anything secret in the template, as a variable read from a Secret with `secretKeyRef`.

## How a claim ends

- **`lifecycle.shutdownTime`** ends the sandbox at an absolute time. It is the claim's own and is not copied to the sandbox; the controller enforces it.
- **`lifecycle.ttlSecondsAfterFinished`** ends it a set time after its pod finishes (succeeds or fails), for sandboxes that run one task to completion.
- **`lifecycle.shutdownPolicy`** decides the claim itself: `Retain` (default) keeps it, marked expired; `Delete` deletes it; `DeleteForeground` deletes it but keeps it visible until its sandbox and pod are gone, so a watcher can follow the shutdown.

`Delete` and `DeleteForeground` remove an object Planton still declares, and the next apply claims again. They suit runtime claims; keep `Retain` for a declared one.

## Finding the sandbox

The controller picks the sandbox's name when it binds: a warm member's `<pool>-<random>` name, or the claim's own name for a cold start. It records it in the claim's status (`sandbox.name`, with the pod IPs and the Service's DNS name when there is one). A caller reads that name and sends it to the controller's router as `X-Sandbox-ID`, with the claim's namespace as `X-Sandbox-Namespace`.

## Parity accounting

Built for 100% parity with the upstream SandboxClaim (`extensions.agents.x-k8s.io/v1beta1`) at agent-sandbox v1.0.7. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.namespace`, `metadata.labels`, `metadata.annotations` | `namespace`, `labels`, `annotations` (the envelope) |
| `spec.warmPoolRef.name` | `warmPoolRef.name`, a reference to a KubernetesAgentSandboxWarmPool |
| `spec.lifecycle.shutdownTime`, `.ttlSecondsAfterFinished`, `.shutdownPolicy` | `lifecycle.shutdownTime` (RFC 3339), `.ttlSecondsAfterFinished`, `.shutdownPolicy` |
| `spec.additionalPodMetadata.labels`, `.annotations` | `additionalPodMetadata.labels`, `.annotations` |
| `spec.env[]` (`name`, `value`, `containerName`) | `env[]` |
| `spec.volumeClaimTemplates[]` | `volumeClaimTemplates[]`, with the full PersistentVolumeClaimSpec |

### What the spec adds beyond upstream's schema

- **The pool as a reference**, so the resource graph draws template, pool, claim, and deploys them in that order.
- **Closed sets and forms**: the shutdown policy, an RFC 3339 expiry, a non-negative time to live, label grammar on the pod labels.
- **Environment variables** named as Kubernetes accepts (printable characters except `=`), with a value, unique per container.
- **A name on every volume claim template**, as a DNS label, unique within the claim.

### Not modeled

- **The status block** (`conditions`, `sandbox.name`, `sandbox.podIPs`, `sandbox.serviceFQDN`) — the controller's report.
- **An empty environment variable value**: upstream requires a value key, and an empty one cannot be told apart from none in the spec's encoding; leave the variable out instead.

## Pairs well with

- **[KubernetesAgentSandboxWarmPool](../kubernetesagentsandboxwarmpool/README.md)** — the pool the claim draws from.
- **[KubernetesAgentSandboxTemplate](../kubernetesagentsandboxtemplate/README.md)** — what the sandbox runs, and what the claim may change.
- **[KubernetesAgentSandboxController](../kubernetesagentsandboxcontroller/README.md)** — binds the claim, carries the allowed label domains, and routes callers to the sandbox.

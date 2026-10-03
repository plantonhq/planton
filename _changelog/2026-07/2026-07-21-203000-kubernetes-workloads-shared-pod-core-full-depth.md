# Kubernetes workloads rebuilt on a shared pod core at full configuration depth

**Date**: 2026-07-21
**Scope**: `apis/dev/planton/provider/kubernetes` (five workload kinds + shared workload protos), `pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod`, `pkg/outputs`, `aa_import`, `e2e`, site catalog, `_rules/catalog-kind/update`

## What changed

### One shared pod core for every workload kind

KubernetesDeployment, KubernetesStatefulSet, KubernetesDaemonSet,
KubernetesJob, and KubernetesCronJob previously each modeled their own
container/pod surface (three divergent per-kind container messages; the batch
kinds flattened a single container into the spec root). All five now compose
two shared provider-root protos:

- `workload_container.proto` — the complete container: image + pull policy,
  command/args/working dir, ports, env, resources, all three probes, volume
  mounts, lifecycle hooks (including the kubelet-native sleep action),
  container security context with capabilities and seccomp.
- `workload_pod.proto` — the pod surface: ServiceAccount reference,
  automount control, image pull secrets, init containers, pod
  labels/annotations, full scheduling (node selector, tolerations, node/pod
  affinity, topology spread), pod security context (fsGroup, sysctls,
  seccomp), termination grace, DNS policy/config, host aliases, host
  namespaces, priority and runtime class.

The same shape is used for app containers, sidecars, and init containers —
anything expressible on one is expressible on all. The legacy limited
sidecar `Container` message is gone. On the Pulumi side, one shared builder
package (`pkg/.../workloadpod`) renders containers, volumes (derived as the
de-duplicated union of every container's mounts), scheduling, and security
for all five kinds, so workload semantics cannot drift between kinds.

### Per-kind depth from the upstream core API types

- **Deployment**: Recreate + RollingUpdate strategies, minReadySeconds,
  revisionHistoryLimit, progressDeadlineSeconds, paused; HPA gains a proper
  `max_replicas` ceiling and is now implemented in BOTH engines (previously
  spec-only on Pulumi).
- **StatefulSet**: update strategy with partition canaries, PVC retention
  policy (whenDeleted/whenScaled), ordinals, pod management policy, deepened
  volume claim templates (access modes, volume mode); the governing Service
  is headless with publishNotReadyAddresses so bootstrapping members can
  discover each other; a `pod_dns_template` output teaches per-replica DNS.
- **DaemonSet**: update strategy with maxSurge, full node-targeting through
  the shared scheduling block; the bespoke inline ServiceAccount/RBAC bundle
  is removed — identity composes from KubernetesServiceAccount and
  KubernetesRbac like every other kind.
- **Job**: complete batch/v1 surface — Indexed completion, per-index retry
  budgets (backoffLimitPerIndex, maxFailedIndexes), podFailurePolicy
  (ordered rules over exit codes and pod conditions), successPolicy,
  activeDeadline, TTL, suspend. Deploys with replace-on-spec-change
  semantics (Job specs are immutable server-side).
- **CronJob**: schedule validation, IANA time zone, startingDeadlineSeconds,
  history limits, and a full nested job template carrying the same batch
  controls. Concurrency defaults to `Forbid` — deliberately safer than the
  upstream `Allow` default, since overlapping cron runs are the classic
  scheduled-workload incident. CronJob is also registered as a service kind
  (pipelines inject the built image at `spec.job_template.container.app.image`).

### Exposure is composed, never embedded

Workload kinds no longer create ingress resources. The embedded
`ingress {enabled, hostname}` blocks — which silently provisioned
certificates, Istio gateways, and routes inside workload modules — are gone,
along with the modules' hardcoded coupling to an istio-ingress namespace.
Workloads export `service`, `kube_endpoint`, and `selector_labels`, and
first-class exposure kinds (Gateway API routes, certificates) attach to
those outputs as visible resource-graph nodes. The inline `config_maps` map
passthrough is likewise gone — configuration composes from the first-class
KubernetesConfigMap kind.

### Cross-engine parity, loudly enforced

Both engines implement the full surface. Where the Terraform provider
genuinely cannot express a field (Job suspend/successPolicy, StatefulSet
rolling-update maxUnavailable/ordinals, the sleep lifecycle hook), the gap
carries twin `PARITY-EXCEPTION:` comments in BOTH engines AND a Terraform
plan-time precondition that rejects the field with an error naming the
working engine — a set field can never be silently dropped.

### Validation

- 150 protovalidate spec-test cases across the five kinds; outputs
  conformance cases for all five; secret-coverage and FK-reference gates
  green; offline tofu plan proofs green per kind including the
  optional-blocks-absent shape.
- Live E2E on both engines (kind cluster): Deployment 13 scenarios,
  StatefulSet 13, DaemonSet 5, Job 5, CronJob 4 — full
  create→verify-outputs→verify-resources→destroy→verify-clean on every
  scenario, zero orphans. Scenario matrices now cover init containers,
  sidecars, hardened security contexts, PVC retention, partition updates,
  Indexed jobs, and pod failure policies.
- The live lanes caught (and this change fixes) a real cross-engine bug:
  the Pulumi env-secret collector read the secret-value oneof through an
  interface assertion that generated wrapper structs never satisfy, so the
  env Secret never materialized and pods wedged at startup. The Terraform
  engine was unaffected — exactly the divergence class the dual-engine E2E
  promise exists to catch.

### Import recipes and docs

All five kinds ship import maps (the provider catalog gains the workload
resource types and satellites); READMEs, research docs, catalog pages, and
presets are rewritten from the new specs — including hardened-production
presets that teach the restricted Pod Security Standard profile. Site
catalog pages regenerated.

### Workflow

The kind-update rule gains two hard-won lessons: never read a proto
oneof through an interface assertion in module code, and fail the Terraform
plan (precondition) when a provider cannot express a spec field.

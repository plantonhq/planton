# KubernetesTekton

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesTektonSpec** declares the cluster's Tekton installation —
which components run (Pipelines, Triggers, Dashboard, Chains), where
they install, the pipeline feature flags and defaults, and the
pruner policy — as a `TektonConfig` custom resource that the Tekton
Operator reconciles.

PREREQUISITE: the Tekton Operator must be installed first (declare a
KubernetesTektonOperator resource). This kind renders the
TektonConfig; the operator turns it into running components via
TektonInstallerSet resources and keeps them converged.

ONE PER CLUSTER: the operator's admission webhook allows exactly one
TektonConfig, and it must be named `config` — the modules render
that fixed name regardless of metadata.name. Declare exactly one
KubernetesTekton per cluster.

TARGET NAMESPACE IS IMMUTABLE: the operator's webhook rejects
changing `target_namespace` on an existing installation — to move
Tekton to a different namespace, destroy this resource and create it
again with the new value.

DESTROY BEHAVIOR: deleting the TektonConfig makes the operator tear
down every component it installed (the TektonInstallerSet resources
carry finalizers only the operator can process). The modules wait
for that teardown to complete on destroy. Never destroy the
KubernetesTektonOperator resource before this one — without a
running operator the finalizers strand and deletion hangs.

IMAGES: every component image comes from where the operator is told
to pull it — set KubernetesTektonOperator's `image_registry` to pull
them all from a mirror; this declaration names no images.

EXPOSURE: when the dashboard is installed (profile `all`) its
Service stays ClusterIP; expose it via first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported service
handle. Note the dashboard has NO built-in authentication — never
expose it without an authenticating layer in front.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed block rendered at once on the `all` profile.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTekton
metadata:
  name: tekton-full
spec:
  profile: all
  targetNamespace: ci-tekton
  targetNamespaceMetadata:
    labels:
      team: platform
    annotations:
      cost-center: ci
  placement:
    nodeSelector:
      role: ci
    tolerations:
      - key: ci
        operator: Exists
        effect: NoSchedule
    priorityClassName: platform-critical
  pipeline:
    cloudEventsSinkUrl: https://ci-events.example.com/tekton
    enableApiFields: beta
    defaultTimeoutMinutes: 90
    defaultServiceAccount: pipelines-runner
    features:
      disableCredsInit: true
      awaitSidecarReadiness: true
      runningInEnvironmentWithInjectedSidecars: false
      requireGitSshSecretKnownHosts: true
      enableCustomTasks: true
      keepPodOnCancel: false
      enableProvenanceInStatus: true
      setSecurityContext: true
      enableCelInWhenexpression: true
      enableStepActions: true
      enableParamEnum: true
      resultsFrom: termination-message
      maxResultSize: 8192
      coschedule: workspaces
    resolvers:
      enableBundlesResolver: true
      enableHubResolver: false
      enableGitResolver: true
      enableClusterResolver: true
    metrics:
      taskrunLevel: task
      taskrunDurationType: histogram
      pipelinerunLevel: pipeline
      pipelinerunDurationType: histogram
      countWithReason: true
    performance:
      replicas: 2
      buckets: 2
      threadsPerController: 4
      kubeApiQps: 50
      kubeApiBurst: 100
  trigger:
    enableApiFields: stable
    defaultServiceAccount: triggers-runner
  dashboard:
    readonly: true
    externalLogs: https://logs.example.com
  chain:
    disabled: false
    generateSigningSecret: true
  pruner:
    schedule: "0 8 * * *"
    resources:
      - pipelinerun
      - taskrun
    keep: 100
    prunePerResource: true
  additionalParams:
    - name: createRbacResource
      value: "true"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.profile` | `string` |  | `all` |  |
| `spec.targetNamespace` | `string` |  | `tekton-pipelines` |  |
| `spec.targetNamespaceMetadata` | `KubernetesTektonNamespaceMetadata` |  |  |  |
| `spec.targetNamespaceMetadata.labels` | `map<string, string>` |  |  |  |
| `spec.targetNamespaceMetadata.annotations` | `map<string, string>` |  |  |  |
| `spec.placement` | `KubernetesTektonPlacement` |  |  |  |
| `spec.placement.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.placement.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.placement.tolerations[].key` | `string` |  |  |  |
| `spec.placement.tolerations[].operator` | `string` |  |  |  |
| `spec.placement.tolerations[].value` | `string` |  |  |  |
| `spec.placement.tolerations[].effect` | `string` |  |  |  |
| `spec.placement.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.placement.priorityClassName` | `string` |  |  |  |
| `spec.pipeline` | `KubernetesTektonPipeline` |  |  |  |
| `spec.pipeline.cloudEventsSinkUrl` | `string` |  |  |  |
| `spec.pipeline.enableApiFields` | `string` |  |  |  |
| `spec.pipeline.defaultTimeoutMinutes` | `int32` |  |  |  |
| `spec.pipeline.defaultServiceAccount` | `string` |  |  |  |
| `spec.pipeline.features` | `KubernetesTektonPipelineFeatures` |  |  |  |
| `spec.pipeline.features.disableCredsInit` | `bool` |  |  |  |
| `spec.pipeline.features.awaitSidecarReadiness` | `bool` |  |  |  |
| `spec.pipeline.features.runningInEnvironmentWithInjectedSidecars` | `bool` |  |  |  |
| `spec.pipeline.features.requireGitSshSecretKnownHosts` | `bool` |  |  |  |
| `spec.pipeline.features.enableCustomTasks` | `bool` |  |  |  |
| `spec.pipeline.features.keepPodOnCancel` | `bool` |  |  |  |
| `spec.pipeline.features.enableProvenanceInStatus` | `bool` |  |  |  |
| `spec.pipeline.features.setSecurityContext` | `bool` |  |  |  |
| `spec.pipeline.features.enableCelInWhenexpression` | `bool` |  |  |  |
| `spec.pipeline.features.enableStepActions` | `bool` |  |  |  |
| `spec.pipeline.features.enableParamEnum` | `bool` |  |  |  |
| `spec.pipeline.features.resultsFrom` | `string` |  |  |  |
| `spec.pipeline.features.maxResultSize` | `int32` |  |  |  |
| `spec.pipeline.features.coschedule` | `string` |  |  |  |
| `spec.pipeline.resolvers` | `KubernetesTektonPipelineResolvers` |  |  |  |
| `spec.pipeline.resolvers.enableBundlesResolver` | `bool` |  |  |  |
| `spec.pipeline.resolvers.enableHubResolver` | `bool` |  |  |  |
| `spec.pipeline.resolvers.enableGitResolver` | `bool` |  |  |  |
| `spec.pipeline.resolvers.enableClusterResolver` | `bool` |  |  |  |
| `spec.pipeline.metrics` | `KubernetesTektonPipelineMetrics` |  |  |  |
| `spec.pipeline.metrics.taskrunLevel` | `string` |  |  |  |
| `spec.pipeline.metrics.taskrunDurationType` | `string` |  |  |  |
| `spec.pipeline.metrics.pipelinerunLevel` | `string` |  |  |  |
| `spec.pipeline.metrics.pipelinerunDurationType` | `string` |  |  |  |
| `spec.pipeline.metrics.countWithReason` | `bool` |  |  |  |
| `spec.pipeline.performance` | `KubernetesTektonPipelinePerformance` |  |  |  |
| `spec.pipeline.performance.replicas` | `int32` |  |  |  |
| `spec.pipeline.performance.buckets` | `int32` |  |  |  |
| `spec.pipeline.performance.threadsPerController` | `int32` |  |  |  |
| `spec.pipeline.performance.kubeApiQps` | `int32` |  |  |  |
| `spec.pipeline.performance.kubeApiBurst` | `int32` |  |  |  |
| `spec.trigger` | `KubernetesTektonTrigger` |  |  |  |
| `spec.trigger.enableApiFields` | `string` |  |  |  |
| `spec.trigger.defaultServiceAccount` | `string` |  |  |  |
| `spec.dashboard` | `KubernetesTektonDashboard` |  |  |  |
| `spec.dashboard.readonly` | `bool` |  |  |  |
| `spec.dashboard.externalLogs` | `string` |  |  |  |
| `spec.chain` | `KubernetesTektonChain` |  |  |  |
| `spec.chain.disabled` | `bool` |  |  |  |
| `spec.chain.generateSigningSecret` | `bool` |  |  |  |
| `spec.pruner` | `KubernetesTektonPruner` |  |  |  |
| `spec.pruner.schedule` | `string` | yes |  |  |
| `spec.pruner.resources` | `[]string` | yes |  |  |
| `spec.pruner.keep` | `int32` |  |  |  |
| `spec.pruner.keepSince` | `int32` |  |  |  |
| `spec.pruner.prunePerResource` | `bool` |  |  |  |
| `spec.additionalParams` | `[]KubernetesTektonParam` |  |  |  |
| `spec.additionalParams[].name` | `string` | yes |  |  |
| `spec.additionalParams[].value` | `string` | yes |  |  |

## Field Details

### spec.profile

`string` · optional (explicit presence)

Which components install. `lite` = Pipelines only. `basic` =
Pipelines + Triggers. `all` = Pipelines + Triggers + Dashboard.
Empty = `all` (the operator's own default). Chains additionally
installs on `basic` and `all` unless `chain.disabled` is set.

- default: `all`
- rule: {"string":{"in":["","lite","basic","all"]}}

### spec.targetNamespace

`string` · optional (explicit presence)

Namespace the Tekton components install into. The operator
creates and owns it — INCLUDING deletion: tearing this resource
down removes the target namespace with the components (verified
live), so never point it at a namespace carrying anything else.
Empty = `tekton-pipelines` (the upstream default). IMMUTABLE:
changing it on an existing installation is rejected by the
operator's webhook — destroy and recreate instead.

- default: `tekton-pipelines`
- rule: {"string":{"pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.targetNamespaceMetadata

`KubernetesTektonNamespaceMetadata`

Extra labels and annotations for the operator-created target
namespace (e.g. pod-security or cost-attribution labels).

### spec.targetNamespaceMetadata.labels

`map<string, string>`

Labels merged onto the target namespace.

### spec.targetNamespaceMetadata.annotations

`map<string, string>`

Annotations merged onto the target namespace.

### spec.placement

`KubernetesTektonPlacement`

Scheduling applied to EVERY Tekton component pod the operator
deploys (controllers, webhooks, dashboard).

### spec.placement.nodeSelector

`map<string, string>`

Node selector for all Tekton component pods.

### spec.placement.tolerations

`[]WorkloadToleration`

Tolerations for all Tekton component pods.

### spec.placement.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.placement.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.placement.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.placement.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.placement.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.placement.priorityClassName

`string`

Priority class name for all Tekton component pods.

### spec.pipeline

`KubernetesTektonPipeline`

Tekton Pipelines configuration — feature flags, execution
defaults, CloudEvents, resolvers, metrics and controller
performance. Installed on every profile.

### spec.pipeline.cloudEventsSinkUrl

`string`

URL that receives CloudEvents for every TaskRun/PipelineRun
lifecycle change (e.g.
"http://receiver.ci.svc.cluster.local/events"). KNOW THIS: Tekton
supports exactly ONE cluster-global sink — every pipeline in
every namespace reports here. Multi-tenant clusters that need
per-team routing put a fan-out service at this URL and route
downstream (the event carries its source namespace).

- rule: cloud_events_sink_url must be an http:// or https:// URL

### spec.pipeline.enableApiFields

`string` · optional (explicit presence)

API stability gate for Tekton CRD fields: `stable` (the
default), `beta`, or `alpha` (enables in-development fields —
never on production clusters).

- rule: {"string":{"in":["","stable","beta","alpha"]}}

### spec.pipeline.defaultTimeoutMinutes

`int32` · optional (explicit presence)

Default timeout in minutes applied to PipelineRuns that declare
none. Empty = Tekton's default (60).

- rule: {"int32":{"gte":1}}

### spec.pipeline.defaultServiceAccount

`string`

ServiceAccount name TaskRun/PipelineRun pods use when their run
declares none. Empty = "default". Point it at a dedicated,
cloud-annotated account when most pipelines need registry or
cloud access.

### spec.pipeline.features

`KubernetesTektonPipelineFeatures`

Feature flags for the pipelines controller. Each maps to one
`feature-flags` ConfigMap key; empty fields keep Tekton's
defaults.

### spec.pipeline.features.disableCredsInit

`bool` · optional (explicit presence)

Stop injecting Git/registry credentials from the run
ServiceAccount's annotated Secrets into step containers
(`disable-creds-init`). Tekton default: false (creds-init runs).

### spec.pipeline.features.awaitSidecarReadiness

`bool` · optional (explicit presence)

Wait for sidecar containers to be ready before starting steps
(`await-sidecar-readiness`). Tekton default: true.

### spec.pipeline.features.runningInEnvironmentWithInjectedSidecars

`bool` · optional (explicit presence)

Tell Tekton the cluster injects sidecars (Istio-class meshes) so
it stops counting them as step containers
(`running-in-environment-with-injected-sidecars`). Tekton
default: true. Set false on mesh-free clusters for a minor
scheduling optimization.

### spec.pipeline.features.requireGitSshSecretKnownHosts

`bool` · optional (explicit presence)

Require a `known_hosts` entry in Git SSH Secrets
(`require-git-ssh-secret-known-hosts`). Tekton default: false.
Turn on to refuse man-in-the-middle-able Git fetches.

### spec.pipeline.features.enableCustomTasks

`bool` · optional (explicit presence)

Allow Custom Task references (`enable-custom-tasks`). Tekton
default: true.

### spec.pipeline.features.keepPodOnCancel

`bool` · optional (explicit presence)

Keep the pod around when a run is cancelled, for debugging
(`keep-pod-on-cancel`). Tekton default: false. Alpha-gated:
requires enable_api_fields "alpha".

### spec.pipeline.features.enableProvenanceInStatus

`bool` · optional (explicit presence)

Record provenance in run statuses (`enable-provenance-in-status`).
Tekton default: true — Chains reads it.

### spec.pipeline.features.setSecurityContext

`bool` · optional (explicit presence)

Apply Tekton's own restricted security context to step containers
(`set-security-context`) so runs schedule in namespaces at the
`restricted` Pod Security level. Tekton default: false.

### spec.pipeline.features.enableCelInWhenexpression

`bool` · optional (explicit presence)

Allow CEL expressions in `when` clauses
(`enable-cel-in-whenexpression`). Tekton default: false.

### spec.pipeline.features.enableStepActions

`bool` · optional (explicit presence)

Allow StepAction CRs (`enable-step-actions`). Tekton default:
true (beta).

### spec.pipeline.features.enableParamEnum

`bool` · optional (explicit presence)

Allow enum constraints on params (`enable-param-enum`). Tekton
default: false.

### spec.pipeline.features.resultsFrom

`string` · optional (explicit presence)

Where step results are read from: `termination-message` (the
default; capped at 4 KB per step) or `sidecar-logs` (larger
results via a log-reading sidecar; alpha).

- rule: {"string":{"in":["","termination-message","sidecar-logs"]}}

### spec.pipeline.features.maxResultSize

`int32` · optional (explicit presence)

Maximum result size in bytes when `results_from` is
"sidecar-logs". Tekton default: 4096.

- rule: {"int32":{"gte":1}}

### spec.pipeline.features.coschedule

`string` · optional (explicit presence)

How PipelineRun pods with shared workspaces co-schedule:
`workspaces` (the default), `pipelineruns`, `isolate-pipelinerun`
or `disabled`.

- rule: {"string":{"in":["","workspaces","pipelineruns","isolate-pipelinerun","disabled"]}}

### spec.pipeline.resolvers

`KubernetesTektonPipelineResolvers`

The built-in remote resolvers (each answers a `taskRef`/
`pipelineRef` resolver type). All four are enabled by Tekton's
default; disable the ones a locked-down cluster must not have
(e.g. `git` and `hub` reach the internet from the resolvers
deployment).

### spec.pipeline.resolvers.enableBundlesResolver

`bool` · optional (explicit presence)

Resolve refs from Tekton bundles (OCI images). Tekton default:
true.

### spec.pipeline.resolvers.enableHubResolver

`bool` · optional (explicit presence)

Resolve refs from the Tekton Hub / Artifact Hub. Tekton default:
true. Reaches the public internet.

### spec.pipeline.resolvers.enableGitResolver

`bool` · optional (explicit presence)

Resolve refs from Git repositories. Tekton default: true.

### spec.pipeline.resolvers.enableClusterResolver

`bool` · optional (explicit presence)

Resolve refs from Tasks/Pipelines already in the cluster. Tekton
default: true.

### spec.pipeline.metrics

`KubernetesTektonPipelineMetrics`

Prometheus metrics shape for the pipelines controller.

### spec.pipeline.metrics.taskrunLevel

`string` · optional (explicit presence)

TaskRun metric granularity: `task`, `taskrun`, or `namespace`
(Tekton default). High-cardinality levels (`taskrun`) can
overwhelm Prometheus on busy clusters.

- rule: {"string":{"in":["","task","taskrun","namespace"]}}

### spec.pipeline.metrics.taskrunDurationType

`string` · optional (explicit presence)

TaskRun duration metric type: `histogram` (Tekton default) or
`lastvalue` / `namespace` gauges.

- rule: {"string":{"in":["","histogram","lastvalue","namespace"]}}

### spec.pipeline.metrics.pipelinerunLevel

`string` · optional (explicit presence)

PipelineRun metric granularity: `pipeline`, `pipelinerun`, or
`namespace` (Tekton default).

- rule: {"string":{"in":["","pipeline","pipelinerun","namespace"]}}

### spec.pipeline.metrics.pipelinerunDurationType

`string` · optional (explicit presence)

PipelineRun duration metric type: `histogram` (Tekton default) or
`lastvalue` / `namespace` gauges.

- rule: {"string":{"in":["","histogram","lastvalue","namespace"]}}

### spec.pipeline.metrics.countWithReason

`bool` · optional (explicit presence)

Add the failure reason as a metric label
(`metrics.count.enable-reason`). Tekton default: false.

### spec.pipeline.performance

`KubernetesTektonPipelinePerformance`

Controller performance tuning for large fleets.

### spec.pipeline.performance.replicas

`int32` · optional (explicit presence)

Pipelines controller replicas. Extra replicas shard reconcile
work via `buckets` (this is Tekton's HA story — replicas without
matching buckets add nothing). Empty = 1.

- rule: {"int32":{"gte":1}}

### spec.pipeline.performance.buckets

`int32` · optional (explicit presence)

Number of reconcile-work shards across replicas (Tekton HA;
upstream maximum 10). Set together with `replicas`.

- rule: {"int32":{"lte":10,"gte":1}}

### spec.pipeline.performance.threadsPerController

`int32` · optional (explicit presence)

Worker threads per controller. Empty = Tekton's default (2).

- rule: {"int32":{"gte":1}}

### spec.pipeline.performance.kubeApiQps

`int32` · optional (explicit presence)

Kubernetes API client QPS for the controller. Empty = Tekton's
default (5.0). Busy clusters raise it together with
`kube_api_burst`.

- rule: {"int32":{"gte":1}}

### spec.pipeline.performance.kubeApiBurst

`int32` · optional (explicit presence)

Kubernetes API client burst for the controller. Empty = Tekton's
default (10).

- rule: {"int32":{"gte":1}}

### spec.trigger

`KubernetesTektonTrigger`

Tekton Triggers configuration. Installed on profiles `basic` and
`all`; ignored on `lite`.

### spec.trigger.enableApiFields

`string` · optional (explicit presence)

API stability gate for Triggers CRD fields: `stable` (the
default), `beta`, or `alpha`.

- rule: {"string":{"in":["","stable","beta","alpha"]}}

### spec.trigger.defaultServiceAccount

`string`

ServiceAccount name EventListener pods use when their listener
declares none. Empty = "default".

### spec.dashboard

`KubernetesTektonDashboard`

Tekton Dashboard configuration. Installed on profile `all` only;
ignored otherwise.

### spec.dashboard.readonly

`bool`

Run the dashboard read-only: viewing runs and logs works, but
creating/re-running/deleting through the UI is disabled. Tekton
default: false (full write access). KNOW THIS before exposing
the dashboard anywhere — it has no authentication of its own, so
a writable dashboard hands pipeline-execution power to anyone who
can reach it.

### spec.dashboard.externalLogs

`string`

External logs provider URL the dashboard links to for logs of
garbage-collected pods (e.g. a log-archive frontend). Empty =
pod logs only.

### spec.chain

`KubernetesTektonChain`

Tekton Chains — the supply-chain security component that observes
completed runs and signs their provenance. Installs on profiles
`basic` and `all` unless disabled here. Signing keys and
attestation storage are configured post-install on the
chains-config ConfigMap (a security workflow, not an install
concern).

### spec.chain.disabled

`bool`

Do not install Chains. Chains observes completed runs and signs
provenance attestations — teams not consuming supply-chain
attestations can disable it to save the controller's footprint.

### spec.chain.generateSigningSecret

`bool`

Generate the x509 signing key automatically (the
`signing-secrets` Secret in the target namespace) so Chains signs
out of the box. Leave false to provision the signing key
yourself (the cosign workflow).

### spec.pruner

`KubernetesTektonPruner`

Automatic cleanup of completed PipelineRuns/TaskRuns. Empty = no
pruner cron — completed runs accumulate until cleaned up manually
or by owner references. Production clusters should declare it:
every run keeps its pods around until pruned.

- rule: declare exactly one retention rule — keep (newest N) or keep_since (younger than N minutes)

### spec.pruner.schedule

`string` · required

Cron schedule for the prune job (e.g. "0 8 * * *" for daily at
08:00).

- rule: {"required":true}

### spec.pruner.resources

`[]string` · required

Which resource types to prune.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["pipelinerun","taskrun"]}}}}

### spec.pruner.keep

`int32` · optional (explicit presence)

Keep the newest N completed runs (mutually exclusive with
`keep_since` — the operator's webhook enforces exactly one).

- rule: {"int32":{"gte":1}}

### spec.pruner.keepSince

`int32` · optional (explicit presence)

Keep runs younger than this many minutes (mutually exclusive with
`keep`).

- rule: {"int32":{"gte":1}}

### spec.pruner.prunePerResource

`bool`

Run the prune job once per resource kind per namespace instead of
one job for everything (`prune-per-resource`). Tekton default:
false.

### spec.additionalParams

`[]KubernetesTektonParam`

Additional operator params beyond the typed fields (rendered onto
the TektonConfig `spec.params` list — e.g. `createRbacResource`).
The escape surface for operator knobs this spec does not model;
never the primary interface.

### spec.additionalParams[].name

`string` · required

Param name.

- rule: {"required":true}

### spec.additionalParams[].value

`string` · required

Param value.

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesTekton, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the Tekton components run in (the TektonConfig targetNamespace — `tekton-pipelines` unless overridden). |
| `status.outputs.profile` | `string` | The installed profile (`lite`, `basic` or `all`). |
| `status.outputs.dashboard_service` | `string` | Name of the dashboard Service (`tekton-dashboard`) in the target namespace — the backend handle exposure kinds (KubernetesIngress, KubernetesHttpRoute) reference. Empty unless profile is `all`. |
| `status.outputs.dashboard_kube_endpoint` | `string` | In-cluster endpoint of the dashboard (e.g. "http://tekton-dashboard.tekton-pipelines.svc.cluster.local:9097"). Empty unless profile is `all`. |
| `status.outputs.port_forward_command` | `string` | Command to port-forward the dashboard to a workstation (http://localhost:9097). Empty unless profile is `all`. |

## See Also

- [Overview](../README.md)

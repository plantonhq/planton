# KubernetesKyverno

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKyvernoSpec** installs Kyverno — the Kubernetes-native
policy engine — from the official `kyverno` chart
(https://kyverno.github.io/kyverno, chart 3.x = Kyverno 1.18+).

Kyverno validates, mutates, generates, and cleans up resources with
policies written as Kubernetes resources (no new language). This
component installs the ENGINE: four controllers (admission,
background, cleanup, reports) plus the policy CRDs. Policies
themselves (ClusterPolicy / Policy and the policies.kyverno.io v1
families) are declared separately — apply them with
KubernetesManifest resources or GitOps once the engine is running.

WEBHOOK LIFECYCLE — the fact everything else follows from: the
chart templates NO webhook configurations. The admission controller
REGISTERS its ValidatingWebhookConfigurations /
MutatingWebhookConfigurations at runtime (and keeps them updated per
installed policies — the autoUpdateWebhooks feature). Uninstall
runs the chart's pre-delete cleanup hook AND a module-owned
post-release cleanup (the chart's delete-webhooks helper at the
pinned Kyverno release deletes the wrong API — ValidatingAdmissionPolicies
instead of ValidatingWebhookConfigurations — so validating configs
would otherwise survive helm uninstall). If a release is force-deleted
without either path, webhook configurations named `kyverno-*` can
strand and must be deleted by hand
(`kubectl delete validatingwebhookconfiguration,mutatingwebhookconfiguration -l webhook.kyverno.io/managed-by=kyverno`).
Kyverno's own resource webhooks default to failurePolicy=Fail per
policy rule, so a stranded webhook with no backing service can block
admission for the resources it matches.

ONE KYVERNO PER CLUSTER: webhook registration and the policy CRDs
are cluster-global. The conventional namespace is `kyverno`.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed block rendered at once (CRD keep, config filter edits, feature
# flags, all four controllers, HPA, cert-manager certificates,
# ServiceMonitor fan-out, registry override, escape hatch).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKyverno
metadata:
  name: kyverno
spec:
  namespace:
    value: kyverno
  createNamespace: true
  chartVersion: 3.8.2
  crds:
    install: true
    keepOnUninstall: true
    migrationEnabled: true
  config:
    webhookExcludeNamespaces:
      - platform-system
    resourceFiltersInclude:
      - "[Secret,ci-cache,*]"
    resourceFiltersExclude:
      - "[Node,*,*]"
    excludeGroups:
      - system:nodes
    excludeUsernames:
      - "!system:kube-scheduler"
    defaultRegistry: mirror.example.com
    enableDefaultRegistryMutation: false
  features:
    forceFailurePolicyIgnore: true
    backgroundScan:
      enabled: true
      workers: 4
      interval: 30m
    generateValidatingAdmissionPolicy: false
    admissionReports: true
    aggregateReports: true
    policyReports: true
    loggingFormat: json
    loggingVerbosity: 4
    omitEventTypes:
      - PolicyApplied
      - PolicySkipped
  admissionController:
    replicas: 3
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 512Mi
    scheduling:
      nodeSelector:
        role: platform
      tolerations:
        - key: platform
          operator: Exists
          effect: NoSchedule
  backgroundController:
    enabled: true
    replicas: 2
    resources:
      requests:
        cpu: 50m
        memory: 128Mi
  cleanupController:
    enabled: true
    replicas: 2
  reportsController:
    enabled: false
  certificates:
    certManager:
      issuerName:
        value: platform-ca
      issuerKind: ClusterIssuer
  metrics:
    serviceMonitor: true
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  webhooksCleanupEnabled: true
  helmValues: |
    admissionController:
      podLabels:
        team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `3.8.2` |  |
| `spec.crds` | `KubernetesKyvernoCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  |  |  |
| `spec.crds.migrationEnabled` | `bool` |  | `true` |  |
| `spec.config` | `KubernetesKyvernoConfig` |  |  |  |
| `spec.config.webhookExcludeNamespaces` | `[]string` |  |  |  |
| `spec.config.resourceFiltersInclude` | `[]string` |  |  |  |
| `spec.config.resourceFiltersExclude` | `[]string` |  |  |  |
| `spec.config.excludeGroups` | `[]string` |  |  |  |
| `spec.config.excludeUsernames` | `[]string` |  |  |  |
| `spec.config.defaultRegistry` | `string` |  |  |  |
| `spec.config.enableDefaultRegistryMutation` | `bool` |  | `true` |  |
| `spec.features` | `KubernetesKyvernoFeatures` |  |  |  |
| `spec.features.forceFailurePolicyIgnore` | `bool` |  |  |  |
| `spec.features.backgroundScan` | `KubernetesKyvernoBackgroundScan` |  |  |  |
| `spec.features.backgroundScan.enabled` | `bool` |  | `true` |  |
| `spec.features.backgroundScan.workers` | `int32` |  | `2` |  |
| `spec.features.backgroundScan.interval` | `string` |  |  |  |
| `spec.features.generateValidatingAdmissionPolicy` | `bool` |  | `true` |  |
| `spec.features.admissionReports` | `bool` |  | `true` |  |
| `spec.features.aggregateReports` | `bool` |  | `true` |  |
| `spec.features.policyReports` | `bool` |  | `true` |  |
| `spec.features.loggingFormat` | `string` |  |  |  |
| `spec.features.loggingVerbosity` | `int32` |  | `2` |  |
| `spec.features.omitEventTypes` | `[]string` |  |  |  |
| `spec.admissionController` | `KubernetesKyvernoAdmissionController` |  |  |  |
| `spec.admissionController.replicas` | `int32` |  | `1` |  |
| `spec.admissionController.resources` | `ContainerResources` |  |  |  |
| `spec.admissionController.resources.limits` | `CpuMemory` |  |  |  |
| `spec.admissionController.resources.limits.cpu` | `string` |  |  |  |
| `spec.admissionController.resources.limits.memory` | `string` |  |  |  |
| `spec.admissionController.resources.requests` | `CpuMemory` |  |  |  |
| `spec.admissionController.resources.requests.cpu` | `string` |  |  |  |
| `spec.admissionController.resources.requests.memory` | `string` |  |  |  |
| `spec.admissionController.scheduling` | `KubernetesKyvernoScheduling` |  |  |  |
| `spec.admissionController.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.admissionController.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.admissionController.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.admissionController.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.admissionController.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.admissionController.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.admissionController.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.admissionController.autoscaling` | `KubernetesKyvernoHpa` |  |  |  |
| `spec.admissionController.autoscaling.minReplicas` | `int32` |  | `1` |  |
| `spec.admissionController.autoscaling.maxReplicas` | `int32` | yes |  |  |
| `spec.admissionController.autoscaling.targetCpuUtilizationPercentage` | `int32` |  | `80` |  |
| `spec.backgroundController` | `KubernetesKyvernoOptionalController` |  |  |  |
| `spec.backgroundController.enabled` | `bool` |  | `true` |  |
| `spec.backgroundController.replicas` | `int32` |  | `1` |  |
| `spec.backgroundController.resources` | `ContainerResources` |  |  |  |
| `spec.backgroundController.resources.limits` | `CpuMemory` |  |  |  |
| `spec.backgroundController.resources.limits.cpu` | `string` |  |  |  |
| `spec.backgroundController.resources.limits.memory` | `string` |  |  |  |
| `spec.backgroundController.resources.requests` | `CpuMemory` |  |  |  |
| `spec.backgroundController.resources.requests.cpu` | `string` |  |  |  |
| `spec.backgroundController.resources.requests.memory` | `string` |  |  |  |
| `spec.backgroundController.scheduling` | `KubernetesKyvernoScheduling` |  |  |  |
| `spec.backgroundController.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.backgroundController.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.backgroundController.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.backgroundController.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.backgroundController.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.backgroundController.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.backgroundController.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.cleanupController` | `KubernetesKyvernoOptionalController` |  |  |  |
| `spec.cleanupController.enabled` | `bool` |  | `true` |  |
| `spec.cleanupController.replicas` | `int32` |  | `1` |  |
| `spec.cleanupController.resources` | `ContainerResources` |  |  |  |
| `spec.cleanupController.resources.limits` | `CpuMemory` |  |  |  |
| `spec.cleanupController.resources.limits.cpu` | `string` |  |  |  |
| `spec.cleanupController.resources.limits.memory` | `string` |  |  |  |
| `spec.cleanupController.resources.requests` | `CpuMemory` |  |  |  |
| `spec.cleanupController.resources.requests.cpu` | `string` |  |  |  |
| `spec.cleanupController.resources.requests.memory` | `string` |  |  |  |
| `spec.cleanupController.scheduling` | `KubernetesKyvernoScheduling` |  |  |  |
| `spec.cleanupController.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.cleanupController.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.cleanupController.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.cleanupController.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.cleanupController.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.cleanupController.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.cleanupController.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.reportsController` | `KubernetesKyvernoOptionalController` |  |  |  |
| `spec.reportsController.enabled` | `bool` |  | `true` |  |
| `spec.reportsController.replicas` | `int32` |  | `1` |  |
| `spec.reportsController.resources` | `ContainerResources` |  |  |  |
| `spec.reportsController.resources.limits` | `CpuMemory` |  |  |  |
| `spec.reportsController.resources.limits.cpu` | `string` |  |  |  |
| `spec.reportsController.resources.limits.memory` | `string` |  |  |  |
| `spec.reportsController.resources.requests` | `CpuMemory` |  |  |  |
| `spec.reportsController.resources.requests.cpu` | `string` |  |  |  |
| `spec.reportsController.resources.requests.memory` | `string` |  |  |  |
| `spec.reportsController.scheduling` | `KubernetesKyvernoScheduling` |  |  |  |
| `spec.reportsController.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.reportsController.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.reportsController.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.reportsController.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.reportsController.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.reportsController.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.reportsController.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.certificates` | `KubernetesKyvernoCertificates` |  |  |  |
| `spec.certificates.certManager` | `KubernetesKyvernoCertManagerCertificates` |  |  |  |
| `spec.certificates.certManager.issuerName` | `string \| valueFrom` |  |  | KubernetesClusterIssuer (`metadata.name`) |
| `spec.certificates.certManager.issuerKind` | `string` |  |  |  |
| `spec.metrics` | `KubernetesKyvernoMetrics` |  |  |  |
| `spec.metrics.serviceMonitor` | `bool` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.webhooksCleanupEnabled` | `bool` |  | `true` |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into (conventionally "kyverno"). Accepts a
literal namespace name or a reference to a KubernetesNamespace
resource. The webhooks the engine registers EXCLUDE this
namespace by default (config.exclude_kyverno_namespace) so a
misbehaving policy cannot lock Kyverno itself out.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "3.8.2" = Kyverno v1.18.2 —
the chart's appVersion pins every controller image). Versions
must exist in the SERVED index at
https://kyverno.github.io/kyverno.

- default: `3.8.2`

### spec.crds

`KubernetesKyvernoCrds`

Policy CRD lifecycle (the chart's `crds` subchart — ~23 CRDs
across kyverno.io, policies.kyverno.io, reports.kyverno.io and
wgpolicyk8s.io).

### spec.crds.install

`bool` · optional (explicit presence)

Install the policy CRDs with the release. Empty = true (the
chart default). Set false only when another install on the
cluster already owns them (kept CRDs carry the owning release's
Helm metadata — a second release must NOT also install them).

- default: `true`

### spec.crds.keepOnUninstall

`bool`

Keep the CRDs on uninstall (adds the `helm.sh/resource-policy:
keep` annotation). Empty = false (the chart default): destroy
deletes the CRDs, which CASCADE-DELETES every ClusterPolicy,
Policy, PolicyException and policy report on the cluster. Set
true when policies must survive an engine reinstall — but note
the kept CRDs still carry this release's ownership metadata (a
later install must set `install: false`).

### spec.crds.migrationEnabled

`bool` · optional (explicit presence)

Run the chart's post-upgrade CRD migration hook (a kyverno-cli
Job that migrates stored resources to the current schema).
Empty = true (the chart default). The hook image is pulled from
reg.kyverno.io by default — air-gap installs must mirror it or
route it via `image_registry`.

- default: `true`

### spec.config

`KubernetesKyvernoConfig`

The engine's runtime ConfigMap: which resources and principals
the engine SKIPS, and how the webhooks select namespaces.

### spec.config.webhookExcludeNamespaces

`[]string`

Namespaces to add to the webhooks' exclusion selector (the
engine's webhooks then never see those namespaces' resources).
The chart default already excludes kube-system and the Kyverno
namespace itself. Excluding a namespace here is stronger than a
resourceFilter: matching requests never reach the engine at all.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.config.resourceFiltersInclude

`[]string`

Extra resource-filter entries APPENDED to the chart's default
skip list. Each entry is Kyverno's `[Kind,namespace,name]`
triple form, e.g. "[Secret,ci-*,*]" (wildcards allowed). Filtered
resources are invisible to policies in both admission and
background scans.

### spec.config.resourceFiltersExclude

`[]string`

Entries REMOVED from the chart's default resource-filter list —
use to make the engine see something the default skips, e.g. the
kube-system wildcard entry (kind "any/any", namespace
kube-system) to police kube-system. Understand the blast radius
first: control-plane pods that fail policy stop admitting. Each
entry must match a default-list entry byte-for-byte.

### spec.config.excludeGroups

`[]string`

Principal groups the engine ignores entirely. Empty = the chart
default ["system:nodes"]. Declaring the field REPLACES the
default — include "system:nodes" unless kubelet-initiated
requests must be policed.

### spec.config.excludeUsernames

`[]string`

Usernames the engine ignores (e.g. a CI service account that
applies platform manifests policies must not block). Prefix with
`!` to negate. Empty = none (the chart default).

### spec.config.defaultRegistry

`string`

Registry hostname the engine's image mutation rewrites bare image
references to. Empty = "docker.io" (the chart default). Only
meaningful with `enable_default_registry_mutation`.

### spec.config.enableDefaultRegistryMutation

`bool` · optional (explicit presence)

Mutate bare container image references to carry the
`default_registry` hostname. Empty = true (the chart default).

- default: `true`

### spec.features

`KubernetesKyvernoFeatures`

Engine feature flags shared by all controllers.

### spec.features.forceFailurePolicyIgnore

`bool`

Force EVERY policy's failure policy to Ignore, regardless of what
the policy declares. The cluster-wide safety valve: admission
never blocks on a Kyverno outage, at the cost of enforcement
gaps during one. Empty = false (the chart default — policies
decide their own failure policy, which defaults to Fail).

### spec.features.backgroundScan

`KubernetesKyvernoBackgroundScan`

Background scanning — re-evaluates existing resources against
policies on an interval (report-only; enforcement happens at
admission). Omit for the chart default (enabled, 2 workers, 1h
interval).

### spec.features.backgroundScan.enabled

`bool` · optional (explicit presence)

Enable background scanning. Empty = true (the chart default).

- default: `true`

### spec.features.backgroundScan.workers

`int32` · optional (explicit presence)

Concurrent scan workers. Empty = 2 (the chart default).

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.features.backgroundScan.interval

`string`

Scan interval as a Go duration (e.g. "1h", "30m"). Empty = "1h"
(the chart default). Shorter intervals surface drift faster and
cost proportionally more API load.

- rule: {"string":{"pattern":"^$|^([0-9]+(h|m|s))+$"}}

### spec.features.generateValidatingAdmissionPolicy

`bool` · optional (explicit presence)

Generate Kubernetes-native ValidatingAdmissionPolicy objects from
Kyverno policies that declare it (offloads validation to the API
server itself — no webhook round-trip). Empty = true (the chart
default).

- default: `true`

### spec.features.admissionReports

`bool` · optional (explicit presence)

Produce PolicyReport results for admission requests. Empty =
true (the chart default).

- default: `true`

### spec.features.aggregateReports

`bool` · optional (explicit presence)

Aggregate ephemeral reports into the wgpolicyk8s.io
PolicyReport / ClusterPolicyReport resources. Empty = true (the
chart default).

- default: `true`

### spec.features.policyReports

`bool` · optional (explicit presence)

Produce PolicyReport resources at all. Empty = true (the chart
default). Disabling stops report generation engine-wide.

- default: `true`

### spec.features.loggingFormat

`string` · optional (explicit presence)

Logging format. Empty = "text" (the chart default).

- rule: {"string":{"in":["","text","json"]}}

### spec.features.loggingVerbosity

`int32` · optional (explicit presence)

Logging verbosity (0 = quietest). Empty = 2 (the chart default).

- default: `2`
- rule: {"int32":{"lte":10,"gte":0}}

### spec.features.omitEventTypes

`[]string`

Event types the engine does NOT emit as Kubernetes Events. Empty
= the chart default ["PolicyApplied","PolicySkipped"] (emitting
those two floods etcd on busy clusters). Declaring the field
REPLACES the default. Valid values: PolicyViolation,
PolicyApplied, PolicyError, PolicySkipped.

- rule: {"repeated":{"items":{"string":{"in":["PolicyViolation","PolicyApplied","PolicyError","PolicySkipped"]}}}}

### spec.admissionController

`KubernetesKyvernoAdmissionController`

The admission controller — the webhook server that validates and
mutates resources at admission time. Always installed.

### spec.admissionController.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1 (the chart default; the chart
REJECTS 0). Run 3 in production — admission sits on the cluster's
write path, and Kyverno requires ≥2 replicas only when
high-availability matters to you, not functionally.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.admissionController.resources

`ContainerResources`

CPU and memory for the controller container. Empty = the chart
defaults.

### spec.admissionController.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.admissionController.resources.limits.cpu

`string`

### spec.admissionController.resources.limits.memory

`string`

### spec.admissionController.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.admissionController.resources.requests.cpu

`string`

### spec.admissionController.resources.requests.memory

`string`

### spec.admissionController.scheduling

`KubernetesKyvernoScheduling`

Scheduling for the controller pods (node selector +
tolerations; affinity and spread ride `helm_values`).

### spec.admissionController.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller pods.

### spec.admissionController.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods.

### spec.admissionController.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.admissionController.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.admissionController.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.admissionController.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.admissionController.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.admissionController.autoscaling

`KubernetesKyvernoHpa`

Horizontal pod autoscaling for the admission controller.
Mutually exclusive with fixing `replicas` above — when declared,
the HPA owns the replica count.

### spec.admissionController.autoscaling.minReplicas

`int32` · optional (explicit presence)

Minimum replicas. Empty = 1 (the chart default).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.admissionController.autoscaling.maxReplicas

`int32` · required

Maximum replicas.

- rule: {"required":true,"int32":{"gte":1}}

### spec.admissionController.autoscaling.targetCpuUtilizationPercentage

`int32` · optional (explicit presence)

Target CPU utilization percentage. Empty = 80 (the chart
default).

- default: `80`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.backgroundController

`KubernetesKyvernoOptionalController`

The background controller — applies mutate-existing and generate
rules outside the admission path. Omit for the chart default
(enabled, 1 replica).

### spec.backgroundController.enabled

`bool` · optional (explicit presence)

Install this controller. Empty = true (the chart default).
Disabling removes the capability it serves: background =
mutate-existing/generate rules, cleanup = CleanupPolicy
reconciliation, reports = PolicyReport aggregation.

- default: `true`

### spec.backgroundController.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1 (the chart default; the chart
rejects 0 — disable with `enabled: false` instead).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.backgroundController.resources

`ContainerResources`

CPU and memory for the controller container. Empty = the chart
defaults.

### spec.backgroundController.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.backgroundController.resources.limits.cpu

`string`

### spec.backgroundController.resources.limits.memory

`string`

### spec.backgroundController.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.backgroundController.resources.requests.cpu

`string`

### spec.backgroundController.resources.requests.memory

`string`

### spec.backgroundController.scheduling

`KubernetesKyvernoScheduling`

Scheduling for the controller pods (node selector +
tolerations; affinity and spread ride `helm_values`).

### spec.backgroundController.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller pods.

### spec.backgroundController.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods.

### spec.backgroundController.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.backgroundController.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.backgroundController.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.backgroundController.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.backgroundController.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.cleanupController

`KubernetesKyvernoOptionalController`

The cleanup controller — reconciles CleanupPolicy resources
(scheduled deletion of matching resources). Runs its own webhook
for cleanup-policy validation. Omit for the chart default
(enabled, 1 replica).

### spec.cleanupController.enabled

`bool` · optional (explicit presence)

Install this controller. Empty = true (the chart default).
Disabling removes the capability it serves: background =
mutate-existing/generate rules, cleanup = CleanupPolicy
reconciliation, reports = PolicyReport aggregation.

- default: `true`

### spec.cleanupController.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1 (the chart default; the chart
rejects 0 — disable with `enabled: false` instead).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.cleanupController.resources

`ContainerResources`

CPU and memory for the controller container. Empty = the chart
defaults.

### spec.cleanupController.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.cleanupController.resources.limits.cpu

`string`

### spec.cleanupController.resources.limits.memory

`string`

### spec.cleanupController.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.cleanupController.resources.requests.cpu

`string`

### spec.cleanupController.resources.requests.memory

`string`

### spec.cleanupController.scheduling

`KubernetesKyvernoScheduling`

Scheduling for the controller pods (node selector +
tolerations; affinity and spread ride `helm_values`).

### spec.cleanupController.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller pods.

### spec.cleanupController.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods.

### spec.cleanupController.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.cleanupController.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.cleanupController.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.cleanupController.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.cleanupController.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.reportsController

`KubernetesKyvernoOptionalController`

The reports controller — aggregates policy reports
(wgpolicyk8s.io PolicyReports). Omit for the chart default
(enabled, 1 replica). Disable to shed report load on very large
clusters that use an external report store instead.

### spec.reportsController.enabled

`bool` · optional (explicit presence)

Install this controller. Empty = true (the chart default).
Disabling removes the capability it serves: background =
mutate-existing/generate rules, cleanup = CleanupPolicy
reconciliation, reports = PolicyReport aggregation.

- default: `true`

### spec.reportsController.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1 (the chart default; the chart
rejects 0 — disable with `enabled: false` instead).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.reportsController.resources

`ContainerResources`

CPU and memory for the controller container. Empty = the chart
defaults.

### spec.reportsController.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.reportsController.resources.limits.cpu

`string`

### spec.reportsController.resources.limits.memory

`string`

### spec.reportsController.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.reportsController.resources.requests.cpu

`string`

### spec.reportsController.resources.requests.memory

`string`

### spec.reportsController.scheduling

`KubernetesKyvernoScheduling`

Scheduling for the controller pods (node selector +
tolerations; affinity and spread ride `helm_values`).

### spec.reportsController.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller pods.

### spec.reportsController.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods.

### spec.reportsController.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.reportsController.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.reportsController.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.reportsController.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.reportsController.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.certificates

`KubernetesKyvernoCertificates`

TLS certificates for the webhook servers.
Omit for the DEFAULT: Kyverno generates and ROTATES its own CA
and serving certificates at runtime (Secrets named
`<service>.<namespace>.svc.*` in the install namespace) — zero
prerequisites. Declare the cert-manager arm to delegate issuance
to cert-manager instead.

### spec.certificates.certManager

`KubernetesKyvernoCertManagerCertificates`

Delegate certificate issuance to cert-manager: the chart renders
Certificate resources for the admission and cleanup webhook
servers. Requires cert-manager on the cluster
(KubernetesCertManager) BEFORE this install.

### spec.certificates.certManager.issuerName

`string | valueFrom`

Use an existing issuer instead of the self-signed ClusterIssuer
the chart creates by default. Accepts a literal issuer name or a
reference to a KubernetesClusterIssuer resource.

- references: KubernetesClusterIssuer (`metadata.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: metadata.name}} -- a bare string does not parse

### spec.certificates.certManager.issuerKind

`string` · optional (explicit presence)

Kind of the issuer referenced by `issuer_name`. Empty =
"ClusterIssuer". Only meaningful when `issuer_name` is set.

- rule: {"string":{"in":["","ClusterIssuer","Issuer"]}}

### spec.metrics

`KubernetesKyvernoMetrics`

Prometheus metrics collection. Setting `service_monitor` creates
a ServiceMonitor for EVERY enabled controller (all four expose
the toggle) — requires the Prometheus operator CRDs on the
cluster (KubernetesKubePrometheusStack).

### spec.metrics.serviceMonitor

`bool`

Create ServiceMonitor resources for every enabled controller
(admission, background, cleanup, reports — each exposes its own
metrics service). Requires the monitoring.coreos.com CRDs
(KubernetesKubePrometheusStack) on the cluster.

### spec.imageRegistry

`string`

Registry that serves ALL Kyverno images (air-gap / mirror path):
sets the chart's `global.image.registry`, overriding the
per-image registry (ghcr.io) across every controller and hook
container. Repository paths and tags stay chart-managed.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the install namespace) for
pulling Kyverno images from a private mirror. Applied to all
controllers and hook jobs (the chart's global list).

### spec.webhooksCleanupEnabled

`bool` · optional (explicit presence)

Run the chart's pre-delete hook that scales controllers to zero
and attempts to remove the runtime-registered webhook
configurations at uninstall. Empty = true (the chart default) —
LEAVE IT ON. The module also deletes `kyverno-*` webhook
configurations by label after the release is gone (the chart
helper alone is not sufficient at the pinned release); turning
this off skips only the chart's scale-to-zero step and is almost
never what you want. Force-deleted releases still need the manual
unstick on the kind's top-level comment.

- default: `true`

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (per-controller probes, tracing, PDBs,
the grafana dashboard subchart, reports-server, ...) — never the
substitute for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKyverno, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the engine runs in. |
| `status.outputs.release_name` | `string` | Helm release name (equals metadata.name). |
| `status.outputs.admission_service_name` | `string` | Name of the admission controller's webhook Service — the backend the runtime-registered webhook configurations point at (`<release>-svc`). |
| `status.outputs.config_map_name` | `string` | Name of the engine's runtime ConfigMap (resource filters, webhook selectors) — the object to inspect when a resource is unexpectedly skipped or policed. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.certificates.certManager.issuerName` | KubernetesClusterIssuer | `metadata.name` |

## See Also

- [Overview](../README.md)

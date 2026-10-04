# KubernetesOtelOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesOtelOperatorSpec** deploys the OpenTelemetry Operator —
the controller that turns OpenTelemetryCollector declarations into
running collector fleets — from the official `opentelemetry-operator`
Helm chart
(https://open-telemetry.github.io/opentelemetry-helm-charts).

GRAIN: one operator per cluster. It watches every namespace and
reconciles the KubernetesOtelCollector kind's CRs (and, unmodeled
here, Instrumentation/OpAMPBridge/TargetAllocator CRs authored
directly).

CRD LIFECYCLE: the chart templates its opentelemetry.io CRDs as
release-owned objects — a Helm uninstall would cascade-delete every
collector in the cluster. This component instead OWNS the CRD
lifecycle: the modules DERIVE the CRD set from the pinned chart at
apply time (rendering it with the release's own values and the
chart's CRD switch on), apply each CRD outside the release as a kept
resource, and install the release with CRDs skipped. The schema
therefore always matches `chart_version`: a bump re-applies the CRDs
at the new pin; destroy keeps them unless `crds.keep_on_uninstall` is
false; a reinstall re-adopts them; a lower `chart_version` than the
CRDs already on the cluster is refused before anything is touched,
with the remedy. Which CRDs exist follows the chart's own gates
(the feature-gated clusterobservabilities CRD appears only when its
gate is enabled through helm_values).

FAILURES EXPLAIN THEMSELVES: every CRD-lifecycle refusal or error
states what was observed, what it most likely means, and the exact
next step (a chart version that is not published, a repository that
cannot be reached at plan time, a schema downgrade, a render that
produced no CRDs).

WEBHOOK CERTIFICATE: the operator's admission webhooks (they default
and validate collector CRs, failurePolicy Fail) need a serving
certificate, and cert-manager issues and rotates it — which makes
KubernetesCertManager a hard prerequisite of this component. This is
a deliberate consequence of the CRD lifecycle above: the collector
CRD carries a version-CONVERSION webhook, and because the CRDs are
retained past the release's lifetime, their conversion trust must be
kept current by a running reconciler (cert-manager's CA injector,
via the CRDs' cert-manager.io/inject-ca-from annotation). A
certificate embedded once at install time (the chart's Helm-generated
arm) goes stale on rotation and silently breaks collector-CR
conversion long after the install succeeded — so no such arm exists
here.

NAMING: keep the resource name at 30 characters or fewer — the chart
derives a `<name>-controller-manager-service-cert` Secret and
truncates composed names at 63 characters; the modules pin the
chart's fullname to the resource name and fail loudly over the
budget.

The typed fields cover the chart's meaningful surface; `helm_values`
remains the escape hatch (merged last, Helm `-f` semantics, identical
on both engines) — kube-rbac-proxy tuning, network policy, PDB — a
safety valve, never the primary interface.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the operator expressed at once — the explicit
# ClusterIssuer webhook arm, a fleet-wide default-collector-image mirror
# (with tag, exercising the image split), the manager-image registry
# mirror, replicas/resources/ServiceMonitor, pull secrets, full
# scheduling, and an escape-hatch document (PDB) — while the module's two
# re-pinned keys (crds.create=false, certManager.enabled=true) stay
# assertable in the rendered plan.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelOperator
metadata:
  name: otel-operator-full
spec:
  namespace:
    value: otel-operator-system
  createNamespace: true
  chartVersion: 0.120.0
  webhook:
    issuerRef:
      kind: ClusterIssuer
      name: internal-ca
  defaultCollectorImage: mirror.example.com/otel/opentelemetry-collector-k8s:0.156.0
  replicas: 2
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  serviceMonitorEnabled: true
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      workload: system
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    pdb:
      create: true
      minAvailable: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.120.0` |  |
| `spec.webhook` | `KubernetesOtelOperatorWebhook` |  |  |  |
| `spec.webhook.issuerRef` | `KubernetesOtelOperatorIssuerRef` |  |  |  |
| `spec.webhook.issuerRef.kind` | `string` |  |  |  |
| `spec.webhook.issuerRef.name` | `string` | yes |  |  |
| `spec.defaultCollectorImage` | `string` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesOtelOperatorScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |
| `spec.crds` | `KubernetesOtelOperatorCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "0.120.0" — chart 0.120.0
pairs with operator v0.156.0). Versions must exist as SERVED charts
in the repository index
(https://open-telemetry.github.io/opentelemetry-helm-charts).
Bumping the chart version re-applies the module-owned CRDs at the
new pin; lowering it below the CRDs already on the cluster is
refused (see `crds`).

- default: `0.120.0`

### spec.webhook

`KubernetesOtelOperatorWebhook`

Admission-webhook certificate configuration. cert-manager issues
and rotates the certificate and keeps the retained CRDs' conversion
trust current (see the WEBHOOK CERTIFICATE note above — this is why
KubernetesCertManager is a prerequisite). Empty = the chart creates
its own self-signed Issuer, the right choice for almost everyone.

### spec.webhook.issuerRef

`KubernetesOtelOperatorIssuerRef`

Issuer to sign the webhook certificate. Empty = the chart creates
its own self-signed Issuer (the default — right for almost
everyone; the webhook cert only needs to be trusted by the API
server, which cert-manager's CA injection handles).

### spec.webhook.issuerRef.kind

`string`

Issuer kind: "Issuer" or "ClusterIssuer".

- rule: {"string":{"in":["Issuer","ClusterIssuer"]}}

### spec.webhook.issuerRef.name

`string` · required

Issuer name.

- rule: {"required":true}

### spec.defaultCollectorImage

`string`

Default collector image the operator injects into
OpenTelemetryCollector CRs that declare none. Empty = the
operator's compiled-in default
(ghcr.io/open-telemetry/opentelemetry-collector-releases/
opentelemetry-collector-k8s at the operator's paired version).
Override for air-gapped registries or a different collector
distribution fleet-wide.

### spec.replicas

`int32` · optional (explicit presence)

Number of operator replicas. 2 gives a warm standby (leader
election picks one active reconciler).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.resources

`ContainerResources`

CPU and memory for the operator container. Empty = the chart
defaults.

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.serviceMonitorEnabled

`bool`

Render a ServiceMonitor for Prometheus discovery of the operator's
own metrics. Requires the monitoring.coreos.com CRDs on the
cluster (deploy KubernetesKubePrometheusStack first).

### spec.imageRegistry

`string`

Registry that replaces the registry part of the operator manager
image — the air-gap/private-mirror path. Empty = the image's
upstream registry (ghcr.io). This does NOT rewrite the default
collector image the operator injects into CRs — mirror that
fleet-wide via `default_collector_image` (the collector pods pull
it, not this component).

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to the operator pods.
The Secrets must already exist in the namespace.

### spec.scheduling

`KubernetesOtelOperatorScheduling`

Scheduling for the operator pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the operator pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the operator pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

Priority class name for the operator pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (kube-rbac-proxy, network policy, PDB, feature gates) —
never the substitute for them.

### spec.crds

`KubernetesOtelOperatorCrds`

CRD installation lifecycle. Both default to true; unset means the
module owns the CRDs and keeps them.

### spec.crds.install

`bool` · optional (explicit presence)

Derive and apply the opentelemetry.io CRDs from the pinned chart
ahead of the release. Default TRUE. Set false ONLY when the CRDs are
owned elsewhere (a GitOps-managed bundle): the release still installs
with CRDs skipped, and with the CRDs absent the operator cannot start,
so this is a bring-your-own-CRDs arm, never a lighter install.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every OpenTelemetryCollector,
Instrumentation, OpAMPBridge and TargetAllocator in the cluster) when
the resource is destroyed. Default TRUE: deleting the CRDs cascades to
the whole collector fleet, a destructive act that must be an explicit
false. A later reinstall re-adopts kept CRDs.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesOtelOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so the child names below derive from it. |
| `status.outputs.webhook_service` | `string` | name of the operator's webhook Service (`<name>-webhook`, port 443) — where the API server sends admission reviews for collector CRs. |
| `status.outputs.webhook_cert_secret_name` | `string` | name of the Secret holding the webhook serving certificate (`<name>-controller-manager-service-cert`). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)

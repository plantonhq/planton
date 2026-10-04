# KubernetesGatekeeper

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesGatekeeperSpec** installs OPA Gatekeeper — the Open
Policy Agent's Kubernetes admission controller — from the official
`gatekeeper` chart
(https://open-policy-agent.github.io/gatekeeper/charts, chart
version = app version).

Gatekeeper enforces policies written as ConstraintTemplates (Rego
or Kubernetes CEL) instantiated by Constraint resources. This
component installs the ENGINE: the webhook controller manager, the
audit controller, and the constraint-framework CRDs.
ConstraintTemplates and Constraints are declared separately — apply
them with KubernetesManifest resources or GitOps once the engine is
running.

WEBHOOK LIFECYCLE: unlike engines that register webhooks at
runtime, the chart OWNS the ValidatingWebhookConfiguration and
MutatingWebhookConfiguration as release objects — uninstall removes
them with everything else. The policy webhook defaults to
failurePolicy=Ignore (fail-open: an engine outage never blocks
admission), but the namespace-label check webhook — the one that
guards Gatekeeper's own exemption label — defaults to
failurePolicy=Fail.

CRD LIFECYCLE: the engine CRDs (constrainttemplates, configs,
expansiontemplates, mutators, ...) ship in the chart's `crds/`
directory — Helm installs them but NEVER upgrades or deletes them
(kept on uninstall, no release ownership). The chart compensates
for upgrades with its own pre-install/pre-upgrade CRD Job. CRDs
that Gatekeeper creates at runtime from your ConstraintTemplates
(one per template, in constraints.gatekeeper.sh) also survive
uninstall until deleted with their templates.

ONE GATEKEEPER PER CLUSTER: webhook configurations and the engine
CRDs are cluster-global. The conventional namespace is
`gatekeeper-system`.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed block rendered at once (webhook postures both ways, audit tuning,
# exemptions, engine capabilities, external cert, hooks, image override,
# scheduling, escape hatch).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGatekeeper
metadata:
  name: gatekeeper
spec:
  namespace:
    value: gatekeeper-system
  createNamespace: true
  chartVersion: 3.23.0
  replicas: 3
  validatingWebhook:
    enabled: true
    failurePolicy: Fail
    timeoutSeconds: 5
    enableDeleteOperations: true
    checkIgnoreFailurePolicy: Ignore
  mutatingWebhook:
    enabled: true
    failurePolicy: Ignore
    timeoutSeconds: 3
    mutationAnnotations: true
  audit:
    intervalSeconds: 120
    constraintViolationsLimit: 50
    fromCache: true
    matchKindOnly: true
    chunkSize: 200
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        memory: 1Gi
  exemptNamespaces:
    - platform-system
  exemptNamespacePrefixes:
    - kube-
  engine:
    enableExternalData: true
    enableK8sNativeValidation: true
    enableGeneratorResourceExpansion: false
    disabledBuiltins:
      - "{http.send}"
    logDenies: true
    logLevel: DEBUG
  resources:
    requests:
      cpu: 100m
      memory: 512Mi
    limits:
      memory: 512Mi
  scheduling:
    nodeSelector:
      role: platform
    tolerations:
      - key: platform
        operator: Exists
        effect: NoSchedule
  externalCert:
    secretName:
      value: gatekeeper-webhook-server-cert
  hooks:
    labelNamespace: true
    probeWebhook: true
    upgradeCrds: true
    deleteWebhookConfigurationsOnUninstall: true
  image:
    repo: mirror.example.com/gatekeeper
    tag: v3.23.0
    pullSecretName: mirror-pull
  helmValues: |
    podLabels:
      team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `3.23.0` |  |
| `spec.replicas` | `int32` |  | `3` |  |
| `spec.validatingWebhook` | `KubernetesGatekeeperValidatingWebhook` |  |  |  |
| `spec.validatingWebhook.enabled` | `bool` |  | `true` |  |
| `spec.validatingWebhook.failurePolicy` | `string` |  |  |  |
| `spec.validatingWebhook.timeoutSeconds` | `int32` |  | `3` |  |
| `spec.validatingWebhook.enableDeleteOperations` | `bool` |  |  |  |
| `spec.validatingWebhook.checkIgnoreFailurePolicy` | `string` |  |  |  |
| `spec.mutatingWebhook` | `KubernetesGatekeeperMutatingWebhook` |  |  |  |
| `spec.mutatingWebhook.enabled` | `bool` |  | `true` |  |
| `spec.mutatingWebhook.failurePolicy` | `string` |  |  |  |
| `spec.mutatingWebhook.timeoutSeconds` | `int32` |  | `1` |  |
| `spec.mutatingWebhook.mutationAnnotations` | `bool` |  |  |  |
| `spec.audit` | `KubernetesGatekeeperAudit` |  |  |  |
| `spec.audit.intervalSeconds` | `int32` |  | `60` |  |
| `spec.audit.constraintViolationsLimit` | `int32` |  | `20` |  |
| `spec.audit.fromCache` | `bool` |  |  |  |
| `spec.audit.matchKindOnly` | `bool` |  |  |  |
| `spec.audit.chunkSize` | `int32` |  | `500` |  |
| `spec.audit.resources` | `ContainerResources` |  |  |  |
| `spec.audit.resources.limits` | `CpuMemory` |  |  |  |
| `spec.audit.resources.limits.cpu` | `string` |  |  |  |
| `spec.audit.resources.limits.memory` | `string` |  |  |  |
| `spec.audit.resources.requests` | `CpuMemory` |  |  |  |
| `spec.audit.resources.requests.cpu` | `string` |  |  |  |
| `spec.audit.resources.requests.memory` | `string` |  |  |  |
| `spec.exemptNamespaces` | `[]string` |  |  |  |
| `spec.exemptNamespacePrefixes` | `[]string` |  |  |  |
| `spec.engine` | `KubernetesGatekeeperEngine` |  |  |  |
| `spec.engine.enableExternalData` | `bool` |  | `true` |  |
| `spec.engine.enableK8sNativeValidation` | `bool` |  | `true` |  |
| `spec.engine.enableGeneratorResourceExpansion` | `bool` |  | `true` |  |
| `spec.engine.disabledBuiltins` | `[]string` |  |  |  |
| `spec.engine.logDenies` | `bool` |  |  |  |
| `spec.engine.logLevel` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.scheduling` | `KubernetesGatekeeperScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.externalCert` | `KubernetesGatekeeperExternalCert` |  |  |  |
| `spec.externalCert.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`spec.secret_name`) |
| `spec.hooks` | `KubernetesGatekeeperHooks` |  |  |  |
| `spec.hooks.labelNamespace` | `bool` |  | `true` |  |
| `spec.hooks.probeWebhook` | `bool` |  | `true` |  |
| `spec.hooks.upgradeCrds` | `bool` |  | `true` |  |
| `spec.hooks.deleteWebhookConfigurationsOnUninstall` | `bool` |  |  |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into (conventionally "gatekeeper-system").
Accepts a literal namespace name or a reference to a
KubernetesNamespace resource. The chart's post-install hook
labels this namespace with `admission.gatekeeper.sh/ignore` so
the engine never polices itself.

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

Helm chart version to install (e.g. "3.23.0" — chart and app
versions move in lockstep). Versions must exist in the SERVED
index at https://open-policy-agent.github.io/gatekeeper/charts.

- default: `3.23.0`

### spec.replicas

`int32` · optional (explicit presence)

Number of webhook controller-manager replicas. Empty = 3 (the
chart default — the webhook sits on the cluster's write path and
ships HA by default). The audit controller is always a single
replica.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.validatingWebhook

`KubernetesGatekeeperValidatingWebhook`

The validating admission webhook — Gatekeeper's enforcement
path. Omit for the chart defaults.

### spec.validatingWebhook.enabled

`bool` · optional (explicit presence)

Install the validating webhook. Empty = true (the chart
default). Disabling leaves audit-only operation — violations are
RECORDED but never blocked.

- default: `true`

### spec.validatingWebhook.failurePolicy

`string` · optional (explicit presence)

Failure policy when the webhook itself is unreachable. Empty =
"Ignore" (the chart default — fail-OPEN: an engine outage never
blocks cluster admission). "Fail" closes the gap attackers could
time, at the price that a Gatekeeper outage blocks every matched
admission cluster-wide — run 3 replicas and understand the
blast radius before choosing it.

- rule: {"string":{"in":["","Ignore","Fail"]}}

### spec.validatingWebhook.timeoutSeconds

`int32` · optional (explicit presence)

Webhook timeout in seconds. Empty = 3 (the chart default). The
API server caps admission webhooks at 30s; long timeouts with
failure_policy=Fail amplify outage impact.

- default: `3`
- rule: {"int32":{"lte":30,"gte":1}}

### spec.validatingWebhook.enableDeleteOperations

`bool`

Also evaluate DELETE operations. Empty = false (the chart
default). NOTE (chart-truth): the webhook rules this expands are
shared — enabling it widens what reaches the engine.

### spec.validatingWebhook.checkIgnoreFailurePolicy

`string` · optional (explicit presence)

Failure policy of the namespace-label CHECK webhook — the one
that guards who may apply Gatekeeper's own exemption label.
Empty = "Fail" (the chart default — fail-CLOSED so the exemption
label cannot be smuggled on during an outage; scoped to the
label operation only, so the blast radius is namespace edits).

- rule: {"string":{"in":["","Ignore","Fail"]}}

### spec.mutatingWebhook

`KubernetesGatekeeperMutatingWebhook`

The mutating admission webhook — applies Assign /
AssignMetadata / ModifySet mutators. Omit for the chart defaults
(installed, fail-open).

### spec.mutatingWebhook.enabled

`bool` · optional (explicit presence)

Install the mutating webhook. Empty = true (the chart default).
Disable when no mutators are used — one less webhook on the
write path.

- default: `true`

### spec.mutatingWebhook.failurePolicy

`string` · optional (explicit presence)

Failure policy when the webhook is unreachable. Empty = "Ignore"
(the chart default — fail-open).

- rule: {"string":{"in":["","Ignore","Fail"]}}

### spec.mutatingWebhook.timeoutSeconds

`int32` · optional (explicit presence)

Webhook timeout in seconds. Empty = 1 (the chart default).

- default: `1`
- rule: {"int32":{"lte":30,"gte":1}}

### spec.mutatingWebhook.mutationAnnotations

`bool`

Annotate mutated objects with what mutated them
(`gatekeeper.sh/mutation-id` + `gatekeeper.sh/mutations`). Empty
= false (the chart default). Invaluable for debugging mutation
pipelines; costs write amplification.

### spec.audit

`KubernetesGatekeeperAudit`

The audit controller — periodically re-evaluates EXISTING
cluster resources against constraints and records violations in
each constraint's status. Omit for the chart defaults (60s
interval, 20 violations reported per constraint).

### spec.audit.intervalSeconds

`int32` · optional (explicit presence)

Audit interval in seconds. Empty = 60 (the chart default). 0
runs audit exactly once at startup.

- default: `60`
- rule: {"int32":{"gte":0}}

### spec.audit.constraintViolationsLimit

`int32` · optional (explicit presence)

Maximum violations RECORDED per constraint status. Empty = 20
(the chart default). Raising it grows constraint objects in
etcd — prefer an export pipeline for full violation inventories.

- default: `20`
- rule: {"int32":{"gte":1}}

### spec.audit.fromCache

`bool`

Audit from the OPA cache instead of listing the API server
(requires syncing the audited kinds via a Config resource).
Empty = false (the chart default — audit lists directly).

### spec.audit.matchKindOnly

`bool`

Only audit kinds actually matched by some constraint (cuts API
load dramatically on large clusters). Empty = false (the chart
default — audits ALL kinds).

### spec.audit.chunkSize

`int32` · optional (explicit presence)

List/scan chunk size. Empty = 500 (the chart default). 0
disables chunking.

- default: `500`
- rule: {"int32":{"gte":0}}

### spec.audit.resources

`ContainerResources`

CPU and memory for the audit container. Empty = the chart
defaults (512Mi memory limit, 100m CPU request).

### spec.audit.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.audit.resources.limits.cpu

`string`

### spec.audit.resources.limits.memory

`string`

### spec.audit.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.audit.resources.requests.cpu

`string`

### spec.audit.resources.requests.memory

`string`

### spec.exemptNamespaces

`[]string`

Namespaces Gatekeeper exempts from admission enforcement
ENTIRELY (in addition to its own namespace). Exemption still
requires the namespace to carry the
`admission.gatekeeper.sh/ignore` label — these values authorize
which namespaces MAY carry it.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.exemptNamespacePrefixes

`[]string`

Namespace PREFIXES authorized for exemption (e.g. "kube-" to
allow every kube-* namespace to opt out with the ignore label).

### spec.engine

`KubernetesGatekeeperEngine`

Policy-engine capabilities. Omit for the chart defaults.

### spec.engine.enableExternalData

`bool` · optional (explicit presence)

External data providers — constraints may call out to registered
providers (image signature services, CMDB lookups) during
evaluation. Empty = true (the chart default).

- default: `true`

### spec.engine.enableK8sNativeValidation

`bool` · optional (explicit presence)

Kubernetes-native CEL validation in ConstraintTemplates
(K8sNativeValidation alongside Rego). Empty = true (the chart
default).

- default: `true`

### spec.engine.enableGeneratorResourceExpansion

`bool` · optional (explicit presence)

Expand workload generators (Deployments, CronJobs, ...) into the
Pods they would create and evaluate constraints against those
BEFORE the generator is admitted (ExpansionTemplate resources).
Empty = true (the chart default).

- default: `true`

### spec.engine.disabledBuiltins

`[]string`

Rego builtins DISABLED in ConstraintTemplates. Empty = the chart
default ["{http.send}"] (Rego must not make arbitrary network
calls from the admission path — use external data providers
instead). Declaring the field REPLACES the default; include
"{http.send}" unless you accept that risk.

### spec.engine.logDenies

`bool`

Log every admission DENY decision to the controller log (an
audit trail of what enforcement actually blocked). Empty = false
(the chart default).

### spec.engine.logLevel

`string` · optional (explicit presence)

Log level. Empty = "INFO" (the chart default).

- rule: {"string":{"in":["","DEBUG","INFO","WARNING","ERROR"]}}

### spec.resources

`ContainerResources`

CPU and memory for the webhook controller-manager containers.
Empty = the chart defaults (512Mi memory limit, 100m CPU
request).

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

### spec.scheduling

`KubernetesGatekeeperScheduling`

Scheduling for the controller-manager pods (node selector +
tolerations; affinity and spread ride `helm_values`). Empty =
the chart defaults (linux node selector, anti-affinity across
hosts, system-cluster-critical priority).

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller-manager and audit pods.
Empty = the chart default (kubernetes.io/os: linux).

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller-manager and audit pods.

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

### spec.externalCert

`KubernetesGatekeeperExternalCert`

Webhook server TLS. Omit for the DEFAULT: Gatekeeper's embedded
cert-controller generates and ROTATES the CA and serving
certificate in the `gatekeeper-webhook-server-cert` Secret —
zero prerequisites. Declare this block to inject certificates
from an external issuer (cert-manager) instead.

### spec.externalCert.secretName

`string | valueFrom` · required

Name of the TLS Secret (in the install namespace) carrying the
webhook serving certificate — typically materialized by a
cert-manager Certificate whose secretName matches. The Secret
must exist BEFORE the install; the CA it chains to is what the
webhook configurations advertise.

- references: KubernetesCertificate (`spec.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: spec.secret_name}} -- a bare string does not parse

### spec.hooks

`KubernetesGatekeeperHooks`

Lifecycle hook jobs. Omit for the chart defaults (namespace
labeling ON, webhook probe ON, CRD upgrade ON, webhook-config
pre-delete cleanup OFF — the chart-owned webhook configurations
already delete with the release).

### spec.hooks.labelNamespace

`bool` · optional (explicit presence)

Post-install job that labels the install namespace with the
Gatekeeper exemption + Pod Security Standards labels. Empty =
true (the chart default). LEAVE IT ON unless the namespace is
label-managed externally — without the ignore label Gatekeeper
polices its own pods and can deadlock itself. When the module
creates the namespace (create_namespace), it also DECLARES the
exemption label on the namespace object itself so day-2 applies
never strip what the hook stamped.

- default: `true`

### spec.hooks.probeWebhook

`bool` · optional (explicit presence)

Post-install probe job (curl) that waits until the webhook
endpoint actually serves before the release reports installed.
Empty = true (the chart default).

- default: `true`

### spec.hooks.upgradeCrds

`bool` · optional (explicit presence)

Pre-install/pre-upgrade job that applies the engine CRDs at the
chart's version (Helm never upgrades `crds/`-directory CRDs on
its own — this hook is the chart's own answer; disabling it
leaves CRDs frozen at first-install schema). Empty = true (the
chart default).

- default: `true`

### spec.hooks.deleteWebhookConfigurationsOnUninstall

`bool`

Pre-delete job that explicitly deletes the webhook
configurations before the release objects go. Empty = false (the
chart default — the chart-owned webhook configurations already
delete with the release; enable for URL-mode webhooks or
belt-and-suspenders teardown ordering). The modules render the
hook's service-account name alongside the toggle — the chart's
own RBAC binding for this job requires it (enabling the raw chart
value alone fails at uninstall).

### spec.image

`ContainerImage`

Override the Gatekeeper image (air-gap / private-mirror path).
Empty = openpolicyagent/gatekeeper at the chart's release tag.
The chart also runs hook containers from
openpolicyagent/gatekeeper-crds and curlimages/curl — mirror
those too for air-gapped installs (set via `helm_values`).

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (PDB, network policy, PSS label sets,
violation export, per-hook images, ...) — never the substitute
for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGatekeeper, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the engine runs in. |
| `status.outputs.release_name` | `string` | Helm release name (equals metadata.name). |
| `status.outputs.webhook_service_name` | `string` | Name of the webhook Service the ValidatingWebhookConfiguration / MutatingWebhookConfiguration point at (`gatekeeper-webhook-service`). |
| `status.outputs.webhook_cert_secret_name` | `string` | Name of the Secret carrying the webhook server certificate (`gatekeeper-webhook-server-cert`) — chart-fixed; rotated by the embedded cert-controller unless an external certificate is injected. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.externalCert.secretName` | KubernetesCertificate | `spec.secret_name` |

## See Also

- [Overview](../README.md)

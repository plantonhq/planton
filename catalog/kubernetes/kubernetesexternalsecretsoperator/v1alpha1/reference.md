# KubernetesExternalSecretsOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesExternalSecretsOperatorSpec** installs the External Secrets
Operator (ESO) — the controller that syncs secrets FROM external stores
(AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao,
...) INTO Kubernetes Secrets — from the official Helm chart
(`external-secrets` at https://charts.external-secrets.io).

This component installs and configures the OPERATOR MACHINERY only. Which
stores exist and which secrets sync are separate first-class resources:
create KubernetesClusterSecretStore / KubernetesSecretStore for the
backend connections and KubernetesExternalSecret for each secret to sync.
One operator installation per cluster serves all of them (the release name
is fixed to "external-secrets": the CRDs and webhook configuration are
cluster-global, an upstream architectural constraint).

The operator runs three components: the controller (reconciles stores and
external secrets), the webhook (validates ESO resources at admission), and
the cert-controller (bootstraps the webhook's serving certificate).

The typed fields cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for values beyond them (merged
last, Helm `-f` semantics, identical on both engines) — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises the CRD lifecycle knobs,
# HA controller with leader election, reconcile tuning, sharding and
# scoping, workload identity, scheduling, sizing, per-kind tuning,
# observability, image override, and the helm_values escape hatch — so the
# offline tofu plan and pulumi preview proofs cover arms the live kind
# lanes exclude. Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalSecretsOperator
metadata:
  name: hack-external-secrets
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  chartVersion: 2.8.0
  crds:
    install: true
    keepOnUninstall: true
  replicas: 2
  leaderElect: true
  concurrent: 5
  controllerClass: platform
  logLevel: debug
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  workloadIdentity:
    gke:
      serviceAccountEmail:
        value: external-secrets@my-project.iam.gserviceaccount.com
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: platform
      effect: NoSchedule
  priorityClassName: system-cluster-critical
  podDisruptionBudget: true
  prometheus:
    serviceMonitor: true
    serviceMonitorInterval: 30s
    serviceMonitorLabels:
      release: kube-prometheus-stack
  webhook:
    enabled: true
    replicas: 2
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
  certController:
    enabled: true
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
  imageRepository: registry.example.com/external-secrets/external-secrets
  helmValues: |
    revisionHistoryLimit: 3
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.8.0` |  |
| `spec.crds` | `KubernetesExternalSecretsOperatorCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.leaderElect` | `bool` |  |  |  |
| `spec.concurrent` | `int32` |  | `1` |  |
| `spec.controllerClass` | `string` |  |  |  |
| `spec.scopedNamespace` | `string` |  |  |  |
| `spec.scopedRbac` | `bool` |  |  |  |
| `spec.logLevel` | `string` |  | `info` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.workloadIdentity.aks.tenantId` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.priorityClassName` | `string` |  |  |  |
| `spec.podDisruptionBudget` | `bool` |  |  |  |
| `spec.prometheus` | `KubernetesExternalSecretsOperatorPrometheus` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitorInterval` | `string` |  | `30s` |  |
| `spec.prometheus.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.webhook` | `KubernetesExternalSecretsOperatorWebhook` |  |  |  |
| `spec.webhook.enabled` | `bool` |  | `true` |  |
| `spec.webhook.replicas` | `int32` |  | `1` |  |
| `spec.webhook.resources` | `ContainerResources` |  |  |  |
| `spec.webhook.resources.limits` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.limits.cpu` | `string` |  |  |  |
| `spec.webhook.resources.limits.memory` | `string` |  |  |  |
| `spec.webhook.resources.requests` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.requests.cpu` | `string` |  |  |  |
| `spec.webhook.resources.requests.memory` | `string` |  |  |  |
| `spec.certController` | `KubernetesExternalSecretsOperatorCertController` |  |  |  |
| `spec.certController.enabled` | `bool` |  | `true` |  |
| `spec.certController.replicas` | `int32` |  | `1` |  |
| `spec.certController.resources` | `ContainerResources` |  |  |  |
| `spec.certController.resources.limits` | `CpuMemory` |  |  |  |
| `spec.certController.resources.limits.cpu` | `string` |  |  |  |
| `spec.certController.resources.limits.memory` | `string` |  |  |  |
| `spec.certController.resources.requests` | `CpuMemory` |  |  |  |
| `spec.certController.resources.requests.cpu` | `string` |  |  |  |
| `spec.certController.resources.requests.memory` | `string` |  |  |  |
| `spec.imageRepository` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the operator into ("external-secrets" by
convention). Accepts a literal namespace name or a reference to a
KubernetesNamespace resource.

Treat the namespace as PERMANENT while CRDs are kept (the
`crds.keep_on_uninstall` default): kept CRDs retain the Helm release's
namespace in their ownership metadata, so re-installing into a
DIFFERENT namespace fails with Helm's release-ownership error on the
surviving CRDs. Moving an install requires first deleting the kept
CRDs — which cascades to every ExternalSecret and SecretStore object
cluster-wide.

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

Helm chart version to install (chart and operator versions are aligned
upstream, e.g. "2.8.0" ships operator v2.8.0). Pin deliberately;
upgrades re-run the release with the new chart. Pick versions from the
chart repository's index (`helm search repo`): the served chart is the
contract — the upstream source tree's Chart.yaml can claim a version at
a tag that was never served.

- default: `2.8.0`

### spec.crds

`KubernetesExternalSecretsOperatorCrds`

CRD lifecycle. ESO's CRDs (ExternalSecret, SecretStore,
ClusterSecretStore, ...) are cluster-scoped and shared: installing them
with the release is the standard single-installation path.

### spec.crds.install

`bool` · optional (explicit presence)

Install the ESO CRDs with the release. Default TRUE — one component
managing both halves is strictly simpler. Disable only when another
installation already owns the CRDs.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every ExternalSecret/SecretStore object in
the cluster) when the release is uninstalled. Default TRUE (Planton
opinion, rendered as the helm.sh/resource-policy=keep annotation on the
CRDs — the chart itself has no keep knob and would DELETE them, which
cascades to every ESO object cluster-wide, a destructive act that should
require an explicit false).

- default: `true`

### spec.replicas

`int32` · optional (explicit presence)

Controller replica count. One replica is standard; with more than one,
enable leader_elect so exactly one reconciles.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.leaderElect

`bool`

Leader election for the controller. Chart default: false (fine for one
replica). Required for replicas > 1 — without it every replica
reconciles and they race.

### spec.concurrent

`int32` · optional (explicit presence)

ExternalSecret resources reconciled concurrently. Chart default: 1.
Raise for clusters with hundreds of ExternalSecrets where sync latency
matters.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.controllerClass

`string`

Only reconcile stores whose spec.controller matches this class — the
sharding knob for running several isolated operator installations in one
cluster. Empty = reconcile every store (the standard single-operator
posture).

### spec.scopedNamespace

`string`

Restrict the operator to ONE namespace (it only watches and reconciles
resources there). Pair with scoped_rbac for a Role instead of a
ClusterRole. Empty = cluster-wide (the standard posture).

### spec.scopedRbac

`bool`

With scoped_namespace: grant a namespace Role instead of a ClusterRole.
The fully-scoped multi-tenant posture.

### spec.logLevel

`string` · optional (explicit presence)

Log verbosity for all three components. Chart default: "info".

- default: `info`
- rule: {"string":{"in":["debug","info","error"]}}

### spec.resources

`ContainerResources`

Controller container CPU/memory requests and limits. Empty = chart
defaults (no explicit resources).

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

### spec.workloadIdentity

`KubernetesWorkloadIdentity`

Binds the CONTROLLER ServiceAccount to a cloud identity for keyless
store access (AWS via EKS IRSA, GCP via GKE Workload Identity, Azure via
AKS Workload Identity). Stores that leave their auth block empty
authenticate through this ambient identity — the simplest posture when
one cloud identity may read every synced secret. Per-store identities
(finer-grained, recommended for multi-team clusters) instead reference
dedicated ServiceAccounts in each store's auth block; those need nothing
here.

### spec.workloadIdentity.gke

`KubernetesWorkloadIdentityGke`

GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.

### spec.workloadIdentity.gke.serviceAccountEmail

`string | valueFrom` · required

GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
email or a reference to a GcpServiceAccount resource's output.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.workloadIdentity.eks

`KubernetesWorkloadIdentityEksIrsa`

EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.

### spec.workloadIdentity.eks.roleArn

`string | valueFrom` · required

AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.workloadIdentity.aks

`KubernetesWorkloadIdentityAks`

Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
(or user-assigned managed identity) client ID.

### spec.workloadIdentity.aks.clientId

`string | valueFrom` · required

Client ID (GUID) of the user-assigned managed identity or Entra application.
Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
GUID or a reference to an AzureUserAssignedIdentity resource's output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.workloadIdentity.aks.tenantId

`string` · optional (explicit presence)

Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
omitted the azure-workload-identity webhook uses its default tenant. Applied as
the `azure.workload.identity/tenant-id` annotation when set.

- rule: {"string":{"uuid":true}}

### spec.nodeSelector

`map<string, string>`

Node selector for all operator pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods (webhook/cert-controller inherit
their own chart defaults; set component-level tolerations via
helm_values when they must differ).

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.priorityClassName

`string`

PriorityClass for the controller pod.

### spec.podDisruptionBudget

`bool`

When true, a PodDisruptionBudget guards the controller (minAvailable 1)
— meaningful only with replicas > 1.

### spec.prometheus

`KubernetesExternalSecretsOperatorPrometheus`

Prometheus metrics exposure. The metrics port is always served; the
ServiceMonitor is opt-in and requires the Prometheus operator CRDs on
the cluster.

### spec.prometheus.serviceMonitor

`bool`

Create ServiceMonitor resources for scrape discovery. Requires the
Prometheus operator CRDs (e.g. kube-prometheus-stack) on the cluster —
the release FAILS to install without them.

### spec.prometheus.serviceMonitorInterval

`string` · optional (explicit presence)

Scrape interval for the ServiceMonitor. Chart default: "30s".

- default: `30s`

### spec.prometheus.serviceMonitorLabels

`map<string, string>`

Extra labels on the ServiceMonitor — how a Prometheus instance's
serviceMonitorSelector finds it (e.g. {"release": "kube-prometheus-stack"}).

### spec.webhook

`KubernetesExternalSecretsOperatorWebhook`

Webhook component tuning. The webhook must be reachable by the API
server; it validates every ESO resource at admission.

### spec.webhook.enabled

`bool` · optional (explicit presence)

Run the webhook. Chart default: true. Disabling leaves ESO resources
unvalidated at admission — misconfigured stores/secrets then fail at
reconcile time instead of at apply time.

- default: `true`

### spec.webhook.replicas

`int32` · optional (explicit presence)

Webhook replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.webhook.resources

`ContainerResources`

Webhook container CPU/memory requests and limits.

### spec.webhook.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.webhook.resources.limits.cpu

`string`

### spec.webhook.resources.limits.memory

`string`

### spec.webhook.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.webhook.resources.requests.cpu

`string`

### spec.webhook.resources.requests.memory

`string`

### spec.certController

`KubernetesExternalSecretsOperatorCertController`

cert-controller component tuning. It bootstraps and rotates the
webhook's serving certificate — disable only when cert-manager (via
helm_values webhook.certManager) or another mechanism owns webhook
certificates.

### spec.certController.enabled

`bool` · optional (explicit presence)

Run the cert-controller. Chart default: true; the webhook's serving
certificate depends on it (unless cert-manager integration via
helm_values takes over — the chart then skips this Deployment on its
own).

- default: `true`

### spec.certController.replicas

`int32` · optional (explicit presence)

cert-controller replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.certController.resources

`ContainerResources`

cert-controller container CPU/memory requests and limits.

### spec.certController.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.certController.resources.limits.cpu

`string`

### spec.certController.resources.limits.memory

`string`

### spec.certController.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.certController.resources.requests.cpu

`string`

### spec.certController.resources.requests.memory

`string`

### spec.imageRepository

`string`

Full image repository override (registry + path) for the operator
images — the air-gapped/mirror knob. Empty = the chart default
(ghcr.io/external-secrets/external-secrets).

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields — never
the substitute for them. Do not put secrets here.

## Validation Rules

- `eso.replicas_require_leader_election`: replicas > 1 requires leader_elect — without leader election every replica reconciles the same resources and they race
- `eso.scoped_rbac_requires_namespace`: scoped_rbac narrows RBAC to scoped_namespace — set scoped_namespace when enabling it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesExternalSecretsOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the operator is installed in. Store resources reference it as the home for credential Secrets that cluster-scoped stores read. |
| `status.outputs.release_name` | `string` | Helm release name (always "external-secrets" — one installation per cluster is an upstream architectural constraint). |
| `status.outputs.controller_service_account` | `string` | Name of the controller ServiceAccount. The cloud-side half of a keyless ambient-identity binding (IAM role trust policy, Workload Identity binding, federated credential) references this name together with the namespace. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.workloadIdentity.gke.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.workloadIdentity.eks.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.workloadIdentity.aks.clientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesClusterSecretStore | `spec.secretsNamespace` | `status.outputs.namespace` |

## See Also

- [Overview](../README.md)

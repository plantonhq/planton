# KubernetesKeda

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesKedaSpec** installs KEDA — Kubernetes Event-Driven Autoscaling
— from the official Helm chart (`keda` at
https://kedacore.github.io/charts). KEDA scales workloads on REAL-WORLD
signals (queue depth, stream lag, database rows, cron schedules, cloud
metrics — 70+ scalers) instead of only CPU/memory: its operator watches
ScaledObject/ScaledJob resources, drives the workload's HPA (including
scale-to-ZERO, which plain HPA cannot do), and serves the
external.metrics.k8s.io API the HPA controller reads.

ONE INSTALLATION PER CLUSTER: KEDA registers the cluster-wide
`v1beta1.external.metrics.k8s.io` APIService, a singleton — a second
installation would fight over it (and Kubernetes allows only one external
metrics provider). The Helm release name is therefore fixed to "keda".

The scaling declarations themselves (ScaledObject, ScaledJob,
TriggerAuthentication) are KEDA custom resources applied per workload —
deploy them alongside the workloads they scale (KubernetesManifest carries
them today). This component installs and configures the ENGINE.

The typed fields below cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for chart values beyond them
(merged last, Helm `-f` semantics, identical on both engines) — a safety
valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Not a realistic
# production shape — see presets for those.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeda
metadata:
  name: hack-keda
spec:
  namespace:
    value: hack-keda
  createNamespace: true
  chartVersion: "2.20.1"
  crds:
    install: true
    keepOnUninstall: true
  watchNamespace: hack-keda-apps
  operator:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: "1"
        memory: 1000Mi
  metricsServer:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        memory: 1000Mi
  webhooks:
    enabled: true
    failurePolicy: Fail
    replicas: 2
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        memory: 200Mi
  podIdentity:
    awsIrsa:
      enabled: true
      roleArn: arn:aws:iam::123456789012:role/hack-keda-scalers
  certificates:
    type: cert_manager
    certManagerIssuer:
      clusterIssuer:
        name:
          value: hack-cluster-issuer
  httpTimeoutMs: 5000
  priorityClassName: system-cluster-critical
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: node-role.kubernetes.io/control-plane
      operator: Exists
      effect: NoSchedule
  prometheus:
    enabled: true
    serviceMonitor: false
  helmValues: |
    asciiArt: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.20.1` |  |
| `spec.crds` | `KubernetesKedaCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.watchNamespace` | `string` |  |  |  |
| `spec.operator` | `KubernetesKedaComponent` |  |  |  |
| `spec.operator.replicas` | `int32` |  | `1` |  |
| `spec.operator.resources` | `ContainerResources` |  |  |  |
| `spec.operator.resources.limits` | `CpuMemory` |  |  |  |
| `spec.operator.resources.limits.cpu` | `string` |  |  |  |
| `spec.operator.resources.limits.memory` | `string` |  |  |  |
| `spec.operator.resources.requests` | `CpuMemory` |  |  |  |
| `spec.operator.resources.requests.cpu` | `string` |  |  |  |
| `spec.operator.resources.requests.memory` | `string` |  |  |  |
| `spec.metricsServer` | `KubernetesKedaComponent` |  |  |  |
| `spec.metricsServer.replicas` | `int32` |  | `1` |  |
| `spec.metricsServer.resources` | `ContainerResources` |  |  |  |
| `spec.metricsServer.resources.limits` | `CpuMemory` |  |  |  |
| `spec.metricsServer.resources.limits.cpu` | `string` |  |  |  |
| `spec.metricsServer.resources.limits.memory` | `string` |  |  |  |
| `spec.metricsServer.resources.requests` | `CpuMemory` |  |  |  |
| `spec.metricsServer.resources.requests.cpu` | `string` |  |  |  |
| `spec.metricsServer.resources.requests.memory` | `string` |  |  |  |
| `spec.webhooks` | `KubernetesKedaWebhooks` |  |  |  |
| `spec.webhooks.enabled` | `bool` |  | `true` |  |
| `spec.webhooks.failurePolicy` | `string` |  | `Ignore` |  |
| `spec.webhooks.replicas` | `int32` |  | `1` |  |
| `spec.webhooks.resources` | `ContainerResources` |  |  |  |
| `spec.webhooks.resources.limits` | `CpuMemory` |  |  |  |
| `spec.webhooks.resources.limits.cpu` | `string` |  |  |  |
| `spec.webhooks.resources.limits.memory` | `string` |  |  |  |
| `spec.webhooks.resources.requests` | `CpuMemory` |  |  |  |
| `spec.webhooks.resources.requests.cpu` | `string` |  |  |  |
| `spec.webhooks.resources.requests.memory` | `string` |  |  |  |
| `spec.podIdentity` | `KubernetesKedaPodIdentity` |  |  |  |
| `spec.podIdentity.awsIrsa` | `KubernetesKedaAwsIrsa` |  |  |  |
| `spec.podIdentity.awsIrsa.enabled` | `bool` |  |  |  |
| `spec.podIdentity.awsIrsa.roleArn` | `string` |  |  |  |
| `spec.podIdentity.azureWorkloadIdentity` | `KubernetesKedaAzureWorkloadIdentity` |  |  |  |
| `spec.podIdentity.azureWorkloadIdentity.enabled` | `bool` |  |  |  |
| `spec.podIdentity.azureWorkloadIdentity.clientId` | `string` |  |  |  |
| `spec.podIdentity.azureWorkloadIdentity.tenantId` | `string` |  |  |  |
| `spec.podIdentity.gcpWorkloadIdentity` | `KubernetesKedaGcpWorkloadIdentity` |  |  |  |
| `spec.podIdentity.gcpWorkloadIdentity.enabled` | `bool` |  |  |  |
| `spec.podIdentity.gcpWorkloadIdentity.serviceAccountEmail` | `string` |  |  |  |
| `spec.certificates` | `KubernetesKedaCertificates` |  |  |  |
| `spec.certificates.type` | `string` |  | `operator` |  |
| `spec.certificates.certManagerIssuer` | `CertManagerIssuerRef` |  |  |  |
| `spec.certificates.certManagerIssuer.issuer` | `CertManagerNamespacedIssuerRef` |  |  |  |
| `spec.certificates.certManagerIssuer.issuer.name` | `string \| valueFrom` | yes |  | KubernetesIssuer (`status.outputs.issuer_name`) |
| `spec.certificates.certManagerIssuer.clusterIssuer` | `CertManagerClusterIssuerRef` |  |  |  |
| `spec.certificates.certManagerIssuer.clusterIssuer.name` | `string \| valueFrom` | yes |  | KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`) |
| `spec.httpTimeoutMs` | `int32` |  | `3000` |  |
| `spec.priorityClassName` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.prometheus` | `KubernetesKedaPrometheus` |  |  |  |
| `spec.prometheus.enabled` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install KEDA into ("keda" is the upstream convention).
Accepts a literal namespace name or a reference to a
KubernetesNamespace resource.

Treat the namespace as PERMANENT while CRDs are kept (the
`crds.keep_on_uninstall` default): kept CRDs retain the Helm release's
namespace in their ownership metadata, so re-installing into a
DIFFERENT namespace fails with Helm's release-ownership error on the
surviving CRDs. Moving an install requires first deleting the kept
CRDs — which cascades to every scaling declaration in the cluster
(ScaledObjects, ScaledJobs, TriggerAuthentications).

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

Helm chart version to install (e.g. "2.20.1", which ships KEDA 2.20.1 —
chart and app versions move together). Pin deliberately; upgrades
re-run the release with the new chart. Pick versions from the chart
repository's index (`helm search repo`): the served chart is the
contract — the upstream source tree's Chart.yaml can claim a version at
a tag that was never served.

- default: `2.20.1`

### spec.crds

`KubernetesKedaCrds`

KEDA custom resource definitions (ScaledObject, ScaledJob,
TriggerAuthentication, ...) lifecycle.

- rule: keep_on_uninstall only applies when the release installs the CRDs — with install false there is nothing to keep

### spec.crds.install

`bool` · optional (explicit presence)

Install the KEDA CRDs with the release. Chart default: true. Disable
only when something else manages them.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every ScaledObject/ScaledJob/
TriggerAuthentication in the cluster) when this release is
uninstalled. Rendered as the `helm.sh/resource-policy: keep` annotation
on the CRDs — the chart has no native keep knob, and WITHOUT this a
plain uninstall cascade-deletes every scaling declaration in the
cluster. Leave true unless you deliberately want uninstall to purge
them.

- default: `true`

### spec.watchNamespace

`string`

Namespace KEDA watches for ScaledObjects/ScaledJobs. Empty (the chart
default) watches ALL namespaces — the normal cluster-wide posture. Set
a single namespace to fence KEDA into one team's space.

### spec.operator

`KubernetesKedaComponent`

keda-operator sizing — the controller that reconciles ScaledObjects
and drives HPAs. Extra replicas are warm standbys (leader-elected),
not horizontal capacity.

### spec.operator.replicas

`int32` · optional (explicit presence)

Replica count. Chart default: 1. Only one instance leads/serves at a
time — extra replicas are failover standbys, per upstream's HA
guidance.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.operator.resources

`ContainerResources`

Container resources. Empty = chart defaults (requests 100m/100Mi,
limits 1/1000Mi).

### spec.operator.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.operator.resources.limits.cpu

`string`

### spec.operator.resources.limits.memory

`string`

### spec.operator.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.operator.resources.requests.cpu

`string`

### spec.operator.resources.requests.memory

`string`

### spec.metricsServer

`KubernetesKedaComponent`

Metrics API server sizing — serves external.metrics.k8s.io to the HPA
controller. Like the operator, one instance serves traffic; extra
replicas reduce failover downtime.

### spec.metricsServer.replicas

`int32` · optional (explicit presence)

Replica count. Chart default: 1. Only one instance leads/serves at a
time — extra replicas are failover standbys, per upstream's HA
guidance.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.metricsServer.resources

`ContainerResources`

Container resources. Empty = chart defaults (requests 100m/100Mi,
limits 1/1000Mi).

### spec.metricsServer.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.metricsServer.resources.limits.cpu

`string`

### spec.metricsServer.resources.limits.memory

`string`

### spec.metricsServer.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.metricsServer.resources.requests.cpu

`string`

### spec.metricsServer.resources.requests.memory

`string`

### spec.webhooks

`KubernetesKedaWebhooks`

Admission webhooks — validate ScaledObjects at apply time (catching
broken scale targets and conflicting HPAs before they misbehave at
runtime).

### spec.webhooks.enabled

`bool` · optional (explicit presence)

Deploy the admission webhooks. Chart default: true. Disabling removes
apply-time validation of ScaledObjects (mistakes then surface only as
runtime scaling failures).

- default: `true`

### spec.webhooks.failurePolicy

`string` · optional (explicit presence)

What the API server does when the webhook is unreachable: "Ignore"
(chart default — applies proceed unvalidated) or "Fail" (applies are
rejected until the webhook is back; stricter, but a webhook outage
then blocks ScaledObject changes).

- default: `Ignore`
- rule: failure_policy must be either 'Ignore' or 'Fail'

### spec.webhooks.replicas

`int32` · optional (explicit presence)

Replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.webhooks.resources

`ContainerResources`

Container resources. Empty = chart defaults.

### spec.webhooks.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.webhooks.resources.limits.cpu

`string`

### spec.webhooks.resources.limits.memory

`string`

### spec.webhooks.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.webhooks.resources.requests.cpu

`string`

### spec.webhooks.resources.requests.memory

`string`

### spec.podIdentity

`KubernetesKedaPodIdentity`

Ambient cloud identity for SCALERS: how KEDA authenticates to cloud
metric sources (SQS queue depth, Azure Service Bus, GCP Pub/Sub, ...)
without stored credentials. This annotates/labels KEDA's own service
accounts; per-trigger authentication beyond it lives in
TriggerAuthentication resources next to the workloads.

### spec.podIdentity.awsIrsa

`KubernetesKedaAwsIrsa`

AWS IRSA: annotate KEDA's service account with an IAM role so scalers
read SQS/CloudWatch/Kinesis without stored keys.

- rule: aws_irsa requires role_arn — without the role annotation there is no identity to assume

### spec.podIdentity.awsIrsa.enabled

`bool`

Enable IRSA annotations on the KEDA service account.

### spec.podIdentity.awsIrsa.roleArn

`string`

ARN of the IAM role to assume (the role's trust policy must allow the
cluster's OIDC provider and KEDA's service account).

- rule: role_arn must be an IAM role ARN (arn:aws:iam::<account>:role/<name>)

### spec.podIdentity.azureWorkloadIdentity

`KubernetesKedaAzureWorkloadIdentity`

Azure Workload Identity: label/annotate for federated Entra identity
so scalers read Service Bus/Event Hubs/Monitor without stored secrets.

- rule: azure_workload_identity requires both client_id and tenant_id — the federated credential is addressed by the pair

### spec.podIdentity.azureWorkloadIdentity.enabled

`bool`

Enable the Azure Workload Identity label/annotations on the KEDA
service account.

### spec.podIdentity.azureWorkloadIdentity.clientId

`string`

Entra application (client) ID KEDA federates as.

### spec.podIdentity.azureWorkloadIdentity.tenantId

`string`

Entra tenant ID of that application.

### spec.podIdentity.gcpWorkloadIdentity

`KubernetesKedaGcpWorkloadIdentity`

GCP Workload Identity: bind a GCP service account so scalers read
Pub/Sub/Stackdriver without stored keys.

- rule: gcp_workload_identity requires service_account_email — without it there is no GCP identity to impersonate

### spec.podIdentity.gcpWorkloadIdentity.enabled

`bool`

Enable the GCP Workload Identity annotation on the KEDA service
account.

### spec.podIdentity.gcpWorkloadIdentity.serviceAccountEmail

`string`

Email of the GCP IAM service account to impersonate (must carry a
workload-identity binding to KEDA's Kubernetes service account).

- rule: service_account_email must be a GCP service-account email (…@…gserviceaccount.com)

### spec.certificates

`KubernetesKedaCertificates`

How KEDA's internal TLS certificates (operator ↔ metrics server ↔
webhooks) are provisioned.

- rule: cert_manager_issuer is only used with certificates type cert_manager — set type accordingly or remove the issuer

### spec.certificates.type

`string` · optional (explicit presence)

Provisioning method. "operator" (chart default): the KEDA operator
self-generates certificates and patches the APIService caBundle —
zero-dependency and fine for almost every cluster. "cert_manager":
cert-manager issues and renews them (requires KubernetesCertManager on
the cluster).

- default: `operator`
- rule: certificates type must be either 'operator' or 'cert_manager'

### spec.certificates.certManagerIssuer

`CertManagerIssuerRef`

cert-manager issuer that signs KEDA's certificates (type cert_manager).
Empty = the chart generates its own self-signed CA + Issuer chain.

The grain is the arm: `issuer` (a namespaced Issuer in the installation
namespace) or `cluster_issuer` (a cluster-scoped ClusterIssuer).

- rule: An issuer is required -- choose 'issuer' (a namespaced Issuer in the installation namespace) or 'cluster_issuer' (a cluster-scoped ClusterIssuer)

### spec.certificates.certManagerIssuer.issuer

`CertManagerNamespacedIssuerRef`

A namespaced Issuer in the installation namespace, by name.

### spec.certificates.certManagerIssuer.issuer.name

`string | valueFrom` · required

Issuer name. Accepts a literal name or a reference to a KubernetesIssuer
resource.

- references: KubernetesIssuer (`status.outputs.issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesIssuer, name: <that resource's name>, fieldPath: status.outputs.issuer_name}} -- a bare string does not parse

### spec.certificates.certManagerIssuer.clusterIssuer

`CertManagerClusterIssuerRef`

A cluster-scoped ClusterIssuer, by name.

### spec.certificates.certManagerIssuer.clusterIssuer.name

`string | valueFrom` · required

ClusterIssuer name. Accepts a literal name or a reference to a
KubernetesClusterIssuer resource.

- references: KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: status.outputs.cluster_issuer_name}} -- a bare string does not parse

### spec.httpTimeoutMs

`int32` · optional (explicit presence)

Default timeout in MILLISECONDS for scalers that reach external
services over raw HTTP. Chart default: 3000.

- default: `3000`
- rule: {"int32":{"gte":1}}

### spec.priorityClassName

`string`

PriorityClass for all KEDA components. The autoscaling engine should
outlive workload evictions — pods that scale on KEDA stop scaling
without it.

### spec.nodeSelector

`map<string, string>`

Node selector applied to all KEDA components.

### spec.tolerations

`[]WorkloadToleration`

Tolerations applied to all KEDA components.

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

### spec.prometheus

`KubernetesKedaPrometheus`

KEDA's own Prometheus telemetry (operator + metrics-server scrape
endpoints and optional ServiceMonitors). This is telemetry ABOUT KEDA —
unrelated to the external metrics it serves to HPAs.

- rule: service_monitor requires prometheus metrics to be enabled — the ServiceMonitor would have no metrics endpoint to scrape

### spec.prometheus.enabled

`bool`

Expose the operator's and metrics API server's own /metrics endpoints
(scaler loop latencies, trigger errors, HPA interactions).

### spec.prometheus.serviceMonitor

`bool`

Create ServiceMonitors for scrape discovery. Requires the Prometheus
operator CRDs on the cluster — the release FAILS to install without
them.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields
(network policies, OpenTelemetry, profiling, per-component scheduling
overrides, image overrides, ...) — never the substitute for them. Do
not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKeda, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace KEDA was installed into (the resolved spec.namespace). ScaledObjects live NEXT TO their workloads, not here — this is where the engine runs. |
| `status.outputs.release_name` | `string` | Helm release name — fixed "keda" (one installation per cluster; the external.metrics.k8s.io APIService is a cluster singleton). |
| `status.outputs.operator_service_account_name` | `string` | Name of the operator's Kubernetes service account (the chart's fixed "keda-operator") — the subject cloud-side keyless bindings (IRSA role trust policies, GCP WI bindings, Entra federated credentials) are written against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.certificates.certManagerIssuer.issuer.name` | KubernetesIssuer | `status.outputs.issuer_name` |
| `spec.certificates.certManagerIssuer.clusterIssuer.name` | KubernetesClusterIssuer | `status.outputs.cluster_issuer_name` |

## See Also

- [Overview](../README.md)

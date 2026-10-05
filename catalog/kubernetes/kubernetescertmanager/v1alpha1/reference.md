# KubernetesCertManager

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesCertManagerSpec** installs cert-manager — the cluster's
certificate machinery — from the official Helm chart (`cert-manager` at
https://charts.jetstack.io). cert-manager runs three components: the
controller (watches Certificates and drives issuance), the webhook
(validates cert-manager resources at admission), and the cainjector
(injects CA bundles into webhook/CRD configurations).

This component installs and configures the CONTROLLER MACHINERY only. Who
signs certificates and what certificates exist are separate first-class
resources: create KubernetesClusterIssuer / KubernetesIssuer for the
signing authorities and KubernetesCertificate for the certificates. One
cert-manager installation per cluster serves all of them.

The typed fields below cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for chart values beyond them
(merged last, Helm `-f` semantics, identical on both engines) — a safety
valve, never the primary interface.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises every typed field plus the helm_values escape
# hatch; both engines must render an identical release from it.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCertManager
metadata:
  name: test-cert-manager
spec:
  namespace:
    value: cert-manager
  createNamespace: true
  chartVersion: v1.20.3
  crds:
    install: true
    keepOnUninstall: false
  replicas: 2
  resources:
    requests:
      cpu: 50m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  logLevel: 4
  clusterResourceNamespace: cert-manager-secrets
  leaderElectionNamespace: cert-manager
  enableCertificateOwnerRef: true
  featureGates:
    ServerSideApply: true
  dns01SelfCheck:
    recursiveNameservers:
      - "1.1.1.1:53"
      - "8.8.8.8:53"
    recursiveNameserversOnly: true
  maxConcurrentChallenges: 30
  workloadIdentity:
    eks:
      roleArn:
        value: arn:aws:iam::123456789012:role/cert-manager-dns01
  imageRegistry: mirror.example.com
  prometheus:
    enabled: true
    serviceMonitor: false
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: platform
      effect: NoSchedule
  podDisruptionBudget: true
  webhook:
    replicas: 2
    timeoutSeconds: 20
    hostNetwork: false
    securePort: 10260
    resources:
      requests:
        cpu: 20m
        memory: 64Mi
  cainjector:
    enabled: true
    replicas: 2
    resources:
      requests:
        cpu: 20m
        memory: 64Mi
  startupapicheck:
    enabled: true
    timeout: 2m
  helmValues: |
    extraArgs:
      - --enable-certificate-owner-ref=true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `v1.20.3` |  |
| `spec.crds` | `KubernetesCertManagerCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.logLevel` | `int32` |  | `2` |  |
| `spec.clusterResourceNamespace` | `string` |  |  |  |
| `spec.leaderElectionNamespace` | `string` |  | `kube-system` |  |
| `spec.enableCertificateOwnerRef` | `bool` |  |  |  |
| `spec.featureGates` | `map<string, bool>` |  |  |  |
| `spec.dns01SelfCheck` | `KubernetesCertManagerDns01SelfCheck` |  |  |  |
| `spec.dns01SelfCheck.recursiveNameservers` | `[]string` |  |  |  |
| `spec.dns01SelfCheck.recursiveNameserversOnly` | `bool` |  |  |  |
| `spec.maxConcurrentChallenges` | `int32` |  | `60` |  |
| `spec.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.workloadIdentity.aks.tenantId` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.prometheus` | `KubernetesCertManagerPrometheus` |  |  |  |
| `spec.prometheus.enabled` | `bool` |  | `true` |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitorInterval` | `string` |  | `60s` |  |
| `spec.prometheus.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.podDisruptionBudget` | `bool` |  |  |  |
| `spec.webhook` | `KubernetesCertManagerWebhook` |  |  |  |
| `spec.webhook.replicas` | `int32` |  | `1` |  |
| `spec.webhook.timeoutSeconds` | `int32` |  | `30` |  |
| `spec.webhook.hostNetwork` | `bool` |  |  |  |
| `spec.webhook.securePort` | `int32` |  | `10250` |  |
| `spec.webhook.resources` | `ContainerResources` |  |  |  |
| `spec.webhook.resources.limits` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.limits.cpu` | `string` |  |  |  |
| `spec.webhook.resources.limits.memory` | `string` |  |  |  |
| `spec.webhook.resources.requests` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.requests.cpu` | `string` |  |  |  |
| `spec.webhook.resources.requests.memory` | `string` |  |  |  |
| `spec.cainjector` | `KubernetesCertManagerCainjector` |  |  |  |
| `spec.cainjector.enabled` | `bool` |  | `true` |  |
| `spec.cainjector.replicas` | `int32` |  | `1` |  |
| `spec.cainjector.resources` | `ContainerResources` |  |  |  |
| `spec.cainjector.resources.limits` | `CpuMemory` |  |  |  |
| `spec.cainjector.resources.limits.cpu` | `string` |  |  |  |
| `spec.cainjector.resources.limits.memory` | `string` |  |  |  |
| `spec.cainjector.resources.requests` | `CpuMemory` |  |  |  |
| `spec.cainjector.resources.requests.cpu` | `string` |  |  |  |
| `spec.cainjector.resources.requests.memory` | `string` |  |  |  |
| `spec.startupapicheck` | `KubernetesCertManagerStartupApiCheck` |  |  |  |
| `spec.startupapicheck.enabled` | `bool` |  | `true` |  |
| `spec.startupapicheck.timeout` | `string` |  | `1m` |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install cert-manager into ("cert-manager" by convention).
Accepts a literal namespace name or a reference to a KubernetesNamespace
resource.

Treat the namespace as PERMANENT while CRDs are kept (the
`crds.keep_on_uninstall` default): kept CRDs retain the Helm release's
namespace in their ownership metadata, so re-installing into a
DIFFERENT namespace fails with Helm's release-ownership error on the
surviving CRDs. Moving an install requires first deleting the kept
CRDs — which cascades to ALL certificate data cluster-wide.

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

Helm chart version to install (also the cert-manager version — chart and
app versions are aligned upstream, e.g. "v1.20.3"). Pin deliberately;
upgrades re-run the release with the new chart. Pick versions from the
chart repository's index (`helm search repo`): the served chart is the
contract — the upstream source tree's Chart.yaml can claim a version at
a tag that was never served.

- default: `v1.20.3`

### spec.crds

`KubernetesCertManagerCrds`

CRD lifecycle. cert-manager's CRDs (Certificate, Issuer, ClusterIssuer,
...) are cluster-scoped and shared: installing them with the release is
the standard single-installation path.

### spec.crds.install

`bool` · optional (explicit presence)

Install the cert-manager CRDs with the release. Default TRUE (Planton
opinion — the chart itself defaults to false and expects a separate
kubectl apply; one component managing both halves is strictly simpler).
Disable only when another installation already owns the CRDs.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every Certificate/Issuer object in the
cluster) when the release is uninstalled. Default TRUE, matching
upstream — deleting CRDs cascades to ALL certificate data cluster-wide,
a destructive act that should require an explicit false.

- default: `true`

### spec.replicas

`int32` · optional (explicit presence)

Controller replica count. One replica is standard (leader election makes
extras hot standbys, not throughput).

- default: `1`
- rule: {"int32":{"gte":1}}

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

### spec.logLevel

`int32` · optional (explicit presence)

Log verbosity, 0 (errors only) to 6 (trace). Chart default: 2.

- default: `2`
- rule: {"int32":{"lte":6,"gte":0}}

### spec.clusterResourceNamespace

`string`

Namespace cert-manager reads Secrets from for CLUSTER-scoped resources
(ClusterIssuer credentials, ACME account keys) — the "cluster resource
namespace". Empty = the installation namespace. KubernetesClusterIssuer
resources materialize their credential Secrets here.

### spec.leaderElectionNamespace

`string` · optional (explicit presence)

Namespace used for leader election leases. Chart default: "kube-system"
— on clusters where kube-system is locked down (some managed platforms),
set the installation namespace instead.

- default: `kube-system`

### spec.enableCertificateOwnerRef

`bool`

When true, issued-certificate Secrets carry an ownerReference to their
Certificate, so deleting the Certificate garbage-collects the Secret.
Off upstream by default (deleting a Certificate keeps its Secret —
safer for accidental deletes).

### spec.featureGates

`map<string, bool>`

Feature gates for the controller (e.g. {"ServerSideApply": true}).
Rendered as the chart's comma-separated featureGates string.

### spec.dns01SelfCheck

`KubernetesCertManagerDns01SelfCheck`

DNS-01 self-check resolution. Before asking the CA to verify a DNS-01
challenge, cert-manager checks the TXT record itself — on clusters whose
in-cluster DNS serves a private view (split-horizon), that self-check
sees stale/private answers and issuance hangs. Point it at recursive
resolvers that see the PUBLIC view.

- rule: recursive_nameservers_only requires recursive_nameservers to be set — 'only' restricts lookups to the configured resolvers, so there must be some

### spec.dns01SelfCheck.recursiveNameservers

`[]string`

Recursive resolvers for the self-check, each "host:port"
(e.g. "8.8.8.8:53", "1.1.1.1:53").

- rule: {"repeated":{"items":{"cel":[{"id":"dns01.nameserver_format","message":"Each nameserver must be host:port, e.g. '8.8.8.8:53'","expression":"this.matches('^.+:[0-9]+$')"}]}}}

### spec.dns01SelfCheck.recursiveNameserversOnly

`bool`

When true, use the configured resolvers for ALL DNS-01 lookups, not just
the initial authoritative-nameserver discovery — the full split-horizon
fix.

### spec.maxConcurrentChallenges

`int32` · optional (explicit presence)

Maximum ACME challenges processed concurrently. Chart default: 60.

- default: `60`
- rule: {"int32":{"gte":1}}

### spec.workloadIdentity

`KubernetesWorkloadIdentity`

Binds the controller ServiceAccount to a cloud identity for KEYLESS
DNS-01 (Route53 via EKS IRSA, Cloud DNS via GKE Workload Identity,
Azure DNS via AKS Workload Identity). Issuers whose DNS-01 providers
leave static credentials empty authenticate through this identity.
Not needed for token-based providers (Cloudflare, DigitalOcean, ...).

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

### spec.imageRegistry

`string`

Registry serving all cert-manager images (controller, webhook,
cainjector, acmesolver, startupapicheck) — the air-gapped/mirror knob.
Empty = quay.io.

### spec.prometheus

`KubernetesCertManagerPrometheus`

Prometheus metrics exposure. Enabled by default upstream (metrics port
open); the ServiceMonitor is opt-in and requires the Prometheus operator
CRDs on the cluster.

### spec.prometheus.enabled

`bool` · optional (explicit presence)

Expose the metrics port on all components. Chart default: true.

- default: `true`

### spec.prometheus.serviceMonitor

`bool`

Create ServiceMonitor resources for scrape discovery. Requires the
Prometheus operator CRDs (e.g. kube-prometheus-stack) on the cluster —
the release FAILS to install without them.

### spec.prometheus.serviceMonitorInterval

`string` · optional (explicit presence)

Scrape interval for the ServiceMonitor (e.g. "60s"). Chart default: 60s.

- default: `60s`

### spec.prometheus.serviceMonitorLabels

`map<string, string>`

Extra labels on the ServiceMonitor — how a Prometheus instance's
serviceMonitorSelector finds it (e.g. {"release": "kube-prometheus-stack"}).

### spec.nodeSelector

`map<string, string>`

Node selector for all cert-manager pods. Chart default:
{"kubernetes.io/os": "linux"}.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods (webhook/cainjector inherit their
own chart defaults; set component-level tolerations via helm_values when
they must differ).

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

### spec.podDisruptionBudget

`bool`

When true, a PodDisruptionBudget guards each cert-manager component
(minAvailable 1) — meaningful only with replicas > 1.

### spec.webhook

`KubernetesCertManagerWebhook`

Webhook component tuning. The webhook must be reachable by the
API server — host_network is the standard fix on clusters whose control
plane cannot reach pod IPs (e.g. EKS with custom CNI).

### spec.webhook.replicas

`int32` · optional (explicit presence)

Webhook replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.webhook.timeoutSeconds

`int32` · optional (explicit presence)

Admission timeout in seconds the API server waits on the webhook.
Chart default: 30 (the API server maximum).

- default: `30`
- rule: {"int32":{"lte":30,"gte":1}}

### spec.webhook.hostNetwork

`bool`

Run the webhook on the host network — required where the control plane
cannot reach pod IPs (EKS with custom/alternative CNI is the canonical
case). Pair with a secure_port that is free on the node.

### spec.webhook.securePort

`int32` · optional (explicit presence)

Port the webhook serves on. Chart default: 10250. With host_network,
pick a port unused on nodes (10250 collides with the kubelet).

- default: `10250`
- rule: {"int32":{"lte":65535,"gte":1}}

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

### spec.cainjector

`KubernetesCertManagerCainjector`

cainjector component tuning. Required by the webhook's own certificate
bootstrap — disable only when another mechanism manages webhook CA
bundles.

### spec.cainjector.enabled

`bool` · optional (explicit presence)

Run the cainjector. Chart default: true; cert-manager's own webhook
depends on it.

- default: `true`

### spec.cainjector.replicas

`int32` · optional (explicit presence)

cainjector replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.cainjector.resources

`ContainerResources`

cainjector container CPU/memory requests and limits.

### spec.cainjector.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.cainjector.resources.limits.cpu

`string`

### spec.cainjector.resources.limits.memory

`string`

### spec.cainjector.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.cainjector.resources.requests.cpu

`string`

### spec.cainjector.resources.requests.memory

`string`

### spec.startupapicheck

`KubernetesCertManagerStartupApiCheck`

startupapicheck: a post-install hook Job that verifies the webhook is
actually serving before the release reports success. Costs ~seconds;
disable on clusters that forbid hook Jobs.

### spec.startupapicheck.enabled

`bool` · optional (explicit presence)

Run the post-install verification Job. Chart default: true.

- default: `true`

### spec.startupapicheck.timeout

`string` · optional (explicit presence)

How long the check Job retries before failing the install (Go duration).
Chart default: "1m".

- default: `1m`

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields —
never the substitute for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesCertManager, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Kubernetes namespace cert-manager was installed into. |
| `status.outputs.release_name` | `string` | Helm release name. |
| `status.outputs.service_account_name` | `string` | Name of the cert-manager controller ServiceAccount — the identity to bind on the cloud side for keyless DNS-01 (IRSA trust policy subject, GKE Workload Identity member, Azure federated credential subject). |
| `status.outputs.cluster_resource_namespace` | `string` | Namespace cert-manager reads Secrets from for cluster-scoped resources (the resolved cluster-resource namespace: spec.cluster_resource_namespace when set, otherwise the installation namespace). KubernetesClusterIssuer resources materialize their credential Secrets here. |

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
| KubernetesClusterIssuer | `spec.certManagerNamespace` | `status.outputs.cluster_resource_namespace` |

## See Also

- [Overview](../README.md)

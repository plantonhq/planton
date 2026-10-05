# KubernetesClusterAutoscaler

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesClusterAutoscalerSpec** installs the Kubernetes Cluster
Autoscaler from the official Helm chart (`cluster-autoscaler` at
https://kubernetes.github.io/autoscaler). The autoscaler grows and
shrinks EXISTING node groups: when pods are unschedulable it raises the
desired size of a matching group (an EC2 Auto Scaling group, an Azure
VMSS, a Cluster API MachineDeployment, ...), and it scales groups back
down when nodes sit underutilized.

WHERE this component earns its keep: clusters whose node capacity is
organized as pre-defined groups — EKS with ASGs, Cluster API /
self-managed clusters, and providers without a managed autoscaler. Note
that GKE and AKS ship a MANAGED autoscaler configured as a toggle on the
node pool itself (on Planton, through the cluster kinds) — deploying
this component there is the exception, not the rule. For AWS clusters
that want right-sized machines launched on demand instead of pre-defined
groups, KubernetesKarpenter is the modern alternative.

ONE INSTALLATION PER CLUSTER: the autoscaler leader-elects and owns the
cluster-wide scaling decision (a second installation would fight the
first over every scale-up). The Helm release name is therefore fixed to
"cluster-autoscaler".

The typed fields below cover the chart's meaningful configuration
surface; `extra_args` carries any of the autoscaler's 100+ tuning flags
beyond the typed set (the chart's own extraArgs contract), and
`helm_values` remains as the escape hatch for chart values (merged last,
Helm `-f` semantics, identical on both engines).

## Example

```yaml
# Full-surface test manifest: exercises the recommended AWS posture
# (tag-based auto-discovery + IRSA) plus every cross-arm block (scaling,
# extra_args, deployment, prometheus, helm_values) so the offline plan
# proofs cover what the live lanes may not. Not a realistic production
# shape — kube-system with create_namespace false is the real-world norm.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterAutoscaler
metadata:
  name: hack-cluster-autoscaler
spec:
  namespace:
    value: hack-autoscaler
  createNamespace: true
  chartVersion: "9.59.0"
  aws:
    region: us-west-2
    autoDiscovery:
      clusterName: hack-eks-cluster
      tags:
        - k8s.io/cluster-autoscaler/enabled
        - k8s.io/cluster-autoscaler/hack-eks-cluster
    irsaRoleArn: arn:aws:iam::123456789012:role/hack-cluster-autoscaler
  scaling:
    expander: priority,least-waste
    balanceSimilarNodeGroups: true
    scanInterval: 30s
    maxNodeProvisionTime: 15m
    skipNodesWithLocalStorage: false
    skipNodesWithSystemPods: true
    scaleDown:
      enabled: true
      utilizationThreshold: "0.5"
      unneededTime: 10m
      delayAfterAdd: 10m
      delayAfterDelete: 0s
      delayAfterFailure: 3m
  extraArgs:
    max-graceful-termination-sec: "600"
    v: "2"
  deployment:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 300Mi
      limits:
        cpu: "1"
        memory: 600Mi
    priorityClassName: system-cluster-critical
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
  prometheus:
    serviceMonitor: true
    serviceMonitorSelectorRelease: kube-prometheus-stack
  helmValues: |
    podAnnotations:
      hack.planton.ai/full-surface: "true"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `9.59.0` |  |
| `spec.aws` | `KubernetesClusterAutoscalerAws` |  |  |  |
| `spec.aws.region` | `string` | yes |  |  |
| `spec.aws.autoDiscovery` | `KubernetesClusterAutoscalerAwsAutoDiscovery` |  |  |  |
| `spec.aws.autoDiscovery.clusterName` | `string` | yes |  |  |
| `spec.aws.autoDiscovery.tags` | `[]string` |  |  |  |
| `spec.aws.nodeGroups` | `[]KubernetesClusterAutoscalerNodeGroup` |  |  |  |
| `spec.aws.nodeGroups[].name` | `string` | yes |  |  |
| `spec.aws.nodeGroups[].minSize` | `int32` |  |  |  |
| `spec.aws.nodeGroups[].maxSize` | `int32` |  |  |  |
| `spec.aws.irsaRoleArn` | `string` |  |  |  |
| `spec.aws.accessKeys` | `KubernetesClusterAutoscalerAwsAccessKeys` |  |  |  |
| `spec.aws.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.aws.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.azure` | `KubernetesClusterAutoscalerAzure` |  |  |  |
| `spec.azure.subscriptionId` | `string` | yes |  |  |
| `spec.azure.resourceGroup` | `string` | yes |  |  |
| `spec.azure.clusterName` | `string` |  |  |  |
| `spec.azure.nodeGroups` | `[]KubernetesClusterAutoscalerNodeGroup` |  |  |  |
| `spec.azure.nodeGroups[].name` | `string` | yes |  |  |
| `spec.azure.nodeGroups[].minSize` | `int32` |  |  |  |
| `spec.azure.nodeGroups[].maxSize` | `int32` |  |  |  |
| `spec.azure.identity` | `KubernetesClusterAutoscalerAzureIdentity` | yes |  |  |
| `spec.azure.identity.useWorkloadIdentity` | `bool` |  |  |  |
| `spec.azure.identity.useManagedIdentity` | `bool` |  |  |  |
| `spec.azure.identity.userAssignedIdentityId` | `string` |  |  |  |
| `spec.azure.identity.servicePrincipal` | `KubernetesClusterAutoscalerAzureServicePrincipal` |  |  |  |
| `spec.azure.identity.servicePrincipal.tenantId` | `string` | yes |  |  |
| `spec.azure.identity.servicePrincipal.clientId` | `string` | yes |  |  |
| `spec.azure.identity.servicePrincipal.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.gce` | `KubernetesClusterAutoscalerGce` |  |  |  |
| `spec.gce.instanceGroupPrefixes` | `[]KubernetesClusterAutoscalerNodeGroup` | yes |  |  |
| `spec.gce.instanceGroupPrefixes[].name` | `string` | yes |  |  |
| `spec.gce.instanceGroupPrefixes[].minSize` | `int32` |  |  |  |
| `spec.gce.instanceGroupPrefixes[].maxSize` | `int32` |  |  |  |
| `spec.gce.workloadIdentityServiceAccountEmail` | `string` |  |  |  |
| `spec.clusterApi` | `KubernetesClusterAutoscalerClusterApi` |  |  |  |
| `spec.clusterApi.mode` | `string` |  | `incluster-incluster` |  |
| `spec.clusterApi.kubeconfigSecret` | `string` |  |  |  |
| `spec.clusterApi.namespace` | `string` |  |  |  |
| `spec.clusterApi.namespaceScopedRbac` | `bool` |  |  |  |
| `spec.civo` | `KubernetesClusterAutoscalerCivo` |  |  |  |
| `spec.civo.clusterId` | `string` | yes |  |  |
| `spec.civo.region` | `string` | yes |  |  |
| `spec.civo.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.civo.apiUrl` | `string` |  | `https://api.civo.com` |  |
| `spec.kwok` | `KubernetesClusterAutoscalerKwok` |  |  |  |
| `spec.kwok.configMapName` | `string` |  | `kwok-provider-config` |  |
| `spec.scaling` | `KubernetesClusterAutoscalerScaling` |  |  |  |
| `spec.scaling.expander` | `string` |  |  |  |
| `spec.scaling.balanceSimilarNodeGroups` | `bool` |  |  |  |
| `spec.scaling.scaleDown` | `KubernetesClusterAutoscalerScaleDown` |  |  |  |
| `spec.scaling.scaleDown.enabled` | `bool` |  |  |  |
| `spec.scaling.scaleDown.utilizationThreshold` | `string` |  |  |  |
| `spec.scaling.scaleDown.unneededTime` | `string` |  |  |  |
| `spec.scaling.scaleDown.delayAfterAdd` | `string` |  |  |  |
| `spec.scaling.scaleDown.delayAfterDelete` | `string` |  |  |  |
| `spec.scaling.scaleDown.delayAfterFailure` | `string` |  |  |  |
| `spec.scaling.scanInterval` | `string` |  |  |  |
| `spec.scaling.maxNodeProvisionTime` | `string` |  |  |  |
| `spec.scaling.skipNodesWithLocalStorage` | `bool` |  |  |  |
| `spec.scaling.skipNodesWithSystemPods` | `bool` |  |  |  |
| `spec.extraArgs` | `map<string, string>` |  |  |  |
| `spec.deployment` | `KubernetesClusterAutoscalerDeployment` |  |  |  |
| `spec.deployment.replicas` | `int32` |  | `1` |  |
| `spec.deployment.resources` | `ContainerResources` |  |  |  |
| `spec.deployment.resources.limits` | `CpuMemory` |  |  |  |
| `spec.deployment.resources.limits.cpu` | `string` |  |  |  |
| `spec.deployment.resources.limits.memory` | `string` |  |  |  |
| `spec.deployment.resources.requests` | `CpuMemory` |  |  |  |
| `spec.deployment.resources.requests.cpu` | `string` |  |  |  |
| `spec.deployment.resources.requests.memory` | `string` |  |  |  |
| `spec.deployment.priorityClassName` | `string` |  | `system-cluster-critical` |  |
| `spec.deployment.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.deployment.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.deployment.tolerations[].key` | `string` |  |  |  |
| `spec.deployment.tolerations[].operator` | `string` |  |  |  |
| `spec.deployment.tolerations[].value` | `string` |  |  |  |
| `spec.deployment.tolerations[].effect` | `string` |  |  |  |
| `spec.deployment.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.prometheus` | `KubernetesClusterAutoscalerPrometheus` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitorSelectorRelease` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the autoscaler into ("kube-system" is the
upstream convention — it keeps the pod under the system-critical
eviction umbrella). Accepts a literal namespace name or a reference
to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist. Leave false when
installing into kube-system.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "9.59.0", which ships autoscaler
1.35 — chart and app versions move separately; the chart pin
governs). Pin deliberately; upgrades re-run the release with the new
chart. Keep the autoscaler's MINOR version aligned with the cluster's
Kubernetes minor per upstream guidance (override the image tag via
helm_values when the cluster runs an older minor). Pick versions from
the chart repository's index (`helm search repo`): the served chart is
the contract — the upstream source tree's Chart.yaml can claim a
version at a tag that was never served.

- default: `9.59.0`

### spec.aws

`KubernetesClusterAutoscalerAws`

AWS: EC2 Auto Scaling groups, discovered by tag or listed
statically.

- rule: configure exactly one of auto_discovery (tag-based, recommended) or node_groups (static list)
- rule: irsa_role_arn and access_keys are alternative credential postures — set at most one

### spec.aws.region

`string` · required

AWS region of the Auto Scaling groups (required by the chart for the
aws provider).

- rule: {"required":true}

### spec.aws.autoDiscovery

`KubernetesClusterAutoscalerAwsAutoDiscovery`

Tag-based ASG auto-discovery: the autoscaler manages every ASG
carrying the standard tags for this cluster name
(k8s.io/cluster-autoscaler/enabled +
k8s.io/cluster-autoscaler/<cluster_name>). The recommended mode —
new node groups enroll by tagging alone. Exactly one of
auto_discovery / node_groups.

### spec.aws.autoDiscovery.clusterName

`string` · required

Cluster name the discovery tags carry
(k8s.io/cluster-autoscaler/<cluster_name>).

- rule: {"required":true}

### spec.aws.autoDiscovery.tags

`[]string`

Override the ASG tags to match (chart default:
k8s.io/cluster-autoscaler/enabled +
k8s.io/cluster-autoscaler/<cluster_name>). Set only for
non-standard tagging schemes.

### spec.aws.nodeGroups

`[]KubernetesClusterAutoscalerNodeGroup`

Statically listed ASGs with explicit size bounds. Exactly one of
auto_discovery / node_groups.

- rule: min_size cannot exceed max_size

### spec.aws.nodeGroups[].name

`string` · required

Group name (or name prefix for GCE).

- rule: {"required":true}

### spec.aws.nodeGroups[].minSize

`int32`

Minimum size the autoscaler may shrink the group to.

- rule: {"int32":{"gte":0}}

### spec.aws.nodeGroups[].maxSize

`int32`

Maximum size the autoscaler may grow the group to.

- rule: {"int32":{"gte":1}}

### spec.aws.irsaRoleArn

`string`

IAM role ARN for IRSA: annotates the autoscaler's service account so
it calls the Auto Scaling APIs without stored keys (the role's trust
policy must allow the cluster's OIDC provider and the autoscaler's
service account). The keyless posture — preferred over access_keys.

- rule: irsa_role_arn must be an IAM role ARN (arn:aws:iam::<account>:role/<name>)

### spec.aws.accessKeys

`KubernetesClusterAutoscalerAwsAccessKeys`

Static AWS access keys — only for clusters without IRSA (self-managed
on EC2). The secret key is stored as a managed secret and materialized
into the chart's credential Secret.

### spec.aws.accessKeys.accessKeyId

`string` · required

AWS access key ID — the public identifier of the key pair, not a
secret (the guard's name heuristic treats *_id names as references,
so no exemption annotation is needed); only the paired secret access
key is a credential.

- rule: {"required":true}

### spec.aws.accessKeys.secretAccessKey

`string` · required · sensitive

AWS secret access key.

- rule: {"required":true}

### spec.azure

`KubernetesClusterAutoscalerAzure`

Azure: VM scale sets, discovered by tag or listed statically —
for AKS clusters that opt out of the managed autoscaler and for
self-managed clusters on VMSS.

- rule: configure exactly one of cluster_name (tag-based auto-discovery) or node_groups (static list)

### spec.azure.subscriptionId

`string` · required

Azure subscription the scale sets live in.

- rule: {"required":true}

### spec.azure.resourceGroup

`string` · required

Resource group of the cluster/scale sets. For AKS this is the NODE
resource group (the MC_* group holding the VMSS instances).

- rule: {"required":true}

### spec.azure.clusterName

`string`

Tag-based VMSS auto-discovery for this cluster name (scale sets
tagged per upstream's Azure auto-discovery setup). Exactly one of
cluster_name / node_groups.

### spec.azure.nodeGroups

`[]KubernetesClusterAutoscalerNodeGroup`

Statically listed scale sets with explicit size bounds. Exactly one
of cluster_name / node_groups.

- rule: min_size cannot exceed max_size

### spec.azure.nodeGroups[].name

`string` · required

Group name (or name prefix for GCE).

- rule: {"required":true}

### spec.azure.nodeGroups[].minSize

`int32`

Minimum size the autoscaler may shrink the group to.

- rule: {"int32":{"gte":0}}

### spec.azure.nodeGroups[].maxSize

`int32`

Maximum size the autoscaler may grow the group to.

- rule: {"int32":{"gte":1}}

### spec.azure.identity

`KubernetesClusterAutoscalerAzureIdentity` · required

How the autoscaler authenticates to Azure. Exactly one arm.

- rule: {"required":true}
- rule: configure exactly one Azure credential posture: use_workload_identity, use_managed_identity, or service_principal
- rule: user_assigned_identity_id only applies with use_managed_identity

### spec.azure.identity.useWorkloadIdentity

`bool`

Federated workload identity (recommended): the autoscaler's service
account exchanges its token for the Entra application below. Set
client_id in service_principal? No — set use_workload_identity with
the pod labels handled by the chart; the federated app is configured
Azure-side.

### spec.azure.identity.useManagedIdentity

`bool`

VM managed identity: the autoscaler uses the node's (or the
specified user-assigned) managed identity.

### spec.azure.identity.userAssignedIdentityId

`string`

Which user-assigned identity to use when the VMSS carries several
(only with use_managed_identity).

### spec.azure.identity.servicePrincipal

`KubernetesClusterAutoscalerAzureServicePrincipal`

Service-principal credentials (client id + secret with contributor
on the cluster and node resource groups) — the declared-credential
fallback.

### spec.azure.identity.servicePrincipal.tenantId

`string` · required

Entra tenant ID.

- rule: {"required":true}

### spec.azure.identity.servicePrincipal.clientId

`string` · required

Application (client) ID of the service principal (a public
identifier — not a secret).

- rule: {"required":true}

### spec.azure.identity.servicePrincipal.clientSecret

`string` · required · sensitive

Client secret of the service principal.

- rule: {"required":true}

### spec.gce

`KubernetesClusterAutoscalerGce`

GCE: managed instance groups by name prefix — for self-managed
clusters on GCE (GKE's managed autoscaler is a node-pool toggle,
not this component).

### spec.gce.instanceGroupPrefixes

`[]KubernetesClusterAutoscalerNodeGroup` · required

Managed instance groups by NAME PREFIX with size bounds (the chart's
GCE contract — no tagging required).

- rule: {"repeated":{"minItems":"1"}}
- rule: min_size cannot exceed max_size

### spec.gce.instanceGroupPrefixes[].name

`string` · required

Group name (or name prefix for GCE).

- rule: {"required":true}

### spec.gce.instanceGroupPrefixes[].minSize

`int32`

Minimum size the autoscaler may shrink the group to.

- rule: {"int32":{"gte":0}}

### spec.gce.instanceGroupPrefixes[].maxSize

`int32`

Maximum size the autoscaler may grow the group to.

- rule: {"int32":{"gte":1}}

### spec.gce.workloadIdentityServiceAccountEmail

`string`

GCP service-account email for Workload Identity: annotates the
autoscaler's Kubernetes service account so it calls the compute APIs
keylessly (the GSA needs compute.instanceGroups permissions and a WI
binding to the autoscaler's KSA). Empty = node default credentials.

- rule: workload_identity_service_account_email must be a GCP service-account email (…@…gserviceaccount.com)

### spec.clusterApi

`KubernetesClusterAutoscalerClusterApi`

Cluster API: MachineDeployments/MachineSets annotated for
autoscaling — the self-managed / multi-cloud arm.

- rule: the chosen mode reads a kubeconfig — set kubeconfig_secret (and mount it per upstream guidance)

### spec.clusterApi.mode

`string` · optional (explicit presence)

Where the autoscaler finds the workload cluster and the Cluster API
management objects: "incluster-incluster" (chart default — both in
this cluster), "incluster-kubeconfig", "kubeconfig-incluster",
"kubeconfig-kubeconfig" or "single-kubeconfig". Kubeconfig modes
additionally need the kubeconfig mounted via helm_values
(extraVolumeSecrets) per upstream's Cluster API guidance.

- default: `incluster-incluster`
- rule: mode must be one of 'incluster-incluster', 'incluster-kubeconfig', 'kubeconfig-incluster', 'kubeconfig-kubeconfig' or 'single-kubeconfig'

### spec.clusterApi.kubeconfigSecret

`string`

NAME of the Kubernetes Secret holding the kubeconfig for the
CAPI-managed workload cluster (required for the kubeconfig-* and
*-kubeconfig modes). The kubeconfig itself lives in that Secret on
the cluster — never here.

### spec.clusterApi.namespace

`string`

Namespace to watch for CAPI machine objects (autoDiscovery.namespace).
Empty watches all namespaces.

### spec.clusterApi.namespaceScopedRbac

`bool`

Restrict RBAC to namespace scope instead of cluster scope
(rbac.clusterScoped=false) — the least-privilege posture when the
autoscaler only manages machines in one namespace.

### spec.civo

`KubernetesClusterAutoscalerCivo`

Civo: managed Kubernetes node pools on Civo.

### spec.civo.clusterId

`string` · required

Civo cluster ID the node pools belong to.

- rule: {"required":true}

### spec.civo.region

`string` · required

Civo region of the cluster.

- rule: {"required":true}

### spec.civo.apiKey

`string` · required · sensitive

Civo API key used to resize node pools.

- rule: {"required":true}

### spec.civo.apiUrl

`string` · optional (explicit presence)

Civo API URL. Chart default: https://api.civo.com.

- default: `https://api.civo.com`

### spec.kwok

`KubernetesClusterAutoscalerKwok`

KWOK simulation provider — nodes are FAKE (created by the KWOK
controller, which must be installed on the cluster). For testing
scaling policies and evaluating the autoscaler without a cloud
account; never a production arm.

### spec.kwok.configMapName

`string` · optional (explicit presence)

Name of the ConfigMap carrying the KWOK provider configuration
(node templates the simulator "launches"). Chart default:
"kwok-provider-config".

- default: `kwok-provider-config`

### spec.scaling

`KubernetesClusterAutoscalerScaling`

Scaling behavior: expander choice, scale-down tuning, node-group
balancing — the flags every real installation ends up tuning,
rendered into the chart's extraArgs.

### spec.scaling.expander

`string`

Node-group selection strategy when several groups could satisfy a
scale-up: comma-separated ordered expanders from "random",
"most-pods", "least-waste", "price", "priority", "grpc" (upstream
default: random; "least-waste" is the common production choice).
The "priority" expander additionally reads the
cluster-autoscaler-priority-expander ConfigMap (ship it via
helm_values expanderPriorities).

- rule: expander must be a comma-separated list drawn from 'random', 'most-pods', 'least-waste', 'price', 'priority', 'grpc'

### spec.scaling.balanceSimilarNodeGroups

`bool`

Treat node groups with identical instance types and labels as one,
keeping their sizes balanced (the multi-AZ-ASG pattern on AWS, where
one group per zone is the norm).

### spec.scaling.scaleDown

`KubernetesClusterAutoscalerScaleDown`

Scale-down behavior.

### spec.scaling.scaleDown.enabled

`bool` · optional (explicit presence)

Master switch for scale-down (upstream default: true). Disabling
turns the autoscaler into scale-up-only.

### spec.scaling.scaleDown.utilizationThreshold

`string`

Utilization (requests/allocatable) below which a node is a
scale-down candidate, as a decimal fraction string (upstream default
"0.5").

- rule: utilization_threshold must be a decimal fraction between 0 and 1 (e.g. "0.5")

### spec.scaling.scaleDown.unneededTime

`string`

How long a node must be unneeded before removal (upstream default
10m).

- rule: unneeded_time must be a Go duration such as "10m"

### spec.scaling.scaleDown.delayAfterAdd

`string`

Cool-down after a scale-UP before scale-down resumes (upstream
default 10m).

- rule: delay_after_add must be a Go duration such as "10m"

### spec.scaling.scaleDown.delayAfterDelete

`string`

Cool-down after a node DELETION before scale-down resumes (upstream
default 0s).

- rule: delay_after_delete must be a Go duration such as "0s"

### spec.scaling.scaleDown.delayAfterFailure

`string`

Cool-down after a FAILED scale-down before retrying (upstream
default 3m).

- rule: delay_after_failure must be a Go duration such as "3m"

### spec.scaling.scanInterval

`string`

How often the cluster is re-evaluated (Go duration; upstream default
10s). Longer intervals reduce API load on very large clusters.

- rule: scan_interval must be a Go duration such as "10s"

### spec.scaling.maxNodeProvisionTime

`string`

How long to wait for a requested node before giving up on the group
and retrying elsewhere (upstream default 15m).

- rule: max_node_provision_time must be a Go duration such as "15m"

### spec.scaling.skipNodesWithLocalStorage

`bool` · optional (explicit presence)

Do not scale down nodes running pods with local storage (emptyDir /
hostPath). Upstream default: true — set false deliberately, knowing
such pods lose their data on consolidation.

### spec.scaling.skipNodesWithSystemPods

`bool` · optional (explicit presence)

Do not scale down nodes running kube-system pods (except
DaemonSets/mirror pods). Upstream default: true.

### spec.extraArgs

`map<string, string>`

Additional `cluster-autoscaler` flags as flag-name → value pairs
(rendered into the chart's extraArgs, without the leading `--`).
This IS the chart's own contract for the autoscaler's long tail of
flags — use the typed `scaling` block first. Flag names are
validated for shape only; unknown flags fail at pod start.

- rule: {"map":{"keys":{"string":{"pattern":"^[a-z0-9]+(-[a-z0-9]+)*$"}}}}

### spec.deployment

`KubernetesClusterAutoscalerDeployment`

Deployment sizing and platform scheduling for the autoscaler pod
itself.

### spec.deployment.replicas

`int32` · optional (explicit presence)

Replica count. Chart default: 1 — replicas leader-elect, extras are
warm standbys.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.deployment.resources

`ContainerResources`

Container resources. Empty = no requests/limits (upstream's example
starting point is 100m/300Mi).

### spec.deployment.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.deployment.resources.limits.cpu

`string`

### spec.deployment.resources.limits.memory

`string`

### spec.deployment.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.deployment.resources.requests.cpu

`string`

### spec.deployment.resources.requests.memory

`string`

### spec.deployment.priorityClassName

`string` · optional (explicit presence)

PriorityClass for the autoscaler pod. Chart default:
"system-cluster-critical" — the component that adds capacity must
not be evicted for lack of it.

- default: `system-cluster-critical`

### spec.deployment.nodeSelector

`map<string, string>`

Node selector for the autoscaler pod (e.g. a management node group —
the autoscaler should not run on a node it may delete).

### spec.deployment.tolerations

`[]WorkloadToleration`

Tolerations for the autoscaler pod.

### spec.deployment.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.deployment.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.deployment.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.deployment.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.deployment.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.prometheus

`KubernetesClusterAutoscalerPrometheus`

The autoscaler's own Prometheus telemetry.

### spec.prometheus.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery of the autoscaler's
metrics (scale decisions, unschedulable pod counts, node-group
sizes). Requires the Prometheus operator CRDs on the cluster — the
release FAILS to install without them.

### spec.prometheus.serviceMonitorSelectorRelease

`string`

Prometheus release-label selector the ServiceMonitor carries (chart
default "prometheus-operator" — set to your Prometheus install's
selector, e.g. the kube-prometheus-stack release name).

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (image overrides, PDB tuning, VPA, extra volumes,
priority-expander ConfigMap annotations, ...) — never the substitute
for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesClusterAutoscaler, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the autoscaler was installed into (the resolved spec.namespace). |
| `status.outputs.release_name` | `string` | Helm release name — fixed "cluster-autoscaler" (one installation per cluster; the leader-elected autoscaler owns the cluster-wide scaling decision). |
| `status.outputs.service_account_name` | `string` | Name of the autoscaler's Kubernetes service account — the subject cloud-side keyless bindings (IRSA trust policies, GCP WI bindings, Azure federated credentials) are written against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)

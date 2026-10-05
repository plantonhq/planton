# KubernetesRabbitMqOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesRabbitMqOperatorSpec** installs the RabbitMQ Cluster
Operator — the MPL-2.0 operator maintained by the RabbitMQ team —
which reconciles `RabbitmqCluster` custom resources into running
RabbitMQ clusters (StatefulSet, Services, generated credentials,
rolling upgrades).

This component installs and configures the ENGINE. RabbitMQ clusters
themselves are declared with KubernetesRabbitMq resources — one per
cluster — which this operator reconciles.

DISTRIBUTION AND VERSION: the operator has NO Helm chart. Its
official distribution is the single-file release manifest
(`cluster-operator.yml` attached to each GitHub release), which the
modules fetch from the release tag they are pinned to and apply per
document. The spec deliberately has NO version field: the installed
operator (and its RabbitmqCluster CRD schema) is pinned to the
release this catalog's typed SDK was generated against (see
pkg/kubernetes/kubernetestypes/Makefile
`rabbitmq_cluster_operator_release`) — a user-selectable version
would silently drift the CRD schema away from the typed resources
built on it.

NAMESPACE IS FIXED: the release manifest installs into
`rabbitmq-system`, and that name is baked into the manifest's own
cross-references (the webhook client configuration, the cert-manager
Certificate DNS names, the CA-injection annotations, the
cluster-role binding subjects). Exactly ONE install per cluster is
the upstream contract — the admission webhooks are cluster-scoped
singletons with fixed names, so a second install cannot coexist.

CERT-MANAGER IS A HARD PREREQUISITE (declared as a registry
prerequisite of this kind): the manifest ships mutating and
validating admission webhooks for RabbitmqCluster with
`failurePolicy: Fail`, and their serving certificate is a
cert-manager `Certificate` (self-signed `Issuer`) with CA-injection
annotations. Installing without a running cert-manager leaves the
webhook certificate unissued and every RabbitmqCluster admission
failing. Declare cert-manager with KubernetesCertManager first.

CRD LIFECYCLE: the RabbitmqCluster CRD is one document of the
applied manifest — it installs and is REMOVED with the resource.
Destroying the operator therefore deletes the CRD, which
cascade-deletes every RabbitmqCluster on the cluster. Never destroy
the operator while KubernetesRabbitMq resources exist.

## Example

```yaml
# Full-surface offline-proof manifest: exercises a fenced watch scope,
# both fleet-wide default-image overrides, an operator image override with
# a pull secret, a resources override, and node selector + tolerations —
# so the offline tofu plan and pulumi preview proofs cover every patched
# surface of the release manifest's operator Deployment. Placeholder
# values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMqOperator
metadata:
  name: rabbitmq-operator-hack
spec:
  watchNamespaces:
    - messaging
    - team-a
  defaultRabbitmqImage: mirror.internal/rabbitmq:4.2.6-management
  defaultUserUpdaterImage: mirror.internal/default-user-credential-updater:1.0.14
  operatorImage:
    repo: mirror.internal/rabbitmq/cluster-operator
    tag: 2.22.3
    pullSecretName: mirror-pull
  resources:
    requests:
      cpu: 200m
      memory: 500Mi
    limits:
      cpu: 500m
      memory: 1Gi
  nodeSelector:
    workload: system
  tolerations:
    - key: system
      operator: Exists
      effect: NoSchedule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.watchNamespaces` | `[]string` |  |  |  |
| `spec.defaultRabbitmqImage` | `string` |  |  |  |
| `spec.defaultUserUpdaterImage` | `string` |  |  |  |
| `spec.operatorImage` | `ContainerImage` |  |  |  |
| `spec.operatorImage.repo` | `string` |  |  |  |
| `spec.operatorImage.tag` | `string` |  |  |  |
| `spec.operatorImage.pullSecretName` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |

## Field Details

### spec.watchNamespaces

`[]string`

Namespaces the operator watches for RabbitmqCluster resources.
EMPTY (the upstream default) = the operator watches ALL
namespaces — note this is the opposite default from
chart-scoped operators like the Altinity or external-secrets
operators. Set one or more namespace names to fence the watch;
every namespace that will hold KubernetesRabbitMq resources must
then be covered (rendered as the operator's
OPERATOR_SCOPE_NAMESPACE environment variable, comma-separated).

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.defaultRabbitmqImage

`string`

Fleet-wide default RabbitMQ server image applied to every
RabbitmqCluster that does not pin its own `image` (the operator's
DEFAULT_RABBITMQ_IMAGE environment variable). Empty = the
operator's compiled-in default (`rabbitmq:4.2.6-management` at the
pinned release — the `-management` variant is required; the
operator's generated configuration expects the management plugin).
Set for air-gapped clusters that mirror the RabbitMQ image.

### spec.defaultUserUpdaterImage

`string`

Default image for the credential-updater sidecar that
Vault-backed clusters run (the operator's
DEFAULT_USER_UPDATER_IMAGE environment variable). Empty = the
operator's compiled-in default
(`ghcr.io/rabbitmq/default-user-credential-updater:1.0.14` at the
pinned release). Only consulted for KubernetesRabbitMq resources
using the Vault secret backend.

### spec.operatorImage

`ContainerImage`

Override the operator image itself (air-gap / private-mirror
path). Empty = the release manifest's pinned image
(`ghcr.io/rabbitmq/cluster-operator` at the release tag).

### spec.operatorImage.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.operatorImage.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.operatorImage.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.resources

`ContainerResources`

Operator container resources. Empty = the release manifest's
defaults (200m CPU / 500Mi memory for both requests and limits).

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

### spec.nodeSelector

`map<string, string>`

Node selector for the operator pod.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the operator pod.

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

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the rabbitmq-system namespace)
for pulling the operator image from a private mirror.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesRabbitMqOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator is installed into (always rabbitmq-system — the release manifest's fixed namespace). |
| `status.outputs.deployment_name` | `string` | name of the operator Deployment (rabbitmq-cluster-operator — the release manifest's fixed name). |
| `status.outputs.metrics_endpoint` | `string` | in-cluster metrics endpoint of the operator (Prometheus format), e.g. http://rabbitmq-cluster-operator-metrics-service.rabbitmq-system.svc.cluster.local:8080/metrics. |
| `status.outputs.crd_name` | `string` | name of the RabbitmqCluster CustomResourceDefinition the operator serves (rabbitmqclusters.rabbitmq.com). Deleted with this resource — see the CRD-lifecycle note on the spec. |

## See Also

- [Overview](../README.md)

# KubernetesKeycloakOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKeycloakOperatorSpec** installs the official Keycloak
Operator from the keycloak-k8s-resources release manifests
(https://github.com/keycloak/keycloak-k8s-resources — Keycloak
ships NO official Helm chart; the operator IS the first-party
Kubernetes distribution). The operator reconciles Keycloak
declarations (KubernetesKeycloak resources) into running Keycloak
StatefulSets.

THE BUNDLE: 16 plain-YAML documents (ServiceAccount, RBAC, Service,
the operator Deployment) plus 4 CRDs — no admission webhooks, no
cert-manager dependency, no install hooks. The module fetches the
release manifests tag-pinned, stamps the target namespace onto
every namespaced document (upstream expects kustomize to do this),
and applies them as ordered groups so creation and teardown order
correctly by construction.

EVERY RESOURCE IS FIXED-NAME (`keycloak-operator` etc. — upstream's
own names, not derived from this resource's name), so exactly ONE
operator install fits per namespace. Installing the operator alone
deploys no Keycloak server.

NO VERSION FIELD by design: the KubernetesKeycloak declaration
kind's CR rendering is built against the CRD schema this bundle
installs — a selectable operator version would drift the schema
away from what the declaration kind renders. The module pins the
release; upgrades arrive as module updates.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed field rendered at once (cluster-wide watch scope, both image
# overrides, resources, scheduling).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeycloakOperator
metadata:
  name: keycloak-operator-full
spec:
  namespace:
    value: keycloak-op-hack
  createNamespace: true
  clusterWide: true
  operatorImage: mirror.example.com/keycloak/keycloak-operator:26.7.0
  defaultKeycloakImage: mirror.example.com/keycloak/keycloak:26.7.0
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
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.clusterWide` | `bool` |  |  |  |
| `spec.operatorImage` | `string` |  |  |  |
| `spec.defaultKeycloakImage` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.scheduling` | `KubernetesKeycloakOperatorScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the operator into (conventionally
"keycloak"). Accepts a literal namespace name or a reference to
a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.clusterWide

`bool`

Watch scope. False (default): the operator watches ONLY its own
namespace — KubernetesKeycloak resources must live beside it,
and several teams can run isolated operator+Keycloak stacks in
separate namespaces. True: the operator watches ALL namespaces
(upstream's cluster-wide variant — per-controller
ClusterRoleBindings and JOSDK_ALL_NAMESPACES); run at most one
cluster-wide operator per cluster. KNOW THIS (an upstream
constraint): in cluster-wide mode the operator refuses custom
ServiceAccounts on Keycloak pod templates.

### spec.operatorImage

`string`

Override the operator image
(default: quay.io/keycloak/keycloak-operator at the pinned
release). Air-gapped clusters point this at their mirror; the
tag must stay at the pinned release or the CRD schema drifts.

### spec.defaultKeycloakImage

`string`

Override the DEFAULT Keycloak server image the operator stamps
into Keycloak StatefulSets whose declaration sets no image
(the bundle's RELATED_IMAGE_KEYCLOAK, default:
quay.io/keycloak/keycloak at the pinned release).

### spec.resources

`ContainerResources`

CPU and memory for the operator container. Defaults are
upstream's own (requests 300m/450Mi, limits 700m/450Mi).

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

`KubernetesKeycloakOperatorScheduling`

Pod scheduling constraints for the operator pod (upstream ships
none; the module patches them onto the Deployment).

### spec.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

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

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKeycloakOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the operator runs in (where namespaced-watch Keycloak declarations must also live). |
| `status.outputs.deployment` | `string` | The operator Deployment name (upstream-fixed: "keycloak-operator"). |
| `status.outputs.service` | `string` | The operator's metrics/health Service name (upstream-fixed: "keycloak-operator", port 80 → 8080). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)

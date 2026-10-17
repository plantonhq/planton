# KubernetesAgentSandboxClaim

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSandboxClaimSpec defines an agent-sandbox SandboxClaim: a
request for one sandbox from a KubernetesAgentSandboxWarmPool. The
agent-sandbox controller (KubernetesAgentSandboxController, extensions on)
hands the claim a ready member of the pool (a warm start, in about the time
it takes to update an object), or starts a fresh sandbox from the pool's
template when none is ready or the claim asks for something a running pod
cannot take. The claimed sandbox then belongs to the claim, and goes with it.

100% fidelity with the upstream SandboxClaim custom resource
(extensions.agents.x-k8s.io/v1beta1), pinned to agent-sandbox v1.0.7 -- the
release KubernetesAgentSandboxController installs -- and held to its served
definition by the mirror test. The upstream spec follows the Planton envelope
directly: the pool, the lifecycle, the pod metadata, and the environment
variables and volumes the claim adds.

Products usually create claims at runtime through agent-sandbox's SDKs.
Declare one here when a sandbox from a pool should exist as long as the
declaration does: a standing sandbox for a named user or team, a demo
environment, a canary that proves the pool hands out working sandboxes.
The claimed sandbox's name is in the claim's status (`sandbox.name`), which
is what a caller sends to the router.

## Example

```yaml
# The full surface of a KubernetesAgentSandboxClaim, offline: every key of the
# SandboxClaim spec. The served-schema test renders it and holds it to the
# pinned definition. The live lanes run e2e/scenarios/.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxClaim
metadata:
  name: alice-workbench
spec:
  namespace:
    value: agents
  labels:
    team: agents
  annotations:
    example.com/owner: platform
  warmPoolRef:
    name:
      value: coding-agent-pool
  lifecycle:
    shutdownTime: "2099-01-01T00:00:00Z"
    ttlSecondsAfterFinished: 3600
    shutdownPolicy: DeleteForeground
  additionalPodMetadata:
    labels:
      sandbox.users.io/user: alice
    annotations:
      cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  env:
    - name: TASK_ID
      value: task-4217
      containerName: agent
    - name: MODEL
      value: frontier-large
  volumeClaimTemplates:
    - metadata:
        name: workspace
        labels:
          example.com/user: alice
        annotations:
          example.com/backup: daily
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName:
          value: standard-rwo
        resources:
          requests:
            storage: 50Gi
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.warmPoolRef` | `KubernetesAgentSandboxClaimWarmPoolRef` | yes |  |  |
| `spec.warmPoolRef.name` | `string \| valueFrom` | yes |  | KubernetesAgentSandboxWarmPool (`status.outputs.sandbox_warm_pool_name`) |
| `spec.lifecycle` | `KubernetesAgentSandboxClaimLifecycle` |  |  |  |
| `spec.lifecycle.shutdownTime` | `string` |  |  |  |
| `spec.lifecycle.ttlSecondsAfterFinished` | `int32` |  |  |  |
| `spec.lifecycle.shutdownPolicy` | `string` |  |  |  |
| `spec.additionalPodMetadata` | `KubernetesAgentSandboxApiPodMetadata` |  |  |  |
| `spec.additionalPodMetadata.labels` | `map<string, string>` |  |  |  |
| `spec.additionalPodMetadata.annotations` | `map<string, string>` |  |  |  |
| `spec.env` | `[]KubernetesAgentSandboxClaimEnvVar` |  |  |  |
| `spec.env[].name` | `string` | yes |  |  |
| `spec.env[].value` | `string` | yes |  |  |
| `spec.env[].containerName` | `string` |  |  |  |
| `spec.volumeClaimTemplates` | `[]KubernetesAgentSandboxApiPersistentVolumeClaimTemplate` |  |  |  |
| `spec.volumeClaimTemplates[].metadata` | `KubernetesAgentSandboxApiEmbeddedObjectMetadata` | yes |  |  |
| `spec.volumeClaimTemplates[].metadata.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].metadata.labels` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].metadata.annotations` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec` | `KubernetesCoreV1ApiPersistentVolumeClaimSpec` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.accessModes` | `[]string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources` | `KubernetesCoreV1ApiVolumeResourceRequirements` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources.limits` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources.requests` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.volumeName` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.storageClassName` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.volumeClaimTemplates[].spec.volumeMode` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource` | `KubernetesCoreV1ApiTypedLocalObjectReference` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.apiGroup` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.kind` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef` | `KubernetesCoreV1ApiTypedObjectReference` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.apiGroup` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.kind` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.namespace` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.volumeAttributesClassName` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the SandboxClaim is created in. A claim draws from a
pool in its own namespace, and its sandbox lives there. Typically a
reference to a KubernetesNamespace resource's `spec.name`.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the SandboxClaim object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).
Labels for the sandbox's pod are `additional_pod_metadata.labels`.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the SandboxClaim object itself (its metadata.annotations).

### spec.warmPoolRef

`KubernetesAgentSandboxClaimWarmPoolRef` · required

The warm pool to take a sandbox from, in the claim's namespace. Upstream
reaches a template only through a pool, so this is how a claim chooses
what its sandbox runs.

- rule: {"required":true}

### spec.warmPoolRef.name

`string | valueFrom` · required

Name of the SandboxWarmPool, in the claim's namespace. Defaults to a
KubernetesAgentSandboxWarmPool reference (its sandbox_warm_pool_name
output), so the pool is created first; pass the literal name with
`value:` for a pool Planton does not manage.

- references: KubernetesAgentSandboxWarmPool (`status.outputs.sandbox_warm_pool_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesAgentSandboxWarmPool, name: <that resource's name>, fieldPath: status.outputs.sandbox_warm_pool_name}} -- a bare string does not parse

### spec.lifecycle

`KubernetesAgentSandboxClaimLifecycle`

When the claim expires and what happens then. Unset, the claim never
expires.

### spec.lifecycle.shutdownTime

`string` · optional (explicit presence)

When the claim expires, as an RFC 3339 time ("2026-12-31T18:00:00Z"). At
that moment the controller deletes the claim's sandbox (pod, Service and
claims) and applies `shutdown_policy` to the claim. The time is the
claim's own and is not copied to the sandbox. Unset, the claim never
expires.

- rule: {"string":{"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}[Tt][0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]+)?([Zz]|[+-][0-9]{2}:[0-9]{2})$"}}

### spec.lifecycle.ttlSecondsAfterFinished

`int32` · optional (explicit presence)

How long a finished claim is kept, in seconds, counted from when its
sandbox's pod finished (succeeded or failed) -- for sandboxes that run a
task to completion. Unset, a finished claim is kept.

- rule: {"int32":{"gte":0}}

### spec.lifecycle.shutdownPolicy

`string` · optional (explicit presence)

What happens to the claim when it expires (its sandbox is deleted
either way): "Retain" keeps the claim, marked expired; "Delete" deletes
it; "DeleteForeground" deletes it but keeps it visible, with a deletion
timestamp, until its sandbox and pod are gone, so a watcher can follow
the shutdown. Unset, the cluster writes "Retain". "Delete" and
"DeleteForeground" remove an object Planton still declares, so the next
apply recreates it; keep "Retain" for a declared claim.

- rule: {"string":{"in":["Delete","DeleteForeground","Retain"]}}

### spec.additionalPodMetadata

`KubernetesAgentSandboxApiPodMetadata`

Labels and annotations added to the claimed sandbox's pod, on top of the
template's. The controller refuses a label whose key lacks one of its
allowed label domains (`sandbox.users.io` and its subdomains unless the
controller's `extensions.allowed_label_domains` says otherwise), and an
annotation in a restricted system domain (except
`cluster-autoscaler.kubernetes.io/safe-to-evict`); the claim then
reports Ready=False with reason InvalidMetadata. A warm member takes
these without a restart.

### spec.additionalPodMetadata.labels

`map<string, string>`

Labels on the pod, which Services, NetworkPolicies and monitors select it
by. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.additionalPodMetadata.annotations

`map<string, string>`

Annotations on the pod: metadata for tools, such as a service mesh's
sidecar switch or a scrape hint. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations

### spec.env

`[]KubernetesAgentSandboxClaimEnvVar`

Environment variables added to the sandbox's pod. Any variable forces a
cold start (a warm member's pod is already running), and the pool's
template must allow it: its `env_vars_injection_policy` must be
"Allowed" (new names only) or "Overrides". Under "Disallowed" the claim
reports Ready=False with reason EnvVarsInjectionRejected. The values land
in the custom resource as written; keep secrets in the template's
`secretKeyRef` variables instead.

- rule: Each environment variable needs a unique name within its container

### spec.env[].name

`string` · required

Name of the environment variable: any printable characters except "=",
as Kubernetes accepts.

- rule: {"string":{"minLen":"1","pattern":"^[ -<>-~]+$"}}

### spec.env[].value

`string` · required

Value of the environment variable, as a literal. It cannot be empty
(upstream requires a value); leave the variable out instead.

- rule: {"string":{"minLen":"1"}}

### spec.env[].containerName

`string`

The container that receives the variable. Unset, the first container of
the pool's template.

### spec.volumeClaimTemplates

`[]KubernetesAgentSandboxApiPersistentVolumeClaimTemplate`

Volumes the claim's sandbox owns in addition to (or, with "Overrides",
instead of) the template's: one PersistentVolumeClaim per template,
named `<template name>-<sandbox name>`. Any template forces a cold start,
and the pool's template must allow it through its
`volume_claim_templates_policy`; otherwise the claim reports Ready=False
with reason VolumeClaimTemplatesError.

- rule: Each volume claim template needs a unique name: the pod mounts each claim by its template's name

### spec.volumeClaimTemplates[].metadata

`KubernetesAgentSandboxApiEmbeddedObjectMetadata` · required

The claim's name, labels and annotations.

- rule: {"required":true}

### spec.volumeClaimTemplates[].metadata.name

`string` · required

The template's name: the pod's volume name (a `volumeMounts` entry names
it) and the first half of the claim's name. Upstream marks it optional,
but the controller names the pod volume and the claim after it, so a
template without one produces no usable volume; it is required here.

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.volumeClaimTemplates[].metadata.labels

`map<string, string>`

Labels on the claim. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.volumeClaimTemplates[].metadata.annotations

`map<string, string>`

Annotations on the claim, such as a storage provisioner's options. More
info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations

### spec.volumeClaimTemplates[].spec

`KubernetesCoreV1ApiPersistentVolumeClaimSpec` · required

The claim's specification: access modes, size, storage class.

- rule: {"required":true}

### spec.volumeClaimTemplates[].spec.accessModes

`[]string`

accessModes contains the desired access modes the volume should have. More
info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1

- rule: {"repeated":{"items":{"string":{"in":["ReadWriteOnce","ReadOnlyMany","ReadWriteMany","ReadWriteOncePod"]}}}}

### spec.volumeClaimTemplates[].spec.selector

`KubernetesCoreV1ApiLabelSelector`

selector is a label query over volumes to consider for binding.

### spec.volumeClaimTemplates[].spec.selector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.volumeClaimTemplates[].spec.selector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.volumeClaimTemplates[].spec.resources

`KubernetesCoreV1ApiVolumeResourceRequirements`

resources represents the minimum resources the volume should have. Users
are allowed to specify resource requirements that are lower than previous
value but must still be higher than capacity recorded in the status field
of the claim. More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

### spec.volumeClaimTemplates[].spec.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.volumeClaimTemplates[].spec.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.volumeClaimTemplates[].spec.volumeName

`string`

volumeName is the binding reference to the PersistentVolume backing this
claim.

### spec.volumeClaimTemplates[].spec.storageClassName

`string | valueFrom`

storageClassName is the name of the StorageClass required by the claim.
More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1

Defaults to a KubernetesStorageClass reference (its storage_class_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.volumeClaimTemplates[].spec.volumeMode

`string` · optional (explicit presence)

volumeMode defines what type of volume is required by the claim. Value of
Filesystem is implied when not included in claim spec.

- rule: {"string":{"in":["Block","Filesystem"]}}

### spec.volumeClaimTemplates[].spec.dataSource

`KubernetesCoreV1ApiTypedLocalObjectReference`

dataSource field can be used to specify either: * An existing
VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An
existing PVC (PersistentVolumeClaim) If the provisioner or an external
controller can support the specified data source, it will create a new
volume based on the contents of the specified data source. dataSource
contents will be copied to dataSourceRef, and dataSourceRef contents will
be copied to dataSource when dataSourceRef.namespace is not specified. If
the namespace is specified, then dataSourceRef will not be copied to
dataSource.

### spec.volumeClaimTemplates[].spec.dataSource.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.volumeClaimTemplates[].spec.dataSource.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSource.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef

`KubernetesCoreV1ApiTypedObjectReference`

dataSourceRef specifies the object from which to populate the volume with
data, if a non-empty volume is desired. This may be any object from a
non-empty API group (non core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type
of the specified object matches some installed volume populator or dynamic
provisioner. This field will replace the functionality of the dataSource
field and as such if both fields are non-empty, they must have the same
value. For backwards compatibility, when namespace isn't specified in
dataSourceRef, both fields (dataSource and dataSourceRef) will be set to
the same value automatically if one of them is empty and the other is
non-empty. When namespace is specified in dataSourceRef, dataSource isn't
set to the same value and must be empty. There are three important
differences between dataSource and dataSourceRef: * While dataSource only
allows two specific types of objects, dataSourceRef allows any non-core
object, as well as PersistentVolumeClaim objects. * While dataSource
ignores disallowed values (dropping them), dataSourceRef preserves all
values, and generates an error if a disallowed value is specified. * While
dataSource only allows local objects, dataSourceRef allows objects in any
namespaces. (Alpha) Using the namespace field of dataSourceRef requires the
CrossNamespaceVolumeDataSource feature gate to be enabled.

### spec.volumeClaimTemplates[].spec.dataSourceRef.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.volumeClaimTemplates[].spec.dataSourceRef.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef.namespace

`string` · optional (explicit presence)

`namespace` is the namespace of resource being referenced Note that when a
namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object
is required in the referent namespace to allow that namespace's owner to
accept the reference. See the ReferenceGrant documentation for details.
(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate
to be enabled.

### spec.volumeClaimTemplates[].spec.volumeAttributesClassName

`string` · optional (explicit presence)

volumeAttributesClassName may be used to set the VolumeAttributesClass used
by this claim. If specified, the CSI driver will create or update the
volume with the attributes defined in the corresponding
VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string or nil value
indicates that no VolumeAttributesClass will be applied to the claim. If
the claim enters an Infeasible error state, this field can be reset to its
previous value (including nil) to cancel the modification. If the resource
referred to by volumeAttributesClass does not exist, this
PersistentVolumeClaim will be set to a Pending state, as reflected by the
modifyVolumeStatus field, until such as a resource exists. More info:
https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSandboxClaim, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sandbox_claim_name` | `string` | Name of the created SandboxClaim (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the SandboxClaim was created in (the resolved spec.namespace); its sandbox lives here. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.warmPoolRef.name` | KubernetesAgentSandboxWarmPool | `status.outputs.sandbox_warm_pool_name` |
| `spec.volumeClaimTemplates[].spec.storageClassName` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## See Also

- [Overview](../README.md)

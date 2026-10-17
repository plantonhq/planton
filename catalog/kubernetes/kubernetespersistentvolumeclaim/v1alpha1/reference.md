# KubernetesPersistentVolumeClaim

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesPersistentVolumeClaimSpec** requests persistent storage from the
cluster — the durable-disk primitive. A claim names how much storage it
needs, how it will be accessed, and (optionally) which StorageClass
provisions it; the cluster binds it to a PersistentVolume that satisfies the
request.

A standalone claim is the right shape for storage whose lifecycle is
independent of any one workload: a volume shared by several pods
(ReadWriteMany), a data volume a Deployment mounts (via the workload's
volume mounts referencing the claim by name), or a pre-provisioned volume
being adopted. Per-replica storage for a StatefulSet should use the
workload's own `volume_claim_templates` instead — those claims are stamped
and tracked by the StatefulSet controller.

BINDING TIMING: with a wait_for_first_consumer StorageClass (the norm for
zonal cloud disks and kind's local-path), a claim stays **Pending until a
pod uses it** — that is correct behavior, not a failure. Neither engine
blocks on the claim reaching Bound (deliberate, both engines, so deploys
never hang waiting for a consumer that arrives later).

The spec covers the complete core/v1 PersistentVolumeClaimSpec surface.
(The feature-gated `volumeAttributesClassName` and cross-namespace data
sources are deliberately unmodeled until they graduate.)

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises access
# modes, storage limits, an explicit storage class, block volume mode, and a
# selector. (data_source is exercised by the Pulumi-only negative proof — the
# terraform module rejects it with a precondition by design.)
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  namespace:
    value: default
  name: test-pvc
  labels:
    team: platform-engineering
  access_modes:
    - ReadWriteOnce
  storage_request: 10Gi
  storage_limit: 20Gi
  storage_class_name:
    value: standard
  volume_mode: filesystem
  selector:
    match_labels:
      tier: fast
    match_expressions:
      - key: zone
        operator: In
        values:
          - us-east-1a
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.accessModes` | `[]string` |  |  |  |
| `spec.storageRequest` | `string` | yes |  |  |
| `spec.storageLimit` | `string` |  |  |  |
| `spec.storageClassName` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.disableDynamicProvisioning` | `bool` |  |  |  |
| `spec.volumeMode` | `enum` |  | `filesystem` |  |
| `spec.volumeName` | `string` |  |  |  |
| `spec.selector` | `KubernetesPersistentVolumeClaimLabelSelector` |  |  |  |
| `spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.selector.matchExpressions` | `[]KubernetesPersistentVolumeClaimLabelSelectorRequirement` |  |  |  |
| `spec.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.selector.matchExpressions[].operator` | `string` |  |  |  |
| `spec.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.dataSource` | `KubernetesPersistentVolumeClaimDataSource` |  |  |  |
| `spec.dataSource.kind` | `enum` |  |  |  |
| `spec.dataSource.name` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom`

The namespace the claim lives in. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. When omitted, the claim
lands in the cluster's `default` namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the PersistentVolumeClaim (its `metadata.name` in the
cluster) — the value workloads reference in their PVC volume mounts.
Must be a valid DNS subdomain: lowercase alphanumeric characters,
hyphens, and dots, at most 253 characters.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the PersistentVolumeClaim object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations to apply to the PersistentVolumeClaim object.

### spec.accessModes

`[]string`

How the volume may be mounted. Defaults to ["ReadWriteOnce"] — one node
at a time, the mode every block-storage driver supports. The other modes
need driver support: "ReadWriteMany" (shared filesystems like EFS/
Filestore/Azure Files), "ReadOnlyMany" (many readers), and
"ReadWriteOncePod" (exactly one pod, the strictest isolation).
Uses the exact vocabulary workload volume-claim templates use.

- rule: {"repeated":{"items":{"string":{"in":["ReadWriteOnce","ReadOnlyMany","ReadWriteMany","ReadWriteOncePod"]}}}}

### spec.storageRequest

`string` · required

Requested storage, as a Kubernetes quantity (e.g. "10Gi", "500Mi").
Growing later requires the claim's StorageClass to allow volume
expansion (and Kubernetes never shrinks a volume).

- rule: Storage request must be a Kubernetes quantity (e.g. "10Gi", "500Mi")
- rule: {"required":true}

### spec.storageLimit

`string`

Upper bound on the volume's size, as a Kubernetes quantity. Rarely
needed — only meaningful to drivers that honor limits; most provision
exactly the request.

- rule: Storage limit must be a Kubernetes quantity (e.g. "20Gi") or empty

### spec.storageClassName

`string | valueFrom`

The StorageClass provisioning this claim. Accepts a literal class name or
a reference to a KubernetesStorageClass resource. When omitted, the
cluster's DEFAULT StorageClass applies. To opt out of dynamic
provisioning entirely (bind only to pre-provisioned volumes), set
`disable_dynamic_provisioning` instead — an empty class name and an
omitted one mean different things to Kubernetes.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.disableDynamicProvisioning

`bool`

Opts the claim OUT of dynamic provisioning by pinning its StorageClass to
the empty string (`storageClassName: ""`) — the claim then binds only to
a matching pre-provisioned PersistentVolume. This exists as its own field
because Kubernetes distinguishes an EMPTY class name (no provisioning)
from an ABSENT one (use the cluster default), a distinction a single
string field cannot carry. Mutually exclusive with `storage_class_name`.

### spec.volumeMode

`enum` · optional (explicit presence)

How the volume is exposed to containers.
Default: filesystem

- default: `filesystem`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_persistent_volume_claim_volume_mode_unspecified` -- Unspecified. Defaults to filesystem.
- `filesystem` -- A mounted filesystem — the default, what nearly every workload wants.
- `block` -- A raw block device, for workloads that manage their own on-disk format (some databases, storage systems). Requires driver support.

### spec.volumeName

`string`

Binds this claim to one specific pre-provisioned PersistentVolume by
name, skipping dynamic provisioning — the adopt-existing-data path.
The volume's capacity, access modes, and class must still satisfy the
claim.

### spec.selector

`KubernetesPersistentVolumeClaimLabelSelector`

Label selector narrowing which pre-provisioned PersistentVolumes this
claim may bind to. Only meaningful for static binding — dynamically
provisioned volumes are created for the claim, not selected.

### spec.selector.matchLabels

`map<string, string>`

Exact-match label requirements: every key/value pair listed must be
present on the PersistentVolume.

### spec.selector.matchExpressions

`[]KubernetesPersistentVolumeClaimLabelSelectorRequirement`

Set-based label requirements, for selections exact-match cannot express
(key existence, value-in-set).

- rule: values must be non-empty for In/NotIn and empty for Exists/DoesNotExist

### spec.selector.matchExpressions[].key

`string` · required

The label key the requirement applies to.

- rule: {"string":{"minLen":"1"}}

### spec.selector.matchExpressions[].operator

`string`

The relationship between the key and the values: "In" (value must be one
of `values`), "NotIn" (must not be), "Exists" (key present, `values` must
be empty), or "DoesNotExist" (key absent, `values` must be empty).

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist"]}}

### spec.selector.matchExpressions[].values

`[]string`

The values compared against the label's value. Required (non-empty) for
In/NotIn; must be empty for Exists/DoesNotExist.

### spec.dataSource

`KubernetesPersistentVolumeClaimDataSource`

Populates the new volume from an existing source instead of provisioning
it empty: clone a PersistentVolumeClaim or restore a VolumeSnapshot.
Requires a CSI driver that implements the operation. Sources are
same-namespace only (cross-namespace sources are feature-gated upstream
and deliberately unmodeled).

The operation's semantics and constraints are the DRIVER's, and they can
be stricter than the API suggests — verified live on the AWS EBS driver:
cloning is implemented internally as snapshot-then-restore (expect
minutes even for small volumes, and budget timeouts accordingly), and
EC2 refuses to clone an UNENCRYPTED source volume outright — on EKS a
clone data source only works when the source claim's StorageClass sets
`encrypted: "true"`. Restoring a VolumeSnapshot has no such encryption
constraint and completes much faster (the snapshot already exists).
Check your driver's documentation for the equivalent constraints before
depending on clones in a workflow.

### spec.dataSource.kind

`enum`

The kind of the source object.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_persistent_volume_claim_data_source_kind_unspecified` -- Unspecified.
- `persistent_volume_claim` -- Clone an existing PersistentVolumeClaim in the same namespace.
- `volume_snapshot` -- Restore a VolumeSnapshot (snapshot.storage.k8s.io) in the same namespace.

### spec.dataSource.name

`string` · required

The name of the source object, in the claim's own namespace.

- rule: {"string":{"minLen":"1"}}

## Validation Rules

- `storage_class.conflict`: disable_dynamic_provisioning pins the StorageClass to the empty string and cannot be combined with storage_class_name — set one or the other

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPersistentVolumeClaim, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pvc_name` | `string` | The name of the PersistentVolumeClaim object as created in the cluster — the value workload volume mounts reference as their claim name. |
| `status.outputs.namespace` | `string` | The namespace the claim was created in. |
| `status.outputs.storage_request` | `string` | The requested storage size, as a Kubernetes quantity (e.g. "10Gi"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storageClassName` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAgentSandbox | `spec.podTemplate.spec.volumes[].persistentVolumeClaim.claimName` | `status.outputs.pvc_name` |
| KubernetesAgentSandboxTemplate | `spec.podTemplate.spec.volumes[].persistentVolumeClaim.claimName` | `status.outputs.pvc_name` |

## See Also

- [Overview](../README.md)

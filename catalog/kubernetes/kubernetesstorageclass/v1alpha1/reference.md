# KubernetesStorageClass

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesStorageClassSpec** defines a class of storage that
PersistentVolumeClaims can request by name — the cluster's storage menu. A
StorageClass names a provisioner (the CSI driver that creates volumes), the
provisioner-specific parameters (disk type, IOPS, encryption, filesystem),
and the lifecycle policies for the volumes it provisions (reclaim, binding
timing, expandability).

StorageClasses are cluster-scoped: one class serves claims from every
namespace. Managed clusters ship default classes (EKS: gp2/gp3 via
ebs.csi.aws.com, GKE: standard-rwo/premium-rwo via pd.csi.storage.gke.io,
AKS: default/managed-csi via disk.csi.azure.com), and a custom class is how
a platform pins performance characteristics — SSD-backed, expandable,
topology-constrained — instead of inheriting whatever the cluster default
happens to be.

The spec covers the complete storage.k8s.io/v1 StorageClass surface.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises
# parameters, retain reclaim, wait_for_first_consumer binding, expansion,
# mount options, zone topology restriction, and the default-class marker.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStorageClass
metadata:
  name: test-storage-class
spec:
  name: test-storage-class
  labels:
    team: platform-engineering
  provisioner: ebs.csi.aws.com
  parameters:
    type: gp3
    iops: "6000"
    encrypted: "true"
  reclaim_policy: retain
  volume_binding_mode: wait_for_first_consumer
  allow_volume_expansion: true
  mount_options:
    - noatime
  allowed_topologies:
    - match_label_expressions:
        - key: topology.kubernetes.io/zone
          values:
            - us-east-1a
            - us-east-1b
  is_default_class: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.provisioner` | `string` | yes |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.reclaimPolicy` | `enum` |  | `delete` |  |
| `spec.volumeBindingMode` | `enum` |  | `immediate` |  |
| `spec.allowVolumeExpansion` | `bool` |  |  |  |
| `spec.mountOptions` | `[]string` |  |  |  |
| `spec.allowedTopologies` | `[]KubernetesStorageClassTopologySelectorTerm` |  |  |  |
| `spec.allowedTopologies[].matchLabelExpressions` | `[]KubernetesStorageClassTopologySelectorLabelRequirement` | yes |  |  |
| `spec.allowedTopologies[].matchLabelExpressions[].key` | `string` | yes |  |  |
| `spec.allowedTopologies[].matchLabelExpressions[].values` | `[]string` | yes |  |  |
| `spec.isDefaultClass` | `bool` |  |  |  |

## Field Details

### spec.name

`string` · required

The name of the StorageClass (its `metadata.name` in the cluster) — the
value PersistentVolumeClaims reference in `storage_class_name`.
Must be a valid DNS subdomain: lowercase alphanumeric characters, hyphens,
and dots, at most 253 characters.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the StorageClass object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations to apply to the StorageClass object. The default-class
annotation is managed by `is_default_class` — set that field instead of
adding `storageclass.kubernetes.io/is-default-class` here.

### spec.provisioner

`string` · required

The volume plugin that provisions volumes for this class — a CSI driver
name (e.g. "ebs.csi.aws.com", "pd.csi.storage.gke.io",
"disk.csi.azure.com", "rancher.io/local-path") or an in-tree plugin name.
IMMUTABLE after creation: changing the provisioner requires replacing the
class (both engines force replacement).

- rule: {"required":true}

### spec.parameters

`map<string, string>`

Provisioner-specific parameters, passed through as the class's
`parameters` — the vocabulary belongs to the provisioner, not to
Kubernetes (e.g. EBS CSI: {type: gp3, iops: "6000", encrypted: "true"};
GCE PD CSI: {type: pd-ssd}; Azure Disk CSI: {skuName: Premium_LRS}).
Kept as an upstream-shaped map deliberately: each CSI driver defines its
own keys, and typed modeling here would couple the class to one driver.
IMMUTABLE after creation. Values that name secrets (e.g.
`csi.storage.k8s.io/provisioner-secret-name`) are references to Secret
objects, never secret material itself.

### spec.reclaimPolicy

`enum` · optional (explicit presence)

Reclaim policy stamped onto volumes this class provisions.
IMMUTABLE after creation (like provisioner and parameters): changing it
replaces the class.
Default: delete

- default: `delete`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_storage_class_reclaim_policy_unspecified` -- Unspecified. Defaults to delete.
- `delete` -- Delete the backing volume with the claim — the default, right for disposable/recreatable data.
- `retain` -- Retain the backing volume after the claim is deleted; it must be reclaimed manually. Right for data that must survive claim deletion.

### spec.volumeBindingMode

`enum` · optional (explicit presence)

When claims of this class are bound and volumes provisioned.
IMMUTABLE after creation (like provisioner and parameters): changing it
replaces the class.
Default: immediate

- default: `immediate`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_storage_class_volume_binding_mode_unspecified` -- Unspecified. Defaults to immediate.
- `immediate` -- Bind and provision as soon as the claim is created — the default. Simple, but on multi-zone clusters the volume may land in a zone where the consuming pod cannot schedule.
- `wait_for_first_consumer` -- Wait until a pod actually uses the claim, then provision in the topology that pod scheduled into. The right choice for zonal storage (EBS, GCE PD, Azure Disk) — the claim stays Pending until a consumer arrives, which is correct behavior, not an error.

### spec.allowVolumeExpansion

`bool`

Allows PersistentVolumeClaims of this class to be resized after creation
(grow only — Kubernetes never shrinks volumes). Requires the CSI driver
to support expansion. Kubernetes defaults to false; enabling it costs
nothing on drivers that support it and saves a migration when a database
outgrows its disk.

### spec.mountOptions

`[]string`

Mount options for volumes of this class (e.g. ["noatime", "nodiratime"]).
NOT validated by Kubernetes — a volume with an invalid option simply
fails to mount at pod start, so keep to options the filesystem actually
supports.

### spec.allowedTopologies

`[]KubernetesStorageClassTopologySelectorTerm`

Restricts the node topologies where volumes of this class may be
provisioned (e.g. pin to specific zones). Each term's requirements AND
together; multiple terms OR. Only honored when `volume_binding_mode` is
wait_for_first_consumer — the API rejects allowed_topologies with
immediate binding for provisioners that don't support it, and the
combination is meaningless anyway (immediate binding cannot see pod
topology).

### spec.allowedTopologies[].matchLabelExpressions

`[]KubernetesStorageClassTopologySelectorLabelRequirement` · required

The topology label requirements, e.g. key
"topology.kubernetes.io/zone" with values ["us-east-1a", "us-east-1b"].

- rule: {"repeated":{"minItems":"1"}}

### spec.allowedTopologies[].matchLabelExpressions[].key

`string` · required

The node topology label key (e.g. "topology.kubernetes.io/zone").

- rule: {"string":{"minLen":"1"}}

### spec.allowedTopologies[].matchLabelExpressions[].values

`[]string` · required

The allowed values for the label — a node matches when its label value is
ANY of these.

- rule: {"repeated":{"minItems":"1"}}

### spec.isDefaultClass

`bool`

Marks this class as the cluster's DEFAULT StorageClass — the class claims
receive when they name none. Rendered as the
`storageclass.kubernetes.io/is-default-class: "true"` annotation (the
upstream mechanism, surfaced as a first-class field). A cluster should
have exactly ONE default: setting this alongside an existing default
(e.g. the managed cluster's built-in class) makes newest-wins behavior
undefined across Kubernetes versions — demote the old default first.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesStorageClass, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.storage_class_name` | `string` | The name of the StorageClass object as created in the cluster — the value claims put in their `storage_class_name`. |
| `status.outputs.provisioner` | `string` | The provisioner (CSI driver) backing this class. |
| `status.outputs.is_default_class` | `bool` | Whether this class is annotated as the cluster's default StorageClass. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesClickHouse | `spec.storageClass` | `status.outputs.storage_class_name` |
| KubernetesClickHouse | `spec.coordination.keeper.storageClass` | `status.outputs.storage_class_name` |
| KubernetesGhaRunnerScaleSet | `spec.containerMode.kubernetesWorkVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesGrafana | `spec.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesKafka | `spec.nodePools[].storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesKafka | `spec.nodePools[].storage.volumes[].storageClass` | `status.outputs.storage_class_name` |
| KubernetesKubePrometheusStack | `spec.prometheus.storageClass` | `status.outputs.storage_class_name` |
| KubernetesKubePrometheusStack | `spec.alertmanager.storageClass` | `status.outputs.storage_class_name` |
| KubernetesKubePrometheusStack | `spec.grafana.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesLoki | `spec.monolithic.storageClass` | `status.outputs.storage_class_name` |
| KubernetesLoki | `spec.simpleScalable.storageClass` | `status.outputs.storage_class_name` |
| KubernetesMongodb | `spec.replicaSets[].storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesMongodb | `spec.sharding.configServer.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesMysql | `spec.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesMysql | `spec.proxy.proxysql.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesMysql | `spec.backup.storages[].pvc.volume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesNats | `spec.jetStream.storageClass` | `status.outputs.storage_class_name` |
| KubernetesNeo4j | `spec.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesOpenBao | `spec.server.raft.dataStorage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesOpenBao | `spec.server.auditStorage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesOpenSearch | `spec.nodePools[].persistence.pvc.storageClass` | `status.outputs.storage_class_name` |
| KubernetesPersistentVolumeClaim | `spec.storageClassName` | `status.outputs.storage_class_name` |
| KubernetesPostgres | `spec.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesPostgres | `spec.walStorage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesQdrant | `spec.storage.storageClass` | `status.outputs.storage_class_name` |
| KubernetesQdrant | `spec.snapshots.storageClass` | `status.outputs.storage_class_name` |
| KubernetesRabbitMq | `spec.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSeaweedFs | `spec.master.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSeaweedFs | `spec.volume.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSeaweedFs | `spec.filer.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSeaweedFs | `spec.admin.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSignoz | `spec.server.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSolr | `spec.zookeeper.provided.persistence.storageClass` | `status.outputs.storage_class_name` |
| KubernetesSolr | `spec.storage.persistent.storageClass` | `status.outputs.storage_class_name` |
| KubernetesTempo | `spec.storageClass` | `status.outputs.storage_class_name` |
| KubernetesValkey | `spec.replication.persistence.storageClass` | `status.outputs.storage_class_name` |
| KubernetesValkey | `spec.persistence.storageClass` | `status.outputs.storage_class_name` |

## See Also

- [Overview](../README.md)

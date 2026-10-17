# KubernetesGkeComputeClass

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesGkeComputeClassSpec defines a GKE ComputeClass: a cluster-scoped,
ordered list of the machines GKE's cluster autoscaler may create nodes from.
When pods that select the class are pending, GKE tries the first priority
rule; when it cannot get that capacity (a stockout, a missing reservation, a
quota), it tries the next, until one succeeds or the list is exhausted and
`when_unsatisfiable` decides. A class is how a GKE cluster keeps getting
nodes when Google runs out of one machine type in a region.

100% fidelity with the ComputeClass custom resource (cloud.google.com/v1)
as GKE serves it at 1.36.4-gke.2046002. Every property of the served spec is
a field here under its upstream key, held to the served definition by the
mirror test (pkg/kubernetes/manifestprojection, testdata/schemas/
cloud.google.com_computeclasses.yaml). The definition's own validation rules
are mirrored below, so a class GKE would refuse fails validation before the
apply. Fields carry the GKE version that introduced them; an older cluster
prunes or refuses a field it does not serve.

How a class takes effect:
  - A pod selects it with the node selector
    `cloud.google.com/compute-class: <name>` (GKE adds the matching
    toleration for the class's node taint).
  - A namespace selects it for every pod that names none with the label
    `cloud.google.com/default-compute-class: <name>`.
  - A class named `default` applies to every pod that selects none, once the
    cluster's default compute class is enabled
    (GcpGkeCluster.cluster_autoscaling.default_compute_class_enabled).

A class only creates node pools when `node_pool_auto_creation.enabled` is
true; without it, GKE only places pods on existing node pools that match a
rule (the `nodepools` rule, or pools labelled for the class). Node pool
auto-creation needs no cluster-level node auto-provisioning on GKE
1.33.3-gke.1136000 or later.

The envelope: `name` is the object's metadata.name (the value pods select),
and `labels` and `annotations` are the object's own metadata. ComputeClass
is cluster-scoped, so there is no namespace.

## Example

```yaml
# Full-surface KubernetesGkeComputeClass: every part of the spec a Standard
# cluster accepts together, so the schema, both engines' rendering and GKE's
# own admission are exercised in one object. Zones assume a us-east1 cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGkeComputeClass
metadata:
  name: general-purpose-class
spec:
  name: general-purpose
  labels:
    team: platform
  annotations:
    example.com/owner: platform-team
  description: General-purpose services. Reserved capacity, then three families on demand.
  priorities:
    - machineType: n2d-standard-8
      reservations:
        affinity: Specific
        specific:
          - name: platform-n2d-reservation
            zones:
              - us-east1-b
      storage:
        localSSDCount: 1 # local SSDs need an exact machine type (or a GPU)
    - machineFamily: n2d
      minCores: 8
      minMemoryGb: 32
      minCpuPlatform: AMD Milan
      location:
        zones:
          - us-east1-b
          - us-east1-c
        locationPolicy: BALANCED
      storage:
        bootDiskType: pd-balanced
        bootDiskSize: 100
      nodeLabels:
        workload-tier: general
      maxPodsPerNode: 110
      allocationStrategy: lowest-cost
    - machineFamily: c3d
      minCores: 8
      spot: true
      nodeSystemConfig:
        kubeletConfig:
          shutdownGracePeriodSeconds: 30
          shutdownGracePeriodCriticalPodsSeconds: 15
    - machineFamily: c4d
      minCores: 8
      storage:
        bootDiskType: hyperdisk-balanced # C4D boots from Hyperdisk only
      location:
        zoneTypes:
          - CLUSTER_DEFAULT
  priorityDefaults:
    nodeSystemConfig:
      kubeletConfig:
        cpuCfsQuota: true
        podPidsLimit: 4096
        containerLogMaxSize: 50Mi
        containerLogMaxFiles: 5
        imageGcLowThresholdPercent: 70
        imageGcHighThresholdPercent: 85
        maxParallelImagePulls: 3
        evictionSoft:
          memoryAvailable: 200Mi
          nodefsAvailable: 15%
        evictionSoftGracePeriod:
          memoryAvailable: 1m
          nodefsAvailable: 1m
        evictionMaxPodGracePeriodSeconds: 60
      linuxNodeConfig:
        sysctls:
          net.core.somaxconn: 4096
          net.netfilter.nf_conntrack_max: 262144
          vm.max_map_count: 262144
        transparentHugepageEnabled: TRANSPARENT_HUGEPAGE_ENABLED_MADVISE
  whenUnsatisfiable: ScaleUpAnyway
  nodePoolAutoCreation:
    enabled: true
    shieldedInstanceConfig:
      enableSecureBoot: true
      enableIntegrityMonitoring: true
  nodePoolConfig:
    imageType: cos_containerd
    serviceAccount:
      value: gke-nodes@my-project.iam.gserviceaccount.com
    workloadMetadata: GKE_METADATA
    nodeLabels:
      compute-class: general-purpose
    taints:
      - key: workload-tier
        value: general
        effect: PreferNoSchedule
    imageStreaming:
      enabled: true
    gvnic:
      enabled: true
    autoRepair: true
    autoUpgrade: true
    loggingConfig:
      loggingVariantConfig:
        variant: DEFAULT
  activeMigration:
    optimizeRulePriority: true
  autoscalingPolicy:
    consolidationDelayMinutes: 10
    consolidationThreshold: 50
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.priorities` | `[]KubernetesGkeComputeClassPriority` |  |  |  |
| `spec.priorities[].machineFamily` | `string` |  |  |  |
| `spec.priorities[].machineType` | `string` |  |  |  |
| `spec.priorities[].minCores` | `int32` |  |  |  |
| `spec.priorities[].minMemoryGb` | `int32` |  |  |  |
| `spec.priorities[].minCpuPlatform` | `string` |  |  |  |
| `spec.priorities[].spot` | `bool` |  |  |  |
| `spec.priorities[].flexStart` | `KubernetesGkeComputeClassFlexStart` |  |  |  |
| `spec.priorities[].flexStart.enabled` | `bool` |  |  |  |
| `spec.priorities[].flexStart.nodeRecycling` | `KubernetesGkeComputeClassNodeRecycling` |  |  |  |
| `spec.priorities[].flexStart.nodeRecycling.leadTimeSeconds` | `int32` | yes |  |  |
| `spec.priorities[].nodepools` | `[]string \| valueFrom` |  |  | GcpGkeNodePool (`status.outputs.node_pool_name`) |
| `spec.priorities[].podFamily` | `string` |  |  |  |
| `spec.priorities[].gpu` | `KubernetesGkeComputeClassGpu` |  |  |  |
| `spec.priorities[].gpu.type` | `string` |  |  |  |
| `spec.priorities[].gpu.count` | `int64` |  |  |  |
| `spec.priorities[].gpu.driverVersion` | `string` |  |  |  |
| `spec.priorities[].gpu.gpuSharing` | `KubernetesGkeComputeClassGpuSharing` |  |  |  |
| `spec.priorities[].gpu.gpuSharing.sharingStrategy` | `string` |  |  |  |
| `spec.priorities[].gpu.gpuSharing.maxSharedClientsPerGPU` | `int64` |  |  |  |
| `spec.priorities[].gpu.gpuSharing.gpuPartitionSize` | `string` |  |  |  |
| `spec.priorities[].gpuDirect` | `string` |  |  |  |
| `spec.priorities[].tpu` | `KubernetesGkeComputeClassTpu` |  |  |  |
| `spec.priorities[].tpu.type` | `string` |  |  |  |
| `spec.priorities[].tpu.count` | `int64` |  |  |  |
| `spec.priorities[].tpu.topology` | `string` |  |  |  |
| `spec.priorities[].reservations` | `KubernetesGkeComputeClassReservations` |  |  |  |
| `spec.priorities[].reservations.affinity` | `string` |  |  |  |
| `spec.priorities[].reservations.specific` | `[]KubernetesGkeComputeClassSpecificReservation` |  |  |  |
| `spec.priorities[].reservations.specific[].name` | `string` | yes |  |  |
| `spec.priorities[].reservations.specific[].project` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.priorities[].reservations.specific[].zones` | `[]string` | yes |  |  |
| `spec.priorities[].reservations.specific[].reservationBlock` | `KubernetesGkeComputeClassReservationBlock` |  |  |  |
| `spec.priorities[].reservations.specific[].reservationBlock.name` | `string` | yes |  |  |
| `spec.priorities[].reservations.specific[].reservationBlock.reservationSubBlock` | `KubernetesGkeComputeClassReservationSubBlock` |  |  |  |
| `spec.priorities[].reservations.specific[].reservationBlock.reservationSubBlock.name` | `string` | yes |  |  |
| `spec.priorities[].location` | `KubernetesGkeComputeClassLocation` |  |  |  |
| `spec.priorities[].location.zones` | `[]string` | yes |  |  |
| `spec.priorities[].location.zoneTypes` | `[]string` | yes |  |  |
| `spec.priorities[].location.locationPolicy` | `string` |  |  |  |
| `spec.priorities[].storage` | `KubernetesGkeComputeClassStorage` |  |  |  |
| `spec.priorities[].storage.bootDiskType` | `string` |  |  |  |
| `spec.priorities[].storage.bootDiskSize` | `int32` |  |  |  |
| `spec.priorities[].storage.bootDiskKMSKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.priorities[].storage.localSSDCount` | `int32` |  |  |  |
| `spec.priorities[].storage.secondaryBootDisks` | `[]KubernetesGkeComputeClassSecondaryBootDisk` |  |  |  |
| `spec.priorities[].storage.secondaryBootDisks[].diskImageName` | `string` | yes |  |  |
| `spec.priorities[].storage.secondaryBootDisks[].project` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.priorities[].storage.secondaryBootDisks[].mode` | `string` |  |  |  |
| `spec.priorities[].placement` | `KubernetesGkeComputeClassPlacement` |  |  |  |
| `spec.priorities[].placement.policyName` | `string` | yes |  |  |
| `spec.priorities[].nodeLabels` | `map<string, string>` |  |  |  |
| `spec.priorities[].taints` | `[]KubernetesGkeComputeClassTaint` |  |  |  |
| `spec.priorities[].taints[].key` | `string` | yes |  |  |
| `spec.priorities[].taints[].value` | `string` |  |  |  |
| `spec.priorities[].taints[].effect` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig` | `KubernetesGkeComputeClassNodeSystemConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig` | `KubernetesGkeComputeClassKubeletConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.cpuCfsQuota` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.cpuCfsQuotaPeriod` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.cpuManagerPolicy` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.podPidsLimit` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.containerLogMaxSize` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.containerLogMaxFiles` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.imageGcHighThresholdPercent` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.imageGcLowThresholdPercent` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.imageMinimumGcAge` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.imageMaximumGcAge` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.allowedUnsafeSysctls` | `[]string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.maxParallelImagePulls` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.singleProcessOOMKill` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft` | `KubernetesGkeComputeClassEvictionSoft` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.memoryAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.nodefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.imagefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.pidAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod` | `KubernetesGkeComputeClassEvictionSoftGracePeriod` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim` | `KubernetesGkeComputeClassEvictionMinimumReclaim` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.crashLoopBackOff` | `KubernetesGkeComputeClassCrashLoopBackOff` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.memoryManager` | `KubernetesGkeComputeClassMemoryManager` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.memoryManager.policy` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager` | `KubernetesGkeComputeClassTopologyManager` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager.policy` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager.scope` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.shutdownGracePeriodSeconds` | `int32` |  |  |  |
| `spec.priorities[].nodeSystemConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds` | `int32` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig` | `KubernetesGkeComputeClassLinuxNodeConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls` | `KubernetesGkeComputeClassSysctls` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.aio-max-nr` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.file-max` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_instances` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_watches` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.nr_open` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.core_pattern` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.dmesg_restrict` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxbytes` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxkeys` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.kptr_restrict` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.perf_event_paranoid` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sched_rt_runtime_us` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmall` | `string` | yes |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmax` | `string` | yes |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmni` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.softlockup_panic` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sysrq` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.yama.ptrace_scope` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_poll` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_read` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.netdev_max_backlog` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.optmem_max` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_default` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_max` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.somaxconn` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_default` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_max` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh1` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh2` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh3` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_congestion_control` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_ecn` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_orphans` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_tw_buckets` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_mtu_probing` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_rmem` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_syn_retries` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_tw_reuse` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_wmem` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.all.disable_ipv6` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.default.disable_ipv6` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_acct` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_buckets` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_max` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_close_wait` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_established` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_time_wait` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_bytes` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_ratio` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_bytes` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_expire_centisecs` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_ratio` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_writeback_centisecs` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.max_map_count` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.min_free_kbytes` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_memory` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_ratio` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.swappiness` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.vfs_cache_pressure` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.watermark_scale_factor` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig` | `KubernetesGkeComputeClassHugepageConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size2m` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size1g` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.transparentHugepageEnabled` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.transparentHugepageDefrag` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig` | `KubernetesGkeComputeClassSwapConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.enabled` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile` | `KubernetesGkeComputeClassBootDiskProfile` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile` | `KubernetesGkeComputeClassEphemeralLocalSsdProfile` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile` | `KubernetesGkeComputeClassDedicatedLocalSsdProfile` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount` | `int64` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig` | `KubernetesGkeComputeClassEncryptionConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled` | `bool` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler` | `KubernetesGkeComputeClassDiskIoScheduler` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeSystemIoScheduler` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeAttachedDiskIoScheduler` | `string` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.nodeVfioConfig` | `KubernetesGkeComputeClassNodeVfioConfig` |  |  |  |
| `spec.priorities[].nodeSystemConfig.linuxNodeConfig.nodeVfioConfig.dmaEntryLimit` | `int32` |  |  |  |
| `spec.priorities[].instanceMetadata` | `map<string, string>` |  |  |  |
| `spec.priorities[].maxPodsPerNode` | `int32` |  |  |  |
| `spec.priorities[].maxRunDurationSeconds` | `int32` |  |  |  |
| `spec.priorities[].capacityCheckWaitTimeSeconds` | `int32` |  |  |  |
| `spec.priorities[].allocationStrategy` | `string` |  |  |  |
| `spec.priorities[].priorityScore` | `int32` |  |  |  |
| `spec.priorities[].minimumCapacity` | `KubernetesGkeComputeClassMinimumCapacity` |  |  |  |
| `spec.priorities[].minimumCapacity.targetNodeCount` | `int32` |  |  |  |
| `spec.priorities[].acceleratorNetworkProfile` | `string` |  |  |  |
| `spec.priorityDefaults` | `KubernetesGkeComputeClassPriorityDefaults` |  |  |  |
| `spec.priorityDefaults.location` | `KubernetesGkeComputeClassLocation` |  |  |  |
| `spec.priorityDefaults.location.zones` | `[]string` | yes |  |  |
| `spec.priorityDefaults.location.zoneTypes` | `[]string` | yes |  |  |
| `spec.priorityDefaults.location.locationPolicy` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig` | `KubernetesGkeComputeClassNodeSystemConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig` | `KubernetesGkeComputeClassKubeletConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuCfsQuota` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuCfsQuotaPeriod` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuManagerPolicy` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.podPidsLimit` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.containerLogMaxSize` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.containerLogMaxFiles` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageGcHighThresholdPercent` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageGcLowThresholdPercent` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageMinimumGcAge` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageMaximumGcAge` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.allowedUnsafeSysctls` | `[]string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.maxParallelImagePulls` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.singleProcessOOMKill` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft` | `KubernetesGkeComputeClassEvictionSoft` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.memoryAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.nodefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.imagefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.pidAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod` | `KubernetesGkeComputeClassEvictionSoftGracePeriod` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim` | `KubernetesGkeComputeClassEvictionMinimumReclaim` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.crashLoopBackOff` | `KubernetesGkeComputeClassCrashLoopBackOff` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.memoryManager` | `KubernetesGkeComputeClassMemoryManager` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.memoryManager.policy` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager` | `KubernetesGkeComputeClassTopologyManager` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager.policy` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager.scope` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.shutdownGracePeriodSeconds` | `int32` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds` | `int32` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig` | `KubernetesGkeComputeClassLinuxNodeConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls` | `KubernetesGkeComputeClassSysctls` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.aio-max-nr` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.file-max` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_instances` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_watches` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.nr_open` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.core_pattern` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.dmesg_restrict` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxbytes` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxkeys` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.kptr_restrict` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.perf_event_paranoid` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sched_rt_runtime_us` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmall` | `string` | yes |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmax` | `string` | yes |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmni` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.softlockup_panic` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sysrq` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.yama.ptrace_scope` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_poll` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_read` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.netdev_max_backlog` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.optmem_max` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_default` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_max` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.somaxconn` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_default` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_max` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh1` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh2` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh3` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_congestion_control` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_ecn` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_orphans` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_tw_buckets` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_mtu_probing` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_rmem` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_syn_retries` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_tw_reuse` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_wmem` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.all.disable_ipv6` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.default.disable_ipv6` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_acct` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_buckets` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_max` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_close_wait` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_established` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_time_wait` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_bytes` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_ratio` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_bytes` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_expire_centisecs` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_ratio` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_writeback_centisecs` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.max_map_count` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.min_free_kbytes` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_memory` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_ratio` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.swappiness` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.vfs_cache_pressure` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.watermark_scale_factor` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig` | `KubernetesGkeComputeClassHugepageConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size2m` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size1g` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.transparentHugepageEnabled` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.transparentHugepageDefrag` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig` | `KubernetesGkeComputeClassSwapConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.enabled` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile` | `KubernetesGkeComputeClassBootDiskProfile` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile` | `KubernetesGkeComputeClassEphemeralLocalSsdProfile` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile` | `KubernetesGkeComputeClassDedicatedLocalSsdProfile` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount` | `int64` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig` | `KubernetesGkeComputeClassEncryptionConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled` | `bool` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler` | `KubernetesGkeComputeClassDiskIoScheduler` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeSystemIoScheduler` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeAttachedDiskIoScheduler` | `string` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.nodeVfioConfig` | `KubernetesGkeComputeClassNodeVfioConfig` |  |  |  |
| `spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.nodeVfioConfig.dmaEntryLimit` | `int32` |  |  |  |
| `spec.whenUnsatisfiable` | `string` |  | `DoNotScaleUp` |  |
| `spec.nodePoolAutoCreation` | `KubernetesGkeComputeClassNodePoolAutoCreation` |  |  |  |
| `spec.nodePoolAutoCreation.enabled` | `bool` |  |  |  |
| `spec.nodePoolAutoCreation.shieldedInstanceConfig` | `KubernetesGkeComputeClassShieldedInstanceConfig` |  |  |  |
| `spec.nodePoolAutoCreation.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  |  |  |
| `spec.nodePoolAutoCreation.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.nodePoolConfig` | `KubernetesGkeComputeClassNodePoolConfig` |  |  |  |
| `spec.nodePoolConfig.imageType` | `string` |  |  |  |
| `spec.nodePoolConfig.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.nodePoolConfig.workloadMetadata` | `string` |  |  |  |
| `spec.nodePoolConfig.nodeLabels` | `map<string, string>` |  |  |  |
| `spec.nodePoolConfig.taints` | `[]KubernetesGkeComputeClassTaint` |  |  |  |
| `spec.nodePoolConfig.taints[].key` | `string` | yes |  |  |
| `spec.nodePoolConfig.taints[].value` | `string` |  |  |  |
| `spec.nodePoolConfig.taints[].effect` | `string` |  |  |  |
| `spec.nodePoolConfig.taintConfig` | `KubernetesGkeComputeClassTaintConfig` |  |  |  |
| `spec.nodePoolConfig.taintConfig.architectureTaintBehavior` | `string` |  |  |  |
| `spec.nodePoolConfig.ipType` | `string` |  |  |  |
| `spec.nodePoolConfig.confidentialNodeType` | `string` |  |  |  |
| `spec.nodePoolConfig.sandbox` | `KubernetesGkeComputeClassSandbox` |  |  |  |
| `spec.nodePoolConfig.sandbox.type` | `string` |  |  |  |
| `spec.nodePoolConfig.imageStreaming` | `KubernetesGkeComputeClassImageStreaming` |  |  |  |
| `spec.nodePoolConfig.imageStreaming.enabled` | `bool` |  |  |  |
| `spec.nodePoolConfig.gvnic` | `KubernetesGkeComputeClassGvnic` |  |  |  |
| `spec.nodePoolConfig.gvnic.enabled` | `bool` |  |  |  |
| `spec.nodePoolConfig.autoRepair` | `bool` |  |  |  |
| `spec.nodePoolConfig.autoUpgrade` | `bool` |  |  |  |
| `spec.nodePoolConfig.maintenanceExclusion` | `string` |  |  |  |
| `spec.nodePoolConfig.resourceManagerTags` | `[]KubernetesGkeComputeClassResourceManagerTag` |  |  |  |
| `spec.nodePoolConfig.resourceManagerTags[].key` | `string` | yes |  |  |
| `spec.nodePoolConfig.resourceManagerTags[].value` | `string` | yes |  |  |
| `spec.nodePoolConfig.loggingConfig` | `KubernetesGkeComputeClassLoggingConfig` |  |  |  |
| `spec.nodePoolConfig.loggingConfig.loggingVariantConfig` | `KubernetesGkeComputeClassLoggingVariantConfig` |  |  |  |
| `spec.nodePoolConfig.loggingConfig.loggingVariantConfig.variant` | `string` |  |  |  |
| `spec.nodePoolConfig.instanceMetadata` | `map<string, string>` |  |  |  |
| `spec.nodePoolConfig.dra` | `KubernetesGkeComputeClassDra` |  |  |  |
| `spec.nodePoolConfig.dra.networking` | `KubernetesGkeComputeClassDraNetworking` |  |  |  |
| `spec.nodePoolConfig.dra.networking.enabled` | `bool` |  |  |  |
| `spec.nodePoolConfig.workloadType` | `string` |  |  |  |
| `spec.nodePoolGroup` | `KubernetesGkeComputeClassNodePoolGroup` |  |  |  |
| `spec.nodePoolGroup.name` | `string` | yes |  |  |
| `spec.activeMigration` | `KubernetesGkeComputeClassActiveMigration` |  |  |  |
| `spec.activeMigration.optimizeRulePriority` | `bool` |  |  |  |
| `spec.activeMigration.ensureAllDaemonSetPodsRunning` | `bool` |  |  |  |
| `spec.autoscalingPolicy` | `KubernetesGkeComputeClassAutoscalingPolicy` |  |  |  |
| `spec.autoscalingPolicy.consolidationDelayMinutes` | `int32` |  |  |  |
| `spec.autoscalingPolicy.consolidationThreshold` | `int32` |  |  |  |
| `spec.autoscalingPolicy.gpuConsolidationThreshold` | `int32` |  |  |  |
| `spec.allocationStrategyDefaults` | `KubernetesGkeComputeClassAllocationStrategyDefaults` |  |  |  |
| `spec.allocationStrategyDefaults.onDemand` | `string` |  |  |  |
| `spec.allocationStrategyDefaults.spot` | `string` |  |  |  |
| `spec.allocationStrategyDefaults.flexStart` | `string` |  |  |  |
| `spec.minimumCapacity` | `KubernetesGkeComputeClassMinimumCapacity` |  |  |  |
| `spec.minimumCapacity.targetNodeCount` | `int32` |  |  |  |
| `spec.autopilot` | `KubernetesGkeComputeClassAutopilot` |  |  |  |
| `spec.autopilot.enabled` | `bool` |  |  |  |

## Field Details

### spec.name

`string` · required

The ComputeClass's name in the cluster (its metadata.name): the value a
pod selects with `cloud.google.com/compute-class: <name>` and a namespace
with `cloud.google.com/default-compute-class: <name>`.

`default` is special: it becomes the cluster's default class, applied to
every pod that selects none, once the cluster's default compute class is
enabled (GcpGkeCluster.cluster_autoscaling.default_compute_class_enabled).
Because the name lives here rather than in the resource's own name, every
cluster in an environment can carry its own `default` class.

A DNS subdomain name. Names beginning with `autopilot` belong to GKE's
built-in classes (autopilot, autopilot-spot, autopilot-arm,
autopilot-arm-spot) and are refused. Renaming a class creates a new object
and deletes the old one; pods that select the old name stop matching.

- rule: name must be a valid DNS subdomain (lowercase alphanumeric, hyphens and dots, starting and ending with an alphanumeric)
- rule: names beginning with 'autopilot' belong to GKE's built-in compute classes
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Labels on the ComputeClass object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).
For policy engines and inventory that select classes by label. They do not
reach nodes; `node_pool_config.node_labels` and each rule's `node_labels`
do.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the ComputeClass object itself (its metadata.annotations):
notes for people and tools reading the object.

### spec.description

`string` · optional (explicit presence)

Free text telling workload owners when to select this class ("batch jobs
that tolerate Spot", "latency-sensitive services"). Shown by
`kubectl describe computeclass`; GKE does not act on it.

### spec.priorities

`[]KubernetesGkeComputeClassPriority`

The priority rules, tried in order (or by `priority_score` when set):
when pods that select the class are pending, GKE creates nodes from the
first rule it can satisfy. Each rule names one way to get a node:
  - a machine family (`machine_family`, optionally with `min_cores` and
    `min_memory_gb`) or an exact `machine_type`;
  - existing node pools (`nodepools`), on their own;
  - an accelerator (`gpu` or `tpu`);
  - a pod family (`pod_family`, Autopilot classes only);
and how to pay for it: `spot`, `flex_start`, or `reservations`.

Order is the whole design. List the shape you want first, then the
families that can run the same pods when the first runs out: the
stockout-resilient default is several general-purpose families of the
same size (n2d, c3d, n2, t2d at 8 cores). A rule GKE cannot satisfy is
skipped, so a reservation or Spot rule first, followed by on-demand
rules, spends the cheapest capacity before the dearest.

Empty is valid upstream (an Autopilot class, or a class used only for its
node pool configuration), and then GKE creates nodes from its default
machine series. Up to 1000 rules. Requires GKE 1.30.3-gke.1451000.

- rule: {"repeated":{"maxItems":"1000"}}
- rule: a nodepools rule may carry only allocation_strategy and priority_score beside it -- the pools' own configuration decides the node
- rule: machine_family and machine_type cannot be set together
- rule: machine_type cannot be set together with min_cores or min_memory_gb -- the machine type fixes both
- rule: the ek machine family is reserved for Autopilot and cannot be named in a rule
- rule: the flex-start provisioning model is incompatible with spot
- rule: capacity_check_wait_time_seconds is supported only with flex_start enabled or a tpu
- rule: kubelet shutdown_grace_period_seconds is supported only on a spot rule
- rule: accelerator_network_profile must be set when gpu_direct is rdma
- rule: a rule-level minimum_capacity target requires a machine_type, gpu, tpu, or Specific reservation

### spec.priorities[].machineFamily

`string` · optional (explicit presence)

The Compute Engine machine family the node comes from (n2d, c3d, n2, t2d,
c4, n4, e2, ...), with GKE choosing the size from the pending pods'
requests, bounded by `min_cores` and `min_memory_gb`. Families differ in
the disks they take: C4, C4D, N4 and other fourth-generation families use
Hyperdisk only, so pair them with `storage.boot_disk_type`
hyperdisk-balanced. Cannot be set with `machine_type`. Up to 10
characters. Requires GKE 1.30.3-gke.1451000.

- rule: {"string":{"maxLen":"10"}}

### spec.priorities[].machineType

`string` · optional (explicit presence)

An exact machine type (n2d-standard-8, a3-highgpu-1g). Use it when the
shape is a requirement; prefer `machine_family` plus minimums when any
size that fits will do, because a family has more capacity to fall back
on. Cannot be set with `machine_family`, `min_cores` or `min_memory_gb`.
Up to 100 characters. Requires GKE 1.30.3-gke.1451000.

- rule: {"string":{"maxLen":"100"}}

### spec.priorities[].minCores

`int32` · optional (explicit presence)

The fewest vCPUs a node from this rule may have. With `machine_family`,
keeps GKE from choosing a smaller machine than the workloads were sized
for (the stockout default uses 8 on every rule so each fallback is the
same size). More than any machine of the family offers marks the class
CrdMisconfigured (NoSuitableMachineExists). Requires GKE
1.30.3-gke.1451000.

- rule: {"int32":{"gte":0}}

### spec.priorities[].minMemoryGb

`int32` · optional (explicit presence)

The least memory, in GB, a node from this rule may have. Requires GKE
1.30.3-gke.1451000.

- rule: {"int32":{"gte":0}}

### spec.priorities[].minCpuPlatform

`string` · optional (explicit presence)

The oldest CPU platform the node may run on, for workloads that need an
instruction set or a performance floor. Requires GKE 1.34.1-gke.2541000.

- rule: {"string":{"in":["Intel Sandy Bridge","Intel Ivy Bridge","Intel Haswell","Intel Broadwell","Intel Skylake","Intel Cascade Lake","Intel Ice Lake","Intel Sapphire Rapids","Intel Emerald Rapids","Intel Granite Rapids","AMD Rome","AMD Milan","AMD Genoa","AMD Turin","Ampere Altra","Google Axion","Nvidia Grace"]}}

### spec.priorities[].spot

`bool` · optional (explicit presence)

Spot VMs: the same machines at a deep discount, reclaimed by Google with
30 seconds' notice. A Spot rule followed by the same shape on demand is
the cheap-first pattern; Spot capacity also runs out, so the on-demand
rule is the fallback, not an extra. Incompatible with `flex_start`.

### spec.priorities[].flexStart

`KubernetesGkeComputeClassFlexStart`

The flex-start provisioning model: GKE queues the request and creates the
nodes (usually accelerators) when capacity frees up, for a run of up to
seven days, at a discount. For training and batch work that can wait to
start. Incompatible with `spot`. Requires GKE 1.32.2-gke.1652000.

### spec.priorities[].flexStart.enabled

`bool` · optional (explicit presence)

Uses flex-start for the rule.

### spec.priorities[].flexStart.nodeRecycling

`KubernetesGkeComputeClassNodeRecycling`

Starts a replacement node before the current one reaches its end of
life.

### spec.priorities[].flexStart.nodeRecycling.leadTimeSeconds

`int32` · required · optional (explicit presence)

How many seconds (1-604800) before a node's termination GKE starts
looking for its replacement.

- rule: {"required":true,"int32":{"lte":604800,"gte":1}}

### spec.priorities[].nodepools

`[]string | valueFrom`

Existing node pools to place pods on, by name, before (or instead of)
creating new ones: the way a class uses capacity the cluster already pays
for. The pools must belong to the cluster the class lives in. A rule with
`nodepools` carries nothing else but `allocation_strategy` and
`priority_score`; the pools' own configuration decides the node. Not
available on Autopilot classes. Requires GKE 1.30.3-gke.1451000.

- references: GcpGkeNodePool (`status.outputs.node_pool_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeNodePool, name: <that resource's name>, fieldPath: status.outputs.node_pool_name}} -- a bare string does not parse

### spec.priorities[].podFamily

`string` · optional (explicit presence)

Pod-based provisioning and billing: GKE sizes and bills the node by the
pods it runs. Autopilot classes only (`autopilot.enabled`). Requires GKE
1.35.2-gke.1485000.

- rule: {"string":{"in":["general-purpose","general-purpose-arm"]}}

### spec.priorities[].gpu

`KubernetesGkeComputeClassGpu`

GPUs to attach to each node of this rule (type, count, driver, sharing).
Requires GKE 1.30.3-gke.1451000.

### spec.priorities[].gpu.type

`string` · optional (explicit presence)

The GPU model, such as nvidia-l4, nvidia-a100-80gb or nvidia-h100-80gb.
Its availability differs by zone; pair the rule with a `location`.

### spec.priorities[].gpu.count

`int64` · optional (explicit presence)

GPUs per node. A machine type with built-in GPUs (A2, A3, G2) fixes the
count.

- rule: {"int64":{"gte":"0"}}

### spec.priorities[].gpu.driverVersion

`string` · optional (explicit presence)

Which NVIDIA driver GKE installs: default (the node version's default),
latest, or autoinstall-disabled (install your own; requires GKE
1.36.2-gke.2064000). Upstream default `default`. Requires GKE
1.31.1-gke.1858000.

- rule: {"string":{"in":["default","latest","autoinstall-disabled"]}}

### spec.priorities[].gpu.gpuSharing

`KubernetesGkeComputeClassGpuSharing`

GPU sharing between containers (time-sharing, MPS, multi-instance
partitions). Requires GKE 1.35.2-gke.1485000.

### spec.priorities[].gpu.gpuSharing.sharingStrategy

`string` · optional (explicit presence)

TIME_SHARING (containers take turns) or MPS (NVIDIA Multi-Process
Service, concurrent).

- rule: {"string":{"in":["MPS","TIME_SHARING"]}}

### spec.priorities[].gpu.gpuSharing.maxSharedClientsPerGPU

`int64` · optional (explicit presence)

How many containers may share one physical GPU (2-48 per GKE; required
with a sharing strategy).

- rule: {"int64":{"gte":"0"}}

### spec.priorities[].gpu.gpuSharing.gpuPartitionSize

`string` · optional (explicit presence)

Multi-instance GPU partition size, such as 1g.5gb; must be one the GPU
model supports. Works with either sharing strategy.

### spec.priorities[].gpuDirect

`string` · optional (explicit presence)

GPUDirect for the rule's nodes; the only value is `rdma`, which needs an
`accelerator_network_profile`.

- rule: {"string":{"const":"rdma"}}

### spec.priorities[].tpu

`KubernetesGkeComputeClassTpu`

TPUs for each node of this rule (type, chip count, topology). Requires
GKE 1.31.2-gke.1518000.

### spec.priorities[].tpu.type

`string` · optional (explicit presence)

The TPU type, such as tpu-v5-lite-podslice.

### spec.priorities[].tpu.count

`int64` · optional (explicit presence)

TPU chips per node.

### spec.priorities[].tpu.topology

`string` · optional (explicit presence)

The slice topology, such as 2x4.

### spec.priorities[].reservations

`KubernetesGkeComputeClassReservations`

Which Compute Engine reservations the rule's nodes consume, and whether
the rule may fall back to on-demand capacity. A Specific reservation
first, then on-demand rules, spends committed capacity before buying
more. Requires GKE 1.31.1-gke.2105000.

- rule: specific reservations can be listed only with affinity Specific
- rule: affinity Specific requires at least one specific reservation

### spec.priorities[].reservations.affinity

`string`

How reservations are consumed:
  - Specific: only the listed reservations, with no fallback (the rule is
    skipped when they are full). GKE checks the reservations exist and
    are reachable from the cluster's project; a missing one marks the
    whole class CrdMisconfigured (ReservationNotFound), not just the
    rule.
  - AnyBestEffort: any matching non-specific reservation, falling back to
    on-demand.
  - AnyThenFail: any matching non-specific reservation, never on-demand.
  - None: never consume a reservation.

- rule: {"string":{"in":["Specific","AnyBestEffort","None","AnyThenFail"]}}

### spec.priorities[].reservations.specific

`[]KubernetesGkeComputeClassSpecificReservation`

The reservations a Specific rule consumes, in no order.

### spec.priorities[].reservations.specific[].name

`string` · required

The reservation's name.

- rule: {"string":{"minLen":"1"}}

### spec.priorities[].reservations.specific[].project

`string | valueFrom`

The project that owns the reservation, when it is shared from another
project. Unset, the cluster's project. The class runs in the cluster, not
in this project, so the reference never nests it there.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.priorities[].reservations.specific[].zones

`[]string` · required

The zones the reservation is consumed in; each must be in the cluster's
region, or GKE refuses the class.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"minItems":"1"}}

### spec.priorities[].reservations.specific[].reservationBlock

`KubernetesGkeComputeClassReservationBlock`

A block of the reservation, for reservations split into blocks. Requires
GKE 1.31.4-gke.1072000.

### spec.priorities[].reservations.specific[].reservationBlock.name

`string` · required

The block's name.

- rule: {"string":{"minLen":"1"}}

### spec.priorities[].reservations.specific[].reservationBlock.reservationSubBlock

`KubernetesGkeComputeClassReservationSubBlock`

A sub-block of the block. Requires GKE 1.33.1-gke.1788000.

### spec.priorities[].reservations.specific[].reservationBlock.reservationSubBlock.name

`string` · required

The sub-block's name.

- rule: {"string":{"minLen":"1"}}

### spec.priorities[].location

`KubernetesGkeComputeClassLocation`

The zones the rule may create nodes in, and how it spreads across them.
Overrides `priority_defaults.location`. Requires GKE 1.33.1-gke.1545000.

### spec.priorities[].location.zones

`[]string` · required

The zones nodes may be created in. GKE refuses a zone outside the
cluster's region (or its node locations). Requires GKE
1.34.1-gke.3759000.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"minItems":"1"}}

### spec.priorities[].location.zoneTypes

`[]string` · required

Sets of zones instead of a list: STANDARD (the region's core zones), AI
(zones built for AI hardware), CLUSTER_DEFAULT (the cluster's
auto-provisioning locations, or its locations). GKE refuses zone types in
`priority_defaults` when any rule consumes a Specific reservation (the
reservation fixes its own zones). Requires GKE 1.35.2-gke.1842000.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"minItems":"1","maxItems":"3","items":{"string":{"in":["STANDARD","AI","CLUSTER_DEFAULT"]}}}}

### spec.priorities[].location.locationPolicy

`string` · optional (explicit presence)

How new nodes spread across the zones: BALANCED (even spread) or ANY
(wherever capacity is, which finds scarce machines and Spot capacity
sooner). Requires GKE 1.34.1-gke.2541000.

- rule: {"string":{"in":["ANY","BALANCED"]}}

### spec.priorities[].storage

`KubernetesGkeComputeClassStorage`

The node's boot disk (size, type, CMEK key), local SSDs and secondary
boot disks. Requires GKE 1.30.3-gke.1639000.

### spec.priorities[].storage.bootDiskType

`string` · optional (explicit presence)

Boot disk type. Families accept different types: C4, C4D, N4 and other
fourth-generation families take hyperdisk-balanced only, while N2D, C3D,
N2 and E2 take pd-balanced, pd-ssd or pd-standard. Requires GKE
1.34.1-gke.1431000.

- rule: {"string":{"in":["pd-balanced","pd-standard","pd-ssd","hyperdisk-balanced"]}}

### spec.priorities[].storage.bootDiskSize

`int32` · optional (explicit presence)

Boot disk size in GB (10 or more). Requires GKE 1.34.1-gke.1431000.

- rule: {"int32":{"gte":10}}

### spec.priorities[].storage.bootDiskKMSKey

`string | valueFrom`

The Cloud KMS key that encrypts the boot disk (CMEK), as
projects/<p>/locations/<l>/keyRings/<r>/cryptoKeys/<k>. The Compute
Engine service agent needs encrypter/decrypter on the key.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.priorities[].storage.localSSDCount

`int32` · optional (explicit presence)

Local SSDs per node (1 or more), for scratch space and caches. GKE
accepts it only on a rule with a `machine_type` or a `gpu`, since the
count a machine can carry depends on its exact shape.

- rule: {"int32":{"gte":1}}

### spec.priorities[].storage.secondaryBootDisks

`[]KubernetesGkeComputeClassSecondaryBootDisk`

Persistent disks preloaded with data (a container image cache) attached
to every node, so large images start without a pull. Requires GKE
1.31.2-gke.1105000.

### spec.priorities[].storage.secondaryBootDisks[].diskImageName

`string` · required

The disk image's name.

- rule: {"string":{"minLen":"1"}}

### spec.priorities[].storage.secondaryBootDisks[].project

`string | valueFrom`

The project that owns the disk image. Unset, the cluster's project. An
access reference: the class never nests in this project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.priorities[].storage.secondaryBootDisks[].mode

`string` · optional (explicit presence)

CONTAINER_IMAGE_CACHE uses the disk as a container image cache.

- rule: {"string":{"in":["MODE_UNSPECIFIED","CONTAINER_IMAGE_CACHE"]}}

### spec.priorities[].placement

`KubernetesGkeComputeClassPlacement`

A Compute Engine resource policy (compact placement or workload policy)
the rule's nodes are created under. Requires GKE 1.33.2-gke.1335000.

### spec.priorities[].placement.policyName

`string` · required

The Compute Engine resource policy's name, such as a compact placement
policy for low-latency multi-node jobs.

- rule: {"string":{"minLen":"1"}}

### spec.priorities[].nodeLabels

`map<string, string>`

Kubernetes labels on every node from this rule, in addition to
`node_pool_config.node_labels`. Requires GKE 1.33.2-gke.1111000.

### spec.priorities[].taints

`[]KubernetesGkeComputeClassTaint`

Kubernetes taints on every node from this rule, in addition to
`node_pool_config.taints`. Up to 100. Requires GKE 1.33.4-gke.1350000.

- rule: {"repeated":{"maxItems":"100"}}

### spec.priorities[].taints[].key

`string` · required

The taint key, in Kubernetes label-key syntax.

- rule: {"string":{"minLen":"1","maxLen":"320"}}

### spec.priorities[].taints[].value

`string` · optional (explicit presence)

The taint value.

- rule: {"string":{"maxLen":"63","pattern":"^([a-z0-9][-A-Za-z0-9_.]{1,61})?[A-Za-z0-9]$"}}

### spec.priorities[].taints[].effect

`string`

What happens to pods that do not tolerate it: NoSchedule,
PreferNoSchedule or NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.priorities[].nodeSystemConfig

`KubernetesGkeComputeClassNodeSystemConfig`

Kubelet and Linux kernel settings for the rule's nodes. Overrides
`priority_defaults.node_system_config`. Requires GKE 1.32.1-gke.1729000.

### spec.priorities[].nodeSystemConfig.kubeletConfig

`KubernetesGkeComputeClassKubeletConfig`

Kubelet settings.

- rule: image_gc_low_threshold_percent must be lower than image_gc_high_threshold_percent
- rule: image_gc_high_threshold_percent alone must be higher than 80, the default image_gc_low_threshold_percent
- rule: shutdown_grace_period_critical_pods_seconds requires shutdown_grace_period_seconds and must not exceed it

### spec.priorities[].nodeSystemConfig.kubeletConfig.cpuCfsQuota

`bool` · optional (explicit presence)

Enforces pods' CPU limits with the CFS quota. False ignores CPU limits,
for latency-sensitive pods that throttle, at the risk of a pod taking
more CPU than intended.

### spec.priorities[].nodeSystemConfig.kubeletConfig.cpuCfsQuotaPeriod

`string` · optional (explicit presence)

The CFS quota period, between 1ms and 1s (for example 100ms).

- rule: {"string":{"pattern":"^([1-9][0-9]*)m?s$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.cpuManagerPolicy

`string` · optional (explicit presence)

none (the default) or static, which gives Guaranteed pods with integer
CPU requests exclusive cores.

- rule: {"string":{"in":["none","static"]}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.podPidsLimit

`int64` · optional (explicit presence)

The most process IDs a pod may use (1024-4194304).

- rule: {"int64":{"lte":"4194304","gte":"1024"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.containerLogMaxSize

`string` · optional (explicit presence)

The size a container log reaches before it is rotated, 10Mi-500Mi (for
example 100Mi). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(Ki|Mi|Gi))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.containerLogMaxFiles

`int64` · optional (explicit presence)

How many rotated log files a container keeps (2-10). Requires GKE
1.33.4-gke.1036000.

- rule: {"int64":{"lte":"10","gte":"2"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.imageGcHighThresholdPercent

`int64` · optional (explicit presence)

Disk usage (11-85 percent) above which image garbage collection always
runs. Requires GKE 1.33.4-gke.1036000.

- rule: {"int64":{"lte":"85","gte":"11"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.imageGcLowThresholdPercent

`int64` · optional (explicit presence)

Disk usage (10-84 percent) image garbage collection frees down to.
Requires GKE 1.33.4-gke.1036000.

- rule: {"int64":{"lte":"84","gte":"10"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.imageMinimumGcAge

`string` · optional (explicit presence)

The youngest an unused image may be when collected, up to 2m (for
example 1m). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.imageMaximumGcAge

`string` · optional (explicit presence)

The longest an image may stay unused before collection (for example
2h45m). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.allowedUnsafeSysctls

`[]string`

Unsafe sysctls or patterns (ending in `*`) pods may set, from the groups
kernel.shm*, kernel.msg*, kernel.sem, fs.mqueue.* and net.*. Up to 100.
Requires GKE 1.33.4-gke.1036000.

- rule: {"repeated":{"maxItems":"100","items":{"string":{"minLen":"1","maxLen":"253","pattern":"^([a-z0-9]([-_a-z0-9]*[a-z0-9])?[./])*([a-z0-9][-_a-z0-9]*)?[a-z0-9*]$"}}}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.maxParallelImagePulls

`int64` · optional (explicit presence)

Image pulls the kubelet runs in parallel (2-5). Requires GKE
1.34.1-gke.3084001.

- rule: {"int64":{"lte":"5","gte":"2"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.singleProcessOOMKill

`bool` · optional (explicit presence)

OOM-kills a container's processes one at a time instead of as a group.
Requires GKE 1.34.1-gke.3084001.

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds

`int64` · optional (explicit presence)

The longest grace period (0-300 seconds) given to pods evicted by a soft
eviction threshold.

- rule: {"int64":{"lte":"300","gte":"0"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft

`KubernetesGkeComputeClassEvictionSoft`

Soft eviction thresholds.

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.memoryAvailable

`string` · optional (explicit presence)

memory.available threshold as a quantity (for example 200Mi), above the
100Mi hard threshold and under half the machine's memory.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ki|Mi|Gi)$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.nodefsAvailable

`string` · optional (explicit presence)

nodefs.available threshold, 10%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.nodefsInodesFree

`string` · optional (explicit presence)

nodefs.inodesFree threshold, 5%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.imagefsAvailable

`string` · optional (explicit presence)

imagefs.available threshold, 15%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.imagefsInodesFree

`string` · optional (explicit presence)

imagefs.inodesFree threshold, 5%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoft.pidAvailable

`string` · optional (explicit presence)

pid.available threshold, 10%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod

`KubernetesGkeComputeClassEvictionSoftGracePeriod`

Grace periods for the soft eviction thresholds.

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable

`string` · optional (explicit presence)

For memory.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable

`string` · optional (explicit presence)

For nodefs.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree

`string` · optional (explicit presence)

For nodefs.inodesFree.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable

`string` · optional (explicit presence)

For imagefs.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree

`string` · optional (explicit presence)

For imagefs.inodesFree.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable

`string` · optional (explicit presence)

For pid.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim

`KubernetesGkeComputeClassEvictionMinimumReclaim`

The least the kubelet reclaims once an eviction starts.

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable

`string` · optional (explicit presence)

For memory.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable

`string` · optional (explicit presence)

For nodefs.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree

`string` · optional (explicit presence)

For nodefs.inodesFree.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable

`string` · optional (explicit presence)

For imagefs.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree

`string` · optional (explicit presence)

For imagefs.inodesFree.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable

`string` · optional (explicit presence)

For pid.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.crashLoopBackOff

`KubernetesGkeComputeClassCrashLoopBackOff`

Container restart back-off.

### spec.priorities[].nodeSystemConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod

`string` · optional (explicit presence)

The longest the restart back-off grows to, under 5m (for example 30s).

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.memoryManager

`KubernetesGkeComputeClassMemoryManager`

The kubelet Memory Manager. Requires GKE 1.36.2-gke.1368000.

### spec.priorities[].nodeSystemConfig.kubeletConfig.memoryManager.policy

`string` · optional (explicit presence)

None (the default, no alignment) or Static, which gives Guaranteed pods
memory from a single NUMA node.

- rule: {"string":{"in":["None","Static"]}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager

`KubernetesGkeComputeClassTopologyManager`

The kubelet Topology Manager. Requires GKE 1.36.2-gke.1368000.

### spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager.policy

`string` · optional (explicit presence)

none (the default), best-effort, restricted, or single-numa-node (pods
that do not fit one NUMA node are refused).

- rule: {"string":{"in":["none","best-effort","restricted","single-numa-node"]}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.topologyManager.scope

`string` · optional (explicit presence)

container (the default, aligns each container) or pod (aligns the pod as
a whole).

- rule: {"string":{"in":["container","pod"]}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.shutdownGracePeriodSeconds

`int32` · optional (explicit presence)

Seconds (0, 30 or 120) a node delays shutdown so pods terminate
gracefully. Spot rules only. Requires GKE control plane
1.36.0-gke.3204000.

- rule: {"int32":{"in":[0,30,120]}}

### spec.priorities[].nodeSystemConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds

`int32` · optional (explicit presence)

Seconds (0-120, at most shutdown_grace_period_seconds) reserved for
critical pods during that shutdown. Requires GKE control plane
1.36.0-gke.3204000.

- rule: {"int32":{"lte":120,"gte":0}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig

`KubernetesGkeComputeClassLinuxNodeConfig`

Linux kernel settings.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls

`KubernetesGkeComputeClassSysctls`

Kernel sysctls GKE allows on nodes.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.aio-max-nr

`int64` · optional (explicit presence)

fs.aio-max-nr: system-wide asynchronous IO requests (65536-4194304).

- rule: {"int64":{"lte":"4194304","gte":"65536"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.file-max

`int64` · optional (explicit presence)

fs.file-max: file handles the kernel allocates (104857-67108864).

- rule: {"int64":{"lte":"67108864","gte":"104857"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_instances

`int64` · optional (explicit presence)

fs.inotify.max_user_instances: inotify instances per user (8192-1048576).

- rule: {"int64":{"lte":"1048576","gte":"8192"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_watches

`int64` · optional (explicit presence)

fs.inotify.max_user_watches: inotify watches per user (8192-1048576).

- rule: {"int64":{"lte":"1048576","gte":"8192"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.fs.nr_open

`int64` · optional (explicit presence)

fs.nr_open: file descriptors per process (1048576-2147483584).

- rule: {"int64":{"lte":"2147483584","gte":"1048576"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.core_pattern

`string` · optional (explicit presence)

kernel.core_pattern: the core dump path, absolute only (no pipes).

- rule: {"string":{"maxLen":"128","pattern":"^/[a-zA-Z0-9/._%-]+$"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.dmesg_restrict

`bool` · optional (explicit presence)

kernel.dmesg_restrict: keeps unprivileged users from reading the kernel
log.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxbytes

`int64` · optional (explicit presence)

kernel.keys.maxbytes: key payload bytes per non-root user
(20000-2097152).

- rule: {"int64":{"lte":"2097152","gte":"20000"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxkeys

`int64` · optional (explicit presence)

kernel.keys.maxkeys: keys per non-root user (200-1048576).

- rule: {"int64":{"lte":"1048576","gte":"200"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.kptr_restrict

`int64` · optional (explicit presence)

kernel.kptr_restrict: hides kernel addresses (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.perf_event_paranoid

`int64` · optional (explicit presence)

kernel.perf_event_paranoid: unprivileged use of perf events (-1-3).

- rule: {"int64":{"lte":"3","gte":"-1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sched_rt_runtime_us

`int64` · optional (explicit presence)

kernel.sched_rt_runtime_us: real-time scheduling budget (-1-1000000).

- rule: {"int64":{"lte":"1000000","gte":"-1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmall

`string` · required · optional (explicit presence)

kernel.shmall: total shared memory pages, as a decimal string (up to 20
digits; the range exceeds a 64-bit signed integer).

- rule: {"string":{"minLen":"1","maxLen":"20","pattern":"^([0-9]+)$"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmax

`string` · required · optional (explicit presence)

kernel.shmmax: largest shared memory segment in bytes, as a decimal
string (up to 20 digits).

- rule: {"string":{"minLen":"1","maxLen":"20","pattern":"^([0-9]+)$"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmni

`int64` · optional (explicit presence)

kernel.shmmni: shared memory segments system-wide (4096-32768).

- rule: {"int64":{"lte":"32768","gte":"4096"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.softlockup_panic

`bool` · optional (explicit presence)

kernel.softlockup_panic: panics the kernel on a soft lockup.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sysrq

`int64` · optional (explicit presence)

kernel.sysrq: allowed SysRq functions as a bitmask (0-511).

- rule: {"int64":{"lte":"511","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.kernel.yama.ptrace_scope

`int64` · optional (explicit presence)

kernel.yama.ptrace_scope: who may ptrace (0 classic, 1 restricted, 2
admin only, 3 none).

- rule: {"int64":{"lte":"3","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_poll

`int64` · optional (explicit presence)

net.core.busy_poll: busy-poll microseconds for poll and select.

- rule: {"int64":{"lte":"2147483647","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_read

`int64` · optional (explicit presence)

net.core.busy_read: busy-poll microseconds for socket reads.

- rule: {"int64":{"lte":"2147483647","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.netdev_max_backlog

`int64` · optional (explicit presence)

net.core.netdev_max_backlog: queued input packets per interface.

- rule: {"int64":{"lte":"2147483647","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.optmem_max

`int64` · optional (explicit presence)

net.core.optmem_max: ancillary buffer bytes per socket.

- rule: {"int64":{"lte":"2147483647","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_default

`int64` · optional (explicit presence)

net.core.rmem_default: default socket receive buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"2304"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_max

`int64` · optional (explicit presence)

net.core.rmem_max: largest socket receive buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"2304"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.somaxconn

`int64` · optional (explicit presence)

net.core.somaxconn: the listen() backlog limit (128 or more).

- rule: {"int64":{"lte":"2147483647","gte":"128"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_default

`int64` · optional (explicit presence)

net.core.wmem_default: default socket send buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"4608"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_max

`int64` · optional (explicit presence)

net.core.wmem_max: largest socket send buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"4608"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh1

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh1: neighbor cache floor (0-262144).

- rule: {"int64":{"lte":"262144","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh2

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh2: neighbor cache soft ceiling
(512-524288).

- rule: {"int64":{"lte":"524288","gte":"512"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh3

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh3: neighbor cache hard ceiling
(1024-1048576).

- rule: {"int64":{"lte":"1048576","gte":"1024"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_congestion_control

`string` · optional (explicit presence)

net.ipv4.tcp_congestion_control: the congestion control algorithm for new
connections (reno is always available). Not on Autopilot classes.

- rule: {"string":{"maxLen":"10","pattern":"^[a-zA-Z0-9_]+$"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_ecn

`int64` · optional (explicit presence)

net.ipv4.tcp_ecn: Explicit Congestion Notification (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_orphans

`int64` · optional (explicit presence)

net.ipv4.tcp_max_orphans: orphaned sockets kept (16384-262144).

- rule: {"int64":{"lte":"262144","gte":"16384"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_tw_buckets

`int64` · optional (explicit presence)

net.ipv4.tcp_max_tw_buckets: TIME-WAIT sockets kept (4096 or more).

- rule: {"int64":{"lte":"2147483647","gte":"4096"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_mtu_probing

`int64` · optional (explicit presence)

net.ipv4.tcp_mtu_probing: path MTU discovery (0 off, 1 on black hole, 2
always).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_rmem

`string` · optional (explicit presence)

net.ipv4.tcp_rmem: TCP receive buffer min, default and max bytes, space
separated ("4096 87380 6291456").

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_syn_retries

`int64` · optional (explicit presence)

net.ipv4.tcp_syn_retries: SYN retransmits for an outgoing connection
(1-127).

- rule: {"int64":{"lte":"127","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_tw_reuse

`int64` · optional (explicit presence)

net.ipv4.tcp_tw_reuse: reuse TIME-WAIT sockets when safe (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_wmem

`string` · optional (explicit presence)

net.ipv4.tcp_wmem: TCP send buffer min, default and max bytes, space
separated ("4096 16384 4194304").

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.all.disable_ipv6

`bool` · optional (explicit presence)

net.ipv6.conf.all.disable_ipv6: disables IPv6 on every interface.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.default.disable_ipv6

`bool` · optional (explicit presence)

net.ipv6.conf.default.disable_ipv6: disables IPv6 on new interfaces.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_acct

`bool` · optional (explicit presence)

net.netfilter.nf_conntrack_acct: connection tracking flow accounting.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_buckets

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_buckets: connection tracking hash table size
(65536-524288).

- rule: {"int64":{"lte":"524288","gte":"65536"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_max

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_max: connection tracking table size
(65536-4194304); raise it for nodes behind heavy NAT or many short
connections.

- rule: {"int64":{"lte":"4194304","gte":"65536"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_close_wait

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_close_wait: seconds a CLOSE_WAIT
connection stays tracked (60-3600).

- rule: {"int64":{"lte":"3600","gte":"60"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_established

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_established: seconds an idle
established connection stays tracked (600-86400).

- rule: {"int64":{"lte":"86400","gte":"600"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_time_wait

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_time_wait: seconds a TIME_WAIT
connection stays tracked (1-600).

- rule: {"int64":{"lte":"600","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_bytes

`int64` · optional (explicit presence)

vm.dirty_background_bytes: dirty memory bytes that start background
writeback; set it or vm.dirty_background_ratio, not both.

- rule: {"int64":{"lte":"68719476736","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_ratio

`int64` · optional (explicit presence)

vm.dirty_background_ratio: percent of memory (1-100) dirty before
background writeback; keep it below vm.dirty_ratio.

- rule: {"int64":{"lte":"100","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_bytes

`int64` · optional (explicit presence)

vm.dirty_bytes: dirty memory bytes at which writers write back
themselves; set it or vm.dirty_ratio, not both.

- rule: {"int64":{"lte":"68719476736","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_expire_centisecs

`int64` · optional (explicit presence)

vm.dirty_expire_centisecs: how old dirty data gets before it is written
(0-6000 hundredths of a second).

- rule: {"int64":{"lte":"6000","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_ratio

`int64` · optional (explicit presence)

vm.dirty_ratio: percent of memory (1-100) dirty before writers block to
write back; keep it above vm.dirty_background_ratio.

- rule: {"int64":{"lte":"100","gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_writeback_centisecs

`int64` · optional (explicit presence)

vm.dirty_writeback_centisecs: how often flusher threads wake (0-1000
hundredths of a second).

- rule: {"int64":{"lte":"1000","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.max_map_count

`int64` · optional (explicit presence)

vm.max_map_count: memory map areas per process (65536 or more);
Elasticsearch and OpenSearch need at least 262144.

- rule: {"int64":{"lte":"2147483647","gte":"65536"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.min_free_kbytes

`int64` · optional (explicit presence)

vm.min_free_kbytes: memory the kernel keeps free (67584-1048576 KB).

- rule: {"int64":{"lte":"1048576","gte":"67584"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_memory

`int64` · optional (explicit presence)

vm.overcommit_memory: 0 heuristic, 1 always overcommit, 2 strict (swap
plus vm.overcommit_ratio percent of RAM).

- rule: {"int64":{"in":["0","1","2"]}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_ratio

`int64` · optional (explicit presence)

vm.overcommit_ratio: percent of RAM (0-100) committable under strict
overcommit.

- rule: {"int64":{"lte":"100","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.swappiness

`int64` · optional (explicit presence)

vm.swappiness: preference for swapping over dropping page cache (0-200).

- rule: {"int64":{"lte":"200","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.vfs_cache_pressure

`int64` · optional (explicit presence)

vm.vfs_cache_pressure: preference for reclaiming dentry and inode caches
(0-100).

- rule: {"int64":{"lte":"100","gte":"0"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.sysctls.vm.watermark_scale_factor

`int64` · optional (explicit presence)

vm.watermark_scale_factor: how early kswapd starts reclaiming (10-3000).

- rule: {"int64":{"lte":"3000","gte":"10"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig

`KubernetesGkeComputeClassHugepageConfig`

Pre-allocated huge pages.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size2m

`int64` · optional (explicit presence)

2 MB huge pages to allocate (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size1g

`int64` · optional (explicit presence)

1 GB huge pages to allocate (1 or more), for large in-memory databases.

- rule: {"int64":{"gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.transparentHugepageEnabled

`string` · optional (explicit presence)

Transparent huge pages for anonymous memory:
TRANSPARENT_HUGEPAGE_ENABLED_ALWAYS, _MADVISE (the kernel default),
_NEVER, or _UNSPECIFIED (GKE leaves the kernel setting). Requires GKE
1.34.1-gke.3084001.

- rule: {"string":{"in":["TRANSPARENT_HUGEPAGE_ENABLED_ALWAYS","TRANSPARENT_HUGEPAGE_ENABLED_MADVISE","TRANSPARENT_HUGEPAGE_ENABLED_NEVER","TRANSPARENT_HUGEPAGE_ENABLED_UNSPECIFIED"]}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.transparentHugepageDefrag

`string` · optional (explicit presence)

Transparent huge page defragmentation: how hard an allocation tries to
get a huge page (ALWAYS stalls to compact, DEFER compacts in the
background, the MADVISE variants limit it to madvise regions, NEVER
never compacts). Requires GKE 1.34.1-gke.3084001.

- rule: {"string":{"in":["TRANSPARENT_HUGEPAGE_DEFRAG_ALWAYS","TRANSPARENT_HUGEPAGE_DEFRAG_DEFER","TRANSPARENT_HUGEPAGE_DEFRAG_DEFER_WITH_MADVISE","TRANSPARENT_HUGEPAGE_DEFRAG_MADVISE","TRANSPARENT_HUGEPAGE_DEFRAG_NEVER","TRANSPARENT_HUGEPAGE_DEFRAG_UNSPECIFIED"]}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig

`KubernetesGkeComputeClassSwapConfig`

Swap memory on the nodes.

- rule: only one of boot_disk_profile, ephemeral_local_ssd_profile or dedicated_local_ssd_profile may be set

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.enabled

`bool` · optional (explicit presence)

Enables swap. Upstream default false.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile

`KubernetesGkeComputeClassBootDiskProfile`

Swap on the boot disk.

- rule: only one of swap_size_gib or swap_size_percent may be set

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib

`int64` · optional (explicit presence)

Swap size in GiB (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage (1-50) of the boot disk.

- rule: {"int32":{"lte":50,"gte":1}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile

`KubernetesGkeComputeClassEphemeralLocalSsdProfile`

Swap on the local SSDs shared with ephemeral storage.

- rule: only one of swap_size_gib or swap_size_percent may be set

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib

`int64` · optional (explicit presence)

Swap size in GiB (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage (1-80) of the ephemeral local SSDs.

- rule: {"int32":{"lte":80,"gte":1}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile

`KubernetesGkeComputeClassDedicatedLocalSsdProfile`

Swap on dedicated local NVMe SSDs.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount

`int64` · optional (explicit presence)

Local NVMe SSDs attached for swap (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig

`KubernetesGkeComputeClassEncryptionConfig`

Swap encryption; unset, swap is encrypted.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled

`bool` · optional (explicit presence)

True turns swap encryption off; swap is encrypted by default.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler

`KubernetesGkeComputeClassDiskIoScheduler`

Disk IO schedulers. Requires GKE 1.36.2-gke.1368000.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeSystemIoScheduler

`string` · optional (explicit presence)

For the boot disk or ephemeral local SSD that runs system workloads:
mq-deadline, bfq, kyber or none.

- rule: {"string":{"in":["mq-deadline","bfq","kyber","none"]}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeAttachedDiskIoScheduler

`string` · optional (explicit presence)

For attached disks: mq-deadline, bfq, kyber or none.

- rule: {"string":{"in":["mq-deadline","bfq","kyber","none"]}}

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.nodeVfioConfig

`KubernetesGkeComputeClassNodeVfioConfig`

VFIO settings for direct device access (TPUs, GPUs). Requires GKE
1.36.2-gke.1368000.

### spec.priorities[].nodeSystemConfig.linuxNodeConfig.nodeVfioConfig.dmaEntryLimit

`int32` · optional (explicit presence)

DMA entries (65535-4194304) the VFIO IOMMU type 1 driver may map per
container (vfio_iommu_type1 dma_entry_limit); raise it for devices that
pin large memory regions.

- rule: {"int32":{"lte":4194304,"gte":65535}}

### spec.priorities[].instanceMetadata

`map<string, string>`

Custom Compute Engine metadata on the rule's VMs; overrides conflicting
keys of `node_pool_config.instance_metadata`. At most 32 pairs across
the class, keys of letters, digits, hyphens and underscores (under 128
characters), values up to 32,768 characters, and never one of GKE's
reserved keys (kube-env, startup-script, user-data, cluster-name, ...).
Requires GKE 1.36.2-gke.1498000.

- rule: instance_metadata cannot set a reserved GKE or Compute Engine metadata key
- rule: {"map":{"keys":{"string":{"maxLen":"127","pattern":"^[a-zA-Z0-9_-]+$"}},"values":{"string":{"maxLen":"32768"}}}}

### spec.priorities[].maxPodsPerNode

`int32` · optional (explicit presence)

The most pods a node from this rule accommodates (8-256); sizes the
node's pod IP range. Requires GKE 1.32.1-gke.1159000.

- rule: {"int32":{"lte":256,"gte":8}}

### spec.priorities[].maxRunDurationSeconds

`int32` · optional (explicit presence)

How long, in seconds, a node from this rule may exist before GKE removes
it. Unset, nodes live indefinitely. Requires GKE 1.32.1-gke.1159000.

### spec.priorities[].capacityCheckWaitTimeSeconds

`int32` · optional (explicit presence)

How long, in seconds (1-86400), GKE keeps trying this rule before moving
to the next. Only with `flex_start` enabled or a `tpu`. Requires GKE
1.33.1-gke.1788000.

- rule: {"int32":{"lte":86400,"gte":1}}

### spec.priorities[].allocationStrategy

`string` · optional (explicit presence)

How GKE picks among the node pool options that satisfy this rule:
`lowest-cost` (cheapest at us-central1 list prices, counting likely
future scale-ups) or `fleet-efficiency`. Unset, the class's
`allocation_strategy_defaults` apply. Requires GKE 1.36.4-gke.1082000 or
1.35.8-gke.1439000.

- rule: {"string":{"in":["lowest-cost","fleet-efficiency"]}}

### spec.priorities[].priorityScore

`int32` · optional (explicit presence)

An explicit priority (1-1000, higher wins) instead of list order; rules
with the same score are treated as equals, at most three per score. Set
it on every rule or on none. Requires GKE 1.35.2-gke.1842000.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.priorities[].minimumCapacity

`KubernetesGkeComputeClassMinimumCapacity`

Nodes GKE keeps for this rule even with no pods pending; commits spend.
Requires a machine_type, gpu, tpu or Specific reservation on the rule.
Requires GKE 1.36.3-gke.1253000.

### spec.priorities[].minimumCapacity.targetNodeCount

`int32` · optional (explicit presence)

The fewest nodes kept present; when the count falls below it, the
cluster autoscaler provisions nodes ahead of demand.

### spec.priorities[].acceleratorNetworkProfile

`string` · optional (explicit presence)

Automated accelerator network provisioning: `auto`, `auto-<suffix>`, or
the name of a user-managed AcceleratorNetworkProfile. Required with
`gpu_direct` rdma. Requires GKE 1.35.2-gke.1842000.

- rule: {"string":{"maxLen":"63","pattern":"^[a-z]([a-z0-9-]*[a-z0-9])?$"}}

### spec.priorityDefaults

`KubernetesGkeComputeClassPriorityDefaults`

Defaults for the `location` and `node_system_config` of every priority
rule that does not set its own. They never apply to a rule that holds only
`nodepools`. Requires GKE 1.32.1-gke.1729000.

### spec.priorityDefaults.location

`KubernetesGkeComputeClassLocation`

Default zones and zone spread for every rule.

### spec.priorityDefaults.location.zones

`[]string` · required

The zones nodes may be created in. GKE refuses a zone outside the
cluster's region (or its node locations). Requires GKE
1.34.1-gke.3759000.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"minItems":"1"}}

### spec.priorityDefaults.location.zoneTypes

`[]string` · required

Sets of zones instead of a list: STANDARD (the region's core zones), AI
(zones built for AI hardware), CLUSTER_DEFAULT (the cluster's
auto-provisioning locations, or its locations). GKE refuses zone types in
`priority_defaults` when any rule consumes a Specific reservation (the
reservation fixes its own zones). Requires GKE 1.35.2-gke.1842000.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"minItems":"1","maxItems":"3","items":{"string":{"in":["STANDARD","AI","CLUSTER_DEFAULT"]}}}}

### spec.priorityDefaults.location.locationPolicy

`string` · optional (explicit presence)

How new nodes spread across the zones: BALANCED (even spread) or ANY
(wherever capacity is, which finds scarce machines and Spot capacity
sooner). Requires GKE 1.34.1-gke.2541000.

- rule: {"string":{"in":["ANY","BALANCED"]}}

### spec.priorityDefaults.nodeSystemConfig

`KubernetesGkeComputeClassNodeSystemConfig`

Default kubelet and Linux kernel settings for every rule.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig

`KubernetesGkeComputeClassKubeletConfig`

Kubelet settings.

- rule: image_gc_low_threshold_percent must be lower than image_gc_high_threshold_percent
- rule: image_gc_high_threshold_percent alone must be higher than 80, the default image_gc_low_threshold_percent
- rule: shutdown_grace_period_critical_pods_seconds requires shutdown_grace_period_seconds and must not exceed it

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuCfsQuota

`bool` · optional (explicit presence)

Enforces pods' CPU limits with the CFS quota. False ignores CPU limits,
for latency-sensitive pods that throttle, at the risk of a pod taking
more CPU than intended.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuCfsQuotaPeriod

`string` · optional (explicit presence)

The CFS quota period, between 1ms and 1s (for example 100ms).

- rule: {"string":{"pattern":"^([1-9][0-9]*)m?s$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.cpuManagerPolicy

`string` · optional (explicit presence)

none (the default) or static, which gives Guaranteed pods with integer
CPU requests exclusive cores.

- rule: {"string":{"in":["none","static"]}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.podPidsLimit

`int64` · optional (explicit presence)

The most process IDs a pod may use (1024-4194304).

- rule: {"int64":{"lte":"4194304","gte":"1024"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.containerLogMaxSize

`string` · optional (explicit presence)

The size a container log reaches before it is rotated, 10Mi-500Mi (for
example 100Mi). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(Ki|Mi|Gi))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.containerLogMaxFiles

`int64` · optional (explicit presence)

How many rotated log files a container keeps (2-10). Requires GKE
1.33.4-gke.1036000.

- rule: {"int64":{"lte":"10","gte":"2"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageGcHighThresholdPercent

`int64` · optional (explicit presence)

Disk usage (11-85 percent) above which image garbage collection always
runs. Requires GKE 1.33.4-gke.1036000.

- rule: {"int64":{"lte":"85","gte":"11"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageGcLowThresholdPercent

`int64` · optional (explicit presence)

Disk usage (10-84 percent) image garbage collection frees down to.
Requires GKE 1.33.4-gke.1036000.

- rule: {"int64":{"lte":"84","gte":"10"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageMinimumGcAge

`string` · optional (explicit presence)

The youngest an unused image may be when collected, up to 2m (for
example 1m). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.imageMaximumGcAge

`string` · optional (explicit presence)

The longest an image may stay unused before collection (for example
2h45m). Requires GKE 1.33.4-gke.1036000.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.allowedUnsafeSysctls

`[]string`

Unsafe sysctls or patterns (ending in `*`) pods may set, from the groups
kernel.shm*, kernel.msg*, kernel.sem, fs.mqueue.* and net.*. Up to 100.
Requires GKE 1.33.4-gke.1036000.

- rule: {"repeated":{"maxItems":"100","items":{"string":{"minLen":"1","maxLen":"253","pattern":"^([a-z0-9]([-_a-z0-9]*[a-z0-9])?[./])*([a-z0-9][-_a-z0-9]*)?[a-z0-9*]$"}}}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.maxParallelImagePulls

`int64` · optional (explicit presence)

Image pulls the kubelet runs in parallel (2-5). Requires GKE
1.34.1-gke.3084001.

- rule: {"int64":{"lte":"5","gte":"2"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.singleProcessOOMKill

`bool` · optional (explicit presence)

OOM-kills a container's processes one at a time instead of as a group.
Requires GKE 1.34.1-gke.3084001.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds

`int64` · optional (explicit presence)

The longest grace period (0-300 seconds) given to pods evicted by a soft
eviction threshold.

- rule: {"int64":{"lte":"300","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft

`KubernetesGkeComputeClassEvictionSoft`

Soft eviction thresholds.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.memoryAvailable

`string` · optional (explicit presence)

memory.available threshold as a quantity (for example 200Mi), above the
100Mi hard threshold and under half the machine's memory.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ki|Mi|Gi)$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.nodefsAvailable

`string` · optional (explicit presence)

nodefs.available threshold, 10%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.nodefsInodesFree

`string` · optional (explicit presence)

nodefs.inodesFree threshold, 5%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.imagefsAvailable

`string` · optional (explicit presence)

imagefs.available threshold, 15%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.imagefsInodesFree

`string` · optional (explicit presence)

imagefs.inodesFree threshold, 5%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoft.pidAvailable

`string` · optional (explicit presence)

pid.available threshold, 10%-50%.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod

`KubernetesGkeComputeClassEvictionSoftGracePeriod`

Grace periods for the soft eviction thresholds.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable

`string` · optional (explicit presence)

For memory.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable

`string` · optional (explicit presence)

For nodefs.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree

`string` · optional (explicit presence)

For nodefs.inodesFree.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable

`string` · optional (explicit presence)

For imagefs.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree

`string` · optional (explicit presence)

For imagefs.inodesFree.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable

`string` · optional (explicit presence)

For pid.available.

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim

`KubernetesGkeComputeClassEvictionMinimumReclaim`

The least the kubelet reclaims once an eviction starts.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable

`string` · optional (explicit presence)

For memory.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable

`string` · optional (explicit presence)

For nodefs.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree

`string` · optional (explicit presence)

For nodefs.inodesFree.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable

`string` · optional (explicit presence)

For imagefs.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree

`string` · optional (explicit presence)

For imagefs.inodesFree.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable

`string` · optional (explicit presence)

For pid.available.

- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?%$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.crashLoopBackOff

`KubernetesGkeComputeClassCrashLoopBackOff`

Container restart back-off.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod

`string` · optional (explicit presence)

The longest the restart back-off grows to, under 5m (for example 30s).

- rule: {"string":{"pattern":"^([0-9]+([.][0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.memoryManager

`KubernetesGkeComputeClassMemoryManager`

The kubelet Memory Manager. Requires GKE 1.36.2-gke.1368000.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.memoryManager.policy

`string` · optional (explicit presence)

None (the default, no alignment) or Static, which gives Guaranteed pods
memory from a single NUMA node.

- rule: {"string":{"in":["None","Static"]}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager

`KubernetesGkeComputeClassTopologyManager`

The kubelet Topology Manager. Requires GKE 1.36.2-gke.1368000.

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager.policy

`string` · optional (explicit presence)

none (the default), best-effort, restricted, or single-numa-node (pods
that do not fit one NUMA node are refused).

- rule: {"string":{"in":["none","best-effort","restricted","single-numa-node"]}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.topologyManager.scope

`string` · optional (explicit presence)

container (the default, aligns each container) or pod (aligns the pod as
a whole).

- rule: {"string":{"in":["container","pod"]}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.shutdownGracePeriodSeconds

`int32` · optional (explicit presence)

Seconds (0, 30 or 120) a node delays shutdown so pods terminate
gracefully. Spot rules only. Requires GKE control plane
1.36.0-gke.3204000.

- rule: {"int32":{"in":[0,30,120]}}

### spec.priorityDefaults.nodeSystemConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds

`int32` · optional (explicit presence)

Seconds (0-120, at most shutdown_grace_period_seconds) reserved for
critical pods during that shutdown. Requires GKE control plane
1.36.0-gke.3204000.

- rule: {"int32":{"lte":120,"gte":0}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig

`KubernetesGkeComputeClassLinuxNodeConfig`

Linux kernel settings.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls

`KubernetesGkeComputeClassSysctls`

Kernel sysctls GKE allows on nodes.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.aio-max-nr

`int64` · optional (explicit presence)

fs.aio-max-nr: system-wide asynchronous IO requests (65536-4194304).

- rule: {"int64":{"lte":"4194304","gte":"65536"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.file-max

`int64` · optional (explicit presence)

fs.file-max: file handles the kernel allocates (104857-67108864).

- rule: {"int64":{"lte":"67108864","gte":"104857"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_instances

`int64` · optional (explicit presence)

fs.inotify.max_user_instances: inotify instances per user (8192-1048576).

- rule: {"int64":{"lte":"1048576","gte":"8192"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.inotify.max_user_watches

`int64` · optional (explicit presence)

fs.inotify.max_user_watches: inotify watches per user (8192-1048576).

- rule: {"int64":{"lte":"1048576","gte":"8192"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.fs.nr_open

`int64` · optional (explicit presence)

fs.nr_open: file descriptors per process (1048576-2147483584).

- rule: {"int64":{"lte":"2147483584","gte":"1048576"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.core_pattern

`string` · optional (explicit presence)

kernel.core_pattern: the core dump path, absolute only (no pipes).

- rule: {"string":{"maxLen":"128","pattern":"^/[a-zA-Z0-9/._%-]+$"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.dmesg_restrict

`bool` · optional (explicit presence)

kernel.dmesg_restrict: keeps unprivileged users from reading the kernel
log.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxbytes

`int64` · optional (explicit presence)

kernel.keys.maxbytes: key payload bytes per non-root user
(20000-2097152).

- rule: {"int64":{"lte":"2097152","gte":"20000"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.keys.maxkeys

`int64` · optional (explicit presence)

kernel.keys.maxkeys: keys per non-root user (200-1048576).

- rule: {"int64":{"lte":"1048576","gte":"200"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.kptr_restrict

`int64` · optional (explicit presence)

kernel.kptr_restrict: hides kernel addresses (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.perf_event_paranoid

`int64` · optional (explicit presence)

kernel.perf_event_paranoid: unprivileged use of perf events (-1-3).

- rule: {"int64":{"lte":"3","gte":"-1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sched_rt_runtime_us

`int64` · optional (explicit presence)

kernel.sched_rt_runtime_us: real-time scheduling budget (-1-1000000).

- rule: {"int64":{"lte":"1000000","gte":"-1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmall

`string` · required · optional (explicit presence)

kernel.shmall: total shared memory pages, as a decimal string (up to 20
digits; the range exceeds a 64-bit signed integer).

- rule: {"string":{"minLen":"1","maxLen":"20","pattern":"^([0-9]+)$"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmax

`string` · required · optional (explicit presence)

kernel.shmmax: largest shared memory segment in bytes, as a decimal
string (up to 20 digits).

- rule: {"string":{"minLen":"1","maxLen":"20","pattern":"^([0-9]+)$"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.shmmni

`int64` · optional (explicit presence)

kernel.shmmni: shared memory segments system-wide (4096-32768).

- rule: {"int64":{"lte":"32768","gte":"4096"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.softlockup_panic

`bool` · optional (explicit presence)

kernel.softlockup_panic: panics the kernel on a soft lockup.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.sysrq

`int64` · optional (explicit presence)

kernel.sysrq: allowed SysRq functions as a bitmask (0-511).

- rule: {"int64":{"lte":"511","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.kernel.yama.ptrace_scope

`int64` · optional (explicit presence)

kernel.yama.ptrace_scope: who may ptrace (0 classic, 1 restricted, 2
admin only, 3 none).

- rule: {"int64":{"lte":"3","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_poll

`int64` · optional (explicit presence)

net.core.busy_poll: busy-poll microseconds for poll and select.

- rule: {"int64":{"lte":"2147483647","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.busy_read

`int64` · optional (explicit presence)

net.core.busy_read: busy-poll microseconds for socket reads.

- rule: {"int64":{"lte":"2147483647","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.netdev_max_backlog

`int64` · optional (explicit presence)

net.core.netdev_max_backlog: queued input packets per interface.

- rule: {"int64":{"lte":"2147483647","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.optmem_max

`int64` · optional (explicit presence)

net.core.optmem_max: ancillary buffer bytes per socket.

- rule: {"int64":{"lte":"2147483647","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_default

`int64` · optional (explicit presence)

net.core.rmem_default: default socket receive buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"2304"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.rmem_max

`int64` · optional (explicit presence)

net.core.rmem_max: largest socket receive buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"2304"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.somaxconn

`int64` · optional (explicit presence)

net.core.somaxconn: the listen() backlog limit (128 or more).

- rule: {"int64":{"lte":"2147483647","gte":"128"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_default

`int64` · optional (explicit presence)

net.core.wmem_default: default socket send buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"4608"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.core.wmem_max

`int64` · optional (explicit presence)

net.core.wmem_max: largest socket send buffer bytes.

- rule: {"int64":{"lte":"2147483647","gte":"4608"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh1

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh1: neighbor cache floor (0-262144).

- rule: {"int64":{"lte":"262144","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh2

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh2: neighbor cache soft ceiling
(512-524288).

- rule: {"int64":{"lte":"524288","gte":"512"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.neigh.default.gc_thresh3

`int64` · optional (explicit presence)

net.ipv4.neigh.default.gc_thresh3: neighbor cache hard ceiling
(1024-1048576).

- rule: {"int64":{"lte":"1048576","gte":"1024"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_congestion_control

`string` · optional (explicit presence)

net.ipv4.tcp_congestion_control: the congestion control algorithm for new
connections (reno is always available). Not on Autopilot classes.

- rule: {"string":{"maxLen":"10","pattern":"^[a-zA-Z0-9_]+$"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_ecn

`int64` · optional (explicit presence)

net.ipv4.tcp_ecn: Explicit Congestion Notification (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_orphans

`int64` · optional (explicit presence)

net.ipv4.tcp_max_orphans: orphaned sockets kept (16384-262144).

- rule: {"int64":{"lte":"262144","gte":"16384"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_max_tw_buckets

`int64` · optional (explicit presence)

net.ipv4.tcp_max_tw_buckets: TIME-WAIT sockets kept (4096 or more).

- rule: {"int64":{"lte":"2147483647","gte":"4096"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_mtu_probing

`int64` · optional (explicit presence)

net.ipv4.tcp_mtu_probing: path MTU discovery (0 off, 1 on black hole, 2
always).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_rmem

`string` · optional (explicit presence)

net.ipv4.tcp_rmem: TCP receive buffer min, default and max bytes, space
separated ("4096 87380 6291456").

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_syn_retries

`int64` · optional (explicit presence)

net.ipv4.tcp_syn_retries: SYN retransmits for an outgoing connection
(1-127).

- rule: {"int64":{"lte":"127","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_tw_reuse

`int64` · optional (explicit presence)

net.ipv4.tcp_tw_reuse: reuse TIME-WAIT sockets when safe (0-2).

- rule: {"int64":{"lte":"2","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv4.tcp_wmem

`string` · optional (explicit presence)

net.ipv4.tcp_wmem: TCP send buffer min, default and max bytes, space
separated ("4096 16384 4194304").

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.all.disable_ipv6

`bool` · optional (explicit presence)

net.ipv6.conf.all.disable_ipv6: disables IPv6 on every interface.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.ipv6.conf.default.disable_ipv6

`bool` · optional (explicit presence)

net.ipv6.conf.default.disable_ipv6: disables IPv6 on new interfaces.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_acct

`bool` · optional (explicit presence)

net.netfilter.nf_conntrack_acct: connection tracking flow accounting.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_buckets

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_buckets: connection tracking hash table size
(65536-524288).

- rule: {"int64":{"lte":"524288","gte":"65536"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_max

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_max: connection tracking table size
(65536-4194304); raise it for nodes behind heavy NAT or many short
connections.

- rule: {"int64":{"lte":"4194304","gte":"65536"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_close_wait

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_close_wait: seconds a CLOSE_WAIT
connection stays tracked (60-3600).

- rule: {"int64":{"lte":"3600","gte":"60"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_established

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_established: seconds an idle
established connection stays tracked (600-86400).

- rule: {"int64":{"lte":"86400","gte":"600"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.net.netfilter.nf_conntrack_tcp_timeout_time_wait

`int64` · optional (explicit presence)

net.netfilter.nf_conntrack_tcp_timeout_time_wait: seconds a TIME_WAIT
connection stays tracked (1-600).

- rule: {"int64":{"lte":"600","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_bytes

`int64` · optional (explicit presence)

vm.dirty_background_bytes: dirty memory bytes that start background
writeback; set it or vm.dirty_background_ratio, not both.

- rule: {"int64":{"lte":"68719476736","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_background_ratio

`int64` · optional (explicit presence)

vm.dirty_background_ratio: percent of memory (1-100) dirty before
background writeback; keep it below vm.dirty_ratio.

- rule: {"int64":{"lte":"100","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_bytes

`int64` · optional (explicit presence)

vm.dirty_bytes: dirty memory bytes at which writers write back
themselves; set it or vm.dirty_ratio, not both.

- rule: {"int64":{"lte":"68719476736","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_expire_centisecs

`int64` · optional (explicit presence)

vm.dirty_expire_centisecs: how old dirty data gets before it is written
(0-6000 hundredths of a second).

- rule: {"int64":{"lte":"6000","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_ratio

`int64` · optional (explicit presence)

vm.dirty_ratio: percent of memory (1-100) dirty before writers block to
write back; keep it above vm.dirty_background_ratio.

- rule: {"int64":{"lte":"100","gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.dirty_writeback_centisecs

`int64` · optional (explicit presence)

vm.dirty_writeback_centisecs: how often flusher threads wake (0-1000
hundredths of a second).

- rule: {"int64":{"lte":"1000","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.max_map_count

`int64` · optional (explicit presence)

vm.max_map_count: memory map areas per process (65536 or more);
Elasticsearch and OpenSearch need at least 262144.

- rule: {"int64":{"lte":"2147483647","gte":"65536"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.min_free_kbytes

`int64` · optional (explicit presence)

vm.min_free_kbytes: memory the kernel keeps free (67584-1048576 KB).

- rule: {"int64":{"lte":"1048576","gte":"67584"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_memory

`int64` · optional (explicit presence)

vm.overcommit_memory: 0 heuristic, 1 always overcommit, 2 strict (swap
plus vm.overcommit_ratio percent of RAM).

- rule: {"int64":{"in":["0","1","2"]}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.overcommit_ratio

`int64` · optional (explicit presence)

vm.overcommit_ratio: percent of RAM (0-100) committable under strict
overcommit.

- rule: {"int64":{"lte":"100","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.swappiness

`int64` · optional (explicit presence)

vm.swappiness: preference for swapping over dropping page cache (0-200).

- rule: {"int64":{"lte":"200","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.vfs_cache_pressure

`int64` · optional (explicit presence)

vm.vfs_cache_pressure: preference for reclaiming dentry and inode caches
(0-100).

- rule: {"int64":{"lte":"100","gte":"0"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.sysctls.vm.watermark_scale_factor

`int64` · optional (explicit presence)

vm.watermark_scale_factor: how early kswapd starts reclaiming (10-3000).

- rule: {"int64":{"lte":"3000","gte":"10"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig

`KubernetesGkeComputeClassHugepageConfig`

Pre-allocated huge pages.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size2m

`int64` · optional (explicit presence)

2 MB huge pages to allocate (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.hugepageConfig.hugepage_size1g

`int64` · optional (explicit presence)

1 GB huge pages to allocate (1 or more), for large in-memory databases.

- rule: {"int64":{"gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.transparentHugepageEnabled

`string` · optional (explicit presence)

Transparent huge pages for anonymous memory:
TRANSPARENT_HUGEPAGE_ENABLED_ALWAYS, _MADVISE (the kernel default),
_NEVER, or _UNSPECIFIED (GKE leaves the kernel setting). Requires GKE
1.34.1-gke.3084001.

- rule: {"string":{"in":["TRANSPARENT_HUGEPAGE_ENABLED_ALWAYS","TRANSPARENT_HUGEPAGE_ENABLED_MADVISE","TRANSPARENT_HUGEPAGE_ENABLED_NEVER","TRANSPARENT_HUGEPAGE_ENABLED_UNSPECIFIED"]}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.transparentHugepageDefrag

`string` · optional (explicit presence)

Transparent huge page defragmentation: how hard an allocation tries to
get a huge page (ALWAYS stalls to compact, DEFER compacts in the
background, the MADVISE variants limit it to madvise regions, NEVER
never compacts). Requires GKE 1.34.1-gke.3084001.

- rule: {"string":{"in":["TRANSPARENT_HUGEPAGE_DEFRAG_ALWAYS","TRANSPARENT_HUGEPAGE_DEFRAG_DEFER","TRANSPARENT_HUGEPAGE_DEFRAG_DEFER_WITH_MADVISE","TRANSPARENT_HUGEPAGE_DEFRAG_MADVISE","TRANSPARENT_HUGEPAGE_DEFRAG_NEVER","TRANSPARENT_HUGEPAGE_DEFRAG_UNSPECIFIED"]}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig

`KubernetesGkeComputeClassSwapConfig`

Swap memory on the nodes.

- rule: only one of boot_disk_profile, ephemeral_local_ssd_profile or dedicated_local_ssd_profile may be set

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.enabled

`bool` · optional (explicit presence)

Enables swap. Upstream default false.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile

`KubernetesGkeComputeClassBootDiskProfile`

Swap on the boot disk.

- rule: only one of swap_size_gib or swap_size_percent may be set

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib

`int64` · optional (explicit presence)

Swap size in GiB (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage (1-50) of the boot disk.

- rule: {"int32":{"lte":50,"gte":1}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile

`KubernetesGkeComputeClassEphemeralLocalSsdProfile`

Swap on the local SSDs shared with ephemeral storage.

- rule: only one of swap_size_gib or swap_size_percent may be set

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib

`int64` · optional (explicit presence)

Swap size in GiB (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage (1-80) of the ephemeral local SSDs.

- rule: {"int32":{"lte":80,"gte":1}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile

`KubernetesGkeComputeClassDedicatedLocalSsdProfile`

Swap on dedicated local NVMe SSDs.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount

`int64` · optional (explicit presence)

Local NVMe SSDs attached for swap (1 or more).

- rule: {"int64":{"gte":"1"}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig

`KubernetesGkeComputeClassEncryptionConfig`

Swap encryption; unset, swap is encrypted.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled

`bool` · optional (explicit presence)

True turns swap encryption off; swap is encrypted by default.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler

`KubernetesGkeComputeClassDiskIoScheduler`

Disk IO schedulers. Requires GKE 1.36.2-gke.1368000.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeSystemIoScheduler

`string` · optional (explicit presence)

For the boot disk or ephemeral local SSD that runs system workloads:
mq-deadline, bfq, kyber or none.

- rule: {"string":{"in":["mq-deadline","bfq","kyber","none"]}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.diskIoScheduler.nodeAttachedDiskIoScheduler

`string` · optional (explicit presence)

For attached disks: mq-deadline, bfq, kyber or none.

- rule: {"string":{"in":["mq-deadline","bfq","kyber","none"]}}

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.nodeVfioConfig

`KubernetesGkeComputeClassNodeVfioConfig`

VFIO settings for direct device access (TPUs, GPUs). Requires GKE
1.36.2-gke.1368000.

### spec.priorityDefaults.nodeSystemConfig.linuxNodeConfig.nodeVfioConfig.dmaEntryLimit

`int32` · optional (explicit presence)

DMA entries (65535-4194304) the VFIO IOMMU type 1 driver may map per
container (vfio_iommu_type1 dma_entry_limit); raise it for devices that
pin large memory regions.

- rule: {"int32":{"lte":4194304,"gte":65535}}

### spec.whenUnsatisfiable

`string` · optional (explicit presence)

What GKE does when no priority rule can be satisfied:
  - DoNotScaleUp (the default): no node is created and the pods stay
    pending until a rule can be met. The class never runs pods on a
    machine you did not list.
  - ScaleUpAnyway: GKE creates a node from its default machine series
    (E2 on Standard) so the pods run, on a machine the class never named.

Choose ScaleUpAnyway for a cluster-wide `default` class whose job is to
keep installs and controllers schedulable through a stockout; keep
DoNotScaleUp where the machine is a requirement (GPUs, confidential nodes,
licensed families). The default is written explicitly, so the class
behaves the same on every GKE version (GKE changed its own default in
1.33).

- default: `DoNotScaleUp`
- rule: {"string":{"in":["ScaleUpAnyway","DoNotScaleUp"]}}

### spec.nodePoolAutoCreation

`KubernetesGkeComputeClassNodePoolAutoCreation`

Whether GKE may create (and later delete) node pools for this class's
rules. Off (the upstream default), the class only places pods on existing
node pools that match a rule, so a family fallback needs it on.

### spec.nodePoolAutoCreation.enabled

`bool` · optional (explicit presence)

Lets GKE create node pools from the class's rules when no existing pool
fits, and delete them when they empty. Upstream default false. Cannot be
false on an Autopilot class.

### spec.nodePoolAutoCreation.shieldedInstanceConfig

`KubernetesGkeComputeClassShieldedInstanceConfig`

Shielded VM settings for the node pools GKE creates. Requires GKE
1.36.3-gke.1244000.

### spec.nodePoolAutoCreation.shieldedInstanceConfig.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Measures the boot sequence against a known-good baseline and reports
drift to Cloud Monitoring. Upstream default true.

### spec.nodePoolAutoCreation.shieldedInstanceConfig.enableSecureBoot

`bool` · optional (explicit presence)

Refuses to boot a node whose boot components are not signed. Upstream
default false; turn it on unless nodes load unsigned kernel modules
(some GPU drivers).

### spec.nodePoolConfig

`KubernetesGkeComputeClassNodePoolConfig`

Node pool configuration every node of the class carries: image, identity,
labels, taints, metadata server, sandbox, confidential computing. Pools
GKE auto-creates are built with it; an existing pool is used for the class
only when its configuration matches. Requires GKE 1.31.4-gke.1072000.

### spec.nodePoolConfig.imageType

`string` · optional (explicit presence)

The node image: cos_containerd (Container-Optimized OS, the default and
the only one Autopilot classes allow) or ubuntu_containerd. Requires GKE
1.32.4-gke.1198000.

- rule: {"string":{"in":["cos_containerd","ubuntu_containerd"]}}

### spec.nodePoolConfig.serviceAccount

`string | valueFrom`

The IAM service account the nodes run as. Use a dedicated least-privilege
account (roles/container.defaultNodeServiceAccount plus registry read),
never the Compute Engine default, which often holds project Editor.
Requires GKE 1.31.4-gke.1072000.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.nodePoolConfig.workloadMetadata

`string` · optional (explicit presence)

How node metadata is exposed to pods: GKE_METADATA (the GKE metadata
server, required for Workload Identity, and it keeps pods from reading
the node's own credentials) or GCE_METADATA. Requires GKE
1.36.2-gke.1368000.

- rule: {"string":{"in":["GCE_METADATA","GKE_METADATA"]}}

### spec.nodePoolConfig.nodeLabels

`map<string, string>`

Kubernetes labels on every node of the class. Distinct from GCE labels.

### spec.nodePoolConfig.taints

`[]KubernetesGkeComputeClassTaint`

Kubernetes taints on every node of the class. Up to 100. Pods tolerate
the class's own taint automatically; a taint here fences the nodes from
pods that do not tolerate it too.

- rule: {"repeated":{"maxItems":"100"}}

### spec.nodePoolConfig.taints[].key

`string` · required

The taint key, in Kubernetes label-key syntax.

- rule: {"string":{"minLen":"1","maxLen":"320"}}

### spec.nodePoolConfig.taints[].value

`string` · optional (explicit presence)

The taint value.

- rule: {"string":{"maxLen":"63","pattern":"^([a-z0-9][-A-Za-z0-9_.]{1,61})?[A-Za-z0-9]$"}}

### spec.nodePoolConfig.taints[].effect

`string`

What happens to pods that do not tolerate it: NoSchedule,
PreferNoSchedule or NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.nodePoolConfig.taintConfig

`KubernetesGkeComputeClassTaintConfig`

How GKE taints ARM nodes: ARM (the default) taints arm64 nodes so only
pods that tolerate them land there; NONE applies no architecture taint.
Requires GKE 1.36.2-gke.1498000.

### spec.nodePoolConfig.taintConfig.architectureTaintBehavior

`string` · optional (explicit presence)

ARM (the default) taints arm64 nodes; NONE applies no architecture taint.

- rule: {"string":{"in":["NONE","ARM"]}}

### spec.nodePoolConfig.ipType

`string` · optional (explicit presence)

Whether nodes have public IP addresses (`public`) or only private ones
(`private`). Unset, the cluster's default applies. Matches the
`cloud.google.com/private-node` node selector. Requires GKE
1.35.0-gke.2398000.

- rule: {"string":{"in":["public","private"]}}

### spec.nodePoolConfig.confidentialNodeType

`string` · optional (explicit presence)

Confidential computing for every node: SEV and SEV_SNP (AMD) or TDX
(Intel). Each technology runs only on certain families, so every rule
must name a machine_family or machine_type that supports it. Requires GKE
1.33.3-gke.1392000.

- rule: {"string":{"in":["CONFIDENTIAL_INSTANCE_TYPE_UNSPECIFIED","SEV","SEV_SNP","TDX"]}}

### spec.nodePoolConfig.sandbox

`KubernetesGkeComputeClassSandbox`

A sandbox for untrusted workloads: gvisor (GKE Sandbox) or microvm.
Requires GKE 1.36.2-gke.1368000.

### spec.nodePoolConfig.sandbox.type

`string` · optional (explicit presence)

gvisor (GKE Sandbox) or microvm.

- rule: {"string":{"in":["gvisor","microvm"]}}

### spec.nodePoolConfig.imageStreaming

`KubernetesGkeComputeClassImageStreaming`

Image streaming: containers start before their whole image is pulled.
Requires GKE 1.34.1-gke.1279000.

### spec.nodePoolConfig.imageStreaming.enabled

`bool` · optional (explicit presence)

Streams images from Artifact Registry so containers start before the
whole image is pulled. Upstream default false.

### spec.nodePoolConfig.gvnic

`KubernetesGkeComputeClassGvnic`

Google Virtual NIC, for higher network bandwidth. Requires GKE
1.34.1-gke.3759000.

### spec.nodePoolConfig.gvnic.enabled

`bool` · optional (explicit presence)

Uses gVNIC, required for the highest network bandwidth tiers. Upstream
default false.

### spec.nodePoolConfig.autoRepair

`bool` · optional (explicit presence)

Whether GKE repairs unhealthy nodes. Requires GKE 1.34.0-gke.2201000.

### spec.nodePoolConfig.autoUpgrade

`bool` · optional (explicit presence)

Whether GKE upgrades the nodes with the cluster. Requires GKE
1.34.0-gke.2201000.

### spec.nodePoolConfig.maintenanceExclusion

`string` · optional (explicit presence)

A maintenance exclusion for the class's pools: UNTIL_END_OF_SUPPORT
holds them on their minor version until it reaches end of support.
Requires GKE 1.36.4-gke.1082000.

### spec.nodePoolConfig.resourceManagerTags

`[]KubernetesGkeComputeClassResourceManagerTag`

Up to five resource manager tags (purpose GCE_FIREWALL) attached to every
node, for tag-based firewall policies. The tags must already exist.
Requires GKE 1.34.1-gke.2541000.

- rule: {"repeated":{"maxItems":"5"}}

### spec.nodePoolConfig.resourceManagerTags[].key

`string` · required

The tag key.

- rule: {"string":{"minLen":"1"}}

### spec.nodePoolConfig.resourceManagerTags[].value

`string` · required

The tag value.

- rule: {"string":{"minLen":"1"}}

### spec.nodePoolConfig.loggingConfig

`KubernetesGkeComputeClassLoggingConfig`

The node logging agent: DEFAULT, or MAX_THROUGHPUT for nodes that write
high log volumes. Requires GKE 1.35.0-gke.2398000.

### spec.nodePoolConfig.loggingConfig.loggingVariantConfig

`KubernetesGkeComputeClassLoggingVariantConfig`

The logging agent variant.

### spec.nodePoolConfig.loggingConfig.loggingVariantConfig.variant

`string` · optional (explicit presence)

DEFAULT, or MAX_THROUGHPUT for nodes writing high log volumes (more
agent CPU and memory). Autopilot classes allow only DEFAULT.

- rule: {"string":{"in":["DEFAULT","MAX_THROUGHPUT"]}}

### spec.nodePoolConfig.instanceMetadata

`map<string, string>`

Custom Compute Engine metadata on every VM of the class. At most 32 pairs
across the class, keys of letters, digits, hyphens and underscores (under
128 characters), values up to 32,768 characters, and never one of GKE's
reserved keys. Requires GKE 1.36.2-gke.1498000.

- rule: instance_metadata cannot set a reserved GKE or Compute Engine metadata key
- rule: {"map":{"keys":{"string":{"maxLen":"127","pattern":"^[a-zA-Z0-9_-]+$"}},"values":{"string":{"maxLen":"32768"}}}}

### spec.nodePoolConfig.dra

`KubernetesGkeComputeClassDra`

Dynamic resource allocation settings for auto-provisioned nodes.
Requires GKE 1.35.2-gke.1842000.

### spec.nodePoolConfig.dra.networking

`KubernetesGkeComputeClassDraNetworking`

DRA networking for auto-provisioned nodes.

### spec.nodePoolConfig.dra.networking.enabled

`bool` · optional (explicit presence)

Enables DRA networking on the nodes. Upstream default false.

### spec.nodePoolConfig.workloadType

`string` · optional (explicit presence)

The workload SLO of a node pool group: HIGH_AVAILABILITY (Collection SLO,
for serving that needs most slices up all the time) or HIGH_THROUGHPUT
(Goodput SLO, for training that needs all slices up most of the time,
only with a `node_pool_group`). Requires GKE 1.32.2-gke.1359000.

- rule: {"string":{"in":["HIGH_AVAILABILITY","HIGH_THROUGHPUT"]}}

### spec.nodePoolGroup

`KubernetesGkeComputeClassNodePoolGroup`

Groups the class's node pools into one collection that shares a
configuration, such as a workload SLO for multi-host TPU or GPU slices
(`node_pool_config.workload_type`). Requires GKE 1.32.2-gke.1359000.

### spec.nodePoolGroup.name

`string` · required

The collection's name (for example MultiMIG).

- rule: {"string":{"minLen":"1"}}

### spec.activeMigration

`KubernetesGkeComputeClassActiveMigration`

Whether GKE moves running pods back to nodes from a higher-priority rule
once that capacity is available again: the way a class that fell back to
its third family during a stockout returns to its first.

### spec.activeMigration.optimizeRulePriority

`bool` · optional (explicit presence)

Moves pods to nodes from a higher-priority rule when that capacity is
available again, replacing the lower-priority nodes they ran on. Pods
are evicted to move, so it suits stateless workloads with
PodDisruptionBudgets. Upstream default false.

### spec.activeMigration.ensureAllDaemonSetPodsRunning

`bool` · optional (explicit presence)

Migrates node pools to larger machines when a DaemonSet's pods cannot be
scheduled on the existing nodes.

### spec.autoscalingPolicy

`KubernetesGkeComputeClassAutoscalingPolicy`

How quickly GKE removes nodes of the class that are no longer needed.

### spec.autoscalingPolicy.consolidationDelayMinutes

`int32` · optional (explicit presence)

How many minutes (1-1440) a node must be unneeded before GKE removes it.
Short values return capacity quickly; long ones avoid churn for bursty
load.

- rule: {"int32":{"lte":1440,"gte":1}}

### spec.autoscalingPolicy.consolidationThreshold

`int32` · optional (explicit presence)

The CPU and memory utilization (0-100 percent) below which a node becomes
a candidate for removal.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscalingPolicy.gpuConsolidationThreshold

`int32` · optional (explicit presence)

The GPU utilization (0-100 percent) below which a GPU node becomes a
candidate for removal; CPU and memory are then ignored. 0 or 100 makes
GKE consolidate nodes that do not use all their GPUs.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.allocationStrategyDefaults

`KubernetesGkeComputeClassAllocationStrategyDefaults`

The allocation strategy each provisioning model uses when a rule sets
none. Requires GKE 1.36.4-gke.1082000 or 1.35.8-gke.1439000.

### spec.allocationStrategyDefaults.onDemand

`string` · optional (explicit presence)

For on-demand rules (neither spot nor flex_start).

- rule: {"string":{"in":["lowest-cost","fleet-efficiency"]}}

### spec.allocationStrategyDefaults.spot

`string` · optional (explicit presence)

For spot rules.

- rule: {"string":{"in":["lowest-cost","fleet-efficiency"]}}

### spec.allocationStrategyDefaults.flexStart

`string` · optional (explicit presence)

For flex-start rules.

- rule: {"string":{"in":["lowest-cost","fleet-efficiency"]}}

### spec.minimumCapacity

`KubernetesGkeComputeClassMinimumCapacity`

Nodes GKE keeps for the class even with no pods pending: the one setting
here that commits spend before any workload asks for it. A target
requires every rule to name a machine_type, gpu, tpu or Specific
reservation. Requires GKE 1.36.3-gke.1253000.

### spec.minimumCapacity.targetNodeCount

`int32` · optional (explicit presence)

The fewest nodes kept present; when the count falls below it, the
cluster autoscaler provisions nodes ahead of demand.

### spec.autopilot

`KubernetesGkeComputeClassAutopilot`

Makes the class's nodes Autopilot-managed, on a Standard cluster too.
Fixed at creation: GKE refuses a change to it, or removing it, once set
(replace the class to change it). Requires GKE 1.34.1-gke.1829001.

### spec.autopilot.enabled

`bool` · optional (explicit presence)

Autopilot-managed nodes for the class. Fixed once the class exists.

## Validation Rules

- `spec.autopilot_forbids_nodepools`: a nodepools priority rule cannot be used when autopilot is enabled -- Autopilot manages its own node pools
- `spec.autopilot_requires_auto_creation`: node_pool_auto_creation cannot be disabled when autopilot is enabled
- `spec.autopilot_image_type`: only the cos_containerd image type can be used when autopilot is enabled
- `spec.autopilot_logging_variant`: only the DEFAULT logging variant can be used when autopilot is enabled
- `spec.workload_type_requires_node_pool_group`: without node_pool_group, node_pool_config.workload_type can only be HIGH_AVAILABILITY
- `spec.pod_family_requires_autopilot`: a pod_family priority rule can be used only when autopilot is enabled
- `spec.confidential_requires_machine`: with node_pool_config.confidential_node_type set, every priority rule must set machine_family or machine_type
- `spec.confidential_sev_families`: confidential_node_type SEV supports only the n2d, c2d, c3d, c4d and g4 families
- `spec.confidential_sev_g4`: confidential_node_type SEV on g4 supports only the g4-standard-48 machine type with the nvidia-rtx-pro-6000 GPU
- `spec.confidential_sev_snp_families`: confidential_node_type SEV_SNP supports only the n2d family
- `spec.confidential_tdx_families`: confidential_node_type TDX supports only c3 standard, c4 standard, a3-highgpu-1g and a4-highgpu-8g
- `spec.confidential_tdx_gpus`: confidential_node_type TDX allows only the nvidia-h100-80gb GPU on c3, c4 and a3, and the nvidia-b200 GPU on a4
- `spec.priority_score_all_or_none`: priority_score must be set on every priority rule or on none of them
- `spec.default_shutdown_grace_requires_spot`: priority_defaults may set kubelet shutdown_grace_period_seconds only when every priority rule is spot
- `spec.minimum_capacity_requires_machine`: a class-level minimum_capacity target requires every priority rule to name a machine_type, gpu, tpu, or Specific reservation
- `spec.autopilot_forbids_tcp_congestion_control`: the net.ipv4.tcp_congestion_control sysctl cannot be set when autopilot is enabled

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGkeComputeClass, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.compute_class_name` | `string` | Name of the created ComputeClass (the spec's `name`, its metadata.name in the cluster). A workload selects the class with the node selector `cloud.google.com/compute-class: <compute_class_name>`; a class named `default` applies to every pod that selects none when the cluster's default compute class is enabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.priorities[].nodepools` | GcpGkeNodePool | `status.outputs.node_pool_name` |
| `spec.priorities[].reservations.specific[].project` | GcpProject | `status.outputs.project_id` |
| `spec.priorities[].storage.bootDiskKMSKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.priorities[].storage.secondaryBootDisks[].project` | GcpProject | `status.outputs.project_id` |
| `spec.nodePoolConfig.serviceAccount` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)

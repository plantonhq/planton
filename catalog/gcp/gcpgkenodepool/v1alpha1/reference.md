# GcpGkeNodePool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGkeNodePoolSpec defines a GKE node pool (`google_container_node_pool`) —
a group of Compute Engine VMs with one shared configuration, attached to a
GcpGkeCluster.

Node pools are the compute half of the GKE composition boundary: the
cluster owns the control plane and cluster-wide configuration; every node
pool is a first-class resource with its own lifecycle, referencing the
cluster by its name output. Production clusters run several pools —
general-purpose on-demand, scale-to-zero Spot for fault-tolerant batch,
GPU pools for ML — each sized and tainted for its workloads. Autopilot
clusters take no node pools at all (GKE manages nodes).

The node pool inherits its project and location from the parent cluster;
both resolve by reference so a manifest never has to repeat what the
cluster already knows.

## Example

```yaml
# Development manifest for GcpGkeNodePool — exercises a broad slice of the
# spec (autoscaling, spot, taints, labels, upgrade settings, kubelet tuning)
# against an existing cluster.
#
# Usage: planton tofu plan --manifest catalog/gcp/gcpgkenodepool/e2e/manifest.yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeNodePool
metadata:
  name: hack-node-pool
  id: gkenp-hack-001
  org: planton-dev
  env: dev
spec:
  # project_id omitted: the pool lands in the provider's default project.
  clusterName:
    value: hack-gke-cluster
  location:
    value: us-central1
  nodePoolName: hack-batch-pool
  nodeLocations:
    - us-central1-a
    - us-central1-b
  autoscaling:
    minNodes: 0
    maxNodes: 5
    locationPolicy: ANY
  management:
    autoRepair: true
    autoUpgrade: true
  upgradeSettings:
    maxSurge: 2
    maxUnavailable: 1
    strategy: SURGE
  deletionPolicy: DELETE
  # Customized node drain requires per-project enablement from GCP support
  # (the API rejects it otherwise) — shown here as the documented example
  # of drain pacing on allowlisted projects.
  nodeDrainConfig:
    graceTerminationDuration: "60s"
    pdbTimeoutDuration: "600s"
    respectPdbDuringNodePoolDeletion: true
  nodeConfig:
    machineType: n2-standard-4
    diskSizeGb: 100
    diskType: pd-balanced
    imageType: COS_CONTAINERD
    spot: true
    labels:
      workload-class: batch
    taints:
      - key: workload-class
        value: batch
        effect: NO_SCHEDULE
    tags:
      - batch-nodes
    gcfsEnabled: true
    gvnicEnabled: true
    shieldedInstanceConfig:
      enableSecureBoot: true
    advancedMachineFeatures:
      threadsPerCore: 2
      performanceMonitoringUnit: STANDARD
    architectureTaintBehavior: ARM
    containerdConfig:
      privateRegistryAccess:
        enabled: true
        certificateAuthorityDomains:
          - fqdns:
              - registry.internal:5000
            gcpSecretManagerCertificateUri: projects/test-project-123/secrets/registry-ca/versions/latest
      registryHosts:
        - server: docker.io
          hosts:
            - host: https://mirror.internal
              capabilities:
                - pull
                - resolve
              dialTimeout: "10s"
    kubeletConfig:
      podPidsLimit: 4096
      insecureKubeletReadonlyPortEnabled: "FALSE"
      imageMinimumGcAge: "120s"
      imageMaximumGcAge: "86400s"
      maxParallelImagePulls: 3
      evictionSoft:
        memoryAvailable: 200Mi
      evictionSoftGracePeriod:
        memoryAvailable: "90s"
      # Percentage-only: GKE rejects absolute quantities for minimum
      # reclaim (the soft thresholds above accept either form).
      evictionMinimumReclaim:
        memoryAvailable: "10%"
      crashLoopBackOff:
        maxContainerRestartPeriod: "300s"
      topologyManager:
        policy: best-effort
        scope: container
    linuxNodeConfig:
      cgroupMode: CGROUP_MODE_V2
      transparentHugepageEnabled: TRANSPARENT_HUGEPAGE_ENABLED_MADVISE
      nodeKernelModuleLoadingPolicy: ENFORCE_SIGNED_MODULES
      swapConfig:
        enabled: true
        bootDiskProfile:
          swapSizePercent: 20
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.clusterName` | `string \| valueFrom` | yes |  | GcpGkeCluster (`status.outputs.name`) |
| `spec.location` | `string \| valueFrom` | yes |  | GcpGkeCluster (`status.outputs.location`) |
| `spec.nodePoolName` | `string` |  |  |  |
| `spec.namePrefix` | `string` |  |  |  |
| `spec.nodeLocations` | `[]string` |  |  |  |
| `spec.version` | `string` |  |  |  |
| `spec.maxPodsPerNode` | `int32` |  |  |  |
| `spec.initialNodeCount` | `uint32` |  |  |  |
| `spec.nodeCount` | `uint32` |  |  |  |
| `spec.autoscaling` | `GcpGkeNodePoolAutoscaling` |  |  |  |
| `spec.autoscaling.minNodes` | `uint32` |  |  |  |
| `spec.autoscaling.maxNodes` | `uint32` |  |  |  |
| `spec.autoscaling.totalMinNodes` | `uint32` |  |  |  |
| `spec.autoscaling.totalMaxNodes` | `uint32` |  |  |  |
| `spec.autoscaling.locationPolicy` | `string` |  |  |  |
| `spec.management` | `GcpGkeNodePoolManagement` |  |  |  |
| `spec.management.autoRepair` | `bool` |  | `true` |  |
| `spec.management.autoUpgrade` | `bool` |  | `true` |  |
| `spec.upgradeSettings` | `GcpGkeNodePoolUpgradeSettings` |  |  |  |
| `spec.upgradeSettings.maxSurge` | `uint32` |  |  |  |
| `spec.upgradeSettings.maxUnavailable` | `uint32` |  |  |  |
| `spec.upgradeSettings.strategy` | `string` |  |  |  |
| `spec.upgradeSettings.blueGreenSettings` | `GcpGkeNodePoolBlueGreenSettings` |  |  |  |
| `spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy` | `GcpGkeNodePoolStandardRolloutPolicy` | yes |  |  |
| `spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchPercentage` | `float` |  |  |  |
| `spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchNodeCount` | `uint32` |  |  |  |
| `spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchSoakDuration` | `string` |  |  |  |
| `spec.upgradeSettings.blueGreenSettings.nodePoolSoakDuration` | `string` |  |  |  |
| `spec.placementPolicy` | `GcpGkeNodePoolPlacementPolicy` |  |  |  |
| `spec.placementPolicy.type` | `string` | yes |  |  |
| `spec.placementPolicy.policyName` | `string` |  |  |  |
| `spec.placementPolicy.tpuTopology` | `string` |  |  |  |
| `spec.queuedProvisioningEnabled` | `bool` |  |  |  |
| `spec.networkConfig` | `GcpGkeNodePoolNetworkConfig` |  |  |  |
| `spec.networkConfig.createPodRange` | `bool` |  |  |  |
| `spec.networkConfig.podRange` | `string` |  |  |  |
| `spec.networkConfig.podIpv4CidrBlock` | `string` |  |  |  |
| `spec.networkConfig.enablePrivateNodes` | `bool` |  |  |  |
| `spec.networkConfig.totalEgressBandwidthTier` | `string` |  |  |  |
| `spec.networkConfig.podCidrOverprovisionDisabled` | `bool` |  |  |  |
| `spec.networkConfig.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfig.acceleratorNetworkProfile` | `string` |  |  |  |
| `spec.networkConfig.additionalNodeNetworks` | `[]GcpGkeNodePoolAdditionalNodeNetwork` |  |  |  |
| `spec.networkConfig.additionalNodeNetworks[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.networkConfig.additionalNodeNetworks[].subnetwork` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfig.additionalPodNetworks` | `[]GcpGkeNodePoolAdditionalPodNetwork` |  |  |  |
| `spec.networkConfig.additionalPodNetworks[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfig.additionalPodNetworks[].secondaryPodRange` | `string` | yes |  |  |
| `spec.networkConfig.additionalPodNetworks[].maxPodsPerNode` | `int32` |  |  |  |
| `spec.nodeConfig` | `GcpGkeNodePoolNodeConfig` |  |  |  |
| `spec.nodeConfig.machineType` | `string` |  | `e2-medium` |  |
| `spec.nodeConfig.diskSizeGb` | `uint32` |  |  |  |
| `spec.nodeConfig.diskType` | `string` |  |  |  |
| `spec.nodeConfig.imageType` | `string` |  | `COS_CONTAINERD` |  |
| `spec.nodeConfig.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.nodeConfig.oauthScopes` | `[]string` |  |  |  |
| `spec.nodeConfig.labels` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.resourceLabels` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.tags` | `[]string` |  |  |  |
| `spec.nodeConfig.metadata` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.taints` | `[]GcpGkeNodePoolTaint` |  |  |  |
| `spec.nodeConfig.taints[].key` | `string` | yes |  |  |
| `spec.nodeConfig.taints[].value` | `string` | yes |  |  |
| `spec.nodeConfig.taints[].effect` | `string` | yes |  |  |
| `spec.nodeConfig.spot` | `bool` |  |  |  |
| `spec.nodeConfig.preemptible` | `bool` |  |  |  |
| `spec.nodeConfig.guestAccelerators` | `[]GcpGkeNodePoolGuestAccelerator` |  |  |  |
| `spec.nodeConfig.guestAccelerators[].type` | `string` | yes |  |  |
| `spec.nodeConfig.guestAccelerators[].count` | `uint32` |  |  |  |
| `spec.nodeConfig.guestAccelerators[].gpuPartitionSize` | `string` |  |  |  |
| `spec.nodeConfig.guestAccelerators[].gpuDriverVersion` | `string` |  |  |  |
| `spec.nodeConfig.guestAccelerators[].gpuSharingConfig` | `GcpGkeNodePoolGpuSharingConfig` |  |  |  |
| `spec.nodeConfig.guestAccelerators[].gpuSharingConfig.gpuSharingStrategy` | `string` | yes |  |  |
| `spec.nodeConfig.guestAccelerators[].gpuSharingConfig.maxSharedClientsPerGpu` | `uint32` |  |  |  |
| `spec.nodeConfig.shieldedInstanceConfig` | `GcpGkeNodePoolShieldedInstanceConfig` |  |  |  |
| `spec.nodeConfig.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.nodeConfig.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  | `true` |  |
| `spec.nodeConfig.confidentialNodes` | `GcpGkeNodePoolConfidentialNodes` |  |  |  |
| `spec.nodeConfig.confidentialNodes.enabled` | `bool` |  |  |  |
| `spec.nodeConfig.confidentialNodes.confidentialInstanceType` | `string` |  |  |  |
| `spec.nodeConfig.minCpuPlatform` | `string` |  |  |  |
| `spec.nodeConfig.localSsdCount` | `uint32` |  |  |  |
| `spec.nodeConfig.ephemeralStorageLocalSsd` | `GcpGkeNodePoolEphemeralStorageLocalSsd` |  |  |  |
| `spec.nodeConfig.ephemeralStorageLocalSsd.localSsdCount` | `uint32` |  |  |  |
| `spec.nodeConfig.ephemeralStorageLocalSsd.dataCacheCount` | `uint32` |  |  |  |
| `spec.nodeConfig.localNvmeSsdBlock` | `GcpGkeNodePoolLocalNvmeSsdBlock` |  |  |  |
| `spec.nodeConfig.localNvmeSsdBlock.localSsdCount` | `uint32` |  |  |  |
| `spec.nodeConfig.gcfsEnabled` | `bool` |  |  |  |
| `spec.nodeConfig.gvnicEnabled` | `bool` |  |  |  |
| `spec.nodeConfig.fastSocketEnabled` | `bool` |  |  |  |
| `spec.nodeConfig.bootDiskKmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.nodeConfig.workloadMetadataMode` | `string` |  |  |  |
| `spec.nodeConfig.reservationAffinity` | `GcpGkeNodePoolReservationAffinity` |  |  |  |
| `spec.nodeConfig.reservationAffinity.consumeReservationType` | `string` | yes |  |  |
| `spec.nodeConfig.reservationAffinity.key` | `string` |  |  |  |
| `spec.nodeConfig.reservationAffinity.values` | `[]string` |  |  |  |
| `spec.nodeConfig.secondaryBootDisks` | `[]GcpGkeNodePoolSecondaryBootDisk` |  |  |  |
| `spec.nodeConfig.secondaryBootDisks[].diskImage` | `string` | yes |  |  |
| `spec.nodeConfig.secondaryBootDisks[].mode` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig` | `GcpGkeNodePoolKubeletConfig` |  |  |  |
| `spec.nodeConfig.kubeletConfig.cpuManagerPolicy` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.cpuCfsQuota` | `bool` |  |  |  |
| `spec.nodeConfig.kubeletConfig.cpuCfsQuotaPeriod` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.podPidsLimit` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.insecureKubeletReadonlyPortEnabled` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.maxParallelImagePulls` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.containerLogMaxSize` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.containerLogMaxFiles` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.imageGcLowThresholdPercent` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.imageGcHighThresholdPercent` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.imageMinimumGcAge` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.imageMaximumGcAge` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.allowedUnsafeSysctls` | `[]string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.singleProcessOomKill` | `bool` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft` | `GcpGkeNodePoolEvictionSignals` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.memoryAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.nodefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.nodefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.imagefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.imagefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoft.pidAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod` | `GcpGkeNodePoolEvictionGracePeriods` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim` | `GcpGkeNodePoolEvictionMinimumReclaim` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.crashLoopBackOff` | `GcpGkeNodePoolCrashLoopBackOff` |  |  |  |
| `spec.nodeConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.memoryManager` | `GcpGkeNodePoolMemoryManager` |  |  |  |
| `spec.nodeConfig.kubeletConfig.memoryManager.policy` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.topologyManager` | `GcpGkeNodePoolTopologyManager` |  |  |  |
| `spec.nodeConfig.kubeletConfig.topologyManager.policy` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.topologyManager.scope` | `string` |  |  |  |
| `spec.nodeConfig.kubeletConfig.shutdownGracePeriodSeconds` | `int64` |  |  |  |
| `spec.nodeConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig` | `GcpGkeNodePoolLinuxNodeConfig` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.sysctls` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.cgroupMode` | `string` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.hugepagesConfig` | `GcpGkeNodePoolHugepagesConfig` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.hugepagesConfig.hugepageSize2m` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.hugepagesConfig.hugepageSize1g` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.transparentHugepageEnabled` | `string` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.transparentHugepageDefrag` | `string` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.nodeKernelModuleLoadingPolicy` | `string` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.enablePtpKvmTimeSync` | `bool` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig` | `GcpGkeNodePoolSwapConfig` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.enabled` | `bool` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile` | `GcpGkeNodePoolSwapSizing` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile` | `GcpGkeNodePoolSwapDedicatedSsd` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile` | `GcpGkeNodePoolSwapSizing` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent` | `int32` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.encryptionConfig` | `GcpGkeNodePoolSwapEncryption` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled` | `bool` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.customNodeInit` | `GcpGkeNodePoolCustomNodeInit` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.customNodeInit.gcsUri` | `string` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.customNodeInit.gcsGeneration` | `int64` |  |  |  |
| `spec.nodeConfig.linuxNodeConfig.customNodeInit.secretManagerSecretUri` | `string` |  |  |  |
| `spec.nodeConfig.loggingVariant` | `string` |  |  |  |
| `spec.nodeConfig.flexStart` | `bool` |  |  |  |
| `spec.nodeConfig.maxRunDuration` | `string` |  |  |  |
| `spec.nodeConfig.enableConfidentialStorage` | `bool` |  |  |  |
| `spec.nodeConfig.localSsdEncryptionMode` | `string` |  |  |  |
| `spec.nodeConfig.gpudirectStrategy` | `string` |  |  |  |
| `spec.nodeConfig.nodeGroup` | `string` |  |  |  |
| `spec.nodeConfig.storagePools` | `[]string` |  |  |  |
| `spec.nodeConfig.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.advancedMachineFeatures` | `GcpGkeNodePoolAdvancedMachineFeatures` |  |  |  |
| `spec.nodeConfig.advancedMachineFeatures.threadsPerCore` | `int64` |  |  |  |
| `spec.nodeConfig.advancedMachineFeatures.enableNestedVirtualization` | `bool` |  |  |  |
| `spec.nodeConfig.advancedMachineFeatures.performanceMonitoringUnit` | `string` |  |  |  |
| `spec.nodeConfig.bootDisk` | `GcpGkeNodePoolBootDisk` |  |  |  |
| `spec.nodeConfig.bootDisk.diskType` | `string` |  |  |  |
| `spec.nodeConfig.bootDisk.sizeGb` | `int64` |  |  |  |
| `spec.nodeConfig.bootDisk.provisionedIops` | `int64` |  |  |  |
| `spec.nodeConfig.bootDisk.provisionedThroughput` | `int64` |  |  |  |
| `spec.nodeConfig.nodeImage` | `GcpGkeNodePoolNodeImage` |  |  |  |
| `spec.nodeConfig.nodeImage.image` | `string` | yes |  |  |
| `spec.nodeConfig.nodeImage.imageProject` | `string` |  |  |  |
| `spec.nodeConfig.soleTenantConfig` | `GcpGkeNodePoolSoleTenantConfig` |  |  |  |
| `spec.nodeConfig.soleTenantConfig.nodeAffinities` | `[]GcpGkeNodePoolSoleTenantAffinity` | yes |  |  |
| `spec.nodeConfig.soleTenantConfig.nodeAffinities[].key` | `string` | yes |  |  |
| `spec.nodeConfig.soleTenantConfig.nodeAffinities[].operator` | `string` | yes |  |  |
| `spec.nodeConfig.soleTenantConfig.nodeAffinities[].values` | `[]string` | yes |  |  |
| `spec.nodeConfig.soleTenantConfig.minNodeCpus` | `int32` |  |  |  |
| `spec.nodeConfig.sandboxType` | `string` |  |  |  |
| `spec.nodeConfig.windowsOsVersion` | `string` |  |  |  |
| `spec.nodeConfig.hostMaintenanceInterval` | `string` |  |  |  |
| `spec.nodeConfig.architectureTaintBehavior` | `string` |  |  |  |
| `spec.nodeConfig.containerdConfig` | `GcpGkeNodePoolContainerdConfig` |  |  |  |
| `spec.nodeConfig.containerdConfig.privateRegistryAccess` | `GcpGkeNodePoolPrivateRegistryAccess` |  |  |  |
| `spec.nodeConfig.containerdConfig.privateRegistryAccess.enabled` | `bool` |  |  |  |
| `spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains` | `[]GcpGkeNodePoolRegistryCaDomain` |  |  |  |
| `spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].fqdns` | `[]string` | yes |  |  |
| `spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].gcpSecretManagerCertificateUri` | `string` | yes |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts` | `[]GcpGkeNodePoolRegistryHost` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].server` | `string` | yes |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts` | `[]GcpGkeNodePoolRegistryHostEndpoint` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].host` | `string` | yes |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].capabilities` | `[]string` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].dialTimeout` | `string` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].overridePath` | `bool` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].caSecretUri` | `string` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].clientCertSecretUri` | `string` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].clientKeySecretUri` | `string` |  |  |  |
| `spec.nodeConfig.containerdConfig.registryHosts[].hosts[].headers` | `map<string, string>` |  |  |  |
| `spec.nodeConfig.containerdConfig.writableCgroupsEnabled` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.ignoreNodeCountChanges` | `bool` |  |  |  |
| `spec.nodeDrainConfig` | `GcpGkeNodePoolNodeDrainConfig` |  |  |  |
| `spec.nodeDrainConfig.graceTerminationDuration` | `string` |  |  |  |
| `spec.nodeDrainConfig.pdbTimeoutDuration` | `string` |  |  |  |
| `spec.nodeDrainConfig.respectPdbDuringNodePoolDeletion` | `bool` |  |  |  |
| `spec.excludeUpgradesUntilEndOfSupport` | `bool` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the node pool is created in. Must be the parent
cluster's project — node pools cannot live in a different project than
their cluster. Accepts a literal project ID or a reference to a
GcpProject resource. If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.clusterName

`string | valueFrom` · required

Name of the parent GKE cluster as created in GCP. Resolves from the
cluster's name output — the name GCP actually assigned — so it stays
correct even when the cluster's cloud name differs from its Planton
metadata.name. Immutable.

- references: GcpGkeCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.location

`string | valueFrom` · required

Location of the parent cluster — a region ("us-central1") for regional
clusters or a zone ("us-central1-a") for zonal ones. Must match the
cluster's own location; resolves from the cluster's location output by
default. Immutable. For a regional cluster, size limits in autoscaling
are PER ZONE (see autoscaling), and the pool gets one managed instance
group per zone.

- references: GcpGkeCluster (`status.outputs.location`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.location}} -- a bare string does not parse

### spec.nodePoolName

`string`

Name of the node pool in GKE. Immutable. If not specified, defaults to
metadata.name. Must be 1-40 characters: lowercase letters, digits, and
hyphens; starting with a letter and ending with a letter or digit.
Example: "general-pool", "spot-batch", "gpu-a100"

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$"}}

### spec.namePrefix

`string`

Prefix for a GENERATED pool name — GKE appends a random suffix, so
every replacement pool gets a fresh unique name (the
create-before-destroy pattern for pools that must be swapped without
a name collision). Mutually exclusive with node_pool_name; when both
are empty, node_pool_name defaults to metadata.name. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]([a-z0-9-]{0,36})?$"}}

### spec.nodeLocations

`[]string`

Zones this pool's nodes run in, e.g. ["us-central1-a", "us-central1-b"].
Must be within the cluster's region. If unspecified, the cluster-level
node_locations apply (all zones in the region for a regional cluster).
Mutable. Narrowing a pool to fewer zones cuts cost and inter-zone
traffic for workloads that tolerate zonal risk.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}}}

### spec.version

`string`

Kubernetes version for the nodes. If empty (recommended), GKE picks the
version and auto-upgrade keeps it current. Setting an explicit version
while management.auto_upgrade is on makes the two fight — pin a version
only with auto-upgrade off, and expect to own upgrades manually from
then on.

### spec.maxPodsPerNode

`int32` · optional (explicit presence)

Maximum pods per node in this pool (8-256), overriding the cluster's
default (110). Lower values shrink the per-node pod CIDR slice so the
pod range stretches across more nodes. Immutable. Only effective on
VPC-native clusters (which every Planton GKE cluster is).

- rule: {"int32":{"lte":256,"gte":8}}

### spec.initialNodeCount

`uint32` · optional (explicit presence)

Number of nodes the pool starts with when autoscaling manages the size
(per zone for regional clusters). Only meaningful with autoscaling —
a fixed-size pool's size IS node_count. Immutable: changing it forces
pool recreation, so set it once and let the autoscaler own the size
from then on.

### spec.nodeCount

`uint32`

Fixed number of nodes (per zone for regional clusters). The pool
stays at this size until you change it.

### spec.autoscaling

`GcpGkeNodePoolAutoscaling`

Cluster-autoscaler management of the pool size between bounds.

- rule: use per-zone limits (min_nodes/max_nodes) or total limits (total_min_nodes/total_max_nodes), not both — they are mutually exclusive addressing modes
- rule: autoscaling needs a maximum: set max_nodes (per-zone mode) or total_max_nodes (total mode)
- rule: min_nodes cannot exceed max_nodes
- rule: total_min_nodes cannot exceed total_max_nodes

### spec.autoscaling.minNodes

`uint32` · optional (explicit presence)

Minimum nodes PER ZONE. 0 allows scale-to-zero — the pattern for Spot
and GPU pools that should cost nothing while idle.

### spec.autoscaling.maxNodes

`uint32` · optional (explicit presence)

Maximum nodes PER ZONE. A regional cluster in 3 zones with max_nodes=4
can reach 12 nodes.

### spec.autoscaling.totalMinNodes

`uint32` · optional (explicit presence)

Minimum nodes across ALL zones (total addressing mode).

### spec.autoscaling.totalMaxNodes

`uint32` · optional (explicit presence)

Maximum nodes across ALL zones (total addressing mode) — an absolute
cost cap regardless of zone spread.

### spec.autoscaling.locationPolicy

`string`

Scale-up algorithm: BALANCED spreads new nodes to even out zone sizes;
ANY prioritizes unused reservations and reduces Spot preemption risk —
prefer ANY for Spot pools.

- rule: location_policy must be empty, BALANCED, or ANY

### spec.management

`GcpGkeNodePoolManagement`

Auto-repair and auto-upgrade. If omitted, both default to true — GKE's
own defaults and the posture release channels expect.

### spec.management.autoRepair

`bool` · optional (explicit presence)

Automatically repair nodes that fail health checks. Keep on; a pool of
broken nodes heals itself instead of paging you.

- default: `true`

### spec.management.autoUpgrade

`bool` · optional (explicit presence)

Automatically upgrade node Kubernetes versions to track the control
plane. Keep on unless you pin `version` and own upgrades manually
(required for pools on a release channel).

- default: `true`

### spec.upgradeSettings

`GcpGkeNodePoolUpgradeSettings`

How node upgrades roll through the pool: surge (default — add
max_surge nodes, drain max_unavailable at a time) or blue-green
(provision a full green pool, shift workloads, soak, delete blue).
If omitted, GKE's default surge settings (max_surge=1,
max_unavailable=0) apply.

- rule: blue_green_settings apply only when strategy is BLUE_GREEN
- rule: max_surge/max_unavailable apply to the SURGE strategy — remove them when strategy is BLUE_GREEN

### spec.upgradeSettings.maxSurge

`uint32` · optional (explicit presence)

Additional nodes added during a surge upgrade (0 or more). Higher means
faster upgrades at temporary extra cost. max_surge + max_unavailable
must be at least 1 and at most 20 combined.

### spec.upgradeSettings.maxUnavailable

`uint32` · optional (explicit presence)

Nodes that may be simultaneously unavailable during a surge upgrade.
Higher trades workload disruption for speed.

### spec.upgradeSettings.strategy

`string`

SURGE (default): rolling replacement, cheapest. BLUE_GREEN: provision a
complete new node set, migrate, soak, then delete the old one — safest
rollback story, double capacity while in flight.

- rule: strategy must be empty, SURGE, or BLUE_GREEN

### spec.upgradeSettings.blueGreenSettings

`GcpGkeNodePoolBlueGreenSettings`

Blue-green rollout pacing. Only with strategy BLUE_GREEN.

### spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy

`GcpGkeNodePoolStandardRolloutPolicy` · required

How the blue pool drains, batch by batch.

- rule: {"required":true}
- rule: set batch_percentage or batch_node_count, not both

### spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchPercentage

`float` · optional (explicit presence)

Fraction of blue nodes drained per batch, 0.0-1.0.

- rule: {"float":{"lte":1,"gte":0}}

### spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchNodeCount

`uint32` · optional (explicit presence)

Number of blue nodes drained per batch.

### spec.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchSoakDuration

`string`

Soak time after each batch drains before the next starts. Duration in
seconds format, e.g. "600s".

- rule: batch_soak_duration must be a seconds-format duration like "600s"

### spec.upgradeSettings.blueGreenSettings.nodePoolSoakDuration

`string`

Time after the entire blue pool is drained before it is deleted —
the rollback window. Duration in seconds format, e.g. "3600s".

- rule: node_pool_soak_duration must be a seconds-format duration like "3600s"

### spec.placementPolicy

`GcpGkeNodePoolPlacementPolicy`

Compact placement: co-locates nodes physically for low inter-node
latency (tightly-coupled HPC/ML workloads) — and carries the TPU
topology for TPU pools. Immutable.

### spec.placementPolicy.type

`string` · required

COMPACT places nodes close together for minimal inter-node latency —
for tightly-coupled HPC and multi-node ML training. Requires machine
families that support compact placement (C2/C3/A2/A3...).

- rule: {"required":true,"string":{"in":["COMPACT"]}}

### spec.placementPolicy.policyName

`string`

Optional user-supplied compute resource policy to place nodes under.
Must be in the same project and region as the pool. When empty, GKE
creates and owns the placement policy.

### spec.placementPolicy.tpuTopology

`string`

TPU placement topology, e.g. "2x2x2" — TPU pools only.
https://cloud.google.com/kubernetes-engine/docs/concepts/plan-tpus#topology

### spec.queuedProvisioningEnabled

`bool`

Nodes are obtainable only through the ProvisioningRequest API (Dynamic
Workload Scheduler queued provisioning) — the pool provisions capacity
in whole batches when it becomes available, for large atomic workloads
like multi-node training jobs. Immutable.

### spec.networkConfig

`GcpGkeNodePoolNetworkConfig`

Pool-level networking overrides (pod range, private nodes, network
performance). If omitted, the cluster-level defaults apply.

- rule: create_pod_range names the NEW range to create — set pod_range (the range name) alongside it, or set neither to use the cluster's pod range

### spec.networkConfig.createPodRange

`bool`

Create a NEW secondary range for this pool's pods (named by pod_range,
sized by pod_ipv4_cidr_block) instead of drawing from the cluster's pod
range. Immutable. Dedicated pod ranges isolate address planning per
pool — useful when one pool's churn would exhaust the shared range.

### spec.networkConfig.podRange

`string`

The secondary range for this pool's pod IPs. With create_pod_range,
the name given to the new range; without it, the name of an EXISTING
secondary range on the cluster's subnetwork. Immutable.

### spec.networkConfig.podIpv4CidrBlock

`string`

CIDR (e.g. "10.96.0.0/14") or netmask size (e.g. "/14") for the new pod
range when create_pod_range is set; empty lets GKE choose. Immutable.

### spec.networkConfig.enablePrivateNodes

`bool` · optional (explicit presence)

Whether this pool's nodes get only internal IPs, overriding the
cluster-level private-nodes setting. Unset inherits from the cluster.

### spec.networkConfig.totalEgressBandwidthTier

`string`

Network bandwidth tier: TIER_1 unlocks up to 100 Gbps total egress on
supported machine families (N2/N2D/C2/C3...). Requires gVNIC on this
pool (node_config.gvnic_enabled).

- rule: total_egress_bandwidth_tier must be empty, TIER_UNSPECIFIED, or TIER_1

### spec.networkConfig.podCidrOverprovisionDisabled

`bool`

Disables the pod CIDR overprovisioning for this pool (GKE normally
doubles the pod range slice per node). Only relevant with a dedicated
pod range. Immutable.

### spec.networkConfig.subnetwork

`string | valueFrom`

Places this pool's nodes on a DIFFERENT subnetwork than the cluster's
default node subnetwork (same VPC). Accepts a subnetwork self link or
a reference to a GcpSubnetwork resource. Immutable. Used to give pools
their own address planning or firewall scope.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfig.acceleratorNetworkProfile

`string`

Accelerator network profile for the pool (e.g. RDMA-capable profiles
on GPU supercomputer shapes). GKE validates the profile against the
machine family at apply. Immutable.

### spec.networkConfig.additionalNodeNetworks

`[]GcpGkeNodePoolAdditionalNodeNetwork`

Additional node network interfaces (multi-networking): each entry
attaches every node to one more VPC network/subnetwork. Requires the
cluster's enable_multi_networking. Immutable.

### spec.networkConfig.additionalNodeNetworks[].network

`string | valueFrom` · required

The VPC network to attach. Accepts a network name/self link or a
reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.networkConfig.additionalNodeNetworks[].subnetwork

`string | valueFrom` · required

The subnetwork on that network the interface draws its IP from.
Accepts a name/self link or a reference to a GcpSubnetwork resource.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfig.additionalPodNetworks

`[]GcpGkeNodePoolAdditionalPodNetwork`

Additional pod ranges (multi-networking): each entry makes a
secondary range on a subnetwork available to this pool's pods beyond
the primary pod range. Requires the cluster's
enable_multi_networking. Immutable.

### spec.networkConfig.additionalPodNetworks[].subnetwork

`string | valueFrom`

The subnetwork holding the secondary range. Accepts a name/self link
or a reference to a GcpSubnetwork resource.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfig.additionalPodNetworks[].secondaryPodRange

`string` · required

Name of the secondary range on that subnetwork used for pod IPs.

- rule: {"required":true}

### spec.networkConfig.additionalPodNetworks[].maxPodsPerNode

`int32` · optional (explicit presence)

Maximum pods per node drawing from this range (8-256).

- rule: {"int32":{"lte":256,"gte":8}}

### spec.nodeConfig

`GcpGkeNodePoolNodeConfig`

The node VM configuration: machine type, disks, identity, scheduling
constraints, accelerators, security posture, and kubelet/OS tuning.
If omitted entirely, GKE defaults apply (e2-medium, 100 GB pd-balanced,
Container-Optimized OS, Compute Engine default service account).

- rule: spot and preemptible are mutually exclusive — spot is the current model (no 24h max lifetime); preemptible is the legacy one
- rule: fast_socket_enabled requires gvnic_enabled — NCCL Fast Socket rides on gVNIC

### spec.nodeConfig.machineType

`string` · optional (explicit presence)

Compute Engine machine type, e.g. "e2-medium", "n2-standard-8",
"a2-highgpu-1g". Defaults to e2-medium (2 vCPU, 4 GB) — fine for
sandboxes, undersized for real workloads. Immutable (changing it
replaces the pool's nodes).

- default: `e2-medium`

### spec.nodeConfig.diskSizeGb

`uint32` · optional (explicit presence)

Boot disk size in GB per node (min 10). GKE's default is 100.

- rule: {"uint32":{"gte":10}}

### spec.nodeConfig.diskType

`string`

Boot disk type. pd-balanced (GKE's current default) suits most pools;
pd-ssd for I/O-heavy node-local work; hyperdisk-balanced on machine
families that require it (C3/C3D and newer).

- rule: disk_type must be empty, pd-standard, pd-balanced, pd-ssd, hyperdisk-balanced, hyperdisk-extreme, or hyperdisk-throughput

### spec.nodeConfig.imageType

`string` · optional (explicit presence)

Node OS image. COS_CONTAINERD (Container-Optimized OS, the default and
GKE's recommendation), UBUNTU_CONTAINERD (when you need Ubuntu
packages/kernel modules), or WINDOWS_LTSC_CONTAINERD for Windows
workloads.

- default: `COS_CONTAINERD`
- rule: image_type must be COS_CONTAINERD, UBUNTU_CONTAINERD, or WINDOWS_LTSC_CONTAINERD

### spec.nodeConfig.serviceAccount

`string | valueFrom`

Service account the node VMs run as. Accepts an email literal or a
reference to a GcpServiceAccount resource. Defaults to the Compute
Engine default SA — create a minimal dedicated SA for production and
grant workload permissions through Workload Identity instead of node
scopes.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.nodeConfig.oauthScopes

`[]string`

OAuth scopes on the node VMs. Empty applies GKE's defaults
(devstorage.read_only, logging.write, monitoring). With Workload
Identity (the Planton cluster default), workload permissions come from
IAM on Kubernetes service accounts — node scopes only gate node-level
agents, so the defaults are usually right. One node-level agent that
matters: the kubelet pulls images from Artifact Registry with the node
service account under devstorage.read_only, so a private image in a
repository that account is granted on needs no pull secret at all.
Dropping that scope, or a repository the account is not granted on, is
why a pod would need a login declared on the workload instead.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.nodeConfig.labels

`map<string, string>`

Kubernetes labels applied to every node in the pool — what nodeSelector
and affinity rules match on. Example: {"workload-class": "batch"}.

### spec.nodeConfig.resourceLabels

`map<string, string>`

GCE resource labels on the node VMs (cloud billing/inventory labels,
not Kubernetes labels). Merged with the standard platform labels;
platform attribution keys win on conflict.

### spec.nodeConfig.tags

`[]string`

GCE network tags on the node VMs — what VPC firewall rules match.
GKE adds its own cluster tag automatically; entries here are additive.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.nodeConfig.metadata

`map<string, string>`

GCE instance metadata key/value pairs. GKE requires
"disable-legacy-endpoints" = "true" and both engines enforce it
beneath any entries set here. Immutable.

### spec.nodeConfig.taints

`[]GcpGkeNodePoolTaint`

Kubernetes taints applied to every node — the scheduling fence that
keeps general workloads off special-purpose pools. Pair each taint
with a matching toleration on the intended workloads. GPU pools get
an automatic nvidia.com/gpu taint from GKE.

### spec.nodeConfig.taints[].key

`string` · required

Taint key, e.g. "workload-class".

- rule: {"required":true}

### spec.nodeConfig.taints[].value

`string` · required

Taint value, e.g. "batch".

- rule: {"required":true}

### spec.nodeConfig.taints[].effect

`string` · required

NO_SCHEDULE fences new pods without a toleration; PREFER_NO_SCHEDULE
is advisory; NO_EXECUTE also evicts running pods without a toleration.

- rule: {"required":true,"string":{"in":["NO_SCHEDULE","PREFER_NO_SCHEDULE","NO_EXECUTE"]}}

### spec.nodeConfig.spot

`bool`

Spot VMs: deeply discounted (60-91%) capacity with no availability
guarantee — nodes can be preempted with 30s notice. The current model
(no 24-hour max lifetime; replaces preemptible). Pair with
scale-to-zero autoscaling, ANY location policy, and taints so only
fault-tolerant workloads land here. Immutable.

### spec.nodeConfig.preemptible

`bool`

Legacy preemptible VMs (24-hour max lifetime). Prefer spot for new
pools. Immutable.

### spec.nodeConfig.guestAccelerators

`[]GcpGkeNodePoolGuestAccelerator`

GPU accelerators attached to every node. The machine type must support
the accelerator (or be an accelerator-optimized family like A2/A3/G2,
where the GPU is implied by the machine type and this block must match
it). GKE taints GPU nodes automatically.

### spec.nodeConfig.guestAccelerators[].type

`string` · required

Accelerator type resource name, e.g. "nvidia-tesla-t4",
"nvidia-l4", "nvidia-a100-80gb". Must be available in the pool's
zones.

- rule: {"required":true}

### spec.nodeConfig.guestAccelerators[].count

`uint32`

Number of accelerator cards per node.

- rule: {"uint32":{"gte":1}}

### spec.nodeConfig.guestAccelerators[].gpuPartitionSize

`string`

NVIDIA MIG partition size, e.g. "1g.5gb" — slices one physical GPU
into isolated instances (A100/H100 families).

### spec.nodeConfig.guestAccelerators[].gpuDriverVersion

`string`

Driver installation: DEFAULT (GKE installs the default driver
version), LATEST (newest available; COS only), or
INSTALLATION_DISABLED (bring your own DaemonSet). If omitted on GKE
1.30.1+, DEFAULT applies.

- rule: gpu_driver_version must be empty, INSTALLATION_DISABLED, DEFAULT, or LATEST

### spec.nodeConfig.guestAccelerators[].gpuSharingConfig

`GcpGkeNodePoolGpuSharingConfig`

GPU sharing: lets multiple pods share one physical GPU.

### spec.nodeConfig.guestAccelerators[].gpuSharingConfig.gpuSharingStrategy

`string` · required

GPU_TIME_SHARING context-switches the GPU between pods; MPS (NVIDIA
Multi-Process Service) runs them concurrently with resource limits.

- rule: {"required":true,"string":{"in":["GPU_TIME_SHARING","MPS"]}}

### spec.nodeConfig.guestAccelerators[].gpuSharingConfig.maxSharedClientsPerGpu

`uint32`

Maximum pods sharing each physical GPU.

- rule: {"uint32":{"gte":1}}

### spec.nodeConfig.shieldedInstanceConfig

`GcpGkeNodePoolShieldedInstanceConfig`

Shielded VM options. GKE's defaults: secure boot off (to tolerate
unsigned third-party kernel modules), integrity monitoring on. Enable
secure boot unless a workload loads unsigned modules. Immutable.

### spec.nodeConfig.shieldedInstanceConfig.enableSecureBoot

`bool`

Verify boot components against a signature baseline. GCP default
false — because unsigned third-party kernel modules fail secure boot.
Turn it on unless you load such modules.

### spec.nodeConfig.shieldedInstanceConfig.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Monitor and attest boot integrity at runtime (GCP default true).

- default: `true`

### spec.nodeConfig.confidentialNodes

`GcpGkeNodePoolConfidentialNodes`

Confidential GKE nodes: hardware memory encryption (AMD SEV / Intel
TDX) for the node VMs. Requires a supporting machine family (N2D/C2D/
C3D...). Immutable.

### spec.nodeConfig.confidentialNodes.enabled

`bool`

Whether confidential nodes are enabled for this pool.

### spec.nodeConfig.confidentialNodes.confidentialInstanceType

`string`

Confidential computing technology: SEV (AMD, the common choice),
SEV_SNP, or TDX (Intel). Empty lets GCP choose for the machine family.

- rule: confidential_instance_type must be empty, SEV, SEV_SNP, or TDX

### spec.nodeConfig.minCpuPlatform

`string`

Minimum CPU platform, e.g. "Intel Ice Lake" — pins scheduling to that
CPU generation or newer for instruction-set or performance floors.
Immutable.

### spec.nodeConfig.localSsdCount

`uint32` · optional (explicit presence)

Local SSDs attached to each node for scratch I/O, automatically
formatted and mounted by GKE (SCSI interface; legacy knob). For NVMe,
prefer ephemeral_storage_local_ssd or local_nvme_ssd_block. Immutable.

### spec.nodeConfig.ephemeralStorageLocalSsd

`GcpGkeNodePoolEphemeralStorageLocalSsd`

Back ephemeral storage (emptyDir, container layers, logs) with local
NVMe SSDs instead of the boot disk — the biggest node-side I/O win for
churn-heavy workloads. Immutable.

### spec.nodeConfig.ephemeralStorageLocalSsd.localSsdCount

`uint32`

Number of local SSDs backing ephemeral storage. Each is 375 GB (or
3000 GB on Z3); count must match what the machine type supports.

- rule: {"uint32":{"gte":1}}

### spec.nodeConfig.ephemeralStorageLocalSsd.dataCacheCount

`uint32` · optional (explicit presence)

Of those, SSDs dedicated to GKE Data Cache (read caching for
persistent volumes).

### spec.nodeConfig.localNvmeSsdBlock

`GcpGkeNodePoolLocalNvmeSsdBlock`

Attach raw-block local NVMe SSDs for workloads that manage their own
filesystem (databases with direct block access). Immutable.

### spec.nodeConfig.localNvmeSsdBlock.localSsdCount

`uint32`

Number of raw-block local NVMe SSDs (375 GB each).

- rule: {"uint32":{"gte":1}}

### spec.nodeConfig.gcfsEnabled

`bool`

Image streaming (GCFS): containers start before the full image is
pulled, pulling data on demand — large-image pools (ML frameworks)
start minutes faster. Requires Container-Optimized OS.

### spec.nodeConfig.gvnicEnabled

`bool`

gVNIC (Google Virtual NIC): higher-throughput networking than virtio;
required for TIER_1 bandwidth and 100+ Gbps machine shapes. Immutable.

### spec.nodeConfig.fastSocketEnabled

`bool`

NCCL Fast Socket: optimizes multi-node collective communication for
distributed GPU training. Requires gvnic_enabled.

### spec.nodeConfig.bootDiskKmsKey

`string | valueFrom`

Customer-managed encryption key (CMEK) for the node boot disks.
Accepts a full crypto key path or a reference to a GcpKmsKey resource.
The Compute Engine service agent needs Encrypter/Decrypter on the key.
Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.nodeConfig.workloadMetadataMode

`string`

How workloads see instance metadata: GKE_METADATA runs the GKE
metadata server per node (required for Workload Identity — the default
on WI clusters and the right answer); GCE_METADATA exposes the raw VM
metadata (legacy, leaks node credentials to pods).

- rule: workload_metadata_mode must be empty, GCE_METADATA, or GKE_METADATA

### spec.nodeConfig.reservationAffinity

`GcpGkeNodePoolReservationAffinity`

Consume Compute Engine reservations: committed capacity for the pool.
ANY_RESERVATION uses any matching reservation; SPECIFIC_RESERVATION
targets one by name (key "compute.googleapis.com/reservation-name",
values [name]); NO_RESERVATION opts out. Immutable.

- rule: SPECIFIC_RESERVATION requires key "compute.googleapis.com/reservation-name" and the reservation name in values

### spec.nodeConfig.reservationAffinity.consumeReservationType

`string` · required

NO_RESERVATION opts out; ANY_RESERVATION consumes any matching
reservation; SPECIFIC_RESERVATION targets one by name;
ANY_RESERVATION_THEN_FAIL consumes matching reservations and fails
scale-up (instead of falling back to on-demand) when none is
available.

- rule: {"required":true,"string":{"in":["NO_RESERVATION","ANY_RESERVATION","SPECIFIC_RESERVATION","ANY_RESERVATION_THEN_FAIL"]}}

### spec.nodeConfig.reservationAffinity.key

`string`

Reservation label key — "compute.googleapis.com/reservation-name" for
SPECIFIC_RESERVATION.

### spec.nodeConfig.reservationAffinity.values

`[]string`

Reservation label values — the reservation name(s).

### spec.nodeConfig.secondaryBootDisks

`[]GcpGkeNodePoolSecondaryBootDisk`

Secondary boot disks that preload container images or data onto every
node — cold-start acceleration for very large images. Immutable.

### spec.nodeConfig.secondaryBootDisks[].diskImage

`string` · required

Disk image to create the secondary boot disk from (a prepared image
containing the container images/data to preload).

- rule: {"required":true}

### spec.nodeConfig.secondaryBootDisks[].mode

`string`

CONTAINER_IMAGE_CACHE serves preloaded container images to the
container runtime. Empty attaches the disk without special handling.

- rule: mode must be empty or CONTAINER_IMAGE_CACHE

### spec.nodeConfig.kubeletConfig

`GcpGkeNodePoolKubeletConfig`

Kubelet tuning: CPU management, PID limits, log rotation, image GC.
Only set what you need; unset fields keep GKE defaults.

- rule: shutdown_grace_period_critical_pods_seconds cannot exceed shutdown_grace_period_seconds

### spec.nodeConfig.kubeletConfig.cpuManagerPolicy

`string`

CPU management: "static" gives Guaranteed-QoS pods exclusive cores
(latency-sensitive workloads); "none" (default) shares cores.

- rule: cpu_manager_policy must be empty, static, or none

### spec.nodeConfig.kubeletConfig.cpuCfsQuota

`bool` · optional (explicit presence)

Enforce CPU CFS quota for containers with CPU limits. Disabling trades
throttling for potential node CPU contention.

### spec.nodeConfig.kubeletConfig.cpuCfsQuotaPeriod

`string`

CFS quota period, e.g. "100ms" (kubelet default) — shorter periods
smooth throttling for latency-sensitive workloads.

### spec.nodeConfig.kubeletConfig.podPidsLimit

`int64` · optional (explicit presence)

Maximum processes per pod — a fork-bomb fence for multi-tenant pools.

- rule: {"int64":{"lte":"4194304","gte":"1024"}}

### spec.nodeConfig.kubeletConfig.insecureKubeletReadonlyPortEnabled

`string`

The kubelet's insecure read-only port 10255: FALSE closes it (the
hardened posture new clusters default to); TRUE keeps it open for
legacy monitoring agents that still scrape it.

- rule: insecure_kubelet_readonly_port_enabled must be empty, TRUE, or FALSE

### spec.nodeConfig.kubeletConfig.maxParallelImagePulls

`int64` · optional (explicit presence)

Parallel image pulls (kubelet default serializes fewer) — speeds up
busy nodes that churn many images.

### spec.nodeConfig.kubeletConfig.containerLogMaxSize

`string`

Container log file size before rotation, e.g. "10Mi" (10Mi-500Mi).

### spec.nodeConfig.kubeletConfig.containerLogMaxFiles

`int64` · optional (explicit presence)

Rotated container log files kept per container (2-10).

- rule: {"int64":{"lte":"10","gte":"2"}}

### spec.nodeConfig.kubeletConfig.imageGcLowThresholdPercent

`int64` · optional (explicit presence)

Disk usage percent below which image GC never runs (must be lower
than the high threshold).

- rule: {"int64":{"lte":"85","gte":"10"}}

### spec.nodeConfig.kubeletConfig.imageGcHighThresholdPercent

`int64` · optional (explicit presence)

Disk usage percent above which image GC always runs.

- rule: {"int64":{"lte":"85","gte":"10"}}

### spec.nodeConfig.kubeletConfig.imageMinimumGcAge

`string`

Minimum age an unused image must reach before GC may remove it,
seconds format, e.g. "120s".

- rule: image_minimum_gc_age must be a seconds-format duration like "120s"

### spec.nodeConfig.kubeletConfig.imageMaximumGcAge

`string`

Maximum age an unused image may reach before GC removes it regardless
of disk pressure, seconds format, e.g. "86400s".

- rule: image_maximum_gc_age must be a seconds-format duration like "86400s"

### spec.nodeConfig.kubeletConfig.allowedUnsafeSysctls

`[]string`

Sysctl patterns pods may set as unsafe sysctls (e.g. "net.*",
"kernel.shm*") — gate carefully: unsafe sysctls affect the whole
node, not just the pod that sets them.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.nodeConfig.kubeletConfig.evictionMaxPodGracePeriodSeconds

`int64` · optional (explicit presence)

Maximum seconds the kubelet grants a pod to terminate during
soft-eviction (caps the pod's own grace period in that path).

- rule: {"int64":{"gte":"0"}}

### spec.nodeConfig.kubeletConfig.singleProcessOomKill

`bool` · optional (explicit presence)

Kill only the offending process instead of the whole container on
OOM — for containers running multiple processes where one leaking
process should not take down its siblings.

### spec.nodeConfig.kubeletConfig.evictionSoft

`GcpGkeNodePoolEvictionSignals`

Soft eviction thresholds: the kubelet starts graceful pod eviction
when a signal stays past its threshold for the paired grace period.

### spec.nodeConfig.kubeletConfig.evictionSoft.memoryAvailable

`string`

Threshold for memory.available.

### spec.nodeConfig.kubeletConfig.evictionSoft.nodefsAvailable

`string`

Threshold for nodefs.available (the filesystem backing pod volumes
and logs).

### spec.nodeConfig.kubeletConfig.evictionSoft.nodefsInodesFree

`string`

Threshold for nodefs.inodesFree.

### spec.nodeConfig.kubeletConfig.evictionSoft.imagefsAvailable

`string`

Threshold for imagefs.available (the filesystem backing container
images and writable layers).

### spec.nodeConfig.kubeletConfig.evictionSoft.imagefsInodesFree

`string`

Threshold for imagefs.inodesFree.

### spec.nodeConfig.kubeletConfig.evictionSoft.pidAvailable

`string`

Threshold for pid.available.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod

`GcpGkeNodePoolEvictionGracePeriods`

Grace periods paired with eviction_soft: how long a signal must stay
past its threshold before eviction starts (durations like "90s").

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.memoryAvailable

`string`

Grace period for memory.available.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.nodefsAvailable

`string`

Grace period for nodefs.available.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.nodefsInodesFree

`string`

Grace period for nodefs.inodesFree.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.imagefsAvailable

`string`

Grace period for imagefs.available.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.imagefsInodesFree

`string`

Grace period for imagefs.inodesFree.

### spec.nodeConfig.kubeletConfig.evictionSoftGracePeriod.pidAvailable

`string`

Grace period for pid.available.

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim

`GcpGkeNodePoolEvictionMinimumReclaim`

Minimum amounts reclaimed per eviction: prevents thrashing by making
each eviction free at least this much of the signal's resource.

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.memoryAvailable

`string`

Minimum reclaim for memory.available, percentage like "10%".

- rule: eviction_minimum_reclaim.memory_available must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.nodefsAvailable

`string`

Minimum reclaim for nodefs.available, percentage like "10%".

- rule: eviction_minimum_reclaim.nodefs_available must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.nodefsInodesFree

`string`

Minimum reclaim for nodefs.inodesFree, percentage like "10%".

- rule: eviction_minimum_reclaim.nodefs_inodes_free must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.imagefsAvailable

`string`

Minimum reclaim for imagefs.available, percentage like "10%".

- rule: eviction_minimum_reclaim.imagefs_available must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.imagefsInodesFree

`string`

Minimum reclaim for imagefs.inodesFree, percentage like "10%".

- rule: eviction_minimum_reclaim.imagefs_inodes_free must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.evictionMinimumReclaim.pidAvailable

`string`

Minimum reclaim for pid.available, percentage like "10%".

- rule: eviction_minimum_reclaim.pid_available must be a percentage like "10%" — GKE rejects absolute quantities here

### spec.nodeConfig.kubeletConfig.crashLoopBackOff

`GcpGkeNodePoolCrashLoopBackOff`

Caps the exponential backoff for restarting crashed containers —
lower caps recover crash-looping containers faster at the cost of
more restart churn.

### spec.nodeConfig.kubeletConfig.crashLoopBackOff.maxContainerRestartPeriod

`string`

Maximum backoff delay between restarts of a crashing container,
seconds format (e.g. "300s"; kubelet default caps at 300s).

- rule: max_container_restart_period must be a seconds-format duration like "300s"

### spec.nodeConfig.kubeletConfig.memoryManager

`GcpGkeNodePoolMemoryManager`

Kubernetes Memory Manager policy: Static reserves exclusive NUMA
memory for Guaranteed-QoS pods; None (default) does not.

### spec.nodeConfig.kubeletConfig.memoryManager.policy

`string`

"Static" reserves exclusive NUMA-aligned memory for Guaranteed-QoS
pods; "None" disables the manager. Capitalized, per the kubelet's
own policy names.

- rule: policy must be empty, None, or Static

### spec.nodeConfig.kubeletConfig.topologyManager

`GcpGkeNodePoolTopologyManager`

Kubernetes Topology Manager: aligns CPU, memory, and device (GPU)
NUMA placement per pod or per container — for latency-critical and
HPC workloads that suffer on cross-NUMA access.

### spec.nodeConfig.kubeletConfig.topologyManager.policy

`string`

Alignment policy: none (default), best-effort, restricted, or
single-numa-node (strictest — pods failing alignment are rejected).

- rule: policy must be empty, none, best-effort, restricted, or single-numa-node

### spec.nodeConfig.kubeletConfig.topologyManager.scope

`string`

Alignment scope: container (default; each container aligned
independently) or pod (all containers of a pod share one alignment).

- rule: scope must be empty, container, or pod

### spec.nodeConfig.kubeletConfig.shutdownGracePeriodSeconds

`int64` · optional (explicit presence)

Graceful node shutdown: the total time, in seconds, a node delays its
shutdown so every pod (critical and non-critical) can terminate
cleanly when the VM is reclaimed. Only configurable on Spot or
preemptible pools (the ones that get reclaimed). Between 10 and
10000. Leave unset for GKE's default; sent only when set because the
API fills the value itself.

- rule: {"int64":{"lte":"10000","gte":"10"}}

### spec.nodeConfig.kubeletConfig.shutdownGracePeriodCriticalPodsSeconds

`int64` · optional (explicit presence)

The portion of shutdown_grace_period_seconds reserved for critical
pods (system-node-critical and system-cluster-critical priority
classes) after ordinary pods have been given their share. Must not
exceed shutdown_grace_period_seconds. Spot or preemptible pools only.
Sent only when set because the API fills the value itself.

- rule: {"int64":{"lte":"10000","gte":"0"}}

### spec.nodeConfig.linuxNodeConfig

`GcpGkeNodePoolLinuxNodeConfig`

Linux node OS tuning: sysctls, cgroup mode, hugepages.

### spec.nodeConfig.linuxNodeConfig.sysctls

`map<string, string>`

Sysctls applied to every node, e.g.
{"net.core.somaxconn": "4096"} — only GKE's allowlisted keys.

### spec.nodeConfig.linuxNodeConfig.cgroupMode

`string`

Container runtime cgroup mode: CGROUP_MODE_V2 (the modern default on
current GKE versions) or CGROUP_MODE_V1 for workloads that still need
v1.

- rule: cgroup_mode must be empty, CGROUP_MODE_UNSPECIFIED, CGROUP_MODE_V1, or CGROUP_MODE_V2

### spec.nodeConfig.linuxNodeConfig.hugepagesConfig

`GcpGkeNodePoolHugepagesConfig`

Hugepage pre-allocation for DPDK/database workloads.

### spec.nodeConfig.linuxNodeConfig.hugepagesConfig.hugepageSize2m

`int64` · optional (explicit presence)

Number of 2MB hugepages.

### spec.nodeConfig.linuxNodeConfig.hugepagesConfig.hugepageSize1g

`int64` · optional (explicit presence)

Number of 1GB hugepages.

### spec.nodeConfig.linuxNodeConfig.transparentHugepageEnabled

`string`

Kernel transparent hugepage mode for anonymous memory:
ALWAYS, MADVISE (only regions that request it), or NEVER.

- rule: transparent_hugepage_enabled must be empty, TRANSPARENT_HUGEPAGE_ENABLED_ALWAYS, TRANSPARENT_HUGEPAGE_ENABLED_MADVISE, TRANSPARENT_HUGEPAGE_ENABLED_NEVER, or TRANSPARENT_HUGEPAGE_ENABLED_UNSPECIFIED

### spec.nodeConfig.linuxNodeConfig.transparentHugepageDefrag

`string`

Kernel defrag behavior when allocating transparent hugepages:
ALWAYS (stall to reclaim), DEFER (kick background reclaim),
DEFER_WITH_MADVISE, MADVISE, or NEVER.

- rule: transparent_hugepage_defrag must be empty, TRANSPARENT_HUGEPAGE_DEFRAG_ALWAYS, TRANSPARENT_HUGEPAGE_DEFRAG_DEFER, TRANSPARENT_HUGEPAGE_DEFRAG_DEFER_WITH_MADVISE, TRANSPARENT_HUGEPAGE_DEFRAG_MADVISE, TRANSPARENT_HUGEPAGE_DEFRAG_NEVER, or TRANSPARENT_HUGEPAGE_DEFRAG_UNSPECIFIED

### spec.nodeConfig.linuxNodeConfig.nodeKernelModuleLoadingPolicy

`string`

Signed-kernel-module enforcement: ENFORCE_SIGNED_MODULES rejects
unsigned module loads (the hardened posture);
DO_NOT_ENFORCE_SIGNED_MODULES allows them.

- rule: node_kernel_module_loading_policy must be empty, POLICY_UNSPECIFIED, ENFORCE_SIGNED_MODULES, or DO_NOT_ENFORCE_SIGNED_MODULES

### spec.nodeConfig.linuxNodeConfig.enablePtpKvmTimeSync

`bool` · optional (explicit presence)

PTP/KVM paravirtual clock sync for sub-millisecond time accuracy —
for workloads needing precise cross-node timestamps (trading,
distributed tracing at fine granularity).

### spec.nodeConfig.linuxNodeConfig.swapConfig

`GcpGkeNodePoolSwapConfig`

Swap on the nodes (Kubernetes swap support): sizing profile plus
encryption. Swap trades OOM kills for latency under memory pressure;
pair with kubelet eviction tuning.

- rule: size swap with at most one profile: boot_disk_profile, dedicated_local_ssd_profile, or ephemeral_local_ssd_profile

### spec.nodeConfig.linuxNodeConfig.swapConfig.enabled

`bool` · optional (explicit presence)

Whether swap is enabled on the nodes.

### spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile

`GcpGkeNodePoolSwapSizing`

Swap carved from the boot disk.

- rule: set swap_size_gib or swap_size_percent, not both

### spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizeGib

`int64` · optional (explicit presence)

Absolute swap size in GiB.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.linuxNodeConfig.swapConfig.bootDiskProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage of the backing storage (1-100).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.nodeConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile

`GcpGkeNodePoolSwapDedicatedSsd`

Swap on local SSDs dedicated entirely to swap.

### spec.nodeConfig.linuxNodeConfig.swapConfig.dedicatedLocalSsdProfile.diskCount

`int64` · optional (explicit presence)

Number of local SSDs dedicated to swap.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile

`GcpGkeNodePoolSwapSizing`

Swap carved from the ephemeral-storage local SSDs.

- rule: set swap_size_gib or swap_size_percent, not both

### spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizeGib

`int64` · optional (explicit presence)

Absolute swap size in GiB.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.linuxNodeConfig.swapConfig.ephemeralLocalSsdProfile.swapSizePercent

`int32` · optional (explicit presence)

Swap size as a percentage of the backing storage (1-100).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.nodeConfig.linuxNodeConfig.swapConfig.encryptionConfig

`GcpGkeNodePoolSwapEncryption`

Swap encryption (on by default; disabling trades confidentiality for
a little throughput).

### spec.nodeConfig.linuxNodeConfig.swapConfig.encryptionConfig.disabled

`bool` · optional (explicit presence)

Set true to DISABLE swap encryption (encrypted by default).

### spec.nodeConfig.linuxNodeConfig.customNodeInit

`GcpGkeNodePoolCustomNodeInit`

A custom initialization script GKE runs on every node at boot,
before the node joins the cluster — for host-level setup that no
DaemonSet can do (kernel parameters that need a reboot-free apply,
vendor agents, custom certificates). Sourced from Cloud Storage or
Secret Manager; exactly one source.

- rule: set exactly one of gcs_uri or secret_manager_secret_uri for the init script
- rule: gcs_generation pins a Cloud Storage object version — set gcs_uri alongside it

### spec.nodeConfig.linuxNodeConfig.customNodeInit.gcsUri

`string`

Cloud Storage object holding the script, e.g.
"gs://my-bucket/node-init.sh". The node's service account needs read
access to the object.

- rule: gcs_uri must be a gs:// object URI

### spec.nodeConfig.linuxNodeConfig.customNodeInit.gcsGeneration

`int64` · optional (explicit presence)

Pin the Cloud Storage object to one generation so a later upload
does not silently change what new nodes run. Leave unset to always
fetch the current object; sent only when set because the API records
the generation it resolved.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.linuxNodeConfig.customNodeInit.secretManagerSecretUri

`string`

Secret Manager secret version holding the script, e.g.
"projects/P/secrets/node-init/versions/latest" — for scripts that
embed credentials. The node's service account needs
secretmanager.versions.access on it.

- rule: secret_manager_secret_uri must be a Secret Manager version resource name: projects/{project}/secrets/{secret}/versions/{version}

### spec.nodeConfig.loggingVariant

`string`

Node system-log throughput: DEFAULT (100 KiB/s) or MAX_THROUGHPUT
(10 MiB/s, costs a little node CPU) for pools with log-heavy workloads.

- rule: logging_variant must be empty, DEFAULT, or MAX_THROUGHPUT

### spec.nodeConfig.flexStart

`bool`

Flex-start (Dynamic Workload Scheduler): nodes are requested when
needed and run up to 7 days at a discount — the on-demand counterpart
to queued provisioning for hard-to-get GPU capacity. Immutable.

### spec.nodeConfig.maxRunDuration

`string`

Maximum runtime of each node, in seconds format (e.g. "3600s"), after
which it is drained and deleted. Pairs with flex-start/Spot batch
pools; leave empty for long-lived pools. Immutable.

- rule: max_run_duration must be a seconds-format duration like "3600s"

### spec.nodeConfig.enableConfidentialStorage

`bool`

Confidential storage on the node disks (hardware-encrypted at the
storage layer; pairs with confidential_nodes for end-to-end
confidentiality). Requires a supporting machine family and hyperdisk
boot disks. Immutable.

### spec.nodeConfig.localSsdEncryptionMode

`string`

Encryption mode for local SSDs: STANDARD_ENCRYPTION (Google-managed)
or EPHEMERAL_KEY_ENCRYPTION (per-node ephemeral keys destroyed with
the node — local data is unrecoverable after preemption/deletion).
Immutable.

- rule: local_ssd_encryption_mode must be empty, STANDARD_ENCRYPTION, or EPHEMERAL_KEY_ENCRYPTION

### spec.nodeConfig.gpudirectStrategy

`string`

GPUDirect strategy for multi-GPU/multi-node communication (e.g.
"GPUDIRECT_TCPX", "GPUDIRECT_TCPXO", "GPUDIRECT_RDMA"). GKE validates
the strategy against the machine family and GPU type at apply; the
provider passes the value through case-insensitively.

### spec.nodeConfig.nodeGroup

`string`

Sole-tenant node group to schedule this pool's nodes onto (nodes on
dedicated physical servers you already provisioned). Prefer
sole_tenant_config's affinity form for flexible matching. Immutable.

### spec.nodeConfig.storagePools

`[]string`

Hyperdisk storage pools the node boot disks are provisioned in —
pre-purchased pooled capacity/IOPS shared across disks. Full resource
paths. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.nodeConfig.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the node VMs (org-policy/firewall
tags, distinct from network tags and labels), as
{"tagKeys/123": "tagValues/456"} pairs. Mutable.

### spec.nodeConfig.advancedMachineFeatures

`GcpGkeNodePoolAdvancedMachineFeatures`

Advanced machine features: SMT control, nested virtualization, and
the performance monitoring unit. Immutable.

### spec.nodeConfig.advancedMachineFeatures.threadsPerCore

`int64`

Threads per physical core: 1 disables SMT (licensing or
side-channel isolation), 2 keeps it on. Only 1 or 2 are meaningful.

- rule: {"int64":{"lte":"2","gte":"1"}}

### spec.nodeConfig.advancedMachineFeatures.enableNestedVirtualization

`bool` · optional (explicit presence)

Nested virtualization on the nodes (running VMs inside pods, e.g.
KubeVirt). Requires Haswell-or-newer non-shared-core machine types.

### spec.nodeConfig.advancedMachineFeatures.performanceMonitoringUnit

`string`

Performance monitoring unit exposure: ARCHITECTURAL (basic counters),
STANDARD (most counters), or ENHANCED (all counters) — for profiling
workloads that read hardware counters.

- rule: performance_monitoring_unit must be empty, ARCHITECTURAL, STANDARD, or ENHANCED

### spec.nodeConfig.bootDisk

`GcpGkeNodePoolBootDisk`

Boot disk shape as a first-class block — required for hyperdisk
tuning (provisioned IOPS/throughput). When set, its disk_type/size
take precedence over the flat disk_type/disk_size_gb fields.

### spec.nodeConfig.bootDisk.diskType

`string`

Boot disk type (pd-standard, pd-balanced, pd-ssd,
hyperdisk-balanced...). Hyperdisk types unlock provisioned_iops /
provisioned_throughput below.

- rule: disk_type must be empty, pd-standard, pd-balanced, pd-ssd, hyperdisk-balanced, hyperdisk-extreme, or hyperdisk-throughput

### spec.nodeConfig.bootDisk.sizeGb

`int64` · optional (explicit presence)

Boot disk size in GB (min 10).

- rule: {"int64":{"gte":"10"}}

### spec.nodeConfig.bootDisk.provisionedIops

`int64` · optional (explicit presence)

Provisioned IOPS — hyperdisk types that support IOPS provisioning
only.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.bootDisk.provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MiB/s — hyperdisk types that support
throughput provisioning only.

- rule: {"int64":{"gte":"1"}}

### spec.nodeConfig.nodeImage

`GcpGkeNodePoolNodeImage`

Custom node OS image family, overriding the stock image_type image.
Both fields must point at an image built for GKE nodes. Immutable.

### spec.nodeConfig.nodeImage.image

`string` · required

Image family or full image name to boot nodes from.

- rule: {"required":true}

### spec.nodeConfig.nodeImage.imageProject

`string`

Project hosting the image. Empty means the pool's own project.

### spec.nodeConfig.soleTenantConfig

`GcpGkeNodePoolSoleTenantConfig`

Sole-tenant scheduling: affinity rules selecting the sole-tenant node
groups this pool's nodes run on, plus a dedicated-vCPU floor.
Immutable.

### spec.nodeConfig.soleTenantConfig.nodeAffinities

`[]GcpGkeNodePoolSoleTenantAffinity` · required

Affinity rules matching sole-tenant node groups. At least one rule
selects the groups (e.g. key "compute.googleapis.com/node-group-name",
operator IN, values [group name]).

- rule: {"repeated":{"minItems":"1"}}

### spec.nodeConfig.soleTenantConfig.nodeAffinities[].key

`string` · required

Affinity label key, e.g. "compute.googleapis.com/node-group-name".

- rule: {"required":true}

### spec.nodeConfig.soleTenantConfig.nodeAffinities[].operator

`string` · required

IN selects hosts matching values; NOT_IN avoids them.

- rule: {"required":true,"string":{"in":["IN","NOT_IN"]}}

### spec.nodeConfig.soleTenantConfig.nodeAffinities[].values

`[]string` · required

Values matched against the key.

- rule: {"repeated":{"minItems":"1"}}

### spec.nodeConfig.soleTenantConfig.minNodeCpus

`int32` · optional (explicit presence)

Minimum dedicated vCPUs per node reserved for this pool's workloads
on the shared physical host.

- rule: {"int32":{"gte":1}}

### spec.nodeConfig.sandboxType

`string`

GKE Sandbox: runs every pod in this pool inside gVisor (user-space
kernel isolation) — for untrusted/multi-tenant workloads. The only
value is "GVISOR". GKE taints sandbox pools automatically. Immutable.

- rule: sandbox_type must be empty or GVISOR

### spec.nodeConfig.windowsOsVersion

`string`

Windows Server version for WINDOWS_LTSC_CONTAINERD pools:
OS_VERSION_LTSC2019 or OS_VERSION_LTSC2022. Immutable.

- rule: windows_os_version must be empty, OS_VERSION_LTSC2019, or OS_VERSION_LTSC2022

### spec.nodeConfig.hostMaintenanceInterval

`string`

Host maintenance cadence for the underlying physical hosts:
AS_NEEDED (default) or PERIODIC (predictable windows — required by
some GPU/TPU shapes). Immutable.

- rule: host_maintenance_interval must be empty, AS_NEEDED, or PERIODIC

### spec.nodeConfig.architectureTaintBehavior

`string`

How GKE taints nodes based on CPU architecture: ARM applies the
kubernetes.io/arch taint to Arm nodes (the default protection that
keeps amd64-only workloads off T2A/Axion pools); NONE disables that
automatic taint.

- rule: architecture_taint_behavior must be empty, NONE, or ARM

### spec.nodeConfig.containerdConfig

`GcpGkeNodePoolContainerdConfig`

containerd runtime configuration: private registry access (custom CA
domains), per-registry host overrides (mirrors, auth, headers), and
writable cgroups.

### spec.nodeConfig.containerdConfig.privateRegistryAccess

`GcpGkeNodePoolPrivateRegistryAccess`

Trust custom certificate authorities for specific registry domains —
required for private registries with self-signed/internal CAs.

### spec.nodeConfig.containerdConfig.privateRegistryAccess.enabled

`bool`

Master toggle for private registry access configuration.

### spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains

`[]GcpGkeNodePoolRegistryCaDomain`

Per-domain CA trust: each entry names registry FQDNs and the Secret
Manager secret holding the CA certificate for them.

### spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].fqdns

`[]string` · required

Registry FQDNs this CA vouches for, e.g. ["registry.internal:5000"].

- rule: {"repeated":{"minItems":"1"}}

### spec.nodeConfig.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].gcpSecretManagerCertificateUri

`string` · required

Secret Manager secret URI holding the CA certificate, in the form
"projects/{project}/secrets/{secret}/versions/{version}".

- rule: {"required":true}

### spec.nodeConfig.containerdConfig.registryHosts

`[]GcpGkeNodePoolRegistryHost`

Per-registry host overrides: mirrors, capabilities, dial timeouts,
client certificates, and custom headers, in containerd hosts.toml
semantics.

### spec.nodeConfig.containerdConfig.registryHosts[].server

`string` · required

The registry server the overrides apply to, e.g. "docker.io" or
"registry.internal:5000".

- rule: {"required":true}

### spec.nodeConfig.containerdConfig.registryHosts[].hosts

`[]GcpGkeNodePoolRegistryHostEndpoint`

Host endpoints serving this registry (mirrors first, in order).

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].host

`string` · required

Endpoint URL, e.g. "https://mirror.internal".

- rule: {"required":true}

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].capabilities

`[]string`

Operations this endpoint can serve — typically "pull" and
"resolve" (containerd hosts.toml capability names). GKE validates
the values at apply.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true}}

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].dialTimeout

`string`

Dial timeout for this endpoint, e.g. "10s".

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].overridePath

`bool` · optional (explicit presence)

Path override on the endpoint host (registry served under a
non-standard path).

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].caSecretUri

`string`

Secret Manager secret URI for the CA certificate to trust for this
endpoint.

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].clientCertSecretUri

`string`

Secret Manager secret URI for the client TLS certificate presented to
this endpoint.

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].clientKeySecretUri

`string`

Secret Manager secret URI for the client TLS key presented to this
endpoint.

### spec.nodeConfig.containerdConfig.registryHosts[].hosts[].headers

`map<string, string>`

Custom HTTP headers sent to this endpoint, e.g. authorization
headers for pull-through caches.

### spec.nodeConfig.containerdConfig.writableCgroupsEnabled

`bool` · optional (explicit presence)

Writable cgroup filesystem inside containers — for workloads that
manage their own sub-cgroups (nested container runtimes, some JVMs).

### spec.deletionPolicy

`string`

Destroy-time stance of the IaC engines toward the pool itself:
DELETE (default) destroys it, PREVENT fails any plan that would
destroy it, ABANDON removes it from state and leaves the pool
running in GCP. This is an engine-side control, not a GKE API field.

- rule: deletion_policy must be empty, DELETE, PREVENT, or ABANDON

### spec.ignoreNodeCountChanges

`bool`

Skips the per-pool Instance Group Manager queries that reconcile the
observed node count on every plan — a quota/performance optimization
for very large pools. While true, node-count drift is invisible to
plans and the instance_group_urls outputs go stale. The engines
already never fight the autoscaler over node_count; this additionally
silences the read-side queries.

### spec.nodeDrainConfig

`GcpGkeNodePoolNodeDrainConfig`

How nodes drain when the POOL ITSELF is deleted or replaced: grace
periods and whether PodDisruptionBudgets are honored during the
teardown. Distinct from upgrade_settings, which paces upgrades of a
pool that continues to exist. NOTE: customized node drain requires
project-level enablement from GCP support — on a project without it,
the API rejects the create with "customized node drain timeout is not
enabled for this project, please contact your account manager or open
a support case to enable it".

### spec.nodeDrainConfig.graceTerminationDuration

`string`

Grace period each node gets to finish draining before it is removed,
seconds format, e.g. "300s".

- rule: grace_termination_duration must be a seconds-format duration like "300s"

### spec.nodeDrainConfig.pdbTimeoutDuration

`string`

How long the drain waits on a blocking PodDisruptionBudget before
proceeding anyway, seconds format, e.g. "3600s".

- rule: pdb_timeout_duration must be a seconds-format duration like "3600s"

### spec.nodeDrainConfig.respectPdbDuringNodePoolDeletion

`bool` · optional (explicit presence)

Honor PodDisruptionBudgets while the pool is being deleted (bounded
by pdb_timeout_duration) instead of evicting immediately.

### spec.excludeUpgradesUntilEndOfSupport

`bool`

Holds this pool on its current Kubernetes version until that
version's end-of-support date, exempting it from GKE's automatic
upgrades (the cluster's maintenance windows and exclusions still
govern everything else). For a workload that must not move minor
versions until it has been re-qualified. GKE reports the resulting
exclusion window (start and end) in the node pool's status; when the
version reaches end of support the exclusion lapses and upgrades
resume.

## Validation Rules

- `initial_node_count_requires_autoscaling`: initial_node_count only applies to autoscaled pools — a fixed-size pool's size is node_count itself
- `name_xor_name_prefix`: set node_pool_name or name_prefix, not both — a prefixed pool gets its full name from GKE

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeNodePool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.node_pool_name` | `string` | Name of the node pool as created in GKE — the handle gcloud commands and in-cluster references use. Matches spec.node_pool_name when set, otherwise metadata.name. |
| `status.outputs.instance_group_urls` | `[]string` | Resource URLs of the managed instance groups backing this pool — one per zone for regional clusters. What instance-group-targeted load-balancer backends and infrastructure automation compose against. |
| `status.outputs.min_nodes` | `string` | Effective minimum size of the pool: the autoscaling minimum (per zone) when autoscaling manages the pool, else the fixed node_count. |
| `status.outputs.max_nodes` | `string` | Effective maximum size of the pool: the autoscaling maximum (per zone) when autoscaling manages the pool, else the fixed node_count. |
| `status.outputs.current_node_count` | `string` | Number of nodes per zone at the time of the last deploy. For autoscaled pools this drifts as the autoscaler works — treat it as a snapshot, not live state. |
| `status.outputs.node_pool_id` | `string` | Fully qualified GKE node pool resource ID: projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{name}. The handle downstream services (e.g. Dataproc on GKE) reference the pool by, and the path API callers address it with. |
| `status.outputs.location` | `string` | The pool's location (the parent cluster's region or zone), exactly as provided in the spec. |
| `status.outputs.version` | `string` | The Kubernetes version running on the pool's nodes. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.clusterName` | GcpGkeCluster | `status.outputs.name` |
| `spec.location` | GcpGkeCluster | `status.outputs.location` |
| `spec.networkConfig.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.networkConfig.additionalNodeNetworks[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.networkConfig.additionalNodeNetworks[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.networkConfig.additionalPodNetworks[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.nodeConfig.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.nodeConfig.bootDiskKmsKey` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDataprocCluster | `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePool` | `status.outputs.node_pool_id` |

## See Also

- [Overview](../README.md)

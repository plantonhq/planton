# GcpGkeCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGkeClusterSpec defines a GKE cluster (`google_container_cluster`) — the
Kubernetes control plane plus cluster-wide configuration.

One GcpGkeCluster resource is one cluster. Node pools are separate
composable resources — GcpGkeNodePool — referencing this cluster by its
name output; the cluster's default node pool is always removed at create
time so every node pool is an explicitly managed, first-class node.
In Autopilot mode (enable_autopilot) GKE manages nodes entirely and no
GcpGkeNodePool resources are attached.

Clusters are always VPC-native (alias-IP): pods and services draw from
secondary ranges on the referenced subnetwork — either ranges you name in
ip_allocation, or ranges GKE creates and manages when none are named.
Legacy routes-based networking is deliberately not modeled.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeCluster
metadata:
  name: test-gke-cluster
spec:
  projectId:
    value: test-project-123
  location: us-central1
  description: hack manifest exercising the broad cluster surface offline
  network:
    value: projects/test-project-123/global/networks/test-vpc
  subnetwork:
    value: projects/test-project-123/regions/us-central1/subnetworks/test-subnet
  nodeLocations:
    - us-central1-a
    - us-central1-b
  deletionProtection: false
  ipAllocation:
    clusterSecondaryRangeName:
      value: pods-range
    servicesSecondaryRangeName:
      value: services-range
  datapathProvider: ADVANCED_DATAPATH
  privateCluster:
    enablePrivateNodes: true
    masterIpv4CidrBlock: "172.16.0.16/28"
    enableMasterGlobalAccess: true
  masterAuthorizedNetworks:
    cidrBlocks:
      - cidrBlock: 203.0.113.0/24
        displayName: office
  controlPlaneEndpoints:
    dnsEndpointAllowExternalTraffic: true
    enableK8sTokensViaDns: true
  issueClientCertificate: false
  nodeCreationMode: VIA_CONTROL_PLANE
  releaseChannel: REGULAR
  gkeAutoUpgradePatchMode: ACCELERATED
  maintenancePolicy:
    dailyWindow:
      startTime: "03:00"
    exclusions:
      - exclusionName: year-end-freeze
        startTime: "2026-12-20T00:00:00Z"
        endTime: "2027-01-05T00:00:00Z"
        scope: NO_MINOR_UPGRADES
        endTimeBehavior: UNTIL_END_OF_SUPPORT
    disruptionBudget:
      patchVersionDisruptionInterval: "604800s"
  clusterAutoscaling:
    enabled: true
    resourceLimits:
      - resourceType: cpu
        minimum: 4
        maximum: 64
      - resourceType: memory
        minimum: 16
        maximum: 256
    autoProvisioningDefaults:
      diskType: pd-balanced
      upgradeSettings:
        strategy: SURGE
        maxSurge: 1
        maxUnavailable: 0
  enableVerticalPodAutoscaling: true
  databaseEncryption:
    state: ENCRYPTED
    keyName:
      value: projects/test-project-123/locations/us-central1/keyRings/test-ring/cryptoKeys/etcd-key
  securityPosture:
    mode: BASIC
    vulnerabilityMode: VULNERABILITY_BASIC
  logging:
    components:
      - SYSTEM_COMPONENTS
      - WORKLOADS
  monitoring:
    components:
      - SYSTEM_COMPONENTS
    managedPrometheusEnabled: true
  enableCostManagement: true
  enableSecretManagerCsi: true
  secretManagerRotation:
    enabled: true
    rotationInterval: "120s"
  secretSync:
    enabled: true
    rotationEnabled: true
    rotationInterval: "300s"
  rbacBindingConfig:
    enableInsecureBindingSystemAuthenticated: false
    enableInsecureBindingSystemUnauthenticated: false
  nodePoolDefaults:
    gcfsEnabled: true
    insecureKubeletReadonlyPortEnabled: "FALSE"
    loggingVariant: DEFAULT
  userManagedKeys:
    clusterCa: projects/test-project-123/locations/us-central1/caPools/test-cluster-ca
    controlPlaneDiskEncryptionKey:
      value: projects/test-project-123/locations/us-central1/keyRings/test-ring/cryptoKeys/cp-disk-key
  addons:
    gcsFuseCsiDriverEnabled: true
    gkeBackupAgentEnabled: true
    parallelstoreCsiDriverEnabled: true
    rayOperatorEnabled: true
    rayClusterLoggingEnabled: true
  skipNodePoolRefresh: true
  deletionPolicy: DELETE
  resourceLabels:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.clusterName` | `string` |  |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetwork` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.nodeLocations` | `[]string` |  |  |  |
| `spec.resourceLabels` | `map<string, string>` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.ipAllocation` | `GcpGkeClusterIpAllocation` |  |  |  |
| `spec.ipAllocation.clusterSecondaryRangeName` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.secondary_ranges.[*].range_name`) |
| `spec.ipAllocation.servicesSecondaryRangeName` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.secondary_ranges.[*].range_name`) |
| `spec.ipAllocation.clusterIpv4CidrBlock` | `string` |  |  |  |
| `spec.ipAllocation.servicesIpv4CidrBlock` | `string` |  |  |  |
| `spec.ipAllocation.stackType` | `string` |  | `IPV4` |  |
| `spec.ipAllocation.additionalPodRangeNames` | `[]string` |  |  |  |
| `spec.ipAllocation.podCidrOverprovisionDisabled` | `bool` |  |  |  |
| `spec.ipAllocation.additionalIpRanges` | `[]GcpGkeClusterAdditionalIpRange` |  |  |  |
| `spec.ipAllocation.additionalIpRanges[].subnetwork` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.ipAllocation.additionalIpRanges[].podIpv4RangeNames` | `[]string` |  |  |  |
| `spec.ipAllocation.additionalIpRanges[].status` | `string` |  |  |  |
| `spec.ipAllocation.autoIpamEnabled` | `bool` |  |  |  |
| `spec.ipAllocation.networkTier` | `string` |  |  |  |
| `spec.datapathProvider` | `string` |  |  |  |
| `spec.defaultMaxPodsPerNode` | `int32` |  |  |  |
| `spec.enableIntranodeVisibility` | `bool` |  |  |  |
| `spec.enableL4IlbSubsetting` | `bool` |  |  |  |
| `spec.enableFqdnNetworkPolicy` | `bool` |  |  |  |
| `spec.enableCiliumClusterwideNetworkPolicy` | `bool` |  |  |  |
| `spec.enableMultiNetworking` | `bool` |  |  |  |
| `spec.privateIpv6GoogleAccess` | `string` |  |  |  |
| `spec.inTransitEncryption` | `string` |  |  |  |
| `spec.disableDefaultSnat` | `bool` |  |  |  |
| `spec.enableNetworkPolicy` | `bool` |  |  |  |
| `spec.dnsConfig` | `GcpGkeClusterDnsConfig` |  |  |  |
| `spec.dnsConfig.clusterDns` | `string` |  |  |  |
| `spec.dnsConfig.clusterDnsScope` | `string` |  |  |  |
| `spec.dnsConfig.clusterDnsDomain` | `string` |  |  |  |
| `spec.dnsConfig.additiveVpcScopeDnsDomain` | `string` |  |  |  |
| `spec.gatewayApiChannel` | `string` |  |  |  |
| `spec.enableServiceExternalIps` | `bool` |  |  |  |
| `spec.totalEgressBandwidthTier` | `string` |  |  |  |
| `spec.disableL4LbFirewallReconciliation` | `bool` |  |  |  |
| `spec.privateCluster` | `GcpGkeClusterPrivateCluster` |  |  |  |
| `spec.privateCluster.enablePrivateNodes` | `bool` |  |  |  |
| `spec.privateCluster.enablePrivateEndpoint` | `bool` |  |  |  |
| `spec.privateCluster.masterIpv4CidrBlock` | `string` |  |  |  |
| `spec.privateCluster.privateEndpointSubnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.privateCluster.enableMasterGlobalAccess` | `bool` |  |  |  |
| `spec.masterAuthorizedNetworks` | `GcpGkeClusterMasterAuthorizedNetworks` |  |  |  |
| `spec.masterAuthorizedNetworks.cidrBlocks` | `[]GcpGkeClusterMasterAuthorizedNetworkCidr` |  |  |  |
| `spec.masterAuthorizedNetworks.cidrBlocks[].cidrBlock` | `string` | yes |  |  |
| `spec.masterAuthorizedNetworks.cidrBlocks[].displayName` | `string` |  |  |  |
| `spec.masterAuthorizedNetworks.gcpPublicCidrsAccessEnabled` | `bool` |  |  |  |
| `spec.masterAuthorizedNetworks.privateEndpointEnforcementEnabled` | `bool` |  |  |  |
| `spec.controlPlaneEndpoints` | `GcpGkeClusterControlPlaneEndpoints` |  |  |  |
| `spec.controlPlaneEndpoints.dnsEndpointAllowExternalTraffic` | `bool` |  |  |  |
| `spec.controlPlaneEndpoints.ipEndpointsEnabled` | `bool` |  | `true` |  |
| `spec.controlPlaneEndpoints.enableK8sTokensViaDns` | `bool` |  |  |  |
| `spec.controlPlaneEndpoints.enableK8sCertsViaDns` | `bool` |  |  |  |
| `spec.releaseChannel` | `enum` |  | `REGULAR` |  |
| `spec.minMasterVersion` | `string` |  |  |  |
| `spec.maintenancePolicy` | `GcpGkeClusterMaintenancePolicy` |  |  |  |
| `spec.maintenancePolicy.dailyWindow` | `GcpGkeClusterDailyMaintenanceWindow` |  |  |  |
| `spec.maintenancePolicy.dailyWindow.startTime` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringWindow` | `GcpGkeClusterRecurringMaintenanceWindow` |  |  |  |
| `spec.maintenancePolicy.recurringWindow.startTime` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringWindow.endTime` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringWindow.recurrence` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringTimeWindow` | `GcpGkeClusterRecurringTimeMaintenanceWindow` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.windowStartTime` | `GcpGkeClusterTimeOfDay` | yes |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.windowStartTime.hours` | `int32` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.windowStartTime.minutes` | `int32` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.windowStartTime.seconds` | `int32` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.windowDuration` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.recurrence` | `string` | yes |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.delayUntil` | `GcpGkeClusterCalendarDate` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.delayUntil.year` | `int32` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.delayUntil.month` | `int32` |  |  |  |
| `spec.maintenancePolicy.recurringTimeWindow.delayUntil.day` | `int32` |  |  |  |
| `spec.maintenancePolicy.exclusions` | `[]GcpGkeClusterMaintenanceExclusion` |  |  |  |
| `spec.maintenancePolicy.exclusions[].exclusionName` | `string` | yes |  |  |
| `spec.maintenancePolicy.exclusions[].startTime` | `string` | yes |  |  |
| `spec.maintenancePolicy.exclusions[].endTime` | `string` | yes |  |  |
| `spec.maintenancePolicy.exclusions[].scope` | `string` |  |  |  |
| `spec.maintenancePolicy.exclusions[].endTimeBehavior` | `string` |  |  |  |
| `spec.maintenancePolicy.disruptionBudget` | `GcpGkeClusterDisruptionBudget` |  |  |  |
| `spec.maintenancePolicy.disruptionBudget.minorVersionDisruptionInterval` | `string` |  |  |  |
| `spec.maintenancePolicy.disruptionBudget.patchVersionDisruptionInterval` | `string` |  |  |  |
| `spec.clusterAutoscaling` | `GcpGkeClusterAutoscaling` |  |  |  |
| `spec.clusterAutoscaling.enabled` | `bool` |  |  |  |
| `spec.clusterAutoscaling.resourceLimits` | `[]GcpGkeClusterAutoscalingResourceLimit` |  |  |  |
| `spec.clusterAutoscaling.resourceLimits[].resourceType` | `string` | yes |  |  |
| `spec.clusterAutoscaling.resourceLimits[].minimum` | `int64` |  |  |  |
| `spec.clusterAutoscaling.resourceLimits[].maximum` | `int64` |  |  |  |
| `spec.clusterAutoscaling.autoscalingProfile` | `string` |  | `BALANCED` |  |
| `spec.clusterAutoscaling.autoProvisioningLocations` | `[]string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults` | `GcpGkeClusterAutoProvisioningDefaults` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.clusterAutoscaling.autoProvisioningDefaults.oauthScopes` | `[]string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.diskSizeGb` | `int32` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.diskType` | `string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.imageType` | `string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.minCpuPlatform` | `string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.bootDiskKmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.clusterAutoscaling.autoProvisioningDefaults.enableSecureBoot` | `bool` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.enableIntegrityMonitoring` | `bool` |  | `true` |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.autoUpgrade` | `bool` |  | `true` |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.autoRepair` | `bool` |  | `true` |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings` | `GcpGkeClusterNapUpgradeSettings` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.maxSurge` | `uint32` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.maxUnavailable` | `uint32` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.strategy` | `string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings` | `GcpGkeClusterNapBlueGreenSettings` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy` | `GcpGkeClusterNapStandardRolloutPolicy` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchPercentage` | `float` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchNodeCount` | `uint32` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchSoakDuration` | `string` |  |  |  |
| `spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.nodePoolSoakDuration` | `string` |  |  |  |
| `spec.clusterAutoscaling.defaultComputeClassEnabled` | `bool` |  |  |  |
| `spec.enableVerticalPodAutoscaling` | `bool` |  |  |  |
| `spec.hpaProfile` | `string` |  |  |  |
| `spec.workloadIdentityEnabled` | `bool` |  | `true` |  |
| `spec.enableShieldedNodes` | `bool` |  |  |  |
| `spec.databaseEncryption` | `GcpGkeClusterDatabaseEncryption` |  |  |  |
| `spec.databaseEncryption.state` | `string` | yes |  |  |
| `spec.databaseEncryption.keyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.binaryAuthorizationEvaluationMode` | `string` |  |  |  |
| `spec.securityPosture` | `GcpGkeClusterSecurityPosture` |  |  |  |
| `spec.securityPosture.mode` | `string` |  |  |  |
| `spec.securityPosture.vulnerabilityMode` | `string` |  |  |  |
| `spec.authenticatorSecurityGroup` | `string` |  |  |  |
| `spec.enableLegacyAbac` | `bool` |  |  |  |
| `spec.enableMeshCertificates` | `bool` |  |  |  |
| `spec.enableSecretManagerCsi` | `bool` |  |  |  |
| `spec.confidentialNodes` | `GcpGkeClusterConfidentialNodes` |  |  |  |
| `spec.confidentialNodes.enabled` | `bool` |  |  |  |
| `spec.confidentialNodes.confidentialInstanceType` | `string` |  |  |  |
| `spec.anonymousAuthenticationMode` | `string` |  |  |  |
| `spec.enableIdentityService` | `bool` |  |  |  |
| `spec.logging` | `GcpGkeClusterLogging` |  |  |  |
| `spec.logging.components` | `[]string` |  |  |  |
| `spec.logging.enabled` | `bool` |  | `true` |  |
| `spec.monitoring` | `GcpGkeClusterMonitoring` |  |  |  |
| `spec.monitoring.components` | `[]string` |  |  |  |
| `spec.monitoring.managedPrometheusEnabled` | `bool` |  | `true` |  |
| `spec.monitoring.autoMonitoringScope` | `string` |  |  |  |
| `spec.monitoring.advancedDatapathMetricsEnabled` | `bool` |  |  |  |
| `spec.monitoring.advancedDatapathRelayEnabled` | `bool` |  |  |  |
| `spec.notificationPubsub` | `GcpGkeClusterNotificationPubSub` |  |  |  |
| `spec.notificationPubsub.enabled` | `bool` |  |  |  |
| `spec.notificationPubsub.topic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.notificationPubsub.eventTypes` | `[]string` |  |  |  |
| `spec.enableCostManagement` | `bool` |  |  |  |
| `spec.resourceUsageExport` | `GcpGkeClusterResourceUsageExport` |  |  |  |
| `spec.resourceUsageExport.bigqueryDatasetId` | `string \| valueFrom` | yes |  | GcpBigQueryDataset (`status.outputs.dataset_id`) |
| `spec.resourceUsageExport.enableNetworkEgressMetering` | `bool` |  |  |  |
| `spec.resourceUsageExport.enableResourceConsumptionMetering` | `bool` |  | `true` |  |
| `spec.addons` | `GcpGkeClusterAddons` |  |  |  |
| `spec.addons.httpLoadBalancingEnabled` | `bool` |  | `true` |  |
| `spec.addons.horizontalPodAutoscalingEnabled` | `bool` |  | `true` |  |
| `spec.addons.gcePersistentDiskCsiDriverEnabled` | `bool` |  | `true` |  |
| `spec.addons.gcpFilestoreCsiDriverEnabled` | `bool` |  |  |  |
| `spec.addons.gcsFuseCsiDriverEnabled` | `bool` |  |  |  |
| `spec.addons.gkeBackupAgentEnabled` | `bool` |  |  |  |
| `spec.addons.dnsCacheEnabled` | `bool` |  |  |  |
| `spec.addons.configConnectorEnabled` | `bool` |  |  |  |
| `spec.addons.statefulHaEnabled` | `bool` |  |  |  |
| `spec.addons.rayOperatorEnabled` | `bool` |  |  |  |
| `spec.addons.rayClusterLoggingEnabled` | `bool` |  |  |  |
| `spec.addons.rayClusterMonitoringEnabled` | `bool` |  |  |  |
| `spec.addons.cloudrunEnabled` | `bool` |  |  |  |
| `spec.addons.cloudrunLoadBalancerType` | `string` |  |  |  |
| `spec.addons.parallelstoreCsiDriverEnabled` | `bool` |  |  |  |
| `spec.addons.lustreCsiDriverEnabled` | `bool` |  |  |  |
| `spec.addons.lustreCsiLegacyPortEnabled` | `bool` |  |  |  |
| `spec.addons.lustreCsiDisableMultiNic` | `bool` |  |  |  |
| `spec.addons.podSnapshotEnabled` | `bool` |  |  |  |
| `spec.addons.agentSandboxEnabled` | `bool` |  |  |  |
| `spec.addons.sliceControllerEnabled` | `bool` |  |  |  |
| `spec.addons.slurmOperatorEnabled` | `bool` |  |  |  |
| `spec.addons.highScaleCheckpointingEnabled` | `bool` |  |  |  |
| `spec.addons.nodeReadinessControllerEnabled` | `bool` |  |  |  |
| `spec.enableAutopilot` | `bool` |  |  |  |
| `spec.allowNetAdmin` | `bool` |  |  |  |
| `spec.fleetProject` | `string` |  |  |  |
| `spec.fleetMembershipType` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.ignoreNodeCountChanges` | `bool` |  |  |  |
| `spec.skipNodePoolRefresh` | `bool` |  |  |  |
| `spec.enableKubernetesAlpha` | `bool` |  |  |  |
| `spec.k8sBetaApis` | `[]string` |  |  |  |
| `spec.dataplaneOptimizationMode` | `string` |  |  |  |
| `spec.issueClientCertificate` | `bool` |  |  |  |
| `spec.nodeCreationMode` | `string` |  |  |  |
| `spec.gkeAutoUpgradePatchMode` | `string` |  |  |  |
| `spec.rbacBindingConfig` | `GcpGkeClusterRbacBindingConfig` |  |  |  |
| `spec.rbacBindingConfig.enableInsecureBindingSystemAuthenticated` | `bool` |  |  |  |
| `spec.rbacBindingConfig.enableInsecureBindingSystemUnauthenticated` | `bool` |  |  |  |
| `spec.autopilotPolicy` | `GcpGkeClusterAutopilotPolicy` |  |  |  |
| `spec.autopilotPolicy.noStandardNodePools` | `bool` |  |  |  |
| `spec.autopilotPolicy.noSystemImpersonation` | `bool` |  |  |  |
| `spec.autopilotPolicy.noSystemMutation` | `bool` |  |  |  |
| `spec.autopilotPolicy.noUnsafeWebhooks` | `bool` |  |  |  |
| `spec.autopilotPrivilegedAdmissionPaths` | `[]string` |  |  |  |
| `spec.nodePoolAutoConfig` | `GcpGkeClusterNodePoolAutoConfig` |  |  |  |
| `spec.nodePoolAutoConfig.networkTags` | `[]string` |  |  |  |
| `spec.nodePoolAutoConfig.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.nodePoolAutoConfig.cgroupMode` | `string` |  |  |  |
| `spec.nodePoolAutoConfig.nodeKernelModuleLoadingPolicy` | `string` |  |  |  |
| `spec.nodePoolAutoConfig.insecureKubeletReadonlyPortEnabled` | `string` |  |  |  |
| `spec.nodePoolDefaults` | `GcpGkeClusterNodePoolDefaults` |  |  |  |
| `spec.nodePoolDefaults.gcfsEnabled` | `bool` |  |  |  |
| `spec.nodePoolDefaults.insecureKubeletReadonlyPortEnabled` | `string` |  |  |  |
| `spec.nodePoolDefaults.loggingVariant` | `string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig` | `GcpGkeClusterContainerdDefaults` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.privateRegistryAccess` | `GcpGkeClusterPrivateRegistryAccess` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.enabled` | `bool` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains` | `[]GcpGkeClusterRegistryCaDomain` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].fqdns` | `[]string` | yes |  |  |
| `spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].gcpSecretManagerCertificateUri` | `string` | yes |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts` | `[]GcpGkeClusterRegistryHost` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].server` | `string` | yes |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts` | `[]GcpGkeClusterRegistryHostEndpoint` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].host` | `string` | yes |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].capabilities` | `[]string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].dialTimeout` | `string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].overridePath` | `bool` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].caSecretUri` | `string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].clientCertSecretUri` | `string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].clientKeySecretUri` | `string` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].headers` | `map<string, string>` |  |  |  |
| `spec.nodePoolDefaults.containerdConfig.writableCgroupsEnabled` | `bool` |  |  |  |
| `spec.userManagedKeys` | `GcpGkeClusterUserManagedKeys` |  |  |  |
| `spec.userManagedKeys.clusterCa` | `string` |  |  |  |
| `spec.userManagedKeys.etcdApiCa` | `string` |  |  |  |
| `spec.userManagedKeys.etcdPeerCa` | `string` |  |  |  |
| `spec.userManagedKeys.aggregationCa` | `string` |  |  |  |
| `spec.userManagedKeys.controlPlaneDiskEncryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.userManagedKeys.gkeopsEtcdBackupEncryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.userManagedKeys.serviceAccountSigningKeys` | `[]string` |  |  |  |
| `spec.userManagedKeys.serviceAccountVerificationKeys` | `[]string` |  |  |  |
| `spec.secretManagerRotation` | `GcpGkeClusterSecretRotation` |  |  |  |
| `spec.secretManagerRotation.enabled` | `bool` |  |  |  |
| `spec.secretManagerRotation.rotationInterval` | `string` |  |  |  |
| `spec.secretSync` | `GcpGkeClusterSecretSync` |  |  |  |
| `spec.secretSync.enabled` | `bool` |  |  |  |
| `spec.secretSync.rotationEnabled` | `bool` |  |  |  |
| `spec.secretSync.rotationInterval` | `string` |  |  |  |
| `spec.rollbackSafeUpgrade` | `GcpGkeClusterRollbackSafeUpgrade` |  |  |  |
| `spec.rollbackSafeUpgrade.controlPlaneSoakDuration` | `string` |  |  |  |
| `spec.desiredEmulatedVersion` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project in which to create this cluster.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Example: "my-prod-project-123"

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.clusterName

`string`

Name of the GKE cluster in GCP. Immutable. If not specified, defaults to
metadata.name. Must be 1-40 characters: lowercase letters, digits, and
hyphens; starting with a letter and ending with a letter or digit.
Example: "prod-primary"

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$"}}

### spec.location

`string` · required

Location of the cluster's control plane. Immutable. A region
("us-central1") creates a REGIONAL cluster — control-plane replicas in
three zones, higher availability, and node_locations defaulting to all
zones in the region. A zone ("us-central1-a") creates a ZONAL cluster —
a single control-plane instance, cheaper, with a brief control-plane
outage during upgrades. Production clusters should be regional.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+(-[a-z])?$"}}

### spec.description

`string`

Human-readable description of the cluster. Immutable.

### spec.network

`string | valueFrom` · required

The VPC network the cluster lives in. Accepts a network self link or a
reference to a GcpVpcNetwork resource. Immutable. An explicit network is
deliberately required — clusters on the legacy auto-created "default"
network do not compose into reviewable infrastructure.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom` · required

The subnetwork nodes are attached to. Accepts a subnetwork self link or
a reference to a GcpSubnetwork resource. Immutable. Must be in the same
region as the cluster location. Pod and service ranges come from this
subnetwork's secondary ranges (see ip_allocation).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.nodeLocations

`[]string`

Zones in which nodes (not the control plane) run, e.g.
["us-central1-a", "us-central1-b"]. For a regional cluster this narrows
node placement from all zones in the region; for a zonal cluster it adds
node zones beyond the control-plane zone (multi-zonal). Mutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}}}

### spec.resourceLabels

`map<string, string>`

GCE resource labels applied to the cluster (merged with the standard
platform labels). Mutable. These are cloud-billing/inventory labels, not
Kubernetes object labels.

### spec.deletionProtection

`bool` · optional (explicit presence)

Engine-side delete guard: while true (the default, matching GCP), both
IaC engines refuse to destroy the cluster — the plan/preview fails until
this is set to false. A cluster deletion destroys every workload on it;
keep this on for anything real.

- default: `true`

### spec.ipAllocation

`GcpGkeClusterIpAllocation`

VPC-native pod/service IP allocation. If omitted, GKE creates and
manages the secondary ranges itself — fine for dev clusters; production
clusters should name planned ranges on the subnetwork for address-space
governance.

- rule: set either cluster_secondary_range_name (use an existing subnetwork range) or cluster_ipv4_cidr_block (GKE creates the range) — not both
- rule: set either services_secondary_range_name (use an existing subnetwork range) or services_ipv4_cidr_block (GKE creates the range) — not both

### spec.ipAllocation.clusterSecondaryRangeName

`string | valueFrom`

Name of an existing secondary range on the subnetwork for POD IPs.
Accepts a literal range name or a reference to a GcpSubnetwork
resource's secondary ranges. Immutable.

- references: GcpSubnetwork (`status.outputs.secondary_ranges.[*].range_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.secondary_ranges.[*].range_name}} -- a bare string does not parse

### spec.ipAllocation.servicesSecondaryRangeName

`string | valueFrom`

Name of an existing secondary range on the subnetwork for SERVICE
(ClusterIP) IPs. Immutable.

- references: GcpSubnetwork (`status.outputs.secondary_ranges.[*].range_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.secondary_ranges.[*].range_name}} -- a bare string does not parse

### spec.ipAllocation.clusterIpv4CidrBlock

`string`

CIDR block for a GKE-created pod range (e.g. "10.4.0.0/14"), or a
netmask size (e.g. "/14") to let GKE pick the space. Immutable.

### spec.ipAllocation.servicesIpv4CidrBlock

`string`

CIDR block for a GKE-created services range (e.g. "10.8.0.0/20"), or a
netmask size. Immutable.

### spec.ipAllocation.stackType

`string` · optional (explicit presence)

IP stack of the cluster: IPV4 (default) or IPV4_IPV6 (dual-stack;
requires a dual-stack subnetwork). Immutable.

- default: `IPV4`
- rule: stack_type must be IPV4 or IPV4_IPV6

### spec.ipAllocation.additionalPodRangeNames

`[]string`

Names of ADDITIONAL existing secondary ranges node pools may use for
pod IPs — the mechanism for growing pod address space after creation
(added ranges are mutable; the primary pod range is not).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.ipAllocation.podCidrOverprovisionDisabled

`bool`

Disables the 2x pod-CIDR overprovisioning GKE applies per node by
default — doubles node density per pod range at the cost of headroom
for pod churn. Immutable.

### spec.ipAllocation.additionalIpRanges

`[]GcpGkeClusterAdditionalIpRange`

Additional SUBNETWORKS whose secondary ranges node pools may draw pod
IPs from — grows pod address space across subnetworks (beyond
additional_pod_range_names, which only names more ranges on the
cluster's own subnetwork).

### spec.ipAllocation.additionalIpRanges[].subnetwork

`string | valueFrom` · required

The subnetwork carrying the ranges. Accepts a self link or a
reference to a GcpSubnetwork resource.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.ipAllocation.additionalIpRanges[].podIpv4RangeNames

`[]string`

Secondary range names on that subnetwork usable for pod IPs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.ipAllocation.additionalIpRanges[].status

`string`

Lifecycle status of the subnetwork for scheduling: set "DRAINING" to
stop NEW node pools from selecting it (existing pools keep running).

### spec.ipAllocation.autoIpamEnabled

`bool`

Automatic IP address management: GKE plans and allocates the pod and
service ranges itself (no manual CIDR planning). Immutable.

### spec.ipAllocation.networkTier

`string`

Network tier for the cluster's IP allocation (e.g. "PREMIUM" or
"STANDARD"). GKE validates accepted tiers at apply.

### spec.datapathProvider

`string`

The cluster dataplane. ADVANCED_DATAPATH is Dataplane V2 (eBPF/Cilium):
built-in NetworkPolicy enforcement without Calico, dataplane
observability, and FQDN/Cilium policy support — the recommended choice
for new clusters. LEGACY_DATAPATH is kube-proxy/iptables. Immutable.

- rule: datapath_provider must be empty, LEGACY_DATAPATH, or ADVANCED_DATAPATH

### spec.defaultMaxPodsPerNode

`int32` · optional (explicit presence)

Default maximum pods per node for the cluster (8-256, default 110).
Lower values shrink the per-node pod CIDR slice and stretch the pod
range across more nodes. Immutable; node pools can override it.

- rule: {"int32":{"lte":256,"gte":8}}

### spec.enableIntranodeVisibility

`bool`

Mirrors pod-to-pod traffic on the same node to the VPC dataplane, making
it visible to VPC flow logs and packet mirroring. Mutable.

### spec.enableL4IlbSubsetting

`bool`

GKE subsetting for internal L4 load balancers: backends become a subset
of nodes instead of all nodes, lifting the 250-node ILB ceiling for
large clusters. Enabling is one-way — it cannot be turned off without
recreating the cluster.

### spec.enableFqdnNetworkPolicy

`bool`

Allows FQDN (domain-name based) NetworkPolicy rules. Requires Dataplane
V2 (datapath_provider ADVANCED_DATAPATH). Mutable.

### spec.enableCiliumClusterwideNetworkPolicy

`bool`

Allows CiliumClusterwideNetworkPolicy objects (cluster-scoped L3/L4
policy). Requires Dataplane V2. Mutable.

### spec.enableMultiNetworking

`bool`

Enables multi-networking: pods can attach to additional node network
interfaces (Network/GKENetworkParamSet objects). Immutable; requires
Dataplane V2.

### spec.privateIpv6GoogleAccess

`string`

Outbound-only private IPv6 access from nodes/pods to Google services
(PRIVATE_IPV6_GOOGLE_ACCESS_TO_GOOGLE) or bidirectional
(PRIVATE_IPV6_GOOGLE_ACCESS_BIDIRECTIONAL); DISABLED turns it off.

- rule: private_ipv6_google_access must be empty, PRIVATE_IPV6_GOOGLE_ACCESS_DISABLED, PRIVATE_IPV6_GOOGLE_ACCESS_TO_GOOGLE, or PRIVATE_IPV6_GOOGLE_ACCESS_BIDIRECTIONAL

### spec.inTransitEncryption

`string`

Encrypts inter-node pod traffic transparently (Dataplane V2 only):
INTER_NODE_TRANSPARENT encrypts, IN_TRANSIT_ENCRYPTION_DISABLED does
not. Zero application changes; small latency cost.

- rule: in_transit_encryption must be empty, IN_TRANSIT_ENCRYPTION_DISABLED, or IN_TRANSIT_ENCRYPTION_INTER_NODE_TRANSPARENT

### spec.disableDefaultSnat

`bool`

Disables the default source NAT for pod IPs leaving the cluster —
required for some routable-pod (non-masquerade) network designs where
pod IPs must be preserved end to end. Mutable.

### spec.enableNetworkPolicy

`bool`

Enables Kubernetes NetworkPolicy enforcement via Calico (the legacy
enforcement path, paired with the network-policy addon). On Dataplane V2
clusters leave this off — NetworkPolicy is enforced natively. Mutable.

### spec.dnsConfig

`GcpGkeClusterDnsConfig`

Cluster DNS provider configuration (Cloud DNS instead of kube-dns).
If omitted, GKE uses its platform default.

- rule: additive_vpc_scope_dns_domain requires cluster_dns CLOUD_DNS
- rule: cluster_dns_scope applies to Cloud DNS — set cluster_dns to CLOUD_DNS

### spec.dnsConfig.clusterDns

`string`

In-cluster DNS provider: CLOUD_DNS (managed, no kube-dns pods to
scale), KUBE_DNS (explicit kube-dns), or PLATFORM_DEFAULT (GKE
chooses).

- rule: cluster_dns must be empty, PROVIDER_UNSPECIFIED, PLATFORM_DEFAULT, CLOUD_DNS, or KUBE_DNS

### spec.dnsConfig.clusterDnsScope

`string`

Cloud DNS scope: CLUSTER_SCOPE (records visible in-cluster only) or
VPC_SCOPE (cluster DNS records resolvable across the whole VPC —
enables VMs to resolve headless services etc.).

- rule: cluster_dns_scope must be empty, DNS_SCOPE_UNSPECIFIED, CLUSTER_SCOPE, or VPC_SCOPE

### spec.dnsConfig.clusterDnsDomain

`string`

Custom cluster DNS suffix (default "cluster.local").

### spec.dnsConfig.additiveVpcScopeDnsDomain

`string`

With CLUSTER_SCOPE Cloud DNS, ALSO publishes services under this domain
VPC-wide — cluster-scoped DNS plus selective VPC visibility.

### spec.gatewayApiChannel

`string`

Gateway API support: CHANNEL_STANDARD installs the Gateway API CRDs
and the GKE Gateway controller (the successor to Ingress);
CHANNEL_EXPERIMENTAL adds experimental-channel CRDs;
CHANNEL_DISABLED turns it off. Mutable.

- rule: gateway_api_channel must be empty, CHANNEL_DISABLED, CHANNEL_EXPERIMENTAL, or CHANNEL_STANDARD

### spec.enableServiceExternalIps

`bool`

Allows Services of type LoadBalancer/ClusterIP to use external IPs.
Off by default as a security posture (CVE-2020-8554 mitigation) —
enable only if a workload genuinely needs external IPs on Services.

### spec.totalEgressBandwidthTier

`string`

Total egress bandwidth tier for node network performance. TIER_1
unlocks up to 100 Gbps on supported machine families (N2/N2D/C2/C3...);
requires gVNIC on the node pools that use it.

- rule: total_egress_bandwidth_tier must be empty, TIER_UNSPECIFIED, or TIER_1

### spec.disableL4LbFirewallReconciliation

`bool`

Stops GKE from reconciling the VPC firewall rules it creates for L4
load balancers — for environments where firewall rules are governed
externally and GKE's reconciliation fights the desired state.

### spec.privateCluster

`GcpGkeClusterPrivateCluster`

Private-cluster topology: private nodes (no public node IPs) and/or a
private-only control-plane endpoint. If omitted, nodes get public IPs
and the control plane has a public endpoint — fine for sandboxes, not
for production.

- rule: enable_private_endpoint requires enable_private_nodes — a public-node cluster cannot have a private-only control plane
- rule: set either master_ipv4_cidr_block (peering-based control-plane range) or private_endpoint_subnetwork (PSC-based endpoint placement) — not both

### spec.privateCluster.enablePrivateNodes

`bool`

Nodes get only internal IPs. Outbound internet (image pulls from
non-Google registries, external APIs) then requires Cloud NAT on the
network — compose a GcpRouterNat.

### spec.privateCluster.enablePrivateEndpoint

`bool`

Removes the control plane's PUBLIC endpoint entirely: kubectl works
only from inside the VPC (or via the DNS endpoint / authorized
networks' private enforcement). The strictest posture; requires
enable_private_nodes.

### spec.privateCluster.masterIpv4CidrBlock

`string`

RFC1918 /28 block for the control plane on peering-based private
clusters, e.g. "172.16.0.16/28". Must not overlap any range in the VPC.
Immutable. Newer PSC-based clusters don't need it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d+\\.\\d+\\.\\d+\\.\\d+/28$"}}

### spec.privateCluster.privateEndpointSubnetwork

`string | valueFrom`

Subnetwork in which the control plane's private endpoint is placed
(PSC-based private clusters). Accepts a self link or a GcpSubnetwork
reference. Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.privateCluster.enableMasterGlobalAccess

`bool`

Makes the private control-plane endpoint reachable from other GCP
regions and on-prem over interconnect/VPN (not just the cluster's
region).

### spec.masterAuthorizedNetworks

`GcpGkeClusterMasterAuthorizedNetworks`

CIDR allowlist for control-plane API access. Without it, the public
endpoint (when enabled) accepts connections from any IP that presents
valid credentials.

### spec.masterAuthorizedNetworks.cidrBlocks

`[]GcpGkeClusterMasterAuthorizedNetworkCidr`

CIDR ranges allowed to reach the control-plane endpoint(s). An empty
list with the block present means "no external networks authorized".

### spec.masterAuthorizedNetworks.cidrBlocks[].cidrBlock

`string` · required

The CIDR range, e.g. "203.0.113.0/24".

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$"}}

### spec.masterAuthorizedNetworks.cidrBlocks[].displayName

`string`

Display label for this entry in the console.

### spec.masterAuthorizedNetworks.gcpPublicCidrsAccessEnabled

`bool` · optional (explicit presence)

Whether Google Cloud public IPs (e.g. Cloud Shell, Cloud Build default
pools) may reach the control plane.

### spec.masterAuthorizedNetworks.privateEndpointEnforcementEnabled

`bool` · optional (explicit presence)

Enforces the allowlist on the PRIVATE endpoint too (by default only the
public endpoint is filtered).

### spec.controlPlaneEndpoints

`GcpGkeClusterControlPlaneEndpoints`

Control-plane endpoint surface: the DNS endpoint (a Google-managed
*.gke.goog name reachable without VPC peering — the modern access path)
and the IP endpoints toggle.

### spec.controlPlaneEndpoints.dnsEndpointAllowExternalTraffic

`bool`

Allows the Google-managed DNS endpoint (*.gke.goog) to accept traffic
from outside the cluster's VPC — IAM-authenticated access without
peering, bastions, or public IPs. The modern alternative to juggling
authorized networks.

### spec.controlPlaneEndpoints.ipEndpointsEnabled

`bool` · optional (explicit presence)

Whether IP-based endpoints (the classic public/private IPs) are served
at all. Set false for DNS-endpoint-only clusters — the strongest
endpoint posture.

- default: `true`

### spec.controlPlaneEndpoints.enableK8sTokensViaDns

`bool` · optional (explicit presence)

Serve Kubernetes ServiceAccount TOKENS via the DNS endpoint —
workloads outside the VPC can authenticate without IP connectivity to
the control plane.

### spec.controlPlaneEndpoints.enableK8sCertsViaDns

`bool` · optional (explicit presence)

Serve Kubernetes client CERTIFICATES via the DNS endpoint.

### spec.releaseChannel

`enum` · optional (explicit presence)

Kubernetes release channel for automatic upgrades. REGULAR (default)
balances freshness and stability; RAPID gets new minors first; STABLE
lags for maximum soak; EXTENDED keeps a minor supported longer for slow
movers; NONE opts out of channel-based auto-upgrade (you own version
management via min_master_version).

- default: `REGULAR`

Allowed values (use exactly as shown):

- `gke_release_channel_unspecified`
- `RAPID`
- `REGULAR`
- `STABLE`
- `NONE`
- `EXTENDED`

### spec.minMasterVersion

`string`

Minimum Kubernetes version for the control plane, e.g. "1.31" or
"1.31.4-gke.1256000". GKE may run a newer patch. Use with release
channel NONE for pinned-version clusters; on a channel, prefer letting
the channel drive versions.

### spec.maintenancePolicy

`GcpGkeClusterMaintenancePolicy`

Maintenance windows and exclusions controlling WHEN GKE may perform
automatic control-plane and node maintenance.

- rule: set exactly one of daily_window, recurring_window, or recurring_time_window

### spec.maintenancePolicy.dailyWindow

`GcpGkeClusterDailyMaintenanceWindow`

Same 4-hour window every day, starting at this UTC time.

### spec.maintenancePolicy.dailyWindow.startTime

`string` · required

Start of the daily 4-hour window, "HH:MM" (UTC), e.g. "03:00".

- rule: {"required":true,"string":{"pattern":"^([0-1][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.maintenancePolicy.recurringWindow

`GcpGkeClusterRecurringMaintenanceWindow`

RRULE-based recurring window anchored on an absolute first
occurrence (start_time/end_time as RFC3339 timestamps) — e.g.
weekends only. Finer control than the daily window.

### spec.maintenancePolicy.recurringWindow.startTime

`string` · required

Window start, RFC3339, e.g. "2026-01-01T02:00:00Z".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.recurringWindow.endTime

`string` · required

Window end (defines the window LENGTH; recurrence drives repetition),
RFC3339.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.recurringWindow.recurrence

`string` · required

RFC5545 RRULE, e.g. "FREQ=WEEKLY;BYDAY=SA,SU" for weekends.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.recurringTimeWindow

`GcpGkeClusterRecurringTimeMaintenanceWindow`

RRULE-based recurring window expressed as a time of day plus a
duration, with an optional date the recurrence may first start. The
same recurrence power as recurring_window without committing to one
absolute first timestamp — pick this when the policy is authored as
"every Saturday at 02:00 for 6 hours, starting next quarter".

### spec.maintenancePolicy.recurringTimeWindow.windowStartTime

`GcpGkeClusterTimeOfDay` · required

Time of day (UTC) each window instance begins.

- rule: {"required":true}

### spec.maintenancePolicy.recurringTimeWindow.windowStartTime.hours

`int32`

Hour of the day, 0-23.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.maintenancePolicy.recurringTimeWindow.windowStartTime.minutes

`int32`

Minute of the hour, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.maintenancePolicy.recurringTimeWindow.windowStartTime.seconds

`int32`

Second of the minute, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.maintenancePolicy.recurringTimeWindow.windowDuration

`string` · required

Length of each window instance as a duration string with a unit
suffix, e.g. "4h", "6h30m", "21600s". Must be positive.

- rule: {"required":true,"string":{"minLen":"1","pattern":"^([0-9]+(\\.[0-9]+)?(ns|us|ms|s|m|h))+$"}}

### spec.maintenancePolicy.recurringTimeWindow.recurrence

`string` · required

RFC5545 RRULE, e.g. "FREQ=WEEKLY;BYDAY=SA,SU" for weekends.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.recurringTimeWindow.delayUntil

`GcpGkeClusterCalendarDate`

Earliest calendar date the recurrence may start; window instances
before it are skipped. Leave unset to start immediately.

### spec.maintenancePolicy.recurringTimeWindow.delayUntil.year

`int32`

Four-digit year, e.g. 2027.

- rule: {"int32":{"lte":9999,"gte":1}}

### spec.maintenancePolicy.recurringTimeWindow.delayUntil.month

`int32`

Month of the year, 1-12.

- rule: {"int32":{"lte":12,"gte":1}}

### spec.maintenancePolicy.recurringTimeWindow.delayUntil.day

`int32`

Day of the month, 1-31.

- rule: {"int32":{"lte":31,"gte":1}}

### spec.maintenancePolicy.exclusions

`[]GcpGkeClusterMaintenanceExclusion`

Date ranges during which non-emergency maintenance is blocked (change
freezes). Maximum 20; the allowed scope/duration depends on the release
channel.

- rule: {"repeated":{"maxItems":"20"}}
- rule: end_time_behavior refines the exclusion scope — set scope alongside it

### spec.maintenancePolicy.exclusions[].exclusionName

`string` · required

Name of the exclusion, e.g. "year-end-freeze".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.exclusions[].startTime

`string` · required

Exclusion start, RFC3339.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.exclusions[].endTime

`string` · required

Exclusion end, RFC3339.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.maintenancePolicy.exclusions[].scope

`string`

What the exclusion blocks: NO_UPGRADES (everything),
NO_MINOR_UPGRADES, or NO_MINOR_OR_NODE_UPGRADES. Empty means
NO_UPGRADES.

- rule: scope must be empty, NO_UPGRADES, NO_MINOR_UPGRADES, or NO_MINOR_OR_NODE_UPGRADES

### spec.maintenancePolicy.exclusions[].endTimeBehavior

`string`

UNTIL_END_OF_SUPPORT stretches the exclusion until the running minor
version's end of support, instead of stopping at end_time — the
"never force this cluster off its minor" stance. Requires scope.

- rule: end_time_behavior must be empty or UNTIL_END_OF_SUPPORT

### spec.maintenancePolicy.disruptionBudget

`GcpGkeClusterDisruptionBudget`

Minimum spacing between consecutive disruptive maintenance events —
a floor on how often GKE may disrupt the cluster with version
changes, independent of windows.

### spec.maintenancePolicy.disruptionBudget.minorVersionDisruptionInterval

`string`

Minimum interval between MINOR version disruptions, seconds format,
e.g. "2419200s" (28 days).

- rule: minor_version_disruption_interval must be a seconds-format duration like "2419200s"

### spec.maintenancePolicy.disruptionBudget.patchVersionDisruptionInterval

`string`

Minimum interval between PATCH version disruptions, seconds format.

- rule: patch_version_disruption_interval must be a seconds-format duration like "604800s"

### spec.clusterAutoscaling

`GcpGkeClusterAutoscaling`

Node auto-provisioning (NAP): GKE creates and deletes node pools
automatically within the resource limits you set — the cluster-level
autoscaler above individual node-pool autoscaling.

- rule: node auto-provisioning requires resource_limits — set at least cpu and memory maximums so provisioning is bounded
- rule: resource_limits apply only when enabled is true

### spec.clusterAutoscaling.enabled

`bool`

Whether node auto-provisioning is on.

### spec.clusterAutoscaling.resourceLimits

`[]GcpGkeClusterAutoscalingResourceLimit`

Cluster-wide bounds per resource type ("cpu", "memory", or an
accelerator like "nvidia-tesla-t4"). NAP never provisions beyond the
maximums — the cost brake.

### spec.clusterAutoscaling.resourceLimits[].resourceType

`string` · required

"cpu", "memory", or an accelerator type such as "nvidia-tesla-t4".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clusterAutoscaling.resourceLimits[].minimum

`int64`

Minimum amount kept provisioned (0 = none).

- rule: {"int64":{"gte":"0"}}

### spec.clusterAutoscaling.resourceLimits[].maximum

`int64`

Maximum amount NAP may provision. Required — an unbounded NAP is an
unbounded bill.

- rule: {"int64":{"gte":"1"}}

### spec.clusterAutoscaling.autoscalingProfile

`string` · optional (explicit presence)

BALANCED (default) or OPTIMIZE_UTILIZATION (scales down harder and
faster — denser packing, more pod churn).

- default: `BALANCED`
- rule: autoscaling_profile must be BALANCED or OPTIMIZE_UTILIZATION

### spec.clusterAutoscaling.autoProvisioningLocations

`[]string`

Zones in which NAP may create node pools (defaults to the cluster's
node zones).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}}}

### spec.clusterAutoscaling.autoProvisioningDefaults

`GcpGkeClusterAutoProvisioningDefaults`

Defaults applied to every node pool NAP creates (identity, disks,
image, shielding, auto-repair/upgrade).

### spec.clusterAutoscaling.autoProvisioningDefaults.serviceAccount

`string | valueFrom`

IAM service account NAP-created nodes run as. Accepts an email or a
reference to a GcpServiceAccount resource. Defaults to the Compute
Engine default SA — create a minimal dedicated SA for production.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.clusterAutoscaling.autoProvisioningDefaults.oauthScopes

`[]string`

OAuth scopes on NAP-created nodes. With Workload Identity, the
cloud-platform scope is the norm (IAM governs actual access).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.clusterAutoscaling.autoProvisioningDefaults.diskSizeGb

`int32` · optional (explicit presence)

Boot disk size in GB for NAP-created nodes (default 100).

- rule: {"int32":{"lte":65536,"gte":10}}

### spec.clusterAutoscaling.autoProvisioningDefaults.diskType

`string`

Boot disk type: pd-standard (default), pd-balanced, pd-ssd, or
hyperdisk-balanced.

- rule: disk_type must be empty, pd-standard, pd-balanced, pd-ssd, or hyperdisk-balanced

### spec.clusterAutoscaling.autoProvisioningDefaults.imageType

`string`

Node image, e.g. "COS_CONTAINERD" (default) or "UBUNTU_CONTAINERD"
(legacy "COS"/"UBUNTU" accepted for old clusters).

- rule: image_type must be empty, COS_CONTAINERD, COS, UBUNTU_CONTAINERD, or UBUNTU

### spec.clusterAutoscaling.autoProvisioningDefaults.minCpuPlatform

`string`

Minimum CPU platform, e.g. "Intel Ice Lake".

### spec.clusterAutoscaling.autoProvisioningDefaults.bootDiskKmsKey

`string | valueFrom`

Customer-managed key encrypting NAP-created node boot disks. Accepts a
full crypto key path or a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.clusterAutoscaling.autoProvisioningDefaults.enableSecureBoot

`bool`

Shielded-VM secure boot on NAP-created nodes (GCP default false).

### spec.clusterAutoscaling.autoProvisioningDefaults.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Shielded-VM integrity monitoring on NAP-created nodes (GCP default
true).

- default: `true`

### spec.clusterAutoscaling.autoProvisioningDefaults.autoUpgrade

`bool` · optional (explicit presence)

Automatic node upgrades on NAP-created pools (GCP default true;
required on a release channel).

- default: `true`

### spec.clusterAutoscaling.autoProvisioningDefaults.autoRepair

`bool` · optional (explicit presence)

Automatic node repair on NAP-created pools (GCP default true).

- default: `true`

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings

`GcpGkeClusterNapUpgradeSettings`

How upgrades roll through NAP-created pools: surge settings or
blue-green, in the same shape as a GcpGkeNodePool's upgrade_settings.

- rule: blue_green_settings apply only when strategy is BLUE_GREEN
- rule: max_surge/max_unavailable apply to the SURGE strategy — remove them when strategy is BLUE_GREEN

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.maxSurge

`uint32` · optional (explicit presence)

Additional nodes added during a surge upgrade.

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.maxUnavailable

`uint32` · optional (explicit presence)

Nodes that may be simultaneously unavailable during a surge upgrade.

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.strategy

`string`

SURGE (rolling replacement) or BLUE_GREEN (full new node set,
migrate, soak, delete).

- rule: strategy must be empty, SURGE, or BLUE_GREEN

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings

`GcpGkeClusterNapBlueGreenSettings`

Blue-green rollout pacing. Only with strategy BLUE_GREEN.

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy

`GcpGkeClusterNapStandardRolloutPolicy`

How the old (blue) node set drains, batch by batch.

- rule: set batch_percentage or batch_node_count, not both

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchPercentage

`float` · optional (explicit presence)

Fraction of blue nodes drained per batch, 0.0-1.0.

- rule: {"float":{"lte":1,"gte":0}}

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchNodeCount

`uint32` · optional (explicit presence)

Number of blue nodes drained per batch.

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.standardRolloutPolicy.batchSoakDuration

`string`

Soak time after each batch, seconds format, e.g. "600s".

- rule: batch_soak_duration must be a seconds-format duration like "600s"

### spec.clusterAutoscaling.autoProvisioningDefaults.upgradeSettings.blueGreenSettings.nodePoolSoakDuration

`string`

Soak time after the blue set is fully drained before deletion,
seconds format, e.g. "3600s".

- rule: node_pool_soak_duration must be a seconds-format duration like "3600s"

### spec.clusterAutoscaling.defaultComputeClassEnabled

`bool` · optional (explicit presence)

Default compute classes: NAP provisions through the cluster's default
ComputeClass definitions instead of the legacy per-resource limits
path — the newer Autopilot-style provisioning model on Standard
clusters.

### spec.enableVerticalPodAutoscaling

`bool`

Enables Vertical Pod Autoscaling: recommends (and can apply) per-pod
CPU/memory requests based on observed usage. Mutable.

### spec.hpaProfile

`string`

Horizontal Pod Autoscaler profile. PERFORMANCE makes HPA react faster
(higher-resolution metrics pipeline); NONE is the classic behavior.

- rule: hpa_profile must be empty, NONE, or PERFORMANCE

### spec.workloadIdentityEnabled

`bool` · optional (explicit presence)

Workload Identity Federation for GKE: pods authenticate to GCP APIs as
IAM service accounts via the cluster's workload pool
(PROJECT_ID.svc.id.goog) — no exported service-account keys. Enabled by
default; disabling it forces workloads back to node service-account
scopes or key files, which is almost never the right call.

- default: `true`

### spec.enableShieldedNodes

`bool` · optional (explicit presence)

Shielded GKE nodes: secure boot + integrity monitoring on node VMs,
verifying node provenance cryptographically. GCP's default (true);
leave unset on Autopilot clusters (always shielded).

### spec.databaseEncryption

`GcpGkeClusterDatabaseEncryption`

Envelope encryption of Kubernetes secrets at the application layer with
a Cloud KMS key (CMEK for etcd secrets).

- rule: state ENCRYPTED requires key_name — the Cloud KMS key that wraps Kubernetes secrets

### spec.databaseEncryption.state

`string` · required

ENCRYPTED (Kubernetes secrets wrapped with key_name),
ALL_OBJECTS_ENCRYPTION_ENABLED (every etcd object wrapped, not just
secrets), or DECRYPTED.

- rule: state must be ENCRYPTED, ALL_OBJECTS_ENCRYPTION_ENABLED, or DECRYPTED
- rule: {"required":true}

### spec.databaseEncryption.keyName

`string | valueFrom`

The Cloud KMS crypto key (full path) used for envelope encryption.
Accepts a literal or a reference to a GcpKmsKey resource. The GKE
service agent needs Encrypter/Decrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.binaryAuthorizationEvaluationMode

`string`

Binary Authorization: PROJECT_SINGLETON_POLICY_ENFORCE admits only
container images that satisfy the project's Binary Authorization policy
(signature/attestation-based supply-chain control); DISABLED turns
enforcement off.

- rule: binary_authorization_evaluation_mode must be empty, DISABLED, or PROJECT_SINGLETON_POLICY_ENFORCE

### spec.securityPosture

`GcpGkeClusterSecurityPosture`

GKE Security Posture dashboard: workload configuration auditing and
vulnerability scanning surfaced in the console.

### spec.securityPosture.mode

`string`

Workload configuration auditing: DISABLED, BASIC, or ENTERPRISE.

- rule: mode must be empty, DISABLED, BASIC, or ENTERPRISE

### spec.securityPosture.vulnerabilityMode

`string`

Workload vulnerability scanning: VULNERABILITY_DISABLED,
VULNERABILITY_BASIC, or VULNERABILITY_ENTERPRISE.

- rule: vulnerability_mode must be empty, VULNERABILITY_DISABLED, VULNERABILITY_BASIC, or VULNERABILITY_ENTERPRISE

### spec.authenticatorSecurityGroup

`string`

Google Group for RBAC: members of this group (and its subgroups) can be
referenced in RBAC bindings. Must be named gke-security-groups@YOURDOMAIN.

### spec.enableLegacyAbac

`bool`

Legacy Attribute-Based Access Control. Leave off: ABAC predates RBAC
and grants coarse permissions that defeat modern authorization. Exists
only for very old workloads being migrated.

### spec.enableMeshCertificates

`bool`

Issues workload mTLS certificates via the mesh CA (used by Anthos
Service Mesh / managed Istio). Requires Workload Identity.

### spec.enableSecretManagerCsi

`bool`

Built-in Secret Manager add-on: mounts Secret Manager secrets into pods
via the CSI driver without third-party operators.

### spec.confidentialNodes

`GcpGkeClusterConfidentialNodes`

Confidential GKE nodes: node VMs run with hardware memory encryption
(AMD SEV / SEV-SNP or Intel TDX). Immutable; restricts machine families.

### spec.confidentialNodes.enabled

`bool`

Whether Confidential GKE Nodes are enabled cluster-wide. Immutable.

### spec.confidentialNodes.confidentialInstanceType

`string`

Confidential computing technology: SEV (default), SEV_SNP, or TDX.
Machine-family support varies by choice.

- rule: confidential_instance_type must be empty, SEV, SEV_SNP, or TDX

### spec.anonymousAuthenticationMode

`string`

Anonymous Kubernetes API authentication posture. LIMITED restricts the
anonymous user to the health-check endpoints only (the hardened
default on new versions); ENABLED preserves full legacy anonymous
access subject to RBAC.

- rule: anonymous_authentication_mode must be empty, ENABLED, or LIMITED

### spec.enableIdentityService

`bool`

GKE Identity Service: authenticate to the Kubernetes API with external
OIDC identity providers (beyond Google accounts).

### spec.logging

`GcpGkeClusterLogging`

Which cluster components ship logs to Cloud Logging. If omitted, GKE's
default (system components + workloads) applies; declare it with
`enabled: false` to turn the integration off.

- rule: logging is on but names no components - list the components that ship logs, or set enabled: false to turn Cloud Logging integration off
- rule: logging is off but still names components - drop components, or set enabled: true to ship their logs

### spec.logging.components

`[]string`

Components exposing logs: SYSTEM_COMPONENTS, WORKLOADS, APISERVER,
CONTROLLER_MANAGER, SCHEDULER, KCP_CONNECTION, KCP_SSHD, KCP_HPA,
KCP_VPA.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"in":["SYSTEM_COMPONENTS","WORKLOADS","APISERVER","CONTROLLER_MANAGER","SCHEDULER","KCP_CONNECTION","KCP_SSHD","KCP_HPA","KCP_VPA"]}}}}

### spec.logging.enabled

`bool` · optional (explicit presence)

Whether Cloud Logging integration is on. Unset means on: declaring the
block with components has always meant shipping their logs, and
`enabled: false` is the explicit OFF (GKE then receives an empty
component list).

- default: `true`

### spec.monitoring

`GcpGkeClusterMonitoring`

Which cluster components ship metrics to Cloud Monitoring, plus managed
Prometheus. If omitted, GKE's defaults apply (system metrics + managed
Prometheus on current versions).

### spec.monitoring.components

`[]string`

Components exposing metrics: SYSTEM_COMPONENTS, APISERVER, SCHEDULER,
CONTROLLER_MANAGER, STORAGE, HPA, POD, DAEMONSET, DEPLOYMENT,
STATEFULSET, KUBELET, CADVISOR, DCGM, JOBSET. An empty list disables
Cloud Monitoring integration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"in":["SYSTEM_COMPONENTS","APISERVER","SCHEDULER","CONTROLLER_MANAGER","STORAGE","HPA","POD","DAEMONSET","DEPLOYMENT","STATEFULSET","KUBELET","CADVISOR","DCGM","JOBSET"]}}}}

### spec.monitoring.managedPrometheusEnabled

`bool` · optional (explicit presence)

Google Cloud Managed Service for Prometheus: managed collection of
Prometheus metrics (GKE's default on current versions). Disabling it
means running your own Prometheus stack.

- default: `true`

### spec.monitoring.autoMonitoringScope

`string`

Managed Prometheus auto-monitoring scope: ALL deploys packaged
PodMonitorings for supported workloads automatically; NONE leaves
monitoring configuration entirely to you.

- rule: auto_monitoring_scope must be empty, ALL, or NONE

### spec.monitoring.advancedDatapathMetricsEnabled

`bool`

Dataplane V2 observability metrics (per-flow network telemetry).

### spec.monitoring.advancedDatapathRelayEnabled

`bool`

Dataplane V2 observability relay (Hubble-compatible flow export for
in-cluster observability tooling).

### spec.notificationPubsub

`GcpGkeClusterNotificationPubSub`

Cluster lifecycle notifications (upgrades, security bulletins)
published to a Pub/Sub topic — the hook for upgrade automation and
fleet dashboards.

- rule: notification_pubsub.enabled requires topic — the Pub/Sub topic GKE publishes to

### spec.notificationPubsub.enabled

`bool`

Whether notifications are published.

### spec.notificationPubsub.topic

`string | valueFrom`

The Pub/Sub topic (projects/{project}/topics/{name}). Accepts a literal
or a reference to a GcpPubSubTopic resource.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.notificationPubsub.eventTypes

`[]string`

Restrict to specific event types; empty publishes all. Values:
UPGRADE_EVENT, UPGRADE_AVAILABLE_EVENT, SECURITY_BULLETIN_EVENT,
UPGRADE_INFO_EVENT.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"in":["UPGRADE_EVENT","UPGRADE_AVAILABLE_EVENT","SECURITY_BULLETIN_EVENT","UPGRADE_INFO_EVENT"]}}}}

### spec.enableCostManagement

`bool`

Cost allocation: exports per-namespace/per-label resource usage into
the billing export, so cluster cost can be attributed to teams.

### spec.resourceUsageExport

`GcpGkeClusterResourceUsageExport`

Continuous resource-usage metering into a BigQuery dataset (network
egress and resource consumption records).

### spec.resourceUsageExport.bigqueryDatasetId

`string | valueFrom` · required

The BigQuery dataset receiving usage records. Accepts a dataset ID or a
reference to a GcpBigQueryDataset resource. The dataset must be in the
cluster's project.

- references: GcpBigQueryDataset (`status.outputs.dataset_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryDataset, name: <that resource's name>, fieldPath: status.outputs.dataset_id}} -- a bare string does not parse

### spec.resourceUsageExport.enableNetworkEgressMetering

`bool`

Also meter pod network egress (deploys a metering agent per node).

### spec.resourceUsageExport.enableResourceConsumptionMetering

`bool` · optional (explicit presence)

Meter resource requests/consumption (GCP default true).

- default: `true`

### spec.addons

`GcpGkeClusterAddons`

Cluster addons. If omitted, GKE defaults apply: HTTP load balancing,
horizontal pod autoscaling, and the PD CSI driver enabled; everything
else disabled.

### spec.addons.httpLoadBalancingEnabled

`bool` · optional (explicit presence)

The GCE ingress controller backing Kubernetes Ingress/Gateway with
Google Cloud Load Balancers. Disable only when bringing your own
ingress stack end to end.

- default: `true`

### spec.addons.horizontalPodAutoscalingEnabled

`bool` · optional (explicit presence)

The Horizontal Pod Autoscaler controller.

- default: `true`

### spec.addons.gcePersistentDiskCsiDriverEnabled

`bool` · optional (explicit presence)

The Compute Engine Persistent Disk CSI driver (dynamic PD volumes).

- default: `true`

### spec.addons.gcpFilestoreCsiDriverEnabled

`bool`

The Filestore CSI driver (managed NFS volumes).

### spec.addons.gcsFuseCsiDriverEnabled

`bool`

The Cloud Storage FUSE CSI driver (mount GCS buckets as volumes).

### spec.addons.gkeBackupAgentEnabled

`bool`

The Backup for GKE agent (workload + volume backup/restore).

### spec.addons.dnsCacheEnabled

`bool`

NodeLocal DNSCache (per-node DNS cache; reduces DNS latency and
kube-dns/Cloud DNS load). Enabling recreates node pools on existing
clusters.

### spec.addons.configConnectorEnabled

`bool`

Config Connector (manage GCP resources through Kubernetes CRDs).

### spec.addons.statefulHaEnabled

`bool`

Stateful HA operator (faster failover for stateful workloads).

### spec.addons.rayOperatorEnabled

`bool`

The Ray operator (KubeRay) for distributed Python/AI workloads.

### spec.addons.rayClusterLoggingEnabled

`bool`

Ray cluster logging integration (with ray_operator_enabled).

### spec.addons.rayClusterMonitoringEnabled

`bool`

Ray cluster monitoring integration (with ray_operator_enabled).

### spec.addons.cloudrunEnabled

`bool`

Cloud Run for Anthos / Knative serving addon.

### spec.addons.cloudrunLoadBalancerType

`string`

Load balancer type for the Cloud Run addon:
LOAD_BALANCER_TYPE_INTERNAL serves it on an internal LB instead of
the default external one.

- rule: cloudrun_load_balancer_type must be empty or LOAD_BALANCER_TYPE_INTERNAL

### spec.addons.parallelstoreCsiDriverEnabled

`bool`

The Parallelstore CSI driver (managed parallel filesystem for
AI/HPC).

### spec.addons.lustreCsiDriverEnabled

`bool`

The Lustre CSI driver (Managed Lustre high-performance filesystem).

### spec.addons.lustreCsiLegacyPortEnabled

`bool`

Serve the Lustre protocol on the legacy port (compatibility with
pre-GA Lustre deployments; with lustre_csi_driver_enabled).

### spec.addons.lustreCsiDisableMultiNic

`bool`

Disable multi-NIC transport for the Lustre CSI driver (with
lustre_csi_driver_enabled).

### spec.addons.podSnapshotEnabled

`bool`

Pod snapshot support (checkpoint/restore of running pods).

### spec.addons.agentSandboxEnabled

`bool`

GKE Sandbox (gVisor) agent addon for Autopilot sandbox pods.

### spec.addons.sliceControllerEnabled

`bool`

The slice controller addon (TPU slice management).

### spec.addons.slurmOperatorEnabled

`bool`

The Slurm operator addon (Slurm-on-GKE for HPC scheduling).

### spec.addons.highScaleCheckpointingEnabled

`bool`

High Scale Checkpointing: the addon that lets large AI/ML training
jobs checkpoint and restore state at scale (multi-tier checkpointing
onto node-local and Cloud Storage tiers) so a job resumes from its
last checkpoint after a preemption or failure instead of restarting.

### spec.addons.nodeReadinessControllerEnabled

`bool`

Node Readiness Controller: the addon that holds a node out of
scheduling until its readiness rules (for example, required daemon
pods or device drivers) pass, so workloads never land on a node
whose accelerators or networking are not yet usable.

### spec.enableAutopilot

`bool`

Creates an Autopilot cluster: GKE provisions and manages nodes, bills
per pod, and enforces a hardened posture. Immutable. Autopilot clusters
take no GcpGkeNodePool resources and reject the node-management fields
guarded by validation rules above.

### spec.allowNetAdmin

`bool`

Autopilot only: permit workloads with NET_ADMIN capability (needed by
some networking agents/service meshes on Autopilot).

### spec.fleetProject

`string`

Registers the cluster with a fleet in the given project (the hub for
multi-cluster features: multi-cluster ingress/services, config
management, team scopes).

### spec.fleetMembershipType

`string`

Fleet membership type. LIGHTWEIGHT registers a lightweight membership
(reduced fleet feature surface, no Connect agent). Empty uses the
fleet default (full membership).

- rule: fleet_membership_type must be empty or LIGHTWEIGHT

### spec.deletionPolicy

`string`

Destroy-time stance of the IaC engines toward the cluster itself:
DELETE (default) destroys it, PREVENT fails any plan that would
destroy it, ABANDON removes it from state and leaves the cluster
running in GCP. This is an engine-side control layered UNDER
deletion_protection (the GKE-native guard): deletion_protection
blocks the API call itself; deletion_policy governs what the engines
even attempt.

- rule: deletion_policy must be empty, DELETE, PREVENT, or ABANDON

### spec.ignoreNodeCountChanges

`bool`

Skips the per-pool Instance Group Manager queries during cluster
reads — a quota/performance optimization for clusters with many
pools. While true, node-count drift is invisible to plans and
managed-instance-group outputs go stale on every pool.

### spec.skipNodePoolRefresh

`bool`

Skips refreshing inline node-pool state from the API during cluster
reads — a substantial plan/apply speedup on clusters with many pools.
Safe in this composition because node pools are always separate
GcpGkeNodePool resources, never inline blocks on the cluster.

### spec.enableKubernetesAlpha

`bool`

Creates an ALPHA cluster: all Kubernetes alpha feature gates enabled,
no SLA, cannot be upgraded, and GKE DELETES the cluster after 30
days. Strictly for short-lived feature evaluation. Immutable.

### spec.k8sBetaApis

`[]string`

Specific Kubernetes BETA API groups enabled on the cluster, e.g.
["resource.k8s.io/v1beta1/deviceclasses"]. Beta APIs are off by
default on new clusters; list exactly the groups a workload needs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.dataplaneOptimizationMode

`string`

Dataplane optimization mode (pass-through to the GKE API; GKE
validates accepted modes per version). Immutable.

### spec.issueClientCertificate

`bool` · optional (explicit presence)

Issues a legacy client certificate for control-plane authentication.
Off on modern clusters — certificate auth bypasses IAM and cannot be
revoked short of rotating the cluster CA; leave unset unless a legacy
client genuinely requires it.

### spec.nodeCreationMode

`string`

How nodes register themselves: VIA_KUBELET (nodes self-register, the
classic path) or VIA_CONTROL_PLANE (the control plane creates node
objects — hardens against node impersonation). Immutable.

- rule: node_creation_mode must be empty, VIA_KUBELET, or VIA_CONTROL_PLANE

### spec.gkeAutoUpgradePatchMode

`string`

Auto-upgrade patch cadence: ACCELERATED upgrades to the latest patch
available in the cluster's minor and channel as soon as it ships
(instead of the channel's default rollout pacing).

- rule: gke_auto_upgrade_patch_mode must be empty or ACCELERATED

### spec.rbacBindingConfig

`GcpGkeClusterRbacBindingConfig`

Locks down the legacy RBAC bindings to system:authenticated /
system:unauthenticated — the hardening that prevents accidentally
granting cluster access to every Google account on earth.

### spec.rbacBindingConfig.enableInsecureBindingSystemAuthenticated

`bool` · optional (explicit presence)

Allow ClusterRoleBindings/RoleBindings that grant to
system:authenticated (every Google-authenticated identity). Leave
false — granting to system:authenticated is almost always a mistake.

### spec.rbacBindingConfig.enableInsecureBindingSystemUnauthenticated

`bool` · optional (explicit presence)

Allow bindings that grant to system:unauthenticated /
system:anonymous. Leave false.

### spec.autopilotPolicy

`GcpGkeClusterAutopilotPolicy`

Autopilot-only conversion/posture policies (standard-pool
prohibition, system mutation/impersonation guards, webhook safety).

### spec.autopilotPolicy.noStandardNodePools

`bool` · optional (explicit presence)

Prohibit standard node pools — the cluster runs pure Autopilot.

### spec.autopilotPolicy.noSystemImpersonation

`bool` · optional (explicit presence)

Disallow impersonating system identities.

### spec.autopilotPolicy.noSystemMutation

`bool` · optional (explicit presence)

Disallow mutating system-managed objects.

### spec.autopilotPolicy.noUnsafeWebhooks

`bool` · optional (explicit presence)

Disallow admission webhooks that intercept system-critical requests.

### spec.autopilotPrivilegedAdmissionPaths

`[]string`

Autopilot-only: Cloud Storage / GKE allowlist paths authorizing
PRIVILEGED workloads on Autopilot (partner agents etc.). Entries must
start with gke:// or gs://; [] allows the default partner allowlists.

- rule: {"repeated":{"items":{"string":{"pattern":"^((gke|gs)://.+)?$"}}}}

### spec.nodePoolAutoConfig

`GcpGkeClusterNodePoolAutoConfig`

Node settings GKE applies to the pools IT manages on an Autopilot
cluster (network tags, Resource Manager tags, kubelet/OS hardening) —
the Autopilot counterpart of per-pool node_config.

### spec.nodePoolAutoConfig.networkTags

`[]string`

GCE network tags applied to Autopilot-managed nodes — what VPC
firewall rules match.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.nodePoolAutoConfig.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to Autopilot-managed node VMs, as
{"tagKeys/123": "tagValues/456"} pairs.

### spec.nodePoolAutoConfig.cgroupMode

`string`

Container runtime cgroup mode on Autopilot-managed nodes.

- rule: cgroup_mode must be empty, CGROUP_MODE_UNSPECIFIED, CGROUP_MODE_V1, or CGROUP_MODE_V2

### spec.nodePoolAutoConfig.nodeKernelModuleLoadingPolicy

`string`

Signed-kernel-module enforcement on Autopilot-managed nodes.

- rule: node_kernel_module_loading_policy must be empty, POLICY_UNSPECIFIED, ENFORCE_SIGNED_MODULES, or DO_NOT_ENFORCE_SIGNED_MODULES

### spec.nodePoolAutoConfig.insecureKubeletReadonlyPortEnabled

`string`

The kubelet's insecure read-only port 10255 on Autopilot-managed
nodes: FALSE closes it (hardened), TRUE keeps it open for legacy
agents.

- rule: insecure_kubelet_readonly_port_enabled must be empty, TRUE, or FALSE

### spec.nodePoolDefaults

`GcpGkeClusterNodePoolDefaults`

Defaults inherited by every node pool at CREATION time on a Standard
cluster (image streaming, kubelet read-only port, logging variant,
containerd registry access). A pool's own node_config overrides
these.

### spec.nodePoolDefaults.gcfsEnabled

`bool` · optional (explicit presence)

Image streaming (GCFS) default for new pools: containers start before
the full image is pulled.

### spec.nodePoolDefaults.insecureKubeletReadonlyPortEnabled

`string`

Default posture of the kubelet's insecure read-only port 10255 on new
pools: FALSE closes it (hardened), TRUE keeps it open.

- rule: insecure_kubelet_readonly_port_enabled must be empty, TRUE, or FALSE

### spec.nodePoolDefaults.loggingVariant

`string`

Default node system-log throughput for new pools: DEFAULT (100 KiB/s)
or MAX_THROUGHPUT (10 MiB/s).

- rule: logging_variant must be empty, DEFAULT, or MAX_THROUGHPUT

### spec.nodePoolDefaults.containerdConfig

`GcpGkeClusterContainerdDefaults`

Default containerd registry configuration for new pools: private-CA
registry trust, per-registry host overrides, writable cgroups.

### spec.nodePoolDefaults.containerdConfig.privateRegistryAccess

`GcpGkeClusterPrivateRegistryAccess`

Trust custom certificate authorities for specific registry domains.

### spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.enabled

`bool`

Master toggle for private registry access configuration.

### spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains

`[]GcpGkeClusterRegistryCaDomain`

Per-domain CA trust entries.

### spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].fqdns

`[]string` · required

Registry FQDNs this CA vouches for, e.g. ["registry.internal:5000"].

- rule: {"repeated":{"minItems":"1"}}

### spec.nodePoolDefaults.containerdConfig.privateRegistryAccess.certificateAuthorityDomains[].gcpSecretManagerCertificateUri

`string` · required

Secret Manager secret URI holding the CA certificate, in the form
"projects/{project}/secrets/{secret}/versions/{version}".

- rule: {"required":true}

### spec.nodePoolDefaults.containerdConfig.registryHosts

`[]GcpGkeClusterRegistryHost`

Per-registry host overrides (mirrors, capabilities, auth, headers).

### spec.nodePoolDefaults.containerdConfig.registryHosts[].server

`string` · required

The registry server the overrides apply to, e.g. "docker.io".

- rule: {"required":true}

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts

`[]GcpGkeClusterRegistryHostEndpoint`

Host endpoints serving this registry (mirrors first, in order).

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].host

`string` · required

Endpoint URL, e.g. "https://mirror.internal".

- rule: {"required":true}

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].capabilities

`[]string`

Operations this endpoint can serve — typically "pull" and "resolve"
(containerd hosts.toml capability names).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true}}

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].dialTimeout

`string`

Dial timeout for this endpoint, e.g. "10s".

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].overridePath

`bool` · optional (explicit presence)

Path override on the endpoint host.

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].caSecretUri

`string`

Secret Manager secret URI for the CA certificate to trust for this
endpoint.

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].clientCertSecretUri

`string`

Secret Manager secret URI for the client TLS certificate presented to
this endpoint.

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].clientKeySecretUri

`string`

Secret Manager secret URI for the client TLS key presented to this
endpoint.

### spec.nodePoolDefaults.containerdConfig.registryHosts[].hosts[].headers

`map<string, string>`

Custom HTTP headers sent to this endpoint.

### spec.nodePoolDefaults.containerdConfig.writableCgroupsEnabled

`bool` · optional (explicit presence)

Writable cgroup filesystem inside containers.

### spec.userManagedKeys

`GcpGkeClusterUserManagedKeys`

Bring-your-own control-plane keys: customer-managed CAs for the
cluster/etcd/aggregation trust domains and KMS keys for control-plane
disk encryption and ServiceAccount JWT signing. For regulated
environments that must own the entire trust chain. Immutable.

### spec.userManagedKeys.clusterCa

`string`

CA Service CaPool issuing the cluster CA
("projects/{p}/locations/{l}/caPools/{pool}").

### spec.userManagedKeys.etcdApiCa

`string`

CA Service CaPool for the etcd API CA.

### spec.userManagedKeys.etcdPeerCa

`string`

CA Service CaPool for the etcd peer CA.

### spec.userManagedKeys.aggregationCa

`string`

CA Service CaPool for the aggregation layer CA.

### spec.userManagedKeys.controlPlaneDiskEncryptionKey

`string | valueFrom`

KMS key encrypting the control-plane disks
("projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}"). Accepts a
literal path or a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.userManagedKeys.gkeopsEtcdBackupEncryptionKey

`string | valueFrom`

KMS key encrypting GKE-ops etcd backups. Accepts a literal path or a
reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.userManagedKeys.serviceAccountSigningKeys

`[]string`

KMS cryptoKeyVersions SIGNING ServiceAccount JWTs issued by the
cluster ("projects/.../cryptoKeyVersions/1").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.userManagedKeys.serviceAccountVerificationKeys

`[]string`

KMS cryptoKeyVersions accepted for VERIFYING ServiceAccount JWTs
(include the previous version during signing-key rotation).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.secretManagerRotation

`GcpGkeClusterSecretRotation`

Rotation of Secret Manager secrets mounted through the built-in CSI
add-on (requires enable_secret_manager_csi): re-fetch cadence for
mounted secret values.

### spec.secretManagerRotation.enabled

`bool`

Whether mounted secrets are periodically re-fetched.

### spec.secretManagerRotation.rotationInterval

`string`

Re-fetch cadence, seconds format (e.g. "120s"; default 120s).

- rule: rotation_interval must be a seconds-format duration like "120s"

### spec.secretSync

`GcpGkeClusterSecretSync`

The Secret Manager SYNC add-on: syncs Secret Manager secrets into
Kubernetes Secret objects (as opposed to the CSI add-on's volume
mounts), with its own rotation cadence.

### spec.secretSync.enabled

`bool`

Whether the sync add-on is enabled.

### spec.secretSync.rotationEnabled

`bool`

Whether synced secrets are periodically refreshed.

### spec.secretSync.rotationInterval

`string`

Refresh cadence, seconds format (e.g. "120s").

- rule: rotation_interval must be a seconds-format duration like "120s"

### spec.rollbackSafeUpgrade

`GcpGkeClusterRollbackSafeUpgrade`

Two-step (rollback-safe) control-plane minor upgrades: the control
plane moves to the new version but keeps emulating the old minor for
a soak period during which the upgrade can be rolled back without
data loss. Leave unset for standard one-step upgrades.

### spec.rollbackSafeUpgrade.controlPlaneSoakDuration

`string`

How long the cluster stays in the rollbackable state after the
control plane upgrades, as a seconds-format duration, e.g. "604800s"
(7 days). Minimum 6 hours ("21600s"), maximum 7 days ("604800s").
Leave empty to skip the two-step flow and perform a standard
one-step upgrade.
The bound is expressed as one pattern (21600 <= seconds <= 604800)
so every validation engine, including the Java one, evaluates it
without string slicing.

- rule: control_plane_soak_duration must be a seconds-format duration between "21600s" (6 hours) and "604800s" (7 days)

### spec.desiredEmulatedVersion

`string`

Completes a rollback-safe upgrade declaratively: set to the target
minor version ("major.minor", e.g. "1.33") once the soak period has
proven the new control plane, and GKE stops emulating the old minor.
Removing the field does not trigger completion; only setting it does.
Only meaningful with rollback_safe_upgrade.

- rule: desired_emulated_version must be in major.minor format, e.g. "1.33"

## Validation Rules

- `autopilot_conflicts_cluster_autoscaling`: cluster_autoscaling (node auto-provisioning) cannot be configured on an Autopilot cluster — Autopilot manages node provisioning itself
- `autopilot_conflicts_max_pods`: default_max_pods_per_node cannot be set on an Autopilot cluster — Autopilot manages pod density itself
- `autopilot_conflicts_intranode_visibility`: enable_intranode_visibility cannot be set on an Autopilot cluster — Autopilot controls dataplane telemetry itself
- `autopilot_conflicts_network_policy`: enable_network_policy (Calico) cannot be set on an Autopilot cluster — Autopilot enforces NetworkPolicy through Dataplane V2 natively
- `autopilot_conflicts_shielded_nodes`: enable_shielded_nodes must be left unset on an Autopilot cluster — Autopilot nodes are always shielded
- `autopilot_conflicts_addons`: dns_cache and stateful_ha addons cannot be enabled on an Autopilot cluster — Autopilot manages the addon set itself
- `net_admin_requires_autopilot`: allow_net_admin applies to Autopilot clusters only — Standard clusters always permit NET_ADMIN
- `autopilot_policy_requires_autopilot`: autopilot_policy applies to Autopilot clusters only
- `privileged_admission_requires_autopilot`: autopilot_privileged_admission_paths applies to Autopilot clusters only
- `node_pool_auto_config_requires_autopilot`: node_pool_auto_config configures the nodes GKE manages on an Autopilot cluster — on Standard clusters configure each GcpGkeNodePool instead

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint` | `string` | Kubernetes API server endpoint (IP address). For a cluster with a private-only control plane this is the private endpoint. |
| `status.outputs.cluster_ca_certificate` | `string` | Base64-encoded CA certificate for the cluster's API server. Clients need this to validate the API server's TLS certificate. Public trust material, not a secret. |
| `status.outputs.workload_identity_pool` | `string` | Workload Identity pool used by this cluster (e.g. "PROJECT_ID.svc.id.goog"). Empty when Workload Identity is disabled. |
| `status.outputs.cluster_id` | `string` | Fully qualified GKE cluster resource ID: projects/{project}/locations/{location}/clusters/{name}. Downstream resources (e.g. Dataproc on GKE) reference the cluster by this ID. |
| `status.outputs.name` | `string` | The cluster name as created in GCP — the handle node pools and gcloud commands use. Matches spec.cluster_name when set, otherwise metadata.name. |
| `status.outputs.location` | `string` | The cluster's location (region for regional clusters, zone for zonal), exactly as provided in the spec. |
| `status.outputs.self_link` | `string` | Server-defined URL of the cluster resource. |
| `status.outputs.master_version` | `string` | The Kubernetes version currently running on the control plane. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.ipAllocation.clusterSecondaryRangeName` | GcpSubnetwork | `status.outputs.secondary_ranges.[*].range_name` |
| `spec.ipAllocation.servicesSecondaryRangeName` | GcpSubnetwork | `status.outputs.secondary_ranges.[*].range_name` |
| `spec.ipAllocation.additionalIpRanges[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.privateCluster.privateEndpointSubnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.clusterAutoscaling.autoProvisioningDefaults.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.clusterAutoscaling.autoProvisioningDefaults.bootDiskKmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.databaseEncryption.keyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.notificationPubsub.topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.resourceUsageExport.bigqueryDatasetId` | GcpBigQueryDataset | `status.outputs.dataset_id` |
| `spec.userManagedKeys.controlPlaneDiskEncryptionKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.userManagedKeys.gkeopsEtcdBackupEncryptionKey` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDataprocCluster | `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget` | `status.outputs.cluster_id` |
| GcpDnsZone | `spec.privateVisibilityConfig.gkeClusters[].gkeClusterName` | `status.outputs.cluster_id` |
| GcpEventarcTrigger | `spec.destination.gke.cluster` | `status.outputs.name` |
| GcpGkeNodePool | `spec.clusterName` | `status.outputs.name` |
| GcpGkeNodePool | `spec.location` | `status.outputs.location` |

## See Also

- [Overview](../README.md)

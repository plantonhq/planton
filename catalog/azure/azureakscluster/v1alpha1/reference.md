# AzureAksCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureAksClusterSpec** defines the configuration for creating an Azure
Kubernetes Service (AKS) managed cluster: the control plane, its identity
and access model, its network fabric, the mandatory default node pool, and
the Azure-managed add-ons that turn a bare cluster into a platform.

The cluster deliberately carries exactly ONE node pool -- the default
(system) pool Azure requires at creation. Every additional pool is its own
composable AzureAksNodePool resource referencing this cluster's
`cluster_id` output: pools have independent lifecycles (scale, upgrade,
spot-evict, delete) and coupling them to the cluster would force cluster
updates for pool changes.

Workload identity composition: enabling `oidc_issuer_enabled` (on by
default) publishes the cluster's OIDC issuer URL as the `oidc_issuer_url`
output, which an AzureFederatedIdentityCredential consumes as its `issuer`
-- the keyless (secret-less) path for pods to act as an Azure managed
identity.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksCluster
metadata:
  name: production-aks-cluster
  labels:
    environment: production
    team: platform
spec:
  region: eastus
  resourceGroup:
    value: production-rg
  name: production-aks-cluster

  # Pin for production; unset lets AKS pick the latest recommended GA version.
  kubernetesVersion: "1.35"

  skuTier: STANDARD

  networkProfile:
    networkPlugin: AZURE_CNI
    networkPluginMode: OVERLAY

  apiServerAccessProfile:
    authorizedIpRanges:
      - "203.0.113.0/24"

  oidcIssuerEnabled: true
  workloadIdentityEnabled: true
  azurePolicyEnabled: true

  defaultNodePool:
    name: system
    vmSize: Standard_D4s_v5
    autoScalingEnabled: true
    minCount: 3
    maxCount: 5
    zones:
      - "1"
      - "2"
      - "3"
    onlyCriticalAddonsEnabled: true

  omsAgent:
    logAnalyticsWorkspaceId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-monitoring/providers/Microsoft.OperationalInsights/workspaces/aks-logs
    msiAuthForMonitoringEnabled: true

  keyVaultSecretsProvider:
    secretRotationEnabled: true

  tags:
    cost-center: platform
    owner: platform-team
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.region` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.dnsPrefix` | `string` |  |  |  |
| `spec.dnsPrefixPrivateCluster` | `string` |  |  |  |
| `spec.kubernetesVersion` | `string` |  |  |  |
| `spec.skuTier` | `enum` |  |  |  |
| `spec.supportPlan` | `enum` |  |  |  |
| `spec.defaultNodePool` | `AzureAksClusterDefaultNodePool` | yes |  |  |
| `spec.defaultNodePool.name` | `string` | yes |  |  |
| `spec.defaultNodePool.vmSize` | `string` | yes | `Standard_D4s_v5` |  |
| `spec.defaultNodePool.nodeCount` | `int32` |  |  |  |
| `spec.defaultNodePool.autoScalingEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.minCount` | `int32` |  |  |  |
| `spec.defaultNodePool.maxCount` | `int32` |  |  |  |
| `spec.defaultNodePool.maxPods` | `int32` |  |  |  |
| `spec.defaultNodePool.zones` | `[]string` |  |  |  |
| `spec.defaultNodePool.vnetSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.defaultNodePool.podSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.defaultNodePool.osDiskSizeGb` | `int32` |  |  |  |
| `spec.defaultNodePool.osDiskType` | `enum` |  |  |  |
| `spec.defaultNodePool.kubeletDiskType` | `enum` |  |  |  |
| `spec.defaultNodePool.osSku` | `enum` |  |  |  |
| `spec.defaultNodePool.orchestratorVersion` | `string` |  |  |  |
| `spec.defaultNodePool.nodeLabels` | `map<string, string>` |  |  |  |
| `spec.defaultNodePool.onlyCriticalAddonsEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.fipsEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.hostEncryptionEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.nodePublicIpEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.nodePublicIpPrefixId` | `string \| valueFrom` |  |  | AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`) |
| `spec.defaultNodePool.gpuInstance` | `enum` |  |  |  |
| `spec.defaultNodePool.gpuDriver` | `enum` |  |  |  |
| `spec.defaultNodePool.proximityPlacementGroupId` | `string` |  |  |  |
| `spec.defaultNodePool.hostGroupId` | `string` |  |  |  |
| `spec.defaultNodePool.capacityReservationGroupId` | `string` |  |  |  |
| `spec.defaultNodePool.scaleDownMode` | `enum` |  |  |  |
| `spec.defaultNodePool.snapshotId` | `string` |  |  |  |
| `spec.defaultNodePool.workloadRuntime` | `enum` |  |  |  |
| `spec.defaultNodePool.ultraSsdEnabled` | `bool` |  |  |  |
| `spec.defaultNodePool.temporaryNameForRotation` | `string` |  |  |  |
| `spec.defaultNodePool.kubeletConfig` | `AzureAksClusterKubeletConfig` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.cpuManagerPolicy` | `enum` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.cpuCfsQuotaEnabled` | `bool` |  | `true` |  |
| `spec.defaultNodePool.kubeletConfig.cpuCfsQuotaPeriod` | `string` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.imageGcHighThreshold` | `int32` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.imageGcLowThreshold` | `int32` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.topologyManagerPolicy` | `enum` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.allowedUnsafeSysctls` | `[]string` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.containerLogMaxSizeMb` | `int32` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.containerLogMaxFiles` | `int32` |  |  |  |
| `spec.defaultNodePool.kubeletConfig.podMaxPid` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig` | `AzureAksClusterLinuxOsConfig` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig` | `AzureAksClusterSysctlConfig` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsAioMaxNr` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsFileMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsInotifyMaxUserWatches` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsNrOpen` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.kernelThreadsMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreNetdevMaxBacklog` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreOptmemMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreRmemDefault` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreRmemMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreSomaxconn` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreWmemDefault` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreWmemMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMin` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh1` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh2` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh3` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpFinTimeout` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveIntvl` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveProbes` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveTime` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpMaxSynBacklog` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpMaxTwBuckets` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpTwReuse` | `bool` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackBuckets` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackMax` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmMaxMapCount` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmSwappiness` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmVfsCachePressure` | `int32` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.transparentHugePage` | `enum` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.transparentHugePageDefrag` | `enum` |  |  |  |
| `spec.defaultNodePool.linuxOsConfig.swapFileSizeMb` | `int32` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile` | `AzureAksClusterNodeNetworkProfile` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts` | `[]AzureAksClusterAllowedHostPorts` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].portStart` | `int32` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].portEnd` | `int32` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].protocol` | `enum` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.applicationSecurityGroupIds` | `[]string` |  |  |  |
| `spec.defaultNodePool.nodeNetworkProfile.nodePublicIpTags` | `map<string, string>` |  |  |  |
| `spec.defaultNodePool.upgradeSettings` | `AzureAksClusterDefaultNodePoolUpgradeSettings` |  |  |  |
| `spec.defaultNodePool.upgradeSettings.maxSurge` | `string` | yes |  |  |
| `spec.defaultNodePool.upgradeSettings.drainTimeoutInMinutes` | `int32` |  |  |  |
| `spec.defaultNodePool.upgradeSettings.nodeSoakDurationInMinutes` | `int32` |  |  |  |
| `spec.defaultNodePool.upgradeSettings.undrainableNodeBehavior` | `enum` |  |  |  |
| `spec.defaultNodePool.tags` | `map<string, string>` |  |  |  |
| `spec.identity` | `AzureAksClusterIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.kubeletIdentity` | `AzureAksClusterKubeletIdentity` |  |  |  |
| `spec.kubeletIdentity.clientId` | `string` | yes |  |  |
| `spec.kubeletIdentity.objectId` | `string` | yes |  |  |
| `spec.kubeletIdentity.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.oidcIssuerEnabled` | `bool` |  | `true` |  |
| `spec.workloadIdentityEnabled` | `bool` |  |  |  |
| `spec.privateClusterEnabled` | `bool` |  |  |  |
| `spec.privateDnsZoneId` | `string \| valueFrom` |  |  | AzurePrivateDnsZone (`status.outputs.zone_id`) |
| `spec.privateClusterPublicFqdnEnabled` | `bool` |  |  |  |
| `spec.apiServerAccessProfile` | `AzureAksClusterApiServerAccessProfile` |  |  |  |
| `spec.apiServerAccessProfile.authorizedIpRanges` | `[]string` |  |  |  |
| `spec.apiServerAccessProfile.virtualNetworkIntegrationEnabled` | `bool` |  |  |  |
| `spec.apiServerAccessProfile.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.roleBasedAccessControlEnabled` | `bool` |  | `true` |  |
| `spec.localAccountDisabled` | `bool` |  |  |  |
| `spec.azureActiveDirectoryRoleBasedAccessControl` | `AzureAksClusterAadRbac` |  |  |  |
| `spec.azureActiveDirectoryRoleBasedAccessControl.tenantId` | `string` |  |  |  |
| `spec.azureActiveDirectoryRoleBasedAccessControl.azureRbacEnabled` | `bool` |  |  |  |
| `spec.azureActiveDirectoryRoleBasedAccessControl.adminGroupObjectIds` | `[]string` |  |  |  |
| `spec.networkProfile` | `AzureAksClusterNetworkProfile` |  |  |  |
| `spec.networkProfile.networkPlugin` | `enum` |  |  |  |
| `spec.networkProfile.networkPluginMode` | `enum` |  |  |  |
| `spec.networkProfile.networkPolicy` | `enum` |  |  |  |
| `spec.networkProfile.networkDataPlane` | `enum` |  |  |  |
| `spec.networkProfile.dnsServiceIp` | `string` |  |  |  |
| `spec.networkProfile.serviceCidr` | `string` |  |  |  |
| `spec.networkProfile.serviceCidrs` | `[]string` |  |  |  |
| `spec.networkProfile.podCidr` | `string` |  |  |  |
| `spec.networkProfile.podCidrs` | `[]string` |  |  |  |
| `spec.networkProfile.ipVersions` | `[]enum` |  |  |  |
| `spec.networkProfile.outboundType` | `enum` |  |  |  |
| `spec.networkProfile.loadBalancerProfile` | `AzureAksClusterLoadBalancerProfile` |  |  |  |
| `spec.networkProfile.loadBalancerProfile.outboundPortsAllocated` | `int32` |  |  |  |
| `spec.networkProfile.loadBalancerProfile.idleTimeoutInMinutes` | `int32` |  |  |  |
| `spec.networkProfile.loadBalancerProfile.managedOutboundIpCount` | `int32` |  |  |  |
| `spec.networkProfile.loadBalancerProfile.managedOutboundIpv6Count` | `int32` |  |  |  |
| `spec.networkProfile.loadBalancerProfile.outboundIpPrefixIds` | `[]string \| valueFrom` |  |  | AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`) |
| `spec.networkProfile.loadBalancerProfile.outboundIpAddressIds` | `[]string \| valueFrom` |  |  | AzurePublicIp (`status.outputs.public_ip_id`) |
| `spec.networkProfile.loadBalancerProfile.backendPoolType` | `enum` |  |  |  |
| `spec.networkProfile.natGatewayProfile` | `AzureAksClusterNatGatewayProfile` |  |  |  |
| `spec.networkProfile.natGatewayProfile.idleTimeoutInMinutes` | `int32` |  |  |  |
| `spec.networkProfile.natGatewayProfile.managedOutboundIpCount` | `int32` |  |  |  |
| `spec.networkProfile.advancedNetworking` | `AzureAksClusterAdvancedNetworking` |  |  |  |
| `spec.networkProfile.advancedNetworking.observabilityEnabled` | `bool` |  |  |  |
| `spec.networkProfile.advancedNetworking.securityEnabled` | `bool` |  |  |  |
| `spec.autoScalerProfile` | `AzureAksClusterAutoScalerProfile` |  |  |  |
| `spec.autoScalerProfile.balanceSimilarNodeGroups` | `bool` |  |  |  |
| `spec.autoScalerProfile.daemonsetEvictionForEmptyNodesEnabled` | `bool` |  |  |  |
| `spec.autoScalerProfile.daemonsetEvictionForOccupiedNodesEnabled` | `bool` |  | `true` |  |
| `spec.autoScalerProfile.expander` | `enum` |  |  |  |
| `spec.autoScalerProfile.ignoreDaemonsetsUtilizationEnabled` | `bool` |  |  |  |
| `spec.autoScalerProfile.maxGracefulTerminationSec` | `int32` |  |  |  |
| `spec.autoScalerProfile.maxNodeProvisioningTime` | `string` |  |  |  |
| `spec.autoScalerProfile.maxUnreadyNodes` | `int32` |  |  |  |
| `spec.autoScalerProfile.maxUnreadyPercentage` | `int32` |  |  |  |
| `spec.autoScalerProfile.newPodScaleUpDelay` | `string` |  |  |  |
| `spec.autoScalerProfile.scanInterval` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownDelayAfterAdd` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownDelayAfterDelete` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownDelayAfterFailure` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownUnneeded` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownUnready` | `string` |  |  |  |
| `spec.autoScalerProfile.scaleDownUtilizationThreshold` | `string` |  |  |  |
| `spec.autoScalerProfile.emptyBulkDeleteMax` | `int32` |  |  |  |
| `spec.autoScalerProfile.skipNodesWithLocalStorage` | `bool` |  |  |  |
| `spec.autoScalerProfile.skipNodesWithSystemPods` | `bool` |  | `true` |  |
| `spec.automaticUpgradeChannel` | `enum` |  |  |  |
| `spec.nodeOsUpgradeChannel` | `enum` |  |  |  |
| `spec.maintenanceWindow` | `AzureAksClusterMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.allowed` | `[]AzureAksClusterMaintenanceWindowAllowed` |  |  |  |
| `spec.maintenanceWindow.allowed[].day` | `enum` |  |  |  |
| `spec.maintenanceWindow.allowed[].hours` | `[]int32` | yes |  |  |
| `spec.maintenanceWindow.notAllowed` | `[]AzureAksClusterMaintenanceWindowNotAllowed` |  |  |  |
| `spec.maintenanceWindow.notAllowed[].start` | `string` | yes |  |  |
| `spec.maintenanceWindow.notAllowed[].end` | `string` | yes |  |  |
| `spec.maintenanceWindowAutoUpgrade` | `AzureAksClusterMaintenanceWindowSchedule` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.frequency` | `enum` | yes |  |  |
| `spec.maintenanceWindowAutoUpgrade.interval` | `int32` | yes |  |  |
| `spec.maintenanceWindowAutoUpgrade.duration` | `int32` | yes |  |  |
| `spec.maintenanceWindowAutoUpgrade.dayOfWeek` | `enum` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.weekIndex` | `enum` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.dayOfMonth` | `int32` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.startDate` | `string` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.startTime` | `string` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.utcOffset` | `string` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.notAllowed` | `[]AzureAksClusterMaintenanceWindowNotAllowed` |  |  |  |
| `spec.maintenanceWindowAutoUpgrade.notAllowed[].start` | `string` | yes |  |  |
| `spec.maintenanceWindowAutoUpgrade.notAllowed[].end` | `string` | yes |  |  |
| `spec.maintenanceWindowNodeOs` | `AzureAksClusterMaintenanceWindowSchedule` |  |  |  |
| `spec.maintenanceWindowNodeOs.frequency` | `enum` | yes |  |  |
| `spec.maintenanceWindowNodeOs.interval` | `int32` | yes |  |  |
| `spec.maintenanceWindowNodeOs.duration` | `int32` | yes |  |  |
| `spec.maintenanceWindowNodeOs.dayOfWeek` | `enum` |  |  |  |
| `spec.maintenanceWindowNodeOs.weekIndex` | `enum` |  |  |  |
| `spec.maintenanceWindowNodeOs.dayOfMonth` | `int32` |  |  |  |
| `spec.maintenanceWindowNodeOs.startDate` | `string` |  |  |  |
| `spec.maintenanceWindowNodeOs.startTime` | `string` |  |  |  |
| `spec.maintenanceWindowNodeOs.utcOffset` | `string` |  |  |  |
| `spec.maintenanceWindowNodeOs.notAllowed` | `[]AzureAksClusterMaintenanceWindowNotAllowed` |  |  |  |
| `spec.maintenanceWindowNodeOs.notAllowed[].start` | `string` | yes |  |  |
| `spec.maintenanceWindowNodeOs.notAllowed[].end` | `string` | yes |  |  |
| `spec.omsAgent` | `AzureAksClusterOmsAgent` |  |  |  |
| `spec.omsAgent.logAnalyticsWorkspaceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.omsAgent.msiAuthForMonitoringEnabled` | `bool` |  |  |  |
| `spec.keyVaultSecretsProvider` | `AzureAksClusterKeyVaultSecretsProvider` |  |  |  |
| `spec.keyVaultSecretsProvider.secretRotationEnabled` | `bool` |  |  |  |
| `spec.keyVaultSecretsProvider.secretRotationInterval` | `string` |  |  |  |
| `spec.keyVaultSecretsProvider.enabled` | `bool` |  | `true` |  |
| `spec.azurePolicyEnabled` | `bool` |  |  |  |
| `spec.microsoftDefender` | `AzureAksClusterMicrosoftDefender` |  |  |  |
| `spec.microsoftDefender.logAnalyticsWorkspaceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.monitorMetrics` | `AzureAksClusterMonitorMetrics` |  |  |  |
| `spec.monitorMetrics.annotationsAllowed` | `string` |  |  |  |
| `spec.monitorMetrics.labelsAllowed` | `string` |  |  |  |
| `spec.monitorMetrics.enabled` | `bool` |  | `true` |  |
| `spec.ingressApplicationGateway` | `AzureAksClusterIngressApplicationGateway` |  |  |  |
| `spec.ingressApplicationGateway.gatewayId` | `string \| valueFrom` |  |  | AzureApplicationGateway (`status.outputs.application_gateway_id`) |
| `spec.ingressApplicationGateway.gatewayName` | `string` |  |  |  |
| `spec.ingressApplicationGateway.subnetCidr` | `string` |  |  |  |
| `spec.ingressApplicationGateway.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.aciConnectorLinux` | `AzureAksClusterAciConnectorLinux` |  |  |  |
| `spec.aciConnectorLinux.subnetName` | `string` | yes |  |  |
| `spec.confidentialComputing` | `AzureAksClusterConfidentialComputing` |  |  |  |
| `spec.confidentialComputing.sgxQuoteHelperEnabled` | `bool` |  |  |  |
| `spec.confidentialComputing.enabled` | `bool` |  | `true` |  |
| `spec.webAppRouting` | `AzureAksClusterWebAppRouting` |  |  |  |
| `spec.webAppRouting.dnsZoneIds` | `[]string \| valueFrom` |  |  | AzureDnsZone (`status.outputs.zone_id`) |
| `spec.webAppRouting.defaultNginxController` | `enum` |  |  |  |
| `spec.serviceMeshProfile` | `AzureAksClusterServiceMeshProfile` |  |  |  |
| `spec.serviceMeshProfile.mode` | `enum` | yes |  |  |
| `spec.serviceMeshProfile.revisions` | `[]string` | yes |  |  |
| `spec.serviceMeshProfile.internalIngressGatewayEnabled` | `bool` |  |  |  |
| `spec.serviceMeshProfile.externalIngressGatewayEnabled` | `bool` |  |  |  |
| `spec.serviceMeshProfile.certificateAuthority` | `AzureAksClusterServiceMeshCertificateAuthority` |  |  |  |
| `spec.serviceMeshProfile.certificateAuthority.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.serviceMeshProfile.certificateAuthority.rootCertObjectName` | `string` | yes |  |  |
| `spec.serviceMeshProfile.certificateAuthority.certChainObjectName` | `string` | yes |  |  |
| `spec.serviceMeshProfile.certificateAuthority.certObjectName` | `string` | yes |  |  |
| `spec.serviceMeshProfile.certificateAuthority.keyObjectName` | `string` | yes |  |  |
| `spec.storageProfile` | `AzureAksClusterStorageProfile` |  |  |  |
| `spec.storageProfile.blobDriverEnabled` | `bool` |  |  |  |
| `spec.storageProfile.diskDriverEnabled` | `bool` |  | `true` |  |
| `spec.storageProfile.fileDriverEnabled` | `bool` |  | `true` |  |
| `spec.storageProfile.snapshotControllerEnabled` | `bool` |  | `true` |  |
| `spec.workloadAutoscalerProfile` | `AzureAksClusterWorkloadAutoscalerProfile` |  |  |  |
| `spec.workloadAutoscalerProfile.kedaEnabled` | `bool` |  |  |  |
| `spec.workloadAutoscalerProfile.verticalPodAutoscalerEnabled` | `bool` |  |  |  |
| `spec.keyManagementService` | `AzureAksClusterKeyManagementService` |  |  |  |
| `spec.keyManagementService.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.key_id`) |
| `spec.keyManagementService.keyVaultNetworkAccess` | `enum` |  |  |  |
| `spec.httpProxyConfig` | `AzureAksClusterHttpProxyConfig` |  |  |  |
| `spec.httpProxyConfig.httpProxy` | `string` |  |  |  |
| `spec.httpProxyConfig.httpsProxy` | `string` |  |  |  |
| `spec.httpProxyConfig.noProxy` | `[]string` |  |  |  |
| `spec.httpProxyConfig.trustedCa` | `string` (sensitive) |  |  |  |
| `spec.linuxProfile` | `AzureAksClusterLinuxProfile` |  |  |  |
| `spec.linuxProfile.adminUsername` | `string` | yes |  |  |
| `spec.linuxProfile.sshPublicKey` | `string` | yes |  |  |
| `spec.windowsProfile` | `AzureAksClusterWindowsProfile` |  |  |  |
| `spec.windowsProfile.adminUsername` | `string` | yes |  |  |
| `spec.windowsProfile.adminPassword` | `string` (sensitive) | yes |  |  |
| `spec.windowsProfile.license` | `enum` |  |  |  |
| `spec.windowsProfile.gmsa` | `AzureAksClusterWindowsGmsa` |  |  |  |
| `spec.windowsProfile.gmsa.dnsServer` | `string` |  |  |  |
| `spec.windowsProfile.gmsa.rootDomain` | `string` |  |  |  |
| `spec.imageCleanerEnabled` | `bool` |  |  |  |
| `spec.imageCleanerIntervalHours` | `int32` |  |  |  |
| `spec.costAnalysisEnabled` | `bool` |  |  |  |
| `spec.runCommandEnabled` | `bool` |  | `true` |  |
| `spec.diskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.nodeResourceGroup` | `string` |  |  |  |
| `spec.customCaTrustCertificatesBase64` | `[]string` |  |  |  |
| `spec.bootstrapProfile` | `AzureAksClusterBootstrapProfile` |  |  |  |
| `spec.bootstrapProfile.artifactSource` | `enum` |  |  |  |
| `spec.bootstrapProfile.containerRegistryId` | `string \| valueFrom` |  |  | AzureContainerRegistry (`status.outputs.container_registry_id`) |
| `spec.nodeProvisioningProfile` | `AzureAksClusterNodeProvisioningProfile` |  |  |  |
| `spec.nodeProvisioningProfile.mode` | `enum` |  |  |  |
| `spec.nodeProvisioningProfile.defaultNodePools` | `enum` |  |  |  |
| `spec.upgradeOverride` | `AzureAksClusterUpgradeOverride` |  |  |  |
| `spec.upgradeOverride.forceUpgradeEnabled` | `bool` | yes |  |  |
| `spec.upgradeOverride.effectiveUntil` | `string` |  |  |  |
| `spec.aiToolchainOperatorEnabled` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group in which to create the cluster.
Can be a literal name or a reference to an AzureResourceGroup output.
Changing the resource group replaces the cluster.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.region

`string` · required

The Azure region for the cluster's control plane and default node pool,
e.g. "eastus", "westeurope". Changing the region replaces the cluster.

- rule: {"required":true}

### spec.name

`string` · required

The name of the managed cluster, unique within the resource group.
1-63 characters: alphanumerics, underscores, and hyphens; must start
and end with an alphanumeric. Changing the name replaces the cluster.

- rule: AKS cluster names start and end with a letter or number and may contain alphanumerics, underscores, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.dnsPrefix

`string`

DNS prefix for the cluster's public API-server FQDN
(<prefix>-<hash>.hcp.<region>.azmk8s.io). 1-54 characters: letters,
numbers, and hyphens; must begin and end with a letter or number.
Leave unset to derive it from the cluster name -- the right call for
nearly everyone. Mutually exclusive with dns_prefix_private_cluster.
Changing it replaces the cluster.

- rule: DNS prefixes begin and end with a letter or number, contain only letters, numbers, and hyphens, and are 1-54 characters long

### spec.dnsPrefixPrivateCluster

`string`

DNS prefix for the PRIVATE cluster's API-server record in its private
DNS zone. Only valid for private clusters (private_cluster_enabled),
and mutually exclusive with dns_prefix. Leave unset for private
clusters too -- the modules derive a prefix from the cluster name.
Changing it replaces the cluster.

### spec.kubernetesVersion

`string`

Kubernetes version for the control plane, e.g. "1.35" (minor-version
aliases pick the latest GA patch) or an exact "1.35.2". Leave unset to
provision the latest AKS-recommended GA version -- but PIN a version
for production clusters so upgrades happen when you choose, not when
you redeploy. Upgrades are in-place but can only move one minor
version at a time; node pools must stay within two minors of the
control plane.

### spec.skuTier

`enum`

Control-plane pricing/support tier. Unspecified applies Azure's
default (Free): no uptime SLA, 1000-node limit -- fine for dev/test.
STANDARD adds the financially-backed 99.95% SLA (with availability
zones) and 5000-node scale that production clusters need. PREMIUM
additionally unlocks AKS Long Term Support. Updatable in place.

Allowed values (use exactly as shown):

- `azure_aks_cluster_sku_tier_unspecified` -- Not specified: Azure's default (Free) -- no uptime SLA.
- `FREE` -- Free control plane: no SLA, up to 1000 nodes. Dev/test only.
- `STANDARD` -- Standard tier: financially-backed 99.95% API-server SLA (with availability zones), 5000-node scale. The production baseline.
- `PREMIUM` -- Premium tier: Standard plus AKS Long Term Support eligibility.

### spec.supportPlan

`enum`

Kubernetes-version support plan. Unspecified applies Azure's default
(KUBERNETES_OFFICIAL): versions supported per upstream community
windows. AKS_LONG_TERM_SUPPORT extends a version's support to 2 years
and requires the PREMIUM sku_tier.

Allowed values (use exactly as shown):

- `azure_aks_cluster_support_plan_unspecified` -- Not specified: Azure's default (KubernetesOfficial).
- `KUBERNETES_OFFICIAL` -- Community-aligned support windows for each Kubernetes version.
- `AKS_LONG_TERM_SUPPORT` -- Two-year long-term support for eligible versions; requires PREMIUM.

### spec.defaultNodePool

`AzureAksClusterDefaultNodePool` · required

The default (system) node pool Azure requires at cluster creation --
the one pool whose lifecycle is genuinely coupled to the cluster's.
It runs critical system pods (CoreDNS, metrics-server, konnectivity)
and is always Linux in System mode, which is why it carries no
os_type/mode/spot knobs. Add every other pool as a standalone
AzureAksNodePool resource referencing this cluster.

- rule: {"required":true}
- rule: With auto_scaling_enabled, set min_count and max_count (min <= max); without it, leave both unset
- rule: Set node_count (1-1000) when autoscaling is disabled -- the pool needs a fixed size
- rule: node_public_ip_prefix_id requires node_public_ip_enabled to be true
- rule: The default node pool is always Linux -- choose a Linux os_sku (Ubuntu or Azure Linux); Windows pools are standalone AzureAksNodePool resources

### spec.defaultNodePool.name

`string` · required

Agent-pool name: 1-12 lowercase letters and numbers, starting with a
letter. Renaming normally replaces the pool -- set
temporary_name_for_rotation to let AKS rotate through a stand-in pool
instead of tearing the cluster's system pool down first.

- rule: Node pool names are 1-12 lowercase letters and numbers and start with a letter
- rule: {"required":true}

### spec.defaultNodePool.vmSize

`string` · required

Azure VM size for the pool's nodes, e.g. "Standard_D4s_v5". System
pools need at least 2 vCPUs and 4 GiB memory. Changing the size
rotates the pool (see temporary_name_for_rotation).

- default: `Standard_D4s_v5`
- rule: {"required":true}

### spec.defaultNodePool.nodeCount

`int32`

Fixed node count (1-1000) when autoscaling is off, or the initial
count when it is on. With autoscaling enabled, leave unset to let the
autoscaler own the count from the start.

- rule: node_count must be between 1 and 1000

### spec.defaultNodePool.autoScalingEnabled

`bool`

Whether the cluster autoscaler manages this pool's node count between
min_count and max_count. Tune the cluster-wide behavior through the
cluster's auto_scaler_profile.

### spec.defaultNodePool.minCount

`int32`

Minimum node count for autoscaling (1-1000; the default pool cannot
scale to zero -- it hosts the system pods). Requires
auto_scaling_enabled.

- rule: min_count must be between 1 and 1000

### spec.defaultNodePool.maxCount

`int32`

Maximum node count for autoscaling (1-1000). Requires
auto_scaling_enabled.

- rule: max_count must be between 1 and 1000

### spec.defaultNodePool.maxPods

`int32`

Maximum pods per node. Unset applies Azure's plugin-dependent default
(250 for Azure CNI overlay, 30 for traditional Azure CNI). Set at
pool creation; raising it later rotates the pool.

### spec.defaultNodePool.zones

`[]string`

Availability zones to spread nodes across, e.g. ["1", "2", "3"].
Three-zone system pools are the production posture (and what the
STANDARD tier's 99.95% SLA assumes). Leave empty for regions without
zones. Changing zones rotates the pool.

- rule: {"repeated":{"items":{"string":{"in":["1","2","3"]}}}}

### spec.defaultNodePool.vnetSubnetId

`string | valueFrom`

The subnet nodes deploy into. Leave unset to let AKS create and
manage its own network -- Azure's default and the simplest correct
start. Reference an AzureSubnet to place nodes in YOUR network
(required for private clusters, custom routing, or NSG control);
the cluster identity then needs Network Contributor on it. Changing
the subnet replaces the pool.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.defaultNodePool.podSubnetId

`string | valueFrom`

A separate subnet for POD IPs (traditional Azure CNI with dynamic pod
IP allocation). Only meaningful with a vnet_subnet_id and the
non-overlay Azure CNI; overlay mode ignores it (pods use pod_cidr).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.defaultNodePool.osDiskSizeGb

`int32`

OS disk size in GiB. Unset applies Azure's VM-size-dependent default.

### spec.defaultNodePool.osDiskType

`enum`

OS disk placement. Unspecified applies Azure's default (MANAGED): a
persistent managed disk. EPHEMERAL places the OS disk on node-local
storage -- faster and free, the right choice whenever the VM size's
cache disk fits the image; nodes are cattle, their OS disks need no
durability.

Allowed values (use exactly as shown):

- `azure_aks_cluster_os_disk_type_unspecified` -- Not specified: Azure's default (Managed).
- `MANAGED` -- Persistent managed disk -- survives node restarts, billed separately.
- `EPHEMERAL` -- Node-local ephemeral disk -- faster, free, right for stateless nodes whenever the VM size's cache disk fits the OS image.

### spec.defaultNodePool.kubeletDiskType

`enum`

Where kubelet state (image layers, emptyDir volumes) lives.
Unspecified applies Azure's default (OS disk). TEMPORARY uses the
VM's temp disk for higher IOPS on image-churn-heavy nodes.

Allowed values (use exactly as shown):

- `azure_aks_cluster_kubelet_disk_type_unspecified` -- Not specified: Azure's default (the OS disk).
- `OS` -- Kubelet state on the OS disk.
- `TEMPORARY` -- Kubelet state on the VM's temporary disk (higher IOPS for image churn; contents lost on deallocation, which kubelet state tolerates).

### spec.defaultNodePool.osSku

`enum`

Node OS image. Unspecified applies Azure's default Linux image
(Ubuntu). AZURE_LINUX is Microsoft's own minimal, security-hardened
distro -- smaller attack surface and faster boots; the direction AKS
is heading. Version-pinned values (UBUNTU_2204/2404, AZURE_LINUX_3)
pin the OS major independent of Kubernetes version.

Allowed values (use exactly as shown):

- `azure_aks_cluster_os_sku_unspecified` -- Not specified: Azure's default image for the pool's OS type (currently Ubuntu for Linux, Windows Server 2022 for Windows).
- `UBUNTU` -- Ubuntu, version following AKS's default for the Kubernetes version.
- `UBUNTU_2204` -- Ubuntu 22.04 LTS, pinned.
- `UBUNTU_2404` -- Ubuntu 24.04 LTS, pinned.
- `AZURE_LINUX` -- Azure Linux (Microsoft's minimal container-host distro), version following AKS's default.
- `AZURE_LINUX_3` -- Azure Linux 3, pinned.
- `WINDOWS_2019` -- Windows Server 2019 (standalone Windows pools only; retired for new pools on Kubernetes >= 1.33).
- `WINDOWS_2022` -- Windows Server 2022 (standalone Windows pools only).

### spec.defaultNodePool.orchestratorVersion

`string`

Kubernetes version for the pool's nodes. Unset follows the control
plane's version -- correct for the default pool in almost all cases.

### spec.defaultNodePool.nodeLabels

`map<string, string>`

Kubernetes labels applied to this pool's nodes, for scheduling
(nodeSelector/affinity) against system-pool nodes.

### spec.defaultNodePool.onlyCriticalAddonsEnabled

`bool`

Whether the pool is tainted CriticalAddonsOnly=true:NoSchedule so
ONLY system pods schedule here. The recommended posture once real
workloads run in dedicated AzureAksNodePool resources: it stops app
pods from starving CoreDNS.

### spec.defaultNodePool.fipsEnabled

`bool`

Whether nodes get FIPS 140-2 validated OS images (compliance
environments). Changing it rotates the pool.

### spec.defaultNodePool.hostEncryptionEnabled

`bool`

Whether host-based encryption is enabled: data on the node's temp
disks and disk caches is encrypted at rest. Requires the
EncryptionAtHost feature on the subscription. Changing it rotates the
pool.

### spec.defaultNodePool.nodePublicIpEnabled

`bool`

Whether each node gets its own public IP -- niche direct-node-ingress
patterns (game servers). Egress normally flows through the cluster
load balancer or NAT gateway instead.

### spec.defaultNodePool.nodePublicIpPrefixId

`string | valueFrom`

Public IP prefix to allocate node public IPs from, so node IPs come
from one known, allowlistable CIDR. Requires node_public_ip_enabled.

- references: AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIpPrefix, name: <that resource's name>, fieldPath: status.outputs.public_ip_prefix_id}} -- a bare string does not parse

### spec.defaultNodePool.gpuInstance

`enum`

GPU Multi-Instance GPU (MIG) profile for A100 sizes, partitioning
each physical GPU into isolated slices (e.g. "MIG1g" = 7 slices).
Only for MIG-capable VM sizes. Changing it rotates the pool.

Allowed values (use exactly as shown):

- `azure_aks_cluster_gpu_instance_unspecified` -- Not specified: no MIG partitioning.
- `MIG1G` -- Seven 1g.5gb slices per GPU.
- `MIG2G` -- Three 2g.10gb slices per GPU.
- `MIG3G` -- Two 3g.20gb slices per GPU.
- `MIG4G` -- One 4g.20gb slice per GPU.
- `MIG7G` -- One 7g.40gb slice (the whole GPU as a single MIG device).

### spec.defaultNodePool.gpuDriver

`enum`

Whether AKS installs the NVIDIA GPU driver on GPU nodes. Unspecified
applies Azure's default (install when the VM size has a GPU). NONE
skips installation for teams that run the GPU operator themselves.

Allowed values (use exactly as shown):

- `azure_aks_cluster_gpu_driver_unspecified` -- Not specified: Azure's default (install the driver on GPU sizes).
- `INSTALL` -- AKS installs the NVIDIA driver.
- `NONE` -- AKS skips driver installation -- for self-managed GPU operators.

### spec.defaultNodePool.proximityPlacementGroupId

`string`

ARM id of a Proximity Placement Group to co-locate nodes for minimal
inter-node latency (HPC). Plain ARM id (no Planton kind). Changing it
rotates the pool.

### spec.defaultNodePool.hostGroupId

`string`

ARM id of a Dedicated Host Group to place nodes on your isolated
physical hosts (compliance isolation). Plain ARM id. Changing it
rotates the pool.

### spec.defaultNodePool.capacityReservationGroupId

`string`

ARM id of a Capacity Reservation Group to draw guaranteed compute
capacity from. Plain ARM id. Changing it rotates the pool.

### spec.defaultNodePool.scaleDownMode

`enum`

What scale-down does with removed nodes. Unspecified applies Azure's
default (DELETE): nodes are deleted and stop billing. DEALLOCATE
stops them but keeps disks for faster scale-up at storage cost.

Allowed values (use exactly as shown):

- `azure_aks_cluster_scale_down_mode_unspecified` -- Not specified: Azure's default (Delete).
- `DELETE` -- Removed nodes are deleted -- billing stops entirely.
- `DEALLOCATE` -- Removed nodes are stopped (deallocated) -- compute billing stops but disks persist for faster scale-up.

### spec.defaultNodePool.snapshotId

`string`

ARM id of a node-pool snapshot to source this pool's configuration
from (replicating a known-good pool config across clusters). Plain
ARM id.

### spec.defaultNodePool.workloadRuntime

`enum`

Container runtime class. Unspecified runs standard containers
(OCIContainer, Azure's default). KATA_MSHV_VM_ISOLATION runs each pod
in a lightweight utility VM for kernel isolation (requires supporting
VM sizes).

Allowed values (use exactly as shown):

- `azure_aks_cluster_workload_runtime_unspecified` -- Not specified: Azure's default (OCIContainer).
- `OCI_CONTAINER` -- Standard OCI containers.
- `KATA_MSHV_VM_ISOLATION` -- Kata Containers on Microsoft Hyper-V: each pod in a lightweight utility VM for kernel-level isolation.

### spec.defaultNodePool.ultraSsdEnabled

`bool`

Whether nodes may use Ultra SSD (zone-pinned, highest-IOPS) data
disks. Requires zones. Changing it rotates the pool.

### spec.defaultNodePool.temporaryNameForRotation

`string`

Stand-in pool name AKS uses to rotate this pool through otherwise
replace-forcing changes (vm_size, os_disk_type, fips_enabled...):
a temporary pool with this name carries the system pods while the
real pool is rebuilt. Same format as name. Set it proactively on
production clusters.

- rule: Node pool names are 1-12 lowercase letters and numbers and start with a letter

### spec.defaultNodePool.kubeletConfig

`AzureAksClusterKubeletConfig`

Kubelet tuning for this pool's nodes -- CPU manager policy, image GC
thresholds, container log rotation, pid limits. Unset fields keep
AKS defaults. Changing kubelet config rotates the pool.

### spec.defaultNodePool.kubeletConfig.cpuManagerPolicy

`enum`

CPU manager policy. Unspecified keeps AKS's default ("none").
STATIC gives Guaranteed-QoS pods exclusive cores -- latency-sensitive
workloads (trading, real-time media).

Allowed values (use exactly as shown):

- `azure_aks_cluster_cpu_manager_policy_unspecified` -- Not specified: kubelet default ("none") -- shared CPU pool.
- `CPU_MANAGER_NONE` -- Shared CPU pool for all pods.
- `CPU_MANAGER_STATIC` -- Exclusive core pinning for Guaranteed-QoS pods with integer CPU requests.

### spec.defaultNodePool.kubeletConfig.cpuCfsQuotaEnabled

`bool` · optional (explicit presence)

Whether CPU CFS quota enforcement applies to containers with CPU
limits. Azure's default is true; disabling trades limit enforcement
for throttling-sensitive latency.

- default: `true`

### spec.defaultNodePool.kubeletConfig.cpuCfsQuotaPeriod

`string`

CPU CFS quota period, e.g. "100ms" (the kubelet default). Shorter
periods smooth throttling for latency-sensitive workloads.

### spec.defaultNodePool.kubeletConfig.imageGcHighThreshold

`int32`

Disk usage percentage that triggers image garbage collection (0-100).
Unset keeps the kubelet default (85).

- rule: image_gc_high_threshold is a percentage between 0 and 100

### spec.defaultNodePool.kubeletConfig.imageGcLowThreshold

`int32`

Disk usage percentage image GC frees down to (0-100). Unset keeps the
kubelet default (80).

- rule: image_gc_low_threshold is a percentage between 0 and 100

### spec.defaultNodePool.kubeletConfig.topologyManagerPolicy

`enum`

NUMA topology alignment policy for pod resources. Unspecified keeps
the kubelet default ("none").

Allowed values (use exactly as shown):

- `azure_aks_cluster_topology_manager_policy_unspecified` -- Not specified: kubelet default ("none") -- no NUMA alignment.
- `TOPOLOGY_NONE` -- No alignment.
- `BEST_EFFORT` -- Prefer aligned placement, admit regardless.
- `RESTRICTED` -- Admit only pods whose preferred alignment is achievable.
- `SINGLE_NUMA_NODE` -- Admit only pods placeable on a single NUMA node.

### spec.defaultNodePool.kubeletConfig.allowedUnsafeSysctls

`[]string`

Unsafe sysctls (or patterns like "net.*") pods may set via their
security context. Empty means none -- the safe default.

### spec.defaultNodePool.kubeletConfig.containerLogMaxSizeMb

`int32`

Maximum container log file size in MB before rotation.

### spec.defaultNodePool.kubeletConfig.containerLogMaxFiles

`int32`

Maximum rotated container log files kept per container (>= 2).

- rule: container_log_max_files must be at least 2 (the active file plus one rotation)

### spec.defaultNodePool.kubeletConfig.podMaxPid

`int32`

Maximum processes per pod. Unset keeps the kubelet default
(unlimited, -1).

### spec.defaultNodePool.linuxOsConfig

`AzureAksClusterLinuxOsConfig`

Linux kernel and OS tuning -- sysctl values, transparent huge pages,
swap file. Unset fields keep AKS defaults. Changing OS config rotates
the pool.

### spec.defaultNodePool.linuxOsConfig.sysctlConfig

`AzureAksClusterSysctlConfig`

Kernel sysctl overrides. Only the sysctls AKS allows are modeled;
each carries the ARM-enforced range in its validation.

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsAioMaxNr

`int32`

fs.aio-max-nr (65536-6553500): max concurrent async I/O requests.

- rule: fs_aio_max_nr must be between 65536 and 6553500

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsFileMax

`int32`

fs.file-max (8192-12000500): system-wide open file handle limit.

- rule: fs_file_max must be between 8192 and 12000500

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsInotifyMaxUserWatches

`int32`

fs.inotify.max_user_watches (781250-2097152): inotify watch limit --
raise for file-watching-heavy workloads (IDEs, log shippers).

- rule: fs_inotify_max_user_watches must be between 781250 and 2097152

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.fsNrOpen

`int32`

fs.nr_open (8192-20000500): per-process open file limit.

- rule: fs_nr_open must be between 8192 and 20000500

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.kernelThreadsMax

`int32`

kernel.threads-max (20-513785): system-wide thread limit.

- rule: kernel_threads_max must be between 20 and 513785

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreNetdevMaxBacklog

`int32`

net.core.netdev_max_backlog (1000-3240000): NIC ingress queue length.

- rule: net_core_netdev_max_backlog must be between 1000 and 3240000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreOptmemMax

`int32`

net.core.optmem_max (20480-4194304): per-socket ancillary buffer max.

- rule: net_core_optmem_max must be between 20480 and 4194304

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreRmemDefault

`int32`

net.core.rmem_default (212992-134217728): default socket receive
buffer.

- rule: net_core_rmem_default must be between 212992 and 134217728

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreRmemMax

`int32`

net.core.rmem_max (212992-134217728): max socket receive buffer.

- rule: net_core_rmem_max must be between 212992 and 134217728

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreSomaxconn

`int32`

net.core.somaxconn (4096-3240000): listen backlog limit.

- rule: net_core_somaxconn must be between 4096 and 3240000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreWmemDefault

`int32`

net.core.wmem_default (212992-134217728): default socket send buffer.

- rule: net_core_wmem_default must be between 212992 and 134217728

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netCoreWmemMax

`int32`

net.core.wmem_max (212992-134217728): max socket send buffer.

- rule: net_core_wmem_max must be between 212992 and 134217728

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMin

`int32`

net.ipv4.ip_local_port_range minimum (1024-60999).

- rule: net_ipv4_ip_local_port_range_min must be between 1024 and 60999

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMax

`int32`

net.ipv4.ip_local_port_range maximum (32768-65535).

- rule: net_ipv4_ip_local_port_range_max must be between 32768 and 65535

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh1

`int32`

net.ipv4.neigh.default.gc_thresh1 (128-80000): ARP cache soft floor.

- rule: net_ipv4_neigh_default_gc_thresh1 must be between 128 and 80000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh2

`int32`

net.ipv4.neigh.default.gc_thresh2 (512-90000): ARP cache soft ceiling.

- rule: net_ipv4_neigh_default_gc_thresh2 must be between 512 and 90000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh3

`int32`

net.ipv4.neigh.default.gc_thresh3 (1024-100000): ARP cache hard limit.

- rule: net_ipv4_neigh_default_gc_thresh3 must be between 1024 and 100000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpFinTimeout

`int32`

net.ipv4.tcp_fin_timeout (5-120): FIN-WAIT-2 hold seconds.

- rule: net_ipv4_tcp_fin_timeout must be between 5 and 120

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveIntvl

`int32`

net.ipv4.tcp_keepalive_intvl (10-90): keepalive probe interval.

- rule: net_ipv4_tcp_keepalive_intvl must be between 10 and 90

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveProbes

`int32`

net.ipv4.tcp_keepalive_probes (1-15): unanswered probes before drop.

- rule: net_ipv4_tcp_keepalive_probes must be between 1 and 15

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveTime

`int32`

net.ipv4.tcp_keepalive_time (30-432000): idle seconds before
keepalives start.

- rule: net_ipv4_tcp_keepalive_time must be between 30 and 432000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpMaxSynBacklog

`int32`

net.ipv4.tcp_max_syn_backlog (128-3240000): half-open connection
queue.

- rule: net_ipv4_tcp_max_syn_backlog must be between 128 and 3240000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpMaxTwBuckets

`int32`

net.ipv4.tcp_max_tw_buckets (8000-1440000): TIME-WAIT socket cap.

- rule: net_ipv4_tcp_max_tw_buckets must be between 8000 and 1440000

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netIpv4TcpTwReuse

`bool`

net.ipv4.tcp_tw_reuse: allow reusing TIME-WAIT sockets for new
outbound connections.

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackBuckets

`int32`

net.netfilter.nf_conntrack_buckets (65536-524288): conntrack hash
size.

- rule: net_netfilter_nf_conntrack_buckets must be between 65536 and 524288

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackMax

`int32`

net.netfilter.nf_conntrack_max (131072-2097152): tracked connection
cap -- raise for high-connection-count proxies.

- rule: net_netfilter_nf_conntrack_max must be between 131072 and 2097152

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmMaxMapCount

`int32`

vm.max_map_count (65530-262144): memory-map areas per process --
Elasticsearch famously needs 262144.

- rule: vm_max_map_count must be between 65530 and 262144

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmSwappiness

`int32`

vm.swappiness (0-100): kernel swap eagerness.

- rule: vm_swappiness must be between 0 and 100

### spec.defaultNodePool.linuxOsConfig.sysctlConfig.vmVfsCachePressure

`int32`

vm.vfs_cache_pressure (0-100): dentry/inode cache reclaim pressure.

- rule: vm_vfs_cache_pressure must be between 0 and 100

### spec.defaultNodePool.linuxOsConfig.transparentHugePage

`enum`

Transparent Huge Pages mode. Unspecified keeps the OS default
("always"). Databases with sparse access patterns often want
MADVISE or NEVER.

Allowed values (use exactly as shown):

- `azure_aks_cluster_transparent_huge_page_unspecified` -- Not specified: the OS default ("always").
- `THP_ALWAYS` -- THP for all memory regions.
- `THP_MADVISE` -- THP only for madvise(MADV_HUGEPAGE) regions.
- `THP_NEVER` -- THP disabled.

### spec.defaultNodePool.linuxOsConfig.transparentHugePageDefrag

`enum`

Transparent Huge Pages defrag behavior. Unspecified keeps the OS
default ("madvise").

Allowed values (use exactly as shown):

- `azure_aks_cluster_transparent_huge_page_defrag_unspecified` -- Not specified: the OS default ("madvise").
- `DEFRAG_ALWAYS` -- Synchronous defrag on every THP allocation.
- `DEFRAG_DEFER` -- Defer defrag to kswapd.
- `DEFRAG_DEFER_MADVISE` -- Defer generally, defrag synchronously for madvise regions.
- `DEFRAG_MADVISE` -- Defrag synchronously only for madvise regions.
- `DEFRAG_NEVER` -- Never defrag.

### spec.defaultNodePool.linuxOsConfig.swapFileSizeMb

`int32`

Swap file size in MB on each node. Unset means no swap -- the
Kubernetes-recommended default.

### spec.defaultNodePool.nodeNetworkProfile

`AzureAksClusterNodeNetworkProfile`

Node-level network hardening: allowed host ports, application
security groups for nodes, and tags on node public IPs.

### spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts

`[]AzureAksClusterAllowedHostPorts`

Host port ranges pods may bind on the node -- each entry opens a
port range/protocol in the node's network security rules.

### spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].portStart

`int32`

Start of the port range (1-65535).

- rule: port_start must be between 1 and 65535

### spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].portEnd

`int32`

End of the port range (1-65535, >= port_start).

- rule: port_end must be between 1 and 65535

### spec.defaultNodePool.nodeNetworkProfile.allowedHostPorts[].protocol

`enum`

Protocol for the range.

Allowed values (use exactly as shown):

- `azure_aks_cluster_host_port_protocol_unspecified` -- Not specified.
- `TCP` -- TCP.
- `UDP` -- UDP.

### spec.defaultNodePool.nodeNetworkProfile.applicationSecurityGroupIds

`[]string`

ARM ids of Application Security Groups the pool's nodes join, so NSG
rules can target "the AKS nodes" as a group. Plain ARM ids.

### spec.defaultNodePool.nodeNetworkProfile.nodePublicIpTags

`map<string, string>`

Azure tags applied to the node PUBLIC IPs (with
node_public_ip_enabled), e.g. routing preference tags. Set at pool
creation.

### spec.defaultNodePool.upgradeSettings

`AzureAksClusterDefaultNodePoolUpgradeSettings`

How node upgrades roll through the pool: surge sizing, drain
behavior, and soak time between nodes.

### spec.defaultNodePool.upgradeSettings.maxSurge

`string` · required

Extra nodes added during an upgrade, as a count ("2") or percentage
("10%" -- AKS's recommended default). More surge = faster upgrades,
more temporary cost.

- rule: {"required":true}

### spec.defaultNodePool.upgradeSettings.drainTimeoutInMinutes

`int32`

Minutes to wait for a node to drain before giving up (honoring pod
disruption budgets). Unset keeps Azure's default (30).

### spec.defaultNodePool.upgradeSettings.nodeSoakDurationInMinutes

`int32`

Minutes to soak (wait) after each upgraded node before the next
(0-30). Unset upgrades continuously.

- rule: node_soak_duration_in_minutes must be between 0 and 30

### spec.defaultNodePool.upgradeSettings.undrainableNodeBehavior

`enum`

What happens to a node that will not drain (PDB-blocked).
Unspecified keeps Azure's default (the upgrade fails). CORDON
quarantines the node and proceeds; SCHEDULE lets pods return to it.

Allowed values (use exactly as shown):

- `azure_aks_cluster_undrainable_node_behavior_unspecified` -- Not specified: Azure's default -- the upgrade errors on an undrainable node.
- `CORDON` -- Cordon the undrainable node into quarantine and continue.
- `SCHEDULE` -- Leave the node schedulable and continue.

### spec.defaultNodePool.tags

`map<string, string>`

Free-form tags applied to the pool's VM scale set, merged over the
Planton-derived resource tags; a user tag with the same key wins.

### spec.identity

`AzureAksClusterIdentity`

The cluster's managed identity -- how the AKS control plane itself
authenticates to Azure (to manage the node resource group, attach
disks, create load balancers). Leave unset for a system-assigned
identity, created and rotated by Azure -- right for most clusters.
Configure USER_ASSIGNED (with identity_ids) when the cluster identity
needs pre-provisioned grants -- e.g. Private DNS Zone Contributor on a
BYO private zone, or Network Contributor on a BYO subnet -- so the
grants exist before the cluster does.

- rule: identity_ids is required for USER_ASSIGNED and must be empty otherwise

### spec.identity.type

`enum`

Identity flavor. Unspecified applies SYSTEM_ASSIGNED -- Azure creates
and rotates the identity with the cluster.

Allowed values (use exactly as shown):

- `azure_aks_cluster_identity_type_unspecified` -- Not specified: SystemAssigned.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the cluster.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identity (set identity_ids).

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED: the user-assigned identities the control plane
runs as (AKS uses the first). Reference AzureUserAssignedIdentity
resources so grants (Private DNS Zone Contributor, Network
Contributor) can be composed before cluster creation.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.kubeletIdentity

`AzureAksClusterKubeletIdentity`

The kubelet identity -- the identity NODES use to pull images from
Azure Container Registry and access other Azure resources. Leave
unset to let AKS manage one. Setting it requires a USER_ASSIGNED
cluster identity, and all three fields must be set together (the
client id and object id of the same user-assigned identity whose ARM
id is referenced). Changing any of them replaces the cluster.

### spec.kubeletIdentity.clientId

`string` · required

Client id of the kubelet's user-assigned identity.

- rule: {"required":true}

### spec.kubeletIdentity.objectId

`string` · required

Object (principal) id of the kubelet's user-assigned identity.

- rule: {"required":true}

### spec.kubeletIdentity.userAssignedIdentityId

`string | valueFrom` · required

ARM id of the kubelet's user-assigned identity.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.oidcIssuerEnabled

`bool` · optional (explicit presence)

Whether the cluster publishes an OIDC issuer endpoint. Azure's
provisioning default is off, but this spec defaults it ON: the issuer
is the foundation of workload identity federation (an
AzureFederatedIdentityCredential's `issuer` is exactly this cluster's
oidc_issuer_url output), it costs nothing, and disabling it after
enabling forces cluster replacement.

- default: `true`

### spec.workloadIdentityEnabled

`bool`

Whether to run the Azure Workload Identity webhook in the cluster,
letting pods exchange Kubernetes service-account tokens for Azure AD
tokens (secret-less access to Azure APIs). Requires the OIDC issuer.
The full trust chain is: this cluster's oidc_issuer_url output → an
AzureFederatedIdentityCredential on a user-assigned identity → Azure
role assignments on that identity.

### spec.privateClusterEnabled

`bool`

Deploy as a PRIVATE cluster: the API server gets only a private IP in
the cluster's VNet, reachable via VNet peering, VPN, or ExpressRoute.
The hardened-enterprise posture; pair with private_dns_zone_id.
Changing it replaces the cluster.

### spec.privateDnsZoneId

`string | valueFrom`

For private clusters: the Private DNS zone hosting the API server's
record. Referencing an AzurePrivateDnsZone (or a literal zone ARM id)
requires the cluster identity to hold Private DNS Zone Contributor on
it -- use a USER_ASSIGNED cluster identity so the grant can pre-exist.
The literals "System" (AKS creates and manages the zone; the default)
and "None" (public DNS resolving to the private IP) are also accepted.
Changing it replaces the cluster.

- references: AzurePrivateDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.privateClusterPublicFqdnEnabled

`bool`

For private clusters: whether to ALSO publish a public FQDN that
resolves to the API server's private IP. Useful when private-network
clients cannot use the private DNS zone (e.g. on-prem resolvers);
the endpoint itself stays private either way.

### spec.apiServerAccessProfile

`AzureAksClusterApiServerAccessProfile`

API-server access hardening for PUBLIC clusters: source-IP allowlist
and API Server VNet Integration.

### spec.apiServerAccessProfile.authorizedIpRanges

`[]string`

CIDR blocks allowed to reach the public API server. Empty admits all
source IPs -- set your office/VPN/CI ranges for any real cluster that
is not private. Ignored for private clusters.

- rule: {"repeated":{"items":{"string":{"pattern":"^(?:25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])(?:\\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])){3}(?:\\/(?:3[0-2]|[12]?[0-9]))?$"}}}}

### spec.apiServerAccessProfile.virtualNetworkIntegrationEnabled

`bool`

Whether API Server VNet Integration projects the API server into a
delegated subnet in YOUR network -- API-server traffic stays private
without full private-cluster mode.

### spec.apiServerAccessProfile.subnetId

`string | valueFrom`

The delegated subnet for API Server VNet Integration (delegation:
Microsoft.ContainerService/managedClusters).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.roleBasedAccessControlEnabled

`bool` · optional (explicit presence)

Whether Kubernetes Role-Based Access Control is enabled. Azure's
default is true, and virtually nothing legitimate runs without RBAC;
the field exists to mirror the provider surface. Disabling it forces
cluster replacement.

- default: `true`

### spec.localAccountDisabled

`bool`

Disable Kubernetes local accounts (the certificate-based admin in
kube_admin_config), forcing ALL access through Microsoft Entra ID.
The production hardening posture -- requires
azure_active_directory_role_based_access_control to be configured so
someone can still get in.

### spec.azureActiveDirectoryRoleBasedAccessControl

`AzureAksClusterAadRbac`

Microsoft Entra ID (Azure AD) integration for Kubernetes
authentication and, optionally, authorization. Configure this for any
cluster humans access: cluster admission via AAD group membership
beats distributing client certificates.

### spec.azureActiveDirectoryRoleBasedAccessControl.tenantId

`string`

Entra tenant to authenticate against. Leave unset to use the
cluster's own tenant -- correct except for cross-tenant setups.

- rule: tenant_id must be a UUID

### spec.azureActiveDirectoryRoleBasedAccessControl.azureRbacEnabled

`bool`

Whether AZURE RBAC (role assignments on the cluster's ARM scope) is
the Kubernetes authorization source, replacing in-cluster
RoleBindings. One control plane for who-can-do-what, managed like
every other Azure grant (composable with AzureRoleAssignment).

### spec.azureActiveDirectoryRoleBasedAccessControl.adminGroupObjectIds

`[]string`

Entra group object ids granted cluster-admin. The break-glass and
platform-team entry point when azure_rbac_enabled is off (and a
sensible baseline even when it is on).

- rule: {"repeated":{"items":{"string":{"pattern":"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"}}}}

### spec.networkProfile

`AzureAksClusterNetworkProfile`

The cluster's network fabric: CNI plugin and mode, network policy
engine, pod/service address spaces, outbound (egress) model, and the
managed load-balancer / NAT-gateway profiles. Leave unset for the
modern AKS default the modules apply: Azure CNI in overlay mode with
a managed Standard load balancer. Most sub-fields replace the cluster
when changed -- decide the network model up front.

- rule: CILIUM network policy requires network_data_plane CILIUM
- rule: The CILIUM data plane requires the AZURE_CNI network plugin (leave network_plugin unset or set it to AZURE_CNI)
- rule: OVERLAY network_plugin_mode requires the AZURE_CNI network plugin

### spec.networkProfile.networkPlugin

`enum`

CNI plugin. Unspecified applies the modern AKS default the modules
write explicitly: AZURE_CNI (with overlay mode unless a pod subnet
says otherwise). KUBENET is deprecated (retires March 2028) and only
kept for pre-existing estates; NONE brings your own CNI.

Allowed values (use exactly as shown):

- `azure_aks_cluster_network_plugin_unspecified` -- Not specified: the modules apply the modern AKS default, Azure CNI.
- `AZURE_CNI` -- Azure CNI -- full VNet integration; pair with OVERLAY mode (the default) or a pod subnet for traditional dynamic allocation.
- `KUBENET` -- Kubenet -- basic NAT networking. Deprecated; AKS retires it March 2028. Do not use for new clusters.
- `NETWORK_PLUGIN_NONE` -- No managed CNI -- bring your own network plugin.

### spec.networkProfile.networkPluginMode

`enum`

Azure CNI addressing mode. Unspecified with AZURE_CNI applies OVERLAY
-- pods get IPs from the private pod_cidr, not the VNet, eliminating
VNet IP exhaustion. Leave truly unset (traditional mode) only when
pods must be first-class VNet endpoints (pair with pod_subnet_id on
the pools).

Allowed values (use exactly as shown):

- `azure_aks_cluster_network_plugin_mode_unspecified` -- Not specified: with AZURE_CNI the modules apply OVERLAY (the modern default); set nothing here AND a pod_subnet_id on pools for traditional dynamic pod-IP allocation.
- `OVERLAY` -- Overlay: pods draw from pod_cidr (not VNet space). Solves IP exhaustion; the right mode for nearly every new cluster.

### spec.networkProfile.networkPolicy

`enum`

Network policy engine enforcing Kubernetes NetworkPolicy objects.
Unspecified means no enforcement -- policies are silently inert;
enable one for any multi-tenant or zero-trust cluster. CILIUM
requires the CILIUM data plane.

Allowed values (use exactly as shown):

- `azure_aks_cluster_network_policy_unspecified` -- Not specified: no NetworkPolicy enforcement.
- `NETWORK_POLICY_AZURE` -- Azure's NetworkPolicy implementation (iptables).
- `CALICO` -- Calico -- the widest NetworkPolicy feature set on the classic dataplane.
- `NETWORK_POLICY_CILIUM` -- Cilium eBPF policy -- requires the CILIUM data plane.

### spec.networkProfile.networkDataPlane

`enum`

Dataplane technology. Unspecified applies Azure's default (iptables-
based Azure dataplane). CILIUM switches to eBPF -- higher throughput,
lower latency, richer observability -- and pairs with (requires)
AZURE_CNI; it is also the only dataplane for CILIUM network policy.

Allowed values (use exactly as shown):

- `azure_aks_cluster_network_data_plane_unspecified` -- Not specified: Azure's default (iptables-based) dataplane.
- `DATA_PLANE_AZURE` -- The classic Azure dataplane.
- `DATA_PLANE_CILIUM` -- Cilium eBPF dataplane -- higher throughput and the foundation for advanced_networking observability/security.

### spec.networkProfile.dnsServiceIp

`string`

IP for Kubernetes DNS (kube-dns/CoreDNS service). Must sit inside
service_cidr. Unset applies Azure's default (".10" of the service
CIDR).

- rule: dns_service_ip must be an IPv4 address, e.g. 10.0.0.10

### spec.networkProfile.serviceCidr

`string`

CIDR for Kubernetes Services (ClusterIPs). Never routed outside the
cluster but must not overlap the VNet or peered networks. Unset
applies Azure's default (10.0.0.0/16).

### spec.networkProfile.serviceCidrs

`[]string`

Additional service CIDR for dual-stack (one IPv4 + one IPv6 block via
service_cidrs takes precedence over service_cidr in ARM; model your
dual-stack blocks here). Single-stack clusters use service_cidr.

### spec.networkProfile.podCidr

`string`

CIDR pods draw IPs from in OVERLAY (or kubenet) mode. Unset applies
Azure's default (10.244.0.0/16). Irrelevant in traditional CNI mode
(pods use subnet IPs).

### spec.networkProfile.podCidrs

`[]string`

Dual-stack pod CIDRs (one IPv4 + one IPv6). Single-stack clusters use
pod_cidr.

### spec.networkProfile.ipVersions

`[]enum`

IP families for the cluster. Unset means IPv4-only. For dual-stack,
list IPV4 then IPV6 and provide dual-stack pod_cidrs/service_cidrs.

Allowed values (use exactly as shown):

- `azure_aks_cluster_ip_version_unspecified` -- Not specified.
- `IPV4` -- IPv4.
- `IPV6` -- IPv6 (dual-stack only -- always alongside IPV4).

### spec.networkProfile.outboundType

`enum`

How cluster egress leaves Azure. Unspecified applies Azure's default
(LOAD_BALANCER): SNAT through the managed Standard load balancer.
MANAGED_NAT_GATEWAY swaps in an AKS-managed NAT gateway (better SNAT
scaling); USER_ASSIGNED_NAT_GATEWAY uses the NAT gateway already
attached to your BYO subnet (compose via AzureSubnet.nat_gateway_id);
USER_DEFINED_ROUTING sends everything to your route table's next hop
(the firewall-egress pattern -- requires a BYO subnet with an
AzureRouteTable attached); NONE provisions no egress path at all
(pair with bootstrap_profile CACHE so nodes can still pull system
images).

Allowed values (use exactly as shown):

- `azure_aks_cluster_outbound_type_unspecified` -- Not specified: Azure's default (loadBalancer).
- `LOAD_BALANCER` -- SNAT through the managed Standard load balancer.
- `MANAGED_NAT_GATEWAY` -- AKS-managed NAT gateway (managed VNet clusters).
- `USER_ASSIGNED_NAT_GATEWAY` -- The NAT gateway attached to your BYO subnet (compose it via AzureSubnet's nat_gateway_id seam).
- `USER_DEFINED_ROUTING` -- Your route table decides (firewall egress / forced tunneling); requires a BYO subnet with an attached AzureRouteTable.
- `OUTBOUND_NONE` -- No managed egress path at all -- network-isolated clusters; pair with bootstrap_profile CACHE.

### spec.networkProfile.loadBalancerProfile

`AzureAksClusterLoadBalancerProfile`

Tuning for the managed STANDARD LOAD BALANCER egress path (outbound
type LOAD_BALANCER): SNAT port allocation, outbound IP scaling or
explicit outbound IPs.

- rule: Choose one outbound IP strategy: managed counts, outbound_ip_prefix_ids, or outbound_ip_address_ids -- they are mutually exclusive

### spec.networkProfile.loadBalancerProfile.outboundPortsAllocated

`int32`

SNAT ports reserved per node (0-64000, multiples of 8). 0 (unset)
lets Azure allocate dynamically by cluster size. Set explicitly for
connection-heavy workloads hitting SNAT exhaustion.

- rule: outbound_ports_allocated must be between 0 and 64000

### spec.networkProfile.loadBalancerProfile.idleTimeoutInMinutes

`int32`

Minutes an idle outbound flow holds its SNAT port (4-100). Unset
applies Azure's default (30).

- rule: idle_timeout_in_minutes must be between 4 and 100

### spec.networkProfile.loadBalancerProfile.managedOutboundIpCount

`int32`

Number of Azure-managed outbound IPv4 addresses (1-100). More IPs =
more SNAT ports. Mutually exclusive with the explicit-IP fields.

- rule: managed_outbound_ip_count must be between 1 and 100

### spec.networkProfile.loadBalancerProfile.managedOutboundIpv6Count

`int32`

Number of Azure-managed outbound IPv6 addresses for dual-stack
clusters (1-100).

- rule: managed_outbound_ipv6_count must be between 1 and 100

### spec.networkProfile.loadBalancerProfile.outboundIpPrefixIds

`[]string | valueFrom`

Explicit public IP PREFIXES to SNAT from -- egress comes from known,
allowlistable CIDRs. Mutually exclusive with managed counts and
outbound_ip_address_ids.

- references: AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIpPrefix, name: <that resource's name>, fieldPath: status.outputs.public_ip_prefix_id}} -- a bare string does not parse

### spec.networkProfile.loadBalancerProfile.outboundIpAddressIds

`[]string | valueFrom`

Explicit public IP ADDRESSES to SNAT from. Mutually exclusive with
managed counts and outbound_ip_prefix_ids.

- references: AzurePublicIp (`status.outputs.public_ip_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIp, name: <that resource's name>, fieldPath: status.outputs.public_ip_id}} -- a bare string does not parse

### spec.networkProfile.loadBalancerProfile.backendPoolType

`enum`

How backend nodes register with the load balancer. Unspecified
applies Azure's default (NODE_IP_CONFIGURATION -- VMSS NIC configs).
NODE_IP registers node IPs directly.

Allowed values (use exactly as shown):

- `azure_aks_cluster_load_balancer_backend_pool_type_unspecified` -- Not specified: Azure's default (NodeIPConfiguration).
- `NODE_IP_CONFIGURATION` -- Nodes join by VMSS NIC IP configuration.
- `NODE_IP` -- Nodes join by IP address.

### spec.networkProfile.natGatewayProfile

`AzureAksClusterNatGatewayProfile`

Tuning for the MANAGED_NAT_GATEWAY egress path.

### spec.networkProfile.natGatewayProfile.idleTimeoutInMinutes

`int32`

Minutes an idle outbound flow holds its SNAT port (4-120). Unset
applies Azure's default (4).

- rule: idle_timeout_in_minutes must be between 4 and 120

### spec.networkProfile.natGatewayProfile.managedOutboundIpCount

`int32`

Number of managed outbound IPs on the NAT gateway (1-100).

- rule: managed_outbound_ip_count must be between 1 and 100

### spec.networkProfile.advancedNetworking

`AzureAksClusterAdvancedNetworking`

Advanced Container Networking Services: OBSERVABILITY streams
pod-level network flow metrics/logs; SECURITY adds FQDN-based egress
filtering. Both require the CILIUM data plane.

### spec.networkProfile.advancedNetworking.observabilityEnabled

`bool`

Pod-level network observability: flow metrics and logs surfaced
through Azure Monitor.

### spec.networkProfile.advancedNetworking.securityEnabled

`bool`

FQDN-based egress policy filtering.

### spec.autoScalerProfile

`AzureAksClusterAutoScalerProfile`

Fine-tuning for the cluster autoscaler that scales node pools with
auto_scaling_enabled. Unset fields keep Azure's defaults, which suit
most clusters; the profile is cluster-wide (all pools share it).

### spec.autoScalerProfile.balanceSimilarNodeGroups

`bool`

Treat node groups with identical instance types/labels as one when
balancing scale-out. Azure default: false.

### spec.autoScalerProfile.daemonsetEvictionForEmptyNodesEnabled

`bool`

Allow evicting DaemonSet pods from EMPTY nodes at scale-down. Azure
default: false.

### spec.autoScalerProfile.daemonsetEvictionForOccupiedNodesEnabled

`bool` · optional (explicit presence)

Allow evicting DaemonSet pods from OCCUPIED nodes at scale-down.
Azure default: true.

- default: `true`

### spec.autoScalerProfile.expander

`enum`

Which node group grows on scale-out. Unspecified keeps Azure's
default (RANDOM). LEAST_WASTE picks the best resource fit;
PRIORITY follows your priority config; MOST_PODS maximizes scheduled
pods.

Allowed values (use exactly as shown):

- `azure_aks_cluster_autoscaler_expander_unspecified` -- Not specified: Azure's default (random).
- `LEAST_WASTE` -- Pick the group wasting the least resources for the pending pods.
- `MOST_PODS` -- Pick the group that schedules the most pending pods.
- `PRIORITY` -- Follow user-configured group priorities.
- `RANDOM` -- Random choice.

### spec.autoScalerProfile.ignoreDaemonsetsUtilizationEnabled

`bool`

Ignore DaemonSet resource usage when computing node utilization.
Azure default: false.

### spec.autoScalerProfile.maxGracefulTerminationSec

`int32`

Seconds the autoscaler waits for graceful pod termination at
scale-down. Azure default: 600.

### spec.autoScalerProfile.maxNodeProvisioningTime

`string`

Longest a provisioning node may take before being abandoned, e.g.
"15m" (Azure's default).

### spec.autoScalerProfile.maxUnreadyNodes

`int32`

Max unready nodes tolerated before autoscaling pauses. Azure
default: 3.

### spec.autoScalerProfile.maxUnreadyPercentage

`int32`

Max unready percentage tolerated before autoscaling pauses (0-100).
Azure default: 45.

- rule: max_unready_percentage must be between 0 and 100

### spec.autoScalerProfile.newPodScaleUpDelay

`string`

Delay before unschedulable pods trigger scale-up, e.g. "10s" (Azure's
default "0s" scales immediately).

### spec.autoScalerProfile.scanInterval

`string`

How often the autoscaler re-evaluates, e.g. "10s" (Azure's default).

### spec.autoScalerProfile.scaleDownDelayAfterAdd

`string`

Cool-down after a scale-UP before scale-down evaluation resumes, e.g.
"10m" (Azure's default).

### spec.autoScalerProfile.scaleDownDelayAfterDelete

`string`

Cool-down after a node DELETION, e.g. "10s".

### spec.autoScalerProfile.scaleDownDelayAfterFailure

`string`

Cool-down after a scale-down FAILURE, e.g. "3m" (Azure's default).

### spec.autoScalerProfile.scaleDownUnneeded

`string`

How long a node must be unneeded before removal, e.g. "10m" (Azure's
default).

### spec.autoScalerProfile.scaleDownUnready

`string`

How long an UNREADY node must be unneeded before removal, e.g. "20m"
(Azure's default).

### spec.autoScalerProfile.scaleDownUtilizationThreshold

`string`

Utilization below which a node is scale-down eligible, as a fraction
string, e.g. "0.5" (Azure's default).

### spec.autoScalerProfile.emptyBulkDeleteMax

`int32`

Max empty nodes deleted in one pass. Azure default: 10.

### spec.autoScalerProfile.skipNodesWithLocalStorage

`bool`

Skip scale-down for nodes whose pods use local storage. Azure
default: false.

### spec.autoScalerProfile.skipNodesWithSystemPods

`bool` · optional (explicit presence)

Skip scale-down for nodes running non-DaemonSet kube-system pods.
Azure default: true.

- default: `true`

### spec.automaticUpgradeChannel

`enum`

Automatic upgrade channel for the cluster's KUBERNETES VERSION.
Unspecified means no automatic upgrades (Azure's default) -- pin
kubernetes_version and upgrade deliberately. PATCH auto-applies patch
releases of the pinned minor; STABLE tracks N-1 minor; RAPID tracks
the newest; NODE_IMAGE only refreshes node images (pair with
node_os_upgrade_channel NODE_IMAGE).

Allowed values (use exactly as shown):

- `azure_aks_cluster_upgrade_channel_unspecified` -- Not specified: no automatic Kubernetes upgrades (Azure's default).
- `PATCH` -- Auto-apply patch releases of the current minor.
- `STABLE` -- Track the latest supported minor (N-1 of newest).
- `RAPID` -- Track the newest supported minor.
- `NODE_IMAGE` -- Only refresh node images to the latest for the current version.

### spec.nodeOsUpgradeChannel

`enum`

Upgrade channel for the NODE OS IMAGE (security patches to the node
OS, independent of Kubernetes version). Unspecified applies Azure's
default (NODE_IMAGE): nodes are re-imaged as AKS publishes patched
images -- the recommended posture. SECURITY_PATCH applies patches
without re-imaging; UNMANAGED leaves patching to you; NONE disables.

Allowed values (use exactly as shown):

- `azure_aks_cluster_node_os_upgrade_channel_unspecified` -- Not specified: Azure's default (NodeImage).
- `NODE_OS_NODE_IMAGE` -- Re-image nodes as AKS publishes patched images (recommended).
- `SECURITY_PATCH` -- Apply OS security patches in place without re-imaging.
- `UNMANAGED` -- You manage node OS patching entirely.
- `NODE_OS_NONE` -- No automatic node OS updates.

### spec.maintenanceWindow

`AzureAksClusterMaintenanceWindow`

Dedicated maintenance window: WHEN Azure may perform routine cluster
maintenance (control-plane updates it initiates). Prefer the two
schedule-based windows below for upgrade operations; this legacy
window shapes everything else.

### spec.maintenanceWindow.allowed

`[]AzureAksClusterMaintenanceWindowAllowed`

Hours of specific weekdays when maintenance MAY run.

### spec.maintenanceWindow.allowed[].day

`enum`

The weekday this allowance covers.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_aks_cluster_week_day_unspecified` -- Not specified.
- `SUNDAY` -- Sunday.
- `MONDAY` -- Monday.
- `TUESDAY` -- Tuesday.
- `WEDNESDAY` -- Wednesday.
- `THURSDAY` -- Thursday.
- `FRIDAY` -- Friday.
- `SATURDAY` -- Saturday.

### spec.maintenanceWindow.allowed[].hours

`[]int32` · required

Permitted hours on that day, 0-23 (cluster-local UTC hours).

- rule: {"repeated":{"minItems":"1","items":{"int32":{"lte":23,"gte":0}}}}

### spec.maintenanceWindow.notAllowed

`[]AzureAksClusterMaintenanceWindowNotAllowed`

Absolute time spans when maintenance MUST NOT run (freeze windows).

### spec.maintenanceWindow.notAllowed[].start

`string` · required

Span start, RFC 3339, e.g. "2035-01-01T00:00:00Z".

- rule: {"required":true}

### spec.maintenanceWindow.notAllowed[].end

`string` · required

Span end, RFC 3339.

- rule: {"required":true}

### spec.maintenanceWindowAutoUpgrade

`AzureAksClusterMaintenanceWindowSchedule`

Schedule for AUTOMATIC KUBERNETES UPGRADES (the
automatic_upgrade_channel work). Configure it whenever an upgrade
channel is set so version bumps land in a window you chose.

### spec.maintenanceWindowAutoUpgrade.frequency

`enum` · required

How the schedule recurs. DAILY/WEEKLY use interval as
days/weeks-between; the monthly frequencies pick a day via
week_index+day_of_week (RELATIVE_MONTHLY) or day_of_month
(ABSOLUTE_MONTHLY).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_aks_cluster_maintenance_frequency_unspecified` -- Not specified.
- `DAILY` -- Every `interval` days.
- `WEEKLY` -- Every `interval` weeks on day_of_week.
- `RELATIVE_MONTHLY` -- Every `interval` months on the week_index-th day_of_week.
- `ABSOLUTE_MONTHLY` -- Every `interval` months on day_of_month.

### spec.maintenanceWindowAutoUpgrade.interval

`int32` · required

Recurrence interval: every N days/weeks/months (>= 1).

- rule: {"required":true,"int32":{"gte":1}}

### spec.maintenanceWindowAutoUpgrade.duration

`int32` · required

Window length in hours (4-24). Azure needs at least 4 hours to make
progress.

- rule: {"required":true,"int32":{"lte":24,"gte":4}}

### spec.maintenanceWindowAutoUpgrade.dayOfWeek

`enum`

Weekday for WEEKLY and RELATIVE_MONTHLY schedules.

Allowed values (use exactly as shown):

- `azure_aks_cluster_week_day_unspecified` -- Not specified.
- `SUNDAY` -- Sunday.
- `MONDAY` -- Monday.
- `TUESDAY` -- Tuesday.
- `WEDNESDAY` -- Wednesday.
- `THURSDAY` -- Thursday.
- `FRIDAY` -- Friday.
- `SATURDAY` -- Saturday.

### spec.maintenanceWindowAutoUpgrade.weekIndex

`enum`

Which occurrence of day_of_week for RELATIVE_MONTHLY (e.g. FIRST
Tuesday).

Allowed values (use exactly as shown):

- `azure_aks_cluster_week_index_unspecified` -- Not specified.
- `FIRST` -- First occurrence of the weekday.
- `SECOND` -- Second occurrence.
- `THIRD` -- Third occurrence.
- `FOURTH` -- Fourth occurrence.
- `LAST` -- Last occurrence.

### spec.maintenanceWindowAutoUpgrade.dayOfMonth

`int32`

Day of month (1-31) for ABSOLUTE_MONTHLY schedules.

- rule: day_of_month must be between 1 and 31

### spec.maintenanceWindowAutoUpgrade.startDate

`string`

Date the schedule takes effect, "yyyy-MM-dd". Unset starts
immediately.

### spec.maintenanceWindowAutoUpgrade.startTime

`string`

Window start time-of-day, "HH:mm" (with utc_offset applied), e.g.
"02:00".

- rule: start_time must be HH:mm, e.g. "02:00"

### spec.maintenanceWindowAutoUpgrade.utcOffset

`string`

UTC offset for start_time, "+HH:mm" or "-HH:mm", e.g. "+05:30".
Unset means UTC.

- rule: utc_offset must be +HH:mm or -HH:mm, e.g. "-08:00"

### spec.maintenanceWindowAutoUpgrade.notAllowed

`[]AzureAksClusterMaintenanceWindowNotAllowed`

Blackout spans that override the recurrence (freezes).

### spec.maintenanceWindowAutoUpgrade.notAllowed[].start

`string` · required

Span start, RFC 3339, e.g. "2035-01-01T00:00:00Z".

- rule: {"required":true}

### spec.maintenanceWindowAutoUpgrade.notAllowed[].end

`string` · required

Span end, RFC 3339.

- rule: {"required":true}

### spec.maintenanceWindowNodeOs

`AzureAksClusterMaintenanceWindowSchedule`

Schedule for NODE OS IMAGE maintenance (the node_os_upgrade_channel
work). Configure it alongside NODE_IMAGE or SECURITY_PATCH channels
so node re-imaging respects business hours.

### spec.maintenanceWindowNodeOs.frequency

`enum` · required

How the schedule recurs. DAILY/WEEKLY use interval as
days/weeks-between; the monthly frequencies pick a day via
week_index+day_of_week (RELATIVE_MONTHLY) or day_of_month
(ABSOLUTE_MONTHLY).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_aks_cluster_maintenance_frequency_unspecified` -- Not specified.
- `DAILY` -- Every `interval` days.
- `WEEKLY` -- Every `interval` weeks on day_of_week.
- `RELATIVE_MONTHLY` -- Every `interval` months on the week_index-th day_of_week.
- `ABSOLUTE_MONTHLY` -- Every `interval` months on day_of_month.

### spec.maintenanceWindowNodeOs.interval

`int32` · required

Recurrence interval: every N days/weeks/months (>= 1).

- rule: {"required":true,"int32":{"gte":1}}

### spec.maintenanceWindowNodeOs.duration

`int32` · required

Window length in hours (4-24). Azure needs at least 4 hours to make
progress.

- rule: {"required":true,"int32":{"lte":24,"gte":4}}

### spec.maintenanceWindowNodeOs.dayOfWeek

`enum`

Weekday for WEEKLY and RELATIVE_MONTHLY schedules.

Allowed values (use exactly as shown):

- `azure_aks_cluster_week_day_unspecified` -- Not specified.
- `SUNDAY` -- Sunday.
- `MONDAY` -- Monday.
- `TUESDAY` -- Tuesday.
- `WEDNESDAY` -- Wednesday.
- `THURSDAY` -- Thursday.
- `FRIDAY` -- Friday.
- `SATURDAY` -- Saturday.

### spec.maintenanceWindowNodeOs.weekIndex

`enum`

Which occurrence of day_of_week for RELATIVE_MONTHLY (e.g. FIRST
Tuesday).

Allowed values (use exactly as shown):

- `azure_aks_cluster_week_index_unspecified` -- Not specified.
- `FIRST` -- First occurrence of the weekday.
- `SECOND` -- Second occurrence.
- `THIRD` -- Third occurrence.
- `FOURTH` -- Fourth occurrence.
- `LAST` -- Last occurrence.

### spec.maintenanceWindowNodeOs.dayOfMonth

`int32`

Day of month (1-31) for ABSOLUTE_MONTHLY schedules.

- rule: day_of_month must be between 1 and 31

### spec.maintenanceWindowNodeOs.startDate

`string`

Date the schedule takes effect, "yyyy-MM-dd". Unset starts
immediately.

### spec.maintenanceWindowNodeOs.startTime

`string`

Window start time-of-day, "HH:mm" (with utc_offset applied), e.g.
"02:00".

- rule: start_time must be HH:mm, e.g. "02:00"

### spec.maintenanceWindowNodeOs.utcOffset

`string`

UTC offset for start_time, "+HH:mm" or "-HH:mm", e.g. "+05:30".
Unset means UTC.

- rule: utc_offset must be +HH:mm or -HH:mm, e.g. "-08:00"

### spec.maintenanceWindowNodeOs.notAllowed

`[]AzureAksClusterMaintenanceWindowNotAllowed`

Blackout spans that override the recurrence (freezes).

### spec.maintenanceWindowNodeOs.notAllowed[].start

`string` · required

Span start, RFC 3339, e.g. "2035-01-01T00:00:00Z".

- rule: {"required":true}

### spec.maintenanceWindowNodeOs.notAllowed[].end

`string` · required

Span end, RFC 3339.

- rule: {"required":true}

### spec.omsAgent

`AzureAksClusterOmsAgent`

Container Insights: stream container logs, metrics, and Kubernetes
events to a Log Analytics workspace -- the standard AKS observability
add-on.

### spec.omsAgent.logAnalyticsWorkspaceId

`string | valueFrom` · required

The Log Analytics workspace logs and metrics stream to.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.omsAgent.msiAuthForMonitoringEnabled

`bool`

Authenticate the agent with the cluster's managed identity instead
of a workspace key -- the modern, secret-less mode; enable it on new
clusters.

### spec.keyVaultSecretsProvider

`AzureAksClusterKeyVaultSecretsProvider`

Azure Key Vault provider for the Secrets Store CSI driver: mount Key
Vault secrets/keys/certificates into pods as volumes, with optional
periodic rotation.

### spec.keyVaultSecretsProvider.secretRotationEnabled

`bool`

Whether mounted secrets re-sync from Key Vault on an interval.
Without rotation, pods see updated secrets only on restart.

### spec.keyVaultSecretsProvider.secretRotationInterval

`string`

Poll interval for rotation, e.g. "2m" (Azure's default). Only
meaningful with secret_rotation_enabled.

### spec.keyVaultSecretsProvider.enabled

`bool` · optional (explicit presence)

Whether the add-on is on. Unset means on: declaring the block has always
meant enabling it, and this switch lets a manifest say the opposite out
loud while keeping the add-on's settings in place.

- default: `true`

### spec.azurePolicyEnabled

`bool`

Whether the Azure Policy add-on enforces policy-based governance
(pod security baselines, allowed registries, resource limits) inside
the cluster via OPA Gatekeeper. Recommended for governed estates.

### spec.microsoftDefender

`AzureAksClusterMicrosoftDefender`

Microsoft Defender for Containers: runtime threat detection streaming
security events to a Log Analytics workspace.

### spec.microsoftDefender.logAnalyticsWorkspaceId

`string | valueFrom` · required

The Log Analytics workspace security events stream to.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.monitorMetrics

`AzureAksClusterMonitorMetrics`

Azure Monitor managed Prometheus metrics collection. Set (even empty)
to enable the metrics profile; the two filter fields extend which
Kubernetes annotations/labels become metric labels.

### spec.monitorMetrics.annotationsAllowed

`string`

Comma-separated Kubernetes ANNOTATION keys exported as metric labels.

### spec.monitorMetrics.labelsAllowed

`string`

Comma-separated Kubernetes LABEL keys exported as metric labels.

### spec.monitorMetrics.enabled

`bool` · optional (explicit presence)

Whether the add-on is on. Unset means on: declaring the block has always
meant enabling it, and this switch lets a manifest say the opposite out
loud while keeping the add-on's settings in place.

- default: `true`

### spec.ingressApplicationGateway

`AzureAksClusterIngressApplicationGateway`

Application Gateway Ingress Controller (AGIC) add-on: program an
Azure Application Gateway as the cluster's ingress. Reference an
existing gateway by id, or have AKS create one from a subnet CIDR or
subnet id. Exactly one of the three anchors must be set.

- rule: Set exactly one of gateway_id, subnet_cidr, or subnet_id to anchor the Application Gateway

### spec.ingressApplicationGateway.gatewayId

`string | valueFrom`

Existing Application Gateway AGIC programs. The gateway must be
reachable from the cluster network (peered or same VNet).

- references: AzureApplicationGateway (`status.outputs.application_gateway_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationGateway, name: <that resource's name>, fieldPath: status.outputs.application_gateway_id}} -- a bare string does not parse

### spec.ingressApplicationGateway.gatewayName

`string`

Name for the AKS-created gateway (with subnet_cidr/subnet_id
anchors). Unset lets AKS derive one.

### spec.ingressApplicationGateway.subnetCidr

`string`

CIDR for a NEW subnet AKS creates in the cluster VNet to host a new
gateway, e.g. "10.225.0.0/24" (at least /27).

### spec.ingressApplicationGateway.subnetId

`string | valueFrom`

Existing subnet to host the new gateway -- must be dedicated to it
and at least /27.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.aciConnectorLinux

`AzureAksClusterAciConnectorLinux`

Virtual nodes (serverless burst) via Azure Container Instances: the
ACI connector schedules overflow pods into a dedicated subnet without
provisioning VMs.

### spec.aciConnectorLinux.subnetName

`string` · required

Name of the subnet (in the cluster's VNet) delegated to ACI
(delegation: Microsoft.ContainerInstance/containerGroups) where
virtual-node pods run.

- rule: {"required":true}

### spec.confidentialComputing

`AzureAksClusterConfidentialComputing`

Confidential computing (Intel SGX) support: deploys the SGX device
plugin, and optionally the quote-helper sidecar, for enclave
workloads on DC-series node pools.

### spec.confidentialComputing.sgxQuoteHelperEnabled

`bool`

Whether the SGX quote-helper sidecar runs for out-of-proc enclave
attestation.

### spec.confidentialComputing.enabled

`bool` · optional (explicit presence)

Whether the add-on is on. Unset means on: declaring the block has always
meant enabling it, and this switch lets a manifest say the opposite out
loud while keeping the add-on's settings in place.

- default: `true`

### spec.webAppRouting

`AzureAksClusterWebAppRouting`

Managed NGINX ingress (the "application routing" add-on): a
production-supported NGINX ingress controller wired to Azure DNS
zones for automatic record management.

### spec.webAppRouting.dnsZoneIds

`[]string | valueFrom`

Azure DNS zones (public and/or private) the add-on manages records
in as Ingress resources come and go. Empty enables the controller
without DNS automation.

- references: AzureDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.webAppRouting.defaultNginxController

`enum`

Posture of the default NGINX controller. Unspecified applies Azure's
default (ANNOTATION_CONTROLLED: per-Ingress annotations decide).
INTERNAL/EXTERNAL pin the default controller's load balancer
visibility; NGINX_NONE deploys no default controller.

Allowed values (use exactly as shown):

- `azure_aks_cluster_nginx_default_controller_unspecified` -- Not specified: Azure's default (AnnotationControlled).
- `ANNOTATION_CONTROLLED` -- Per-Ingress annotations choose internal/external.
- `INTERNAL` -- Default controller behind an internal (VNet-only) load balancer.
- `EXTERNAL` -- Default controller behind an external (public) load balancer.
- `NGINX_NONE` -- No default controller.

### spec.serviceMeshProfile

`AzureAksClusterServiceMeshProfile`

Managed Istio service mesh: Azure-operated Istio control plane with
optional managed ingress gateways and a bring-your-own certificate
authority from Key Vault.

### spec.serviceMeshProfile.mode

`enum` · required

Mesh mode -- ISTIO is the only supported value; the field exists for
future modes.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_aks_cluster_service_mesh_mode_unspecified` -- Not specified.
- `ISTIO` -- Azure-managed Istio.

### spec.serviceMeshProfile.revisions

`[]string` · required

Istio control-plane revisions in the cluster, e.g. ["asm-1-24"].
Two entries only during a canary control-plane upgrade.

- rule: {"repeated":{"minItems":"1","maxItems":"2"}}

### spec.serviceMeshProfile.internalIngressGatewayEnabled

`bool`

Whether the managed INTERNAL (VNet-only) Istio ingress gateway runs.

### spec.serviceMeshProfile.externalIngressGatewayEnabled

`bool`

Whether the managed EXTERNAL (public) Istio ingress gateway runs.

### spec.serviceMeshProfile.certificateAuthority

`AzureAksClusterServiceMeshCertificateAuthority`

Bring-your-own root certificate authority for mesh mTLS, sourced
from Key Vault. Omit for Istio's self-signed CA.

### spec.serviceMeshProfile.certificateAuthority.keyVaultId

`string | valueFrom` · required

The Key Vault holding the CA objects. The cluster identity needs
get/list on its certificates and secrets.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.serviceMeshProfile.certificateAuthority.rootCertObjectName

`string` · required

Key Vault object name of the root certificate.

- rule: {"required":true}

### spec.serviceMeshProfile.certificateAuthority.certChainObjectName

`string` · required

Key Vault object name of the certificate chain.

- rule: {"required":true}

### spec.serviceMeshProfile.certificateAuthority.certObjectName

`string` · required

Key Vault object name of the intermediate certificate.

- rule: {"required":true}

### spec.serviceMeshProfile.certificateAuthority.keyObjectName

`string` · required

Key Vault object name of the intermediate's private key.

- rule: {"required":true}

### spec.storageProfile

`AzureAksClusterStorageProfile`

Which CSI storage drivers run in the cluster. Unset fields keep
Azure's defaults (disk on, file on, snapshot controller on, blob
OFF). Enable the blob driver for workloads mounting Blob Storage via
NFS/FUSE.

### spec.storageProfile.blobDriverEnabled

`bool`

Azure Blob CSI driver (NFS/FUSE blob mounts). Azure default: false.

### spec.storageProfile.diskDriverEnabled

`bool` · optional (explicit presence)

Azure Disk CSI driver. Azure default: true -- most PVCs depend on it.

- default: `true`

### spec.storageProfile.fileDriverEnabled

`bool` · optional (explicit presence)

Azure Files CSI driver. Azure default: true.

- default: `true`

### spec.storageProfile.snapshotControllerEnabled

`bool` · optional (explicit presence)

CSI snapshot controller. Azure default: true.

- default: `true`

### spec.workloadAutoscalerProfile

`AzureAksClusterWorkloadAutoscalerProfile`

Managed workload autoscalers: KEDA (event-driven pod autoscaling) and
the Vertical Pod Autoscaler, both run by Azure instead of self-hosted.

### spec.workloadAutoscalerProfile.kedaEnabled

`bool`

Azure-managed KEDA (event-driven autoscaling from queue depths,
Kafka lag, custom metrics).

### spec.workloadAutoscalerProfile.verticalPodAutoscalerEnabled

`bool`

Azure-managed Vertical Pod Autoscaler.

### spec.keyManagementService

`AzureAksClusterKeyManagementService`

Key Management Service etcd encryption: envelope-encrypt Kubernetes
secrets at rest in etcd with YOUR Key Vault key (customer-managed
key), instead of Azure's platform keys.

### spec.keyManagementService.keyVaultKeyId

`string | valueFrom` · required

VERSIONED Key Vault key id used to envelope-encrypt Kubernetes
secrets in etcd, e.g.
https://myvault.vault.azure.net/keys/etcd-cmk/<version>.
AKS pins a specific key version: rotate by updating to the new
version's id. Defaults to referencing an AzureKeyVaultKey's key_id
output (the versioned id) in composed environments. The cluster
identity needs encrypt/decrypt on the key.

- references: AzureKeyVaultKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.keyManagementService.keyVaultNetworkAccess

`enum`

Whether the Key Vault is reached over public network or private
link. Unspecified applies Azure's default (PUBLIC).

Allowed values (use exactly as shown):

- `azure_aks_cluster_key_vault_network_access_unspecified` -- Not specified: Azure's default (Public).
- `KMS_PUBLIC` -- Key Vault reachable over its public endpoint.
- `KMS_PRIVATE` -- Key Vault reachable over private endpoints only.

### spec.httpProxyConfig

`AzureAksClusterHttpProxyConfig`

Corporate HTTP/HTTPS proxy configuration applied to nodes and pods --
for enterprises whose egress must traverse an inspection proxy.

### spec.httpProxyConfig.httpProxy

`string`

Proxy URL for HTTP traffic, e.g. "http://proxy.corp.example:3128".

### spec.httpProxyConfig.httpsProxy

`string`

Proxy URL for HTTPS traffic.

### spec.httpProxyConfig.noProxy

`[]string`

Hosts/domains/CIDRs that bypass the proxy. AKS automatically adds
the cluster-internal ranges.

### spec.httpProxyConfig.trustedCa

`string` · sensitive

Base64-encoded CA certificate of a TLS-intercepting proxy, trusted
by nodes. Marked sensitive: internal CA material identifies and
impersonates the interception layer.

### spec.linuxProfile

`AzureAksClusterLinuxProfile`

SSH access configuration for Linux nodes. Set it to enable direct SSH
(debugging via node IP); omit it and use `kubectl debug node` --
AKS generates and manages keys either way.

### spec.linuxProfile.adminUsername

`string` · required

Admin username for SSH, beginning with a letter; letters, numbers,
hyphens, and underscores.

- rule: admin_username begins with a letter and contains only letters, numbers, underscores, and hyphens
- rule: {"required":true}

### spec.linuxProfile.sshPublicKey

`string` · required

SSH PUBLIC key (e.g. "ssh-rsa AAAA...") installed for the admin
user. Public material -- the private half never leaves you.

- rule: {"required":true}

### spec.windowsProfile

`AzureAksClusterWindowsProfile`

Windows node administrator credentials and licensing -- required
before any Windows AzureAksNodePool can join this cluster.

### spec.windowsProfile.adminUsername

`string` · required

Windows administrator username.

- rule: {"required":true}

### spec.windowsProfile.adminPassword

`string` · required · sensitive

Windows administrator password, 8-123 characters with the usual
complexity requirements.

- rule: {"required":true,"string":{"minLen":"8","maxLen":"123"}}

### spec.windowsProfile.license

`enum`

License flavor. Set WINDOWS_SERVER to apply the Azure Hybrid Use
Benefit (bring your own Windows Server licenses).

Allowed values (use exactly as shown):

- `azure_aks_cluster_windows_license_unspecified` -- Not specified: pay-as-you-go Windows licensing.
- `WINDOWS_SERVER` -- Azure Hybrid Use Benefit -- bring your own Windows Server licenses.

### spec.windowsProfile.gmsa

`AzureAksClusterWindowsGmsa`

Group Managed Service Account (gMSA) support for Windows containers
authenticating to Active Directory.

- rule: Set dns_server and root_domain together, or leave both empty to inherit the VNet DNS configuration

### spec.windowsProfile.gmsa.dnsServer

`string`

DNS server for the Active Directory domain. Set both fields, or
leave both empty to inherit from the VNet DNS configuration.

### spec.windowsProfile.gmsa.rootDomain

`string`

Root domain name for gMSA, e.g. "corp.example.com". Set with
dns_server.

### spec.imageCleanerEnabled

`bool`

Whether Image Cleaner removes unused, vulnerable container images
from nodes on an interval. A cheap hardening win for long-lived
clusters.

### spec.imageCleanerIntervalHours

`int32`

Image Cleaner scan interval, in hours (24-2160). Only meaningful with
image_cleaner_enabled; unset applies Azure's default interval.

- rule: Image Cleaner interval must be between 24 and 2160 hours

### spec.costAnalysisEnabled

`bool`

Whether Microsoft Cost Management shows namespace- and
deployment-level cost breakdowns for this cluster. Requires STANDARD
or PREMIUM sku_tier.

### spec.runCommandEnabled

`bool` · optional (explicit presence)

Whether `az aks command invoke` (run-command) is allowed -- executing
commands inside the cluster through the Azure API without direct
network reach. Azure's default is enabled; hardened environments turn
it off to force all access through audited kubeconfig paths.

- default: `true`

### spec.diskEncryptionSetId

`string | valueFrom`

A Disk Encryption Set for encrypting node OS disks and persistent
volumes with a customer-managed key. A disk encryption set by ARM ID,
or a reference to an AzureDiskEncryptionSet's output. Changing it
replaces the cluster.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.edgeZone

`string`

Deploy the cluster into an Azure Extended Zone (edge location) close
to end users, e.g. "losangeles". Leave unset for the standard region
-- the overwhelmingly common case. Changing it replaces the cluster.

### spec.nodeResourceGroup

`string`

Name for the NODE resource group -- the Azure-managed group holding
the cluster's infrastructure (VMSS instances, managed load balancer,
managed public IPs). Must NOT already exist. Leave unset for Azure's
default ("MC_<rg>_<cluster>_<region>"). Changing it replaces the
cluster.

### spec.customCaTrustCertificatesBase64

`[]string`

Up to 10 base64-encoded CA certificates added to the trust store of
every node -- for private registries or proxies signed by an internal
CA. Requires custom CA trust (nodes re-image on change).

- rule: {"repeated":{"maxItems":"10"}}

### spec.bootstrapProfile

`AzureAksClusterBootstrapProfile`

Where nodes pull AKS bootstrap artifacts (system images) from.
Unspecified pulls directly from the Microsoft Container Registry --
fine whenever nodes have internet egress. Configure CACHE with an
Azure Container Registry (ACR cache rules) for network-isolated
clusters whose outbound_type is NONE or fully firewalled.

- rule: artifact_source CACHE requires container_registry_id (the ACR that caches Microsoft Container Registry content)

### spec.bootstrapProfile.artifactSource

`enum`

Where nodes pull AKS system images. Unspecified applies Azure's
default (DIRECT: Microsoft Container Registry over the internet).

Allowed values (use exactly as shown):

- `azure_aks_cluster_bootstrap_artifact_source_unspecified` -- Not specified: Azure's default (Direct from Microsoft Container Registry).
- `DIRECT` -- Pull directly from Microsoft Container Registry.
- `CACHE` -- Pull through your caching Azure Container Registry (network-isolated clusters).

### spec.bootstrapProfile.containerRegistryId

`string | valueFrom`

For CACHE: the Azure Container Registry (with ACR cache rules for
MCR) nodes pull through.

- references: AzureContainerRegistry (`status.outputs.container_registry_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerRegistry, name: <that resource's name>, fieldPath: status.outputs.container_registry_id}} -- a bare string does not parse

### spec.nodeProvisioningProfile

`AzureAksClusterNodeProvisioningProfile`

Node auto-provisioning (Karpenter): let AKS create and remove
right-sized node pools automatically from pending pod requirements,
instead of you pre-planning pools. AUTO mode hands node management to
the platform; leave unset for classic manual pool management.

### spec.nodeProvisioningProfile.mode

`enum`

Provisioning mode. Unspecified applies Azure's default (MANUAL): you
size and manage node pools. AUTO lets AKS create right-sized pools
from pending pod requirements.

Allowed values (use exactly as shown):

- `azure_aks_cluster_node_provisioning_mode_unspecified` -- Not specified: Azure's default (Manual).
- `MANUAL` -- You manage node pools.
- `AUTO` -- AKS (Karpenter) provisions and removes pools automatically.

### spec.nodeProvisioningProfile.defaultNodePools

`enum`

In AUTO mode: whether AKS seeds default system node pools.
Unspecified applies Azure's default (NODE_POOLS_AUTO).

Allowed values (use exactly as shown):

- `azure_aks_cluster_node_provisioning_default_pools_unspecified` -- Not specified: Azure's default (Auto).
- `NODE_POOLS_AUTO` -- AKS seeds managed default pools.
- `NODE_POOLS_NONE` -- No seeded pools -- everything auto-provisions from workload demand.

### spec.upgradeOverride

`AzureAksClusterUpgradeOverride`

Escape hatch for cluster upgrades: force an upgrade to proceed even
through drain failures or API-deprecation checks, until an expiry
timestamp. Use only to unblock a stuck upgrade, never as steady state.

### spec.upgradeOverride.forceUpgradeEnabled

`bool` · required

Whether upgrades ignore drain failures and deprecated-API checks.

- rule: {"required":true}

### spec.upgradeOverride.effectiveUntil

`string`

RFC 3339 timestamp the override stops applying, e.g.
"2035-01-01T00:00:00Z". Always set one -- an unbounded force
override is a standing foot-gun.

### spec.aiToolchainOperatorEnabled

`bool`

Whether the AI toolchain operator (KAITO) manages large-model
inference workloads on the cluster (GPU provisioning, model serving).

### spec.tags

`map<string, string>`

Free-form tags applied to the cluster, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Tags are Azure's governance
surface -- Azure Policy enforces them and Microsoft Cost Management
groups by them. Updatable in place.

## Validation Rules

- `aks_dns_prefix_exclusive`: Set at most one of dns_prefix or dns_prefix_private_cluster -- a cluster has either a public API FQDN prefix or a private-zone prefix, never both
- `aks_private_dns_prefix_requires_private_cluster`: dns_prefix_private_cluster is only valid when private_cluster_enabled is true
- `aks_workload_identity_requires_oidc`: workload_identity_enabled requires the OIDC issuer -- leave oidc_issuer_enabled unset (defaults to true) or set it to true
- `aks_local_account_disable_requires_aad`: Disabling local accounts requires azure_active_directory_role_based_access_control to be configured, or nobody can authenticate to the cluster
- `aks_cost_analysis_requires_paid_tier`: cost_analysis_enabled requires sku_tier STANDARD or PREMIUM
- `aks_lts_requires_premium`: The AKS_LONG_TERM_SUPPORT support plan requires sku_tier PREMIUM
- `aks_image_cleaner_interval_requires_enabled`: image_cleaner_interval_hours is only meaningful when image_cleaner_enabled is true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureAksCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | The Azure Resource Manager ID of the managed cluster. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{name} This is the parent reference AzureAksNodePool resources consume. |
| `status.outputs.cluster_name` | `string` | The name of the managed cluster. |
| `status.outputs.fqdn` | `string` | The public FQDN of the Kubernetes API server (empty for private clusters without a public FQDN). |
| `status.outputs.private_fqdn` | `string` | The private FQDN of the API server, populated for private clusters. |
| `status.outputs.portal_fqdn` | `string` | The FQDN used by the Azure Portal to reach the cluster when API Server VNet Integration is configured. |
| `status.outputs.oidc_issuer_url` | `string` | The cluster's OIDC issuer URL, populated when oidc_issuer_enabled is true (the default). An AzureFederatedIdentityCredential's `issuer` field consumes this value directly -- the trust anchor for workload identity federation. |
| `status.outputs.node_resource_group` | `string` | The name of the Azure-managed NODE resource group holding the cluster's infrastructure (VM scale sets, managed load balancer, managed public IPs). |
| `status.outputs.node_resource_group_id` | `string` | The Azure Resource Manager ID of the node resource group -- handy as a scope for role assignments over the cluster's infrastructure. |
| `status.outputs.cluster_kubeconfig` | `string` | Base64-encoded kubeconfig for the cluster (the user credential; with Entra ID integration it triggers the AAD login flow). Treat as a secret. |
| `status.outputs.cluster_identity_principal_id` | `string` | The principal (object) ID of the cluster's managed identity -- grant this identity Azure roles (e.g. Network Contributor on a BYO subnet, Private DNS Zone Contributor on a BYO private zone) via AzureRoleAssignment. |
| `status.outputs.kubelet_identity_object_id` | `string` | The object ID of the kubelet identity -- grant it AcrPull on container registries so nodes can pull images. |
| `status.outputs.kubelet_identity_client_id` | `string` | The client ID of the kubelet identity -- what pods see as the node's identity when using legacy IMDS-based access. |
| `status.outputs.current_kubernetes_version` | `string` | The Kubernetes version the control plane is actually running -- useful when kubernetes_version was left unset (AKS picked the latest recommended GA version). |
| `status.outputs.cluster_ca_certificate` | `string` | Base64-encoded cluster Certificate Authority (CA) certificate -- the standard kubeconfig certificate-authority-data format. Public cluster identity (TLS trust anchor), not credential material: it is exported as a plain value even though the providers derive it from their sensitive kubeconfig attribute. Consumed (with fqdn) by the platform's cluster-connection materializer, mirroring the EKS/GKE contract. |
| `status.outputs.entra_integration_enabled` | `string` | "true" when the cluster is Entra ID (Azure AD) integrated -- i.e. the spec configures azure_active_directory_role_based_access_control -- else "false". Entra integration is what makes the API server honor short-lived Entra bearer tokens, so this is the applicability gate the platform's cluster-connection materializer reads: token-based connections are only published for Entra-integrated clusters (local-accounts clusters connect through their kubeconfig instead). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.defaultNodePool.vnetSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.defaultNodePool.podSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.defaultNodePool.nodePublicIpPrefixId` | AzurePublicIpPrefix | `status.outputs.public_ip_prefix_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.kubeletIdentity.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.privateDnsZoneId` | AzurePrivateDnsZone | `status.outputs.zone_id` |
| `spec.apiServerAccessProfile.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.networkProfile.loadBalancerProfile.outboundIpPrefixIds` | AzurePublicIpPrefix | `status.outputs.public_ip_prefix_id` |
| `spec.networkProfile.loadBalancerProfile.outboundIpAddressIds` | AzurePublicIp | `status.outputs.public_ip_id` |
| `spec.omsAgent.logAnalyticsWorkspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.microsoftDefender.logAnalyticsWorkspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.ingressApplicationGateway.gatewayId` | AzureApplicationGateway | `status.outputs.application_gateway_id` |
| `spec.ingressApplicationGateway.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.webAppRouting.dnsZoneIds` | AzureDnsZone | `status.outputs.zone_id` |
| `spec.serviceMeshProfile.certificateAuthority.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.keyManagementService.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.key_id` |
| `spec.diskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.bootstrapProfile.containerRegistryId` | AzureContainerRegistry | `status.outputs.container_registry_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAksNodePool | `spec.kubernetesClusterId` | `status.outputs.cluster_id` |
| AzureDataProtectionBackupInstance | `spec.kubernetesCluster.kubernetesClusterId` | `status.outputs.cluster_id` |
| AzureFederatedIdentityCredential | `spec.issuer` | `status.outputs.oidc_issuer_url` |

## See Also

- [Overview](../README.md)

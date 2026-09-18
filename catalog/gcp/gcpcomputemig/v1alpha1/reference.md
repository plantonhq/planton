# GcpComputeMig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpComputeMigSpec defines a Google Compute Engine Managed Instance
Group (MIG) — a self-healing, optionally auto-scaling fleet of
identical VMs behind one declarative surface. One resource manages the
whole stack: the instance TEMPLATE (what each VM looks like), the
GROUP MANAGER (how many run, how they roll out changes, how failed
VMs are repaired), an optional AUTOSCALER (when the fleet grows and
shrinks), PER-INSTANCE CONFIGS (stateful name/disk/IP overrides for
individual instances), and RESIZE REQUESTS (queued one-shot capacity
asks).

Scope: exactly one of zone (a ZONAL group — all VMs in one zone) or
region (a REGIONAL group — VMs spread across the region's zones for
higher availability) must be set. The scope is immutable: moving a
group between zones/regions replaces every resource in the kind.

Template immutability: instance templates cannot be modified in place
(only their labels can change). Every change to the template block
ROTATES the template — a new template is created, the group manager
is pointed at it, and the old one is deleted after the switch. How
running VMs pick up the new template is governed by update_policy:
PROACTIVE rolls the fleet automatically within the surge/unavailable
budget; OPPORTUNISTIC waits for manual or lifecycle-driven refreshes.

The group's instance_group stack output is the load-balancer backend
handle: a GcpBackendService backend's group takes exactly that value.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeMig
metadata:
  name: my-sample-mig
spec:
  # Exactly one of zone/region. Zonal is the simplest shape; region
  # spreads the fleet across zones for availability.
  zone: us-central1-a

  # The VM shape — IMMUTABLE in GCP: every change here rotates the
  # template and rolls the group per updatePolicy.
  template:
    machineType: e2-micro
    disks:
      - boot: true
        sourceImage: debian-cloud/debian-12
    networkInterfaces:
      # Private fleet on the E2E VPC fixture (auto-mode, so the bare
      # network reference is a sufficient attachment point). No
      # accessConfigs = no external IPs — the secure default.
      - network:
          valueFrom:
            kind: GcpVpcNetwork
            name: planton-oss-e2e-gcpvpcnetwork-prereq
            fieldPath: status.outputs.network_self_link

  # Demand-driven sizing: the autoscaler owns the fleet size (mutually
  # exclusive with targetSize).
  autoscaler:
    minReplicas: 1
    maxReplicas: 3
    cpuTarget: 0.6

  # Template changes roll automatically, one extra instance of surge.
  updatePolicy:
    minimalAction: REPLACE
    type: PROACTIVE
    maxSurgeFixed: 1

  # What a destroy does: DELETE (the whole stack), PREVENT (destroy
  # fails), or ABANDON (leave the fleet running unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.migName` | `string` |  |  |  |
| `spec.zone` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.baseInstanceName` | `string` |  |  |  |
| `spec.template` | `GcpComputeMigTemplate` | yes |  |  |
| `spec.template.machineType` | `string` | yes |  |  |
| `spec.template.description` | `string` |  |  |  |
| `spec.template.instanceDescription` | `string` |  |  |  |
| `spec.template.disks` | `[]GcpComputeMigTemplateDisk` | yes |  |  |
| `spec.template.disks[].boot` | `bool` |  |  |  |
| `spec.template.disks[].sourceImage` | `string` |  |  |  |
| `spec.template.disks[].sourceSnapshot` | `string` |  |  |  |
| `spec.template.disks[].source` | `string \| valueFrom` |  |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.template.disks[].sizeGb` | `int32` |  |  |  |
| `spec.template.disks[].diskType` | `string` |  |  |  |
| `spec.template.disks[].type` | `string` |  |  |  |
| `spec.template.disks[].autoDelete` | `bool` |  | `true` |  |
| `spec.template.disks[].deviceName` | `string` |  |  |  |
| `spec.template.disks[].diskName` | `string` |  |  |  |
| `spec.template.disks[].mode` | `string` |  |  |  |
| `spec.template.disks[].interface` | `string` |  |  |  |
| `spec.template.disks[].diskLabels` | `map<string, string>` |  |  |  |
| `spec.template.disks[].provisionedIops` | `int64` |  |  |  |
| `spec.template.disks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.template.disks[].architecture` | `string` |  |  |  |
| `spec.template.disks[].guestOsFeatures` | `[]string` |  |  |  |
| `spec.template.disks[].resourcePolicies` | `[]string` |  |  |  |
| `spec.template.disks[].resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.template.disks[].storagePool` | `string` |  |  |  |
| `spec.template.disks[].diskEncryption` | `GcpComputeMigEncryptionKey` |  |  |  |
| `spec.template.disks[].diskEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.template.disks[].diskEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.template.disks[].sourceImageEncryption` | `GcpComputeMigEncryptionKey` |  |  |  |
| `spec.template.disks[].sourceImageEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.template.disks[].sourceImageEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.template.disks[].sourceSnapshotEncryption` | `GcpComputeMigEncryptionKey` |  |  |  |
| `spec.template.disks[].sourceSnapshotEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.template.disks[].sourceSnapshotEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.template.networkInterfaces` | `[]GcpComputeMigTemplateNetworkInterface` | yes |  |  |
| `spec.template.networkInterfaces[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.template.networkInterfaces[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.template.networkInterfaces[].subnetworkProject` | `string` |  |  |  |
| `spec.template.networkInterfaces[].networkIp` | `string` |  |  |  |
| `spec.template.networkInterfaces[].accessConfigs` | `[]GcpComputeMigAccessConfig` |  |  |  |
| `spec.template.networkInterfaces[].accessConfigs[].natIp` | `string` |  |  |  |
| `spec.template.networkInterfaces[].accessConfigs[].networkTier` | `string` |  |  |  |
| `spec.template.networkInterfaces[].ipv6AccessConfigs` | `[]GcpComputeMigIpv6AccessConfig` |  |  |  |
| `spec.template.networkInterfaces[].ipv6AccessConfigs[].networkTier` | `string` | yes |  |  |
| `spec.template.networkInterfaces[].stackType` | `string` |  |  |  |
| `spec.template.networkInterfaces[].nicType` | `string` |  |  |  |
| `spec.template.networkInterfaces[].queueCount` | `int32` |  |  |  |
| `spec.template.networkInterfaces[].aliasIpRanges` | `[]GcpComputeMigAliasIpRange` |  |  |  |
| `spec.template.networkInterfaces[].aliasIpRanges[].ipCidrRange` | `string` | yes |  |  |
| `spec.template.networkInterfaces[].aliasIpRanges[].subnetworkRangeName` | `string` |  |  |  |
| `spec.template.networkInterfaces[].networkAttachment` | `string` |  |  |  |
| `spec.template.networkInterfaces[].vlan` | `int32` |  |  |  |
| `spec.template.networkInterfaces[].igmpQuery` | `string` |  |  |  |
| `spec.template.networkInterfaces[].ipv6Address` | `string` |  |  |  |
| `spec.template.networkInterfaces[].internalIpv6PrefixLength` | `int32` |  |  |  |
| `spec.template.serviceAccount` | `GcpComputeMigServiceAccount` |  |  |  |
| `spec.template.serviceAccount.email` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.template.serviceAccount.scopes` | `[]string` | yes |  |  |
| `spec.template.scheduling` | `GcpComputeMigScheduling` |  |  |  |
| `spec.template.scheduling.provisioningModel` | `string` |  |  |  |
| `spec.template.scheduling.automaticRestart` | `bool` |  | `true` |  |
| `spec.template.scheduling.onHostMaintenance` | `string` |  |  |  |
| `spec.template.scheduling.instanceTerminationAction` | `string` |  |  |  |
| `spec.template.scheduling.maxRunDurationSeconds` | `int64` |  |  |  |
| `spec.template.scheduling.terminationTime` | `string` |  |  |  |
| `spec.template.scheduling.discardLocalSsdsOnStop` | `bool` |  |  |  |
| `spec.template.scheduling.availabilityDomain` | `int32` |  |  |  |
| `spec.template.scheduling.minNodeCpus` | `int32` |  |  |  |
| `spec.template.scheduling.nodeAffinities` | `[]GcpComputeMigNodeAffinity` |  |  |  |
| `spec.template.scheduling.nodeAffinities[].key` | `string` | yes |  |  |
| `spec.template.scheduling.nodeAffinities[].operator` | `string` | yes |  |  |
| `spec.template.scheduling.nodeAffinities[].values` | `[]string` | yes |  |  |
| `spec.template.scheduling.localSsdRecoveryTimeoutSeconds` | `int64` |  |  |  |
| `spec.template.scheduling.hostErrorTimeoutSeconds` | `int32` |  |  |  |
| `spec.template.shieldedInstanceConfig` | `GcpComputeMigShieldedConfig` |  |  |  |
| `spec.template.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.template.shieldedInstanceConfig.enableVtpm` | `bool` |  | `true` |  |
| `spec.template.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  | `true` |  |
| `spec.template.confidentialInstanceConfig` | `GcpComputeMigConfidentialConfig` |  |  |  |
| `spec.template.confidentialInstanceConfig.confidentialInstanceType` | `string` | yes |  |  |
| `spec.template.advancedMachineFeatures` | `GcpComputeMigAdvancedMachineFeatures` |  |  |  |
| `spec.template.advancedMachineFeatures.enableNestedVirtualization` | `bool` |  |  |  |
| `spec.template.advancedMachineFeatures.threadsPerCore` | `int32` |  |  |  |
| `spec.template.advancedMachineFeatures.visibleCoreCount` | `int32` |  |  |  |
| `spec.template.advancedMachineFeatures.enableUefiNetworking` | `bool` |  |  |  |
| `spec.template.advancedMachineFeatures.performanceMonitoringUnit` | `string` |  |  |  |
| `spec.template.advancedMachineFeatures.turboMode` | `string` |  |  |  |
| `spec.template.guestAccelerators` | `[]GcpComputeMigGuestAccelerator` |  |  |  |
| `spec.template.guestAccelerators[].type` | `string` | yes |  |  |
| `spec.template.guestAccelerators[].count` | `int32` | yes |  |  |
| `spec.template.reservationAffinity` | `GcpComputeMigReservationAffinity` |  |  |  |
| `spec.template.reservationAffinity.type` | `string` | yes |  |  |
| `spec.template.reservationAffinity.specificReservation` | `GcpComputeMigSpecificReservation` |  |  |  |
| `spec.template.reservationAffinity.specificReservation.key` | `string` | yes |  |  |
| `spec.template.reservationAffinity.specificReservation.values` | `[]string` | yes |  |  |
| `spec.template.totalEgressBandwidthTier` | `string` |  |  |  |
| `spec.template.metadata` | `map<string, string>` |  |  |  |
| `spec.template.startupScript` | `string` |  |  |  |
| `spec.template.tags` | `[]string` |  |  |  |
| `spec.template.labels` | `map<string, string>` |  |  |  |
| `spec.template.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.template.minCpuPlatform` | `string` |  |  |  |
| `spec.template.canIpForward` | `bool` |  |  |  |
| `spec.template.keyRevocationActionType` | `string` |  |  |  |
| `spec.template.resourcePolicies` | `[]string` |  |  |  |
| `spec.template.workloadIdentityConfig` | `GcpComputeMigWorkloadIdentityConfig` |  |  |  |
| `spec.template.workloadIdentityConfig.identity` | `string` | yes |  |  |
| `spec.template.workloadIdentityConfig.identityCertificateEnabled` | `bool` |  |  |  |
| `spec.versions` | `[]GcpComputeMigVersion` |  |  |  |
| `spec.versions[].versionName` | `string` |  |  |  |
| `spec.versions[].templateSelfLink` | `string` |  |  |  |
| `spec.versions[].targetSizeFixed` | `int32` |  |  |  |
| `spec.versions[].targetSizePercent` | `int32` |  |  |  |
| `spec.targetSize` | `int32` |  |  |  |
| `spec.namedPorts` | `[]GcpComputeMigNamedPort` |  |  |  |
| `spec.namedPorts[].name` | `string` | yes |  |  |
| `spec.namedPorts[].port` | `int32` | yes |  |  |
| `spec.updatePolicy` | `GcpComputeMigUpdatePolicy` |  |  |  |
| `spec.updatePolicy.minimalAction` | `string` | yes |  |  |
| `spec.updatePolicy.type` | `string` | yes |  |  |
| `spec.updatePolicy.mostDisruptiveAllowedAction` | `string` |  |  |  |
| `spec.updatePolicy.replacementMethod` | `string` |  |  |  |
| `spec.updatePolicy.maxSurgeFixed` | `int32` |  |  |  |
| `spec.updatePolicy.maxSurgePercent` | `int32` |  |  |  |
| `spec.updatePolicy.maxUnavailableFixed` | `int32` |  |  |  |
| `spec.updatePolicy.maxUnavailablePercent` | `int32` |  |  |  |
| `spec.updatePolicy.instanceRedistributionType` | `string` |  |  |  |
| `spec.autoHealing` | `GcpComputeMigAutoHealing` |  |  |  |
| `spec.autoHealing.healthCheck` | `string \| valueFrom` | yes |  | GcpHealthCheck (`status.outputs.self_link`) |
| `spec.autoHealing.initialDelaySec` | `int32` | yes |  |  |
| `spec.standbyPolicy` | `GcpComputeMigStandbyPolicy` |  |  |  |
| `spec.standbyPolicy.initialDelaySec` | `int32` |  |  |  |
| `spec.standbyPolicy.mode` | `string` |  |  |  |
| `spec.targetSuspendedSize` | `int32` |  |  |  |
| `spec.targetStoppedSize` | `int32` |  |  |  |
| `spec.statefulDisks` | `[]GcpComputeMigStatefulDisk` |  |  |  |
| `spec.statefulDisks[].deviceName` | `string` | yes |  |  |
| `spec.statefulDisks[].deleteRule` | `string` |  |  |  |
| `spec.statefulExternalIps` | `[]GcpComputeMigStatefulIp` |  |  |  |
| `spec.statefulExternalIps[].interfaceName` | `string` |  |  |  |
| `spec.statefulExternalIps[].deleteRule` | `string` |  |  |  |
| `spec.statefulInternalIps` | `[]GcpComputeMigStatefulIp` |  |  |  |
| `spec.statefulInternalIps[].interfaceName` | `string` |  |  |  |
| `spec.statefulInternalIps[].deleteRule` | `string` |  |  |  |
| `spec.instanceLifecyclePolicy` | `GcpComputeMigInstanceLifecyclePolicy` |  |  |  |
| `spec.instanceLifecyclePolicy.defaultActionOnFailure` | `string` |  |  |  |
| `spec.instanceLifecyclePolicy.forceUpdateOnRepair` | `string` |  |  |  |
| `spec.instanceLifecyclePolicy.onFailedHealthCheck` | `string` |  |  |  |
| `spec.instanceLifecyclePolicy.onRepairAllowChangingZone` | `string` |  |  |  |
| `spec.allInstancesConfig` | `GcpComputeMigAllInstancesConfig` |  |  |  |
| `spec.allInstancesConfig.labels` | `map<string, string>` |  |  |  |
| `spec.allInstancesConfig.metadata` | `map<string, string>` |  |  |  |
| `spec.listManagedInstancesResults` | `string` |  |  |  |
| `spec.workloadPolicy` | `string` |  |  |  |
| `spec.targetPools` | `[]string` |  |  |  |
| `spec.waitForInstances` | `bool` |  |  |  |
| `spec.waitForInstancesStatus` | `string` |  |  |  |
| `spec.distributionPolicy` | `GcpComputeMigDistributionPolicy` |  |  |  |
| `spec.distributionPolicy.zones` | `[]string` |  |  |  |
| `spec.distributionPolicy.targetShape` | `string` |  |  |  |
| `spec.instanceFlexibilityPolicy` | `GcpComputeMigInstanceFlexibilityPolicy` |  |  |  |
| `spec.instanceFlexibilityPolicy.instanceSelections` | `[]GcpComputeMigInstanceSelection` | yes |  |  |
| `spec.instanceFlexibilityPolicy.instanceSelections[].name` | `string` | yes |  |  |
| `spec.instanceFlexibilityPolicy.instanceSelections[].machineTypes` | `[]string` | yes |  |  |
| `spec.instanceFlexibilityPolicy.instanceSelections[].rank` | `int32` |  |  |  |
| `spec.targetSizePolicyMode` | `string` |  |  |  |
| `spec.autoscaler` | `GcpComputeMigAutoscaler` |  |  |  |
| `spec.autoscaler.autoscalerName` | `string` |  |  |  |
| `spec.autoscaler.description` | `string` |  |  |  |
| `spec.autoscaler.minReplicas` | `int32` |  |  |  |
| `spec.autoscaler.maxReplicas` | `int32` | yes |  |  |
| `spec.autoscaler.cooldownPeriod` | `int32` |  |  |  |
| `spec.autoscaler.mode` | `string` |  |  |  |
| `spec.autoscaler.cpuTarget` | `double` |  |  |  |
| `spec.autoscaler.cpuPredictiveMethod` | `string` |  |  |  |
| `spec.autoscaler.loadBalancingTarget` | `double` |  |  |  |
| `spec.autoscaler.metrics` | `[]GcpComputeMigAutoscalerMetric` |  |  |  |
| `spec.autoscaler.metrics[].name` | `string` | yes |  |  |
| `spec.autoscaler.metrics[].target` | `double` |  |  |  |
| `spec.autoscaler.metrics[].type` | `string` |  |  |  |
| `spec.autoscaler.metrics[].filter` | `string` |  |  |  |
| `spec.autoscaler.metrics[].singleInstanceAssignment` | `double` |  |  |  |
| `spec.autoscaler.scaleInControl` | `GcpComputeMigScaleInControl` |  |  |  |
| `spec.autoscaler.scaleInControl.maxScaledInReplicasFixed` | `int32` |  |  |  |
| `spec.autoscaler.scaleInControl.maxScaledInReplicasPercent` | `int32` |  |  |  |
| `spec.autoscaler.scaleInControl.timeWindowSec` | `int32` |  |  |  |
| `spec.autoscaler.schedules` | `[]GcpComputeMigScalingSchedule` |  |  |  |
| `spec.autoscaler.schedules[].scheduleName` | `string` | yes |  |  |
| `spec.autoscaler.schedules[].schedule` | `string` | yes |  |  |
| `spec.autoscaler.schedules[].durationSec` | `int32` | yes |  |  |
| `spec.autoscaler.schedules[].minRequiredReplicas` | `int32` | yes |  |  |
| `spec.autoscaler.schedules[].disabled` | `bool` |  |  |  |
| `spec.autoscaler.schedules[].timeZone` | `string` |  |  |  |
| `spec.autoscaler.schedules[].description` | `string` |  |  |  |
| `spec.autoscaler.stabilizationPeriod` | `int32` |  |  |  |
| `spec.perInstanceConfigs` | `[]GcpComputeMigPerInstanceConfig` |  |  |  |
| `spec.perInstanceConfigs[].configName` | `string` | yes |  |  |
| `spec.perInstanceConfigs[].preservedState` | `GcpComputeMigPreservedState` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.metadata` | `map<string, string>` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.disks` | `[]GcpComputeMigPreservedDisk` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.disks[].deviceName` | `string` | yes |  |  |
| `spec.perInstanceConfigs[].preservedState.disks[].source` | `string \| valueFrom` | yes |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.perInstanceConfigs[].preservedState.disks[].mode` | `string` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.disks[].deleteRule` | `string` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.externalIps` | `[]GcpComputeMigPreservedIp` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.externalIps[].interfaceName` | `string` | yes |  |  |
| `spec.perInstanceConfigs[].preservedState.externalIps[].address` | `string \| valueFrom` |  |  | GcpAddress (`status.outputs.address`) |
| `spec.perInstanceConfigs[].preservedState.externalIps[].autoDelete` | `string` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.internalIps` | `[]GcpComputeMigPreservedIp` |  |  |  |
| `spec.perInstanceConfigs[].preservedState.internalIps[].interfaceName` | `string` | yes |  |  |
| `spec.perInstanceConfigs[].preservedState.internalIps[].address` | `string \| valueFrom` |  |  | GcpAddress (`status.outputs.address`) |
| `spec.perInstanceConfigs[].preservedState.internalIps[].autoDelete` | `string` |  |  |  |
| `spec.perInstanceConfigs[].minimalAction` | `string` |  |  |  |
| `spec.perInstanceConfigs[].mostDisruptiveAllowedAction` | `string` |  |  |  |
| `spec.perInstanceConfigs[].removeInstanceOnDestroy` | `bool` |  |  |  |
| `spec.perInstanceConfigs[].removeInstanceStateOnDestroy` | `bool` |  |  |  |
| `spec.resizeRequests` | `[]GcpComputeMigResizeRequest` |  |  |  |
| `spec.resizeRequests[].requestName` | `string` | yes |  |  |
| `spec.resizeRequests[].description` | `string` |  |  |  |
| `spec.resizeRequests[].resizeBy` | `int32` | yes |  |  |
| `spec.resizeRequests[].requestedRunDurationSeconds` | `int64` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns every resource in the group.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.migName

`string`

Name of the managed instance group. 1-63 characters, lowercase
letters, numbers, and hyphens; must start with a letter and cannot
end with a hyphen. When omitted, metadata.name is used. The name
also seeds the managed resources around the group (template name
prefix, autoscaler name). Immutable after creation.

- rule: mig_name must be 1-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.zone

`string`

Zone for a ZONAL group (e.g. "us-central1-a") — every VM runs in
this one zone. Exactly one of zone or region must be set.
Immutable: a group cannot move between scopes or locations.

- rule: zone must look like us-central1-a

### spec.region

`string`

Region for a REGIONAL group (e.g. "us-central1") — VMs are spread
across the region's zones (see distribution_policy) so a zone
outage takes down only part of the fleet. Exactly one of zone or
region must be set. Immutable: a group cannot move between scopes
or locations.

- rule: region must look like us-central1

### spec.description

`string`

Human-readable description of the group shown in the console.
Immutable after creation.

### spec.baseInstanceName

`string`

Base name for VMs created by the group — instances are named
"<base_instance_name>-<random suffix>" (e.g. "web-x7kq"). RFC-1035
format, 1-58 characters. When omitted, the group name (mig_name or
metadata.name) is used.

- rule: base_instance_name must be lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.template

`GcpComputeMigTemplate` · required

The instance template — what every VM in the group looks like:
machine type, disks, networking, identity, and scheduling. The
template is IMMUTABLE in GCP (only labels can change in place):
every other change here creates a NEW template and rolls the group
to it per update_policy. See GcpComputeMigTemplate for the
replace-on-change semantics of each field.

- rule: {"required":true}
- rule: SPOT and FLEX_START VMs cannot automatically restart after reclamation — leave scheduling.automatic_restart unset or set it to false
- rule: scheduling.instance_termination_action applies to SPOT/FLEX_START VMs and to timed-run VMs (max_run_duration_seconds or termination_time) — configure one of those or remove the termination action
- rule: confidential VMs cannot live-migrate — set scheduling.on_host_maintenance to TERMINATE when confidential_instance_config is configured
- rule: VMs with guest accelerators (GPUs) cannot live-migrate — set scheduling.on_host_maintenance to TERMINATE when guest_accelerators are attached
- rule: RESERVATION_BOUND VMs consume one named reservation — set reservation_affinity.type to SPECIFIC_RESERVATION and name the reservation
- rule: max_run_duration_seconds and termination_time both bound the VM's lifetime — set at most one
- rule: exactly one disk must have boot set to true — the disk the VMs boot from

### spec.template.machineType

`string` · required

Machine type for every VM, e.g. "e2-micro", "e2-medium",
"n2-standard-4", or a custom shape like "custom-6-20480".
Changing it rotates the template.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.description

`string`

Human-readable description of the TEMPLATE resource itself.
Changing it rotates the template.

### spec.template.instanceDescription

`string`

Description stamped onto each INSTANCE created from the template
(visible on the VMs, distinct from the template's own
description). Changing it rotates the template.

### spec.template.disks

`[]GcpComputeMigTemplateDisk` · required

Disks attached to every VM created from the template. Exactly one
disk must be the boot disk. Each entry either creates a new disk
per instance (from an image, a snapshot, or blank) or attaches one
existing disk (source) — see GcpComputeMigTemplateDisk.
Changing disks rotates the template.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: a disk takes at most one source: source_image (fresh install), source_snapshot (restore), or source (attach an existing disk); leave all unset for a blank disk
- rule: the boot disk needs an OS to boot from — set source_image (e.g. debian-cloud/debian-12), source_snapshot, or source
- rule: source_image_encryption is only valid together with source_image
- rule: source_snapshot_encryption is only valid together with source_snapshot

### spec.template.disks[].boot

`bool`

Whether this is the boot disk — the disk the VMs boot from.
Exactly one disk in the template must set this.

### spec.template.disks[].sourceImage

`string`

Source image for a fresh disk on each VM. Accepts an image family
("debian-cloud/debian-12", "ubuntu-os-cloud/ubuntu-2404-lts-amd64")
or a specific image self link. Families resolve to the newest image
at template creation.

### spec.template.disks[].sourceSnapshot

`string`

Source snapshot each VM's disk is restored from (name or self
link).

### spec.template.disks[].source

`string | valueFrom`

Existing disk to attach (all VMs share it), referenced as a
GcpComputeDisk or a literal disk name/self link. Shared attachment
requires mode READ_ONLY.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.template.disks[].sizeGb

`int32`

Size of the disk in GB. When omitted, the image or snapshot size is
used (blank disks require a size).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65536,"gte":1}}

### spec.template.disks[].diskType

`string`

Disk type: "pd-standard" (HDD), "pd-balanced" (default in GCP and
the sensible choice), "pd-ssd" (high IOPS), local "local-ssd"
(ephemeral scratch), or a hyperdisk type on supported machine
families ("hyperdisk-balanced").

### spec.template.disks[].type

`string`

Disk role: "PERSISTENT" (default) or "SCRATCH" (ephemeral local
SSD; contents lost when the VM stops — pair with disk_type
"local-ssd" and 375 GB units).

- rule: type must be PERSISTENT or SCRATCH

### spec.template.disks[].autoDelete

`bool` · optional (explicit presence)

Delete each VM's disk automatically when the VM is deleted.
Defaults to true (matching GCP). Stateful disks (see the spec's
stateful_disks) override this per device.

- default: `true`

### spec.template.disks[].deviceName

`string`

Device name exposed under /dev/disk/by-id/google-*. When omitted
GCP assigns one. Stateful disk rules match on this name.

### spec.template.disks[].diskName

`string`

Name for disks created per VM. When omitted GCP derives one from
the instance name. Rarely needed.

### spec.template.disks[].mode

`string`

Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (required
when attaching one existing source disk across the fleet).

- rule: mode must be READ_WRITE or READ_ONLY

### spec.template.disks[].interface

`string`

Disk attachment interface: "SCSI" or "NVME". GCP normally selects
the right interface from the machine type and disk type — the
provider's own guidance is to leave it unset without advice from
Google.

- rule: interface must be SCSI or NVME

### spec.template.disks[].diskLabels

`map<string, string>`

Labels applied to the per-VM disks (distinct from VM labels) —
disk-level cost attribution and snapshot policies.

### spec.template.disks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned IOPS for hyperdisk types that support tuning
(e.g. hyperdisk-extreme, hyperdisk-balanced). Leave unset for pd-*
types.

- rule: {"int64":{"gt":"0"}}

### spec.template.disks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for hyperdisk types that support
tuning (e.g. hyperdisk-throughput, hyperdisk-balanced). Leave unset
for pd-* types.

- rule: {"int64":{"gt":"0"}}

### spec.template.disks[].architecture

`string`

CPU architecture of the disk/image: "X86_64" or "ARM64" (e.g. for
Tau T2A/Axion machine types). Normally inferred from the image.

- rule: architecture must be X86_64 or ARM64

### spec.template.disks[].guestOsFeatures

`[]string`

Guest OS features to enable on the disk, e.g. ["UEFI_COMPATIBLE",
"SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET", "WINDOWS"]. The accepted
set evolves with GCP — see "Enabling guest operating system
features" in the Compute Engine docs.

### spec.template.disks[].resourcePolicies

`[]string`

Self links of resource policies attached to each per-VM disk (e.g.
a snapshot schedule). GCP currently allows at most one per disk.

- rule: {"repeated":{"maxItems":"1"}}

### spec.template.disks[].resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the per-VM disks. Keys in the form
"tagKeys/{id}", values "tagValues/{id}".

### spec.template.disks[].storagePool

`string`

URL of the storage pool to create per-VM disks in (hyperdisk
storage pools).

### spec.template.disks[].diskEncryption

`GcpComputeMigEncryptionKey`

Customer-managed encryption key (CMEK) for the per-VM disks,
referenced as a GcpKmsKey. The Compute Engine service agent
(service-<project-number>@compute-system.iam.gserviceaccount.com)
must hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.

### spec.template.disks[].diskEncryption.kmsKey

`string | valueFrom` · required

The KMS key, referenced as a GcpKmsKey or a literal self link. The
service agent performing the operation needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.template.disks[].diskEncryption.kmsKeyServiceAccount

`string`

Service account used for the encryption/decryption request. When
omitted, the Compute Engine default service agent is used.

### spec.template.disks[].sourceImageEncryption

`GcpComputeMigEncryptionKey`

Decrypts the source image when it is itself CMEK-encrypted. Only
valid together with source_image.

### spec.template.disks[].sourceImageEncryption.kmsKey

`string | valueFrom` · required

The KMS key, referenced as a GcpKmsKey or a literal self link. The
service agent performing the operation needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.template.disks[].sourceImageEncryption.kmsKeyServiceAccount

`string`

Service account used for the encryption/decryption request. When
omitted, the Compute Engine default service agent is used.

### spec.template.disks[].sourceSnapshotEncryption

`GcpComputeMigEncryptionKey`

Decrypts the source snapshot when it is itself CMEK-encrypted.
Only valid together with source_snapshot.

### spec.template.disks[].sourceSnapshotEncryption.kmsKey

`string | valueFrom` · required

The KMS key, referenced as a GcpKmsKey or a literal self link. The
service agent performing the operation needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.template.disks[].sourceSnapshotEncryption.kmsKeyServiceAccount

`string`

Service account used for the encryption/decryption request. When
omitted, the Compute Engine default service agent is used.

### spec.template.networkInterfaces

`[]GcpComputeMigTemplateNetworkInterface` · required

Network interfaces for every VM. At least one is required; multiple
NICs must each attach to a different VPC network. Changing
interfaces rotates the template.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: each network interface needs an attachment point: a network (auto-mode VPC), a subnetwork (custom-mode VPC), or a network_attachment (Private Service Connect) — set at least one

### spec.template.networkInterfaces[].network

`string | valueFrom`

VPC network for this interface, referenced as a GcpVpcNetwork.
Sufficient alone only for auto-mode VPCs; custom-mode VPCs need
subnetwork.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.template.networkInterfaces[].subnetwork

`string | valueFrom`

Subnetwork for this interface, referenced as a GcpSubnetwork. The
subnetwork's region must contain the group's location.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.template.networkInterfaces[].subnetworkProject

`string`

Project owning the subnetwork — set when attaching to a Shared VPC
host project's subnetwork from a service project.

### spec.template.networkInterfaces[].networkIp

`string`

Static internal IP shared configuration is not meaningful for a
fleet (every VM needs its own address) — this field pins the
PRIMARY internal IP only for single-instance groups or specialized
setups. When omitted (the norm), GCP assigns each VM an ephemeral
internal IP from the subnetwork range.

### spec.template.networkInterfaces[].accessConfigs

`[]GcpComputeMigAccessConfig`

External IPv4 access configs. Empty means no external IP (private
fleet — pair with Cloud NAT for egress; the secure default). GCP
supports at most one access config per interface.

- rule: {"repeated":{"maxItems":"1"}}

### spec.template.networkInterfaces[].accessConfigs[].natIp

`string`

Static external IP to pin — meaningful only for single-instance
groups (a fleet cannot share one IP; use stateful_external_ips for
per-instance identity). When omitted (the norm), each VM gets an
ephemeral external IP.

### spec.template.networkInterfaces[].accessConfigs[].networkTier

`string`

Network service tier for the external IPs: "PREMIUM" (default;
Google's global backbone) or "STANDARD" (regional, cheaper).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PREMIUM","STANDARD"]}}

### spec.template.networkInterfaces[].ipv6AccessConfigs

`[]GcpComputeMigIpv6AccessConfig`

External IPv6 access configs. Requires stack_type "IPV4_IPV6" and a
subnetwork with an external IPv6 range. At most one per interface.

- rule: {"repeated":{"maxItems":"1"}}

### spec.template.networkInterfaces[].ipv6AccessConfigs[].networkTier

`string` · required

Network service tier for IPv6 traffic. Only "PREMIUM" is valid.

- rule: {"required":true,"string":{"in":["PREMIUM"]}}

### spec.template.networkInterfaces[].stackType

`string`

IP stack of the interface:
  ""            -- same as "IPV4_ONLY" (GCP default)
  "IPV4_ONLY"   -- IPv4 addresses only
  "IPV4_IPV6"   -- dual stack (subnetwork must have an IPv6 range)
  "IPV6_ONLY"   -- IPv6 only (supported on IPv6-enabled
                   subnetworks)

- rule: stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY

### spec.template.networkInterfaces[].nicType

`string`

vNIC type: "" (GCP picks), "GVNIC" (recommended on modern machine
families; required for TIER_1 bandwidth), "VIRTIO_NET" (legacy), or
the RDMA types "IDPF", "MRDMA", "IRDMA" on specialized shapes.

- rule: nic_type must be one of GVNIC, VIRTIO_NET, IDPF, MRDMA, IRDMA

### spec.template.networkInterfaces[].queueCount

`int32` · optional (explicit presence)

Networking queue count for Rx and Tx (1-32). When omitted GCP sizes
queues from vCPU count.

- rule: {"int32":{"lte":32,"gte":1}}

### spec.template.networkInterfaces[].aliasIpRanges

`[]GcpComputeMigAliasIpRange`

Alias IP ranges served by this interface on every VM — per-VM
secondary ranges for multi-IP workloads.

### spec.template.networkInterfaces[].aliasIpRanges[].ipCidrRange

`string` · required

The alias range: a CIDR ("10.1.2.0/24"), a single IP ("10.1.2.3"),
or a netmask ("/24") to auto-allocate per VM from the range.

- rule: {"required":true}

### spec.template.networkInterfaces[].aliasIpRanges[].subnetworkRangeName

`string`

Secondary range name on the subnetwork to allocate from. When
omitted the primary range is used.

### spec.template.networkInterfaces[].networkAttachment

`string`

URL of a Private Service Connect NETWORK ATTACHMENT this interface
connects to, in the form
"projects/{projectNumber}/regions/{region}/networkAttachments/{name}"
— connects the fleet into a producer's VPC. An attachment-only
interface is legal (no network or subnetwork).

### spec.template.networkInterfaces[].vlan

`int32` · optional (explicit presence)

VLAN tag (2-255) making this a DYNAMIC network interface — a
sub-interface multiplexed onto a parent NIC.

- rule: {"int32":{"lte":255,"gte":2}}

### spec.template.networkInterfaces[].igmpQuery

`string`

IGMP multicast query support on this interface:
  ""                    -- GCP default (disabled)
  "IGMP_QUERY_V2"       -- IGMPv2 queries enabled (multicast)
  "IGMP_QUERY_DISABLED" -- explicitly disabled

- rule: igmp_query must be IGMP_QUERY_V2 or IGMP_QUERY_DISABLED

### spec.template.networkInterfaces[].ipv6Address

`string`

Static INTERNAL IPv6 address for the interface — meaningful only
for single-instance groups (a fleet cannot share one address).
When omitted, GCP assigns per-VM addresses from the subnetwork's
internal IPv6 range.

### spec.template.networkInterfaces[].internalIpv6PrefixLength

`int32` · optional (explicit presence)

Prefix length of the primary internal IPv6 range assigned to this
interface. When omitted, GCP assigns its default.

- rule: {"int32":{"lte":128,"gte":1}}

### spec.template.serviceAccount

`GcpComputeMigServiceAccount`

Service account the VMs' workloads authenticate as. When omitted,
the Compute Engine default service account is used with its default
scopes — prefer a dedicated least-privilege account for production.
Changing it rotates the template.

### spec.template.serviceAccount.email

`string | valueFrom`

Service account email, referenced as a GcpServiceAccount or a
literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.template.serviceAccount.scopes

`[]string` · required

OAuth scopes for the attached account. The modern practice is a
single "https://www.googleapis.com/auth/cloud-platform" scope with
access controlled entirely by IAM roles; narrower legacy scopes
remain supported. Required when this block is set.

- rule: {"repeated":{"minItems":"1"}}

### spec.template.scheduling

`GcpComputeMigScheduling`

Scheduling policy: Spot vs standard provisioning, maintenance
behavior, run-duration limits, and sole-tenant node placement.
Changing it rotates the template.

### spec.template.scheduling.provisioningModel

`string`

Provisioning model:
  ""                  -- same as "STANDARD"
  "STANDARD"          -- on-demand capacity
  "SPOT"              -- deeply discounted preemptible capacity;
                         GCP may reclaim VMs at any time (the group
                         recreates them per its repair policy)
  "FLEX_START"        -- discounted capacity with a flexible start
                         time (Dynamic Workload Scheduler); pairs
                         with resize_requests
  "RESERVATION_BOUND" -- runs only on capacity from one specific
                         reservation (pair with reservation_affinity
                         type SPECIFIC_RESERVATION)

- rule: provisioning_model must be STANDARD, SPOT, FLEX_START, or RESERVATION_BOUND

### spec.template.scheduling.automaticRestart

`bool` · optional (explicit presence)

Restart a VM automatically when Compute Engine (not a user) stops
it. Defaults to true for standard VMs; must be false (or unset) for
Spot/FLEX_START.

- default: `true`

### spec.template.scheduling.onHostMaintenance

`string`

Host maintenance behavior: "" (GCP default "MIGRATE"), "MIGRATE"
(live-migrate; zero downtime), or "TERMINATE" (stop during
maintenance — required for GPUs and confidential VMs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MIGRATE","TERMINATE"]}}

### spec.template.scheduling.instanceTerminationAction

`string`

What GCP does when a VM is reclaimed (Spot preemption, FLEX_START
expiry) or a run-duration limit fires: "STOP" (keep the stopped VM
and disks) or "DELETE" (remove the VM — the usual choice in a
managed group, which recreates capacity itself).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STOP","DELETE"]}}

### spec.template.scheduling.maxRunDurationSeconds

`int64` · optional (explicit presence)

Maximum run duration in seconds, after which
instance_termination_action is executed. The duration clock starts
at every VM start. Mutually exclusive with termination_time.

- rule: {"int64":{"lte":"315576000000","gte":"1"}}

### spec.template.scheduling.terminationTime

`string`

Absolute timestamp (RFC 3339) at which VMs are terminated.
Mutually exclusive with max_run_duration_seconds.

### spec.template.scheduling.discardLocalSsdsOnStop

`bool` · optional (explicit presence)

Discard local-SSD contents when a VM is stopped by a lifetime limit
(max_run_duration/termination_time) instead of preserving them.

### spec.template.scheduling.availabilityDomain

`int32` · optional (explicit presence)

Availability domain for spread-placement within the zone (used with
spread placement policies).

- rule: {"int32":{"gte":1}}

### spec.template.scheduling.minNodeCpus

`int32` · optional (explicit presence)

Minimum vCPUs on the sole-tenant node the VMs can be scheduled
onto. Sole-tenancy only.

- rule: {"int32":{"gte":1}}

### spec.template.scheduling.nodeAffinities

`[]GcpComputeMigNodeAffinity`

Sole-tenant node affinities selecting which node groups the VMs may
run on. Setting any affinity places the fleet on sole-tenant
hardware.

### spec.template.scheduling.nodeAffinities[].key

`string` · required

Node-group label key to match (e.g.
"compute.googleapis.com/node-group-name").

- rule: {"required":true}

### spec.template.scheduling.nodeAffinities[].operator

`string` · required

Match operator: "IN" or "NOT_IN".

- rule: {"required":true,"string":{"in":["IN","NOT_IN"]}}

### spec.template.scheduling.nodeAffinities[].values

`[]string` · required

Label values to match.

- rule: {"repeated":{"minItems":"1"}}

### spec.template.scheduling.localSsdRecoveryTimeoutSeconds

`int64` · optional (explicit presence)

How long Compute Engine waits for a local-SSD-preserving recovery
when a host fails, in seconds, before falling back to default
recovery.

- rule: {"int64":{"lte":"604800","gte":"0"}}

### spec.template.scheduling.hostErrorTimeoutSeconds

`int32` · optional (explicit presence)

How long Compute Engine waits, in seconds, before declaring a host
failed and starting host-error recovery for the VM on it. A lower
value recovers a hung host faster at the cost of more false
positives; leave unset for Compute Engine's default recovery timing.
Must be 90..330 in steps of 30 (90, 120, ..., 330).

- rule: host_error_timeout_seconds must be a multiple of 30 between 90 and 330
- rule: {"int32":{"lte":330,"gte":90}}

### spec.template.shieldedInstanceConfig

`GcpComputeMigShieldedConfig`

Shielded VM configuration (secure boot, vTPM, integrity
monitoring). Requires an image with Shielded VM support (all recent
Google-provided images qualify). Changing it rotates the template.

### spec.template.shieldedInstanceConfig.enableSecureBoot

`bool` · optional (explicit presence)

Verify the boot loader's signature chain; blocks boot on tampering.
GCP default is false because some third-party images are unsigned.

### spec.template.shieldedInstanceConfig.enableVtpm

`bool` · optional (explicit presence)

Virtual Trusted Platform Module. GCP default is true.

- default: `true`

### spec.template.shieldedInstanceConfig.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Boot-integrity measurement and monitoring via the vTPM. GCP default
is true.

- default: `true`

### spec.template.confidentialInstanceConfig

`GcpComputeMigConfidentialConfig`

Confidential VM configuration — hardware memory encryption (AMD
SEV / SEV-SNP or Intel TDX). Requires a supported machine family
(e.g. N2D, C2D, C3) and scheduling.on_host_maintenance =
"TERMINATE". Changing it rotates the template.

### spec.template.confidentialInstanceConfig.confidentialInstanceType

`string` · required

Confidential computing technology:
  "SEV"     -- AMD Secure Encrypted Virtualization (N2D/C2D/C3D)
  "SEV_SNP" -- AMD SEV Secure Nested Paging (requires an AMD Milan+
               min_cpu_platform)
  "TDX"     -- Intel Trust Domain Extensions (C3)

- rule: {"required":true,"string":{"in":["SEV","SEV_SNP","TDX"]}}

### spec.template.advancedMachineFeatures

`GcpComputeMigAdvancedMachineFeatures`

Advanced machine features: nested virtualization, SMT control,
visible core count, UEFI networking, performance monitoring unit,
and turbo mode. Changing it rotates the template.

### spec.template.advancedMachineFeatures.enableNestedVirtualization

`bool` · optional (explicit presence)

Expose nested virtualization support (VMX) to the guest — run VMs
inside the VMs.

### spec.template.advancedMachineFeatures.threadsPerCore

`int32` · optional (explicit presence)

Threads per physical core: 1 disables simultaneous multithreading
(SMT) — common for licensing and security isolation; 2 is the
hardware default.

- rule: {"int32":{"in":[1,2]}}

### spec.template.advancedMachineFeatures.visibleCoreCount

`int32` · optional (explicit presence)

Number of physical cores exposed to the guest (core visibility for
per-core licensing). When unset all cores are visible.

- rule: {"int32":{"gte":1}}

### spec.template.advancedMachineFeatures.enableUefiNetworking

`bool` · optional (explicit presence)

Enable UEFI networking in the guest firmware.

### spec.template.advancedMachineFeatures.performanceMonitoringUnit

`string`

Performance monitoring unit exposure level: "STANDARD", "ENHANCED",
or "ARCHITECTURAL".

- rule: performance_monitoring_unit must be STANDARD, ENHANCED, or ARCHITECTURAL

### spec.template.advancedMachineFeatures.turboMode

`string`

Turbo frequency mode. "ALL_CORE_MAX" runs all cores at maximum
turbo frequency (supported machine families only).

- rule: turbo_mode must be ALL_CORE_MAX

### spec.template.guestAccelerators

`[]GcpComputeMigGuestAccelerator`

GPU accelerator cards attached to every VM. Requires a GPU-capable
location and scheduling.on_host_maintenance = "TERMINATE".
Changing it rotates the template.

### spec.template.guestAccelerators[].type

`string` · required

Accelerator type available in the group's location, e.g.
"nvidia-tesla-t4", "nvidia-l4", "nvidia-a100-80gb".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.guestAccelerators[].count

`int32` · required

Number of cards of this type per VM.

- rule: {"required":true,"int32":{"gte":1}}

### spec.template.reservationAffinity

`GcpComputeMigReservationAffinity`

Reservation affinity — whether VMs consume capacity from any
matching reservation, a specific reservation, or none. Changing it
rotates the template.

- rule: specific_reservation must be set when (and only when) type is SPECIFIC_RESERVATION

### spec.template.reservationAffinity.type

`string` · required

Reservation consumption mode:
  "ANY_RESERVATION"      -- consume any matching reservation (GCP
                            default)
  "SPECIFIC_RESERVATION" -- consume only the named reservation
  "NO_RESERVATION"       -- never consume reserved capacity

- rule: {"required":true,"string":{"in":["ANY_RESERVATION","SPECIFIC_RESERVATION","NO_RESERVATION"]}}

### spec.template.reservationAffinity.specificReservation

`GcpComputeMigSpecificReservation`

The specific reservation to consume (type SPECIFIC_RESERVATION).

### spec.template.reservationAffinity.specificReservation.key

`string` · required

Reservation label key — use
"compute.googleapis.com/reservation-name" to target by name.

- rule: {"required":true}

### spec.template.reservationAffinity.specificReservation.values

`[]string` · required

Reservation label values (the reservation name when using the
reservation-name key).

- rule: {"repeated":{"minItems":"1"}}

### spec.template.totalEgressBandwidthTier

`string`

Per-VM egress bandwidth tier. "TIER_1" raises the bandwidth cap on
supported machine shapes (N2/N2D/C2/C3 with >= 30 vCPUs and gVNIC);
"DEFAULT" is the standard cap. Changing it rotates the template.

- rule: total_egress_bandwidth_tier must be DEFAULT or TIER_1

### spec.template.metadata

`map<string, string>`

Custom metadata key/value pairs made available to the guest OS via
the metadata server. Well-known keys configure agents and features
(e.g. "enable-oslogin"). Changing metadata rotates the template.

### spec.template.startupScript

`string`

Startup script executed by the guest agent on every boot of every
VM. Maps to the metadata_startup_script surface, which keeps it
distinct from user metadata. Changing it rotates the template.

### spec.template.tags

`[]string`

Network tags used by firewall rules and network routes to select
the group's VMs. Changing tags rotates the template.

### spec.template.labels

`map<string, string>`

User labels stamped onto every VM (merged with Planton attribution
labels, which win on key conflicts). The ONLY template surface GCP
allows to change in place — label edits do not rotate the template.

### spec.template.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the TEMPLATE for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values
"tagValues/{id}". Changing them rotates the template.

### spec.template.minCpuPlatform

`string`

Minimum CPU platform for the VMs, e.g. "Intel Ice Lake" or
"AMD Milan". Constrains scheduling to hosts with at least this
platform. Changing it rotates the template.

### spec.template.canIpForward

`bool`

Allow sending/receiving packets with source or destination IPs
that do not match the VM's own addresses — required for VMs acting
as routers, NAT gateways, or VPN endpoints. Changing it rotates the
template.

### spec.template.keyRevocationActionType

`string`

Action GCP takes on the VMs when a Cloud KMS key protecting them is
revoked: "NONE" (default) or "STOP". Changing it rotates the
template.

- rule: key_revocation_action_type must be NONE or STOP

### spec.template.resourcePolicies

`[]string`

Self links of compute resource policies attached to every VM (e.g.
an instance schedule). GCP currently allows at most one policy per
instance. Changing it rotates the template.

- rule: {"repeated":{"maxItems":"1"}}

### spec.template.workloadIdentityConfig

`GcpComputeMigWorkloadIdentityConfig`

Managed workload identity for every VM in the group: a SPIFFE
identity issued to each instance (and, optionally, X.509 identity
certificates) so workloads authenticate to each other by identity
instead of shared secrets or network position. Part of the template,
so changing it rotates the template and rolls the group.

### spec.template.workloadIdentityConfig.identity

`string` · required

The SPIFFE ID Compute Engine issues to each instance, e.g.
"spiffe://PROJECT.svc.id.goog/ns/NAMESPACE/sa/SERVICE_ACCOUNT" or a
workload-identity-pool identity of the form
"spiffe://POOL.global.PROJECT_NUMBER.workload.id.goog/ns/NS/sa/SA".

- rule: {"required":true,"string":{"minLen":"1","prefix":"spiffe://"}}

### spec.template.workloadIdentityConfig.identityCertificateEnabled

`bool`

Whether Compute Engine also issues and rotates X.509 certificates
bound to the identity, made available on each VM for mutual TLS.

### spec.versions

`[]GcpComputeMigVersion`

Named application versions for CANARY rollouts. When empty (the
normal case), the group runs one version on this kind's own
template. Add entries to split the fleet across templates — e.g.
the kind's template as the stable version plus an external
template URL pinned as a canary with a small target_size.

- rule: a version's target size is either fixed or percent — set at most one

### spec.versions[].versionName

`string`

Name of the version (e.g. "stable", "canary") — appears in the
instance metadata of VMs created from it.

### spec.versions[].templateSelfLink

`string`

Instance template for this version. Leave EMPTY to run this
kind's own template (the default and the normal case). Set a full
template self link URL to pin an EXTERNAL template — the canary
escape hatch for splitting the fleet across templates this kind
does not manage.

### spec.versions[].targetSizeFixed

`int32` · optional (explicit presence)

Fixed number of instances running this version. Unset on the
stable version (it absorbs the remainder of the fleet).

- rule: {"int32":{"gte":0}}

### spec.versions[].targetSizePercent

`int32` · optional (explicit presence)

Percent of the fleet (0-100) running this version, rounded up.
Unset on the stable version (it absorbs the remainder).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.targetSize

`int32` · optional (explicit presence)

Fixed number of running VMs in the group. Set this for manually
sized groups; leave it unset when the autoscaler manages the size
(the two are mutually exclusive — the autoscaler would fight a
fixed size on every apply). When neither target_size nor autoscaler
is set, the group is created with 0 instances.

- rule: {"int32":{"gte":0}}

### spec.namedPorts

`[]GcpComputeMigNamedPort`

Named ports published by the group — the mechanism backend services
use to map a logical service name ("http") to a port number (8080)
on every VM. The same name must be used by the backend service's
port_name.

### spec.namedPorts[].name

`string` · required

The port name backend services reference via port_name (e.g.
"http").

- rule: {"required":true}

### spec.namedPorts[].port

`int32` · required

The port number the service listens on (e.g. 8080).

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.updatePolicy

`GcpComputeMigUpdatePolicy`

How the group rolls out template and configuration changes to
running VMs: automatically (PROACTIVE) or on-demand
(OPPORTUNISTIC), and within what surge/unavailability budget.
When omitted, the group manager applies changes opportunistically
with GCP's default budget.

- rule: the surge budget is either fixed or percent — set at most one of max_surge_fixed / max_surge_percent
- rule: the unavailability budget is either fixed or percent — set at most one of max_unavailable_fixed / max_unavailable_percent
- rule: RECREATE preserves instance names by replacing in place, so it needs room to take instances down — set max_unavailable_fixed or max_unavailable_percent above 0 (the provider's own requirement)

### spec.updatePolicy.minimalAction

`string` · required

The most disruptive action the rollout may take on its own:
  "NONE"    -- no automatic action (changes wait for manual
               refresh)
  "REFRESH" -- apply updates that need no restart
  "RESTART" -- stop/start instances to apply updates
  "REPLACE" -- recreate instances from the new template (the usual
               choice for template rotations)

- rule: {"required":true,"string":{"in":["NONE","REFRESH","RESTART","REPLACE"]}}

### spec.updatePolicy.type

`string` · required

Rollout mode:
  "PROACTIVE"     -- the group rolls changes out automatically
                     within the surge/unavailability budget
  "OPPORTUNISTIC" -- changes apply only when instances are
                     recreated anyway (manual refresh, autoscaler
                     churn, repairs)

- rule: {"required":true,"string":{"in":["OPPORTUNISTIC","PROACTIVE"]}}

### spec.updatePolicy.mostDisruptiveAllowedAction

`string`

Cap on how disruptive an individual update is allowed to be —
updates needing more disruption than this wait instead. Same value
set as minimal_action; defaults to "REPLACE" (no cap).

- rule: most_disruptive_allowed_action must be NONE, REFRESH, RESTART, or REPLACE

### spec.updatePolicy.replacementMethod

`string`

How replaced instances are recreated:
  ""           -- same as "SUBSTITUTE"
  "SUBSTITUTE" -- new instances get fresh random names (allows
                  surge; zero-unavailability rollouts possible)
  "RECREATE"   -- instance NAMES are preserved (stateful-friendly);
                  requires an unavailability budget above 0

- rule: replacement_method must be RECREATE or SUBSTITUTE

### spec.updatePolicy.maxSurgeFixed

`int32` · optional (explicit presence)

Extra instances the rollout may create above target_size (fixed
count). Higher surge = faster rollout, more temporary cost.
Regional groups accept ONLY 0 or a value >= the group's zone count
(live-verified 400: "Fixed updatePolicy.maxSurge for regional
managed instance group has to be either 0 or at least equal to the
number of zones" — a regional group spreads over 3 zones by
default, so 1 and 2 are rejected; use percent for finer budgets).

- rule: {"int32":{"gte":0}}

### spec.updatePolicy.maxSurgePercent

`int32` · optional (explicit presence)

Extra instances the rollout may create above target_size, as a
percent of the group (0-100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.updatePolicy.maxUnavailableFixed

`int32` · optional (explicit presence)

Instances the rollout may take below target_size (fixed count).

- rule: {"int32":{"gte":0}}

### spec.updatePolicy.maxUnavailablePercent

`int32` · optional (explicit presence)

Instances the rollout may take below target_size, as a percent of
the group (0-100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.updatePolicy.instanceRedistributionType

`string`

REGIONAL groups only: whether the group proactively redistributes
instances to keep the zone balance ("PROACTIVE", the GCP default)
or leaves them where they are ("NONE" — required for stateful
regional groups).

- rule: instance_redistribution_type must be PROACTIVE or NONE

### spec.autoHealing

`GcpComputeMigAutoHealing`

Auto-healing: recreate VMs that fail an application-level health
check (not just VM liveness). Requires a GcpHealthCheck; the
initial delay gives freshly booted VMs time to become healthy
before repairs kick in.

### spec.autoHealing.healthCheck

`string | valueFrom` · required

The health check that decides instance health, referenced as a
GcpHealthCheck or a literal self link. Use a health check tuned for
auto-healing (conservative thresholds) — aggressive LB health
checks cause repair storms.

- references: GcpHealthCheck (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHealthCheck, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.autoHealing.initialDelaySec

`int32` · required

Seconds a freshly created VM gets to boot and become healthy before
auto-healing counts failures against it (0-3600). Size it to your
application's cold-start time — too short and the group repair-loops
healthy-but-slow instances.

- rule: {"required":true,"int32":{"lte":3600,"gte":0}}

### spec.standbyPolicy

`GcpComputeMigStandbyPolicy`

Standby pool configuration: keep pre-created VMs SUSPENDED or
STOPPED so scale-outs resume them (seconds) instead of booting
from scratch (minutes).

### spec.standbyPolicy.initialDelaySec

`int32` · optional (explicit presence)

Seconds a newly created standby VM runs before being suspended or
stopped (0-3600) — time for boot and warmup so resume is instant.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.standbyPolicy.mode

`string`

How standby VMs are consumed:
  ""               -- same as "MANUAL" (GCP default)
  "MANUAL"         -- standby VMs activate only by explicit API
                      action
  "SCALE_OUT_POOL" -- scale-outs resume standby VMs first (the
                      fast-scale-out mode)

- rule: mode must be MANUAL or SCALE_OUT_POOL

### spec.targetSuspendedSize

`int32` · optional (explicit presence)

Target number of SUSPENDED VMs held in the standby pool
(suspended VMs keep memory state; fastest resume, disks and memory
keep billing).

- rule: {"int32":{"gte":0}}

### spec.targetStoppedSize

`int32` · optional (explicit presence)

Target number of STOPPED VMs held in the standby pool (stopped VMs
keep only disks; cheaper than suspended, slower to start).

- rule: {"int32":{"gte":0}}

### spec.statefulDisks

`[]GcpComputeMigStatefulDisk`

Stateful persistent disks preserved across instance recreation and
updates — for VMs whose disks carry irreplaceable data (databases,
brokers). The device names must match disks defined in the
template.

### spec.statefulDisks[].deviceName

`string` · required

Device name of the template disk to make stateful — must match a
device_name in the template's disks.

- rule: {"required":true}

### spec.statefulDisks[].deleteRule

`string`

What happens to the preserved disk when its instance is PERMANENTLY
deleted (group deletion, explicit instance delete — not repairs):
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the disk is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the disk is deleted with the
                                      instance

- rule: delete_rule must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.statefulExternalIps

`[]GcpComputeMigStatefulIp`

Stateful EXTERNAL IPs preserved across instance recreation — each
VM keeps its public IP identity through repairs and updates.

### spec.statefulExternalIps[].interfaceName

`string`

Network interface name whose IP to preserve. When omitted, "nic0"
(the first interface) is used.

### spec.statefulExternalIps[].deleteRule

`string`

What happens to the preserved IP when its instance is PERMANENTLY
deleted:
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the address is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
                                      the instance

- rule: delete_rule must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.statefulInternalIps

`[]GcpComputeMigStatefulIp`

Stateful INTERNAL IPs preserved across instance recreation — each
VM keeps its private IP identity through repairs and updates.

### spec.statefulInternalIps[].interfaceName

`string`

Network interface name whose IP to preserve. When omitted, "nic0"
(the first interface) is used.

### spec.statefulInternalIps[].deleteRule

`string`

What happens to the preserved IP when its instance is PERMANENTLY
deleted:
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the address is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
                                      the instance

- rule: delete_rule must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.instanceLifecyclePolicy

`GcpComputeMigInstanceLifecyclePolicy`

What the group does when instances fail or need repair — the
repair-vs-do-nothing switches and the health-check failure action.

### spec.instanceLifecyclePolicy.defaultActionOnFailure

`string`

What the group does with failed instances (crashed, preempted,
failed to start):
  ""           -- same as "REPAIR" (GCP default)
  "REPAIR"     -- recreate/restart failed instances automatically
  "DO_NOTHING" -- leave failed instances alone (manual operations
                  mode)

- rule: default_action_on_failure must be REPAIR or DO_NOTHING

### spec.instanceLifecyclePolicy.forceUpdateOnRepair

`string`

Whether a repair may apply the LATEST template version instead of
the instance's current one: "YES" or "NO" (GCP default NO —
repairs preserve the running version; YES turns repairs into
opportunistic update vectors).

- rule: force_update_on_repair must be YES or NO

### spec.instanceLifecyclePolicy.onFailedHealthCheck

`string`

What the group does when the auto-healing health check fails (only
meaningful with auto_healing configured):
  ""               -- same as "DEFAULT_ACTION" (GCP default)
  "DEFAULT_ACTION" -- follow default_action_on_failure
  "REPAIR"         -- repair on health-check failure even when
                      default_action_on_failure is DO_NOTHING
  "DO_NOTHING"     -- never repair on health-check failure

- rule: on_failed_health_check must be DEFAULT_ACTION, REPAIR, or DO_NOTHING

### spec.instanceLifecyclePolicy.onRepairAllowChangingZone

`string`

Whether a repair may recreate the instance in a DIFFERENT zone of a
regional group: "YES" or "NO" (GCP default NO — repairs stay
in-zone).

- rule: on_repair_allow_changing_zone must be YES or NO

### spec.allInstancesConfig

`GcpComputeMigAllInstancesConfig`

Labels and metadata stamped onto ALL instances IN ADDITION to what
the template defines — changing these does NOT rotate the template;
the group patches running instances per update_policy instead. Use
for fleet-wide toggles that should not force template rotation.

### spec.allInstancesConfig.labels

`map<string, string>`

Labels added to every instance (on top of template labels).

### spec.allInstancesConfig.metadata

`map<string, string>`

Metadata added to every instance (on top of template metadata).

### spec.listManagedInstancesResults

`string`

Pagination behavior of the group's listManagedInstances API:
  ""          -- same as "PAGELESS" (GCP default)
  "PAGELESS"  -- ignores pagination parameters (legacy behavior)
  "PAGINATED" -- respects maxResults/pageToken (use for very large
                 groups)

- rule: list_managed_instances_results must be PAGELESS or PAGINATED

### spec.workloadPolicy

`string`

Self link of a WORKLOAD resource policy applied to the group (e.g.
a high-throughput or placement workload policy). GCP accepts at
most one workload policy per group.

### spec.targetPools

`[]string`

Self links of legacy target pools (network load balancer) whose
member list the group manages. Modern L7/L4 load balancing uses
backend services pointed at the group's instance_group output
instead — target pools remain for the legacy NLB path.

### spec.waitForInstances

`bool` · optional (explicit presence)

Wait for all managed instances to reach the wait_for_instances_status
before the apply completes (and before dependent resources deploy).
Turns "the group exists" into "the fleet is actually up" — useful
when a chart deploys consumers right behind the group. Failed VMs
make the apply fail at timeout.

### spec.waitForInstancesStatus

`string`

Which state wait_for_instances waits for:
  ""         -- same as "STABLE" (GCP default)
  "STABLE"   -- instances are running or repaired to running
  "UPDATED"  -- additionally, all instances are on the target
                template version (rollouts fully converged)

- rule: wait_for_instances_status must be STABLE or UPDATED

### spec.distributionPolicy

`GcpComputeMigDistributionPolicy`

REGIONAL groups only: which zones VMs spread across and the target
distribution shape. When omitted, GCP spreads evenly across the
region's zones.

### spec.distributionPolicy.zones

`[]string`

The zones instances may run in (e.g. ["us-central1-a",
"us-central1-f"]). When omitted, GCP uses the region's zones.
Immutable: changing zones replaces the group.

### spec.distributionPolicy.targetShape

`string`

The distribution shape the group converges to (per the API's
documented set):
  ""                -- same as "EVEN" (GCP default)
  "EVEN"            -- equal instance counts across zones (highest
                       availability)
  "BALANCED"        -- spread across zones subject to capacity
  "ANY"             -- any zone with capacity (best for scarce
                       shapes)
  "ANY_SINGLE_ZONE" -- all instances in one zone GCP picks
The provider does not pre-validate this list — it is the API's
documented vocabulary, enforced here so typos fail at validate
time.

- rule: target_shape must be EVEN, BALANCED, ANY, or ANY_SINGLE_ZONE

### spec.instanceFlexibilityPolicy

`GcpComputeMigInstanceFlexibilityPolicy`

REGIONAL groups only: ranked machine-type alternatives the group
may fall back to when the primary shape is out of capacity —
useful for large fleets on capacity-constrained shapes.

### spec.instanceFlexibilityPolicy.instanceSelections

`[]GcpComputeMigInstanceSelection` · required

Named, ranked machine-type selections. Lower rank is preferred.

- rule: {"repeated":{"minItems":"1"}}

### spec.instanceFlexibilityPolicy.instanceSelections[].name

`string` · required

Name of this selection (e.g. "primary", "fallback").

- rule: {"required":true}

### spec.instanceFlexibilityPolicy.instanceSelections[].machineTypes

`[]string` · required

Machine types in this selection (e.g. ["e2-standard-4",
"n2-standard-4"]).

- rule: {"repeated":{"minItems":"1"}}

### spec.instanceFlexibilityPolicy.instanceSelections[].rank

`int32` · optional (explicit presence)

Preference rank — selections with lower rank are consumed first.

- rule: {"int32":{"gte":0}}

### spec.targetSizePolicyMode

`string`

How the group creates VMs to reach its target size:
  ""           -- GCP default (individual creation)
  "INDIVIDUAL" -- create VMs one by one; partial success possible
  "BULK"       -- all-or-nothing atomic creation of the whole
                  target size
Immutable: changing it replaces the group.

- rule: target_size_policy_mode must be BULK or INDIVIDUAL

### spec.autoscaler

`GcpComputeMigAutoscaler`

Autoscaler for the group — grows and shrinks the fleet between
min/max replicas from CPU, load-balancer serving capacity, custom
Cloud Monitoring metrics, or calendar schedules. Mutually
exclusive with target_size.

- rule: max_replicas must be greater than or equal to min_replicas

### spec.autoscaler.autoscalerName

`string`

Name of the autoscaler resource. When omitted, the group name is
used.

- rule: autoscaler_name must be lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.autoscaler.description

`string`

Human-readable description of the autoscaler.

### spec.autoscaler.minReplicas

`int32`

Minimum number of replicas the autoscaler maintains (>= 0).

- rule: {"int32":{"gte":0}}

### spec.autoscaler.maxReplicas

`int32` · required

Maximum number of replicas the autoscaler may create.

- rule: {"required":true,"int32":{"gte":1}}

### spec.autoscaler.cooldownPeriod

`int32` · optional (explicit presence)

Seconds the autoscaler waits before collecting usage from a NEW
instance — covers boot and warmup so unreliable early samples do
not drive scaling. GCP default: 60.

- rule: {"int32":{"gte":0}}

### spec.autoscaler.mode

`string`

Operating mode (the API's documented vocabulary; the provider
defaults it to "ON" and does not pre-validate the value):
  ""               -- same as "ON"
  "ON"             -- scale out and in
  "OFF"            -- autoscaler holds everything (group keeps its
                      current size)
  "ONLY_SCALE_OUT" -- grow but never shrink

- rule: mode must be ON, OFF, or ONLY_SCALE_OUT

### spec.autoscaler.cpuTarget

`double` · optional (explicit presence)

Target CPU utilization as a fraction in (0, 1] — e.g. 0.6 keeps
average fleet CPU at 60%. The default signal when no other target
is set (GCP applies 0.6 when the policy has no targets at all).

- rule: {"double":{"lte":1,"gt":0}}

### spec.autoscaler.cpuPredictiveMethod

`string`

Predictive autoscaling for the CPU signal (the provider's own
documented values):
  ""                      -- same as "NONE"
  "NONE"                  -- react to real-time metrics only
  "OPTIMIZE_AVAILABILITY" -- learn daily/weekly patterns and scale
                             out AHEAD of anticipated demand

- rule: cpu_predictive_method must be NONE or OPTIMIZE_AVAILABILITY

### spec.autoscaler.loadBalancingTarget

`double` · optional (explicit presence)

Target backend utilization fraction (0, 1] of the load balancer's
configured serving capacity — scales the fleet to hold the
balancer's per-backend utilization at this level. Only meaningful
when the group serves an EXTERNAL HTTP(S) load balancer with
UTILIZATION balancing mode.

- rule: {"double":{"lte":1,"gt":0}}

### spec.autoscaler.metrics

`[]GcpComputeMigAutoscalerMetric`

Custom Cloud Monitoring metric signals.

- rule: a metric is either utilization-targeted (target + type) or workload-proportional (single_instance_assignment) — set at most one

### spec.autoscaler.metrics[].name

`string` · required

The metric identifier, e.g.
"pubsub.googleapis.com/subscription/num_undelivered_messages" or
"custom.googleapis.com/myapp/queue_depth". The metric must export
values for the fleet's instances (or be a per-group metric used
with filter + single_instance_assignment).

- rule: {"required":true}

### spec.autoscaler.metrics[].target

`double` · optional (explicit presence)

Target value of the metric the autoscaler maintains per instance.
Pair with type to say how the metric is interpreted.

- rule: {"double":{"gt":0}}

### spec.autoscaler.metrics[].type

`string`

How the metric's values are interpreted against target (the
provider validates this list):
  "GAUGE"             -- instantaneous value
  "DELTA_PER_SECOND"  -- rate per second
  "DELTA_PER_MINUTE"  -- rate per minute

- rule: type must be GAUGE, DELTA_PER_SECOND, or DELTA_PER_MINUTE

### spec.autoscaler.metrics[].filter

`string`

Monitoring filter expression selecting the metric's time series
(e.g. a per-group Pub/Sub subscription metric). GCP default:
"resource.type = gce_instance".

### spec.autoscaler.metrics[].singleInstanceAssignment

`double` · optional (explicit presence)

For per-GROUP workload metrics (queue depth, pending jobs): the
amount of work one instance handles. The autoscaler keeps
instances proportional to metric_value / single_instance_assignment.
Mutually exclusive with target/type.

- rule: {"double":{"gt":0}}

### spec.autoscaler.scaleInControl

`GcpComputeMigScaleInControl`

Limits how fast the autoscaler SHRINKS the fleet after load drops —
the guard against cascading scale-in on temporary dips.

- rule: scale_in_control needs at least one bound: max_scaled_in_replicas_fixed, max_scaled_in_replicas_percent, or time_window_sec
- rule: the scale-in cap is either fixed or percent — set at most one

### spec.autoscaler.scaleInControl.maxScaledInReplicasFixed

`int32` · optional (explicit presence)

Max instances that may be removed within the trailing time window
(fixed count).

- rule: {"int32":{"gte":0}}

### spec.autoscaler.scaleInControl.maxScaledInReplicasPercent

`int32` · optional (explicit presence)

Max instances that may be removed within the trailing time window,
as a percent of the group (0-100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.autoscaler.scaleInControl.timeWindowSec

`int32` · optional (explicit presence)

The trailing time window (seconds) the scale-in cap applies over.

- rule: {"int32":{"gte":0}}

### spec.autoscaler.schedules

`[]GcpComputeMigScalingSchedule`

Calendar-based capacity schedules (cron) that set a minimum replica
count for recurring time windows — e.g. business-hours capacity
floors. Metric-based scaling still adds capacity above the
schedule's floor.

### spec.autoscaler.schedules[].scheduleName

`string` · required

Name of the schedule (unique within the autoscaler).

- rule: {"required":true}

### spec.autoscaler.schedules[].schedule

`string` · required

Cron expression for when the window STARTS (e.g. "0 8 * * MON-FRI"
for weekday mornings), interpreted in time_zone.

- rule: {"required":true}

### spec.autoscaler.schedules[].durationSec

`int32` · required

How long the window lasts, in seconds. The provider documents a
minimum of 300.

- rule: {"required":true,"int32":{"gte":300}}

### spec.autoscaler.schedules[].minRequiredReplicas

`int32` · required

Minimum replicas the group holds during the window.

- rule: {"required":true,"int32":{"gte":0}}

### spec.autoscaler.schedules[].disabled

`bool` · optional (explicit presence)

Keep the schedule defined but inactive. GCP default: false.

### spec.autoscaler.schedules[].timeZone

`string`

IANA time zone the cron expression is evaluated in (e.g.
"America/New_York"). GCP default: "UTC".

### spec.autoscaler.schedules[].description

`string`

Human-readable description of the schedule.

### spec.autoscaler.stabilizationPeriod

`int32` · optional (explicit presence)

Seconds of load stabilization the autoscaler considers before
scale-in decisions — effectively how long it "remembers" peak load.

- rule: {"int32":{"gte":0}}

### spec.perInstanceConfigs

`[]GcpComputeMigPerInstanceConfig`

Stateful per-instance overrides: pin a specific instance NAME to
preserved disks, IPs, and metadata. The group treats configured
instances as stateful — their identity and preserved state survive
recreation. The config name IS the instance name it applies to.

### spec.perInstanceConfigs[].configName

`string` · required

The per-instance config's name — which IS the name of the managed
instance it applies to (for RECREATE-method groups the instance
names are "<base_instance_name>-<suffix>"; a config for a
not-yet-existing name creates an instance with that name).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.perInstanceConfigs[].preservedState

`GcpComputeMigPreservedState`

Preserved state pinned to the instance: disks, IPs, and metadata
that survive recreation.

### spec.perInstanceConfigs[].preservedState.metadata

`map<string, string>`

Preserved metadata key/value pairs pinned to the instance (merged
over template/all-instances metadata).

### spec.perInstanceConfigs[].preservedState.disks

`[]GcpComputeMigPreservedDisk`

Preserved disks pinned to the instance.

### spec.perInstanceConfigs[].preservedState.disks[].deviceName

`string` · required

Device name the disk is exposed under on the instance — matches
the template's disk device_name when overriding a template disk.

- rule: {"required":true}

### spec.perInstanceConfigs[].preservedState.disks[].source

`string | valueFrom` · required

The disk to attach, referenced as a GcpComputeDisk or a literal
self link. The disk must live in the instance's zone.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.perInstanceConfigs[].preservedState.disks[].mode

`string`

Attachment mode: "READ_WRITE" (GCP default) or "READ_ONLY".

- rule: mode must be READ_WRITE or READ_ONLY

### spec.perInstanceConfigs[].preservedState.disks[].deleteRule

`string`

What happens to the disk when the instance is PERMANENTLY deleted:
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the disk is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the disk is deleted with the
                                      instance

- rule: delete_rule must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.perInstanceConfigs[].preservedState.externalIps

`[]GcpComputeMigPreservedIp`

Preserved EXTERNAL IPs pinned to the instance's interfaces.

### spec.perInstanceConfigs[].preservedState.externalIps[].interfaceName

`string` · required

Interface name the address is pinned to (e.g. "nic0").

- rule: {"required":true}

### spec.perInstanceConfigs[].preservedState.externalIps[].address

`string | valueFrom`

The literal IP address to preserve, or a reference to a reserved
GcpAddress. When omitted, the instance's current address is
adopted as preserved state.

- references: GcpAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.perInstanceConfigs[].preservedState.externalIps[].autoDelete

`string`

What happens to the address when the instance is PERMANENTLY
deleted:
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the address is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
                                      the instance

- rule: auto_delete must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.perInstanceConfigs[].preservedState.internalIps

`[]GcpComputeMigPreservedIp`

Preserved INTERNAL IPs pinned to the instance's interfaces.

### spec.perInstanceConfigs[].preservedState.internalIps[].interfaceName

`string` · required

Interface name the address is pinned to (e.g. "nic0").

- rule: {"required":true}

### spec.perInstanceConfigs[].preservedState.internalIps[].address

`string | valueFrom`

The literal IP address to preserve, or a reference to a reserved
GcpAddress. When omitted, the instance's current address is
adopted as preserved state.

- references: GcpAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.perInstanceConfigs[].preservedState.internalIps[].autoDelete

`string`

What happens to the address when the instance is PERMANENTLY
deleted:
  ""                               -- same as "NEVER" (GCP default)
  "NEVER"                          -- the address is kept
  "ON_PERMANENT_INSTANCE_DELETION" -- the address is released with
                                      the instance

- rule: auto_delete must be NEVER or ON_PERMANENT_INSTANCE_DELETION

### spec.perInstanceConfigs[].minimalAction

`string`

The LEAST disruptive action the group may take to apply this
config to the instance (the update escalates to what the change
actually needs, but never below this): "NONE" (GCP default),
"REFRESH", "RESTART", or "REPLACE".

- rule: minimal_action must be NONE, REFRESH, RESTART, or REPLACE

### spec.perInstanceConfigs[].mostDisruptiveAllowedAction

`string`

The MOST disruptive action the group may take to apply this config
— updates needing more wait instead. "NONE", "REFRESH", "RESTART",
or "REPLACE" (GCP default).

- rule: most_disruptive_allowed_action must be NONE, REFRESH, RESTART, or REPLACE

### spec.perInstanceConfigs[].removeInstanceOnDestroy

`bool` · optional (explicit presence)

When true, removing this config from the spec DELETES the managed
instance itself. Default false: the instance keeps running and
only its stateful config is removed (state application follows
remove_instance_state_on_destroy).

### spec.perInstanceConfigs[].removeInstanceStateOnDestroy

`bool` · optional (explicit presence)

When true, removing this config also removes its preserved state
(disks per their delete rules, IPs, metadata) from the running
instance immediately. Default false: the instance keeps the state
until it is recreated. Irrelevant when remove_instance_on_destroy
deletes the instance altogether.

### spec.resizeRequests

`[]GcpComputeMigResizeRequest`

Queued one-shot capacity requests (Dynamic Workload Scheduler):
ask GCP to add N instances for a bounded duration when capacity
becomes available — the batch/HPC path for scarce shapes. Each
request is immutable once created; deleting an ACCEPTED request
cancels it.

### spec.resizeRequests[].requestName

`string` · required

Name of the resize request (unique within the group).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resizeRequests[].description

`string`

Human-readable description of the request.

### spec.resizeRequests[].resizeBy

`int32` · required

Number of instances to ADD to the group when capacity is granted.

- rule: {"required":true,"int32":{"gte":1}}

### spec.resizeRequests[].requestedRunDurationSeconds

`int64` · optional (explicit presence)

How long the granted instances run before GCP reclaims them, in
seconds (10 minutes to 7 days: 600-604800 — the provider's own
documented bounds). Requires the template's
scheduling.provisioning_model to support bounded runs (FLEX_START).
When omitted, the request asks for unbounded capacity.

- rule: {"int64":{"lte":"604800","gte":"600"}}

### spec.deletionPolicy

`string`

Deletion policy — what happens to the group's resources when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- every resource is deleted (VMs are terminated; disks
               follow their template auto_delete / stateful rules)
  "PREVENT" -- destroy FAILS before touching anything
  "ABANDON" -- resources are removed from management but keep
               running in GCP
Applies to every resource in the kind that supports it (group
manager, autoscaler, regional template, per-instance configs,
resize requests). The ZONAL instance template carries no deletion
policy in the provider — it is always deleted on destroy.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `location_exactly_one`: exactly one of zone (zonal group) or region (regional group) must be set
- `distribution_policy_is_regional_only`: distribution_policy applies only to REGIONAL groups — set region instead of zone
- `instance_flexibility_is_regional_only`: instance_flexibility_policy applies only to REGIONAL groups — set region instead of zone
- `redistribution_type_is_regional_only`: update_policy.instance_redistribution_type applies only to REGIONAL groups — remove it or set region
- `autoscaler_owns_sizing`: target_size and autoscaler fight over the group size on every apply — set autoscaler min/max replicas OR a fixed target_size, never both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpComputeMig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_group` | `string` | The full URL of the group's INSTANCE GROUP — the load-balancer backend handle: a GcpBackendService backend's group takes exactly this value. Zonal: https://www.googleapis.com/compute/v1/projects/{project}/zones/{zone}/instanceGroups/{name} Regional: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/instanceGroups/{name} |
| `status.outputs.self_link` | `string` | Self link of the instance group MANAGER resource (the management surface — distinct from the instance group it manages). |
| `status.outputs.current_template_self_link` | `string` | The unique self link of the template the group currently runs (carries the template's uniqueId, so it changes on every template rotation — compare across applies to confirm a rollout actually rolled). |
| `status.outputs.mig_name` | `string` | Name of the managed instance group as it exists in GCP. |
| `status.outputs.location` | `string` | The group's location: the zone of a zonal group or the region of a regional one. Downstream composition can use this to confirm scope compatibility (a regional backend service needs backends in its own region). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.template.disks[].source` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.template.disks[].diskEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.template.disks[].sourceImageEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.template.disks[].sourceSnapshotEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.template.networkInterfaces[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.template.networkInterfaces[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.template.serviceAccount.email` | GcpServiceAccount | `status.outputs.email` |
| `spec.autoHealing.healthCheck` | GcpHealthCheck | `status.outputs.self_link` |
| `spec.perInstanceConfigs[].preservedState.disks[].source` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.perInstanceConfigs[].preservedState.externalIps[].address` | GcpAddress | `status.outputs.address` |
| `spec.perInstanceConfigs[].preservedState.internalIps[].address` | GcpAddress | `status.outputs.address` |

## See Also

- [Overview](../README.md)

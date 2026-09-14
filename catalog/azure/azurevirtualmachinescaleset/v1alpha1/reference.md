# AzureVirtualMachineScaleSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureVirtualMachineScaleSetSpec** defines the configuration for
creating an Azure Virtual Machine Scale Set: a fleet of identical VMs
managed as one resource -- the template (image, size, OS profile,
disks, network), the fleet controls (instance count, zones, upgrade
policy, spot economics, instance repair), and the orchestration mode.

ARM has exactly ONE scale-set resource type with an orchestration-mode
property, and this component models it that way:
- FLEXIBLE (the default, and Azure's recommendation for new workloads)
  spreads instances across fault domains like a resilient VM group;
  individual VMs can even attach to the set (an AzureVirtualMachine's
  availability.virtual_machine_scale_set_id). It carries per-OS patch
  orchestration, mixed-SKU profiles, and spot/on-demand priority
  mixing.
- UNIFORM is the classic mode behind large stateless fleets: identical
  instances, overprovisioning, automatic OS-image upgrades, spot
  restore, scale-in policy, and gallery applications.
Mode-specific capabilities are gated by validation, so the spec tells
the truth about what each mode supports.

The scale set is a template, so its network and data-disk shapes are
deliberately INLINE (unlike the single VM, which references first-class
NICs and disks): every instance stamps its own NIC and disks from the
template, and they live and die with the instance. What IS shared is
referenced: subnets, load-balancer backend pools (by the load
balancer's name-keyed pool-ID outputs), NSGs, identities, and the
rolling-upgrade health probe.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualMachineScaleSet
metadata:
  name: test-vmss
spec:
  region: eastus

  resourceGroup:
    value: test-rg

  name: web-fleet

  # FLEXIBLE applies by default; this manifest exercises the UNIFORM
  # branch and its mode-gated capabilities (overprovision, scale-in,
  # health probe, automatic OS upgrades, rolling policy) so the plan
  # proves the conditionally-mapped enum seams.
  orchestrationMode: UNIFORM

  skuName: Standard_B2s
  instances: 2

  osProfile:
    computerNamePrefix: web
    linux:
      adminUsername: azureuser
      sshPublicKeys:
        # A throwaway key that exists only so offline plans carry valid
        # key material -- it grants access to nothing.
        - publicKey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBvUEIH5UU/EzaIyRcGhmQNt63nyi+zWgfzYFElbfxVd

  osDisk:
    caching: READ_WRITE
    storageAccountType: STANDARD_LRS
    diskSizeGb: 64

  dataDisks:
    - lun: 0
      caching: NONE
      diskSizeGb: 128
      storageAccountType: DATA_STANDARD_SSD_LRS
      createOption: EMPTY

  sourceImageReference:
    publisher: Canonical
    offer: ubuntu-24_04-lts
    sku: server
    version: latest

  networkInterfaces:
    - name: primary
      primary: true
      acceleratedNetworkingEnabled: false
      ipConfigurations:
        - name: internal
          primary: true
          subnetId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/app
          loadBalancerBackendAddressPoolIds:
            - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/loadBalancers/test-lb/backendAddressPools/web

  upgradePolicy:
    mode: ROLLING
    rolling:
      maxBatchInstancePercent: 20
      maxUnhealthyInstancePercent: 20
      maxUnhealthyUpgradedInstancePercent: 20
      pauseTimeBetweenBatches: PT30S
      prioritizeUnhealthyInstancesEnabled: true
    automaticOsUpgrade:
      enabled: true
    healthProbeId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/loadBalancers/test-lb/probes/http-health

  automaticInstanceRepair:
    enabled: true
    gracePeriod: PT30M
    action: REPLACE

  terminationNotification:
    timeout: PT10M

  extensions:
    - name: health
      publisher: Microsoft.ManagedServices
      type: ApplicationHealthLinux
      typeHandlerVersion: "1.0"
      settings: '{"protocol":"http","port":80,"requestPath":"/healthz"}'

  scaleIn:
    rule: OLDEST_VM

  overprovision: false

  bootDiagnostics:
    enabled: true

  zones: ["1"]

  identity:
    type: USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/fleet-identity

  tags:
    cost-center: compute
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.orchestrationMode` | `enum` |  |  |  |
| `spec.skuName` | `string` | yes |  |  |
| `spec.instances` | `int32` |  |  |  |
| `spec.skuProfile` | `AzureVirtualMachineScaleSetSkuProfile` |  |  |  |
| `spec.skuProfile.allocationStrategy` | `enum` | yes |  |  |
| `spec.skuProfile.vmSizes` | `[]AzureVirtualMachineScaleSetVmSize` | yes |  |  |
| `spec.skuProfile.vmSizes[].name` | `string` | yes |  |  |
| `spec.skuProfile.vmSizes[].rank` | `int32` |  |  |  |
| `spec.osProfile` | `AzureVirtualMachineScaleSetOsProfile` | yes |  |  |
| `spec.osProfile.computerNamePrefix` | `string` |  |  |  |
| `spec.osProfile.linux` | `AzureVirtualMachineScaleSetLinuxProfile` |  |  |  |
| `spec.osProfile.linux.adminUsername` | `string` |  |  |  |
| `spec.osProfile.linux.sshPublicKeys` | `[]AzureVirtualMachineScaleSetSshPublicKey` |  |  |  |
| `spec.osProfile.linux.sshPublicKeys[].publicKey` | `string` | yes |  |  |
| `spec.osProfile.linux.sshPublicKeys[].username` | `string` |  |  |  |
| `spec.osProfile.linux.adminPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.osProfile.linux.disablePasswordAuthentication` | `bool` |  | `true` |  |
| `spec.osProfile.linux.patchMode` | `enum` |  |  |  |
| `spec.osProfile.linux.patchAssessmentMode` | `enum` |  |  |  |
| `spec.osProfile.windows` | `AzureVirtualMachineScaleSetWindowsProfile` |  |  |  |
| `spec.osProfile.windows.adminUsername` | `string` |  |  |  |
| `spec.osProfile.windows.adminPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.osProfile.windows.patchMode` | `enum` |  |  |  |
| `spec.osProfile.windows.patchAssessmentMode` | `enum` |  |  |  |
| `spec.osProfile.windows.automaticUpdatesEnabled` | `bool` |  | `true` |  |
| `spec.osProfile.windows.hotpatchingEnabled` | `bool` |  |  |  |
| `spec.osProfile.windows.timezone` | `string` |  |  |  |
| `spec.osProfile.windows.winrmListeners` | `[]AzureVirtualMachineScaleSetWinrmListener` |  |  |  |
| `spec.osProfile.windows.winrmListeners[].protocol` | `enum` | yes |  |  |
| `spec.osProfile.windows.winrmListeners[].certificateUrl` | `string` |  |  |  |
| `spec.osProfile.windows.additionalUnattendContents` | `[]AzureVirtualMachineScaleSetAdditionalUnattendContent` |  |  |  |
| `spec.osProfile.windows.additionalUnattendContents[].setting` | `enum` | yes |  |  |
| `spec.osProfile.windows.additionalUnattendContents[].content` | `string` (sensitive) | yes |  |  |
| `spec.osProfile.windows.licenseType` | `enum` |  |  |  |
| `spec.osDisk` | `AzureVirtualMachineScaleSetOsDisk` | yes |  |  |
| `spec.osDisk.caching` | `enum` | yes |  |  |
| `spec.osDisk.storageAccountType` | `enum` | yes |  |  |
| `spec.osDisk.diskSizeGb` | `int32` |  |  |  |
| `spec.osDisk.diffDiskSettings` | `AzureVirtualMachineScaleSetDiffDiskSettings` |  |  |  |
| `spec.osDisk.diffDiskSettings.placement` | `enum` |  |  |  |
| `spec.osDisk.diskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.osDisk.secureVmDiskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.osDisk.securityEncryptionType` | `enum` |  |  |  |
| `spec.osDisk.writeAcceleratorEnabled` | `bool` |  |  |  |
| `spec.dataDisks` | `[]AzureVirtualMachineScaleSetDataDisk` |  |  |  |
| `spec.dataDisks[].lun` | `int32` | yes |  |  |
| `spec.dataDisks[].caching` | `enum` | yes |  |  |
| `spec.dataDisks[].diskSizeGb` | `int32` | yes |  |  |
| `spec.dataDisks[].storageAccountType` | `enum` | yes |  |  |
| `spec.dataDisks[].createOption` | `enum` |  |  |  |
| `spec.dataDisks[].name` | `string` |  |  |  |
| `spec.dataDisks[].writeAcceleratorEnabled` | `bool` |  |  |  |
| `spec.dataDisks[].diskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.dataDisks[].ultraSsdDiskIopsReadWrite` | `int64` |  |  |  |
| `spec.dataDisks[].ultraSsdDiskMbpsReadWrite` | `int64` |  |  |  |
| `spec.sourceImageReference` | `AzureVirtualMachineScaleSetSourceImageReference` |  |  |  |
| `spec.sourceImageReference.publisher` | `string` | yes |  |  |
| `spec.sourceImageReference.offer` | `string` | yes |  |  |
| `spec.sourceImageReference.sku` | `string` | yes |  |  |
| `spec.sourceImageReference.version` | `string` | yes |  |  |
| `spec.sourceImageId` | `string` |  |  |  |
| `spec.networkInterfaces` | `[]AzureVirtualMachineScaleSetNetworkInterface` | yes |  |  |
| `spec.networkInterfaces[].name` | `string` | yes |  |  |
| `spec.networkInterfaces[].primary` | `bool` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations` | `[]AzureVirtualMachineScaleSetIpConfiguration` | yes |  |  |
| `spec.networkInterfaces[].ipConfigurations[].name` | `string` | yes |  |  |
| `spec.networkInterfaces[].ipConfigurations[].primary` | `bool` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.networkInterfaces[].ipConfigurations[].version` | `enum` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].loadBalancerBackendAddressPoolIds` | `[]string \| valueFrom` |  |  | AzureLoadBalancer (`status.outputs.backend_pool_ids`) |
| `spec.networkInterfaces[].ipConfigurations[].loadBalancerInboundNatRuleIds` | `[]string \| valueFrom` |  |  | AzureLoadBalancer (`status.outputs.nat_rule_ids`) |
| `spec.networkInterfaces[].ipConfigurations[].applicationGatewayBackendAddressPoolIds` | `[]string \| valueFrom` |  |  | AzureApplicationGateway (`status.outputs.backend_address_pool_ids`) |
| `spec.networkInterfaces[].ipConfigurations[].applicationSecurityGroupIds` | `[]string \| valueFrom` |  |  | AzureApplicationSecurityGroup (`status.outputs.application_security_group_id`) |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress` | `AzureVirtualMachineScaleSetPublicIpAddress` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.name` | `string` | yes |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.domainNameLabel` | `string` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.idleTimeoutInMinutes` | `int32` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.version` | `enum` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.publicIpPrefixId` | `string \| valueFrom` |  |  | AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`) |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags` | `[]AzureVirtualMachineScaleSetIpTag` |  |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags[].type` | `string` | yes |  |  |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags[].tag` | `string` | yes |  |  |
| `spec.networkInterfaces[].dnsServers` | `[]string` |  |  |  |
| `spec.networkInterfaces[].acceleratedNetworkingEnabled` | `bool` |  |  |  |
| `spec.networkInterfaces[].ipForwardingEnabled` | `bool` |  |  |  |
| `spec.networkInterfaces[].networkSecurityGroupId` | `string \| valueFrom` |  |  | AzureNetworkSecurityGroup (`status.outputs.network_security_group_id`) |
| `spec.networkInterfaces[].auxiliaryMode` | `enum` |  |  |  |
| `spec.networkInterfaces[].auxiliarySku` | `enum` |  |  |  |
| `spec.upgradePolicy` | `AzureVirtualMachineScaleSetUpgradePolicy` |  |  |  |
| `spec.upgradePolicy.mode` | `enum` |  |  |  |
| `spec.upgradePolicy.rolling` | `AzureVirtualMachineScaleSetRollingUpgradePolicy` |  |  |  |
| `spec.upgradePolicy.rolling.maxBatchInstancePercent` | `int32` | yes |  |  |
| `spec.upgradePolicy.rolling.maxUnhealthyInstancePercent` | `int32` | yes |  |  |
| `spec.upgradePolicy.rolling.maxUnhealthyUpgradedInstancePercent` | `int32` | yes |  |  |
| `spec.upgradePolicy.rolling.pauseTimeBetweenBatches` | `string` | yes |  |  |
| `spec.upgradePolicy.rolling.crossZoneUpgradesEnabled` | `bool` |  |  |  |
| `spec.upgradePolicy.rolling.prioritizeUnhealthyInstancesEnabled` | `bool` |  |  |  |
| `spec.upgradePolicy.rolling.maximumSurgeInstancesEnabled` | `bool` |  |  |  |
| `spec.upgradePolicy.automaticOsUpgrade` | `AzureVirtualMachineScaleSetAutomaticOsUpgrade` |  |  |  |
| `spec.upgradePolicy.automaticOsUpgrade.enabled` | `bool` |  |  |  |
| `spec.upgradePolicy.automaticOsUpgrade.disableAutomaticRollback` | `bool` |  |  |  |
| `spec.upgradePolicy.healthProbeId` | `string \| valueFrom` |  |  | AzureLoadBalancer (`status.outputs.probe_ids`) |
| `spec.spot` | `AzureVirtualMachineScaleSetSpot` |  |  |  |
| `spec.spot.evictionPolicy` | `enum` | yes |  |  |
| `spec.spot.maxBidPrice` | `double` |  |  |  |
| `spec.spot.restore` | `AzureVirtualMachineScaleSetSpotRestore` |  |  |  |
| `spec.spot.restore.timeout` | `string` |  |  |  |
| `spec.spot.priorityMix` | `AzureVirtualMachineScaleSetPriorityMix` |  |  |  |
| `spec.spot.priorityMix.baseRegularCount` | `int32` |  |  |  |
| `spec.spot.priorityMix.regularPercentageAboveBase` | `int32` |  |  |  |
| `spec.identity` | `AzureVirtualMachineScaleSetIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.security` | `AzureVirtualMachineScaleSetSecurity` |  |  |  |
| `spec.security.secureBootEnabled` | `bool` |  |  |  |
| `spec.security.vtpmEnabled` | `bool` |  |  |  |
| `spec.security.encryptionAtHostEnabled` | `bool` |  |  |  |
| `spec.automaticInstanceRepair` | `AzureVirtualMachineScaleSetAutomaticInstanceRepair` |  |  |  |
| `spec.automaticInstanceRepair.enabled` | `bool` | yes |  |  |
| `spec.automaticInstanceRepair.gracePeriod` | `string` |  |  |  |
| `spec.automaticInstanceRepair.action` | `enum` |  |  |  |
| `spec.terminationNotification` | `AzureVirtualMachineScaleSetTerminationNotification` |  |  |  |
| `spec.terminationNotification.timeout` | `string` |  |  |  |
| `spec.extensions` | `[]AzureVirtualMachineScaleSetExtension` |  |  |  |
| `spec.extensions[].name` | `string` | yes |  |  |
| `spec.extensions[].publisher` | `string` | yes |  |  |
| `spec.extensions[].type` | `string` | yes |  |  |
| `spec.extensions[].typeHandlerVersion` | `string` | yes |  |  |
| `spec.extensions[].autoUpgradeMinorVersionEnabled` | `bool` |  | `true` |  |
| `spec.extensions[].automaticUpgradeEnabled` | `bool` |  |  |  |
| `spec.extensions[].settings` | `string` |  |  |  |
| `spec.extensions[].protectedSettings` | `string` (sensitive) |  |  |  |
| `spec.extensions[].protectedSettingsFromKeyVault` | `AzureVirtualMachineScaleSetExtensionProtectedSettingsFromKeyVault` |  |  |  |
| `spec.extensions[].protectedSettingsFromKeyVault.secretUrl` | `string` | yes |  |  |
| `spec.extensions[].protectedSettingsFromKeyVault.sourceVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.extensions[].provisionAfterExtensions` | `[]string` |  |  |  |
| `spec.extensions[].forceUpdateTag` | `string` |  |  |  |
| `spec.extensions[].failureSuppressionEnabled` | `bool` |  |  |  |
| `spec.extensionsTimeBudget` | `string` |  |  |  |
| `spec.bootDiagnostics` | `AzureVirtualMachineScaleSetBootDiagnostics` |  |  |  |
| `spec.bootDiagnostics.storageAccountUri` | `string` |  |  |  |
| `spec.bootDiagnostics.enabled` | `bool` |  | `true` |  |
| `spec.zones` | `[]string` |  |  |  |
| `spec.zoneBalance` | `bool` |  | `false` |  |
| `spec.platformFaultDomainCount` | `int32` |  |  |  |
| `spec.placement` | `AzureVirtualMachineScaleSetPlacement` |  |  |  |
| `spec.placement.proximityPlacementGroupId` | `string` |  |  |  |
| `spec.placement.capacityReservationGroupId` | `string` |  |  |  |
| `spec.placement.hostGroupId` | `string` |  |  |  |
| `spec.placement.singlePlacementGroup` | `bool` |  |  |  |
| `spec.overprovision` | `bool` |  |  |  |
| `spec.scaleIn` | `AzureVirtualMachineScaleSetScaleIn` |  |  |  |
| `spec.scaleIn.rule` | `enum` |  |  |  |
| `spec.scaleIn.forceDeletionEnabled` | `bool` |  |  |  |
| `spec.doNotRunExtensionsOnOverprovisionedMachines` | `bool` |  |  |  |
| `spec.customData` | `string` (sensitive) |  |  |  |
| `spec.userData` | `string` |  |  |  |
| `spec.provisionVmAgent` | `bool` |  | `true` |  |
| `spec.extensionOperationsEnabled` | `bool` |  | `true` |  |
| `spec.secrets` | `[]AzureVirtualMachineScaleSetSecret` |  |  |  |
| `spec.secrets[].keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.secrets[].certificates` | `[]AzureVirtualMachineScaleSetSecretCertificate` | yes |  |  |
| `spec.secrets[].certificates[].url` | `string` | yes |  |  |
| `spec.secrets[].certificates[].store` | `string` |  |  |  |
| `spec.networkApiVersion` | `string` |  |  |  |
| `spec.plan` | `AzureVirtualMachineScaleSetPlan` |  |  |  |
| `spec.plan.name` | `string` | yes |  |  |
| `spec.plan.product` | `string` | yes |  |  |
| `spec.plan.publisher` | `string` | yes |  |  |
| `spec.galleryApplications` | `[]AzureVirtualMachineScaleSetGalleryApplication` |  |  |  |
| `spec.galleryApplications[].versionId` | `string` | yes |  |  |
| `spec.galleryApplications[].order` | `int32` |  |  |  |
| `spec.galleryApplications[].tag` | `string` |  |  |  |
| `spec.galleryApplications[].configurationBlobUri` | `string` |  |  |  |
| `spec.additionalCapabilities` | `AzureVirtualMachineScaleSetAdditionalCapabilities` |  |  |  |
| `spec.additionalCapabilities.ultraSsdEnabled` | `bool` |  |  |  |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the scale set runs in, e.g. "eastus". Must match
the region of the referenced subnets and load balancer. Changing the
region replaces the scale set.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the scale set will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the scale set, unique within the resource group.
Instance computer names derive from computer_name_prefix (or this
name), so keep it short for Windows fleets (the 15-character
computer-name limit, minus the instance suffix). Changing the name
replaces the scale set.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64"}}

### spec.orchestrationMode

`enum`

How the scale set orchestrates its instances. Unspecified applies
FLEXIBLE -- Azure's recommendation for new workloads. The mode is
fixed at creation and gates mode-specific capabilities (spec-level
validation tells you which).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_orchestration_mode_unspecified` -- Not specified: FLEXIBLE.
- `FLEXIBLE` -- Instances are spread like a resilient VM group: per-OS patch orchestration, mixed-SKU profiles, spot/on-demand mixing, and standalone VMs can attach to the set.
- `UNIFORM` -- The classic large-fleet mode: identical instances, overprovisioning, automatic OS image upgrades, spot restore, scale-in policy, gallery applications, trusted launch.

### spec.skuName

`string` · required

The VM size every instance uses, e.g. "Standard_D2s_v3",
"Standard_B2s". On a FLEXIBLE scale set the special value "Mix"
activates sku_profile (mixed sizes with an allocation strategy).
Resizing updates in place and rolls instances per the upgrade
policy.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.instances

`int32` · optional (explicit presence)

The number of instances. Unset lets the platform manage the count --
right when an autoscaler owns it; set it for a fixed-size fleet.
0-1000.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.skuProfile

`AzureVirtualMachineScaleSetSkuProfile`

FLEXIBLE + sku_name "Mix" only: the mixed VM sizes instances may use
and the strategy for picking among them -- the capacity-resilient
shape for spot fleets and constrained regions.

- rule: vm_sizes ranks are only meaningful with the PRIORITIZED allocation strategy

### spec.skuProfile.allocationStrategy

`enum` · required

How Azure picks among the sizes: LOWEST_PRICE for cost,
CAPACITY_OPTIMIZED for allocation success, PRIORITIZED to honor the
per-size ranks.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_allocation_strategy_unspecified` -- Not specified -- invalid; the strategy is an explicit choice.
- `LOWEST_PRICE` -- Pick the cheapest available size first.
- `CAPACITY_OPTIMIZED` -- Pick the size most likely to allocate successfully.
- `PRIORITIZED` -- Honor the per-size ranks.

### spec.skuProfile.vmSizes

`[]AzureVirtualMachineScaleSetVmSize` · required

The candidate VM sizes (up to 5). Ranks are only meaningful with
the PRIORITIZED strategy.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.skuProfile.vmSizes[].name

`string` · required

The VM size, e.g. "Standard_D2s_v3".

- rule: {"required":true}

### spec.skuProfile.vmSizes[].rank

`int32` · optional (explicit presence)

The size's priority under the PRIORITIZED strategy (0 is highest).

- rule: {"int32":{"lte":3,"gte":0}}

### spec.osProfile

`AzureVirtualMachineScaleSetOsProfile` · required

The operating-system profile: exactly one of `linux` or `windows`,
carrying the OS's authentication and OS-specific management surface
(SSH-first for Linux, password + WinRM/unattend for Windows, per-OS
patch vocabularies on FLEXIBLE sets).

- rule: {"required":true}
- rule: set exactly one OS profile: linux or windows

### spec.osProfile.computerNamePrefix

`string`

The prefix instance computer names derive from (Azure appends a
unique suffix per instance). Unset defaults to the scale-set name.
Keep it at most 9 characters for Windows fleets (the 15-character
computer-name limit minus the 6-character suffix). Fixed at
creation.

### spec.osProfile.linux

`AzureVirtualMachineScaleSetLinuxProfile`

Linux configuration. Exactly one of linux/windows.

### spec.osProfile.linux.adminUsername

`string`

The admin account's username. Fixed at creation.

- rule: {"string":{"maxLen":"64"}}

### spec.osProfile.linux.sshPublicKeys

`[]AzureVirtualMachineScaleSetSshPublicKey`

SSH public keys installed for the admin account -- the production
authentication path. Each key's username defaults to admin_username.

### spec.osProfile.linux.sshPublicKeys[].publicKey

`string` · required

The OpenSSH-format public key (at least 2048-bit RSA or an Ed25519
key), e.g. "ssh-ed25519 AAAA...". Public material -- not a secret.

- rule: {"required":true}

### spec.osProfile.linux.sshPublicKeys[].username

`string`

The account the key is installed for. Unset defaults to the
profile's admin_username -- the common case.

### spec.osProfile.linux.adminPassword

`string | valueFrom` · sensitive

The admin account's password. Only meaningful when
disable_password_authentication is explicitly false; SSH keys are
the production path. Can be a literal or a reference to a secret
(e.g. a Config Manager entry). Fixed at creation.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.osProfile.linux.disablePasswordAuthentication

`bool` · optional (explicit presence)

Whether SSH password authentication is disabled. Azure's default is
true (keys only) -- the right posture; setting false requires
admin_password. Fixed at creation.

- default: `true`

### spec.osProfile.linux.patchMode

`enum`

FLEXIBLE only: how instance OSes are patched. Unspecified applies
Azure's default (LINUX_IMAGE_DEFAULT). LINUX_AUTOMATIC_BY_PLATFORM
hands orchestration to Azure Update Manager and requires the
application health extension.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_linux_patch_mode_unspecified` -- Not specified: Azure's default (ImageDefault).
- `LINUX_IMAGE_DEFAULT` -- The image's own update configuration governs patching.
- `LINUX_AUTOMATIC_BY_PLATFORM` -- Azure Update Manager orchestrates patching (requires the health extension).

### spec.osProfile.linux.patchAssessmentMode

`enum`

FLEXIBLE only: how patch assessment runs. Unspecified applies
Azure's default (ASSESSMENT_IMAGE_DEFAULT);
ASSESSMENT_AUTOMATIC_BY_PLATFORM assesses pending patches daily.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_patch_assessment_mode_unspecified` -- Not specified: Azure's default (ImageDefault).
- `ASSESSMENT_IMAGE_DEFAULT` -- The image's own assessment behavior.
- `ASSESSMENT_AUTOMATIC_BY_PLATFORM` -- Azure assesses pending patches daily.

### spec.osProfile.windows

`AzureVirtualMachineScaleSetWindowsProfile`

Windows configuration. Exactly one of linux/windows.

### spec.osProfile.windows.adminUsername

`string`

The admin account's username. Fixed at creation.

- rule: {"string":{"maxLen":"20"}}

### spec.osProfile.windows.adminPassword

`string | valueFrom` · sensitive

The admin account's password (8-123 characters, 3 of 4 complexity
classes -- ARM enforces). Can be a literal or a reference to a
secret (e.g. a Config Manager entry). Fixed at creation.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.osProfile.windows.patchMode

`enum`

FLEXIBLE only: how instance OSes are patched. Unspecified applies
Azure's default (AUTOMATIC_BY_OS: Windows Update as configured in
the image). WINDOWS_AUTOMATIC_BY_PLATFORM hands orchestration to
Azure Update Manager, requires the application health extension,
and is a prerequisite for hotpatching.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_windows_patch_mode_unspecified` -- Not specified: Azure's default (AutomaticByOS).
- `WINDOWS_MANUAL` -- Windows Update is fully manual.
- `AUTOMATIC_BY_OS` -- Windows Update as configured in the image (Azure's default).
- `WINDOWS_AUTOMATIC_BY_PLATFORM` -- Azure Update Manager orchestrates patching (prerequisite for hotpatching; requires the health extension).

### spec.osProfile.windows.patchAssessmentMode

`enum`

FLEXIBLE only: how patch assessment runs.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_patch_assessment_mode_unspecified` -- Not specified: Azure's default (ImageDefault).
- `ASSESSMENT_IMAGE_DEFAULT` -- The image's own assessment behavior.
- `ASSESSMENT_AUTOMATIC_BY_PLATFORM` -- Azure assesses pending patches daily.

### spec.osProfile.windows.automaticUpdatesEnabled

`bool` · optional (explicit presence)

Whether Windows Update's automatic updates are enabled. Azure's
default is true. Fixed at creation.

- default: `true`

### spec.osProfile.windows.hotpatchingEnabled

`bool`

FLEXIBLE only: hotpatching -- security updates applied without
reboots, on supported Windows Server Azure Edition images only.
Requires patch_mode WINDOWS_AUTOMATIC_BY_PLATFORM and the health
extension.

### spec.osProfile.windows.timezone

`string`

The Windows time zone, e.g. "Pacific Standard Time". Unset uses
UTC. Fixed at creation.

### spec.osProfile.windows.winrmListeners

`[]AzureVirtualMachineScaleSetWinrmListener`

WinRM remote-management listeners. HTTPS listeners reference the
certificate by its Key Vault secret URL.

- rule: an HTTPS WinRM listener requires certificate_url (and HTTP forbids it)

### spec.osProfile.windows.winrmListeners[].protocol

`enum` · required

The listener protocol. HTTPS requires certificate_url.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_winrm_protocol_unspecified` -- Not specified -- invalid; the protocol is an explicit choice.
- `HTTP` -- Unencrypted HTTP (port 5985) -- VNet-internal management only.
- `HTTPS` -- TLS (port 5986); requires certificate_url.

### spec.osProfile.windows.winrmListeners[].certificateUrl

`string`

For HTTPS: the Key Vault secret URL of the listener's certificate,
e.g. "https://{vault}.vault.azure.net/secrets/{name}/{version}". The
vault must be enabled for deployment.

### spec.osProfile.windows.additionalUnattendContents

`[]AzureVirtualMachineScaleSetAdditionalUnattendContent`

Raw unattend.xml fragments injected into Windows setup (AutoLogon /
FirstLogonCommands) for pre-agent bootstrap. The content may embed
credentials, so it is treated as secret material. Fixed at creation.

### spec.osProfile.windows.additionalUnattendContents[].setting

`enum` · required

Which setup pass the fragment configures.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_unattend_setting_unspecified` -- Not specified -- invalid; the pass is an explicit choice.
- `AUTO_LOGON` -- Automatic logon configuration (carries credentials).
- `FIRST_LOGON_COMMANDS` -- Commands run at first logon.

### spec.osProfile.windows.additionalUnattendContents[].content

`string` · required · sensitive

The raw XML fragment. May embed credentials (AutoLogon carries the
admin password), so it is treated as secret material.

- rule: {"required":true}

### spec.osProfile.windows.licenseType

`enum`

Azure Hybrid Benefit: bring an existing Windows license instead of
paying the image's Windows price. Unspecified means no benefit
(regular pay-as-you-go).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_windows_license_type_unspecified` -- Not specified: regular pay-as-you-go image billing.
- `WINDOWS_LICENSE_NONE` -- Explicitly no benefit (ARM's literal None).
- `WINDOWS_CLIENT` -- Bring a Windows Client license.
- `WINDOWS_SERVER` -- Bring a Windows Server license.

### spec.osDisk

`AzureVirtualMachineScaleSetOsDisk` · required

The OS disk template every instance stamps. Inline by definition --
an instance's OS disk is born and dies with it.

- rule: {"required":true}
- rule: disk_encryption_set_id and secure_vm_disk_encryption_set_id are mutually exclusive
- rule: secure_vm_disk_encryption_set_id requires security_encryption_type
- rule: an ephemeral OS disk (diff_disk_settings) requires caching READ_ONLY

### spec.osDisk.caching

`enum` · required

The host-caching mode. READ_WRITE is right for general OS disks;
READ_ONLY suits high-IOPS workloads that re-read hot data (and is
required for ephemeral OS disks); NONE for write-heavy disks.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_disk_caching_unspecified` -- Not specified -- invalid; caching is an explicit choice per disk.
- `NONE` -- No host caching: write-heavy disks, and required for Ultra/ PremiumV2 data disks.
- `READ_ONLY` -- Read caching only: re-read-heavy data, and required for ephemeral OS disks.
- `READ_WRITE` -- Read/write caching: the general-purpose OS-disk mode.

### spec.osDisk.storageAccountType

`enum` · required

The disk's storage SKU. PREMIUM_LRS is the production default; the
ZRS variants survive a zone outage. OS disks cannot use
PremiumV2/Ultra. Changing it replaces instances' disks.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_os_disk_storage_account_type_unspecified` -- Not specified -- invalid; the SKU is an explicit choice.
- `STANDARD_LRS` -- HDD -- dev/test only.
- `STANDARD_SSD_LRS` -- Standard SSD, locally redundant.
- `PREMIUM_LRS` -- Premium SSD, locally redundant -- the production default.
- `STANDARD_SSD_ZRS` -- Standard SSD, zone redundant.
- `PREMIUM_ZRS` -- Premium SSD, zone redundant.

### spec.osDisk.diskSizeGb

`int32` · optional (explicit presence)

The OS disk size in GiB (up to 4095). Unset inherits the image's
size -- correct for almost everything. Can only increase.

- rule: {"int32":{"lte":4095,"gte":1}}

### spec.osDisk.diffDiskSettings

`AzureVirtualMachineScaleSetDiffDiskSettings`

Ephemeral OS disk: instances' OS disks live on local cache/temp
storage instead of remote storage -- free, fast, and WIPED on every
stop/reimage. The natural fit for stateless, image-driven fleets.
Presence makes the OS disk ephemeral; requires caching READ_ONLY.
Fixed at creation.

### spec.osDisk.diffDiskSettings.placement

`enum`

Which local storage hosts the ephemeral OS disk. Unspecified applies
Azure's default (CACHE_DISK when the size's cache is big enough).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_diff_disk_placement_unspecified` -- Not specified: Azure's default (the cache disk when big enough).
- `CACHE_DISK` -- The VM size's cache disk.
- `RESOURCE_DISK` -- The VM size's temp/resource disk.
- `NVME_DISK` -- The VM size's local NVMe disks.

### spec.osDisk.diskEncryptionSetId

`string | valueFrom`

Customer-managed-key encryption: the disk encryption set encrypting
instances' OS disks. A disk encryption set by ARM ID, or a reference
to an AzureDiskEncryptionSet's output. Conflicts with
secure_vm_disk_encryption_set_id.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.osDisk.secureVmDiskEncryptionSetId

`string | valueFrom`

UNIFORM only, for confidential instances with customer-key
guest-state encryption: the disk encryption set for the
VMGuestState blob. A disk encryption set by ARM ID, or a reference to
an AzureDiskEncryptionSet's output. Requires security_encryption_type;
conflicts with disk_encryption_set_id. Fixed at creation.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.osDisk.securityEncryptionType

`enum`

UNIFORM only: confidential-VM encryption of the instance guest
state. VM_GUEST_STATE_ONLY encrypts just the guest-state blob;
DISK_WITH_VM_GUEST_STATE also encrypts the OS disk (and requires
security.secure_boot_enabled). Both require security.vtpm_enabled
and a confidential-capable size. Fixed at creation.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_security_encryption_type_unspecified` -- Not specified: not confidential instances.
- `VM_GUEST_STATE_ONLY` -- Only the instance guest state is encrypted.
- `DISK_WITH_VM_GUEST_STATE` -- The OS disk and guest state are encrypted (requires secure boot).

### spec.osDisk.writeAcceleratorEnabled

`bool`

Write Accelerator for M-series sizes with Premium disks and caching
NONE -- sub-millisecond write latency for database logs.

### spec.dataDisks

`[]AzureVirtualMachineScaleSetDataDisk`

Data-disk templates every instance stamps. Inline (template) disks,
unlike the single VM's referenced first-class disks: each instance
gets its own copies, created and deleted with the instance.

- rule: ultra_ssd_disk_iops_read_write / ultra_ssd_disk_mbps_read_write require storage_account_type ULTRA_SSD_LRS or PREMIUM_V2_LRS

### spec.dataDisks[].lun

`int32` · required · optional (explicit presence)

The logical unit number the disk mounts at (0-2000), unique per
instance -- the stable identity the OS addresses the disk by.
Explicit presence (optional + required) so LUN 0 -- the most common
-- survives proto-JSON serialization, which drops plain zero values.

- rule: {"required":true,"int32":{"lte":2000,"gte":0}}

### spec.dataDisks[].caching

`enum` · required

The host-caching mode. READ_ONLY suits read-heavy data; NONE for
write-heavy volumes and required for Ultra/PremiumV2 disks.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_disk_caching_unspecified` -- Not specified -- invalid; caching is an explicit choice per disk.
- `NONE` -- No host caching: write-heavy disks, and required for Ultra/ PremiumV2 data disks.
- `READ_ONLY` -- Read caching only: re-read-heavy data, and required for ephemeral OS disks.
- `READ_WRITE` -- Read/write caching: the general-purpose OS-disk mode.

### spec.dataDisks[].diskSizeGb

`int32` · required

The disk size in GiB (1-32767).

- rule: {"required":true,"int32":{"lte":32767,"gte":1}}

### spec.dataDisks[].storageAccountType

`enum` · required

The disk's storage SKU. UltraSSD/PremiumV2 unlock dialed IOPS and
throughput below (and require zonal instances).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_data_disk_storage_account_type_unspecified` -- Not specified -- invalid; the SKU is an explicit choice.
- `DATA_STANDARD_LRS` -- HDD -- dev/test only.
- `DATA_STANDARD_SSD_LRS` -- Standard SSD, locally redundant.
- `DATA_PREMIUM_LRS` -- Premium SSD, locally redundant.
- `DATA_PREMIUM_ZRS` -- Premium SSD, zone redundant.
- `ULTRA_SSD_LRS` -- Ultra SSD: dialed IOPS/throughput, zonal instances only.
- `PREMIUM_V2_LRS` -- Premium SSD v2: dialed IOPS/throughput, zonal instances only.
- `DATA_STANDARD_SSD_ZRS` -- Standard SSD, zone redundant.

### spec.dataDisks[].createOption

`enum`

How the disk is created. Unspecified applies Azure's default
(EMPTY); FROM_IMAGE stamps the disk from the source image's data
disks (marketplace images that ship data volumes).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_data_disk_create_option_unspecified` -- Not specified: Azure's default (EMPTY).
- `EMPTY` -- A fresh empty disk per instance.
- `FROM_IMAGE` -- Stamped from the source image's data disks.

### spec.dataDisks[].name

`string`

UNIFORM only: an explicit name for instances' disks. Unset lets
Azure derive one (FLEXIBLE sets always derive -- ARM's orchestrated
surface carries no per-disk name).

### spec.dataDisks[].writeAcceleratorEnabled

`bool`

Write Accelerator for this disk (M-series + Premium + caching NONE).

### spec.dataDisks[].diskEncryptionSetId

`string | valueFrom`

Customer-managed-key encryption for instances' data disks. A disk
encryption set by ARM ID, or a reference to an AzureDiskEncryptionSet's
output.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.dataDisks[].ultraSsdDiskIopsReadWrite

`int64` · optional (explicit presence)

For ULTRA_SSD_LRS / PREMIUM_V2_LRS: the dialed IOPS.

- rule: {"int64":{"gte":"1"}}

### spec.dataDisks[].ultraSsdDiskMbpsReadWrite

`int64` · optional (explicit presence)

For ULTRA_SSD_LRS / PREMIUM_V2_LRS: the dialed throughput in MBps.

- rule: {"int64":{"gte":"1"}}

### spec.sourceImageReference

`AzureVirtualMachineScaleSetSourceImageReference`

Marketplace/platform image to boot from, by its four coordinates
(publisher/offer/sku/version). Exactly one image source: this or
source_image_id.

### spec.sourceImageReference.publisher

`string` · required

The image publisher, e.g. "Canonical".

- rule: {"required":true}

### spec.sourceImageReference.offer

`string` · required

The image offer, e.g. "ubuntu-24_04-lts".

- rule: {"required":true}

### spec.sourceImageReference.sku

`string` · required

The image SKU, e.g. "server".

- rule: {"required":true}

### spec.sourceImageReference.version

`string` · required

The image version, e.g. "latest" or a pinned "24.04.202506100".
With automatic OS upgrades, "latest" keeps the fleet on new image
releases; otherwise it resolves at creation only.

- rule: {"required":true}

### spec.sourceImageId

`string`

A custom or gallery image to boot from, by ARM ID (a managed image,
or a Shared Image Gallery image/version -- community and direct
shared gallery IDs included). Exactly one image source.

### spec.networkInterfaces

`[]AzureVirtualMachineScaleSetNetworkInterface` · required

The network-interface templates every instance stamps -- at least
one; exactly one marked primary (the first, when several are
declared). Subnets, NSGs, load-balancer pools, and public IP
prefixes are referenced; the NICs themselves are per-instance
template stampings.

- rule: {"repeated":{"minItems":"1"}}
- rule: auxiliary_mode and auxiliary_sku must be set together (both or neither)
- rule: when a NIC template has multiple ip_configurations, the first must be marked primary

### spec.networkInterfaces[].name

`string` · required

A label for this NIC template, unique within the scale set.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.networkInterfaces[].primary

`bool`

Whether this is the primary NIC. Exactly one NIC is primary; with a
single NIC ARM treats it as primary automatically, and with multiple
the FIRST must be marked (spec-level validation enforces both).

### spec.networkInterfaces[].ipConfigurations

`[]AzureVirtualMachineScaleSetIpConfiguration` · required

The NIC's IP configurations -- at least one; the first is primary
when several are declared.

- rule: {"repeated":{"minItems":"1"}}

### spec.networkInterfaces[].ipConfigurations[].name

`string` · required

A label for this configuration, unique within the NIC template.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.networkInterfaces[].ipConfigurations[].primary

`bool`

Whether this is the NIC's primary configuration.

### spec.networkInterfaces[].ipConfigurations[].subnetId

`string | valueFrom`

The subnet instances' private addresses live in, by ARM ID.
Required for IPv4 configurations (ARM's contract).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].version

`enum`

The address family. Unspecified applies Azure's default (IPV4).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_ip_version_unspecified` -- Not specified: IPv4.
- `IPV4` -- IPv4.
- `IPV6` -- IPv6 (dual-stack templates pair it with an IPv4 configuration).

### spec.networkInterfaces[].ipConfigurations[].loadBalancerBackendAddressPoolIds

`[]string | valueFrom`

Load-balancer backend pools every instance joins, by pool ARM ID --
membership is expressed from the member side in Azure's model.
Reference a pool through the load balancer's name-keyed map output,
e.g. valueFrom fieldPath "status.outputs.backend_pool_ids.web".

- references: AzureLoadBalancer (`status.outputs.backend_pool_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.backend_pool_ids}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].loadBalancerInboundNatRuleIds

`[]string | valueFrom`

UNIFORM only: pool-style inbound NAT rules whose port ranges map
onto instances, by rule ARM ID. Reference a rule through the load
balancer's name-keyed map output, e.g. valueFrom fieldPath
"status.outputs.nat_rule_ids.per-instance-ssh".

- references: AzureLoadBalancer (`status.outputs.nat_rule_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.nat_rule_ids}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].applicationGatewayBackendAddressPoolIds

`[]string | valueFrom`

Application Gateway backend pools every instance joins, by pool ARM
ID. Reference a pool through the gateway's name-keyed map output,
e.g. valueFrom fieldPath "status.outputs.backend_address_pool_ids.web".

- references: AzureApplicationGateway (`status.outputs.backend_address_pool_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationGateway, name: <that resource's name>, fieldPath: status.outputs.backend_address_pool_ids}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].applicationSecurityGroupIds

`[]string | valueFrom`

Application security groups instances join (up to 20), so NSG rules
can target the fleet as a workload group. Each entry is an application
security group by ARM ID, or a reference to an
AzureApplicationSecurityGroup's output.

- references: AzureApplicationSecurityGroup (`status.outputs.application_security_group_id`)
- rule: {"repeated":{"maxItems":"20"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.application_security_group_id}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress

`AzureVirtualMachineScaleSetPublicIpAddress`

Give every instance its own public IP, stamped from this template.
For fleets reached through a load balancer or gateway leave it
unset -- the production shape.

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.name

`string` · required

A label for the instances' public IPs.

- rule: {"required":true}

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.domainNameLabel

`string`

The DNS label prefix for instances' addresses; Azure appends a
per-instance suffix. Leave unset for IP-only addressing.

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.idleTimeoutInMinutes

`int32` · optional (explicit presence)

The TCP idle timeout for the addresses, in minutes (4-32). Unset
applies Azure's default.

- rule: {"int32":{"lte":32,"gte":4}}

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.version

`enum`

The address family. Unspecified applies Azure's default (IPV4).
Fixed at creation.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_ip_version_unspecified` -- Not specified: IPv4.
- `IPV4` -- IPv4.
- `IPV6` -- IPv6 (dual-stack templates pair it with an IPv4 configuration).

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.publicIpPrefixId

`string | valueFrom`

Draw instance addresses from a reserved public IP prefix, by ARM
ID -- the fleet egresses/ingresses from a known CIDR partners can
allowlist. Fixed at creation.

- references: AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIpPrefix, name: <that resource's name>, fieldPath: status.outputs.public_ip_prefix_id}} -- a bare string does not parse

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags

`[]AzureVirtualMachineScaleSetIpTag`

Azure IP tags applied to instances' addresses (routing metadata,
e.g. RoutingPreference=Internet). Fixed at creation.

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags[].type

`string` · required

The tag type, e.g. "RoutingPreference".

- rule: {"required":true}

### spec.networkInterfaces[].ipConfigurations[].publicIpAddress.ipTags[].tag

`string` · required

The tag value, e.g. "Internet".

- rule: {"required":true}

### spec.networkInterfaces[].dnsServers

`[]string`

DNS servers instances behind this NIC use, overriding the virtual
network's DNS. Rarely set -- prefer configuring DNS on the virtual
network.

### spec.networkInterfaces[].acceleratedNetworkingEnabled

`bool`

Whether accelerated networking (SR-IOV) is enabled. Azure's default
is false, but production fleets on supported sizes should enable it.

### spec.networkInterfaces[].ipForwardingEnabled

`bool`

Whether the NIC forwards traffic not addressed to it -- network
virtual appliance fleets only.

### spec.networkInterfaces[].networkSecurityGroupId

`string | valueFrom`

The network security group filtering these NICs' traffic, by ARM
ID -- the per-fleet complement to the subnet-level NSG.

- references: AzureNetworkSecurityGroup (`status.outputs.network_security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureNetworkSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.network_security_group_id}} -- a bare string does not parse

### spec.networkInterfaces[].auxiliaryMode

`enum`

The auxiliary mode for network-virtual-appliance acceleration (a
preview feature; on FLEXIBLE sets it requires network_api_version
"2022-11-01"). Must be set together with auxiliary_sku.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_auxiliary_mode_unspecified` -- Not specified: no auxiliary acceleration.
- `ACCELERATED_CONNECTIONS` -- Optimizes connections-per-second for appliance fleets.
- `FLOATING` -- Floating IP support for the auxiliary path.

### spec.networkInterfaces[].auxiliarySku

`enum`

The auxiliary SKU sizing the NVA acceleration. Must be set together
with auxiliary_mode.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_auxiliary_sku_unspecified` -- Not specified: no auxiliary SKU.
- `A1` -- The smallest acceleration tier.
- `A2` -- The second acceleration tier.
- `A4` -- The mid acceleration tier.
- `A8` -- The largest acceleration tier.

### spec.upgradePolicy

`AzureVirtualMachineScaleSetUpgradePolicy`

How instances receive template changes and OS image upgrades.
Leave unset for MANUAL (changes apply to new instances only).

### spec.upgradePolicy.mode

`enum`

The upgrade mode. Unspecified applies Azure's default (MANUAL:
template changes apply to NEW instances only; existing ones wait
for a manual upgrade). AUTOMATIC upgrades all instances immediately
(brief fleet-wide disruption); ROLLING upgrades in health-checked
batches -- the production choice, requiring health monitoring.
Fixed at creation.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_upgrade_mode_unspecified` -- Not specified: Azure's default (MANUAL).
- `MANUAL` -- Template changes apply to new instances only; existing instances wait for a manual upgrade.
- `AUTOMATIC` -- All instances upgrade immediately on template changes.
- `ROLLING` -- Instances upgrade in health-checked batches (requires health monitoring and a rolling policy) -- the production choice.

### spec.upgradePolicy.rolling

`AzureVirtualMachineScaleSetRollingUpgradePolicy`

ROLLING mode's batching contract: how much of the fleet upgrades at
once and how much unhealth pauses the rollout. Required when mode
is ROLLING.

### spec.upgradePolicy.rolling.maxBatchInstancePercent

`int32` · required

The maximum percent of the fleet upgraded in one batch (5-100).

- rule: {"required":true,"int32":{"lte":100,"gte":5}}

### spec.upgradePolicy.rolling.maxUnhealthyInstancePercent

`int32` · required

The maximum percent of the fleet that may be unhealthy (from any
cause) before the rollout pauses (5-100).

- rule: {"required":true,"int32":{"lte":100,"gte":5}}

### spec.upgradePolicy.rolling.maxUnhealthyUpgradedInstancePercent

`int32` · required

The maximum percent of UPGRADED instances that may be unhealthy
before the rollout pauses (0-100).

- rule: {"required":true,"int32":{"lte":100,"gte":0}}

### spec.upgradePolicy.rolling.pauseTimeBetweenBatches

`string` · required

The wait between batches, as an ISO 8601 duration (e.g. "PT30S",
"PT5M") -- health checks run during the pause.

- rule: {"required":true}

### spec.upgradePolicy.rolling.crossZoneUpgradesEnabled

`bool`

Upgrade zone by zone instead of mixing zones in a batch. Requires
zones.

### spec.upgradePolicy.rolling.prioritizeUnhealthyInstancesEnabled

`bool`

Upgrade unhealthy instances first (they are already degraded, so
upgrading them costs nothing).

### spec.upgradePolicy.rolling.maximumSurgeInstancesEnabled

`bool`

Surge new instances instead of upgrading in place: each batch
creates fresh instances and deletes old ones after they pass health
checks. Requires overprovision to be explicitly false.

### spec.upgradePolicy.automaticOsUpgrade

`AzureVirtualMachineScaleSetAutomaticOsUpgrade`

UNIFORM only: automatic OS image upgrades -- Azure rolls new image
versions (e.g. "latest" marketplace releases) across the fleet
automatically. Requires mode AUTOMATIC or ROLLING and health
monitoring.

### spec.upgradePolicy.automaticOsUpgrade.enabled

`bool`

Whether Azure automatically rolls new OS image versions across the
fleet.

### spec.upgradePolicy.automaticOsUpgrade.disableAutomaticRollback

`bool`

Disable the automatic rollback that restores the previous image
when an upgrade batch fails health checks. Azure's default keeps
rollback enabled -- the safer posture.

### spec.upgradePolicy.healthProbeId

`string | valueFrom`

UNIFORM only: the load-balancer health probe that reports instance
health for rolling upgrades and instance repair, by probe ARM ID.
Reference a probe through the load balancer's name-keyed map
output, e.g. valueFrom fieldPath
"status.outputs.probe_ids.http-health". FLEXIBLE sets use the
application health extension instead.

- references: AzureLoadBalancer (`status.outputs.probe_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.probe_ids}} -- a bare string does not parse

### spec.spot

`AzureVirtualMachineScaleSetSpot`

Run the fleet on spot capacity: deeply discounted, evictable when
Azure needs the capacity back. Presence makes instances spot;
absence is a regular on-demand fleet. Fixed at creation.

### spec.spot.evictionPolicy

`enum` · required

What happens when Azure evicts an instance: DEALLOCATE stops it
(compute billing stops, disks persist, restore can bring it back);
DELETE removes it and its disks -- the usual fleet choice, letting
autoscaling replace capacity. Fixed at creation.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_eviction_policy_unspecified` -- Not specified -- invalid; eviction behavior is an explicit choice.
- `DEALLOCATE` -- Stop the instance (billing stops, disks persist, restorable).
- `DELETE` -- Delete the instance and its disks -- the usual fleet choice.

### spec.spot.maxBidPrice

`double` · optional (explicit presence)

The maximum hourly price per instance in US dollars, or -1 (the
default) to pay up to the on-demand price and never be evicted on
price.

- rule: {"double":{"gte":-1}}

### spec.spot.restore

`AzureVirtualMachineScaleSetSpotRestore`

UNIFORM only: automatically try to restore evicted spot instances
when capacity returns. Presence enables it.

### spec.spot.restore.timeout

`string`

How long Azure keeps trying to restore evicted instances, as an ISO
8601 duration between PT15M and PT2H. Empty applies Azure's default
(PT1H).

### spec.spot.priorityMix

`AzureVirtualMachineScaleSetPriorityMix`

FLEXIBLE only: mix spot and on-demand instances -- a guaranteed
on-demand base plus a spot/on-demand ratio above it. The
cost/availability middle ground for fleets that cannot be all-spot.

### spec.spot.priorityMix.baseRegularCount

`int32` · optional (explicit presence)

Instances guaranteed to run on-demand before any spot instance is
considered (0-1000). Unset applies Azure's default (0).

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.spot.priorityMix.regularPercentageAboveBase

`int32` · optional (explicit presence)

The percentage of instances ABOVE the base that run on-demand
(0-100; the rest are spot). Unset applies Azure's default (0 -- all
spot above the base).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.identity

`AzureVirtualMachineScaleSetIdentity`

The scale set's managed identity: how instances authenticate to
Azure services without stored credentials. On FLEXIBLE sets only
user-assigned identities are supported (ARM's contract).

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the scale set (UNIFORM only -- FLEXIBLE sets support
USER_ASSIGNED only, ARM's contract); USER_ASSIGNED brings
identities you manage and share; SYSTEM_AND_USER_ASSIGNED carries
both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_identity_type_unspecified` -- Not specified: no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the scale set (UNIFORM only).
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones (UNIFORM only).

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to instances, by ARM ID. Reference
AzureUserAssignedIdentity resources so grants can be composed
before the fleet exists.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.security

`AzureVirtualMachineScaleSetSecurity`

The trusted-launch / encryption security posture. secure_boot and
vtpm are UNIFORM-only capabilities; encryption_at_host works in both
modes.

### spec.security.secureBootEnabled

`bool`

UNIFORM only: UEFI secure boot -- only signed boot components load.
With vtpm_enabled this is "trusted launch". Fixed at creation.

### spec.security.vtpmEnabled

`bool`

UNIFORM only: virtual TPM -- measured boot and attestation;
required for confidential guest-state encryption. Fixed at
creation.

### spec.security.encryptionAtHostEnabled

`bool`

Encryption at host: instance data is encrypted on the compute host
itself, covering temp disks and disk caches. The subscription must
have the EncryptionAtHost feature registered.

### spec.automaticInstanceRepair

`AzureVirtualMachineScaleSetAutomaticInstanceRepair`

Automatic instance repair: unhealthy instances (per the health
extension, or the health probe on UNIFORM sets) are replaced
automatically. Requires health monitoring to be configured.

### spec.automaticInstanceRepair.enabled

`bool` · required

Whether automatic repair is on. Requires health monitoring (the
application health extension, or the health probe on UNIFORM sets).

- rule: {"required":true}

### spec.automaticInstanceRepair.gracePeriod

`string`

The grace period after an instance becomes available before repair
kicks in, as an ISO 8601 duration between PT10M and PT90M. Empty
applies Azure's default.

### spec.automaticInstanceRepair.action

`enum`

What repair does. Unspecified applies Azure's default (REPLACE:
delete and recreate). RESTART and REIMAGE are lighter-weight
alternatives.

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_repair_action_unspecified` -- Not specified: Azure's default (REPLACE).
- `REPLACE` -- Delete the unhealthy instance and create a fresh one.
- `RESTART` -- Restart the unhealthy instance in place.
- `REIMAGE` -- Reimage the unhealthy instance (fresh OS disk, same identity).

### spec.terminationNotification

`AzureVirtualMachineScaleSetTerminationNotification`

Emits a scheduled event before an instance is terminated, giving the
workload up to 15 minutes to drain. Presence enables it.

### spec.terminationNotification.timeout

`string`

How long before termination the event fires, as an ISO 8601
duration between PT5M and PT15M. Empty applies Azure's default
(PT5M).

### spec.extensions

`[]AzureVirtualMachineScaleSetExtension`

VM extensions installed onto every instance -- the health extension
(ApplicationHealthLinux / ApplicationHealthWindows) is the load-
bearing one: rolling upgrades, automatic instance repair, and
hotpatching all key on instance health.

- rule: protected_settings and protected_settings_from_key_vault are mutually exclusive

### spec.extensions[].name

`string` · required

The extension's name, unique within the scale set.

- rule: {"required":true}

### spec.extensions[].publisher

`string` · required

The extension publisher, e.g. "Microsoft.ManagedServices".

- rule: {"required":true}

### spec.extensions[].type

`string` · required

The extension type, e.g. "ApplicationHealthLinux" -- the health
extension rolling upgrades and instance repair key on.

- rule: {"required":true}

### spec.extensions[].typeHandlerVersion

`string` · required

The extension version, e.g. "1.0" (major.minor).

- rule: {"required":true}

### spec.extensions[].autoUpgradeMinorVersionEnabled

`bool` · optional (explicit presence)

Whether the platform picks up new MINOR versions automatically.
Azure's default is true.

- default: `true`

### spec.extensions[].automaticUpgradeEnabled

`bool`

UNIFORM only: whether the platform upgrades the extension across
MAJOR versions automatically, for extensions that support it (ARM's
orchestrated extension surface does not carry the toggle).

### spec.extensions[].settings

`string`

The extension's public settings, as a JSON object string.

### spec.extensions[].protectedSettings

`string` · sensitive

The extension's protected settings, as a JSON object string --
credentials, connection strings, keys. Secret material. Conflicts
with protected_settings_from_key_vault.

### spec.extensions[].protectedSettingsFromKeyVault

`AzureVirtualMachineScaleSetExtensionProtectedSettingsFromKeyVault`

Source the protected settings from a Key Vault secret instead of
inlining them. Conflicts with protected_settings.

### spec.extensions[].protectedSettingsFromKeyVault.secretUrl

`string` · required

The secret's versioned URL, e.g.
"https://{vault}.vault.azure.net/secrets/{name}/{version}".

- rule: {"required":true}

### spec.extensions[].protectedSettingsFromKeyVault.sourceVaultId

`string | valueFrom` · required

The vault holding the secret, by ARM ID. Can be a literal or a
reference to an AzureKeyVault's id output.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.extensions[].provisionAfterExtensions

`[]string`

Extensions that must provision before this one, by name.

### spec.extensions[].forceUpdateTag

`string`

A value that, when changed, forces the extension to re-run even if
its settings did not change.

### spec.extensions[].failureSuppressionEnabled

`bool`

FLEXIBLE only: whether an extension failure is suppressed instead
of failing the instance deployment. Azure's default is false.

### spec.extensionsTimeBudget

`string`

How long ALL extensions on an instance may collectively take to
provision, as an ISO 8601 duration between PT15M and PT2H. Unset
applies Azure's default (PT1H30M).

### spec.bootDiagnostics

`AzureVirtualMachineScaleSetBootDiagnostics`

Boot diagnostics: serial console output and boot screenshots.
Presence enables it; an empty message uses Azure's managed storage
(the right default), or point storage_account_uri at your own
storage account.

### spec.bootDiagnostics.storageAccountUri

`string`

The storage account to write console logs/screenshots to, by blob
endpoint URI. Empty uses Azure's MANAGED storage -- the right
default (no storage account to operate).

### spec.bootDiagnostics.enabled

`bool` · optional (explicit presence)

Whether boot diagnostics are on. Unset means on: declaring the block
has always meant enabling it, and this switch lets a manifest say the
opposite out loud.

- default: `true`

### spec.zones

`[]string`

Availability zones instances spread across, e.g. ["1","2","3"] for
zone redundancy. Removing a zone from an existing set replaces it.

### spec.zoneBalance

`bool` · optional (explicit presence)

Strictly balance instance counts across zones. Azure's default is
false (best-effort); true requires zones and trades scale-out speed
for exact balance. Fixed at creation.

- default: `false`

### spec.platformFaultDomainCount

`int32` · optional (explicit presence)

The number of fault domains instances spread across. REQUIRED on
FLEXIBLE sets (1 with zones -- zones are the resilience unit -- or
the region's max for regional spreading); optional on UNIFORM sets
(Azure picks). Fixed at creation.

- rule: {"int32":{"gte":1}}

### spec.placement

`AzureVirtualMachineScaleSetPlacement`

Placement constraints: proximity groups, capacity reservations,
dedicated host groups, and placement-group sizing.

- rule: capacity_reservation_group_id and proximity_placement_group_id are mutually exclusive
- rule: capacity_reservation_group_id requires single_placement_group to be explicitly false

### spec.placement.proximityPlacementGroupId

`string`

Co-locates instances with the group for minimal inter-VM latency
(HPC/low-latency clusters), by ARM ID. Plain ARM ID. Conflicts with
capacity_reservation_group_id. Fixed at creation.

### spec.placement.capacityReservationGroupId

`string`

Consumes reserved capacity from a capacity reservation group, by
ARM ID. Conflicts with proximity_placement_group_id and requires
single_placement_group to be explicitly false. Fixed at creation.

### spec.placement.hostGroupId

`string`

UNIFORM only: lets Azure place instances on a dedicated host group,
by ARM ID (single-tenant physical isolation). Fixed at creation.

### spec.placement.singlePlacementGroup

`bool` · optional (explicit presence)

Whether the fleet is confined to a single placement group (max 100
instances). Unset lets Azure decide -- the right default. Large
UNIFORM fleets set false; once false it can never return to true.

### spec.overprovision

`bool` · optional (explicit presence)

UNIFORM only: provision extra instances during scale-out and keep
the first healthy ones -- faster, more reliable scale-out at no
extra cost (overprovisioned instances are not billed). Unset on a
UNIFORM set applies Azure's default (true). No default annotation:
the loader would materialize it onto FLEXIBLE sets, where the field
is illegal.

### spec.scaleIn

`AzureVirtualMachineScaleSetScaleIn`

UNIFORM only: which instances scale-in removes first, and whether
removal force-deletes.

### spec.scaleIn.rule

`enum`

Which instances scale-in removes first. Unspecified applies Azure's
default (DEFAULT: balance zones/domains, then highest instance
IDs).

Allowed values (use exactly as shown):

- `azure_virtual_machine_scale_set_scale_in_rule_unspecified` -- Not specified: Azure's default (DEFAULT).
- `DEFAULT` -- Balance zones/domains, then remove the highest instance IDs.
- `NEWEST_VM` -- Remove the newest instances first.
- `OLDEST_VM` -- Remove the oldest instances first.

### spec.scaleIn.forceDeletionEnabled

`bool`

Force-delete removed instances instead of the graceful path --
faster scale-in, no drain.

### spec.doNotRunExtensionsOnOverprovisionedMachines

`bool` · optional (explicit presence)

UNIFORM only: skip running extensions on overprovisioned instances
that will be discarded anyway. Unset applies Azure's default
(false). No default annotation: the loader would materialize it
onto FLEXIBLE sets, where the field is illegal.

### spec.customData

`string` · sensitive

Cloud-init / provisioning data, base64-encoded, delivered once at
first boot of every instance. May embed bootstrap secrets, so it is
treated as secret material.

- rule: {"string":{"maxBytes":"65536"}}

### spec.userData

`string`

Arbitrary machine-readable data, base64-encoded, retrievable from
inside instances via the Instance Metadata Service at any time --
unlike custom_data it is UPDATABLE in place and readable back, so
never put secrets here.

- rule: {"string":{"maxBytes":"65536"}}

### spec.provisionVmAgent

`bool` · optional (explicit presence)

Whether the Azure VM agent is provisioned on instances. Azure's
default is true; false is for appliance images that ship without an
agent -- it disables extensions and most platform management. Fixed
at creation.

- default: `true`

### spec.extensionOperationsEnabled

`bool` · optional (explicit presence)

Whether extension operations are allowed. Azure's default is true;
false requires the set to declare no extensions. Fixed at creation.

- default: `true`

### spec.secrets

`[]AzureVirtualMachineScaleSetSecret`

Certificates from Key Vault installed onto every instance at
provisioning time. Each entry names a vault and the certificate
secret URLs to install (Windows fleets also name the certificate
store).

### spec.secrets[].keyVaultId

`string | valueFrom` · required

The vault holding the certificates, by ARM ID. Can be a literal or
a reference to an AzureKeyVault's id output. The vault must be
enabled for deployment.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.secrets[].certificates

`[]AzureVirtualMachineScaleSetSecretCertificate` · required

The certificates to install from the vault.

- rule: {"repeated":{"minItems":"1"}}

### spec.secrets[].certificates[].url

`string` · required

The certificate's Key Vault secret URL (versioned), e.g.
"https://{vault}.vault.azure.net/secrets/{name}/{version}".

- rule: {"required":true}

### spec.secrets[].certificates[].store

`string`

For WINDOWS fleets: the certificate store to install into (e.g.
"My"). Must stay empty on Linux.

### spec.networkApiVersion

`string`

FLEXIBLE only: the Microsoft.Network API version used for the
instances' networking resources. Unset applies Azure's default
("2020-11-01"); "2022-11-01" unlocks NIC auxiliary acceleration.

- rule: network_api_version must be "2020-11-01" or "2022-11-01" (or unset for Azure's default)

### spec.plan

`AzureVirtualMachineScaleSetPlan`

The marketplace plan for images that require purchase-plan
acceptance (third-party marketplace images). Leave unset for
platform and custom images. Fixed at creation.

### spec.plan.name

`string` · required

The plan name (the image SKU's plan id). Fixed at creation.

- rule: {"required":true}

### spec.plan.product

`string` · required

The product (offer id). Fixed at creation.

- rule: {"required":true}

### spec.plan.publisher

`string` · required

The publisher id. Fixed at creation.

- rule: {"required":true}

### spec.galleryApplications

`[]AzureVirtualMachineScaleSetGalleryApplication`

UNIFORM only: VM Applications (gallery applications) installed onto
every instance, ordered by `order`. Up to 100.

- rule: {"repeated":{"maxItems":"100"}}

### spec.galleryApplications[].versionId

`string` · required

The gallery application VERSION's ARM ID.
Format: .../galleries/{g}/applications/{app}/versions/{v}

- rule: {"required":true}

### spec.galleryApplications[].order

`int32` · optional (explicit presence)

Installation order across the applications (lower installs first);
0 leaves ordering to Azure.

- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.galleryApplications[].tag

`string`

A free-form tag passed to the application's install script.

### spec.galleryApplications[].configurationBlobUri

`string`

A per-instance configuration blob overriding the version's default
configuration, by URI.

### spec.additionalCapabilities

`AzureVirtualMachineScaleSetAdditionalCapabilities`

Niche capability toggles: Ultra SSD attachability for instances.

### spec.additionalCapabilities.ultraSsdEnabled

`bool`

Whether Ultra SSD data disks can attach to instances (requires
zonal placement and a supported size).

### spec.edgeZone

`string`

UNIFORM only: the Azure Edge Zone the scale set is deployed in, for
edge-computing workloads. Leave unset for regular regional
deployment. Fixed at creation.

### spec.tags

`map<string, string>`

Free-form tags applied to the scale set, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Tags are Azure's governance
surface -- Azure Policy enforces them and Microsoft Cost Management
groups by them. Updatable in place.

## Validation Rules

- `vmss_image_source_exactly_one`: set exactly one image source: source_image_reference (marketplace/platform) or source_image_id (custom/gallery)
- `vmss_linux_auth_required`: a Linux scale set requires admin_username and at least one credential -- SSH keys when password authentication is disabled (the default), admin_password when it is enabled
- `vmss_windows_auth_required`: a Windows scale set requires admin_username and admin_password
- `vmss_overprovision_is_uniform_only`: overprovision is a UNIFORM-orchestration capability (FLEXIBLE sets never overprovision)
- `vmss_scale_in_is_uniform_only`: scale_in policy is a UNIFORM-orchestration capability
- `vmss_gallery_applications_are_uniform_only`: gallery_applications are a UNIFORM-orchestration capability
- `vmss_overprovision_extensions_skip_is_uniform_only`: do_not_run_extensions_on_overprovisioned_machines is a UNIFORM-orchestration capability
- `vmss_secure_boot_vtpm_are_uniform_only`: secure_boot_enabled and vtpm_enabled (trusted launch) are UNIFORM-orchestration capabilities on scale sets
- `vmss_os_disk_security_encryption_is_uniform_only`: os_disk confidential-VM encryption (security_encryption_type / secure_vm_disk_encryption_set_id) is a UNIFORM-orchestration capability on scale sets
- `vmss_health_probe_is_uniform_only`: upgrade_policy.health_probe_id is a UNIFORM-orchestration capability (FLEXIBLE sets monitor health via the application health extension)
- `vmss_automatic_os_upgrade_is_uniform_only`: upgrade_policy.automatic_os_upgrade is a UNIFORM-orchestration capability
- `vmss_spot_restore_is_uniform_only`: spot.restore is a UNIFORM-orchestration capability
- `vmss_host_group_is_uniform_only`: placement.host_group_id is a UNIFORM-orchestration capability
- `vmss_edge_zone_is_uniform_only`: edge_zone is a UNIFORM-orchestration capability on scale sets
- `vmss_nat_rules_are_uniform_only`: ip_configuration load_balancer_inbound_nat_rule_ids are a UNIFORM-orchestration capability (on FLEXIBLE sets, NAT attaches to the instance NICs directly)
- `vmss_data_disk_name_is_uniform_only`: data_disks[].name is a UNIFORM-orchestration capability (FLEXIBLE sets derive instance disk names; ARM's orchestrated surface carries no per-disk name)
- `vmss_extension_automatic_upgrade_is_uniform_only`: extensions[].automatic_upgrade_enabled is a UNIFORM-orchestration capability (ARM's orchestrated extension surface does not carry it)
- `vmss_extension_failure_suppression_is_flexible_only`: extensions[].failure_suppression_enabled is a FLEXIBLE-orchestration capability
- `vmss_sku_profile_is_flexible_only`: sku_profile (mixed VM sizes) is a FLEXIBLE-orchestration capability
- `vmss_priority_mix_is_flexible_only`: spot.priority_mix is a FLEXIBLE-orchestration capability
- `vmss_network_api_version_is_flexible_only`: network_api_version is a FLEXIBLE-orchestration setting
- `vmss_patch_modes_are_flexible_only`: per-OS patch_mode / patch_assessment_mode / hotpatching are FLEXIBLE-orchestration capabilities on scale sets
- `vmss_flexible_requires_fault_domain_count`: FLEXIBLE orchestration requires platform_fault_domain_count (1 when spreading across zones, or the region's maximum for regional spreading)
- `vmss_flexible_identity_is_user_assigned_only`: FLEXIBLE-orchestration scale sets support only USER_ASSIGNED managed identities (ARM's contract)
- `vmss_sku_profile_pairs_with_mix`: sku_profile requires sku_name "Mix", and sku_name "Mix" requires sku_profile
- `vmss_zone_balance_requires_zones`: zone_balance requires zones to be specified
- `vmss_rolling_upgrade_policy_pairing`: upgrade_policy.rolling is required when mode is ROLLING and forbidden when mode is MANUAL
- `vmss_rolling_upgrades_need_health`: ROLLING upgrade mode requires health monitoring: an ApplicationHealthLinux/ApplicationHealthWindows extension, or upgrade_policy.health_probe_id on a UNIFORM set
- `vmss_automatic_os_upgrade_needs_automatic_or_rolling`: upgrade_policy.automatic_os_upgrade requires upgrade mode AUTOMATIC or ROLLING
- `vmss_cross_zone_upgrades_require_zones`: upgrade_policy.rolling.cross_zone_upgrades_enabled requires zones
- `vmss_maximum_surge_requires_no_overprovision`: upgrade_policy.rolling.maximum_surge_instances_enabled requires overprovision to be explicitly false
- `vmss_instance_repair_needs_health`: automatic_instance_repair requires health monitoring: an ApplicationHealthLinux/ApplicationHealthWindows extension, or upgrade_policy.health_probe_id on a UNIFORM set
- `vmss_hotpatching_needs_platform_patching_and_health`: hotpatching_enabled requires patch_mode WINDOWS_AUTOMATIC_BY_PLATFORM and an ApplicationHealthWindows extension
- `vmss_platform_patching_needs_health_extension`: patch_mode AUTOMATIC_BY_PLATFORM requires an ApplicationHealthLinux/ApplicationHealthWindows extension
- `vmss_guest_state_encryption_needs_secure_boot_and_vtpm`: os_disk.security_encryption_type requires security.vtpm_enabled (and DISK_WITH_VM_GUEST_STATE additionally requires security.secure_boot_enabled)
- `vmss_encryption_at_host_conflicts_guest_state_disk`: security.encryption_at_host_enabled cannot be combined with os_disk.security_encryption_type DISK_WITH_VM_GUEST_STATE
- `vmss_extension_operations_need_agent`: extension_operations_enabled requires provision_vm_agent
- `vmss_first_network_interface_primary_when_multiple`: when a scale set declares multiple network interfaces, the first must be marked primary
- `vmss_at_most_one_primary_network_interface`: at most one network interface may be marked primary

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualMachineScaleSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.scale_set_id` | `string` | The Azure Resource Manager ID of the scale set. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachineScaleSets/{name} |
| `status.outputs.scale_set_name` | `string` | The name of the scale set. |
| `status.outputs.unique_id` | `string` | The scale set's globally unique ARM-assigned identifier. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The system-assigned managed identity's principal (object) ID -- what AzureRoleAssignment grants reference. Empty unless the identity type includes SYSTEM_ASSIGNED (UNIFORM sets only). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.osDisk.diskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.osDisk.secureVmDiskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.dataDisks[].diskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.networkInterfaces[].ipConfigurations[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.networkInterfaces[].ipConfigurations[].loadBalancerBackendAddressPoolIds` | AzureLoadBalancer | `status.outputs.backend_pool_ids` |
| `spec.networkInterfaces[].ipConfigurations[].loadBalancerInboundNatRuleIds` | AzureLoadBalancer | `status.outputs.nat_rule_ids` |
| `spec.networkInterfaces[].ipConfigurations[].applicationGatewayBackendAddressPoolIds` | AzureApplicationGateway | `status.outputs.backend_address_pool_ids` |
| `spec.networkInterfaces[].ipConfigurations[].applicationSecurityGroupIds` | AzureApplicationSecurityGroup | `status.outputs.application_security_group_id` |
| `spec.networkInterfaces[].ipConfigurations[].publicIpAddress.publicIpPrefixId` | AzurePublicIpPrefix | `status.outputs.public_ip_prefix_id` |
| `spec.networkInterfaces[].networkSecurityGroupId` | AzureNetworkSecurityGroup | `status.outputs.network_security_group_id` |
| `spec.upgradePolicy.healthProbeId` | AzureLoadBalancer | `status.outputs.probe_ids` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.extensions[].protectedSettingsFromKeyVault.sourceVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.secrets[].keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualMachine | `spec.availability.virtualMachineScaleSetId` | `status.outputs.scale_set_id` |

## See Also

- [Overview](../README.md)

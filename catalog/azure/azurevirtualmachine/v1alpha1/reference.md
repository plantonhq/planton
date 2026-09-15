# AzureVirtualMachine

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureVirtualMachineSpec** defines the configuration for creating an
Azure Virtual Machine: the compute instance itself -- its size, image,
OS profile, disks, identity, placement, and security posture.

The VM is deliberately just the machine. Everything it composes with is
referenced, never created here -- matching Azure's own model, where a
VM is a compute shell wired to first-class resources:
- Its network presence is one or more referenced AzureNetworkInterface
  resources (network_interface_ids; a VM can carry several -- management
  + data planes, appliance arms). Public IPs, NSG filtering, and subnet
  placement all live on the NIC.
- Its data volumes are referenced AzureManagedDisk resources attached
  with a LUN and caching mode (data_disk_attachments) -- the data
  outlives the machine, and a shared disk can attach to several VMs.
- Its identities are referenced AzureUserAssignedIdentity resources;
  grants are composed with AzureRoleAssignment against the identity or
  the VM's own system-assigned principal.
Only the OS disk is inline (os_disk): it is born and dies with the VM
by definition -- unless the VM boots from an EXISTING referenced OS
disk (os_managed_disk_id), the disk-swap/golden-disk recovery path.

The OS choice is explicit: os_profile carries exactly one of `linux` or
`windows`, each with its own authentication contract (SSH-first for
Linux, password + WinRM/unattend for Windows) and its own patch-mode
vocabulary -- mirroring ARM's own per-OS surfaces.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualMachine
metadata:
  name: test-vm
  labels:
    environment: production
    team: platform
spec:
  region: eastus

  # The resource group the VM lives in (literal value here; a manifest
  # can also reference an AzureResourceGroup's name output via valueFrom).
  resourceGroup:
    value: test-rg

  name: app-vm
  size: Standard_D2s_v3

  # The VM's network presence: referenced first-class NICs (the first is
  # primary).
  networkInterfaceIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/app-nic

  # Explicit OS choice: SSH-key-only Ubuntu with Azure Update Manager
  # patch orchestration (exercises the per-OS patch-mode mapping).
  osProfile:
    linux:
      adminUsername: azureuser
      sshPublicKeys:
        - publicKey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKp1QgHKux0e/js6p7UBR4jYOtb5aeedkl+0cNr5RB6Q planton-oss-e2e
      patchMode: LINUX_AUTOMATIC_BY_PLATFORM

  osDisk:
    caching: READ_WRITE
    storageAccountType: PREMIUM_LRS

  sourceImageReference:
    publisher: Canonical
    offer: ubuntu-24_04-lts
    sku: server
    version: latest

  # A referenced first-class data disk mounted at LUN 0.
  dataDiskAttachments:
    - managedDiskId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/disks/orders-db-data
      lun: 0
      caching: READ_ONLY

  # Platform-patching orchestration knobs (valid only with the OS
  # profile's AUTOMATIC_BY_PLATFORM patch mode, set above).
  patching:
    assessmentMode: ASSESSMENT_AUTOMATIC_BY_PLATFORM
    rebootSetting: IF_REQUIRED

  identity:
    type: SYSTEM_ASSIGNED

  availability:
    zone: "1"

  security:
    secureBootEnabled: true
    vtpmEnabled: true

  bootDiagnostics:
    enabled: true

  # User tags merged over the metadata-derived tags.
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.size` | `string` | yes |  |  |
| `spec.networkInterfaceIds` | `[]string \| valueFrom` | yes |  | AzureNetworkInterface (`status.outputs.network_interface_id`) |
| `spec.osProfile` | `AzureVirtualMachineOsProfile` | yes |  |  |
| `spec.osProfile.computerName` | `string` |  |  |  |
| `spec.osProfile.linux` | `AzureVirtualMachineLinuxProfile` |  |  |  |
| `spec.osProfile.linux.adminUsername` | `string` |  |  |  |
| `spec.osProfile.linux.sshPublicKeys` | `[]AzureVirtualMachineSshPublicKey` |  |  |  |
| `spec.osProfile.linux.sshPublicKeys[].publicKey` | `string` | yes |  |  |
| `spec.osProfile.linux.sshPublicKeys[].username` | `string` |  |  |  |
| `spec.osProfile.linux.adminPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.osProfile.linux.disablePasswordAuthentication` | `bool` |  | `true` |  |
| `spec.osProfile.linux.patchMode` | `enum` |  |  |  |
| `spec.osProfile.linux.licenseType` | `enum` |  |  |  |
| `spec.osProfile.windows` | `AzureVirtualMachineWindowsProfile` |  |  |  |
| `spec.osProfile.windows.adminUsername` | `string` |  |  |  |
| `spec.osProfile.windows.adminPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.osProfile.windows.patchMode` | `enum` |  |  |  |
| `spec.osProfile.windows.automaticUpdatesEnabled` | `bool` |  | `true` |  |
| `spec.osProfile.windows.hotpatchingEnabled` | `bool` |  |  |  |
| `spec.osProfile.windows.timezone` | `string` |  |  |  |
| `spec.osProfile.windows.winrmListeners` | `[]AzureVirtualMachineWinrmListener` |  |  |  |
| `spec.osProfile.windows.winrmListeners[].protocol` | `enum` | yes |  |  |
| `spec.osProfile.windows.winrmListeners[].certificateUrl` | `string` |  |  |  |
| `spec.osProfile.windows.additionalUnattendContents` | `[]AzureVirtualMachineAdditionalUnattendContent` |  |  |  |
| `spec.osProfile.windows.additionalUnattendContents[].setting` | `enum` | yes |  |  |
| `spec.osProfile.windows.additionalUnattendContents[].content` | `string` (sensitive) | yes |  |  |
| `spec.osProfile.windows.licenseType` | `enum` |  |  |  |
| `spec.osDisk` | `AzureVirtualMachineOsDisk` | yes |  |  |
| `spec.osDisk.caching` | `enum` | yes |  |  |
| `spec.osDisk.storageAccountType` | `enum` | yes |  |  |
| `spec.osDisk.diskSizeGb` | `int32` |  |  |  |
| `spec.osDisk.name` | `string` |  |  |  |
| `spec.osDisk.diffDiskSettings` | `AzureVirtualMachineDiffDiskSettings` |  |  |  |
| `spec.osDisk.diffDiskSettings.placement` | `enum` |  |  |  |
| `spec.osDisk.diskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.osDisk.secureVmDiskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.osDisk.securityEncryptionType` | `enum` |  |  |  |
| `spec.osDisk.writeAcceleratorEnabled` | `bool` |  |  |  |
| `spec.sourceImageReference` | `AzureVirtualMachineSourceImageReference` |  |  |  |
| `spec.sourceImageReference.publisher` | `string` | yes |  |  |
| `spec.sourceImageReference.offer` | `string` | yes |  |  |
| `spec.sourceImageReference.sku` | `string` | yes |  |  |
| `spec.sourceImageReference.version` | `string` | yes |  |  |
| `spec.sourceImageId` | `string` |  |  |  |
| `spec.osManagedDiskId` | `string \| valueFrom` |  |  | AzureManagedDisk (`status.outputs.disk_id`) |
| `spec.dataDiskAttachments` | `[]AzureVirtualMachineDataDiskAttachment` |  |  |  |
| `spec.dataDiskAttachments[].managedDiskId` | `string \| valueFrom` | yes |  | AzureManagedDisk (`status.outputs.disk_id`) |
| `spec.dataDiskAttachments[].lun` | `int32` | yes |  |  |
| `spec.dataDiskAttachments[].caching` | `enum` | yes |  |  |
| `spec.dataDiskAttachments[].writeAcceleratorEnabled` | `bool` |  |  |  |
| `spec.identity` | `AzureVirtualMachineIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.spot` | `AzureVirtualMachineSpot` |  |  |  |
| `spec.spot.evictionPolicy` | `enum` | yes |  |  |
| `spec.spot.maxBidPrice` | `double` |  |  |  |
| `spec.availability` | `AzureVirtualMachineAvailability` |  |  |  |
| `spec.availability.zone` | `string` |  |  |  |
| `spec.availability.availabilitySetId` | `string \| valueFrom` |  |  | AzureAvailabilitySet (`status.outputs.availability_set_id`) |
| `spec.availability.proximityPlacementGroupId` | `string` |  |  |  |
| `spec.availability.capacityReservationGroupId` | `string` |  |  |  |
| `spec.availability.dedicatedHostId` | `string` |  |  |  |
| `spec.availability.dedicatedHostGroupId` | `string` |  |  |  |
| `spec.availability.virtualMachineScaleSetId` | `string \| valueFrom` |  |  | AzureVirtualMachineScaleSet (`status.outputs.scale_set_id`) |
| `spec.availability.platformFaultDomain` | `int32` |  |  |  |
| `spec.security` | `AzureVirtualMachineSecurity` |  |  |  |
| `spec.security.secureBootEnabled` | `bool` |  |  |  |
| `spec.security.vtpmEnabled` | `bool` |  |  |  |
| `spec.security.encryptionAtHostEnabled` | `bool` |  |  |  |
| `spec.patching` | `AzureVirtualMachinePatching` |  |  |  |
| `spec.patching.assessmentMode` | `enum` |  |  |  |
| `spec.patching.rebootSetting` | `enum` |  |  |  |
| `spec.patching.bypassPlatformSafetyChecksOnUserScheduleEnabled` | `bool` |  |  |  |
| `spec.bootDiagnostics` | `AzureVirtualMachineBootDiagnostics` |  |  |  |
| `spec.bootDiagnostics.storageAccountUri` | `string` |  |  |  |
| `spec.bootDiagnostics.enabled` | `bool` |  | `true` |  |
| `spec.galleryApplications` | `[]AzureVirtualMachineGalleryApplication` |  |  |  |
| `spec.galleryApplications[].versionId` | `string` | yes |  |  |
| `spec.galleryApplications[].order` | `int32` |  |  |  |
| `spec.galleryApplications[].tag` | `string` |  |  |  |
| `spec.galleryApplications[].configurationBlobUri` | `string` |  |  |  |
| `spec.galleryApplications[].automaticUpgradeEnabled` | `bool` |  |  |  |
| `spec.galleryApplications[].treatFailureAsDeploymentFailureEnabled` | `bool` |  |  |  |
| `spec.terminationNotification` | `AzureVirtualMachineTerminationNotification` |  |  |  |
| `spec.terminationNotification.timeout` | `string` |  |  |  |
| `spec.osImageNotification` | `AzureVirtualMachineOsImageNotification` |  |  |  |
| `spec.osImageNotification.timeout` | `string` |  |  |  |
| `spec.plan` | `AzureVirtualMachinePlan` |  |  |  |
| `spec.plan.name` | `string` | yes |  |  |
| `spec.plan.product` | `string` | yes |  |  |
| `spec.plan.publisher` | `string` | yes |  |  |
| `spec.customData` | `string` (sensitive) |  |  |  |
| `spec.userData` | `string` |  |  |  |
| `spec.extensionsTimeBudget` | `string` |  |  |  |
| `spec.provisionVmAgent` | `bool` |  | `true` |  |
| `spec.allowExtensionOperations` | `bool` |  | `true` |  |
| `spec.diskControllerType` | `enum` |  |  |  |
| `spec.additionalCapabilities` | `AzureVirtualMachineAdditionalCapabilities` |  |  |  |
| `spec.additionalCapabilities.ultraSsdEnabled` | `bool` |  |  |  |
| `spec.additionalCapabilities.hibernationEnabled` | `bool` |  |  |  |
| `spec.secrets` | `[]AzureVirtualMachineSecret` |  |  |  |
| `spec.secrets[].keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.secrets[].certificates` | `[]AzureVirtualMachineSecretCertificate` | yes |  |  |
| `spec.secrets[].certificates[].url` | `string` | yes |  |  |
| `spec.secrets[].certificates[].store` | `string` |  |  |  |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the VM runs in, e.g. "eastus". Must match the region
of every referenced NIC and disk. Changing the region replaces the
VM.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the VM will be created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the VM, unique within the resource group. 1-64 characters
for Linux, 1-15 for Windows (ARM's limits; the OS's computer name
defaults to this, which is where the Windows limit bites). Changing
the name replaces the VM.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64"}}

### spec.size

`string` · required

The VM size (SKU), e.g. "Standard_D2s_v3", "Standard_B2s". The size
determines vCPUs, memory, temp-disk, accelerated-networking and
ultra-disk support, and hourly cost. Resizing updates in place but
reboots the VM (and may deallocate it when moving size families).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.networkInterfaceIds

`[]string | valueFrom` · required

The network interfaces attached to the VM, by ARM ID -- at least one;
the FIRST is the primary. References first-class
AzureNetworkInterface resources: subnet placement, public IPs, NSG
filtering, and (in Azure's model) load-balancer pool membership all
live NIC-side. Multiple NICs serve appliances and split
management/data planes; the VM size caps how many.

- references: AzureNetworkInterface (`status.outputs.network_interface_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureNetworkInterface, name: <that resource's name>, fieldPath: status.outputs.network_interface_id}} -- a bare string does not parse

### spec.osProfile

`AzureVirtualMachineOsProfile` · required

The operating-system profile: exactly one of `linux` or `windows`,
carrying the OS's authentication and OS-specific management surface.
When the VM boots from an existing OS disk (os_managed_disk_id), the
chosen profile still selects the OS but must carry NO authentication
fields -- the disk already contains its users.

- rule: {"required":true}
- rule: set exactly one OS profile: linux or windows

### spec.osProfile.computerName

`string`

The OS hostname (computer name). Unset defaults to the VM's name --
set it only when the hostname must differ (e.g. a VM name longer
than Windows' 15-character computer-name limit). Fixed at creation.

### spec.osProfile.linux

`AzureVirtualMachineLinuxProfile`

Linux configuration. Exactly one of linux/windows.

### spec.osProfile.linux.adminUsername

`string`

The admin account's username. Required when booting from an image;
must stay empty when booting from an existing OS disk. Fixed at
creation.

- rule: {"string":{"maxLen":"64"}}

### spec.osProfile.linux.sshPublicKeys

`[]AzureVirtualMachineSshPublicKey`

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

How the OS is patched. Unspecified applies Azure's default
(IMAGE_DEFAULT: whatever the image's own update configuration does).
AUTOMATIC_BY_PLATFORM hands patch orchestration to Azure Update
Manager and unlocks patching.reboot_setting and safe scheduled
patching.

Allowed values (use exactly as shown):

- `azure_virtual_machine_linux_patch_mode_unspecified` -- Not specified: Azure's default (ImageDefault).
- `LINUX_IMAGE_DEFAULT` -- The image's own update configuration governs patching.
- `LINUX_AUTOMATIC_BY_PLATFORM` -- Azure Update Manager orchestrates patching.

### spec.osProfile.linux.licenseType

`enum`

Bring-your-own-subscription licensing for commercial distros (Red
Hat, SUSE, Ubuntu Pro). Leave unspecified for regular pay-as-you-go
images.

Allowed values (use exactly as shown):

- `azure_virtual_machine_linux_license_type_unspecified` -- Not specified: regular pay-as-you-go image billing.
- `RHEL_BYOS` -- Red Hat bring-your-own-subscription.
- `RHEL_BASE` -- Red Hat base pay-as-you-go conversion.
- `RHEL_EUS` -- Red Hat Extended Update Support.
- `RHEL_SAPAPPS` -- Red Hat for SAP Applications.
- `RHEL_SAPHA` -- Red Hat for SAP with High Availability.
- `RHEL_BASESAPAPPS` -- Red Hat base for SAP Applications.
- `RHEL_BASESAPHA` -- Red Hat base for SAP with HA.
- `SLES_BYOS` -- SUSE bring-your-own-subscription.
- `SLES_SAP` -- SUSE for SAP.
- `SLES_HPC` -- SUSE for HPC.
- `UBUNTU_PRO` -- Ubuntu Pro attach.

### spec.osProfile.windows

`AzureVirtualMachineWindowsProfile`

Windows configuration. Exactly one of linux/windows.

- rule: hotpatching_enabled requires patch_mode AUTOMATIC_BY_PLATFORM

### spec.osProfile.windows.adminUsername

`string`

The admin account's username. Required when booting from an image;
must stay empty when booting from an existing OS disk. Fixed at
creation.

- rule: {"string":{"maxLen":"20"}}

### spec.osProfile.windows.adminPassword

`string | valueFrom` · sensitive

The admin account's password (8-123 characters, 3 of 4 complexity
classes -- ARM enforces). Can be a literal or a reference to a
secret (e.g. a Config Manager entry). Fixed at creation.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.osProfile.windows.patchMode

`enum`

How the OS is patched. Unspecified applies Azure's default
(AUTOMATIC_BY_OS: Windows Update as configured in the image).
AUTOMATIC_BY_PLATFORM hands orchestration to Azure Update Manager
and is a prerequisite for hotpatching and patching.reboot_setting.

Allowed values (use exactly as shown):

- `azure_virtual_machine_windows_patch_mode_unspecified` -- Not specified: Azure's default (AutomaticByOS).
- `MANUAL` -- Windows Update is fully manual.
- `AUTOMATIC_BY_OS` -- Windows Update as configured in the image (Azure's default).
- `WINDOWS_AUTOMATIC_BY_PLATFORM` -- Azure Update Manager orchestrates patching (prerequisite for hotpatching and reboot control).

### spec.osProfile.windows.automaticUpdatesEnabled

`bool` · optional (explicit presence)

Whether Windows Update's automatic updates are enabled. Azure's
default is true. Fixed at creation.

- default: `true`

### spec.osProfile.windows.hotpatchingEnabled

`bool`

Hotpatching: security updates applied without reboots, on supported
Windows Server Azure Edition images only. Requires patch_mode
AUTOMATIC_BY_PLATFORM.

### spec.osProfile.windows.timezone

`string`

The Windows time zone, e.g. "Pacific Standard Time". Unset uses
UTC. Fixed at creation.

### spec.osProfile.windows.winrmListeners

`[]AzureVirtualMachineWinrmListener`

WinRM remote-management listeners. HTTPS listeners reference the
certificate by its Key Vault secret URL.

- rule: an HTTPS WinRM listener requires certificate_url (and HTTP forbids it)

### spec.osProfile.windows.winrmListeners[].protocol

`enum` · required

The listener protocol. HTTPS requires certificate_url.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_winrm_protocol_unspecified` -- Not specified -- invalid; the protocol is an explicit choice.
- `HTTP` -- Unencrypted HTTP (port 5985) -- VNet-internal management only.
- `HTTPS` -- TLS (port 5986); requires certificate_url.

### spec.osProfile.windows.winrmListeners[].certificateUrl

`string`

For HTTPS: the Key Vault secret URL of the listener's certificate,
e.g. "https://{vault}.vault.azure.net/secrets/{name}/{version}". The
vault must be enabled for deployment.

### spec.osProfile.windows.additionalUnattendContents

`[]AzureVirtualMachineAdditionalUnattendContent`

Raw unattend.xml fragments injected into Windows setup (AutoLogon /
FirstLogonCommands) for pre-agent bootstrap. The content may embed
credentials, so it is treated as secret material. Fixed at creation.

### spec.osProfile.windows.additionalUnattendContents[].setting

`enum` · required

Which setup pass the fragment configures.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_unattend_setting_unspecified` -- Not specified -- invalid; the pass is an explicit choice.
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
(regular pay-as-you-go). Updatable in place.

Allowed values (use exactly as shown):

- `azure_virtual_machine_windows_license_type_unspecified` -- Not specified: regular pay-as-you-go image billing.
- `WINDOWS_LICENSE_NONE` -- Explicitly no benefit (ARM's literal None).
- `WINDOWS_CLIENT` -- Bring a Windows Client license.
- `WINDOWS_SERVER` -- Bring a Windows Server license.

### spec.osDisk

`AzureVirtualMachineOsDisk` · required

The OS disk created with the VM. Always required: it describes the
disk's caching and storage even when the VM boots from an existing
disk. The OS disk is the one deliberately inline disk -- data volumes
are first-class AzureManagedDisk resources (data_disk_attachments).

- rule: {"required":true}
- rule: disk_encryption_set_id and secure_vm_disk_encryption_set_id are mutually exclusive
- rule: secure_vm_disk_encryption_set_id requires security_encryption_type

### spec.osDisk.caching

`enum` · required

The host-caching mode. READ_WRITE is right for general OS disks;
READ_ONLY suits high-IOPS workloads that re-read hot data; NONE for
write-heavy disks where caching only adds latency.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_disk_caching_unspecified` -- Not specified -- invalid; caching is an explicit choice per disk.
- `NONE` -- No host caching: right for write-heavy disks and required for disks larger than 4 TiB.
- `READ_ONLY` -- Read caching only: high-IOPS workloads re-reading hot data.
- `READ_WRITE` -- Read/write caching: the general-purpose OS-disk mode.

### spec.osDisk.storageAccountType

`enum` · required

The disk's storage SKU. PREMIUM_LRS is the production default (and
required for single-VM SLAs); the ZRS variants survive a zone outage
but forbid zone-pinning the VM. OS disks cannot use PremiumV2/Ultra.
Changing it replaces the disk.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_os_disk_storage_account_type_unspecified` -- Not specified -- invalid; the SKU is an explicit choice.
- `STANDARD_LRS` -- HDD -- dev/test only.
- `STANDARD_SSD_LRS` -- Standard SSD, locally redundant.
- `PREMIUM_LRS` -- Premium SSD, locally redundant -- the production default.
- `STANDARD_SSD_ZRS` -- Standard SSD, zone redundant (conflicts with zone-pinning the VM).
- `PREMIUM_ZRS` -- Premium SSD, zone redundant (conflicts with zone-pinning the VM).

### spec.osDisk.diskSizeGb

`int32` · optional (explicit presence)

The OS disk's size in GiB (up to 4095 for OS disks). Unset inherits
the image's size -- correct for almost everything; grow it only when
the OS volume itself (not data -- use data disks) needs room. Can
only increase.

- rule: {"int32":{"lte":4095,"gte":1}}

### spec.osDisk.name

`string`

An explicit name for the OS disk resource. Unset lets Azure derive
one from the VM name. Changing it replaces the disk.

### spec.osDisk.diffDiskSettings

`AzureVirtualMachineDiffDiskSettings`

Ephemeral OS disk: the OS disk lives on the VM's local cache/temp/
NVMe storage instead of remote storage -- free, fast, and WIPED on
every stop/deallocate. For stateless, image-driven fleets only.
Presence makes the OS disk ephemeral. Fixed at creation.

### spec.osDisk.diffDiskSettings.placement

`enum`

Which local storage hosts the ephemeral OS disk. Unspecified applies
Azure's default (CACHE_DISK when the size's cache is big enough).

Allowed values (use exactly as shown):

- `azure_virtual_machine_diff_disk_placement_unspecified` -- Not specified: Azure's default (the cache disk when big enough).
- `CACHE_DISK` -- The VM size's cache disk.
- `RESOURCE_DISK` -- The VM size's temp/resource disk.
- `NVME_DISK` -- The VM size's local NVMe disks.

### spec.osDisk.diskEncryptionSetId

`string | valueFrom`

Customer-managed-key encryption: the disk encryption set encrypting
the OS disk. A disk encryption set by ARM ID, or a reference to an
AzureDiskEncryptionSet's output. Conflicts with
secure_vm_disk_encryption_set_id.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.osDisk.secureVmDiskEncryptionSetId

`string | valueFrom`

For confidential VMs with customer-key guest-state encryption: the
disk encryption set for the VMGuestState blob. A disk encryption set
by ARM ID, or a reference to an AzureDiskEncryptionSet's output.
Requires security_encryption_type; conflicts with
disk_encryption_set_id. Fixed at creation.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.osDisk.securityEncryptionType

`enum`

Confidential-VM encryption of the VM guest state: VM_GUEST_STATE_ONLY
encrypts just the guest-state blob; DISK_WITH_VM_GUEST_STATE also
encrypts the OS disk (and requires security.secure_boot_enabled).
Both require security.vtpm_enabled and a confidential-capable size.
Fixed at creation.

Allowed values (use exactly as shown):

- `azure_virtual_machine_security_encryption_type_unspecified` -- Not specified: not a confidential VM.
- `VM_GUEST_STATE_ONLY` -- Only the VM guest state is encrypted.
- `DISK_WITH_VM_GUEST_STATE` -- The OS disk and guest state are encrypted (requires secure boot).

### spec.osDisk.writeAcceleratorEnabled

`bool`

Write Accelerator for M-series VMs with Premium disks and caching
NONE -- sub-millisecond write latency for database logs.

### spec.sourceImageReference

`AzureVirtualMachineSourceImageReference`

Marketplace/platform image to boot from, by its four coordinates
(publisher/offer/sku/version). Exactly one image source: this,
source_image_id, or os_managed_disk_id.

### spec.sourceImageReference.publisher

`string` · required

The image publisher, e.g. "Canonical". Fixed at creation.

- rule: {"required":true}

### spec.sourceImageReference.offer

`string` · required

The image offer, e.g. "ubuntu-24_04-lts". Fixed at creation.

- rule: {"required":true}

### spec.sourceImageReference.sku

`string` · required

The image SKU, e.g. "server". Fixed at creation.

- rule: {"required":true}

### spec.sourceImageReference.version

`string` · required

The image version, e.g. "latest" or a pinned "24.04.202506100".
"latest" resolves at CREATION only -- the VM does not follow new
image releases afterward. Fixed at creation.

- rule: {"required":true}

### spec.sourceImageId

`string`

A custom or gallery image to boot from, by ARM ID (a managed image,
or a Shared Image Gallery image/version -- community and direct
shared gallery IDs included). Exactly one image source. Fixed at
creation.

### spec.osManagedDiskId

`string | valueFrom`

An EXISTING OS disk to boot from, by ARM ID -- the disk-swap /
golden-disk path. References a first-class AzureManagedDisk that
already carries an operating system. The os_profile must then carry
no authentication fields (the disk has its users), and patching stays
at the image default. Exactly one image source. Fixed at creation.

- references: AzureManagedDisk (`status.outputs.disk_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureManagedDisk, name: <that resource's name>, fieldPath: status.outputs.disk_id}} -- a bare string does not parse

### spec.dataDiskAttachments

`[]AzureVirtualMachineDataDiskAttachment`

Data disks attached to the VM. Each entry references a first-class
AzureManagedDisk by ARM ID and mounts it at a LUN with a caching
mode -- realized as attachment resources on both engines, so the
disk (and its data) outlives the VM. Disks can be attached and
detached in place.

### spec.dataDiskAttachments[].managedDiskId

`string | valueFrom` · required

The managed disk to attach, by ARM ID. References a first-class
AzureManagedDisk so the data outlives the VM.

- references: AzureManagedDisk (`status.outputs.disk_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureManagedDisk, name: <that resource's name>, fieldPath: status.outputs.disk_id}} -- a bare string does not parse

### spec.dataDiskAttachments[].lun

`int32` · required · optional (explicit presence)

The logical unit number the disk mounts at (0-63), unique per VM --
the stable identity the OS addresses the disk by (
/dev/disk/azure/scsi1/lun{n}). Keep LUNs stable across changes.
Explicit presence (optional + required) so LUN 0 -- the most common
-- survives proto-JSON serialization, which drops plain zero values.

- rule: {"required":true,"int32":{"lte":63,"gte":0}}

### spec.dataDiskAttachments[].caching

`enum` · required

The host-caching mode. READ_ONLY suits read-heavy data; NONE is
required for disks larger than 4 TiB and right for write-heavy
volumes (database logs).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_disk_caching_unspecified` -- Not specified -- invalid; caching is an explicit choice per disk.
- `NONE` -- No host caching: right for write-heavy disks and required for disks larger than 4 TiB.
- `READ_ONLY` -- Read caching only: high-IOPS workloads re-reading hot data.
- `READ_WRITE` -- Read/write caching: the general-purpose OS-disk mode.

### spec.dataDiskAttachments[].writeAcceleratorEnabled

`bool`

Write Accelerator for this attachment (M-series + Premium + caching
NONE).

### spec.identity

`AzureVirtualMachineIdentity`

The VM's managed identity: how the workload authenticates to Azure
services without stored credentials. Grants are composed with
AzureRoleAssignment against the system-assigned principal (surfaced
in the outputs) or the referenced user-assigned identities.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the VM (its principal surfaces in the outputs for
AzureRoleAssignment grants); USER_ASSIGNED brings identities you
manage and share across resources; SYSTEM_AND_USER_ASSIGNED carries
both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_identity_type_unspecified` -- Not specified: the VM has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the VM.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the VM, by ARM ID. Reference
AzureUserAssignedIdentity resources so grants can be composed before
the VM exists.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.spot

`AzureVirtualMachineSpot`

Run the VM on spot capacity: deeply discounted, evictable when Azure
needs the capacity back. Presence makes the VM a spot instance;
absence is a regular on-demand VM. For interruption-tolerant
workloads only. Fixed at creation.

### spec.spot.evictionPolicy

`enum` · required

What happens when Azure evicts the VM: DEALLOCATE stops it (compute
billing stops, disks persist, it can restart later); DELETE removes
it and its disks. Fixed at creation.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_virtual_machine_eviction_policy_unspecified` -- Not specified -- invalid; eviction behavior is an explicit choice.
- `DEALLOCATE` -- Stop the VM (billing stops, disks persist, restartable later).
- `DELETE` -- Delete the VM and its disks.

### spec.spot.maxBidPrice

`double` · optional (explicit presence)

The maximum hourly price in US dollars, or -1 (the default) to pay
up to the on-demand price and never be evicted on price. Set a cap
only when cost predictability beats availability.

- rule: {"double":{"gte":-1}}

### spec.availability

`AzureVirtualMachineAvailability`

Where the VM runs relative to Azure's fault machinery: availability
zone, availability set, proximity placement, dedicated hosts,
capacity reservations, or a Flexible scale set. Leave unset for
regional placement with no constraints.

- rule: zone and availability_set_id are mutually exclusive placement strategies
- rule: dedicated_host_id and dedicated_host_group_id are mutually exclusive (pin a host or let Azure pick within the group)
- rule: capacity_reservation_group_id cannot be combined with availability_set_id or proximity_placement_group_id
- rule: platform_fault_domain requires virtual_machine_scale_set_id

### spec.availability.zone

`string`

The availability zone to pin the VM to ("1", "2", or "3"). Zonal
placement is the modern resilience unit -- NICs' public IPs and
zonal disks must match the zone. Conflicts with availability_set_id.
Fixed at creation.

- rule: {"string":{"in":["","1","2","3"]}}

### spec.availability.availabilitySetId

`string | valueFrom`

The classic pre-zones fault/update-domain grouping. Prefer zones
in zoned regions. Can be a literal ARM ID or a reference to an
AzureAvailabilitySet's availability_set_id output. Conflicts with
zone and capacity_reservation_group_id. Fixed at creation.

- references: AzureAvailabilitySet (`status.outputs.availability_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureAvailabilitySet, name: <that resource's name>, fieldPath: status.outputs.availability_set_id}} -- a bare string does not parse

### spec.availability.proximityPlacementGroupId

`string`

Co-locates the VM with its group for minimal inter-VM latency
(HPC/low-latency clusters), by ARM ID. Plain ARM ID.

### spec.availability.capacityReservationGroupId

`string`

Consumes reserved capacity from a capacity reservation group, by ARM
ID -- guaranteed capacity for burst/DR events. Conflicts with
availability_set_id and proximity_placement_group_id.

### spec.availability.dedicatedHostId

`string`

Pins the VM to a specific dedicated host, by ARM ID (single-tenant
physical isolation). Conflicts with dedicated_host_group_id.

### spec.availability.dedicatedHostGroupId

`string`

Lets Azure pick a host within a dedicated host group, by ARM ID.
Conflicts with dedicated_host_id.

### spec.availability.virtualMachineScaleSetId

`string | valueFrom`

Attaches the VM to a FLEXIBLE-orchestration scale set, by ARM ID --
scale-set-managed fault spreading for an individually-managed VM.
Can be a literal ARM ID or a reference to an
AzureVirtualMachineScaleSet's scale_set_id output (the set must be
FLEXIBLE -- UNIFORM sets do not accept attached VMs). Fixed at
creation.

- references: AzureVirtualMachineScaleSet (`status.outputs.scale_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualMachineScaleSet, name: <that resource's name>, fieldPath: status.outputs.scale_set_id}} -- a bare string does not parse

### spec.availability.platformFaultDomain

`int32` · optional (explicit presence)

The fault domain to pin the VM to within virtual_machine_scale_set_id
(requires it). Unset lets Azure choose. Fixed at creation.

- rule: {"int32":{"gte":0}}

### spec.security

`AzureVirtualMachineSecurity`

The trusted-launch / encryption security posture. Leave unset for a
standard VM; production fleets on Gen2 images should enable secure
boot + vTPM (trusted launch).

### spec.security.secureBootEnabled

`bool`

UEFI secure boot: only signed boot components load. With
vtpm_enabled this is "trusted launch" -- the right posture for
production Gen2 images. Fixed at creation.

### spec.security.vtpmEnabled

`bool`

Virtual TPM: measured boot and attestation; required for
confidential-VM guest-state encryption. Fixed at creation.

### spec.security.encryptionAtHostEnabled

`bool`

Encryption at host: data is encrypted on the compute host itself, so
temp disks and disk caches are covered too (the gap platform
encryption leaves). The subscription must have the EncryptionAtHost
feature registered.

### spec.patching

`AzureVirtualMachinePatching`

OS patch orchestration shared across both OSes (the per-OS patch
MODE lives in the linux/windows profile, because ARM's mode
vocabularies differ per OS).

### spec.patching.assessmentMode

`enum`

How patch assessment runs. Unspecified applies Azure's default
(IMAGE_DEFAULT); AUTOMATIC_BY_PLATFORM has Azure assess pending
patches daily.

Allowed values (use exactly as shown):

- `azure_virtual_machine_patch_assessment_mode_unspecified` -- Not specified: Azure's default (ImageDefault).
- `ASSESSMENT_IMAGE_DEFAULT` -- The image's own assessment behavior.
- `ASSESSMENT_AUTOMATIC_BY_PLATFORM` -- Azure assesses pending patches daily.

### spec.patching.rebootSetting

`enum`

When platform patching may reboot the VM. Requires the OS profile's
patch_mode to be AUTOMATIC_BY_PLATFORM.

Allowed values (use exactly as shown):

- `azure_virtual_machine_reboot_setting_unspecified` -- Not specified: no reboot preference expressed.
- `ALWAYS` -- Always reboot after patching.
- `IF_REQUIRED` -- Reboot only when a patch requires it.
- `NEVER` -- Never reboot (patches needing one wait for a manual reboot).

### spec.patching.bypassPlatformSafetyChecksOnUserScheduleEnabled

`bool`

Allows customer-scheduled platform patching to bypass certain
platform safety checks. Requires patch_mode AUTOMATIC_BY_PLATFORM.

### spec.bootDiagnostics

`AzureVirtualMachineBootDiagnostics`

Boot diagnostics: serial console output and boot screenshots, the
first tool for debugging a VM that will not boot. Presence enables
it; an empty message uses Azure's managed storage (the right
default), or point storage_account_uri at your own storage account.

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

### spec.galleryApplications

`[]AzureVirtualMachineGalleryApplication`

VM Applications (gallery applications) installed onto the VM at
deployment -- versioned application packages from an Azure Compute
Gallery, ordered by `order`. Up to 100.

- rule: {"repeated":{"maxItems":"100"}}

### spec.galleryApplications[].versionId

`string` · required

The gallery application VERSION's ARM ID.
Format: .../galleries/{g}/applications/{app}/versions/{v}

- rule: {"required":true}

### spec.galleryApplications[].order

`int32` · optional (explicit presence)

Installation order across the VM's applications (lower installs
first); 0 leaves ordering to Azure.

- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.galleryApplications[].tag

`string`

A free-form tag passed to the application's install script.

### spec.galleryApplications[].configurationBlobUri

`string`

A per-VM configuration blob overriding the version's default
configuration, by URI.

### spec.galleryApplications[].automaticUpgradeEnabled

`bool`

Whether the VM automatically picks up new versions of the
application.

### spec.galleryApplications[].treatFailureAsDeploymentFailureEnabled

`bool`

Whether a failed application deployment fails the whole VM
deployment (Azure's default treats it as best-effort).

### spec.terminationNotification

`AzureVirtualMachineTerminationNotification`

Emits a scheduled event before the VM is terminated, giving the
workload up to 15 minutes to drain. Presence enables it.

### spec.terminationNotification.timeout

`string`

How long before termination the event fires, as an ISO 8601
duration between PT5M and PT15M. Empty applies Azure's default
(PT5M).

### spec.osImageNotification

`AzureVirtualMachineOsImageNotification`

Emits a scheduled event before a platform-initiated OS image
upgrade. Presence enables it.

### spec.osImageNotification.timeout

`string`

How long before the upgrade the event fires. Azure only supports
PT15M; empty applies it.

### spec.plan

`AzureVirtualMachinePlan`

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

### spec.customData

`string` · sensitive

Cloud-init / provisioning data, base64-encoded, delivered once at
first boot. May embed bootstrap secrets, so it is treated as secret
material. Fixed at creation (changing it replaces the VM).

- rule: {"string":{"maxBytes":"65536"}}

### spec.userData

`string`

Arbitrary machine-readable data, base64-encoded, retrievable from
inside the VM via the Instance Metadata Service at any time --
unlike custom_data it is UPDATABLE in place and readable back, so
never put secrets here.

- rule: {"string":{"maxBytes":"65536"}}

### spec.extensionsTimeBudget

`string`

How long ALL extensions on the VM may collectively take to provision,
as an ISO 8601 duration between PT15M and PT2H. Unset applies Azure's
default (PT1H30M).

### spec.provisionVmAgent

`bool` · optional (explicit presence)

Whether the Azure VM agent is provisioned. Azure's default is true;
false is for appliance images that ship without an agent -- it
disables extensions and most platform management, and is fixed at
creation.

- default: `true`

### spec.allowExtensionOperations

`bool` · optional (explicit presence)

Whether extension operations are allowed on the VM. Azure's default
is true; false hard-locks the VM against any extension install.

- default: `true`

### spec.diskControllerType

`enum`

The disk controller the VM presents to the OS. Unspecified applies
Azure's default for the size/image (SCSI today). NVME requires a
supported size + Gen2 image and delivers higher disk throughput.

Allowed values (use exactly as shown):

- `azure_virtual_machine_disk_controller_type_unspecified` -- Not specified: Azure's default for the size/image (SCSI today).
- `SCSI` -- The default, universally supported controller.
- `NVME` -- Higher disk throughput; needs a supported size and a Gen2 image.

### spec.additionalCapabilities

`AzureVirtualMachineAdditionalCapabilities`

Niche capability toggles: Ultra SSD attachability and hibernation
support.

### spec.additionalCapabilities.ultraSsdEnabled

`bool`

Whether Ultra SSD data disks can attach to this VM (requires zonal
placement and a supported size).

### spec.additionalCapabilities.hibernationEnabled

`bool`

Whether the VM supports hibernation (suspend-to-disk; the OS state
persists across deallocation).

### spec.secrets

`[]AzureVirtualMachineSecret`

Certificates from Key Vault installed onto the VM at provisioning
time. Each entry names a vault and the certificate secret URLs to
install (Windows VMs also name the certificate store).

### spec.secrets[].keyVaultId

`string | valueFrom` · required

The vault holding the certificates, by ARM ID. Can be a literal or a
reference to an AzureKeyVault's id output. The vault must be enabled
for deployment.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.secrets[].certificates

`[]AzureVirtualMachineSecretCertificate` · required

The certificates to install from the vault.

- rule: {"repeated":{"minItems":"1"}}

### spec.secrets[].certificates[].url

`string` · required

The certificate's Key Vault secret URL (versioned), e.g.
"https://{vault}.vault.azure.net/secrets/{name}/{version}".

- rule: {"required":true}

### spec.secrets[].certificates[].store

`string`

For WINDOWS VMs: the certificate store to install into (e.g. "My").
Must stay empty on Linux, where certificates land under
/var/lib/waagent.

### spec.edgeZone

`string`

The Azure Edge Zone the VM is deployed in, for edge-computing
workloads. Leave unset for regular regional deployment. Fixed at
creation.

### spec.tags

`map<string, string>`

Free-form tags applied to the VM, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins. Tags are Azure's governance surface -- Azure
Policy enforces them and Microsoft Cost Management groups by them.
Updatable in place.

## Validation Rules

- `vm_image_source_exactly_one`: set exactly one image source: source_image_reference (marketplace/platform), source_image_id (custom/gallery), or os_managed_disk_id (boot from an existing OS disk)
- `vm_linux_auth_required`: a Linux VM booting from an image requires admin_username and at least one credential -- SSH keys when password authentication is disabled (the default), admin_password when it is enabled
- `vm_windows_auth_required`: a Windows VM booting from an image requires admin_username and admin_password
- `vm_boot_from_disk_forbids_auth`: a VM booting from an existing OS disk (os_managed_disk_id) must carry no authentication fields -- the disk already contains its users
- `vm_reboot_setting_needs_platform_patching`: patching.reboot_setting requires the OS profile's patch_mode to be AUTOMATIC_BY_PLATFORM
- `vm_bypass_safety_checks_needs_platform_patching`: patching.bypass_platform_safety_checks_on_user_schedule_enabled requires the OS profile's patch_mode to be AUTOMATIC_BY_PLATFORM
- `vm_guest_state_encryption_needs_secure_boot_and_vtpm`: os_disk.security_encryption_type requires security.vtpm_enabled (and DISK_WITH_VM_GUEST_STATE additionally requires security.secure_boot_enabled)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualMachine, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vm_id` | `string` | The Azure Resource Manager ID of the VM -- what role assignments, diagnostics, and backup policies scope to. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{name} |
| `status.outputs.vm_name` | `string` | The name of the VM. |
| `status.outputs.virtual_machine_guid` | `string` | The 128-bit unique GUID Azure assigns the VM -- what licensing and inventory systems key on (stable across restarts, unlike the ARM id across recreate). |
| `status.outputs.private_ip_address` | `string` | The primary private IP address across the VM's attached NICs -- a convenience echo of the primary AzureNetworkInterface's address. |
| `status.outputs.public_ip_address` | `string` | The primary public IP address across the VM's attached NICs, when any ip configuration is fronted by one (empty for private-only VMs). |
| `status.outputs.computer_name` | `string` | The OS hostname (computer name) the VM booted with. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the VM's system-assigned identity, populated only when the identity type includes SYSTEM_ASSIGNED -- what AzureRoleAssignment grants reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.networkInterfaceIds` | AzureNetworkInterface | `status.outputs.network_interface_id` |
| `spec.osDisk.diskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.osDisk.secureVmDiskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.osManagedDiskId` | AzureManagedDisk | `status.outputs.disk_id` |
| `spec.dataDiskAttachments[].managedDiskId` | AzureManagedDisk | `status.outputs.disk_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.availability.availabilitySetId` | AzureAvailabilitySet | `status.outputs.availability_set_id` |
| `spec.availability.virtualMachineScaleSetId` | AzureVirtualMachineScaleSet | `status.outputs.scale_set_id` |
| `spec.secrets[].keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureBackupProtectedVm | `spec.sourceVmId` | `status.outputs.vm_id` |

## See Also

- [Overview](../README.md)

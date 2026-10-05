# Azure Virtual Machine

Deploys an Azure Virtual Machine — the compute instance itself: its size, boot image, OS profile, OS disk, managed identity, placement, and security posture. The VM is deliberately just the machine, matching Azure's own model: its network presence lives on referenced **AzureNetworkInterface** resources, its data volumes on referenced **AzureManagedDisk** resources, and its identities on referenced **AzureUserAssignedIdentity** resources — everything it composes with is wired through ValueFromRef, never created inline.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Linux or Windows Virtual Machine** -- exactly one OS profile (the spec enforces it): the VM with its size, boot source, admin credentials, patch mode, and licensing
- **OS Disk** -- the one deliberately inline disk, born and dying with the VM: caching, storage SKU, optional explicit size, optional ephemeral mode, and encryption posture
- **Data Disk Attachments** -- created only when `dataDiskAttachments` entries are configured; each mounts a referenced AzureManagedDisk at a LUN with a host-caching mode (the disk and its data outlive the VM)
- **VM Applications** -- created only when `galleryApplications` entries are configured; versioned packages from an Azure Compute Gallery installed at deployment
- **Key Vault Certificates** -- created only when `secrets` entries are configured; certificates installed onto the VM at provisioning from a deployment-enabled vault
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The network interface, managed data disks, and user-assigned identities are NOT created here — they are first-class Infra Components this VM references.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the VM will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **At least one Network Interface** in the same region. Reference an AzureNetworkInterface Infra Component via ValueFromRef (its subnet, IPs, and NSG are configured there), or provide an ARM ID.
- **An SSH public key** (Linux) or an **admin password** (Windows). Passwords are secret material — store them as org secrets and reference them; the platform rejects plaintext.

## Deploy

### Console

Open the deployment store, find **Azure Virtual Machine**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Ubuntu Server with SSH Keys** preset in the [Presets](#presets) tab for a zonal Ubuntu 24.04 LTS VM with SSH-key-only authentication.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualMachine
metadata:
  name: app-server
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: app-vm
  size: Standard_D2s_v5
  networkInterfaceIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Network/networkInterfaces/app-nic
  osProfile:
    linux:
      adminUsername: azureuser
      sshPublicKeys:
        - publicKey: "ssh-ed25519 AAAAC3..."
  osDisk:
    caching: READ_WRITE
    storageAccountType: PREMIUM_LRS
  sourceImageReference:
    publisher: Canonical
    offer: ubuntu-24_04-lts
    sku: server
    version: latest
  identity:
    type: SYSTEM_ASSIGNED
  bootDiagnostics: {}
```

```shell
planton apply -f virtual-machine.yaml
```

This creates an Ubuntu 24.04 LTS VM (Standard_D2s_v5 — 2 vCPUs, 8 GiB), Premium SSD OS disk inheriting the image's size, SSH-key-only authentication, a system-assigned managed identity, and managed-storage boot diagnostics. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the VM to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  networkInterfaceIds:
    - valueFrom:
        kind: AzureNetworkInterface
        name: app-nic
        fieldPath: status.outputs.network_interface_id
  dataDiskAttachments:
    - managedDiskId:
        valueFrom:
          kind: AzureManagedDisk
          name: app-data
          fieldPath: status.outputs.disk_id
      lun: 0
      caching: READ_ONLY
```

The InfraPipeline resolves the dependency graph, deploys the resource group, NIC, and disks first, then provisions the VM with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Virtual Machine. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Boot source** -- exactly one of three: `sourceImageReference` (marketplace/platform image by publisher/offer/sku/version), `sourceImageId` (a custom or Compute Gallery image by ARM ID), or `osManagedDiskId` (boot from an EXISTING referenced OS disk — the golden-disk path, which forbids all authentication fields because the disk carries its users).

**Operating system** -- `osProfile` carries exactly one of `linux` or `windows`, each with its own contract: Linux is SSH-first (`sshPublicKeys`; password auth disabled by Azure default), Windows is password-based with WinRM and unattend.xml as its management channels. Patch modes use per-OS vocabularies.

**VM size** -- `size` fixes vCPUs, memory, temp disk, NIC/disk caps, and hourly cost. `Standard_D2s_v5` is the balanced production entry point. Resizing later updates in place but reboots the VM.

**Networking** -- at least one entry in `networkInterfaceIds`; the FIRST is the primary. Subnet placement, public IPs, and NSG filtering all live on the referenced AzureNetworkInterface — not on the VM.

**Managed identity** -- `identity.type: SYSTEM_ASSIGNED` gives workloads keyless access to Azure services; the principal ID surfaces in the outputs for AzureRoleAssignment grants. User-assigned flavors reference AzureUserAssignedIdentity resources.

**Spot capacity** -- the presence of `spot` makes the VM evictable (up to ~90% off, 30 seconds' notice): set `evictionPolicy` (DEALLOCATE or DELETE) and optionally cap `maxBidPrice` (-1 pays up to the on-demand price). Pair it with `terminationNotification` for a drain signal.

**Security posture** -- `security.secureBootEnabled` + `security.vtpmEnabled` is trusted launch (free, the production posture for Gen2 images); `osDisk.securityEncryptionType` upgrades to a confidential VM on capable sizes.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureNetworkInterface** | `networkInterfaceIds` | `status.outputs.network_interface_id` |
| **AzureManagedDisk** | `osManagedDiskId`, `dataDiskAttachments[].managedDiskId` | `status.outputs.disk_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds` | `status.outputs.identity_id` |
| **AzureVirtualMachineScaleSet** (optional) | `availability.virtualMachineScaleSetId` | `status.outputs.scale_set_id` |
| **AzureDiskEncryptionSet** (optional) | `osDisk.diskEncryptionSetId`, `osDisk.secureVmDiskEncryptionSetId` | `status.outputs.disk_encryption_set_id` |
| **AzureKeyVault** (optional) | `secrets[].keyVaultId` | `status.outputs.key_vault_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vm_id` | Azure Resource Manager ID of the Virtual Machine | Monitor diagnostic settings, Azure Policy assignments, backup policies |
| `private_ip_address` | Private IP of the primary NIC | Application configuration, internal DNS records |
| `public_ip_address` | Public IP of the primary NIC (when one is configured NIC-side) | SSH/RDP access, DNS A records |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity (when enabled) | AzureRoleAssignment grants on Azure resources |

The VM also surfaces `vm_name`, `computer_name`, and `virtual_machine_guid` (the 128-bit GUID licensing and inventory systems key on, stable across restarts) for reference; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Ubuntu SSH server** -- a zonal Ubuntu 24.04 LTS VM authenticated by SSH keys only, attached to a referenced network interface, with managed boot diagnostics. Suitable for application servers, jump boxes, and development VMs. Start from the **Ubuntu Server with SSH Keys** preset.

**Windows Server with trusted launch** -- Windows Server 2022 with secure boot + vTPM, the admin password sourced from a secret reference, and Azure Hybrid Benefit licensing. Start from the **Windows Server with Trusted Launch** preset.

**Spot worker with a data disk** -- an evictable spot VM with a termination notification for drain, a referenced managed data disk at LUN 0, and DEALLOCATE eviction so the disks persist. Start from the **Spot Worker with a Persistent Data Disk** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the VM is created
- [**Azure Network Interface**](/infra-catalog/azure-network-interface) -- provides the VM's network presence (subnet, IPs, NSG)
- [**Azure Managed Disk**](/infra-catalog/azure-managed-disk) -- provides boot and data disks that outlive the machine
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- provides shared workload identities
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants the VM's identities access to Azure resources
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- provides certificates installed at provisioning
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- provides customer-managed keys for the OS disk

# Azure Virtual Machine Scale Set

Deploys an Azure Virtual Machine Scale Set — the fleet primitive: instances stamped from one image, scaled as one unit, upgraded in health-checked batches, and healed automatically. One spec models Azure's own view of the resource (a single scale-set type with an orchestration-mode property), and the IaC modules dispatch onto the right deployment surface: **FLEXIBLE** (Azure's recommendation for new workloads — mixed sizes, spot/on-demand blending, standalone-VM attachment) or **UNIFORM** (the classic identical-instances model — automatic OS image upgrades, overprovisioning, scale-in policies). The mode is fixed at creation and gates roughly sixteen capabilities, each enforced by spec-level validation exactly as ARM enforces them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Linux or Windows Virtual Machine Scale Set** -- exactly one OS profile (the spec enforces it), realized as the FLEXIBLE orchestrated resource or the UNIFORM Linux/Windows resource per the orchestration mode
- **Per-Instance NICs** -- stamped from the declared NIC templates: subnets, NSGs, load-balancer/application-gateway pool membership (member-side wiring through name-keyed map outputs), optional per-instance public IPs
- **OS Disk Template** -- every instance's OS disk, optionally EPHEMERAL (free, fast, wiped on stop/reimage — the stateless-fleet signature) with customer-managed-key encryption
- **Data Disk Templates** -- created only when `dataDisks` entries are configured; per-instance copies that live and die with their instance
- **Upgrade & Health Machinery** -- rolling upgrade batching, automatic instance repair, the termination notification, and (UNIFORM) automatic OS image upgrades and the LB health probe binding
- **Spot Configuration** -- created only when `spot` is configured; eviction policy, bid cap, (UNIFORM) restore, (FLEXIBLE) the spot/on-demand priority mix
- **VM Extensions** -- installed onto every instance; the application health extension is the load-bearing one (rolling upgrades, repair, and platform patching all key on it)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with the user tags, and flowed to every instance

The subnets, load balancer, user-assigned identities, and disk encryption sets are NOT created here — they are first-class Infra Components this fleet references.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the scale set will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A subnet in the same region** for the instances' NICs. Reference an AzureSubnet Infra Component via ValueFromRef.
- **An SSH public key** (Linux) or an **admin password** (Windows). Passwords, cloud-init custom data, and unattend content are secret material — store them as org secrets and reference them; the platform rejects plaintext.
- **For rolling upgrades, repair, or platform patching**: the application health extension (declared in `extensions`), or — on UNIFORM sets — a load-balancer health probe.

## Deploy

### Console

Open the deployment store, find **Azure Virtual Machine Scale Set**, and click **Deploy**. The creation wizard leads with the orchestration mode — the fork every later capability gates on — then walks placement, capacity (with the FLEXIBLE fault-domain contract), the OS and image unions, disks, networking, lifecycle, spot economics, identity, security, and extensions. Start from the **Stateless Web Fleet (Flexible)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualMachineScaleSet
metadata:
  name: web-fleet
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: web-fleet
  skuName: Standard_D2s_v3
  instances: 3
  platformFaultDomainCount: 1
  osProfile:
    computerNamePrefix: web
    linux:
      adminUsername: azureuser
      sshPublicKeys:
        - publicKey: "ssh-ed25519 AAAAC3..."
  osDisk:
    caching: READ_ONLY
    storageAccountType: PREMIUM_LRS
    diffDiskSettings: {}
  sourceImageReference:
    publisher: Canonical
    offer: ubuntu-24_04-lts
    sku: server
    version: latest
  networkInterfaces:
    - name: primary
      primary: true
      acceleratedNetworkingEnabled: true
      ipConfigurations:
        - name: internal
          primary: true
          subnetId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/web-subnet
  zones: ["1", "2", "3"]
```

```shell
planton apply -f scale-set.yaml
```

This creates a FLEXIBLE (the unspecified default — Azure's recommendation) three-instance Ubuntu fleet on ephemeral OS disks, zone-spread with the fault-domain contract satisfied (1 — zones are the resilience unit), SSH-key-only authentication. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the fleet to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  networkInterfaces:
    - name: primary
      primary: true
      ipConfigurations:
        - name: internal
          primary: true
          subnetId:
            valueFrom:
              kind: AzureSubnet
              name: web-subnet
              fieldPath: status.outputs.subnet_id
          loadBalancerBackendAddressPoolIds:
            - valueFrom:
                kind: AzureLoadBalancer
                name: web-lb
                fieldPath: status.outputs.backend_pool_ids.web
  identity:
    type: USER_ASSIGNED
    identityIds:
      - valueFrom:
          kind: AzureUserAssignedIdentity
          name: fleet-identity
          fieldPath: status.outputs.identity_id
```

The InfraPipeline resolves the dependency graph, deploys the group, network, balancer, and identity first, then provisions the fleet with the resolved values — booting already wired into its pools and already authorized.

## Key Configuration

These are the most important decisions when configuring a Virtual Machine Scale Set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Orchestration mode** -- `orchestrationMode` is THE fork, fixed at creation. Unspecified applies FLEXIBLE (Azure's recommendation): mixed sizes (`skuName: Mix` + `skuProfile`), spot priority mixing, per-OS patch orchestration, standalone-VM attachment — but USER_ASSIGNED identities only and a REQUIRED `platformFaultDomainCount` (1 with zones, or the region's max). UNIFORM unlocks automatic OS image upgrades, overprovisioning, scale-in policies, LB health probes, trusted launch, gallery applications, and edge zones.

**Operating system** -- `osProfile` carries exactly one of `linux` (SSH-first; password auth disabled by Azure default) or `windows` (password + WinRM/unattend). Admin passwords, `customData`, and unattend content are secret references.

**Image** -- exactly one of `sourceImageReference` (marketplace, four coordinates) or `sourceImageId` (managed image / Shared Image Gallery — the golden-image path). The image is the fleet's deployment artifact: bump it and roll.

**Upgrades & health** -- `upgradePolicy.mode` is fixed at creation: MANUAL (Azure's default), AUTOMATIC, or ROLLING (health-checked batches — the production choice). Rolling upgrades, `automaticInstanceRepair`, and platform patching ALL require health monitoring: the ApplicationHealthLinux/Windows extension, or (UNIFORM) `upgradePolicy.healthProbeId`.

**Spot** -- the presence of `spot` makes the whole fleet evictable (fixed at creation): `evictionPolicy` DELETE is the fleet norm, `maxBidPrice: -1` pays up to on-demand. FLEXIBLE adds `priorityMix` (a guaranteed on-demand base); UNIFORM adds `restore`.

**Ephemeral OS disks** -- `osDisk.diffDiskSettings` puts instance OS disks on local cache/temp storage: free, faster, wiped on stop/reimage — the stateless-fleet signature. Requires `caching: READ_ONLY`.

**Managed identity** -- `identity.type: USER_ASSIGNED` referencing AzureUserAssignedIdentity resources is the fleet norm (grants compose before the fleet exists); the system-assigned flavors are UNIFORM-only, their principal surfacing in the outputs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `networkInterfaces[].ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| **AzureLoadBalancer** (optional) | `...loadBalancerBackendAddressPoolIds`, `...loadBalancerInboundNatRuleIds`, `upgradePolicy.healthProbeId` | `status.outputs.backend_pool_ids.<name>`, `status.outputs.nat_rule_ids.<name>`, `status.outputs.probe_ids.<name>` |
| **AzureApplicationGateway** (optional) | `...applicationGatewayBackendAddressPoolIds` | `status.outputs.backend_address_pool_ids.<name>` |
| **AzureApplicationSecurityGroup** (optional) | `...applicationSecurityGroupIds` | `status.outputs.application_security_group_id` |
| **AzureNetworkSecurityGroup** (optional) | `networkInterfaces[].networkSecurityGroupId` | `status.outputs.network_security_group_id` |
| **AzurePublicIpPrefix** (optional) | `...publicIpAddress.publicIpPrefixId` | `status.outputs.public_ip_prefix_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds` | `status.outputs.identity_id` |
| **AzureDiskEncryptionSet** (optional) | `osDisk.diskEncryptionSetId`, `osDisk.secureVmDiskEncryptionSetId`, `dataDisks[].diskEncryptionSetId` | `status.outputs.disk_encryption_set_id` |
| **AzureKeyVault** (optional) | `secrets[].keyVaultId`, `extensions[].protectedSettingsFromKeyVault.sourceVaultId` | `status.outputs.key_vault_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `scale_set_id` | Azure Resource Manager ID of the scale set | AzureVirtualMachine `availability.virtualMachineScaleSetId` (attaching a standalone VM to a FLEXIBLE set), autoscale settings, monitoring scopes |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity — empty unless a UNIFORM set carries one | AzureRoleAssignment grants |

The set also surfaces `scale_set_name` and `unique_id` (the globally unique ARM-assigned identifier) for reference; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Stateless web fleet (FLEXIBLE)** -- Ubuntu LTS on ephemeral OS disks, zone-spread, SSH keys, joined to a load-balancer pool. Start from the **Stateless Web Fleet (Flexible)** preset.

**Spot batch fleet** -- mixed sizes at CAPACITY_OPTIMIZED allocation, DELETE eviction with an on-demand priority-mix base, the termination notification for drain, cloud-init bootstrap. Start from the **Spot Batch Fleet (Flexible, Mixed SKUs)** preset.

**Windows fleet with automatic OS upgrades (UNIFORM)** -- Windows Server Azure Edition, rolling upgrades keyed on the LB health probe, automatic OS image upgrades on "latest", automatic repair, Azure Hybrid Benefit. Start from the **Windows Fleet with Automatic OS Upgrades (Uniform)** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the fleet is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the instances' network placement
- [**Azure Load Balancer**](/infra-catalog/azure-load-balancer) -- provides the pools instances join, per-instance NAT, and (UNIFORM) the health probe
- [**Azure Application Gateway**](/infra-catalog/azure-application-gateway) -- provides L7 pools instances join
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- a standalone VM can ATTACH to a FLEXIBLE set by referencing this fleet's `scale_set_id`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- provides the fleet's keyless workload identity
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- provides customer-managed keys for the OS and data disk templates
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- provides certificates installed at provisioning and extension protected settings

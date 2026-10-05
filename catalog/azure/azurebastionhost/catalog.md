# Azure Bastion Host

Deploys Azure Bastion -- the managed jump service for RDP/SSH sessions to virtual machines over their private addresses, with no public IPs or inbound NSG rules on the machines themselves. Four SKUs span two deployment shapes: Developer attaches to a virtual network on free, Azure-shared infrastructure; Basic, Standard, and Premium deploy dedicated infrastructure into the network's `AzureBastionSubnet` with a Standard static public IP (Premium may omit the IP for a private-only host). Billing starts at provisioning, creates run about 10 minutes, and SKU upgrades are in-place while downgrades replace the host.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bastion host** -- the jump service itself: on Basic/Standard/Premium, dedicated infrastructure bound to the `AzureBastionSubnet` and (unless private-only Premium) a Standard static public IP, with the SKU's feature knobs and 2-50 scale units; on Developer, a shared-infrastructure host attached directly to the virtual network
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource ID); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A subnet named EXACTLY `AzureBastionSubnet`** (/26 or larger, carrying nothing else) in the target virtual network -- required for Basic/Standard/Premium via `ipConfiguration.subnetId`. ARM validates the name at deploy time, so a wrongly-named subnet fails the deploy, not the manifest.
- **A Standard-SKU static public IP** the host binds exclusively -- required on Basic and Standard via `ipConfiguration.publicIpAddressId`; Premium may omit it for a private-only host. Sharing the address with a NAT gateway, load balancer, or VPN gateway fails at deploy.
- **Developer SKU instead references the virtual network directly** through `virtualNetworkId` -- no subnet, no public IP, and a limited region list; check availability before choosing it.

## Deploy

### Console

Open the deployment store, find **Azure Bastion Host**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Host**, **Standard with Tunneling**, or **Developer Host** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBastionHost
metadata:
  name: prod-bastion
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: prod-bastion
  sku: BASIC
  ipConfiguration:
    name: bastion-ip-config
    subnetId:
      value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Network/virtualNetworks/prod-hub-vnet/subnets/AzureBastionSubnet"
    publicIpAddressId:
      value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Network/publicIPAddresses/prod-bastion-pip"
```

```shell
planton apply -f bastion.yaml
```

This creates a Basic host on dedicated infrastructure -- fixed 2 scale units, browser-based sessions with copy/paste -- reachable through the bound public IP; expect the create to run 10-14 minutes. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the network alongside the host, wire the subnet and public IP by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-bastion
  sku: BASIC
  ipConfiguration:
    name: bastion-ip-config
    subnetId:
      valueFrom:
        kind: AzureSubnet
        name: bastion-subnet
        fieldPath: status.outputs.subnet_id
    publicIpAddressId:
      valueFrom:
        kind: AzurePublicIp
        name: bastion-pip
        fieldPath: status.outputs.public_ip_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, and public IP first, then provisions the host with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a Bastion host. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Choose the long-term SKU up front -- downgrades replace the host** -- Upgrading (Basic to Standard to Premium) is an in-place, roughly 10-minute configuration change. Downgrading has no path: the host is deleted and recreated, dropping every active session. When session recording or private-only deployment is plausibly in your future, start at Premium -- the cost difference over Standard is marginal.

**Developer is a real tier with real limits** -- Free, no subnet or IP ceremony -- but shared infrastructure: one connection per user, no NSG support on the shared path, no reach across virtual-network peerings, a limited region list, zero feature knobs. Right for a developer poking at a dev VM; wrong the moment two people need concurrent sessions or traffic crosses a peering. Upgrading later brings the full dedicated-infrastructure ceremony with it.

**The subnet is a contract, not a suggestion** -- Dedicated-infrastructure hosts deploy only into a subnet named exactly `AzureBastionSubnet`, /26 or larger, carrying nothing else. The name lives on the referenced subnet, so it surfaces as a deploy-time failure, not a manifest error. Carve the subnet when you design the network, not when you need the host.

**Private-only is Premium's door off the internet** -- A Premium host whose `ipConfiguration` omits the public IP is reachable only from connected networks -- the shape for environments where even the jump service must not face the internet. The choice is fixed at creation (`ipConfiguration` changes replace the host), and the `private_only_enabled` output reports it.

**Plan Kerberos at create time** -- `kerberosEnabled` is honored at create only; the provider silently ignores later changes -- there is no update path. If domain-joined sign-in is on the roadmap, enable it when the host is born; retrofitting means replacing the host.

**Session hygiene is a feature choice** -- Clipboard (`copyPasteEnabled`, on by default) and file copy (`fileCopyEnabled`, Standard+) are exfiltration channels as much as conveniences -- compliance environments commonly disable the clipboard. Shareable links (`shareableLinkEnabled`, Standard+) bypass Azure RBAC for the link holder: treat every link as a credential and prefer them off unless a workflow demands them.

**Scale units are a capacity dial, not an afterthought** -- Each unit carries roughly 20 concurrent sessions; Standard and Premium scale 2-50 in place (Basic and Developer are fixed at 2). Scaling is an update, not a replacement, so start modest and grow before users queue -- there is no reason to over-provision on day one.

**Budget real minutes for every lifecycle operation** -- Measured live, a Basic host's create runs 10-14 minutes and its delete 6-9 minutes; paid tiers are similar or slower. Plan rollouts, teardowns, and SKU changes as maintenance-window work, never as a quick pre-meeting change.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (Basic/Standard/Premium) | `ipConfiguration.subnetId` | `status.outputs.subnet_id` |
| **AzurePublicIp** (Basic/Standard; optional on Premium) | `ipConfiguration.publicIpAddressId` | `status.outputs.public_ip_id` |
| **AzureVirtualNetwork** (Developer only) | `virtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

The host is a session endpoint, not a building block: no downstream Infra Component consumes its outputs via ValueFromRef. `status.outputs` carries `bastion_host_id` and `bastion_host_name` for identification, `dns_name` -- the endpoint sessions connect through (empty until Azure assigns it), and `private_only_enabled` -- whether the host deployed without a public IP. Sessions open from the Azure portal's Connect blade, or from a local terminal via `az network bastion ssh/rdp/tunnel` when tunneling is enabled.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Audited access without per-VM exposure** -- one Basic host per production network, browser sessions for everyone, no public IPs on any VM. Fixed capacity (~40 concurrent sessions), upgrade in place when the team outgrows it. Start from the **Basic Host** preset.

**Terminal-native engineering access** -- a Standard host with tunneling, file copy, and IP connect, sized at 4 scale units. Real ssh/scp ergonomics from a local terminal, and IP connect reaches peered spokes Bastion cannot enumerate -- the shape for a hub network serving many spokes. Start from the **Standard with Tunneling** preset.

**Free dev/test access** -- a Developer host attached straight to one virtual network: no subnet to carve, no IP to allocate, no hourly bill. One connection per user, no peering reach. Start from the **Developer Host** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the host is created in
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the network the host serves; Developer hosts attach to it directly
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- the dedicated `AzureBastionSubnet` that dedicated-infrastructure hosts deploy into
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the Standard static address the host binds exclusively
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the machines sessions reach over their private addresses

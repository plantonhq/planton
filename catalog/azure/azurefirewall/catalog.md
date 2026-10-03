# Azure Firewall

Deploys an Azure Firewall — the managed, stateful network firewall data plane that enforces an AzureFirewallPolicy. The firewall carries WHERE enforcement runs (the dedicated subnet, public IPs, availability zones, deployment model) while the attached policy carries WHAT is enforced (rules, threat intelligence, TLS inspection, IDPS). Its `private_ip_address` output anchors the hub-spoke pattern: spoke route tables send 0.0.0.0/0 to it as a VIRTUAL_APPLIANCE next hop, and everything egresses through one inspected chokepoint. Two deployment models are fixed at creation — AZFW_VNET deploys into a dedicated subnet of your virtual network (named exactly `AzureFirewallSubnet`, /26 or larger), AZFW_HUB into a Virtual WAN hub where Azure manages the addressing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Azure Firewall** -- the firewall instance with its deployment model, tier, zones, IP configurations (or Virtual WAN hub binding), optional management path, and policy attachment
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The subnet, public IPs, and policy are NOT created here — they are referenced Infra Components with their own lifecycles. Azure Firewall provisions and deletes SLOWLY (10-20+ minutes each way) — design changes to avoid replacement.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureSubnet named exactly `AzureFirewallSubnet`** (/26+, no other workloads) in the hub VNet — ARM rejects any other name.
- **A Standard-SKU static AzurePublicIp** for the data path (zone-redundant to match the firewall's zones).
- **An AzureFirewallPolicy whose tier MATCHES** the firewall's — the pair moves together.
- **For forced tunneling or BASIC tier**: a second subnet named exactly `AzureFirewallManagementSubnet` (/26+) with its own public IP — the management block is FIXED at creation.

## Deploy

### Console

Open the deployment store, find **Azure Firewall**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Hub-Spoke Egress Firewall** preset in the [Presets](#presets) tab for the classic chokepoint.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFirewall
metadata:
  name: hub-egress-fw
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-network-hub"
  name: hub-egress-fw
  skuName: AZFW_VNET
  skuTier: STANDARD
  zones: ["1", "2", "3"]
  ipConfigurations:
    - name: primary
      subnetId:
        valueFrom:
          name: hub-vnet-firewall-subnet
      publicIpAddressId:
        valueFrom:
          name: fw-pip
  firewallPolicyId:
    valueFrom:
      name: egress-baseline
  tags:
    cost-center: network-security
```

```shell
planton apply -f firewall.yaml
```

This creates a zone-redundant STANDARD VNet firewall in the hub, enforcing the `egress-baseline` policy through one public IP. An Infra Job tracks the provisioning in real time. After deploy, point spoke route tables' 0.0.0.0/0 at the firewall's `private_ip_address` output — the last step that closes the hub-spoke loop.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the firewall to its dependencies:

```yaml
spec:
  ipConfigurations:
    - name: primary
      subnetId:
        valueFrom:
          kind: AzureSubnet
          name: hub-vnet-firewall-subnet
          fieldPath: status.outputs.subnet_id
      publicIpAddressId:
        valueFrom:
          kind: AzurePublicIp
          name: fw-pip
          fieldPath: status.outputs.public_ip_id
  firewallPolicyId:
    valueFrom:
      kind: AzureFirewallPolicy
      name: egress-baseline
      fieldPath: status.outputs.firewall_policy_id
```

The InfraPipeline resolves the dependency graph — VNet, subnet, public IP, policy, then the firewall — and the route tables that steer traffic reference this firewall's `private_ip_address`.

## Key Configuration

These are the most important decisions when configuring an Azure Firewall. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Deployment model (`skuName`)** -- AZFW_VNET (your subnet, your route tables — the default) or AZFW_HUB (Virtual WAN hub, Azure-managed addressing). Fixed at creation; the wizard reshapes around the choice.

**Tier (`skuTier`)** -- must MATCH the attached policy's tier. BASIC additionally requires the management IP configuration. The tier updates in place; the policy's does not.

**Zones** -- zone redundancy is FREE on Azure Firewall (you pay only the normal deployment cost); spanning all three is the production posture. Fixed at creation.

**IP configurations** -- exactly ONE carries the subnet; every additional configuration adds a public IP (more SNAT ports for high-connection egress, more DNAT frontends). Public IPs add/remove in place; the subnet does not.

**Management path** -- the forced-tunneling enabler: with it, the data path may be fully private (0.0.0.0/0 leaves via ExpressRoute/VPN). FIXED at creation — it cannot be retrofitted.

**Policy attachment** -- the hot-swappable major part: attach, detach, and re-point in place. With a policy attached, ARM rejects firewall-level DNS parameters — the policy's dns block owns them.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `ipConfigurations[].subnetId`, `managementIpConfiguration.subnetId` | `status.outputs.subnet_id` |
| **AzurePublicIp** | `ipConfigurations[].publicIpAddressId`, `managementIpConfiguration.publicIpAddressId` | `status.outputs.public_ip_id` |
| **AzureFirewallPolicy** | `firewallPolicyId` | `status.outputs.firewall_policy_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `private_ip_address` | The firewall's private IP in AzureFirewallSubnet | Route tables' VIRTUAL_APPLIANCE next hop (`routes[].nextHopInIpAddress`) — the hub-spoke seam |
| `management_private_ip_address` | The management path's private IP (empty without the block) | Forced-tunneling diagnostics |
| `virtual_hub_public_ip_addresses` | Azure-assigned public IPs (hub model only) | Partner allowlists |
| `virtual_hub_private_ip_address` | The hub firewall's private IP (hub model only) | Hub route intent |

`firewall_id` and `firewall_name` are also exported for tooling that addresses the firewall as an ARM resource; they have no ValueFromRef consumers.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hub-spoke egress** -- the zone-redundant VNet firewall enforcing the baseline policy, with spoke route tables steering 0.0.0.0/0 through it. Start from the **Hub-Spoke Egress Firewall** preset.

**Forced tunneling** -- a fully-private data path with the management block: all egress leaves via ExpressRoute/VPN to on-premises inspection. Start from the **Forced-Tunneling Firewall (Private Data Path)** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the firewall is created
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- hosts the dedicated AzureFirewallSubnet
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- the exact-name firewall (and management) subnets
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the Standard-SKU static addresses the firewall fronts
- [**Azure Firewall Policy**](/infra-catalog/azure-firewall-policy) -- the rule-and-inspection document this instance enforces
- [**Azure Route Table**](/infra-catalog/azure-route-table) -- steers spoke traffic to this firewall's private IP (the VIRTUAL_APPLIANCE next hop)

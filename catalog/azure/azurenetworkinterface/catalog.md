# Azure Network Interface

Deploys an Azure Network Interface (NIC) — the attachment point that gives a virtual machine its presence in a subnet. The NIC is a first-class resource in Azure's own model: a VM does not contain its network configuration, it **references** one or more NICs (an **AzureVirtualMachine**'s `networkInterfaceIds` consume this NIC's `network_interface_id` output), so network identity — the private address, pool memberships, filtering — outlives any one machine. Load-balancer and Application Gateway membership is expressed HERE, from the member side, and every NSG/ASG/pool membership is realized as its own ARM association resource that never touches the NIC itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Network Interface** -- with one or more IP configurations (a private address in a referenced subnet, dynamic or pinned static, IPv4 or IPv6, optionally fronted by a referenced AzurePublicIp), accelerated networking, IP forwarding, DNS overrides, and the preview NVA-acceleration pair
- **NSG association** -- when a NIC-level network security group is referenced (the per-workload complement to subnet-level filtering)
- **ASG associations** -- one per application security group membership, so NSG rules can target workload groups
- **Pool and NAT-rule associations** -- one per load-balancer backend pool, inbound NAT rule, and Application Gateway pool membership declared on each IP configuration
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The VM-side attachment is NOT created here — which VM holds this NIC lives on the referencing AzureVirtualMachine.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the NIC will be created (usually the workload's own group, beside the VM).
- **A subnet** for each IPv4 configuration — reference an AzureSubnet's `subnet_id` output. All of a NIC's configurations must address subnets of ONE virtual network, in the NIC's region.
- **Optionally**: an AzurePublicIp to front a configuration, an AzureNetworkSecurityGroup for NIC-level filtering, AzureApplicationSecurityGroups to join, and the load balancer / Application Gateway whose name-keyed pool outputs the memberships reference.

## Deploy

### Console

Open the deployment store, find **Azure Network Interface**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Private NIC** preset in the [Presets](#presets) tab for the everyday workload NIC.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureNetworkInterface
metadata:
  name: orders-api-nic
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "orders-rg"
  name: orders-api-nic
  ipConfigurations:
    - name: primary
      subnetId:
        valueFrom:
          kind: AzureSubnet
          name: app-subnet
          fieldPath: status.outputs.subnet_id
      primary: true
  acceleratedNetworkingEnabled: true
```

```shell
planton apply -f nic.yaml
```

This creates a private-only NIC with a dynamic IPv4 address in the referenced subnet and SR-IOV on — ready for the VM that references it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the NIC to its dependencies — and the VM to the NIC:

```yaml
spec:
  ipConfigurations:
    - name: primary
      subnetId:
        valueFrom:
          kind: AzureSubnet
          name: app-subnet
          fieldPath: status.outputs.subnet_id
      loadBalancerBackendAddressPoolIds:
        - valueFrom:
            kind: AzureLoadBalancer
            name: web-lb
            fieldPath: status.outputs.backend_pool_ids.web
```

The InfraPipeline resolves the dependency graph, deploys the subnet and load balancer first, then the NIC with its memberships — and the VM that attaches it deploys after, referencing this NIC's `network_interface_id`.

## Key Configuration

These are the most important decisions when configuring a NIC. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**IP configurations** -- at least one; most NICs carry exactly one. Each is a private address in a subnet: DYNAMIC allocation (Azure picks; stable for the NIC's lifetime — right for virtually all workloads) or STATIC (`privateIpAddress` required exactly then — appliances whose IP is configuration elsewhere). Dual-stack NICs pair an IPv4 configuration with an IPV6 one (which inherits the NIC's subnet placement). With multiple configurations, ARM requires the FIRST to be marked `primary`.

**Member-side load balancing** -- each configuration lists the LB backend pools it joins (`status.outputs.backend_pool_ids.<pool-name>`), the inbound NAT rules it completes, and the Application Gateway pools it joins — Azure's own model, where the pool never lists its members.

**Accelerated networking** -- SR-IOV bypasses the host's virtual switch. Azure defaults it off; production NICs on supported VM sizes (most current sizes with 2+ vCPUs) should enable it.

**IP forwarding** -- enable ONLY on network virtual appliances. A route table pointing at a non-forwarding NIC silently blackholes.

**Filtering** -- `networkSecurityGroupId` attaches a NIC-level NSG (inbound passes the subnet's NSG first, then the NIC's); `applicationSecurityGroupIds` joins workload groups so NSG rules target roles instead of IP ranges.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (per configuration) | `ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| **AzurePublicIp** (optional, per configuration) | `ipConfigurations[].publicIpAddressId` | `status.outputs.public_ip_id` |
| **AzureNetworkSecurityGroup** (optional) | `networkSecurityGroupId` | `status.outputs.network_security_group_id` |
| **AzureApplicationSecurityGroup** (optional, repeated) | `applicationSecurityGroupIds` | `status.outputs.application_security_group_id` |
| **AzureLoadBalancer** (optional, per membership) | `ipConfigurations[].loadBalancerBackendAddressPoolIds`, `ipConfigurations[].loadBalancerInboundNatRuleIds` | `status.outputs.backend_pool_ids.<pool>`, `status.outputs.nat_rule_ids.<rule>` |
| **AzureApplicationGateway** (optional, per membership) | `ipConfigurations[].applicationGatewayBackendAddressPoolIds` | `status.outputs.backend_address_pool_ids.<pool>` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `network_interface_id` | Azure Resource Manager ID of the NIC | AzureVirtualMachine `networkInterfaceIds` — the attachment seam |
| `private_ip_address` | The primary configuration's private IP | Backends, firewall rules, DNS records |
| `private_ip_addresses` | ALL configurations' private IPs, in order | Multi-IP and dual-stack wiring |
| `mac_address` | The NIC's MAC — populated once attached to a RUNNING VM | License servers, appliance registrations |
| `internal_domain_name_suffix` | The VNet's internal DNS suffix | Completes `internal_dns_name_label` into a resolvable FQDN |

`status.outputs` also echoes `network_interface_name` back; attachments and associations all travel by ARM ID.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard workload NIC** -- one dynamic IPv4 configuration in a referenced subnet, accelerated networking on, private-only. Start from the **Standard Private NIC** preset.

**Public-facing edge NIC** -- a configuration fronted by a referenced AzurePublicIp, with a NIC-level NSG. Start from the **Public-Facing NIC with NIC-Level NSG** preset.

**Appliance forwarding NIC** -- a STATIC pinned address with IP forwarding on — the NVA shape route tables point at. Start from the **Network Virtual Appliance NIC (Forwarding + Static IP)** preset.

## Works With

- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- attaches this NIC by its `network_interface_id` output (the FIRST entry in `networkInterfaceIds` is the primary interface)
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- where each configuration's private address lives
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- fronts a configuration for inbound internet traffic
- [**Azure Load Balancer**](/infra-catalog/azure-load-balancer) -- this NIC joins its pools and completes its NAT rules from the member side, through the name-keyed map outputs
- [**Azure Network Security Group**](/infra-catalog/azure-network-security-group) -- NIC-level filtering, in series with the subnet's NSG
- [**Azure Application Security Group**](/infra-catalog/azure-application-security-group) -- workload groups this NIC joins so NSG rules target roles instead of IP ranges
- [**Azure Application Gateway**](/infra-catalog/azure-application-gateway) -- this NIC joins its backend pools from the member side
